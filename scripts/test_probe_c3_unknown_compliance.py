import json
import subprocess
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts/probe_c3_unknown_compliance.py"


class ProbeC3UnknownComplianceTests(unittest.TestCase):
    def run_probe(self) -> subprocess.CompletedProcess[str]:
        self.tempdir = tempfile.TemporaryDirectory()
        root = Path(self.tempdir.name)
        manifest = root / "manifest.json"
        manifest.write_text(
            json.dumps(
                {
                    "issues": [
                        {"entity": "adr-private", "family": "unknown_compliance_ref", "message": "Compliance Refs references unknown ref: ref-missing"}
                    ]
                }
            ),
            encoding="utf-8",
        )
        probe = root / "entities.json"
        probe.write_text(
            json.dumps({"rows": [{"entity": "adr-private", "type": "adr", "status": "open", "exit_code": 0}]}),
            encoding="utf-8",
        )
        c3_dir = root / ".c3"
        c3_dir.mkdir()
        binary = root / "binary"
        binary.write_bytes(b"fake")
        wrapper = root / "wrapper.sh"
        wrapper.write_text(
            textwrap.dedent(
                """\
                #!/usr/bin/env bash
                set -euo pipefail
                id="$4"
                if [[ "$id" == "ref-missing" ]]; then exit 7; fi
                printf '%s\n' 'type: adr' 'body: "## Compliance Refs\\n\\n| Ref | Why required | Evidence | Action |\\n|---|---|---|---|\\n| ref-missing | Boundary policy applies. | N.A - create pending. | review |"'
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
                str(probe),
                "--family",
                "unknown_compliance_ref",
                "--c3-dir",
                str(c3_dir),
                "--wrapper",
                str(wrapper),
                "--local-binary",
                str(binary),
                "--local-version",
                "test",
                "--private-output",
                str(root / "result.json"),
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

    def test_classifies_missing_fact_without_create_intent(self) -> None:
        result = self.run_probe()
        self.assertEqual(result.returncode, 0, result.stderr)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["selected_issue_count"], 1)
        self.assertEqual(summary["target_missing_count"], 1)
        self.assertEqual(summary["explicit_create_action_count"], 0)
        self.assertEqual(summary["semantic_judgement_row_count"], 1)
        self.assertEqual(summary["mechanical_repair_candidate_count"], 0)
        self.assertNotIn("adr-private", result.stdout)
        self.assertNotIn("ref-missing", result.stdout)


if __name__ == "__main__":
    unittest.main()
