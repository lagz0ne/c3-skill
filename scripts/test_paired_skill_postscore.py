#!/usr/bin/env python3

from __future__ import annotations

import hashlib
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

from scripts import paired_skill_eval as paired


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts" / "paired_skill_postscore.py"


def base_result(condition: str) -> dict:
    result = paired.empty_generic_result(
        study_id="STUDY-001", case_id="BR-001", family="blast_radius",
        condition=condition, agent="codex", model="gpt-5.6-sol", trial=1,
        price_version="prices-2026-07-18", reasoning_effort="low",
    )
    result.update({
        "status": "collected", "exit_code": 0, "input_tokens": 100,
        "cached_input_tokens": 20, "output_tokens": 30,
        "reasoning_output_tokens": 10, "total_tokens": 130,
        "effective_tokens": 110, "model_turns": 1, "tool_calls": 3,
        "runtime_guard_reason": "none", "tool_result_bytes": 1000,
        "elapsed_ms": 1000, "setup_elapsed_ms": 10, "cost_usd": 0.1,
        "total_cost_usd": 0.1, "max_tool_calls": 40,
        "max_tool_result_bytes": 200000, "max_output_bytes": 524288,
        "max_answer_words": 400,
    })
    if condition == "with_c3":
        result.update({
            "skill_sha256": "d" * 64, "c3_invocation_count": 4,
            "c3_route_command_count": 1, "c3_impact_command_count": 2,
            "c3_evidence_command_count": 1, "c3_success_count": 4,
            "c3_route_success_count": 1, "c3_impact_success_count": 2,
            "c3_evidence_success_count": 1,
        })
    return result


class PairedSkillPostscoreTest(unittest.TestCase):
    def test_scores_private_manifest_and_emits_generic_rows(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            answer = root / "answer.md"
            answer.write_text('{"current_truth":[],"change_impact":[],"proposed_invariants":[],"unknowns":[]}\n', encoding="utf-8")
            answer_sha = hashlib.sha256(answer.read_bytes()).hexdigest()
            manifest = root / "answers.jsonl"
            manifest.write_text(json.dumps({
                "study_id": "STUDY-001",
                "case_id": "BR-001",
                "trial": 1,
                "condition": "with_c3",
                "answer_sha256": answer_sha,
                "answer_path": str(answer),
                "base_result": base_result("with_c3"),
            }) + "\n", encoding="utf-8")
            cases = root / "cases.jsonl"
            cases.write_text('{"case_id":"BR-001","family":"blast_radius","prompt":"private"}\n', encoding="utf-8")
            score = root / "score.py"
            score.write_text(
                "import json; print(json.dumps({"
                "'quality_score':4.2,'correctness_score':4,'trace_completeness_score':4,"
                "'reasoning_depth_score':4,'grounding_score':4,'no_hallucination_score':5,"
                "'change_usefulness_score':4,'passed':True,'independent_review_count':2,"
                "'deterministic_evidence_count':1,'review_receipt_sha256':['a'*64,'b'*64],"
                "'deterministic_evidence_sha256':['c'*64],'scoring_cost_usd':0.25}))",
                encoding="utf-8",
            )
            output = root / "scores.jsonl"
            result = subprocess.run(
                [
                    sys.executable, str(SCRIPT), "--manifest", str(manifest),
                    "--cases", str(cases), "--score-command",
                    f"{sys.executable} {score} {{answer}} {{cases}} {{case_id}}",
                    "--output", str(output), "--study-id", "STUDY-001",
                ],
                text=True, capture_output=True, check=False,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            row = json.loads(output.read_text(encoding="utf-8"))
            self.assertEqual(row["case_id"], "BR-001")
            self.assertEqual(row["condition"], "with_c3")
            self.assertEqual(row["quality_score"], 4.2)
            self.assertNotIn("answer_path", row)
            paired.validate_generic_result(row)
            self.assertEqual(set(row), paired.GENERIC_RESULT_FIELDS)

    def test_rejects_answer_hash_mismatch(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            answer = root / "answer.md"
            answer.write_text("private\n", encoding="utf-8")
            manifest = root / "answers.jsonl"
            manifest.write_text(json.dumps({
                "study_id": "STUDY-001", "case_id": "BR-001", "trial": 1,
                "condition": "without_c3", "answer_sha256": "0" * 64,
                "answer_path": str(answer),
                "base_result": base_result("without_c3"),
            }) + "\n", encoding="utf-8")
            cases = root / "cases.jsonl"
            cases.write_text('{"case_id":"BR-001","family":"blast_radius","prompt":"private"}\n', encoding="utf-8")
            result = subprocess.run(
                [sys.executable, str(SCRIPT), "--manifest", str(manifest),
                 "--cases", str(cases), "--score-command", "true",
                 "--output", str(root / "scores.jsonl"), "--study-id", "STUDY-001"],
                text=True, capture_output=True, check=False,
            )
            self.assertNotEqual(result.returncode, 0)
            self.assertIn("answer hash", result.stderr)

    def test_removes_partial_output_when_later_review_fails(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            answer = root / "answer.md"
            answer.write_text("private\n", encoding="utf-8")
            digest = hashlib.sha256(answer.read_bytes()).hexdigest()
            manifest = root / "answers.jsonl"
            records = []
            for case_id in ("BR-001", "BR-002"):
                base = base_result("with_c3")
                base["case_id"] = case_id
                records.append({
                    "study_id": "STUDY-001", "case_id": case_id, "trial": 1,
                    "condition": "with_c3", "answer_sha256": digest,
                    "answer_path": str(answer),
                    "base_result": base,
                })
            manifest.write_text("\n".join(json.dumps(row) for row in records) + "\n", encoding="utf-8")
            cases = root / "cases.jsonl"
            cases.write_text("\n".join(
                json.dumps({"case_id": case_id, "family": "blast_radius", "prompt": "private"})
                for case_id in ("BR-001", "BR-002")
            ) + "\n", encoding="utf-8")
            score = root / "score.py"
            score.write_text(
                "import sys; sys.exit(1) if sys.argv[-1] == 'BR-002' else print('{}')",
                encoding="utf-8",
            )
            output = root / "scores.jsonl"
            result = subprocess.run(
                [sys.executable, str(SCRIPT), "--manifest", str(manifest),
                 "--cases", str(cases), "--score-command",
                 f"{sys.executable} {score} {{answer}} {{cases}} {{case_id}}",
                 "--output", str(output), "--study-id", "STUDY-001"],
                text=True, capture_output=True, check=False,
            )
            self.assertNotEqual(result.returncode, 0)
            self.assertFalse(output.exists())
            self.assertFalse(output.with_name(output.name + ".partial").exists())

    def test_rejects_non_object_manifest_and_removes_partial_output(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            manifest = root / "answers.jsonl"
            manifest.write_text("[]\n", encoding="utf-8")
            cases = root / "cases.jsonl"
            cases.write_text("{}\n", encoding="utf-8")
            output = root / "scores.jsonl"
            result = subprocess.run(
                [sys.executable, str(SCRIPT), "--manifest", str(manifest),
                 "--cases", str(cases), "--score-command", "true",
                 "--output", str(output), "--study-id", "STUDY-001"],
                text=True, capture_output=True, check=False,
            )
            self.assertNotEqual(result.returncode, 0)
            self.assertIn("must be an object", result.stderr)
            self.assertFalse(output.exists())
            self.assertFalse(output.with_name(output.name + ".partial").exists())


if __name__ == "__main__":
    unittest.main()
