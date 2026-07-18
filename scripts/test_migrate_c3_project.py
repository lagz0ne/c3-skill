import json
import os
import subprocess
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts/migrate_c3_project.py"
LOCAL_WRAPPER = ROOT / "skills/c3/bin/c3x.sh"


class MigrateC3ProjectTests(unittest.TestCase):
    def make_target(self, root: Path) -> Path:
        target = root / "target"
        target.mkdir()
        subprocess.run(["git", "init", "-q", str(target)], check=True)
        subprocess.run(["git", "-C", str(target), "config", "user.email", "eval@example.invalid"], check=True)
        subprocess.run(["git", "-C", str(target), "config", "user.name", "Eval"], check=True)
        (target / ".c3").mkdir()
        (target / ".c3/README.md").write_text("before\n", encoding="utf-8")
        (target / "src").mkdir()
        (target / "src/app.ts").write_text("export {}\n", encoding="utf-8")
        subprocess.run(["git", "-C", str(target), "add", "."], check=True)
        subprocess.run(["git", "-C", str(target), "commit", "-qm", "seed"], check=True)
        return target

    def make_wrapper(
        self,
        root: Path,
        check_exit: int = 0,
        check_issues: int = 0,
        eval_drift: int = 0,
        collapse_adrs: bool = False,
        delete_carriers: bool = False,
        require_dependency_link: bool = False,
    ) -> tuple[Path, Path]:
        calls = root / "calls.jsonl"
        wrapper = root / "wrapper.sh"
        migration_mutation = ""
        if collapse_adrs:
            migration_mutation += textwrap.dedent(
                """\
                rm -f "$c3_dir/adr/"*.md
                printf '%s\n' '---' 'id: adr-20260101-first' '---' > "$c3_dir/adr/adr-20260101-.md"
                """
            )
        if delete_carriers:
            migration_mutation += 'rm -rf "$c3_dir/changes/unit-a"\n'
        eval_guard = ""
        if require_dependency_link:
            eval_guard = textwrap.dedent(
                """\
                if [[ ! -L "$c3_dir/../node_modules" ]]; then
                  printf 'missing temporary dependency link\n' >&2
                  exit 8
                fi
                """
            )
        wrapper.write_text(
            textwrap.dedent(
                f"""\
                #!/usr/bin/env bash
                set -u
                printf '{{"argv":"%s"}}\n' "$*" >> "$FAKE_C3_CALLS"
                c3_dir=""
                command=""
                while [[ $# -gt 0 ]]; do
                  case "$1" in
                    --c3-dir) c3_dir="$2"; shift 2 ;;
                    --version|-v|-V) printf 'fake-c3 1.0\n'; exit 0 ;;
                    *) command="$1"; shift; break ;;
                  esac
                done
                case "$command" in
                  migrate) printf 'migrated\n' >> "$c3_dir/README.md"; {migration_mutation} printf 'MIGRATION COMPLETE: 1 entities swept, 0 canvases reconciled\n' ;;
                  repair) printf 'repair ok\n' ;;
                  check) printf 'total: 4\nissues[{check_issues}]:\n'; exit {check_exit} ;;
                  eval) {eval_guard} printf 'total: 4\nholds: {4 - eval_drift}\ndrift: {eval_drift}\nneeds_judgement: 0\n' ;;
                  *) printf 'unexpected command: %s\n' "$command" >&2; exit 9 ;;
                esac
                """
            ),
            encoding="utf-8",
        )
        wrapper.chmod(0o755)
        return wrapper, calls

    def seed_identity_fixture(self, target: Path) -> None:
        adr = target / ".c3/adr"
        changes = target / ".c3/changes/unit-a"
        adr.mkdir(parents=True)
        changes.mkdir(parents=True)
        for slug in ("first", "second"):
            (adr / f"adr-20260101-{slug}.md").write_text(
                f"---\nid: adr-20260101-{slug}\n---\n",
                encoding="utf-8",
            )
        (changes / "01.patch.md").write_text("carrier\n", encoding="utf-8")
        subprocess.run(["git", "-C", str(target), "add", ".c3"], check=True)
        subprocess.run(["git", "-C", str(target), "commit", "-qm", "identity fixture"], check=True)

    def run_script(self, *args: str, env: dict[str, str] | None = None) -> subprocess.CompletedProcess[str]:
        merged = os.environ.copy()
        if env:
            merged.update(env)
        return subprocess.run(
            [sys.executable, str(SCRIPT), *args],
            cwd=ROOT,
            env=merged,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )

    def test_preview_migrates_shadow_without_touching_target(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            target = self.make_target(root)
            wrapper, calls = self.make_wrapper(root)
            output = root / "output"

            result = self.run_script(
                "--target", str(target),
                "--wrapper", str(wrapper),
                "--runtime", "packaged",
                "--output-dir", str(output),
                env={"FAKE_C3_CALLS": str(calls)},
            )

            report = json.loads((output / "summary.json").read_text(encoding="utf-8"))
            call_rows = [json.loads(line)["argv"] for line in calls.read_text().splitlines()]
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual((target / ".c3/README.md").read_text(), "before\n")
            self.assertEqual(report["decision"], "ready_to_apply")
            self.assertFalse(report["target_mutated"])
            self.assertEqual(report["shadow"]["commands"], ["migrate", "repair", "check", "eval"])
            self.assertTrue(all(str(target / ".c3") not in row for row in call_rows))

    def test_eval_uses_temporary_dependency_link_and_cleans_it(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            target = self.make_target(root)
            dependency_source = root / "dependency-source"
            dependency_source.mkdir()
            wrapper, calls = self.make_wrapper(root, require_dependency_link=True)
            output = root / "output"

            result = self.run_script(
                "--target", str(target),
                "--wrapper", str(wrapper),
                "--runtime", "packaged",
                "--output-dir", str(output),
                "--keep-shadow",
                "--dependency-link", f"node_modules={dependency_source}",
                env={"FAKE_C3_CALLS": str(calls)},
            )

            report = json.loads((output / "summary.json").read_text(encoding="utf-8"))
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(report["decision"], "ready_to_apply")
            self.assertFalse((output / "shadow-worktree/node_modules").exists())

    def test_red_shadow_blocks_apply_and_preserves_target(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            target = self.make_target(root)
            wrapper, calls = self.make_wrapper(root, check_exit=1, check_issues=3)
            output = root / "output"

            result = self.run_script(
                "--target", str(target),
                "--wrapper", str(wrapper),
                "--runtime", "packaged",
                "--output-dir", str(output),
                "--apply",
                env={"FAKE_C3_CALLS": str(calls)},
            )

            report = json.loads((output / "summary.json").read_text(encoding="utf-8"))
            self.assertEqual(result.returncode, 2)
            self.assertEqual((target / ".c3/README.md").read_text(), "before\n")
            self.assertEqual(report["decision"], "blocked_validation_debt")
            self.assertFalse(report["target_mutated"])
            self.assertEqual(report["shadow"]["check_issue_count"], 3)

    def test_zero_exit_check_issues_still_block_apply(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            target = self.make_target(root)
            wrapper, calls = self.make_wrapper(root, check_issues=2)
            output = root / "output"

            result = self.run_script(
                "--target", str(target),
                "--wrapper", str(wrapper),
                "--runtime", "packaged",
                "--output-dir", str(output),
                "--apply",
                env={"FAKE_C3_CALLS": str(calls)},
            )

            report = json.loads((output / "summary.json").read_text(encoding="utf-8"))
            self.assertEqual(result.returncode, 2)
            self.assertEqual(report["decision"], "blocked_validation_debt")
            self.assertEqual(report["shadow"]["check_issue_count"], 2)
            self.assertFalse(report["target_mutated"])

    def test_zero_exit_eval_drift_still_blocks_apply(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            target = self.make_target(root)
            wrapper, calls = self.make_wrapper(root, eval_drift=1)
            output = root / "output"

            result = self.run_script(
                "--target", str(target),
                "--wrapper", str(wrapper),
                "--runtime", "packaged",
                "--output-dir", str(output),
                "--apply",
                env={"FAKE_C3_CALLS": str(calls)},
            )

            report = json.loads((output / "summary.json").read_text(encoding="utf-8"))
            self.assertEqual(result.returncode, 2)
            self.assertEqual(report["decision"], "blocked_validation_debt")
            self.assertEqual(report["shadow"]["eval_drift_count"], 1)
            self.assertFalse(report["target_mutated"])

    def test_apply_runs_only_after_green_shadow_and_limits_changes_to_c3(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            target = self.make_target(root)
            wrapper, calls = self.make_wrapper(root)
            output = root / "output"

            result = self.run_script(
                "--target", str(target),
                "--wrapper", str(wrapper),
                "--runtime", "packaged",
                "--output-dir", str(output),
                "--apply",
                env={"FAKE_C3_CALLS": str(calls)},
            )

            report = json.loads((output / "summary.json").read_text(encoding="utf-8"))
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual((target / ".c3/README.md").read_text(), "before\nmigrated\n")
            self.assertEqual(report["decision"], "applied_ready")
            self.assertTrue(report["target_mutated"])
            self.assertEqual(report["changed_tracked_paths"], [".c3/README.md"])

    def test_apply_rejects_untracked_user_files_but_not_runtime_cache(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            target = self.make_target(root)
            wrapper, calls = self.make_wrapper(root)
            output = root / "output"
            (target / ".c3/c3.db").write_text("disposable\n", encoding="utf-8")
            (target / "notes.txt").write_text("user work\n", encoding="utf-8")

            result = self.run_script(
                "--target", str(target),
                "--wrapper", str(wrapper),
                "--runtime", "packaged",
                "--output-dir", str(output),
                "--apply",
                env={"FAKE_C3_CALLS": str(calls)},
            )

            report = json.loads((output / "summary.json").read_text(encoding="utf-8"))
            self.assertEqual(result.returncode, 1)
            self.assertEqual(report["decision"], "runner_error")
            self.assertIn("notes.txt", report["error"])
            self.assertNotIn("c3.db", report["error"])
            self.assertEqual((target / ".c3/README.md").read_text(), "before\n")

    def test_identity_loss_blocks_even_allow_red_apply(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            target = self.make_target(root)
            self.seed_identity_fixture(target)
            wrapper, calls = self.make_wrapper(root, collapse_adrs=True)
            output = root / "output"

            result = self.run_script(
                "--target", str(target),
                "--wrapper", str(wrapper),
                "--runtime", "packaged",
                "--output-dir", str(output),
                "--apply",
                "--allow-red",
                env={"FAKE_C3_CALLS": str(calls)},
            )

            report = json.loads((output / "summary.json").read_text(encoding="utf-8"))
            self.assertEqual(result.returncode, 1)
            self.assertEqual(report["decision"], "blocked_identity_loss")
            self.assertEqual(report["shadow"]["identity_loss_count"], 1)
            self.assertFalse(report["shadow"]["identity_preserved"])
            self.assertTrue((target / ".c3/adr/adr-20260101-first.md").exists())
            self.assertTrue((target / ".c3/adr/adr-20260101-second.md").exists())

    def test_change_carrier_loss_blocks_preview(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            target = self.make_target(root)
            self.seed_identity_fixture(target)
            wrapper, calls = self.make_wrapper(root, delete_carriers=True)
            output = root / "output"

            result = self.run_script(
                "--target", str(target),
                "--wrapper", str(wrapper),
                "--runtime", "packaged",
                "--output-dir", str(output),
                env={"FAKE_C3_CALLS": str(calls)},
            )

            report = json.loads((output / "summary.json").read_text(encoding="utf-8"))
            self.assertEqual(result.returncode, 1)
            self.assertEqual(report["decision"], "blocked_identity_loss")
            self.assertEqual(report["shadow"]["carrier_changed_file_count"], 1)
            self.assertFalse(report["shadow"]["identity_preserved"])

    def test_local_wrapper_honors_fresh_source_binary_override(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            binary = Path(td) / "c3x-source"
            binary.write_text("#!/usr/bin/env bash\nprintf 'source-runtime:%s\\n' \"$*\"\n", encoding="utf-8")
            binary.chmod(0o755)
            env = os.environ.copy()
            env["C3X_LOCAL_BINARY"] = str(binary)
            env["C3X_MODE"] = "agent"

            result = subprocess.run(
                ["bash", str(LOCAL_WRAPPER), "check"],
                cwd=ROOT,
                env=env,
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
            )

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(result.stdout, "source-runtime:check\n")


if __name__ == "__main__":
    unittest.main()
