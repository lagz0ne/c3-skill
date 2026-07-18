import json
import subprocess
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts/apply_c3_section_plan.py"


class ApplyC3SectionPlanTests(unittest.TestCase):
    def test_applies_exact_multi_family_reduction_without_leaking_ids(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            project = root / "shadow"
            primary = root / "primary"
            for path in (project, primary):
                (path / ".c3/adr").mkdir(parents=True)
                (path / ".c3/adr/example.md").write_text("old\n", encoding="utf-8")
                (path / "product.txt").write_text("unchanged\n", encoding="utf-8")
                subprocess.run(["git", "init", "-q", str(path)], check=True)
                subprocess.run(["git", "-C", str(path), "add", "."], check=True)
                subprocess.run(["git", "-C", str(path), "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-qm", "base"], check=True)
            plan = root / "plan.json"
            plan.write_text(
                json.dumps(
                    {
                        "plans": [
                            {"source": "adr-private", "section": "Affected Topology", "old_section": "old", "new_section": "new"}
                        ]
                    }
                ),
                encoding="utf-8",
            )
            binary = root / "binary"
            binary.write_bytes(b"fake")
            dependencies = root / "shared-node-modules"
            dependencies.mkdir()
            wrapper = root / "wrapper.sh"
            wrapper.write_text(
                textwrap.dedent(
                    """\
                    #!/usr/bin/env bash
                    set -euo pipefail
                    c3dir="$2"
                    project="$(dirname "$c3dir")"
                    command="$3"
                    shift 3
                    state="$c3dir/section.done"
                    if [[ "$command" == "write" ]]; then
                      input="$(</dev/stdin)"
                      [[ "$input" == "new" || "$input" == "old" ]]
                      if [[ "$input" == "new" ]]; then printf '%s\n' done > "$state"; else rm -f "$state"; fi
                      printf '%s\n' 'ok: true'
                    elif [[ "$command" == "check" ]]; then
                      if [[ -f "$state" ]]; then
                        printf '%s\n' 'issues[0]:'
                      else
                        printf '%s\n' 'issues[2]:' '  -' '    severity: warning' '    entity: adr-private' '    message: missing required column "Evidence" in table: Affected Topology' '  -' '    severity: warning' '    entity: adr-private' '    message: Affected Topology row for component-private must include Evidence citation'
                      fi
                    elif [[ "$command" == "eval" ]]; then
                      [[ -L "$project/node_modules" ]]
                      printf '%s\n' 'total: 1' 'holds: 1' 'drift: 0' 'needs_judgement: 0'
                    elif [[ "$command" == "repair" ]]; then
                      printf '%s\n' 'ok: true'
                    else
                      exit 64
                    fi
                    """
                ),
                encoding="utf-8",
            )
            result = subprocess.run(
                [
                    sys.executable,
                    str(SCRIPT),
                    "--project",
                    str(project),
                    "--primary-target",
                    str(primary),
                    "--plan",
                    str(plan),
                    "--output-dir",
                    str(root / "evidence"),
                    "--wrapper",
                    str(wrapper),
                    "--local-binary",
                    str(binary),
                    "--local-version",
                    "test",
                    "--expect-before",
                    "2",
                    "--expect-after",
                    "0",
                    "--expected-family-reduction",
                    "missing_required_column=1",
                    "--expected-family-reduction",
                    "missing_topology_evidence=1",
                    "--dependency-link",
                    f"node_modules={dependencies}",
                ],
                cwd=ROOT,
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            summary = json.loads(result.stdout)
            self.assertEqual(summary["decision"], "accepted")
            self.assertEqual(summary["before_issue_count"], 2)
            self.assertEqual(summary["after_issue_count"], 0)
            self.assertEqual(summary["family_reductions"], {"missing_required_column": 1, "missing_topology_evidence": 1})
            self.assertEqual(summary["non_c3_source_change_count"], 0)
            self.assertNotIn("adr-private", result.stdout)
            self.assertFalse((project / "node_modules").exists())


if __name__ == "__main__":
    unittest.main()
