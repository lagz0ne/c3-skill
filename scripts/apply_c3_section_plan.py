#!/usr/bin/env python3
"""Apply one exact private C3 section plan with count, semantic, and scope gates."""

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

from migrate_c3_project import canonical_manifest
from run_c3_repair_step import classify, git_changed_paths, non_c3_paths, sha256_bytes, sha256_file
from c3_dependency_links import dependency_link, with_dependency_links


ROOT = Path(__file__).resolve().parents[1]


def family_reduction(value: str) -> tuple[str, int]:
    family, separator, raw_count = value.partition("=")
    if not separator or not family or not raw_count.isdigit() or int(raw_count) < 1:
        raise argparse.ArgumentTypeError("expected FAMILY=positive-integer")
    return family, int(raw_count)


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description=__doc__)
    result.add_argument("--project", type=Path, required=True)
    result.add_argument("--primary-target", type=Path, required=True)
    result.add_argument("--plan", type=Path, required=True)
    result.add_argument("--output-dir", type=Path, required=True)
    result.add_argument("--wrapper", type=Path, required=True)
    result.add_argument("--local-binary", type=Path, required=True)
    result.add_argument("--local-version", required=True)
    result.add_argument("--expect-before", type=int, required=True)
    result.add_argument("--expect-after", type=int, required=True)
    result.add_argument("--expected-family-reduction", type=family_reduction, action="append", required=True)
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


def command_record(
    wrapper: Path,
    c3_dir: Path,
    command: str,
    command_args: list[str],
    output_prefix: Path,
    env: dict[str, str],
    timeout: int,
    stdin: bytes | None = None,
) -> dict[str, object]:
    started = time.monotonic_ns()
    try:
        result = subprocess.run(
            ["bash", str(wrapper), "--c3-dir", str(c3_dir), command, *command_args],
            cwd=ROOT,
            env=env,
            input=stdin,
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
        "args_sha256": hashlib.sha256(json.dumps(command_args).encode()).hexdigest(),
        "exit_code": exit_code,
        "elapsed_ms": elapsed_ms,
        "stdout_bytes": len(stdout),
        "stderr_bytes": len(stderr),
        "stdout_sha256": sha256_bytes(stdout),
        "stderr_sha256": sha256_bytes(stderr),
        "stdin_sha256": sha256_bytes(stdin or b""),
    }


def main(argv: Sequence[str] | None = None) -> int:
    args = parser().parse_args(argv)
    project = args.project.resolve()
    primary = args.primary_target.resolve()
    plan_path = args.plan.resolve()
    output_dir = args.output_dir.resolve()
    wrapper = args.wrapper.resolve()
    binary = args.local_binary.resolve()
    expected_reductions = dict(args.expected_family_reduction)
    if len(expected_reductions) != len(args.expected_family_reduction):
        raise ValueError("duplicate expected family")
    for path, label in ((project, "project"), (primary, "primary target")):
        if not path.is_dir() or not (path / ".c3").is_dir():
            raise ValueError(f"{label} must contain .c3")
        if output_dir == path or output_dir.is_relative_to(path):
            raise ValueError("private output directory must stay outside both worktrees")
    if not plan_path.is_file() or not wrapper.is_file() or not binary.is_file():
        raise ValueError("plan, wrapper, and binary must exist")
    if non_c3_paths(project):
        raise ValueError("repair worktree already has changes outside .c3")
    if git_changed_paths(primary):
        raise ValueError("primary target must be clean before repair")
    plan_document = json.loads(plan_path.read_text(encoding="utf-8"))
    plans = plan_document.get("plans", [])
    if not isinstance(plans, list) or not plans:
        raise ValueError("plan must contain at least one section repair")
    for plan in plans:
        if not isinstance(plan, dict) or not all(
            isinstance(plan.get(key), str) and plan[key] for key in ("source", "section", "old_section", "new_section")
        ):
            raise ValueError("each repair requires source, section, old_section, and new_section")

    output_dir.mkdir(parents=True, exist_ok=True)
    private = output_dir / "private"
    private.mkdir(parents=True, exist_ok=True)
    env = os.environ.copy()
    env.update(
        {"C3X_MODE": "agent", "C3X_LOCAL_BINARY": str(binary), "C3X_LOCAL_VERSION": args.local_version}
    )
    records: list[dict[str, object]] = []

    def run(label: str, command: str, command_args: list[str], stdin: bytes | None = None) -> dict[str, object]:
        execute = lambda: command_record(
            wrapper, project / ".c3", command, command_args, private / label, env, args.timeout, stdin
        )
        record = with_dependency_links(project, args.dependency_link, execute) if command == "eval" else execute()
        records.append(record)
        return record

    before_tree = canonical_manifest(project / ".c3")
    before_check = run("01-before-check", "check", ["--include-adr"])
    before_eval = run("02-before-eval", "eval", [])
    before_summary, before_classifier = classify(
        private / "01-before-check.stdout.toon", private / "02-before-eval.stdout.toon", private / "before-debt.json"
    )
    baseline_ok = (
        before_check["exit_code"] == 0
        and before_eval["exit_code"] == 0
        and before_classifier["exit_code"] == 0
        and before_summary.get("structural_issue_count") == args.expect_before
        and before_summary.get("unclassified_issue_count") == 0
        and before_summary.get("parse_mismatch_count") == 0
        and all(int(before_summary.get("families", {}).get(family, 0)) >= count for family, count in expected_reductions.items())
    )
    write_records: list[dict[str, object]] = []
    if baseline_ok:
        for index, plan in enumerate(plans, 1):
            write_records.append(
                run(
                    f"03-write-{index:02d}",
                    "write",
                    [str(plan["source"]), "--section", str(plan["section"])],
                    str(plan["new_section"]).encode(),
                )
            )
    repair = run("04-repair", "repair", ["--include-adr"]) if baseline_ok else {"exit_code": 2, "elapsed_ms": 0}
    after_check = run("05-after-check", "check", ["--include-adr"]) if baseline_ok else {"exit_code": 2, "elapsed_ms": 0}
    after_eval = run("06-after-eval", "eval", []) if baseline_ok else {"exit_code": 2, "elapsed_ms": 0}
    after_summary: dict[str, object] = {}
    after_classifier: dict[str, object] = {"exit_code": 2}
    if baseline_ok:
        after_summary, after_classifier = classify(
            private / "05-after-check.stdout.toon", private / "06-after-eval.stdout.toon", private / "after-debt.json"
        )
    after_tree = canonical_manifest(project / ".c3")
    actual_reductions = {
        family: int(before_summary.get("families", {}).get(family, 0))
        - int(after_summary.get("families", {}).get(family, 0))
        for family in expected_reductions
    }
    before_drift = before_summary.get("semantic_drift_count")
    after_drift = after_summary.get("semantic_drift_count")
    before_judgement = before_summary.get("semantic_needs_judgement_count")
    after_judgement = after_summary.get("semantic_needs_judgement_count")
    scope_changes = non_c3_paths(project)
    primary_changes = git_changed_paths(primary)
    accepted = bool(
        baseline_ok
        and all(record["exit_code"] == 0 for record in write_records)
        and repair["exit_code"] == 0
        and after_check["exit_code"] == 0
        and after_eval["exit_code"] == 0
        and after_classifier["exit_code"] == 0
        and after_summary.get("structural_issue_count") == args.expect_after
        and actual_reductions == expected_reductions
        and after_summary.get("unclassified_issue_count") == 0
        and after_summary.get("parse_mismatch_count") == 0
        and isinstance(before_drift, int)
        and isinstance(after_drift, int)
        and after_drift <= before_drift
        and isinstance(before_judgement, int)
        and isinstance(after_judgement, int)
        and after_judgement <= before_judgement
        and not scope_changes
        and not primary_changes
    )

    rollback_attempted = False
    rollback_success = False
    if baseline_ok and not accepted and write_records:
        rollback_attempted = True
        rollback_records = []
        for index, plan in enumerate(reversed(plans), 1):
            rollback_records.append(
                run(
                    f"90-rollback-{index:02d}",
                    "write",
                    [str(plan["source"]), "--section", str(plan["section"])],
                    str(plan["old_section"]).encode(),
                )
            )
        rollback_repair = run("91-rollback-repair", "repair", ["--include-adr"])
        rollback_tree = canonical_manifest(project / ".c3")
        rollback_success = all(record["exit_code"] == 0 for record in rollback_records) and rollback_repair[
            "exit_code"
        ] == 0 and rollback_tree["tree_sha256"] == before_tree["tree_sha256"]

    summary = {
        "schema_version": 1,
        "decision": "accepted" if accepted else ("blocked_admissibility" if not baseline_ok else "blocked"),
        "before_issue_count": before_summary.get("structural_issue_count"),
        "after_issue_count": after_summary.get("structural_issue_count"),
        "family_reductions": actual_reductions,
        "before_semantic_drift_count": before_drift,
        "after_semantic_drift_count": after_drift,
        "before_semantic_needs_judgement_count": before_judgement,
        "after_semantic_needs_judgement_count": after_judgement,
        "before_unclassified_issue_count": before_summary.get("unclassified_issue_count"),
        "after_unclassified_issue_count": after_summary.get("unclassified_issue_count"),
        "section_write_count": len(write_records),
        "non_c3_source_change_count": len(scope_changes),
        "primary_target_change_count": len(primary_changes),
        "canonical_c3_changed": before_tree["tree_sha256"] != after_tree["tree_sha256"],
        "rollback_attempted": rollback_attempted,
        "rollback_success": rollback_success,
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
        "plan_sha256": sha256_file(plan_path),
        "before_c3_tree_sha256": before_tree["tree_sha256"],
        "after_c3_tree_sha256": after_tree["tree_sha256"],
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
