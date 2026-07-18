import json
import subprocess
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts/probe_c3_compliance_repair.py"
HANDLE = "target-ref#n1@v1:sha256:" + "a" * 64 + ' "governance"'


class ProbeC3ComplianceRepairTests(unittest.TestCase):
    def run_probe(self, *, missing_target: bool = False) -> subprocess.CompletedProcess[str]:
        self.tempdir = tempfile.TemporaryDirectory()
        root = Path(self.tempdir.name)
        manifest = root / "manifest.json"
        manifest.write_text(
            json.dumps(
                {
                    "issues": [
                        {"entity": "private-adr-one", "family": "missing_adr_compliance_ref", "message": "ADR missing compliance ref target-ref (scoped to private-scope)"},
                        {"entity": "private-adr-two", "family": "missing_adr_compliance_ref", "message": "ADR missing compliance ref target-ref (scoped to private-scope)"},
                        {"entity": "private-adr-two", "family": "other", "message": "ignored"},
                    ]
                }
            ),
            encoding="utf-8",
        )
        entity_probe = root / "entities.json"
        entity_probe.write_text(
            json.dumps(
                {
                    "rows": [
                        {"entity": "private-adr-one", "type": "adr", "status": "accepted", "exit_code": 0},
                        {"entity": "private-adr-two", "type": "adr", "status": "open", "exit_code": 0},
                    ]
                }
            ),
            encoding="utf-8",
        )
        c3_dir = root / "project/.c3"
        c3_dir.mkdir(parents=True)
        binary = root / "c3x-source"
        binary.write_bytes(b"fake")
        wrapper = root / "c3x.sh"
        wrapper.write_text(
            textwrap.dedent(
                f"""\
                #!/usr/bin/env bash
                set -euo pipefail
                id="$4"
                if [[ "$id" == "target-ref" ]]; then
                  if [[ "{int(missing_target)}" == "1" ]]; then exit 7; fi
                  if [[ "$5" == "--full" ]]; then
                    printf '%s\n' 'type: ref' 'body: "## Goal\\n\\nGovern the boundary.\\n\\n## Decision\\n\\nReview changes."'
                  else
                    printf '%s\n' '{HANDLE}'
                  fi
                  exit 0
                fi
                printf '%s\n' 'type: adr' 'status: accepted' 'body: "## Compliance Refs\\n\\n| Ref | Why required | Evidence | Action |"'
                """
            ),
            encoding="utf-8",
        )
        return subprocess.run(
            [
                sys.executable,
                str(SCRIPT),
                "--private-manifest",
                str(manifest),
                "--entity-probe",
                str(entity_probe),
                "--family",
                "missing_adr_compliance_ref",
                "--c3-dir",
                str(c3_dir),
                "--wrapper",
                str(wrapper),
                "--local-binary",
                str(binary),
                "--local-version",
                "test-source",
                "--private-output",
                str(root / "plan.json"),
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

    def test_classifies_complete_trace_without_leaking_ids(self) -> None:
        result = self.run_probe()
        self.assertEqual(result.returncode, 0, result.stderr)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["mutable_issue_count"], 2)
        self.assertEqual(summary["source_entity_count"], 2)
        self.assertEqual(summary["target_entity_count"], 1)
        self.assertEqual(summary["source_with_compliance_section_count"], 2)
        self.assertEqual(summary["target_with_citation_count"], 1)
        self.assertEqual(summary["semantic_judgement_row_count"], 2)
        self.assertEqual(summary["mechanical_repair_candidate_count"], 0)
        self.assertEqual(summary["wrapper_read_count"], 5)
        self.assertNotIn("private-adr", result.stdout)
        self.assertNotIn("target-ref", result.stdout)

    def test_missing_target_fails_closed(self) -> None:
        result = self.run_probe(missing_target=True)
        self.assertEqual(result.returncode, 2)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["target_read_failure_count"], 1)
        self.assertEqual(summary["target_with_citation_count"], 0)
        self.assertNotIn("target-ref", result.stdout)


if __name__ == "__main__":
    unittest.main()
