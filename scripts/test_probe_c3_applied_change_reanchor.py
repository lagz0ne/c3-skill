import json
import subprocess
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts/probe_c3_applied_change_reanchor.py"


class ProbeC3AppliedChangeReanchorTests(unittest.TestCase):
    def run_probe(
        self,
        *,
        duplicate_current_hash: bool = False,
        omit_carrier: bool = False,
        refresh_pair: bool = False,
        unrelated_invalid_carrier: bool = False,
    ) -> subprocess.CompletedProcess[str]:
        self.tempdir = tempfile.TemporaryDirectory()
        root = Path(self.tempdir.name)
        source = "adr-private-source"
        target = "component-private-target"
        wrong_owner = "component-private-owner"
        old_hash = "1" * 64
        patch_content = "Current applied detail."
        import hashlib

        current_hash = hashlib.sha256(patch_content.encode()).hexdigest()
        old_handle = f"{target}#n7@v1:sha256:{old_hash}"
        source_body = (
            "# ADR\n\n## Affected Topology\n\n"
            "| Entity | Evidence |\n|---|---|\n"
            f"| {target} | {old_handle} |\n"
        )
        target_body = "# Target\n\n## Goal\n\nLead.\n\nCurrent applied detail.\n"
        manifest = root / "manifest.json"
        manifest.write_text(
            json.dumps(
                {
                    "issues": (
                        [
                            {
                                "entity": source,
                                "family": "stale_citation",
                                "message": f"Evidence for Compliance Refs row {target} has a stale cite (no node seals to that hash)",
                            },
                            {
                                "entity": source,
                                "family": "empty_citation_snippet",
                                "message": f"citation to {target} has empty snippet",
                            },
                        ]
                        if refresh_pair
                        else [
                            {
                                "entity": source,
                                "family": "citation_entity_mismatch",
                                "message": f"citation to {target} cites node 7 from {wrong_owner}",
                            },
                            {
                                "entity": source,
                                "family": "citation_entity_mismatch",
                                "message": f"Evidence for Affected Topology row {target} cites node 7 from {wrong_owner}",
                            },
                        ]
                    )
                }
            ),
            encoding="utf-8",
        )
        entity_probe = root / "entities.json"
        entity_probe.write_text(
            json.dumps({"rows": [{"entity": source, "type": "adr", "status": "accepted", "exit_code": 0}]}),
            encoding="utf-8",
        )
        c3_dir = root / "project/.c3"
        carrier_dir = c3_dir / "changes" / source
        carrier_dir.mkdir(parents=True)
        if not omit_carrier:
            (carrier_dir / "0001.patch.md").write_text(
                f"---\ntarget: {target}\nscope: block\nbase: {old_handle}\n---\n{patch_content}\n",
                encoding="utf-8",
            )
        if unrelated_invalid_carrier:
            (carrier_dir / "9999-unrelated.patch.md").write_text(
                "---\ntarget: another-target\nnot-frontmatter\n---\n",
                encoding="utf-8",
            )
        binary = root / "c3x-source"
        binary.write_bytes(b"fake")
        wrapper = root / "c3x.sh"
        citations = [
            f'{target}#n9@v1:sha256:{current_hash} "Current applied detail."',
        ]
        if duplicate_current_hash:
            citations.append(f'{target}#n10@v1:sha256:{current_hash} "Current applied detail."')
        rendered_citations = "','".join(citations)
        wrapper.write_text(
            textwrap.dedent(
                f"""\
                #!/usr/bin/env bash
                set -euo pipefail
                id="$4"
                shift 4
                if [[ "$id" == "{source}" ]]; then
                  printf '%s\n' 'id: {source}' 'type: adr' 'status: accepted' 'body: {json.dumps(source_body)}'
                elif [[ " $* " == *" --section Goal "* ]]; then
                  printf '%s\n' 'section: Goal' 'content: {json.dumps("Lead.\\n\\n" + patch_content)}' 'citations: "{rendered_citations}"'
                else
                  printf '%s\n' 'id: {target}' 'type: component' 'body: {json.dumps(target_body)}'
                fi
                """
            ),
            encoding="utf-8",
        )
        argv = [
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
                "test-source",
                "--private-output",
                str(root / "plan.json"),
            ]
        if refresh_pair:
            argv.extend(["--family", "stale_citation", "--family", "empty_citation_snippet"])
        return subprocess.run(
            argv,
            cwd=ROOT,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )

    def tearDown(self) -> None:
        if hasattr(self, "tempdir"):
            self.tempdir.cleanup()

    def test_builds_one_exact_section_plan_for_two_checker_rows(self) -> None:
        result = self.run_probe()
        self.assertEqual(result.returncode, 0, result.stderr)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["issue_row_count"], 2)
        self.assertEqual(summary["source_group_count"], 1)
        self.assertEqual(summary["safe_reanchor_count"], 1)
        self.assertEqual(summary["reconciled_issue_row_count"], 2)
        self.assertEqual(summary["ambiguous_current_match_count"], 0)
        self.assertNotIn("private-source", result.stdout)
        self.assertNotIn("private-target", result.stdout)

    def test_rejects_duplicate_current_content_hashes(self) -> None:
        result = self.run_probe(duplicate_current_hash=True)
        self.assertEqual(result.returncode, 2)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["safe_reanchor_count"], 0)
        self.assertEqual(summary["ambiguous_current_match_count"], 1)

    def test_rejects_a_missing_change_carrier(self) -> None:
        result = self.run_probe(omit_carrier=True)
        self.assertEqual(result.returncode, 2)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["safe_reanchor_count"], 0)
        self.assertEqual(summary["missing_carrier_count"], 1)

    def test_reanchors_one_carrier_for_a_stale_and_empty_citation_pair(self) -> None:
        result = self.run_probe(refresh_pair=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["issue_row_count"], 2)
        self.assertEqual(summary["source_group_count"], 1)
        self.assertEqual(summary["safe_reanchor_count"], 1)
        self.assertEqual(summary["reconciled_issue_row_count"], 2)

    def test_ignores_an_unrelated_carrier_format(self) -> None:
        result = self.run_probe(unrelated_invalid_carrier=True)
        self.assertEqual(result.returncode, 0, result.stderr)
        summary = json.loads(result.stdout)
        self.assertEqual(summary["carrier_parse_failure_count"], 0)
        self.assertEqual(summary["safe_reanchor_count"], 1)


if __name__ == "__main__":
    unittest.main()
