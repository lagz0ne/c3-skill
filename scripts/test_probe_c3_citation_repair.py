import json
import subprocess
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts/probe_c3_citation_repair.py"


class ProbeC3CitationRepairTests(unittest.TestCase):
    def run_probe(self, *, ambiguous: bool = False, moved_node_with_snippet: bool = False) -> subprocess.CompletedProcess[str]:
        self.tempdir = tempfile.TemporaryDirectory()
        root = Path(self.tempdir.name)
        old_hash = "1" * 64
        new_hash = "2" * 64
        other_hash = "3" * 64
        old_handle = (
            f'private-target#n7@v1:sha256:{old_hash} "current evidence"'
            if moved_node_with_snippet
            else f"private-target#n7@v1:sha256:{old_hash}"
        )
        current_node = 9 if moved_node_with_snippet else 7
        current = f'private-target#n{current_node}@v2:sha256:{new_hash} "current evidence"'
        duplicate = f'private-target#n7@v2:sha256:{other_hash} "other evidence"'
        source_body = f"# ADR\n\n## Evidence\n\n| Claim | Cite |\n|---|---|\n| private | {old_handle} |\n"
        target_body = "# Target\n\n## Goal\n\nCurrent evidence.\n"
        manifest = root / "manifest.json"
        manifest.write_text(
            json.dumps(
                {
                    "issues": [
                        {
                            "entity": "private-source",
                            "family": "stale_citation" if moved_node_with_snippet else "empty_citation_snippet",
                            "message": (
                                "citation to private-target cites version 1, current version is 2"
                                if moved_node_with_snippet
                                else "citation to private-target has empty snippet"
                            ),
                        }
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
                        {"entity": "private-source", "type": "adr", "status": "accepted", "exit_code": 0}
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
        citations = "\\n  - ".join([current, duplicate] if ambiguous else [current])
        wrapper.write_text(
            textwrap.dedent(
                f"""\
                #!/usr/bin/env bash
                set -euo pipefail
                id="$4"
                shift 4
                if [[ "$id" == "private-source" ]]; then
                  printf '%s\\n' 'id: private-source' 'type: adr' 'status: accepted' 'body: {json.dumps(source_body)}'
                elif [[ " $* " == *" --section Goal "* ]]; then
                  printf '%s\\n' 'section: Goal' 'content: Current evidence.' 'citations[{2 if ambiguous else 1}]:' '  - {citations}'
                else
                  printf '%s\\n' 'id: private-target' 'type: component' 'body: {json.dumps(target_body)}'
                fi
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
                "stale_citation" if moved_node_with_snippet else "empty_citation_snippet",
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

    def test_finds_exact_same_node_replacement_and_source_section(self) -> None:
        result = self.run_probe()
        self.assertEqual(result.returncode, 0, result.stderr)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["mutable_row_count"], 1)
        self.assertEqual(summary["safe_replacement_count"], 1)
        self.assertEqual(summary["ambiguous_replacement_count"], 0)
        self.assertEqual(summary["missing_replacement_count"], 0)
        self.assertNotIn("private-source", result.stdout)
        self.assertNotIn("private-target", result.stdout)

    def test_rejects_more_than_one_current_handle_for_the_same_node(self) -> None:
        result = self.run_probe(ambiguous=True)
        self.assertEqual(result.returncode, 2)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["safe_replacement_count"], 0)
        self.assertEqual(summary["ambiguous_replacement_count"], 1)

    def test_uses_an_exact_unique_snippet_when_the_node_id_moved(self) -> None:
        result = self.run_probe(moved_node_with_snippet=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["safe_replacement_count"], 1)
        self.assertEqual(summary["same_node_match_count"], 0)
        self.assertEqual(summary["exact_snippet_match_count"], 1)


if __name__ == "__main__":
    unittest.main()
