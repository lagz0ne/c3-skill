#!/usr/bin/env python3

from __future__ import annotations

import importlib.util
import tempfile
import unittest
from pathlib import Path


MODULE_PATH = Path(__file__).with_name("judge.py")
SPEC = importlib.util.spec_from_file_location("judge", MODULE_PATH)
assert SPEC and SPEC.loader
judge = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(judge)


class JudgeHarnessTest(unittest.TestCase):
    def test_extract_case_excerpt(self) -> None:
        excerpt = judge.extract_case_excerpt("PROPERTY-FILE-IDEMPOTENCY-1")
        self.assertIn("Expected trace", excerpt)
        self.assertIn("Import idempotency", excerpt)

    def test_build_prompt_includes_case_and_candidate(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            answer = Path(td) / "answer.md"
            answer.write_text("## Evidence commands\nc3 search x\n\n## Answer\ncandidate text\n", encoding="utf-8")
            prompt = judge.build_prompt("AUTH-1", answer)
        self.assertIn("STRICT INDEPENDENT", prompt)
        self.assertIn("How is authentication handled", prompt)
        self.assertIn("candidate text", prompt)

    def test_normalize_recomputes_overall_and_verdict(self) -> None:
        verdict = {
            "case_id": "X",
            "dimensions": {
                name: {"score": 4, "justification": "ok"} for name in judge.DIMENSIONS
            },
            "overall": 1,
            "verdict": "fail",
            "summary": "ok",
            "quality_gaps": [],
        }
        normalized = judge.normalize_verdict(verdict)
        self.assertEqual(normalized["overall"], 4.0)
        self.assertEqual(normalized["verdict"], "pass")


if __name__ == "__main__":
    unittest.main()
