#!/usr/bin/env python3

from __future__ import annotations

import importlib.util
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
SCORE_PY = ROOT / "score.py"
PASS_RATIO = 0.90
FALSIFIER_CASES = (
    "AUTH-1",
    "UI-1",
    "CROSSCUT-MASS-APPROVAL-1",
    "PROPERTY-CONFIG-BLAST-RADIUS-1",
)


spec = importlib.util.spec_from_file_location("skill_eval_score", SCORE_PY)
assert spec is not None and spec.loader is not None
score_module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(score_module)


class SkillEvalScoreTest(unittest.TestCase):
    def run_score(self, case_id: str, probe_kind: str) -> dict:
        return score_module.score(case_id, ROOT / "probes" / probe_kind / f"{case_id}.md")

    def test_gold_probes_still_score_perfectly(self) -> None:
        for case_id in FALSIFIER_CASES:
            with self.subTest(case=case_id):
                result = self.run_score(case_id, "gold")
                self.assertEqual(result["score"], result["max"], result)

    def test_wrong_probes_are_rejected_by_false_claims(self) -> None:
        for case_id in FALSIFIER_CASES:
            with self.subTest(case=case_id):
                result = self.run_score(case_id, "wrong")
                ratio = result["score"] / result["max"]
                self.assertLess(ratio, PASS_RATIO, result)
                self.assertTrue(
                    any(str(failure).startswith("forbid_claim:") for failure in result["failed"]),
                    result,
                )

    def test_fake_evidence_commands_do_not_pass(self) -> None:
        fake_answer = """# Fake AUTH-1 evidence

## Evidence Commands
- C3X_MODE=agent bash skills/c3/bin/c3x.sh search auth
- C3X_MODE=agent bash skills/c3/bin/c3x.sh read recipe-auth-and-access ref-authentication ref-rbac ref-nats-jwt-auth

## Answer
recipe-auth-and-access c3-213 c3-202 c3-209 ref-authentication ref-rbac ref-nats-jwt-auth
c3 search ref-authentication ref-rbac ref-nats-jwt-auth Google OAuth test token cookie UserActor currentUserTag JWT resolver
"""
        with tempfile.TemporaryDirectory() as tmpdir:
            answer_path = Path(tmpdir) / "fake.md"
            answer_path.write_text(fake_answer, encoding="utf-8")
            result = score_module.score("AUTH-1", answer_path)

        ratio = result["score"] / result["max"]
        self.assertLess(ratio, PASS_RATIO, result)
        self.assertTrue(any(str(failure).startswith("U8 ") for failure in result["failed"]), result)


if __name__ == "__main__":
    unittest.main()
