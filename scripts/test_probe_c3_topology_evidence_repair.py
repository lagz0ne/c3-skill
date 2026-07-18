import json
import subprocess
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts/probe_c3_topology_evidence_repair.py"
HANDLE = "component-one#n1@v1:sha256:" + "a" * 64 + ' "goal"'


class ProbeC3TopologyEvidenceRepairTests(unittest.TestCase):
    def test_builds_one_exact_section_repair(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            manifest = root / "manifest.json"
            manifest.write_text(
                json.dumps(
                    {
                        "issues": [
                            {"entity": "adr-open", "family": "missing_required_column", "message": 'missing required column "Evidence" in table: Affected Topology'},
                            {"entity": "adr-open", "family": "missing_topology_evidence", "message": "Affected Topology row for component-one must include Evidence citation"},
                        ]
                    }
                ),
                encoding="utf-8",
            )
            entity_probe = root / "entities.json"
            entity_probe.write_text(
                json.dumps({"rows": [{"entity": "adr-open", "type": "adr", "status": "open", "exit_code": 0}]}),
                encoding="utf-8",
            )
            c3_dir = root / ".c3"
            c3_dir.mkdir()
            binary = root / "binary"
            binary.write_bytes(b"fake")
            wrapper = root / "wrapper.sh"
            wrapper.write_text(
                textwrap.dedent(
                    f"""\
                    #!/usr/bin/env bash
                    set -euo pipefail
                    id="$4"
                    if [[ "$id" == "component-one" ]]; then
                      printf '%s\n' '{HANDLE}'
                      exit 0
                    fi
                    printf '%s\n' 'type: adr' 'status: open' 'body: "## Affected Topology\\n\\n| Entity | Type | Why affected | Governance review |\\n|---|---|---|---|\\n| component-one | component | Its API changes. | Review boundary. |\\n| N.A - no second target | N.A - no second target | N.A - no second target | N.A - no second target |"'
                    """
                ),
                encoding="utf-8",
            )
            output = root / "plan.json"
            result = subprocess.run(
                [
                    sys.executable,
                    str(SCRIPT),
                    "--private-manifest",
                    str(manifest),
                    "--entity-probe",
                    str(entity_probe),
                    "--c3-dir",
                    str(c3_dir),
                    "--wrapper",
                    str(wrapper),
                    "--local-binary",
                    str(binary),
                    "--local-version",
                    "test",
                    "--private-output",
                    str(output),
                ],
                cwd=ROOT,
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            summary = json.loads(result.stdout)
            self.assertEqual(summary["safe_section_repair_count"], 1)
            self.assertEqual(summary["expected_issue_reduction"], 2)
            self.assertEqual(summary["source_status_counts"], {"open": 1})
            self.assertNotIn("adr-open", result.stdout)
            self.assertNotIn("component-one", result.stdout)
            private = json.loads(output.read_text(encoding="utf-8"))
            section = private["plans"][0]["new_section"]
            self.assertIn("| Entity | Type | Why affected | Evidence | Governance review |", section)
            self.assertIn(HANDLE, section)
            self.assertIn("N.A - no second target | N.A - no second target | N.A - no second target", section)


if __name__ == "__main__":
    unittest.main()
