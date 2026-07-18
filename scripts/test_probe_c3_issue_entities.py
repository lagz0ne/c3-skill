import json
import subprocess
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts/probe_c3_issue_entities.py"


class ProbeC3IssueEntitiesTests(unittest.TestCase):
    def run_probe(self, *, fail_second: bool = False) -> subprocess.CompletedProcess[str]:
        self.tempdir = tempfile.TemporaryDirectory()
        root = Path(self.tempdir.name)
        manifest = root / "manifest.json"
        manifest.write_text(
            json.dumps(
                {
                    "issues": [
                        {"entity": "private-one", "family": "empty_citation_snippet"},
                        {"entity": "private-two", "family": "empty_citation_snippet"},
                        {"entity": "private-other", "family": "another_family"},
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
                if [[ "$id" == "private-two" && "{int(fail_second)}" == "1" ]]; then
                  exit 7
                fi
                printf '%s\\n' "id: $id" 'type: adr' 'status: accepted' 'body: private body'
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
                "--family",
                "empty_citation_snippet",
                "--c3-dir",
                str(c3_dir),
                "--wrapper",
                str(wrapper),
                "--local-binary",
                str(binary),
                "--local-version",
                "test-source",
                "--private-output",
                str(root / "probe.json"),
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

    def test_reports_only_generic_type_and_status_counts(self) -> None:
        result = self.run_probe()
        self.assertEqual(result.returncode, 0, result.stderr)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["entity_count"], 2)
        self.assertEqual(summary["type_counts"], {"adr": 2})
        self.assertEqual(summary["status_counts"], {"accepted": 2})
        self.assertEqual(summary["issue_status_counts"], {"accepted": 2})
        self.assertEqual(summary["mutable_issue_count"], 2)
        self.assertEqual(summary["terminal_issue_count"], 0)
        self.assertEqual(summary["frozen_fact_count"], 0)
        self.assertNotIn("private-one", result.stdout)
        self.assertNotIn("private-two", result.stdout)

    def test_read_failure_fails_closed_without_leaking_id(self) -> None:
        result = self.run_probe(fail_second=True)
        self.assertEqual(result.returncode, 2)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["read_failure_count"], 1)
        self.assertNotIn("private-two", result.stdout)


if __name__ == "__main__":
    unittest.main()
