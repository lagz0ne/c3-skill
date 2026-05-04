import json
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import agent_efficiency_eval as ev


class AgentEfficiencyEvalTests(unittest.TestCase):
    def test_default_matrix_has_six_cases_and_two_agents(self):
        self.assertEqual(
            [case.id for case in ev.default_cases()],
            [
                "skill_task_session",
                "skill_content_limit_adr",
                "adr_create",
                "task_session",
                "debug_session",
                "system_design_change",
            ],
        )
        self.assertEqual([agent.id for agent in ev.default_agents()], ["claude", "codex"])

    def test_adr_cases_require_pressure_response(self):
        cases = {case.id: case for case in ev.default_cases()}

        self.assertIn("mentions_pressure_response", cases["adr_create"].accuracy_checks)
        self.assertIn("mentions_pressure_response", cases["skill_content_limit_adr"].accuracy_checks)
        self.assertIn("mentions_component_delta", cases["adr_create"].accuracy_checks)
        self.assertIn("mentions_component_delta", cases["skill_content_limit_adr"].accuracy_checks)
        self.assertIn("Up Cap", cases["adr_create"].prompt)
        self.assertIn("c3x read c3-108 --section Up Cap", cases["adr_create"].prompt)
        self.assertIn("Do not use rg/find/sed/cat", cases["adr_create"].prompt)
        self.assertIn("decompose", cases["skill_content_limit_adr"].prompt)
        self.assertIn("component_delta", cases["skill_content_limit_adr"].prompt)

    def test_dry_run_builds_full_matrix_without_execution(self):
        plan = ev.build_plan(ev.default_cases(), ev.default_agents(), dry_run=True)

        self.assertEqual(len(plan), 12)
        self.assertTrue(all(item.dry_run for item in plan))
        self.assertTrue(all(item.command for item in plan))
        self.assertEqual(
            {(item.agent_id, item.case_id) for item in plan},
            {(agent.id, case.id) for agent in ev.default_agents() for case in ev.default_cases()},
        )

    def test_score_result_handles_missing_token_metadata(self):
        result = ev.RunResult(
            agent_id="codex",
            case_id="debug_session",
            command=["codex", "exec", "prompt"],
            dry_run=False,
            exit_code=0,
            elapsed_ms=1200,
            stdout="task complete\nturn 1\nturn 2\n",
            stderr="",
            output_dir="/tmp/eval",
            artifact_dir="/tmp/artifacts",
            accuracy_checks={"mentioned_root_cause": True, "verified": False},
            token_usage=None,
            turn_count=None,
            trace_metrics={},
        )

        scored = ev.score_result(result)

        self.assertEqual(scored["agent"], "codex")
        self.assertEqual(scored["case"], "debug_session")
        self.assertIsNone(scored["tokens_total"])
        self.assertEqual(scored["turn_count"], 2)
        self.assertEqual(scored["accuracy_score"], 0.5)

    def test_extract_token_usage_reads_codex_json_events(self):
        text = (
            '{"type":"turn.completed","usage":'
            '{"input_tokens":10,"cached_input_tokens":4,'
            '"output_tokens":3,"reasoning_output_tokens":2}}\n'
        )

        usage = ev.extract_token_usage(text)

        self.assertEqual(usage["input_tokens"], 10)
        self.assertEqual(usage["cached_input_tokens"], 4)
        self.assertEqual(usage["output_tokens"], 3)
        self.assertEqual(usage["reasoning_output_tokens"], 2)
        self.assertEqual(usage["total_tokens"], 13)
        self.assertEqual(usage["effective_tokens"], 9)

    def test_extract_turn_count_reads_codex_turn_completion(self):
        text = '{"type":"turn.completed","usage":{"input_tokens":1,"output_tokens":1}}\n'

        self.assertEqual(ev.extract_turn_count(text), 1)

    def test_extract_trace_metrics_finds_c3_commands(self):
        text = 'C3X_MODE=agent bash skills/c3/bin/c3x.sh lookup cli/cmd/list.go\n{"type":"turn.completed"}\n'

        metrics = ev.extract_trace_metrics(text)

        self.assertEqual(metrics["json_event_count"], 1)
        self.assertEqual(metrics["c3_command_count"], 1)
        self.assertIn("lookup cli/cmd/list.go", metrics["c3_command_sequence"][0])

    def test_extract_trace_metrics_ignores_markdown_command_mentions(self):
        text = (
            '{"type":"item.completed","item":{"type":"command_execution","command":"/bin/bash -lc \'rg --files .c3 | head\'",'
            '"aggregated_output":"| c3x check focused mode | Verifies one ADR. |"}}\n'
            "| Surface | Behavior |\n"
            "| c3x check focused mode | Verifies one ADR. |\n"
            "Do not run broad searches with rg or find . during eval.\n"
            "C3X_MODE=agent bash skills/c3/bin/c3x.sh schema adr\n"
        )

        metrics = ev.extract_trace_metrics(text)

        self.assertEqual(metrics["c3_command_sequence"], ["C3X_MODE=agent bash skills/c3/bin/c3x.sh schema adr"])
        self.assertEqual(metrics["c3_command_count"], 1)
        self.assertEqual(metrics["broad_search_count"], 1)

    def test_extract_trace_metrics_counts_actual_command_output_bytes(self):
        text = (
            '{"type":"item.completed","item":{"type":"command_execution","command":"/bin/bash -lc \'C3X_MODE=agent bash skills/c3/bin/c3x.sh schema adr\'",'
            '"aggregated_output":"schema-output","status":"completed"}}\n'
            '{"type":"item.completed","item":{"type":"command_execution","command":"/bin/bash -lc \'rg --files .c3 | head\'",'
            '"aggregated_output":"search-output","status":"completed"}}\n'
            "| c3x check focused mode | Verifies one ADR. |\n"
        )

        metrics = ev.extract_trace_metrics(text)

        self.assertEqual(metrics["tool_output_bytes_total"], len("schema-output") + len("search-output"))
        self.assertEqual(metrics["c3_output_bytes_total"], len("schema-output"))
        self.assertEqual(metrics["transcript_bytes_total"], len(text.encode()))
        self.assertGreater(metrics["c3_command_bytes_total"], 0)

    def test_run_writes_jsonl_records_for_dry_run(self):
        with tempfile.TemporaryDirectory() as tmp:
            out = Path(tmp) / "results.jsonl"

            code = ev.main(["--dry-run", "--output", str(out)])

            self.assertEqual(code, 0)
            records = [json.loads(line) for line in out.read_text().splitlines()]
            self.assertEqual(len(records), 12)
            self.assertTrue(all(record["dry_run"] for record in records))
            self.assertEqual({record["accuracy_score"] for record in records}, {0.0})
            self.assertEqual(
                {record["metric_basis"] for record in records},
                {"tokens_total, turn_count, accuracy_score, elapsed_ms"},
            )
            self.assertTrue(all(Path(record["artifact_dir"]).exists() for record in records))

    def test_live_run_scores_eval_result_before_cleanup(self):
        original_copy = ev._copy_controlled_workspace
        ev._copy_controlled_workspace = lambda workspace: None
        try:
            case = next(case for case in ev.default_cases() if case.id == "task_session")
            item = ev.PlanItem(
                agent_id="fake",
                case_id=case.id,
                command=[
                    sys.executable,
                    "-c",
                    (
                        "import json; "
                        "open('eval_result.json','w').write(json.dumps("
                        "{'verified': True, 'artifacts': ['owner:c3-112']}))"
                    ),
                ],
                prompt=case.prompt,
                dry_run=False,
            )

            scored = ev.score_result(ev.run_plan_item(item, case, keep_workspace=False))

            self.assertEqual(scored["accuracy_score"], 1.0)
            self.assertFalse(Path(scored["output_dir"]).exists())
            self.assertTrue(Path(scored["artifact_dir"], "eval_result.json").exists())
        finally:
            ev._copy_controlled_workspace = original_copy

    def test_evaluate_adr_quality_scores_specific_work_order(self):
        with tempfile.TemporaryDirectory() as tmp:
            workspace = Path(tmp)
            adr = workspace / "adr-20260503-limit-list-content.md"
            adr.write_text(
                "# ADR\n\n"
                "## Goal\nReduce c3x list token cost.\n\n"
                "## Decision\nUse one limited entity projection for agent-mode structured list output.\n\n"
                "## Alternatives Considered\n| Alternative | Rejected because |\n| --- | --- |\n"
                "| Keep broad output | It preserves the token regression. |\n\n"
                "## Risks\n| Risk | Mitigation |\n| --- | --- |\n"
                "| JSON regression | Keep explicit JSON tests. |\n\n"
                "## Verification\n| Check | Result |\n| --- | --- |\n"
                "| go test ./cmd -run TestRunList | proves list output stays bounded |\n"
            )
            (workspace / "eval_result.json").write_text(
                json.dumps(
                    {
                        "summary": "Traced c3x list ownership and created a design ADR.",
                        "verified": [
                            "C3X_MODE=agent bash skills/c3/bin/c3x.sh lookup cli/cmd/list.go",
                            "C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr --only adr-20260503-limit-list-content",
                        ],
                        "artifacts": ["adr-20260503-limit-list-content.md"],
                        "root_cause": "RunList emits too much detail from cli/cmd/list.go by default.",
                        "design_change": "Update cli/cmd/list.go and cli/cmd/list_test.go to keep default list output bounded while preserving explicit detail output.",
                        "pressure_response": "c3-112 should stay additive for now because Up Cap current load is below the reference cap; split if list output refs exceed the cap.",
                        "component_delta": "No new component now: keep c3-112 as owner, but extract c3-121 list-output-budget if Up Cap current load exceeds the references cap.",
                        "adr_id": "adr-20260503-limit-list-content",
                    }
                )
            )

            quality = ev.evaluate_adr_quality(workspace, "irrelevant output")

        self.assertGreaterEqual(quality["score"], 0.85)
        self.assertTrue(quality["checks"]["owner_correct"])
        self.assertTrue(quality["checks"]["verification_executable"])
        self.assertTrue(quality["checks"]["pressure_response_specific"])
        self.assertTrue(quality["checks"]["component_delta_specific"])

    def test_evaluate_adr_quality_penalizes_vague_adr_result(self):
        with tempfile.TemporaryDirectory() as tmp:
            workspace = Path(tmp)
            (workspace / "eval_result.json").write_text(
                json.dumps(
                    {
                        "summary": "Created an ADR.",
                        "verified": True,
                        "artifacts": ["adr-vague.md"],
                        "root_cause": "Improve quality.",
                        "design_change": "Use best practices.",
                        "pressure_response": "Handle pressure as needed.",
                        "component_delta": "Change components as needed.",
                        "adr_id": "adr-20260503-vague",
                    }
                )
            )

            quality = ev.evaluate_adr_quality(workspace, "")

        self.assertLess(quality["score"], 0.70)
        self.assertFalse(quality["checks"]["root_cause_specific"])
        self.assertFalse(quality["checks"]["decision_concrete"])
        self.assertFalse(quality["checks"]["pressure_response_specific"])
        self.assertFalse(quality["checks"]["component_delta_specific"])

    def test_evaluate_accuracy_requires_pressure_response(self):
        case = next(case for case in ev.default_cases() if case.id == "adr_create")
        with tempfile.TemporaryDirectory() as tmp:
            workspace = Path(tmp)
            (workspace / "eval_result.json").write_text(
                json.dumps(
                    {
                        "verified": True,
                        "adr_id": "adr-20260504-token-cost",
                        "pressure_response": "Up Cap pressure means c3-112 should split.",
                    }
                )
            )

            checks = ev.evaluate_accuracy(workspace, case, "", "")

        self.assertTrue(checks["mentions_pressure_response"])
        self.assertFalse(checks["mentions_component_delta"])

    def test_score_result_includes_adr_quality_metric(self):
        with tempfile.TemporaryDirectory() as tmp:
            workspace = Path(tmp)
            (workspace / "eval_result.json").write_text(
                json.dumps(
                    {
                        "summary": "Created an ADR.",
                        "verified": True,
                        "artifacts": ["adr-vague.md"],
                        "root_cause": "Improve quality.",
                        "design_change": "Use best practices.",
                        "adr_id": "adr-20260503-vague",
                    }
                )
            )
            result = ev.RunResult(
                agent_id="codex",
                case_id="adr_create",
                command=["codex", "exec", "prompt"],
                dry_run=False,
                exit_code=0,
                elapsed_ms=1200,
                stdout="",
                stderr="",
                output_dir=str(workspace),
                artifact_dir="/tmp/artifacts",
                accuracy_checks={"has_adr_id": True},
                token_usage={"total_tokens": 100},
                turn_count=1,
                trace_metrics={},
            )

            scored = ev.score_result(result)

        self.assertIn("adr_quality_score", scored)
        self.assertIn("adr_quality_checks", scored)

    def test_score_result_uses_artifact_dir_when_workspace_was_removed(self):
        with tempfile.TemporaryDirectory() as tmp:
            artifact_dir = Path(tmp) / "artifacts"
            artifact_dir.mkdir()
            (artifact_dir / "eval_result.json").write_text(
                json.dumps(
                    {
                        "summary": "Created a bounded ADR.",
                        "verified": True,
                        "artifacts": ["adr-20260504-token-cost.md"],
                        "root_cause": "Agent-mode c3x output repeats token-heavy command prose.",
                        "design_change": "Update cli/cmd output helpers to return compact TOON repair evidence.",
                        "pressure_response": "c3-112 stays additive because Up Cap pressure is below the cap.",
                        "component_delta": "no-delta: keep c3-112 as owner until output helper duplication appears.",
                        "adr_id": "adr-20260504-token-cost",
                    }
                )
            )
            (artifact_dir / "adr-20260504-token-cost.md").write_text(
                "## Alternatives Considered\n\n| Alternative | Rejected because |\n| --- | --- |\n"
                "| Keep broad output | It preserves token pressure. |\n\n"
                "## Risks\n\n| Risk | Mitigation |\n| --- | --- |\n"
                "| Lost repair context | Keep failing entity IDs. |\n\n"
                "## Verification\n\n| Check | Result |\n| --- | --- |\n"
                "| c3x check --include-adr --only adr-20260504-token-cost | passed |\n"
            )
            result = ev.RunResult(
                agent_id="codex",
                case_id="adr_create",
                command=["codex", "exec", "prompt"],
                dry_run=False,
                exit_code=0,
                elapsed_ms=1200,
                stdout="",
                stderr="",
                output_dir=str(Path(tmp) / "removed-workspace"),
                artifact_dir=str(artifact_dir),
                accuracy_checks={"has_adr_id": True},
                token_usage={"total_tokens": 100},
                turn_count=1,
                trace_metrics={},
            )

            scored = ev.score_result(result)

        self.assertGreater(scored["adr_quality_score"], 0.0)

    def test_evaluate_adr_quality_ignores_prompt_without_artifact(self):
        quality = ev.evaluate_adr_quality(
            None,
            "Create one valid ADR with c3x check and go test verification.",
        )

        self.assertEqual(quality["score"], 0.0)
        self.assertFalse(any(quality["checks"].values()))

    def test_threshold_pressure_marks_no_go_runaway(self):
        pressure = ev.evaluate_threshold_pressure(
            case_id="skill_content_limit_adr",
            tokens_total=833186,
            accuracy_score=1.0,
            adr_quality_score=0.8,
            trace_metrics={"broad_search_count": 4, "tool_output_bytes_total": 228676},
        )

        self.assertEqual(pressure["status"], "no_go")
        self.assertEqual(pressure["action"], "stop_or_split")
        self.assertEqual(pressure["target_tokens"], ev.TOKEN_THRESHOLDS["no_go"])
        self.assertEqual(pressure["potential_savings"], 583186)
        self.assertIn("broad_search", pressure["reasons"])

    def test_threshold_pressure_marks_quality_blocked_adr(self):
        pressure = ev.evaluate_threshold_pressure(
            case_id="adr_create",
            tokens_total=177144,
            accuracy_score=1.0,
            adr_quality_score=0.3,
            trace_metrics={"broad_search_count": 0, "tool_output_bytes_total": 12632},
        )

        self.assertEqual(pressure["status"], "soft")
        self.assertEqual(pressure["action"], "fail_quality_or_split")
        self.assertIn("adr_quality_below_floor", pressure["reasons"])

    def test_score_result_includes_threshold_pressure(self):
        result = ev.RunResult(
            agent_id="codex",
            case_id="task_session",
            command=["codex", "exec", "prompt"],
            dry_run=False,
            exit_code=0,
            elapsed_ms=1200,
            stdout="",
            stderr="",
            output_dir="",
            artifact_dir="/tmp/artifacts",
            accuracy_checks={"verified": True},
            token_usage={"total_tokens": 58681, "effective_tokens": 9000},
            turn_count=1,
            trace_metrics={"broad_search_count": 0, "tool_output_bytes_total": 4264},
        )

        scored = ev.score_result(result)

        self.assertEqual(scored["effective_tokens_total"], 9000)
        self.assertEqual(scored["threshold_status"], "ok")
        self.assertEqual(scored["threshold_action"], "accept")
        self.assertEqual(scored["threshold_potential_savings"], 0)

    def test_threshold_pressure_never_accepts_failed_accuracy(self):
        pressure = ev.evaluate_threshold_pressure(
            case_id="task_session",
            tokens_total=50000,
            accuracy_score=0.0,
            adr_quality_score=1.0,
            trace_metrics={},
        )

        self.assertEqual(pressure["status"], "ok")
        self.assertEqual(pressure["action"], "fail_accuracy")
        self.assertIn("accuracy_below_guard", pressure["reasons"])


if __name__ == "__main__":
    unittest.main()
