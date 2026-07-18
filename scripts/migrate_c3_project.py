#!/usr/bin/env python3
"""Preview or apply a C3 project migration with the current local checkout.

The default is a shadow-worktree preview. Project-specific logs stay in a private
output directory; the repository only contains this generic runner and its tests.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
import time
from collections import Counter
from dataclasses import asdict, dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Sequence

from c3_dependency_links import dependency_link, with_dependency_links


ROOT = Path(__file__).resolve().parents[1]
DEFAULT_WRAPPER = ROOT / "skills/c3/bin/c3x.sh"
RUNTIME_NAMES = {
    "c3.db",
    "c3.db-shm",
    "c3.db-wal",
    ".c3.import.tmp.db",
    ".c3.import.tmp.db-shm",
    ".c3.import.tmp.db-wal",
}


@dataclass
class CommandRecord:
    command: str
    argv: list[str]
    exit_code: int
    elapsed_ms: int
    stdout_bytes: int
    stderr_bytes: int
    stdout_sha256: str
    stderr_sha256: str


def sha256_bytes(value: bytes) -> str:
    return hashlib.sha256(value).hexdigest()


def sha256_file(path: Path) -> str:
    return sha256_bytes(path.read_bytes())


def is_runtime_path(path: Path) -> bool:
    return path.name in RUNTIME_NAMES or path.name.startswith(".c3-write-") or path.name.endswith(".lock")


def canonical_manifest(c3_dir: Path) -> dict:
    rows: list[dict[str, object]] = []
    for path in sorted(c3_dir.rglob("*")):
        if not path.is_file() or is_runtime_path(path):
            continue
        rel = path.relative_to(c3_dir).as_posix()
        data = path.read_bytes()
        rows.append({"path": rel, "sha256": sha256_bytes(data), "bytes": len(data)})
    canonical = json.dumps(rows, sort_keys=True, separators=(",", ":")).encode()
    return {"file_count": len(rows), "tree_sha256": sha256_bytes(canonical), "files": rows}


def canonical_identity_manifest(c3_dir: Path) -> tuple[dict[str, object], Counter[str]]:
    """Return a private-free identity summary for canonical entity documents."""
    identities: Counter[str] = Counter()
    excluded_roots = {"changes", "canvases", "adr-templates", "_index"}
    for path in sorted(c3_dir.rglob("*.md")):
        rel = path.relative_to(c3_dir)
        if rel.parts and rel.parts[0] in excluded_roots:
            continue
        text = path.read_text(encoding="utf-8", errors="replace")
        if not text.startswith("---"):
            continue
        end = text.find("\n---", 3)
        if end < 0:
            continue
        match = re.search(r"(?m)^id:\s*([^\s#]+)\s*$", text[3:end])
        if not match:
            continue
        identity = match.group(1).strip("\"'")
        if identity:
            identities[identity] += 1
    rows = sorted(identities.items())
    canonical = json.dumps(rows, separators=(",", ":")).encode()
    return {
        "entity_document_count": sum(identities.values()),
        "unique_entity_id_count": len(identities),
        "duplicate_entity_id_count": sum(count - 1 for count in identities.values()),
        "entity_id_multiset_sha256": sha256_bytes(canonical),
    }, identities


def change_carrier_manifest(c3_dir: Path) -> tuple[dict[str, object], dict[str, str]]:
    rows: dict[str, str] = {}
    changes = c3_dir / "changes"
    if changes.is_dir():
        for path in sorted(changes.rglob("*")):
            if path.is_file() and not is_runtime_path(path):
                rows[path.relative_to(changes).as_posix()] = sha256_file(path)
    canonical = json.dumps(sorted(rows.items()), separators=(",", ":")).encode()
    return {
        "change_carrier_file_count": len(rows),
        "change_carrier_tree_sha256": sha256_bytes(canonical),
    }, rows


def copy_canonical_c3(source: Path, destination: Path) -> None:
    if destination.exists():
        shutil.rmtree(destination)

    def ignore(_dir: str, names: list[str]) -> set[str]:
        return {name for name in names if name in RUNTIME_NAMES or name.startswith(".c3-write-") or name.endswith(".lock")}

    shutil.copytree(source, destination, ignore=ignore)


def run_process(argv: Sequence[str], *, cwd: Path, env: dict[str, str], timeout: int) -> subprocess.CompletedProcess[bytes]:
    return subprocess.run(
        list(argv),
        cwd=cwd,
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        timeout=timeout,
        check=False,
    )


def toon_count(text: str, key: str) -> int | None:
    match = re.search(rf"(?m)^{re.escape(key)}(?:\[(\d+)\]|:\s*(\d+))", text)
    if not match:
        return None
    return int(match.group(1) or match.group(2))


class MigrationRunner:
    def __init__(self, args: argparse.Namespace, output_dir: Path):
        self.args = args
        self.target = args.target.resolve()
        self.output_dir = output_dir.resolve()
        self.wrapper = args.wrapper.resolve()
        self.env = os.environ.copy()
        self.env["C3X_MODE"] = "agent"
        self.runtime: dict[str, object] = {"mode": args.runtime}

    def build_runtime(self) -> None:
        if self.args.runtime == "packaged":
            self.runtime.update({"wrapper_sha256": sha256_file(self.wrapper), "binary_sha256": None})
            return

        runtime_dir = self.output_dir / "runtime"
        runtime_dir.mkdir(parents=True, exist_ok=True)
        binary = runtime_dir / "c3x-source"
        version_file = ROOT / "skills/c3/bin/VERSION"
        version = version_file.read_text(encoding="utf-8").strip() if version_file.exists() else "dev"
        argv = [
            "go", "build", "-C", str(ROOT / "cli"), "-buildvcs=false",
            f"-ldflags=-s -w -X main.version={version}-source", "-o", str(binary), ".",
        ]
        result = run_process(argv, cwd=ROOT, env=self.env, timeout=self.args.timeout)
        (runtime_dir / "build.stdout.txt").write_bytes(result.stdout)
        (runtime_dir / "build.stderr.txt").write_bytes(result.stderr)
        if result.returncode != 0:
            raise RuntimeError(f"local C3 source build failed; see {runtime_dir}")
        binary.chmod(binary.stat().st_mode | 0o111)
        self.env["C3X_LOCAL_BINARY"] = str(binary)
        self.env["C3X_LOCAL_VERSION"] = f"{version}-source"
        self.runtime.update({
            "wrapper_sha256": sha256_file(self.wrapper),
            "binary_sha256": sha256_file(binary),
            "version": f"{version}-source",
        })

    def c3(self, c3_dir: Path, command: str, args: list[str], phase_dir: Path, ordinal: int) -> CommandRecord:
        argv = ["bash", str(self.wrapper), "--c3-dir", str(c3_dir), command, *args]
        started = time.monotonic_ns()
        try:
            execute = lambda: run_process(argv, cwd=ROOT, env=self.env, timeout=self.args.timeout)
            result = (
                with_dependency_links(c3_dir.parent, self.args.dependency_link, execute)
                if command == "eval"
                else execute()
            )
            stdout, stderr, exit_code = result.stdout, result.stderr, result.returncode
        except subprocess.TimeoutExpired as exc:
            stdout = exc.stdout or b""
            stderr = (exc.stderr or b"") + f"\ntimeout after {self.args.timeout}s\n".encode()
            exit_code = 124
        elapsed_ms = (time.monotonic_ns() - started) // 1_000_000
        prefix = phase_dir / f"{ordinal:02d}-{command}"
        prefix.with_suffix(".stdout.txt").write_bytes(stdout)
        prefix.with_suffix(".stderr.txt").write_bytes(stderr)
        return CommandRecord(
            command=command,
            argv=["bash", "<local-wrapper>", "--c3-dir", "<target-or-shadow>/.c3", command, *args],
            exit_code=exit_code,
            elapsed_ms=elapsed_ms,
            stdout_bytes=len(stdout),
            stderr_bytes=len(stderr),
            stdout_sha256=sha256_bytes(stdout),
            stderr_sha256=sha256_bytes(stderr),
        )

    def run_phase(self, project: Path, name: str) -> dict:
        phase_dir = self.output_dir / name
        phase_dir.mkdir(parents=True, exist_ok=True)
        c3_dir = project / ".c3"
        before = canonical_manifest(c3_dir)
        before_identity, before_ids = canonical_identity_manifest(c3_dir)
        before_carriers, before_carrier_rows = change_carrier_manifest(c3_dir)
        specs: list[tuple[str, list[str]]] = [
            ("migrate", []),
            ("repair", ["--include-adr"]),
            ("check", ["--include-adr"]),
        ]
        if not self.args.skip_eval:
            specs.append(("eval", []))

        records: list[CommandRecord] = []
        for ordinal, (command, command_args) in enumerate(specs, 1):
            record = self.c3(c3_dir, command, command_args, phase_dir, ordinal)
            records.append(record)
            if command == "migrate" and record.exit_code != 0:
                break

        after = canonical_manifest(c3_dir)
        after_identity, after_ids = canonical_identity_manifest(c3_dir)
        after_carriers, after_carrier_rows = change_carrier_manifest(c3_dir)
        identity_loss_count = sum((before_ids - after_ids).values())
        identity_addition_count = sum((after_ids - before_ids).values())
        carrier_changed_file_count = sum(
            before_carrier_rows.get(path) != after_carrier_rows.get(path)
            for path in set(before_carrier_rows) | set(after_carrier_rows)
        )
        identity_preserved = (
            before_ids == after_ids
            and before_carrier_rows == after_carrier_rows
        )
        def command_text(command: str) -> str:
            record = next((item for item in records if item.command == command), None)
            if not record:
                return ""
            ordinal = records.index(record) + 1
            stdout = (phase_dir / f"{ordinal:02d}-{command}.stdout.txt").read_text(errors="replace")
            stderr = (phase_dir / f"{ordinal:02d}-{command}.stderr.txt").read_text(errors="replace")
            return stdout + stderr

        check_text = command_text("check")
        eval_text = command_text("eval")
        check_issue_count = toon_count(check_text, "issues")
        check_clean = check_issue_count == 0 or (check_issue_count is None and bool(re.search(r"(?m)^ok:\s*true\s*$", check_text)))
        eval_drift_count = toon_count(eval_text, "drift") if not self.args.skip_eval else 0
        eval_judgement_count = toon_count(eval_text, "needs_judgement") if not self.args.skip_eval else 0
        eval_clean = eval_drift_count == 0 and eval_judgement_count == 0
        migrate_ok = bool(records) and records[0].command == "migrate" and records[0].exit_code == 0
        ready = (
            migrate_ok
            and len(records) == len(specs)
            and all(record.exit_code == 0 for record in records)
            and check_clean
            and eval_clean
            and identity_preserved
        )
        return {
            "commands": [record.command for record in records],
            "records": [asdict(record) for record in records],
            "before": {"file_count": before["file_count"], "tree_sha256": before["tree_sha256"]},
            "after": {"file_count": after["file_count"], "tree_sha256": after["tree_sha256"]},
            "before_identity": {**before_identity, **before_carriers},
            "after_identity": {**after_identity, **after_carriers},
            "identity_preserved": identity_preserved,
            "identity_loss_count": identity_loss_count,
            "identity_addition_count": identity_addition_count,
            "carrier_changed_file_count": carrier_changed_file_count,
            "canonical_changed": before["tree_sha256"] != after["tree_sha256"],
            "migrate_ok": migrate_ok,
            "ready": ready,
            "check_issue_count": check_issue_count,
            "eval_drift_count": eval_drift_count,
            "eval_needs_judgement_count": eval_judgement_count,
        }

    def create_shadow(self, shadow: Path) -> tuple[str, bool]:
        git_dir = self.target / ".git"
        if git_dir.exists():
            head = subprocess.check_output(["git", "-C", str(self.target), "rev-parse", "HEAD"], text=True).strip()
            subprocess.run(
                ["git", "-C", str(self.target), "worktree", "add", "--detach", "--quiet", str(shadow), head],
                check=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
            )
            copy_canonical_c3(self.target / ".c3", shadow / ".c3")
            return head, True
        shutil.copytree(
            self.target,
            shadow,
            ignore=shutil.ignore_patterns(".git", "node_modules", ".turbo", "dist", "build"),
        )
        return "non-git", False

    def remove_shadow(self, shadow: Path, worktree: bool) -> None:
        if self.args.keep_shadow:
            return
        if worktree:
            subprocess.run(
                ["git", "-C", str(self.target), "worktree", "remove", "--force", str(shadow)],
                check=False,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
            )
        elif shadow.exists():
            shutil.rmtree(shadow)

    def working_tree_blockers(self) -> list[str]:
        if not (self.target / ".git").exists():
            return []
        output = subprocess.check_output(
            ["git", "-C", str(self.target), "status", "--porcelain=v1", "-z", "--untracked-files=all"]
        )
        blockers: list[str] = []
        for raw in output.split(b"\0"):
            if not raw:
                continue
            path_bytes = raw[3:] if len(raw) >= 3 and raw[2:3] == b" " else raw
            path = path_bytes.decode(errors="replace")
            candidate = Path(path)
            if candidate.parts and candidate.parts[0] == ".c3" and is_runtime_path(candidate):
                continue
            blockers.append(path)
        return sorted(set(blockers))

    def changed_tracked_paths(self) -> list[str]:
        if not (self.target / ".git").exists():
            return []
        out = subprocess.check_output(["git", "-C", str(self.target), "diff", "--name-only"], text=True)
        return sorted(line for line in out.splitlines() if line)

    def execute(self) -> tuple[dict, int]:
        if not self.target.is_dir() or not (self.target / ".c3").is_dir():
            raise ValueError(f"target has no .c3 directory: {self.target}")
        if self.output_dir == self.target or self.output_dir.is_relative_to(self.target):
            raise ValueError("output directory must stay outside the target repository")
        if self.args.allow_red and not self.args.apply:
            raise ValueError("--allow-red is only valid with --apply")
        blockers = self.working_tree_blockers() if self.args.apply else []
        if blockers:
            raise ValueError(f"--apply requires a clean target; blockers: {', '.join(blockers)}")

        self.build_runtime()
        shadow = self.output_dir / "shadow-worktree"
        head, worktree = self.create_shadow(shadow)
        try:
            shadow_result = self.run_phase(shadow, "shadow")
        finally:
            self.remove_shadow(shadow, worktree)

        report: dict[str, object] = {
            "schema_version": 1,
            "recorded_at": datetime.now(timezone.utc).isoformat(),
            "mode": "apply" if self.args.apply else "preview",
            "target": str(self.target),
            "target_head": head,
            "runtime": self.runtime,
            "shadow": shadow_result,
            "target_mutated": False,
            "changed_tracked_paths": [],
        }

        if not shadow_result["migrate_ok"]:
            report["decision"] = "migration_failed"
            return report, 1
        if not shadow_result["identity_preserved"]:
            report["decision"] = "blocked_identity_loss"
            return report, 1
        if not shadow_result["ready"] and not (self.args.apply and self.args.allow_red):
            report["decision"] = "blocked_validation_debt"
            return report, 2
        if not self.args.apply:
            report["decision"] = "ready_to_apply"
            return report, 0

        before = canonical_manifest(self.target / ".c3")
        target_result = self.run_phase(self.target, "apply")
        after = canonical_manifest(self.target / ".c3")
        changed = self.changed_tracked_paths()
        outside = [path for path in changed if not (path == ".c3" or path.startswith(".c3/"))]
        report.update({
            "apply": target_result,
            "target_mutated": before["tree_sha256"] != after["tree_sha256"],
            "changed_tracked_paths": changed,
        })
        if outside:
            report["decision"] = "blocked_out_of_scope_write"
            report["out_of_scope_paths"] = outside
            return report, 1
        if not target_result["identity_preserved"]:
            report["decision"] = "blocked_identity_loss"
            return report, 1
        if target_result["ready"]:
            report["decision"] = "applied_ready"
            return report, 0
        report["decision"] = "applied_with_validation_debt"
        return report, 2


def default_output_dir(target: Path) -> Path:
    data_home = Path(os.environ.get("XDG_DATA_HOME", Path.home() / ".local/share"))
    timestamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    target_key = sha256_bytes(str(target.resolve()).encode())[:12]
    return data_home / "c3-migrations" / f"{timestamp}-{target_key}"


def parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--target", type=Path, required=True, help="project containing .c3/")
    p.add_argument("--wrapper", type=Path, default=DEFAULT_WRAPPER, help="repository-local C3 wrapper")
    p.add_argument("--runtime", choices=("source", "packaged"), default="source", help="source builds this checkout privately; packaged uses the bundled binary")
    p.add_argument("--output-dir", type=Path, help="private report directory outside the target")
    p.add_argument("--apply", action="store_true", help="apply only after the shadow migration is ready")
    p.add_argument("--allow-red", action="store_true", help="with --apply, permit mechanical migration despite recorded check/eval debt")
    p.add_argument("--skip-eval", action="store_true", help="skip semantic conformance evaluation")
    p.add_argument(
        "--dependency-link",
        action="append",
        type=dependency_link,
        default=[],
        metavar="REL=ABS",
        help="temporarily link a dependency into each project only while eval runs; repeatable",
    )
    p.add_argument("--keep-shadow", action="store_true", help="retain the private shadow worktree for diagnosis")
    p.add_argument("--timeout", type=int, default=300, help="seconds allowed per C3 command or source build")
    return p


def main(argv: Sequence[str] | None = None) -> int:
    args = parser().parse_args(argv)
    output_dir = (args.output_dir or default_output_dir(args.target)).expanduser().resolve()
    output_dir.mkdir(parents=True, exist_ok=True)
    try:
        report, code = MigrationRunner(args, output_dir).execute()
    except (OSError, ValueError, RuntimeError, subprocess.SubprocessError) as exc:
        report = {
            "schema_version": 1,
            "recorded_at": datetime.now(timezone.utc).isoformat(),
            "decision": "runner_error",
            "error": str(exc),
            "target_mutated": False,
        }
        code = 1
    (output_dir / "summary.json").write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps({
        "decision": report["decision"],
        "output_dir": str(output_dir),
        "target_mutated": report.get("target_mutated", False),
    }, sort_keys=True))
    return code


if __name__ == "__main__":
    sys.exit(main())
