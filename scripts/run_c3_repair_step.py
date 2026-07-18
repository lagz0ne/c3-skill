#!/usr/bin/env python3
"""Run one bounded C3 repair step and emit only generic evidence publicly.

Raw C3 output and project-specific fields stay under the caller-provided private
output directory. The command fails closed unless the requested debt family
shrinks by the exact amount and no path outside `.c3` changes.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import subprocess
import sys
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Sequence

from migrate_c3_project import canonical_identity_manifest, canonical_manifest, change_carrier_manifest
from c3_dependency_links import dependency_link, with_dependency_links


ROOT = Path(__file__).resolve().parents[1]
CLASSIFIER = ROOT / "scripts/classify_c3_migration_debt.py"


def sha256_bytes(value: bytes) -> str:
    return hashlib.sha256(value).hexdigest()


def sha256_file(path: Path) -> str:
    return sha256_bytes(path.read_bytes())


def family_delta(value: str) -> tuple[str, int]:
    family, separator, raw_delta = value.partition("=")
    if not separator or not family:
        raise argparse.ArgumentTypeError("expected FAMILY=REDUCTION")
    try:
        reduction = int(raw_delta)
    except ValueError as exc:
        raise argparse.ArgumentTypeError("family reduction must be an integer") from exc
    if reduction == 0:
        raise argparse.ArgumentTypeError("family reduction must be non-zero")
    return family, reduction


def git_changed_paths(project: Path) -> list[str]:
    if not (project / ".git").exists():
        return []
    result = subprocess.run(
        ["git", "-C", str(project), "status", "--porcelain=v1", "-z", "--untracked-files=all"],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )
    if result.returncode != 0:
        raise RuntimeError("git status failed for repair scope")
    paths: list[str] = []
    for raw in result.stdout.split(b"\0"):
        if not raw:
            continue
        value = raw[3:] if len(raw) >= 3 and raw[2:3] == b" " else raw
        paths.append(value.decode(errors="replace"))
    return sorted(set(paths))


def non_c3_paths(project: Path) -> list[str]:
    return [path for path in git_changed_paths(project) if not (Path(path).parts and Path(path).parts[0] == ".c3")]


def command_record(
    wrapper: Path,
    c3_dir: Path,
    command: str,
    command_args: list[str],
    output_prefix: Path,
    env: dict[str, str],
    timeout: int,
) -> dict[str, object]:
    started = time.monotonic_ns()
    try:
        result = subprocess.run(
            ["bash", str(wrapper), "--c3-dir", str(c3_dir), command, *command_args],
            cwd=ROOT,
            env=env,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            timeout=timeout,
            check=False,
        )
        stdout, stderr, exit_code = result.stdout, result.stderr, result.returncode
    except subprocess.TimeoutExpired as exc:
        stdout = exc.stdout or b""
        stderr = (exc.stderr or b"") + f"\ntimeout after {timeout}s\n".encode()
        exit_code = 124
    elapsed_ms = (time.monotonic_ns() - started) // 1_000_000
    output_prefix.parent.mkdir(parents=True, exist_ok=True)
    output_prefix.with_suffix(".stdout.toon").write_bytes(stdout)
    output_prefix.with_suffix(".stderr.txt").write_bytes(stderr)
    return {
        "command": command,
        "args": command_args,
        "exit_code": exit_code,
        "elapsed_ms": elapsed_ms,
        "stdout_bytes": len(stdout),
        "stderr_bytes": len(stderr),
        "stdout_sha256": sha256_bytes(stdout),
        "stderr_sha256": sha256_bytes(stderr),
    }


def classify(check_path: Path, eval_path: Path, private_output: Path) -> tuple[dict[str, object], dict[str, object]]:
    result = subprocess.run(
        [
            sys.executable,
            str(CLASSIFIER),
            "--check",
            str(check_path),
            "--eval",
            str(eval_path),
            "--private-output",
            str(private_output),
        ],
        cwd=ROOT,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )
    summary = json.loads(result.stdout) if result.stdout else {}
    record = {
        "exit_code": result.returncode,
        "stdout_sha256": sha256_bytes(result.stdout),
        "stderr_sha256": sha256_bytes(result.stderr),
    }
    return summary, record


def identity_snapshot(c3_dir: Path) -> dict[str, object]:
    identity, ids = canonical_identity_manifest(c3_dir)
    carriers, carrier_rows = change_carrier_manifest(c3_dir)
    return {
        "summary": {**identity, **carriers},
        "entity_ids": [list(item) for item in sorted(ids.items())],
        "change_carriers": [list(item) for item in sorted(carrier_rows.items())],
    }


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description=__doc__)
    result.add_argument("--project", type=Path, required=True, help="isolated repair worktree")
    result.add_argument("--primary-target", type=Path, required=True, help="primary worktree that must stay clean")
    result.add_argument("--output-dir", type=Path, required=True, help="private evidence directory outside either project")
    result.add_argument("--wrapper", type=Path, required=True, help="explicit local C3 wrapper")
    result.add_argument("--local-binary", type=Path, required=True, help="pinned local source binary")
    result.add_argument("--local-version", required=True)
    result.add_argument("--family", required=True)
    result.add_argument("--expect-before", type=int, required=True)
    result.add_argument("--expect-after", type=int, required=True)
    result.add_argument("--expected-family-reduction", type=int, default=1)
    result.add_argument("--expected-identity-loss", action="append", default=[], metavar="ID")
    result.add_argument("--expected-identity-addition", action="append", default=[], metavar="ID")
    result.add_argument("--expected-carrier-change", action="append", default=[], metavar="RELATIVE_PATH")
    result.add_argument(
        "--expected-family-delta",
        type=family_delta,
        action="append",
        default=[],
        metavar="FAMILY=REDUCTION",
        help="declare every non-zero before-to-after family reduction; repeatable",
    )
    result.add_argument(
        "--require-exact-family-deltas",
        action="store_true",
        help="require the complete non-zero family-delta map to match declarations, including an empty map",
    )
    result.add_argument("--baseline-only", action="store_true", help="record the fresh baseline and stop before mutation")
    result.add_argument(
        "--set-field",
        nargs=3,
        metavar=("ID", "FIELD", "VALUE"),
        help="use one guarded c3 set mutation instead of check --fix",
    )
    result.add_argument("--timeout", type=int, default=300)
    result.add_argument(
        "--dependency-link",
        type=dependency_link,
        action="append",
        default=[],
        metavar="REL=ABS",
        help="temporary project-relative dependency link used only while eval runs",
    )
    return result


def main(argv: Sequence[str] | None = None) -> int:
    args = parser().parse_args(argv)
    expected_family_deltas = dict(args.expected_family_delta)
    if len(expected_family_deltas) != len(args.expected_family_delta):
        raise ValueError("duplicate --expected-family-delta family")
    project = args.project.resolve()
    primary = args.primary_target.resolve()
    output_dir = args.output_dir.resolve()
    wrapper = args.wrapper.resolve()
    binary = args.local_binary.resolve()
    for path, label in ((project, "project"), (primary, "primary target")):
        if not path.is_dir() or not (path / ".c3").is_dir():
            raise ValueError(f"{label} must contain .c3")
        if output_dir == path or output_dir.is_relative_to(path):
            raise ValueError("private output directory must stay outside both worktrees")
    if not wrapper.is_file() or not binary.is_file():
        raise ValueError("local wrapper and binary must exist")
    if non_c3_paths(project):
        raise ValueError("repair worktree already has changes outside .c3")
    if git_changed_paths(primary):
        raise ValueError("primary target must be clean before repair")

    output_dir.mkdir(parents=True, exist_ok=True)
    private = output_dir / "private"
    private.mkdir(parents=True, exist_ok=True)
    env = os.environ.copy()
    env.update(
        {
            "C3X_MODE": "agent",
            "C3X_LOCAL_BINARY": str(binary),
            "C3X_LOCAL_VERSION": args.local_version,
        }
    )
    records: list[dict[str, object]] = []
    dependency_contract = [
        {"relative": relative.as_posix(), "source": str(source)}
        for relative, source in args.dependency_link
    ]

    def run(label: str, command: str, command_args: list[str]) -> dict[str, object]:
        execute = lambda: command_record(
            wrapper,
            project / ".c3",
            command,
            command_args,
            private / label,
            env,
            args.timeout,
        )
        record = with_dependency_links(project, args.dependency_link, execute) if command == "eval" else execute()
        records.append(record)
        return record

    before_tree = canonical_manifest(project / ".c3")
    before_identity = identity_snapshot(project / ".c3")
    before_identity_path = private / "before-identity.json"
    before_identity_path.write_text(json.dumps(before_identity, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    before_check = run("01-before-check", "check", ["--include-adr"])
    before_eval = run("02-before-eval", "eval", [])
    before_summary, before_classifier = classify(
        private / "01-before-check.stdout.toon",
        private / "02-before-eval.stdout.toon",
        private / "before-debt.json",
    )

    before_family = int(before_summary.get("families", {}).get(args.family, 0))
    pre_non_c3 = non_c3_paths(project)
    pre_primary_changes = git_changed_paths(primary)
    baseline_admissible = (
        before_check["exit_code"] == 0
        and before_eval["exit_code"] == 0
        and before_classifier["exit_code"] == 0
        and before_summary.get("structural_issue_count") == args.expect_before
        and before_family >= args.expected_family_reduction
        and before_summary.get("unclassified_issue_count") == 0
        and before_summary.get("parse_mismatch_count") == 0
        and isinstance(before_summary.get("semantic_drift_count"), int)
        and isinstance(before_summary.get("semantic_needs_judgement_count"), int)
        and not pre_non_c3
        and not pre_primary_changes
    )
    if not baseline_admissible:
        summary = {
            "schema_version": 1,
            "decision": "blocked_admissibility",
            "before_issue_count": before_summary.get("structural_issue_count"),
            "after_issue_count": None,
            "family_reduction": 0,
            "family_deltas": None,
            "before_family_count": before_family,
            "after_family_count": None,
            "before_semantic_drift_count": before_summary.get("semantic_drift_count"),
            "after_semantic_drift_count": None,
            "before_semantic_needs_judgement_count": before_summary.get("semantic_needs_judgement_count"),
            "after_semantic_needs_judgement_count": None,
            "before_unclassified_issue_count": before_summary.get("unclassified_issue_count"),
            "after_unclassified_issue_count": None,
            "non_c3_source_change_count": len(pre_non_c3),
            "primary_target_change_count": len(pre_primary_changes),
            "canonical_c3_changed": False,
            "max_command_elapsed_seconds": round(max(int(record["elapsed_ms"]) for record in records) / 1000, 3),
            "command_failure_count": sum(record["exit_code"] != 0 for record in records),
            "classifier_failure_count": int(before_classifier["exit_code"] != 0),
            "runtime": {
                "version": args.local_version,
                "binary_sha256": sha256_file(binary),
                "wrapper_sha256": sha256_file(wrapper),
            },
        }
        report = {
            "recorded_at": datetime.now(timezone.utc).isoformat(),
            "summary": summary,
            "before_c3_tree_sha256": before_tree["tree_sha256"],
            "after_c3_tree_sha256": before_tree["tree_sha256"],
            "records": records,
            "before_classifier": before_classifier,
        }
        (private / "repair-report.json").write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
        (output_dir / "summary.json").write_text(json.dumps(summary, indent=2, sort_keys=True) + "\n", encoding="utf-8")
        print(json.dumps(summary, sort_keys=True))
        return 2

    if args.baseline_only:
        summary = {
            "schema_version": 1,
            "decision": "snapshot",
            "before_issue_count": before_summary.get("structural_issue_count"),
            "after_issue_count": None,
            "family_reduction": 0,
            "family_deltas": None,
            "before_family_count": before_family,
            "after_family_count": None,
            "before_semantic_drift_count": before_summary.get("semantic_drift_count"),
            "after_semantic_drift_count": None,
            "before_semantic_needs_judgement_count": before_summary.get("semantic_needs_judgement_count"),
            "after_semantic_needs_judgement_count": None,
            "before_unclassified_issue_count": before_summary.get("unclassified_issue_count"),
            "after_unclassified_issue_count": None,
            "non_c3_source_change_count": len(pre_non_c3),
            "primary_target_change_count": len(pre_primary_changes),
            "canonical_c3_changed": False,
            "max_command_elapsed_seconds": round(max(int(record["elapsed_ms"]) for record in records) / 1000, 3),
            "command_failure_count": 0,
            "classifier_failure_count": 0,
            "runtime": {
                "version": args.local_version,
                "binary_sha256": sha256_file(binary),
                "wrapper_sha256": sha256_file(wrapper),
            },
        }
        report = {
            "recorded_at": datetime.now(timezone.utc).isoformat(),
            "summary": summary,
            "before_c3_tree_sha256": before_tree["tree_sha256"],
            "after_c3_tree_sha256": before_tree["tree_sha256"],
            "records": records,
            "before_classifier": before_classifier,
        }
        (private / "repair-report.json").write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
        (output_dir / "summary.json").write_text(json.dumps(summary, indent=2, sort_keys=True) + "\n", encoding="utf-8")
        print(json.dumps(summary, sort_keys=True))
        return 0

    if args.set_field:
        mutation = run("03-set-field", "set", list(args.set_field))
    else:
        mutation = run("03-check-fix", "check", ["--include-adr", "--fix"])
    repair = run("04-repair", "repair", ["--include-adr"])
    after_check = run("05-after-check", "check", ["--include-adr"])
    after_eval = run("06-after-eval", "eval", [])
    after_summary, after_classifier = classify(
        private / "05-after-check.stdout.toon",
        private / "06-after-eval.stdout.toon",
        private / "after-debt.json",
    )
    after_tree = canonical_manifest(project / ".c3")
    after_identity = identity_snapshot(project / ".c3")
    after_identity_path = private / "after-identity.json"
    after_identity_path.write_text(json.dumps(after_identity, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    before_ids = dict(before_identity["entity_ids"])
    after_ids = dict(after_identity["entity_ids"])
    identity_losses = sorted(
        entity_id
        for entity_id, count in before_ids.items()
        for _ in range(max(count - after_ids.get(entity_id, 0), 0))
    )
    identity_additions = sorted(
        entity_id
        for entity_id, count in after_ids.items()
        for _ in range(max(count - before_ids.get(entity_id, 0), 0))
    )
    before_carrier_rows = dict(before_identity["change_carriers"])
    after_carrier_rows = dict(after_identity["change_carriers"])
    carrier_changes = sorted(
        path for path in set(before_carrier_rows) | set(after_carrier_rows)
        if before_carrier_rows.get(path) != after_carrier_rows.get(path)
    )

    after_family = int(after_summary.get("families", {}).get(args.family, 0))
    family_reduction = before_family - after_family
    before_families = before_summary.get("families", {})
    after_families = after_summary.get("families", {})
    actual_family_deltas = {
        family: int(before_families.get(family, 0)) - int(after_families.get(family, 0))
        for family in sorted(set(before_families) | set(after_families))
        if int(before_families.get(family, 0)) != int(after_families.get(family, 0))
    }
    exact_family_deltas_ok = (
        not expected_family_deltas and not args.require_exact_family_deltas
    ) or actual_family_deltas == expected_family_deltas
    repair_non_c3 = non_c3_paths(project)
    primary_changes = git_changed_paths(primary)
    before_drift = before_summary.get("semantic_drift_count")
    after_drift = after_summary.get("semantic_drift_count")
    before_judgement = before_summary.get("semantic_needs_judgement_count")
    after_judgement = after_summary.get("semantic_needs_judgement_count")
    command_exit_ok = all(record["exit_code"] == 0 for record in records)
    classifiers_ok = before_classifier["exit_code"] == 0 and after_classifier["exit_code"] == 0
    counts_ok = (
        before_summary.get("structural_issue_count") == args.expect_before
        and after_summary.get("structural_issue_count") == args.expect_after
        and family_reduction == args.expected_family_reduction
        and before_summary.get("unclassified_issue_count") == 0
        and after_summary.get("unclassified_issue_count") == 0
        and before_summary.get("parse_mismatch_count") == 0
        and after_summary.get("parse_mismatch_count") == 0
        and exact_family_deltas_ok
    )
    semantic_ok = (
        isinstance(before_drift, int)
        and isinstance(after_drift, int)
        and after_drift <= before_drift
        and isinstance(before_judgement, int)
        and isinstance(after_judgement, int)
        and after_judgement <= before_judgement
    )
    scope_ok = not repair_non_c3 and not primary_changes
    dependency_links_cleaned = all(not os.path.lexists(project / item["relative"]) for item in dependency_contract)
    identity_ok = (
        identity_losses == sorted(args.expected_identity_loss)
        and identity_additions == sorted(args.expected_identity_addition)
        and carrier_changes == sorted(args.expected_carrier_change)
    )
    accepted = command_exit_ok and classifiers_ok and counts_ok and semantic_ok and scope_ok and dependency_links_cleaned and identity_ok

    summary = {
        "schema_version": 1,
        "decision": "accepted" if accepted else "blocked",
        "before_issue_count": before_summary.get("structural_issue_count"),
        "after_issue_count": after_summary.get("structural_issue_count"),
        "family_reduction": family_reduction,
        "family_deltas": actual_family_deltas,
        "before_family_count": before_family,
        "after_family_count": after_family,
        "before_semantic_drift_count": before_drift,
        "after_semantic_drift_count": after_drift,
        "before_semantic_needs_judgement_count": before_judgement,
        "after_semantic_needs_judgement_count": after_judgement,
        "before_unclassified_issue_count": before_summary.get("unclassified_issue_count"),
        "after_unclassified_issue_count": after_summary.get("unclassified_issue_count"),
        "non_c3_source_change_count": len(repair_non_c3),
        "primary_target_change_count": len(primary_changes),
        "dependency_link_count": len(dependency_contract),
        "dependency_links_cleaned": dependency_links_cleaned,
        "identity_loss_count": len(identity_losses),
        "identity_addition_count": len(identity_additions),
        "carrier_change_count": len(carrier_changes),
        "identity_losses_sha256": sha256_bytes(json.dumps(identity_losses, separators=(",", ":")).encode()),
        "identity_additions_sha256": sha256_bytes(json.dumps(identity_additions, separators=(",", ":")).encode()),
        "carrier_changes_sha256": sha256_bytes(json.dumps(carrier_changes, separators=(",", ":")).encode()),
        "expected_identity_losses_sha256": sha256_bytes(json.dumps(sorted(args.expected_identity_loss), separators=(",", ":")).encode()),
        "expected_identity_additions_sha256": sha256_bytes(json.dumps(sorted(args.expected_identity_addition), separators=(",", ":")).encode()),
        "expected_carrier_changes_sha256": sha256_bytes(json.dumps(sorted(args.expected_carrier_change), separators=(",", ":")).encode()),
        "before_identity_sha256": sha256_file(before_identity_path),
        "after_identity_sha256": sha256_file(after_identity_path),
        "canonical_c3_changed": before_tree["tree_sha256"] != after_tree["tree_sha256"],
        "max_command_elapsed_seconds": round(max(int(record["elapsed_ms"]) for record in records) / 1000, 3),
        "command_failure_count": sum(record["exit_code"] != 0 for record in records),
        "classifier_failure_count": int(before_classifier["exit_code"] != 0) + int(after_classifier["exit_code"] != 0),
        "runtime": {
            "version": args.local_version,
            "binary_sha256": sha256_file(binary),
            "wrapper_sha256": sha256_file(wrapper),
        },
    }
    report = {
        "recorded_at": datetime.now(timezone.utc).isoformat(),
        "summary": summary,
        "before_c3_tree_sha256": before_tree["tree_sha256"],
        "after_c3_tree_sha256": after_tree["tree_sha256"],
        "before_identity_sha256": sha256_file(before_identity_path),
        "after_identity_sha256": sha256_file(after_identity_path),
        "dependency_links": dependency_contract,
        "expected_family_deltas": expected_family_deltas,
        "require_exact_family_deltas": args.require_exact_family_deltas,
        "expected_identity_losses": sorted(args.expected_identity_loss),
        "expected_identity_additions": sorted(args.expected_identity_addition),
        "expected_carrier_changes": sorted(args.expected_carrier_change),
        "records": records,
        "before_classifier": before_classifier,
        "after_classifier": after_classifier,
    }
    (private / "repair-report.json").write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    (output_dir / "summary.json").write_text(json.dumps(summary, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(summary, sort_keys=True))
    return 0 if accepted else 2


if __name__ == "__main__":
    sys.exit(main())
