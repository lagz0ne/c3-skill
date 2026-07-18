#!/usr/bin/env python3
"""Read-only verifier for one saved C3 repair-step receipt."""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from collections import Counter
from datetime import datetime, timezone
from pathlib import Path
from typing import Sequence

from classify_c3_migration_debt import classify, count_value, parse_issues
from c3_dependency_links import dependency_link
from run_c3_repair_step import family_delta, git_changed_paths, identity_snapshot, non_c3_paths, sha256_bytes, sha256_file


COMMON_RAW_PREFIXES = (
    "01-before-check",
    "02-before-eval",
)
TAIL_RAW_PREFIXES = ("04-repair", "05-after-check", "06-after-eval")
EXPECTED_COMMAND_SEQUENCES = {
    ("check", "eval", "check", "repair", "check", "eval"): "03-check-fix",
    ("check", "eval", "set", "repair", "check", "eval"): "03-set-field",
}


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description=__doc__)
    result.add_argument("--evidence-dir", type=Path, required=True)
    result.add_argument("--project", type=Path, required=True)
    result.add_argument("--primary-target", type=Path, required=True)
    result.add_argument("--family", required=True)
    result.add_argument("--expect-before", type=int, required=True)
    result.add_argument("--expect-after", type=int, required=True)
    result.add_argument("--expected-family-reduction", type=int, default=1)
    result.add_argument("--expected-identity-loss", action="append", default=[], metavar="ID")
    result.add_argument("--expected-identity-addition", action="append", default=[], metavar="ID")
    result.add_argument("--expected-carrier-change", action="append", default=[], metavar="RELATIVE_PATH")
    result.add_argument("--expected-family-delta", type=family_delta, action="append", default=[])
    result.add_argument("--require-exact-family-deltas", action="store_true")
    result.add_argument("--local-binary", type=Path, required=True)
    result.add_argument("--local-version", required=True)
    result.add_argument("--wrapper", type=Path, required=True)
    result.add_argument("--dependency-link", type=dependency_link, action="append", default=[])
    return result


def direct_summary(check_path: Path, eval_path: Path) -> dict[str, object]:
    check_bytes = check_path.read_bytes()
    eval_bytes = eval_path.read_bytes()
    declared, issues = parse_issues(check_bytes.decode(errors="replace"))
    families = Counter(classify(issue.get("message", "")) for issue in issues)
    return {
        "structural_issue_count": len(issues),
        "declared_issue_count": declared,
        "unclassified_issue_count": families.get("unclassified", 0),
        "families": dict(families),
        "semantic_drift_count": count_value(eval_bytes.decode(errors="replace"), "drift"),
        "semantic_needs_judgement_count": count_value(eval_bytes.decode(errors="replace"), "needs_judgement"),
        "check_sha256": sha256_bytes(check_bytes),
        "eval_sha256": sha256_bytes(eval_bytes),
    }


def main(argv: Sequence[str] | None = None) -> int:
    args = parser().parse_args(argv)
    expected_family_deltas = dict(args.expected_family_delta)
    if len(expected_family_deltas) != len(args.expected_family_delta):
        raise ValueError("duplicate --expected-family-delta family")
    evidence = args.evidence_dir.resolve()
    private = evidence / "private"
    summary_path = evidence / "summary.json"
    report_path = private / "repair-report.json"
    failures: list[str] = []
    try:
        saved = json.loads(summary_path.read_text(encoding="utf-8"))
        report = json.loads(report_path.read_text(encoding="utf-8"))
        before_debt = json.loads((private / "before-debt.json").read_text(encoding="utf-8"))
        after_debt = json.loads((private / "after-debt.json").read_text(encoding="utf-8"))
        before_identity = json.loads((private / "before-identity.json").read_text(encoding="utf-8"))
        after_identity = json.loads((private / "after-identity.json").read_text(encoding="utf-8"))
        before = direct_summary(private / "01-before-check.stdout.toon", private / "02-before-eval.stdout.toon")
        after = direct_summary(private / "05-after-check.stdout.toon", private / "06-after-eval.stdout.toon")
    except (OSError, json.JSONDecodeError, KeyError, TypeError, ValueError):
        output = {
            "schema_version": 1,
            "decision": "rejected",
            "failure_count": 1,
            "before_issue_count": None,
            "after_issue_count": None,
            "family_reduction": None,
            "non_c3_source_change_count": None,
            "primary_target_change_count": None,
        }
        print(json.dumps(output, sort_keys=True))
        return 2

    records = report.get("records")
    if not isinstance(records, list) or len(records) != 6:
        failures.append("command_record_count")
        records = []
    command_sequence = tuple(record.get("command") for record in records) if records else ()
    mutation_prefix = EXPECTED_COMMAND_SEQUENCES.get(command_sequence)
    if records and mutation_prefix is None:
        failures.append("command_sequence")
    if records and command_sequence[2] == "set" and len(records[2].get("args", [])) != 3:
        failures.append("set_field_args")
    if records and any(record.get("exit_code") != 0 for record in records):
        failures.append("command_exit")
    raw_prefixes = (*COMMON_RAW_PREFIXES, mutation_prefix or "03-invalid", *TAIL_RAW_PREFIXES)
    for index, prefix in enumerate(raw_prefixes):
        if index >= len(records):
            break
        stdout_path = private / f"{prefix}.stdout.toon"
        stderr_path = private / f"{prefix}.stderr.txt"
        if not stdout_path.is_file() or sha256_file(stdout_path) != records[index].get("stdout_sha256"):
            failures.append(f"stdout_hash_{index + 1}")
        if not stderr_path.is_file() or sha256_file(stderr_path) != records[index].get("stderr_sha256"):
            failures.append(f"stderr_hash_{index + 1}")

    before_family = int(before["families"].get(args.family, 0))
    after_family = int(after["families"].get(args.family, 0))
    family_reduction = before_family - after_family
    actual_family_deltas = {
        family: int(before["families"].get(family, 0)) - int(after["families"].get(family, 0))
        for family in sorted(set(before["families"]) | set(after["families"]))
        if int(before["families"].get(family, 0)) != int(after["families"].get(family, 0))
    }
    expected_direct = {
        "before_issue_count": before["structural_issue_count"],
        "after_issue_count": after["structural_issue_count"],
        "before_family_count": before_family,
        "after_family_count": after_family,
        "family_reduction": family_reduction,
        "before_semantic_drift_count": before["semantic_drift_count"],
        "after_semantic_drift_count": after["semantic_drift_count"],
        "before_semantic_needs_judgement_count": before["semantic_needs_judgement_count"],
        "after_semantic_needs_judgement_count": after["semantic_needs_judgement_count"],
        "before_unclassified_issue_count": before["unclassified_issue_count"],
        "after_unclassified_issue_count": after["unclassified_issue_count"],
    }
    if "family_deltas" in saved:
        expected_direct["family_deltas"] = actual_family_deltas
    if before["declared_issue_count"] != before["structural_issue_count"]:
        failures.append("before_parse_count")
    if after["declared_issue_count"] != after["structural_issue_count"]:
        failures.append("after_parse_count")
    for key, value in expected_direct.items():
        if saved.get(key) != value:
            failures.append(f"saved_{key}")
    if report.get("summary") != saved:
        failures.append("report_summary")
    if saved.get("decision") != "accepted":
        failures.append("worker_decision")
    if before["structural_issue_count"] != args.expect_before:
        failures.append("before_target")
    if after["structural_issue_count"] != args.expect_after:
        failures.append("after_target")
    if family_reduction != args.expected_family_reduction:
        failures.append("family_reduction")
    if (expected_family_deltas or args.require_exact_family_deltas) and actual_family_deltas != expected_family_deltas:
        failures.append("exact_family_deltas")
    if (expected_family_deltas or args.require_exact_family_deltas) and report.get("expected_family_deltas") != expected_family_deltas:
        failures.append("reported_expected_family_deltas")
    if args.require_exact_family_deltas and report.get("require_exact_family_deltas") is not True:
        failures.append("reported_exact_family_mode")
    if before["unclassified_issue_count"] or after["unclassified_issue_count"]:
        failures.append("unclassified")
    if not isinstance(before["semantic_drift_count"], int) or not isinstance(after["semantic_drift_count"], int):
        failures.append("semantic_drift_missing")
    elif after["semantic_drift_count"] > before["semantic_drift_count"]:
        failures.append("semantic_drift_increase")
    if not isinstance(before["semantic_needs_judgement_count"], int) or not isinstance(after["semantic_needs_judgement_count"], int):
        failures.append("semantic_judgement_missing")
    elif after["semantic_needs_judgement_count"] > before["semantic_needs_judgement_count"]:
        failures.append("semantic_judgement_increase")

    for debt, direct, label in ((before_debt, before, "before"), (after_debt, after, "after")):
        if debt.get("check_sha256") != direct["check_sha256"]:
            failures.append(f"{label}_debt_check_hash")
        if debt.get("eval_sha256") != direct["eval_sha256"]:
            failures.append(f"{label}_debt_eval_hash")
        debt_summary = debt.get("summary", {})
        if debt_summary.get("structural_issue_count") != direct["structural_issue_count"]:
            failures.append(f"{label}_debt_issue_count")
        if debt_summary.get("unclassified_issue_count") != direct["unclassified_issue_count"]:
            failures.append(f"{label}_debt_unclassified")

    repair_non_c3 = non_c3_paths(args.project.resolve())
    primary_changes = git_changed_paths(args.primary_target.resolve())
    if repair_non_c3:
        failures.append("repair_scope")
    if primary_changes:
        failures.append("primary_scope")
    if saved.get("non_c3_source_change_count") != len(repair_non_c3):
        failures.append("saved_repair_scope")
    if saved.get("primary_target_change_count") != len(primary_changes):
        failures.append("saved_primary_scope")

    runtime = saved.get("runtime", {})
    if not args.local_binary.is_file() or runtime.get("binary_sha256") != sha256_file(args.local_binary):
        failures.append("runtime_binary")
    if not args.wrapper.is_file() or runtime.get("wrapper_sha256") != sha256_file(args.wrapper):
        failures.append("runtime_wrapper")
    if runtime.get("version") != args.local_version:
        failures.append("runtime_version")
    runtime_env = os.environ.copy()
    runtime_env.update({"C3X_MODE": "agent", "C3X_LOCAL_BINARY": str(args.local_binary), "C3X_LOCAL_VERSION": args.local_version})
    try:
        direct_version = subprocess.run(
            [str(args.local_binary), "--version"], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, check=False
        )
        direct_version_ok = direct_version.returncode == 0 and direct_version.stdout.strip() == args.local_version
    except OSError:
        direct_version_ok = False
    try:
        wrapped_version = subprocess.run(
            ["bash", str(args.wrapper), "--c3-dir", str(args.project.resolve() / ".c3"), "--version"],
            cwd=args.project.resolve(), env=runtime_env, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, check=False,
        )
        wrapped_version_ok = wrapped_version.returncode == 0 and wrapped_version.stdout.strip() == args.local_version
    except OSError:
        wrapped_version_ok = False
    if not direct_version_ok:
        failures.append("runtime_direct_version")
    if not wrapped_version_ok:
        failures.append("runtime_wrapper_selection")

    expected_links = [
        {"relative": relative.as_posix(), "source": str(source)}
        for relative, source in args.dependency_link
    ]
    links_cleaned = all(not os.path.lexists(args.project.resolve() / item["relative"]) for item in expected_links)
    if report.get("dependency_links") != expected_links:
        failures.append("dependency_link_contract")
    if not links_cleaned or saved.get("dependency_links_cleaned") is not True:
        failures.append("dependency_link_cleanup")
    if saved.get("dependency_link_count") != len(expected_links):
        failures.append("dependency_link_count")

    if saved.get("before_identity_sha256") != sha256_file(private / "before-identity.json") or report.get("before_identity_sha256") != saved.get("before_identity_sha256"):
        failures.append("before_identity_hash")
    if saved.get("after_identity_sha256") != sha256_file(private / "after-identity.json") or report.get("after_identity_sha256") != saved.get("after_identity_sha256"):
        failures.append("after_identity_hash")
    if after_identity != identity_snapshot(args.project.resolve() / ".c3"):
        failures.append("after_identity_manifest")
    before_ids = dict(before_identity.get("entity_ids", []))
    after_ids = dict(after_identity.get("entity_ids", []))
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
    before_carriers = dict(before_identity.get("change_carriers", []))
    after_carriers = dict(after_identity.get("change_carriers", []))
    carrier_changes = sorted(
        path for path in set(before_carriers) | set(after_carriers)
        if before_carriers.get(path) != after_carriers.get(path)
    )
    expected_losses = sorted(args.expected_identity_loss)
    expected_additions = sorted(args.expected_identity_addition)
    expected_carriers = sorted(args.expected_carrier_change)
    if identity_losses != expected_losses or identity_additions != expected_additions or carrier_changes != expected_carriers:
        failures.append("identity_deltas")
    identity_actual = {
        "identity_loss_count": len(identity_losses),
        "identity_addition_count": len(identity_additions),
        "carrier_change_count": len(carrier_changes),
        "identity_losses_sha256": sha256_bytes(json.dumps(identity_losses, separators=(",", ":")).encode()),
        "identity_additions_sha256": sha256_bytes(json.dumps(identity_additions, separators=(",", ":")).encode()),
        "carrier_changes_sha256": sha256_bytes(json.dumps(carrier_changes, separators=(",", ":")).encode()),
        "expected_identity_losses_sha256": sha256_bytes(json.dumps(expected_losses, separators=(",", ":")).encode()),
        "expected_identity_additions_sha256": sha256_bytes(json.dumps(expected_additions, separators=(",", ":")).encode()),
        "expected_carrier_changes_sha256": sha256_bytes(json.dumps(expected_carriers, separators=(",", ":")).encode()),
    }
    for key, value in identity_actual.items():
        if saved.get(key) != value:
            failures.append(f"saved_{key}")
    if report.get("expected_identity_losses") != expected_losses:
        failures.append("reported_expected_identity_losses")
    if report.get("expected_identity_additions") != expected_additions:
        failures.append("reported_expected_identity_additions")
    if report.get("expected_carrier_changes") != expected_carriers:
        failures.append("reported_expected_carrier_changes")

    output = {
        "schema_version": 1,
        "decision": "accepted" if not failures else "rejected",
        "failure_count": len(failures),
        "failure_codes": failures,
        "before_issue_count": before["structural_issue_count"],
        "after_issue_count": after["structural_issue_count"],
        "family_reduction": family_reduction,
        "before_semantic_drift_count": before["semantic_drift_count"],
        "after_semantic_drift_count": after["semantic_drift_count"],
        "before_semantic_needs_judgement_count": before["semantic_needs_judgement_count"],
        "after_semantic_needs_judgement_count": after["semantic_needs_judgement_count"],
        "non_c3_source_change_count": len(repair_non_c3),
        "primary_target_change_count": len(primary_changes),
        "summary_sha256": sha256_file(summary_path),
        "report_sha256": sha256_file(report_path),
        "observed_at": datetime.now(timezone.utc).isoformat(),
    }
    print(json.dumps(output, sort_keys=True))
    return 0 if not failures else 2


if __name__ == "__main__":
    sys.exit(main())
