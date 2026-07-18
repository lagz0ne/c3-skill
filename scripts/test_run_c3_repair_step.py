import json
import os
import subprocess
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SCRIPT = ROOT / "scripts/run_c3_repair_step.py"
VERIFIER = ROOT / "scripts/verify_c3_repair_step.py"


class RunC3RepairStepTests(unittest.TestCase):
    def make_project(self, root: Path) -> tuple[Path, Path, Path]:
        project = root / "shadow"
        primary = root / "primary"
        for path in (project, primary):
            (path / ".c3/adr").mkdir(parents=True)
            (path / ".c3/adr/example.md").write_text(
                "---\nid: adr-example\ntype: adr\nstatus: accepted\n---\n\n# Example\n",
                encoding="utf-8",
            )
            (path / "product.txt").write_text("unchanged\n", encoding="utf-8")
            subprocess.run(["git", "init", "-q", str(path)], check=True)
            subprocess.run(["git", "-C", str(path), "add", "."], check=True)
            subprocess.run(
                ["git", "-C", str(path), "-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-qm", "base"],
                check=True,
            )

        binary = root / "c3x-source"
        binary.write_text("#!/usr/bin/env bash\nprintf 'test-source\\n'\n", encoding="utf-8")
        binary.chmod(0o755)
        return project, primary, binary

    def make_wrapper(
        self,
        root: Path,
        *,
        change_product: bool = False,
        unexpected_after_family: bool = False,
        delete_identity: bool = False,
    ) -> Path:
        wrapper = root / "c3x.sh"
        product_change = 'printf "changed\\n" > "$project/product.txt"' if change_product else ":"
        identity_change = 'rm "$c3dir/adr/example.md"' if delete_identity else ":"
        after_issues = (
            "printf '%s\\n' 'issues[2]:' '  -' '    severity: warning' '    entity: private-after' "
            "'    message: ADR missing compliance ref ref-private' '  -' '    severity: warning' "
            "'    entity: private-new' '    message: citation to ref-private cites node 1 from private-new'"
            if unexpected_after_family
            else "printf '%s\\n' 'issues[1]:' '  -' '    severity: warning' '    entity: private-after' '    message: ADR missing compliance ref ref-private'"
        )
        wrapper.write_text(
            textwrap.dedent(
                f"""\
                #!/usr/bin/env bash
                set -euo pipefail
                test "$1" = "--c3-dir"
                c3dir="$2"
                project="$(dirname "$c3dir")"
                command="$3"
                shift 3
                state="$c3dir/lifecycle.done"
                if [[ "$command" == "check" ]]; then
                  if [[ " $* " == *" --fix "* ]]; then
                    printf "done\\n" > "$state"
                    {product_change}
                    {identity_change}
                  fi
                  if [[ -f "$state" ]]; then
                    {after_issues}
                  else
                    printf '%s\\n' 'issues[2]:' '  -' '    severity: warning' '    entity: private-before' \"    message: private-before ready to auto-done: all After cites resolve fresh; run 'c3x check --fix' to actualize accepted->done\" '  -' '    severity: warning' '    entity: private-after' '    message: ADR missing compliance ref ref-private'
                  fi
                elif [[ "$command" == "eval" ]]; then
                  printf '%s\\n' 'total: 2' 'holds: 1' 'drift: 1' 'needs_judgement: 0'
                elif [[ "$command" == "repair" ]]; then
                  printf '%s\\n' 'ok: true'
                elif [[ "$command" == "set" ]]; then
                  test "$1" = "private-change"
                  test "$2" = "status"
                  if [[ "$3" == "done" ]]; then
                    printf "done\\n" > "$state"
                  else
                    test "$3" = "accepted"
                  fi
                elif [[ "$command" == "--version" ]]; then
                  printf 'test-source\n'
                else
                  exit 64
                fi
                """
            ),
            encoding="utf-8",
        )
        wrapper.chmod(0o755)
        return wrapper

    def run_step(
        self,
        project: Path,
        primary: Path,
        binary: Path,
        wrapper: Path,
        output: Path,
        *,
        expect_before: int = 2,
        expect_after: int = 1,
        baseline_only: bool = False,
        set_status_done: bool = False,
        exact_family_deltas: bool = False,
        set_status_accepted: bool = False,
        dependency_link: tuple[str, Path] | None = None,
        expected_identity_loss: str | None = None,
    ) -> subprocess.CompletedProcess[str]:
        argv = [
                sys.executable,
                str(SCRIPT),
                "--project",
                str(project),
                "--primary-target",
                str(primary),
                "--output-dir",
                str(output),
                "--wrapper",
                str(wrapper),
                "--local-binary",
                str(binary),
                "--local-version",
                "test-source",
                "--family",
                "ready_to_auto_done",
                "--expect-before",
                str(expect_before),
                "--expect-after",
                str(expect_after),
                "--expected-family-reduction",
                "1",
            ]
        if baseline_only:
            argv.append("--baseline-only")
        if set_status_done:
            argv.extend(["--set-field", "private-change", "status", "done"])
        if exact_family_deltas:
            argv.extend(["--expected-family-delta", "ready_to_auto_done=1"])
        if set_status_accepted:
            argv.extend([
                "--set-field", "private-change", "status", "accepted",
                "--expect-after", str(expect_before),
                "--expected-family-reduction", "0",
                "--require-exact-family-deltas",
            ])
        if dependency_link:
            relative, source = dependency_link
            argv.extend(["--dependency-link", f"{relative}={source}"])
        if expected_identity_loss:
            argv.extend(["--expected-identity-loss", expected_identity_loss])
        return subprocess.run(
            argv,
            cwd=ROOT,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )

    def verify_step(
        self,
        project: Path,
        primary: Path,
        output: Path,
        *,
        binary: Path | None = None,
        wrapper: Path | None = None,
        expected_family_delta: str | None = None,
        dependency_link: tuple[str, Path] | None = None,
        expected_identity_loss: str | None = None,
    ) -> subprocess.CompletedProcess[str]:
        argv = [
            sys.executable,
            str(VERIFIER),
            "--evidence-dir", str(output),
            "--project", str(project),
            "--primary-target", str(primary),
            "--family", "ready_to_auto_done",
            "--expect-before", "2",
            "--expect-after", "1",
            "--expected-family-reduction", "1",
        ]
        if binary:
            argv.extend(["--local-binary", str(binary), "--local-version", "test-source"])
        if wrapper:
            argv.extend(["--wrapper", str(wrapper)])
        if expected_family_delta:
            argv.extend(["--expected-family-delta", expected_family_delta])
        if dependency_link:
            relative, source = dependency_link
            argv.extend(["--dependency-link", f"{relative}={source}"])
        if expected_identity_loss:
            argv.extend(["--expected-identity-loss", expected_identity_loss])
        return subprocess.run(
            argv,
            cwd=ROOT,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )

    def test_accepts_exact_scope_clean_reduction_and_keeps_raw_output_private(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            project, primary, binary = self.make_project(root)
            wrapper = self.make_wrapper(root)
            result = self.run_step(project, primary, binary, wrapper, root / "evidence")

            self.assertEqual(result.returncode, 0, result.stderr)
            summary = json.loads(result.stdout)
            self.assertEqual(summary["decision"], "accepted")
            self.assertEqual(summary["before_issue_count"], 2)
            self.assertEqual(summary["after_issue_count"], 1)
            self.assertEqual(summary["family_reduction"], 1)
            self.assertEqual(summary["non_c3_source_change_count"], 0)
            self.assertEqual(summary["primary_target_change_count"], 0)
            self.assertNotIn("private-before", result.stdout)
            self.assertTrue((root / "evidence/private/before-debt.json").is_file())
            self.assertTrue((root / "evidence/private/after-debt.json").is_file())
            verification = self.verify_step(project, primary, root / "evidence", binary=binary, wrapper=wrapper)
            self.assertEqual(verification.returncode, 0, verification.stderr)
            self.assertEqual(json.loads(verification.stdout)["decision"], "accepted")

    def test_fails_closed_when_the_repair_changes_product_source(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            project, primary, binary = self.make_project(root)
            result = self.run_step(
                project,
                primary,
                binary,
                self.make_wrapper(root, change_product=True),
                root / "evidence",
            )

            self.assertEqual(result.returncode, 2)
            summary = json.loads(result.stdout)
            self.assertEqual(summary["decision"], "blocked")
            self.assertEqual(summary["non_c3_source_change_count"], 1)
            self.assertNotIn("product.txt", result.stdout)

    def test_fails_closed_when_identity_is_deleted(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            project, primary, binary = self.make_project(root)
            result = self.run_step(
                project, primary, binary,
                self.make_wrapper(root, delete_identity=True),
                root / "evidence",
            )
            self.assertEqual(result.returncode, 2)
            summary = json.loads(result.stdout)
            self.assertEqual(summary["decision"], "blocked")
            self.assertEqual(summary["identity_loss_count"], 1)
            self.assertEqual(summary["carrier_change_count"], 0)

    def test_verifier_rejects_wrong_identity_with_same_delta_count(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            project, primary, binary = self.make_project(root)
            wrapper = self.make_wrapper(root, delete_identity=True)
            output = root / "evidence"
            result = self.run_step(
                project, primary, binary, wrapper, output,
                expected_identity_loss="adr-example",
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            accepted = self.verify_step(
                project, primary, output, binary=binary, wrapper=wrapper,
                expected_identity_loss="adr-example",
            )
            self.assertEqual(accepted.returncode, 0, accepted.stdout)
            rejected = self.verify_step(
                project, primary, output, binary=binary, wrapper=wrapper,
                expected_identity_loss="adr-wrong",
            )
            self.assertEqual(rejected.returncode, 2, rejected.stdout)

    def test_accepts_guarded_set_field_mutation(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            project, primary, binary = self.make_project(root)
            wrapper = self.make_wrapper(root)
            result = self.run_step(
                project,
                primary,
                binary,
                wrapper,
                root / "evidence",
                set_status_done=True,
                exact_family_deltas=True,
            )

            self.assertEqual(result.returncode, 0, result.stderr)
            summary = json.loads(result.stdout)
            self.assertEqual(summary["decision"], "accepted")
            self.assertEqual(summary["family_reduction"], 1)
            self.assertEqual(summary["family_deltas"], {"ready_to_auto_done": 1})
            report = json.loads((root / "evidence/private/repair-report.json").read_text())
            self.assertEqual(report["records"][2]["command"], "set")
            self.assertEqual(report["records"][2]["args"], ["private-change", "status", "done"])
            verification = self.verify_step(project, primary, root / "evidence", binary=binary, wrapper=wrapper)
            self.assertEqual(verification.returncode, 0, verification.stderr)
            self.assertEqual(json.loads(verification.stdout)["decision"], "accepted")

    def test_baseline_mismatch_blocks_before_mutation(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            project, primary, binary = self.make_project(root)
            result = self.run_step(
                project,
                primary,
                binary,
                self.make_wrapper(root),
                root / "evidence",
                expect_before=3,
            )

            self.assertEqual(result.returncode, 2)
            summary = json.loads(result.stdout)
            self.assertEqual(summary["decision"], "blocked_admissibility")
            self.assertFalse((project / ".c3/lifecycle.done").exists())

    def test_exact_family_deltas_reject_an_unexpected_family(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            project, primary, binary = self.make_project(root)
            result = self.run_step(
                project,
                primary,
                binary,
                self.make_wrapper(root, unexpected_after_family=True),
                root / "evidence",
                expect_after=2,
                exact_family_deltas=True,
            )

            self.assertEqual(result.returncode, 2)
            summary = json.loads(result.stdout)
            self.assertEqual(summary["decision"], "blocked")
            self.assertEqual(summary["family_deltas"]["ready_to_auto_done"], 1)
            self.assertEqual(summary["family_deltas"]["citation_entity_mismatch"], -1)

    def test_accepts_a_declared_no_delta_transition(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            project, primary, binary = self.make_project(root)
            result = self.run_step(
                project,
                primary,
                binary,
                self.make_wrapper(root),
                root / "evidence",
                set_status_accepted=True,
            )

            self.assertEqual(result.returncode, 0, result.stderr)
            summary = json.loads(result.stdout)
            self.assertEqual(summary["decision"], "accepted")
            self.assertEqual(summary["family_deltas"], {})

    def test_baseline_only_records_snapshot_without_mutation(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            project, primary, binary = self.make_project(root)
            result = self.run_step(
                project,
                primary,
                binary,
                self.make_wrapper(root),
                root / "evidence",
                baseline_only=True,
            )

            self.assertEqual(result.returncode, 0, result.stderr)
            summary = json.loads(result.stdout)
            self.assertEqual(summary["decision"], "snapshot")
            self.assertEqual(summary["before_issue_count"], 2)
            self.assertIsNone(summary["after_issue_count"])
            self.assertFalse((project / ".c3/lifecycle.done").exists())

    def test_verifier_rejects_tampered_raw_receipt(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            project, primary, binary = self.make_project(root)
            output = root / "evidence"
            wrapper = self.make_wrapper(root)
            result = self.run_step(project, primary, binary, wrapper, output)
            self.assertEqual(result.returncode, 0, result.stderr)
            (output / "private/05-after-check.stdout.toon").write_text("issues[0]:\n", encoding="utf-8")

            verification = self.verify_step(project, primary, output, binary=binary, wrapper=wrapper)
            self.assertEqual(verification.returncode, 2)
            self.assertEqual(json.loads(verification.stdout)["decision"], "rejected")

    def test_verifier_checks_runtime_exact_deltas_identity_and_link_cleanup(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            project, primary, binary = self.make_project(root)
            wrapper = self.make_wrapper(root)
            dependency = root / "shared-deps"
            dependency.mkdir()
            output = root / "evidence"
            result = self.run_step(
                project, primary, binary, wrapper, output,
                exact_family_deltas=True,
                dependency_link=("node_modules", dependency),
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertFalse((project / "node_modules").exists())

            verification = self.verify_step(
                project, primary, output,
                binary=binary,
                wrapper=wrapper,
                expected_family_delta="ready_to_auto_done=1",
                dependency_link=("node_modules", dependency),
            )
            self.assertEqual(verification.returncode, 0, verification.stderr)

            binary.write_bytes(b"tampered")
            rejected = self.verify_step(
                project, primary, output,
                binary=binary,
                wrapper=wrapper,
                expected_family_delta="ready_to_auto_done=1",
                dependency_link=("node_modules", dependency),
            )
            self.assertEqual(rejected.returncode, 2)

    def test_verifier_rejects_wrong_exact_family_delta(self) -> None:
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            project, primary, binary = self.make_project(root)
            wrapper = self.make_wrapper(root)
            output = root / "evidence"
            result = self.run_step(project, primary, binary, wrapper, output, exact_family_deltas=True)
            self.assertEqual(result.returncode, 0, result.stderr)
            verification = self.verify_step(
                project, primary, output,
                expected_family_delta="ready_to_auto_done=2",
            )
            self.assertEqual(verification.returncode, 2)


if __name__ == "__main__":
    unittest.main()
