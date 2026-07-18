import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts/validate_c3_compliance_review.py"
HANDLE = "ref-one#n1@v1:sha256:" + "a" * 64 + ' "governance"'


class ValidateC3ComplianceReviewTests(unittest.TestCase):
    def run_validator(self, *, omit_second: bool = False) -> subprocess.CompletedProcess[str]:
        self.tempdir = tempfile.TemporaryDirectory()
        root = Path(self.tempdir.name)
        plan = root / "plan.json"
        plan.write_text(
            json.dumps(
                {
                    "rows": [
                        {"source": "adr-one", "target": "ref-one", "reason": "scoped"},
                        {"source": "adr-two", "target": "ref-one", "reason": "scoped"},
                    ],
                    "target_records": [{"entity": "ref-one", "citation_handles": [HANDLE]}],
                }
            ),
            encoding="utf-8",
        )
        rows = [
            {
                "source": "adr-one",
                "target": "ref-one",
                "why_required": "The ref governs this ADR boundary.",
                "evidence": HANDLE,
                "action": "review",
                "source_evidence": ["Decision", "ref Goal"],
                "confidence": 0.9,
            },
            {
                "source": "adr-two",
                "target": "ref-one",
                "why_required": "The ref constrains this ADR boundary.",
                "evidence": HANDLE,
                "action": "comply",
                "source_evidence": ["Affected Topology", "ref Goal"],
                "confidence": 0.8,
            },
        ]
        if omit_second:
            rows.pop()
        response = root / "response.json"
        response.write_text(
            json.dumps(
                {
                    "type": "result",
                    "is_error": False,
                    "result": json.dumps(
                        {
                            "review_status": "complete",
                            "rows": rows,
                            "unresolved": [],
                            "unsupported_count": 0,
                            "notes": [],
                        }
                    ),
                    "total_cost_usd": 0.02,
                    "duration_ms": 1200,
                    "num_turns": 3,
                    "permission_denials": [],
                    "usage": {"input_tokens": 10, "output_tokens": 20},
                }
            ),
            encoding="utf-8",
        )
        return subprocess.run(
            [
                sys.executable,
                str(SCRIPT),
                "--private-plan",
                str(plan),
                "--response",
                str(response),
                "--private-output",
                str(root / "validated.json"),
                "--max-cost-usd",
                "0.07",
                "--max-tool-calls",
                "6",
                "--max-output-bytes",
                "524288",
            ],
            cwd=ROOT,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )

    def tearDown(self) -> None:
        if hasattr(self, "tempdir"):
            self.tempdir.cleanup()

    def test_accepts_exact_source_grounded_row_set(self) -> None:
        result = self.run_validator()
        self.assertEqual(result.returncode, 0, result.stderr)
        summary = json.loads(result.stdout)
        self.assertTrue(summary["complete"])
        self.assertEqual(summary["selected_row_count"], 2)
        self.assertEqual(summary["proposed_row_count"], 2)
        self.assertEqual(summary["missing_row_count"], 0)
        self.assertEqual(summary["invalid_evidence_count"], 0)
        self.assertEqual(summary["tool_call_upper_bound"], 2)
        self.assertNotIn("adr-one", result.stdout)
        self.assertNotIn("ref-one", result.stdout)

    def test_rejects_missing_row(self) -> None:
        result = self.run_validator(omit_second=True)
        self.assertEqual(result.returncode, 2)
        summary = json.loads(result.stdout)
        self.assertFalse(summary["complete"])
        self.assertEqual(summary["missing_row_count"], 1)


if __name__ == "__main__":
    unittest.main()
