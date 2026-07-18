#!/usr/bin/env python3

from __future__ import annotations

import json
import hashlib
import os
import subprocess
import tempfile
import unittest
from unittest import mock
from pathlib import Path

from scripts import paired_skill_eval as paired


class PairedSkillEvalTest(unittest.TestCase):
    def pricing(self) -> dict[str, object]:
        return {
            "version": "prices-2026-07-14",
            "currency": "USD",
            "observed_at": "2026-07-14T00:00:00Z",
            "models": [
                {
                    "id": "cheap-model",
                    "provider": "codex",
                    "input_per_million_usd": 0.20,
                    "cached_input_per_million_usd": 0.02,
                    "output_per_million_usd": 0.80,
                    "eligible": True,
                    "live_allowed": True,
                },
                {
                    "id": "expensive-model",
                    "provider": "codex",
                    "input_per_million_usd": 2.00,
                    "cached_input_per_million_usd": 0.20,
                    "output_per_million_usd": 8.00,
                    "eligible": True,
                    "live_allowed": True,
                },
                {
                    "id": "example-only-zero-price",
                    "provider": "codex",
                    "input_per_million_usd": 0.0,
                    "cached_input_per_million_usd": 0.0,
                    "output_per_million_usd": 0.0,
                    "eligible": True,
                    "live_allowed": False,
                },
            ],
        }

    def cases(self) -> list[paired.EvalCase]:
        return [
            paired.EvalCase("BR-001", "blast_radius", "What changes if X moves?"),
            paired.EvalCase("PI-001", "pre_initiative_change_unit", "Shape the next change unit."),
        ]

    def test_selects_cheapest_live_eligible_model(self) -> None:
        model = paired.select_cheapest_model(
            self.pricing(),
            provider="codex",
            estimated_input_tokens=100_000,
            estimated_output_tokens=20_000,
        )

        self.assertEqual(model.model_id, "cheap-model")
        self.assertAlmostEqual(model.estimated_cost_usd, 0.036)

    def test_selects_requested_live_model_instead_of_cheapest(self) -> None:
        model = paired.select_model(
            self.pricing(),
            provider="codex",
            model_id="expensive-model",
            estimated_input_tokens=100_000,
            estimated_output_tokens=20_000,
        )

        self.assertEqual(model.model_id, "expensive-model")
        self.assertAlmostEqual(model.estimated_cost_usd, 0.36)

        with self.assertRaisesRegex(ValueError, "not live-allowed"):
            paired.select_model(
                self.pricing(),
                provider="codex",
                model_id="example-only-zero-price",
                estimated_input_tokens=100_000,
                estimated_output_tokens=20_000,
            )

    def test_claude_run_cost_prefers_provider_report_over_lower_estimate(self) -> None:
        usage = {"input_tokens": 100, "cached_input_tokens": 0, "output_tokens": 10}
        model = paired.SelectedModel("opus", "claude", "v1", 5.0, 0.5, 25.0, 0.1)

        self.assertEqual(paired.observed_cost_usd(usage, model, reported_cost_usd=0.42), 0.42)

    def test_budget_preflight_rejects_run_count_per_run_and_total(self) -> None:
        policy = paired.BudgetPolicy(
            max_runs=4,
            max_total_cost_usd=1.00,
            max_cost_per_run_usd=0.30,
            max_tokens_per_run=250_000,
            timeout_seconds=900,
        )

        with self.assertRaisesRegex(paired.BudgetExceeded, "planned runs"):
            paired.check_budget(policy, planned_runs=5, estimated_cost_per_run=0.10)
        with self.assertRaisesRegex(paired.BudgetExceeded, "per-run"):
            paired.check_budget(policy, planned_runs=2, estimated_cost_per_run=0.31)
        with self.assertRaisesRegex(paired.BudgetExceeded, "total"):
            paired.check_budget(policy, planned_runs=4, estimated_cost_per_run=0.26)

    def test_observe_cost_only_disables_cost_and_token_admission_walls(self) -> None:
        policy = paired.BudgetPolicy(
            max_runs=100,
            max_total_cost_usd=0.0,
            max_cost_per_run_usd=0.0,
            max_tokens_per_run=1,
            observe_cost_only=True,
        )

        paired.check_budget(policy, planned_runs=100, estimated_cost_per_run=100.0)

    def test_pair_manifests_differ_only_by_treatment(self) -> None:
        common = paired.CommonRunManifest(
            case_id="BR-001",
            family="blast_radius",
            trial=1,
            provider="codex",
            model="cheap-model",
            repo_snapshot_sha256="a" * 64,
            task_sha256="b" * 64,
            timeout_seconds=900,
            max_tokens=250_000,
            baseline_instruction_sha256="d" * 64,
            max_tool_calls=6,
            max_output_bytes=524_288,
        )
        control = paired.build_arm_manifest(common, "without_c3", skill_sha256=None)
        treatment = paired.build_arm_manifest(common, "with_c3", skill_sha256="c" * 64)

        self.assertEqual(paired.non_treatment_differences(control, treatment), {})
        self.assertTrue(control["treatment"]["provider_tool_set_parity"])
        self.assertTrue(treatment["treatment"]["provider_tool_set_parity"])
        self.assertEqual(
            treatment["treatment"]["uptake_acceptance"],
            "supervisor_transcript_exact_first_command_exit_zero",
        )
        self.assertEqual(treatment["treatment"]["estimand"], "c3_treatment_package")
        self.assertEqual(
            treatment["treatment"]["bootstrap_internal_c3_call_count"],
            {"minimum": 4, "search_miss_fallback": 5},
        )
        self.assertRegex(treatment["treatment"]["bootstrap_sha256"], r"^[0-9a-f]{64}$")
        changed = json.loads(json.dumps(treatment))
        changed["common"]["timeout_seconds"] = 800
        self.assertEqual(
            paired.non_treatment_differences(control, changed),
            {"timeout_seconds": [900, 800]},
        )

    def test_runner_freeze_environment_pins_target_prompt_and_full_c3(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            repo = root / "seed"
            repo.mkdir()
            subprocess.run(["git", "init", "-q", str(repo)], check=True)
            subprocess.run(["git", "-C", str(repo), "config", "user.email", "eval@example.invalid"], check=True)
            subprocess.run(["git", "-C", str(repo), "config", "user.name", "Eval"], check=True)
            (repo / "README.md").write_text("seed\n", encoding="utf-8")
            subprocess.run(["git", "-C", str(repo), "add", "."], check=True)
            subprocess.run(["git", "-C", str(repo), "commit", "-qm", "seed"], check=True)
            prompt = root / "prompt.md"
            prompt.write_text("Assess the change.\n", encoding="utf-8")

            freeze = paired.runner_freeze_environment(repo, prompt)

        self.assertEqual(
            set(freeze),
            {
                "C3_EXPECT_PROMPT_SHA256",
                "C3_EXPECT_SEED_HEAD_SHA256",
                "C3_EXPECT_SELECTED_BINARY_SHA256",
                "C3_EXPECT_SKILL_MD_SHA256",
                "C3_EXPECT_SKILL_TREE_SHA256",
                "C3_EXPECT_WRAPPER_SHA256",
                "C3_EXPECT_VERSION_SHA256",
            },
        )
        for value in freeze.values():
            self.assertRegex(value, r"^[0-9a-f]{64}$")

    def test_runner_freeze_environment_uses_explicit_frozen_binary(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            repo = root / "seed"
            repo.mkdir()
            subprocess.run(["git", "init", "-q", str(repo)], check=True)
            subprocess.run(["git", "-C", str(repo), "config", "user.email", "eval@example.invalid"], check=True)
            subprocess.run(["git", "-C", str(repo), "config", "user.name", "Eval"], check=True)
            (repo / "README.md").write_text("seed\n", encoding="utf-8")
            subprocess.run(["git", "-C", str(repo), "add", "."], check=True)
            subprocess.run(["git", "-C", str(repo), "commit", "-qm", "seed"], check=True)
            frozen = root / "frozen-c3x"
            frozen.write_bytes(b"portable-freeze")
            with mock.patch.dict(os.environ, {"C3_FROZEN_BINARY": str(frozen)}):
                freeze = paired.runner_freeze_environment(repo)

        self.assertEqual(
            freeze["C3_EXPECT_SELECTED_BINARY_SHA256"],
            hashlib.sha256(b"portable-freeze").hexdigest(),
        )

    def test_both_prompts_share_the_same_bounded_discovery_rule(self) -> None:
        case = self.cases()[0]

        control = paired.render_prompt(case, "without_c3")
        treatment = paired.render_prompt(case, "with_c3")

        rule = "Use at most 6 discovery tool calls"
        self.assertIn(rule, control)
        self.assertIn(rule, treatment)
        self.assertIn("both the research", control)
        self.assertIn("runtime ceiling", control)
        self.assertIn("do not name, quote", control)
        self.assertIn("analysis tool", treatment)

    def test_both_prompts_share_current_truth_first_answer_contract(self) -> None:
        case = self.cases()[0]

        control = paired.render_prompt(case, "without_c3")
        treatment = paired.render_prompt(case, "with_c3")

        contract = (
            "current_truth, change_impact, proposed_invariants, and unknowns"
        )
        self.assertIn(contract, control)
        self.assertIn(contract, treatment)
        self.assertLess(control.index("current_truth"), control.index("change_impact"))
        self.assertLess(control.index("change_impact"), control.index("proposed_invariants"))
        self.assertIn("Target at most 220 words", control)
        self.assertIn("Target at most 220 words", treatment)
        self.assertIn("source anchor", control)
        self.assertIn("Do not present a proposal as current behavior", control)

    def test_prompt_leak_audit_detects_enumerated_registered_lanes(self) -> None:
        prompt = (
            "Assess temporary access across assignment writes, active reads, "
            "deletion checks, owner checks, revocation, and audit behavior."
        )
        lanes = [
            "assignment writes",
            "active reads",
            "deletion checks",
            "owner checks",
            "revocation",
            "audit behavior",
        ]

        audit = paired.audit_prompt_leakage(prompt, lanes)

        self.assertFalse(audit["passed"])
        self.assertEqual(audit["matched_lane_count"], 6)
        self.assertEqual(audit["registered_lane_count"], 6)
        self.assertGreaterEqual(audit["matched_lane_ratio"], 0.8)

    def test_prompt_leak_audit_allows_natural_initiative_wording(self) -> None:
        audit = paired.audit_prompt_leakage(
            "We want temporary access that expires safely. What could this change affect?",
            [
                "assignment writes",
                "active reads",
                "deletion checks",
                "owner checks",
                "revocation",
                "audit behavior",
            ],
        )

        self.assertTrue(audit["passed"])
        self.assertEqual(audit["matched_lane_count"], 0)

    def test_case_loader_rejects_prompt_that_leaks_registered_gold_lanes(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            cases = Path(temp) / "cases.jsonl"
            cases.write_text(
                json.dumps({
                    "case_id": "BR-001",
                    "family": "blast_radius",
                    "prompt": "Trace assignment writes, active reads, deletion checks, and owner checks.",
                    "registered_gold_lanes": [
                        "assignment writes",
                        "active reads",
                        "deletion checks",
                        "owner checks",
                    ],
                }) + "\n",
                encoding="utf-8",
            )

            with self.assertRaisesRegex(ValueError, "prompt leaks 4 of 4"):
                paired.load_cases(cases)

    def test_run_parser_accepts_private_answer_retention_pair(self) -> None:
        args = paired.parse_args([
            "run", "--run", "--private-answer-manifest", "/tmp/private.jsonl",
            "--private-answer-dir", "/tmp/private-answers",
        ])

        self.assertEqual(args.private_answer_manifest, "/tmp/private.jsonl")
        self.assertEqual(args.private_answer_dir, "/tmp/private-answers")
        self.assertEqual(args.max_answer_words, 250)

    def test_answer_word_limit_is_configurable_and_fails_closed(self) -> None:
        with tempfile.TemporaryDirectory() as temp:
            answer = Path(temp) / "answer.md"
            answer.write_text("one two\nthree four five\n", encoding="utf-8")

            self.assertEqual(paired.answer_word_limit_deviations(answer, 5), [])
            self.assertEqual(
                paired.answer_word_limit_deviations(answer, 4),
                ["ANSWER_WORD_LIMIT_BREACH"],
            )

        args = paired.parse_args(["run", "--run", "--max-answer-words", "125"])
        self.assertEqual(args.max_answer_words, 125)

    def test_answer_limit_breach_is_checked_before_private_retention(self) -> None:
        source = (paired.ROOT / "scripts" / "paired_skill_eval.py").read_text(
            encoding="utf-8"
        )

        guard = source.index("answer_word_limit_deviations(answer_path")
        retention = source.index("if private_manifest_path and private_answer_dir:")
        self.assertLess(guard, retention)

    def test_prompt_uses_the_runtime_limit_as_discovery_authority(self) -> None:
        prompt = paired.render_prompt(self.cases()[0], "without_c3", max_tool_calls=3)

        self.assertIn("Use at most 3 discovery tool calls", prompt)
        self.assertIn("both the research", prompt)
        self.assertIn("runtime ceiling", prompt)
        self.assertNotIn("accounting margin", prompt)

    def test_treatment_prompt_has_one_unambiguous_c3_entrypoint(self) -> None:
        prompt = paired.render_prompt(self.cases()[0], "with_c3")

        self.assertIn('bash /opt/c3/bin/c3-impact-bootstrap "<short behavior or domain>"', prompt)
        self.assertIn("Run that bootstrap once before ordinary repository tools", prompt)
        self.assertNotIn("use `$c3`", prompt)
        self.assertNotIn("C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh <command>", prompt)

    def test_generic_retention_rejects_unknown_fields_and_descriptive_case_ids(self) -> None:
        valid = paired.empty_generic_result(
            study_id="STUDY-001",
            case_id="BR-001",
            family="blast_radius",
            condition="with_c3",
            agent="claude",
            model="cheap-model",
            trial=1,
            price_version="prices-2026-07-14",
        )
        valid["skill_sha256"] = "c" * 64

        paired.validate_generic_result(valid)
        self.assertEqual(valid["agent"], "claude")
        self.assertEqual(valid["max_answer_words"], 250)

        leaked = dict(valid)
        leaked["answer"] = "The invoice service owns this."
        with self.assertRaisesRegex(ValueError, "unknown retained fields"):
            paired.validate_generic_result(leaked)

        descriptive = dict(valid)
        descriptive["case_id"] = "CHANGE-NATS-SUBJECT-PREFIX"
        with self.assertRaisesRegex(ValueError, "opaque"):
            paired.validate_generic_result(descriptive)

    def test_c3_uptake_is_required_only_for_the_treatment(self) -> None:
        complete = {
            "c3_invocation_count": "7",
            "c3_route_command_count": "3",
            "c3_impact_command_count": "1",
            "c3_evidence_command_count": "2",
            "c3_success_count": "4",
            "c3_route_success_count": "1",
            "c3_impact_success_count": "1",
            "c3_evidence_success_count": "1",
        }

        transcript = (
            '{"type":"item.completed","item":{"type":"command_execution",'
            '"command":"/usr/bin/zsh -lc \\"bash /opt/c3/bin/c3-impact-bootstrap approval\\"",'
            '"exit_code":0,"aggregated_output":"ok"}}'
        )
        self.assertEqual(paired.c3_uptake_deviations("with_c3", complete, transcript), [])
        self.assertEqual(
            paired.c3_uptake_deviations("with_c3", {}, transcript),
            [],
        )
        self.assertEqual(
            paired.c3_uptake_deviations("without_c3", complete, ""),
            ["C3_CONTROL_CONTAMINATION"],
        )
        missing_evidence_success = dict(complete)
        missing_evidence_success["c3_evidence_success_count"] = "0"
        self.assertEqual(
            paired.c3_uptake_deviations("with_c3", missing_evidence_success, transcript),
            [],
        )

    def test_c3_uptake_requires_a_successful_bootstrap_in_the_supervisor_transcript(self) -> None:
        complete = {
            "c3_invocation_count": "3",
            "c3_route_command_count": "1",
            "c3_impact_command_count": "1",
            "c3_evidence_command_count": "1",
            "c3_success_count": "3",
            "c3_route_success_count": "1",
            "c3_impact_success_count": "1",
            "c3_evidence_success_count": "1",
        }
        forged_shell_output = (
            '{"type":"item.completed","item":{"type":"command_execution",'
            '"command":"/usr/bin/zsh -lc \\"printf bootstrap\\"","exit_code":0,'
            '"aggregated_output":"bash /opt/c3/bin/c3-impact-bootstrap approval"}}'
        )
        failed_bootstrap = (
            '{"type":"item.completed","item":{"type":"command_execution",'
            '"command":"/usr/bin/zsh -lc \\"bash /opt/c3/bin/c3-impact-bootstrap approval\\"",'
            '"exit_code":1,"aggregated_output":"failed"}}'
        )

        self.assertEqual(
            paired.c3_uptake_deviations("with_c3", complete, forged_shell_output),
            ["C3_UPTAKE_MISSING"],
        )
        self.assertEqual(
            paired.c3_uptake_deviations("with_c3", complete, failed_bootstrap),
            ["C3_UPTAKE_MISSING"],
        )

        prefixed_override = (
            '{"type":"item.completed","item":{"type":"command_execution",'
            '"command":"/usr/bin/zsh -lc \\"C3_BOOTSTRAP_WRAPPER=/tmp/fake bash '
            '/opt/c3/bin/c3-impact-bootstrap approval\\"","exit_code":0,'
            '"aggregated_output":"ok"}}'
        )
        prior_inspection = (
            '{"type":"item.completed","item":{"type":"command_execution",'
            '"command":"/usr/bin/zsh -lc \\"rg approval .\\"","exit_code":0}}\n'
            + (
                '{"type":"item.completed","item":{"type":"command_execution",'
                '"command":"/usr/bin/zsh -lc \\"bash /opt/c3/bin/c3-impact-bootstrap approval\\"",'
                '"exit_code":0}}'
            )
        )
        self.assertEqual(
            paired.c3_uptake_deviations("with_c3", complete, prefixed_override),
            ["C3_UPTAKE_MISSING"],
        )
        self.assertEqual(
            paired.c3_uptake_deviations("with_c3", complete, prior_inspection),
            ["C3_UPTAKE_MISSING"],
        )

        for command in (
            'bash /opt/c3/bin/c3-impact-bootstrap approval; cat README.md',
            'bash /opt/c3/bin/c3-impact-bootstrap "$(cat README.md)"',
            'bash /opt/c3/bin/c3-impact-bootstrap approval && cat README.md',
            'bash /opt/c3/bin/c3-impact-bootstrap approval | cat',
        ):
            chained = json.dumps(
                {
                    "type": "item.completed",
                    "item": {
                        "type": "command_execution",
                        "command": f"/usr/bin/zsh -lc {json.dumps(command)}",
                        "exit_code": 0,
                    },
                }
            )
            self.assertEqual(
                paired.c3_uptake_deviations("with_c3", complete, chained),
                ["C3_UPTAKE_MISSING"],
                command,
            )

    def test_score_receipts_require_distinct_replayable_hashes(self) -> None:
        score = {
            "quality_score": 4.0,
            "correctness_score": 4,
            "trace_completeness_score": 4,
            "reasoning_depth_score": 4,
            "grounding_score": 4,
            "no_hallucination_score": 4,
            "change_usefulness_score": 4,
            "passed": True,
            "independent_review_count": 2,
            "deterministic_evidence_count": 1,
            "review_receipt_sha256": ["a" * 64, "b" * 64],
            "deterministic_evidence_sha256": ["c" * 64],
            "scoring_cost_usd": 0.01,
        }

        paired.validate_score(score)
        duplicate = dict(score)
        duplicate["review_receipt_sha256"] = ["a" * 64, "a" * 64]
        with self.assertRaisesRegex(ValueError, "distinct review receipt"):
            paired.validate_score(duplicate)

    def test_sol_high_single_review_with_deterministic_evidence_is_valid(self) -> None:
        score = {
            "quality_score": 4.0,
            "correctness_score": 4,
            "trace_completeness_score": 4,
            "reasoning_depth_score": 4,
            "grounding_score": 4,
            "no_hallucination_score": 4,
            "change_usefulness_score": 4,
            "passed": True,
            "independent_review_count": 1,
            "deterministic_evidence_count": 1,
            "review_receipt_sha256": ["a" * 64],
            "deterministic_evidence_sha256": ["c" * 64],
            "scoring_cost_usd": 0.01,
        }
        paired.validate_score(score)

    def test_plan_uses_one_model_for_both_arms_and_stays_under_budget(self) -> None:
        policy = paired.BudgetPolicy(
            max_runs=24,
            max_total_cost_usd=5.0,
            max_cost_per_run_usd=0.5,
            max_tokens_per_run=250_000,
            timeout_seconds=900,
        )

        plan = paired.build_plan(
            self.cases(),
            pricing=self.pricing(),
            provider="codex",
            repeats=2,
            policy=policy,
            estimated_input_tokens=100_000,
            estimated_output_tokens=20_000,
        )

        self.assertTrue(plan["dry_run"])
        self.assertEqual(plan["planned_runs"], 8)
        self.assertEqual(plan["selected_model"], "cheap-model")
        self.assertLessEqual(plan["estimated_total_cost_usd"], 5.0)
        self.assertEqual({row["model"] for row in plan["runs"]}, {"cheap-model"})
        self.assertEqual({row["condition"] for row in plan["runs"]}, {"with_c3", "without_c3"})
        self.assertEqual({row["max_tool_calls"] for row in plan["runs"]}, {6})
        self.assertEqual({row["max_output_bytes"] for row in plan["runs"]}, {524_288})

    def test_plan_records_fixed_model_and_reasoning_effort_for_both_arms(self) -> None:
        plan = paired.build_plan(
            [self.cases()[0]],
            pricing=self.pricing(),
            provider="codex",
            model_id="expensive-model",
            reasoning_effort="high",
            repeats=1,
            policy=paired.BudgetPolicy(max_runs=2, max_total_cost_usd=1.0),
            estimated_input_tokens=100_000,
            estimated_output_tokens=20_000,
        )

        self.assertEqual(plan["selected_model"], "expensive-model")
        self.assertEqual(plan["reasoning_effort"], "high")
        self.assertEqual({row["model"] for row in plan["runs"]}, {"expensive-model"})
        self.assertEqual({row["reasoning_effort"] for row in plan["runs"]}, {"high"})

    def test_cli_plan_does_not_print_repo_path_or_prompts(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            repo = root / "private-project-name"
            repo.mkdir()
            subprocess.run(["git", "init", "-q", str(repo)], check=True)
            subprocess.run(["git", "-C", str(repo), "config", "user.email", "eval@example.invalid"], check=True)
            subprocess.run(["git", "-C", str(repo), "config", "user.name", "Eval"], check=True)
            (repo / "README.md").write_text("private project\n", encoding="utf-8")
            subprocess.run(["git", "-C", str(repo), "add", "README.md"], check=True)
            subprocess.run(["git", "-C", str(repo), "commit", "-qm", "seed"], check=True)
            cases = root / "private-cases.jsonl"
            cases.write_text(
                json.dumps({"case_id": "BR-001", "family": "blast_radius", "prompt": "Secret project prompt"}) + "\n",
                encoding="utf-8",
            )
            pricing = root / "pricing.json"
            pricing.write_text(json.dumps(self.pricing()), encoding="utf-8")

            result = subprocess.run(
                [
                    "python3",
                    "scripts/paired_skill_eval.py",
                    "plan",
                    "--repo",
                    str(repo),
                    "--cases",
                    str(cases),
                    "--pricing",
                    str(pricing),
                    "--repeat",
                    "1",
                ],
                cwd=Path(__file__).resolve().parents[1],
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
            )

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertNotIn("private-project-name", result.stdout)
        self.assertNotIn("private-cases", result.stdout)
        self.assertNotIn("Secret project prompt", result.stdout)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["planned_runs"], 2)

    def test_live_run_requires_explicit_run_flag(self) -> None:
        result = subprocess.run(
            ["python3", "scripts/paired_skill_eval.py", "run"],
            cwd=Path(__file__).resolve().parents[1],
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )

        self.assertNotEqual(result.returncode, 0)
        self.assertIn("requires --run", result.stderr)

    def test_fake_live_pair_scores_then_retains_generic_rows_only(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            repo = root / "private-project-name"
            repo.mkdir()
            subprocess.run(["git", "init", "-q", str(repo)], check=True)
            subprocess.run(["git", "-C", str(repo), "config", "user.email", "eval@example.invalid"], check=True)
            subprocess.run(["git", "-C", str(repo), "config", "user.name", "Eval"], check=True)
            (repo / "README.md").write_text("private project\n", encoding="utf-8")
            subprocess.run(["git", "-C", str(repo), "add", "README.md"], check=True)
            subprocess.run(["git", "-C", str(repo), "commit", "-qm", "seed"], check=True)
            cases = root / "private-cases.jsonl"
            cases.write_text(
                json.dumps({"case_id": "BR-001", "family": "blast_radius", "prompt": "Secret project prompt"}) + "\n",
                encoding="utf-8",
            )
            pricing = root / "pricing.json"
            pricing.write_text(json.dumps(self.pricing()), encoding="utf-8")
            runner = root / "fake-runner.sh"
            baseline_sha = __import__("hashlib").sha256(paired.BASELINE_INSTRUCTIONS.read_bytes()).hexdigest()
            runner.write_text(
                """#!/usr/bin/env bash
set -euo pipefail
run_dir=""
label=""
condition=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --run-dir) run_dir="$2"; shift 2 ;;
    --label) label="$2"; shift 2 ;;
    --condition) condition="$2"; shift 2 ;;
    *) shift ;;
  esac
done
mkdir -p "$run_dir"
printf 'A grounded private answer.\n' > "$run_dir/$label.md"
printf '%s\n' '{"type":"turn.completed","usage":{"input_tokens":1000,"cached_input_tokens":100,"output_tokens":200,"reasoning_output_tokens":50,"total_tokens":1200}}' > "$run_dir/$label.stdout.txt"
: > "$run_dir/$label.stderr.txt"
printf 'condition=fake\n' > "$run_dir/$label.meta.txt"
printf 'baseline_instruction_sha256=__BASELINE_SHA__\n' >> "$run_dir/$label.meta.txt"
printf 'runtime_guard_status=completed\n' >> "$run_dir/$label.meta.txt"
printf 'runtime_guard_reason=none\n' >> "$run_dir/$label.meta.txt"
printf 'runtime_guard_tool_calls=1\n' >> "$run_dir/$label.meta.txt"
if [[ "$condition" == "with_c3" ]]; then
  printf '%s\n' '{"type":"item.completed","item":{"type":"command_execution","command":"/usr/bin/zsh -lc \\"bash /opt/c3/bin/c3-impact-bootstrap approval\\"","exit_code":0,"aggregated_output":"ok"}}' >> "$run_dir/$label.stdout.txt"
  printf 'treatment_instruction_sha256=__TREATMENT_SHA__\n' >> "$run_dir/$label.meta.txt"
  printf 'treatment_runtime_layer=codex_developer_instructions\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_invocation_count=3\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_route_command_count=1\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_impact_command_count=1\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_evidence_command_count=1\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_success_count=3\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_route_success_count=1\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_impact_success_count=1\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_evidence_success_count=1\n' >> "$run_dir/$label.meta.txt"
else
  printf 'treatment_instruction_sha256=unmounted\n' >> "$run_dir/$label.meta.txt"
  printf 'treatment_runtime_layer=none\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_invocation_count=0\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_route_command_count=0\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_impact_command_count=0\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_evidence_command_count=0\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_success_count=0\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_route_success_count=0\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_impact_success_count=0\n' >> "$run_dir/$label.meta.txt"
  printf 'c3_evidence_success_count=0\n' >> "$run_dir/$label.meta.txt"
fi
""".replace("__BASELINE_SHA__", baseline_sha).replace(
                    "__TREATMENT_SHA__",
                    __import__("hashlib").sha256(paired.C3_TREATMENT_INSTRUCTIONS.read_bytes()).hexdigest(),
                ),
                encoding="utf-8",
            )
            runner.chmod(0o755)
            scorer = root / "fake-scorer.py"
            scorer.write_text(
                """#!/usr/bin/env python3
import json
print(json.dumps({
  "quality_score": 4.2,
  "correctness_score": 4,
  "trace_completeness_score": 4,
  "reasoning_depth_score": 4,
  "grounding_score": 4,
  "no_hallucination_score": 5,
  "change_usefulness_score": 4,
  "passed": True,
  "independent_review_count": 2,
  "deterministic_evidence_count": 1,
  "review_receipt_sha256": ["aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"],
  "deterministic_evidence_sha256": ["cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"],
  "scoring_cost_usd": 0.01
}))
""",
                encoding="utf-8",
            )
            scorer.chmod(0o755)
            results = root / "generic-results.jsonl"

            run = subprocess.run(
                [
                    "python3",
                    "scripts/paired_skill_eval.py",
                    "run",
                    "--run",
                    "--repo",
                    str(repo),
                    "--cases",
                    str(cases),
                    "--pricing",
                    str(pricing),
                    "--runner",
                    str(runner),
                    "--score-command",
                    f"{scorer} {{answer}} {{cases}} {{case_id}}",
                    "--results",
                    str(results),
                    "--study-id",
                    "STUDY-001",
                    "--repeat",
                    "1",
                    "--auth",
                    "env",
                    "--random-seed",
                    "7",
                ],
                cwd=Path(__file__).resolve().parents[1],
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
            )

            self.assertEqual(run.returncode, 0, run.stderr)
            result_text = results.read_text(encoding="utf-8")
            rows = [json.loads(line) for line in result_text.splitlines()]

        self.assertEqual(len(rows), 2)
        self.assertEqual({row["condition"] for row in rows}, {"with_c3", "without_c3"})
        self.assertEqual({row["model"] for row in rows}, {"cheap-model"})
        self.assertTrue(all(row["status"] == "scored" for row in rows))
        self.assertTrue(all(row["cost_usd"] > 0 for row in rows))
        self.assertEqual({row["scoring_cost_usd"] for row in rows}, {0.01})
        self.assertEqual({row["independent_review_count"] for row in rows}, {2})
        self.assertTrue(all(len(row["review_receipt_sha256"]) == 2 for row in rows))
        self.assertTrue(all(row["total_cost_usd"] == row["cost_usd"] + 0.01 for row in rows))
        self.assertTrue(all(set(row) == paired.GENERIC_RESULT_FIELDS for row in rows))
        self.assertEqual({row["reasoning_effort"] for row in rows}, {"medium"})
        self.assertNotIn("Secret project prompt", result_text)
        self.assertNotIn("private-project-name", result_text)

    def test_deferred_scoring_collects_answers_without_fake_review_receipts(self) -> None:
        source = paired.ROOT / "scripts" / "paired_skill_eval.py"
        text = source.read_text(encoding="utf-8")

        self.assertIn('--defer-scoring', text)
        self.assertIn('"collected" if args.defer_scoring else "scored"', text)
        self.assertIn("if not args.defer_scoring:", text)


if __name__ == "__main__":
    unittest.main()
