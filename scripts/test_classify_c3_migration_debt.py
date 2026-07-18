import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts/classify_c3_migration_debt.py"


class ClassifyC3MigrationDebtTests(unittest.TestCase):
    def run_classifier(self, check_text: str, eval_text: str) -> tuple[subprocess.CompletedProcess[str], dict, dict]:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            check_path = root / "check.toon"
            eval_path = root / "eval.toon"
            private_path = root / "private.json"
            check_path.write_text(check_text, encoding="utf-8")
            eval_path.write_text(eval_text, encoding="utf-8")
            result = subprocess.run(
                [
                    sys.executable,
                    str(SCRIPT),
                    "--check",
                    str(check_path),
                    "--eval",
                    str(eval_path),
                    "--private-output",
                    str(private_path),
                ],
                cwd=ROOT,
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
            )
            summary = json.loads(result.stdout) if result.stdout else {}
            private = json.loads(private_path.read_text(encoding="utf-8")) if private_path.exists() else {}
            return result, summary, private

    def test_classifies_every_row_and_keeps_raw_fields_private(self) -> None:
        result, summary, private = self.run_classifier(
            """total: 3
issues[3]:
  -
    severity: warning
    entity: adr-private-one
    message: ADR missing compliance ref ref-private-a
  -
    severity: warning
    entity: adr-private-two
    message: change doc touches nothing: the affected topology is empty or entirely N.A
  -
    severity: warning
    entity: adr-private-three
    message: adr-private-three ready to auto-done: all After cites resolve fresh; run 'c3x check --fix' to actualize accepted->done
""",
            """total: 2
holds: 1
drift: 1
needs_judgement: 0
""",
        )

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(summary["structural_issue_count"], 3)
        self.assertEqual(summary["classified_issue_count"], 3)
        self.assertEqual(summary["unclassified_issue_count"], 0)
        self.assertEqual(summary["semantic_drift_count"], 1)
        self.assertEqual(summary["first_repair_family"], "ready_to_auto_done")
        self.assertNotIn("private", result.stdout)
        self.assertEqual(len(private["issues"]), 3)
        self.assertEqual(private["issues"][0]["entity"], "adr-private-one")

    def test_unclassified_row_fails_closed(self) -> None:
        result, summary, private = self.run_classifier(
            """total: 1
issues[1]:
  -
    severity: warning
    entity: private-entity
    message: a new warning shape that has no classifier
""",
            """total: 1
holds: 1
drift: 0
needs_judgement: 0
""",
        )

        self.assertEqual(result.returncode, 2)
        self.assertEqual(summary["unclassified_issue_count"], 1)
        self.assertEqual(private["issues"][0]["family"], "unclassified")
        self.assertNotIn("private-entity", result.stdout)

    def test_accepts_clean_check_without_an_issues_array(self) -> None:
        result, summary, private = self.run_classifier(
            """total: 4
ok: true
""",
            """total: 2
holds: 2
drift: 0
needs_judgement: 0
""",
        )

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(summary["structural_issue_count"], 0)
        self.assertEqual(summary["parse_mismatch_count"], 0)
        self.assertEqual(private["declared_issue_count"], 0)

    def test_classifies_citation_entity_mismatch_rows(self) -> None:
        result, summary, private = self.run_classifier(
            """total: 2
issues[2]:
  -
    severity: warning
    entity: adr-private
    message: citation to c3-private-a cites node 17 from c3-private-b
  -
    severity: warning
    entity: adr-private
    message: Evidence for Affected Topology row c3-private-a cites node 17 from c3-private-b
""",
            """total: 1
holds: 1
drift: 0
needs_judgement: 0
""",
        )

        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(summary["families"], {"citation_entity_mismatch": 2})
        self.assertEqual([row["family"] for row in private["issues"]], ["citation_entity_mismatch"] * 2)
        self.assertNotIn("private", result.stdout)


if __name__ == "__main__":
    unittest.main()
