#!/usr/bin/env python3

from __future__ import annotations

import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

from scripts import paired_skill_eval as paired


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts" / "paired_skill_analyze.py"


def row(case_id: str, family: str, condition: str, quality: float, trial: int = 1) -> dict:
    record = paired.empty_generic_result(
        study_id="STUDY-001", case_id=case_id, family=family, condition=condition,
        agent="codex", model="gpt-5.6-sol", trial=trial,
        price_version="prices-2026-07-18", reasoning_effort="low",
    )
    record.update({
        "status": "scored", "exit_code": 0, "quality_score": quality,
        "correctness_score": 4, "trace_completeness_score": 4,
        "reasoning_depth_score": 4, "grounding_score": 4,
        "no_hallucination_score": 4, "change_usefulness_score": 4,
        "passed": quality >= 4, "input_tokens": 100, "cached_input_tokens": 20,
        "output_tokens": 30, "reasoning_output_tokens": 10, "total_tokens": 130,
        "effective_tokens": 110, "model_turns": 1, "tool_calls": 3 if condition == "with_c3" else 5,
        "runtime_guard_reason": "none", "tool_result_bytes": 1000, "elapsed_ms": 1000,
        "setup_elapsed_ms": 10, "cost_usd": 0.1 if condition == "with_c3" else 0.2,
        "scoring_cost_usd": 0.01, "independent_review_count": 2,
        "deterministic_evidence_count": 1, "review_receipt_sha256": ["a" * 64, "b" * 64],
        "deterministic_evidence_sha256": ["c" * 64], "total_cost_usd": 0.11,
    })
    if condition == "with_c3":
        record.update({
            "skill_sha256": "d" * 64, "c3_invocation_count": 4,
            "c3_route_command_count": 1, "c3_impact_command_count": 2,
            "c3_evidence_command_count": 1, "c3_success_count": 4,
            "c3_route_success_count": 1, "c3_impact_success_count": 2,
            "c3_evidence_success_count": 1,
        })
    return record


class PairedSkillAnalyzeTest(unittest.TestCase):
    def test_computes_paired_quality_and_efficiency(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            source = root / "results.jsonl"
            rows = [
                row("BR-001", "blast_radius", "with_c3", 4.5),
                row("BR-001", "blast_radius", "without_c3", 3.5),
                row("PI-001", "pre_initiative_change_unit", "with_c3", 3.5),
                row("PI-001", "pre_initiative_change_unit", "without_c3", 3.5),
            ]
            source.write_text("\n".join(json.dumps(item) for item in rows) + "\n", encoding="utf-8")
            result = subprocess.run(
                [sys.executable, str(SCRIPT), "--input", str(source), "--study-id", "STUDY-001", "--seed", "7", "--bootstrap", "2000"],
                text=True, capture_output=True, check=False,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            report = json.loads(result.stdout)
            self.assertEqual(report["pair_count"], 2)
            self.assertAlmostEqual(report["quality"]["mean_paired_delta"], 0.5)
            self.assertEqual(report["quality"]["bootstrap"]["seed"], 7)
            self.assertEqual(report["study_validity"]["status"], "below_confirmatory_minimum")
            self.assertFalse(report["study_validity"]["quality_claim_eligible"])
            self.assertEqual(report["study_validity"]["observed_case_families"], ["blast_radius", "pre_initiative_change_unit"])
            self.assertFalse(report["study_validity"]["ci_half_width_target_met"])
            self.assertLess(report["efficiency"]["with_c3"]["tool_calls"], report["efficiency"]["without_c3"]["tool_calls"])

    def test_rejects_unpaired_rows(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            source = root / "results.jsonl"
            source.write_text(json.dumps(row("BR-001", "blast_radius", "with_c3", 4.0)) + "\n", encoding="utf-8")
            result = subprocess.run(
                [sys.executable, str(SCRIPT), "--input", str(source), "--study-id", "STUDY-001"],
                text=True, capture_output=True, check=False,
            )
            self.assertNotEqual(result.returncode, 0)
            self.assertIn("paired", result.stderr)

    def test_rejects_scored_treatment_without_c3_uptake(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            source = root / "results.jsonl"
            treatment = row("BR-001", "blast_radius", "with_c3", 4.0)
            for field in (
                "c3_invocation_count", "c3_route_command_count", "c3_impact_command_count",
                "c3_evidence_command_count", "c3_success_count", "c3_route_success_count",
                "c3_impact_success_count", "c3_evidence_success_count",
            ):
                treatment[field] = 0
            source.write_text(
                "\n".join(json.dumps(item) for item in [
                    treatment,
                    row("BR-001", "blast_radius", "without_c3", 4.0),
                ]) + "\n",
                encoding="utf-8",
            )
            result = subprocess.run(
                [sys.executable, str(SCRIPT), "--input", str(source), "--study-id", "STUDY-001"],
                text=True, capture_output=True, check=False,
            )
            self.assertNotEqual(result.returncode, 0)
            self.assertIn("uptake", result.stderr)

    def test_rejects_cross_pair_run_setting_drift(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            source = root / "results.jsonl"
            rows = [
                row("BR-001", "blast_radius", "with_c3", 4.0),
                row("BR-001", "blast_radius", "without_c3", 4.0),
                row("PI-001", "pre_initiative_change_unit", "with_c3", 4.0),
                row("PI-001", "pre_initiative_change_unit", "without_c3", 4.0),
            ]
            rows[-1]["model"] = "other-model"
            rows[-2]["model"] = "other-model"
            source.write_text("\n".join(json.dumps(item) for item in rows) + "\n", encoding="utf-8")
            result = subprocess.run(
                [sys.executable, str(SCRIPT), "--input", str(source), "--study-id", "STUDY-001"],
                text=True, capture_output=True, check=False,
            )
            self.assertNotEqual(result.returncode, 0)
            self.assertIn("frozen run settings", result.stderr)

    def test_marks_confirmatory_target_only_when_cases_families_and_ci_pass(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            source = root / "results.jsonl"
            rows = []
            for prefix, family in (("BR", "blast_radius"), ("PI", "pre_initiative_change_unit")):
                for number in range(1, 11):
                    case_id = f"{prefix}-{number:03d}"
                    rows.extend([
                        row(case_id, family, "with_c3", 4.0),
                        row(case_id, family, "without_c3", 4.0),
                    ])
            source.write_text("\n".join(json.dumps(item) for item in rows) + "\n", encoding="utf-8")
            result = subprocess.run(
                [sys.executable, str(SCRIPT), "--input", str(source), "--study-id", "STUDY-001", "--bootstrap", "1000"],
                text=True, capture_output=True, check=False,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            report = json.loads(result.stdout)
            self.assertEqual(report["study_validity"]["status"], "eligible_for_confirmatory_quality_gate")
            self.assertTrue(report["study_validity"]["quality_claim_eligible"])


if __name__ == "__main__":
    unittest.main()
