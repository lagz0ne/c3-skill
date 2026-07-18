#!/usr/bin/env python3
"""Classify C3 check/eval debt without leaking raw project fields to stdout."""

from __future__ import annotations

import argparse
import hashlib
import json
import re
import sys
from collections import Counter
from datetime import datetime, timezone
from pathlib import Path
from typing import Sequence


FAMILY_PATTERNS = (
    ("missing_required_column", re.compile(r"missing required column")),
    ("missing_topology_evidence", re.compile(r"Affected Topology row.*Evidence citation")),
    ("unknown_compliance_ref", re.compile(r"Compliance Refs references unknown ref")),
    ("missing_adr_compliance_ref", re.compile(r"ADR missing compliance ref")),
    ("empty_affected_topology", re.compile(r"change doc touches nothing")),
    ("ready_to_auto_done", re.compile(r"ready to auto-done")),
    ("top_down_incomplete", re.compile(r"top-down incomplete")),
    ("citation_entity_mismatch", re.compile(r"(?:citation to|Affected Topology row).*cites node \d+ from")),
    ("stale_citation", re.compile(r"cites version .* current version|stale cite")),
    ("empty_citation_snippet", re.compile(r"citation to .* has empty snippet")),
    ("unknown_entity_citation", re.compile(r"citation references unknown entity")),
)

REPAIR_PRIORITY = (
    "ready_to_auto_done",
    "citation_entity_mismatch",
    "empty_citation_snippet",
    "stale_citation",
    "unknown_entity_citation",
    "unknown_compliance_ref",
    "empty_affected_topology",
    "top_down_incomplete",
    "missing_topology_evidence",
    "missing_required_column",
    "missing_adr_compliance_ref",
)


def sha256_bytes(value: bytes) -> str:
    return hashlib.sha256(value).hexdigest()


def scalar(value: str) -> str:
    value = value.strip()
    if value.startswith('"'):
        try:
            parsed = json.loads(value)
            if isinstance(parsed, str):
                return parsed
        except json.JSONDecodeError:
            pass
    return value


def parse_issues(text: str) -> tuple[int | None, list[dict[str, str]]]:
    declared_match = re.search(r"(?m)^issues\[(\d+)\]:", text)
    declared = int(declared_match.group(1)) if declared_match else (
        0 if re.search(r"(?m)^ok:\s*true\s*$", text) else None
    )
    issues: list[dict[str, str]] = []
    current: dict[str, str] | None = None
    in_issues = False
    for line in text.splitlines():
        if re.fullmatch(r"issues\[\d+\]:", line):
            in_issues = True
            continue
        if not in_issues:
            continue
        if line == "  -":
            if current is not None:
                issues.append(current)
            current = {}
            continue
        if current is None:
            if line and not line.startswith(" "):
                break
            continue
        match = re.match(r"^    ([a-z_]+):\s*(.*)$", line)
        if match:
            current[match.group(1)] = scalar(match.group(2))
            continue
        if line and not line.startswith(" "):
            break
    if current is not None:
        issues.append(current)
    return declared, issues


def count_value(text: str, key: str) -> int | None:
    match = re.search(rf"(?m)^{re.escape(key)}:\s*(\d+)\s*$", text)
    return int(match.group(1)) if match else None


def classify(message: str) -> str:
    for family, pattern in FAMILY_PATTERNS:
        if pattern.search(message):
            return family
    return "unclassified"


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description=__doc__)
    result.add_argument("--check", type=Path, required=True, help="private C3 check TOON output")
    result.add_argument("--eval", type=Path, required=True, help="private C3 eval TOON output")
    result.add_argument("--private-output", type=Path, required=True, help="private manifest path outside the public repository")
    return result


def main(argv: Sequence[str] | None = None) -> int:
    args = parser().parse_args(argv)
    check_bytes = args.check.read_bytes()
    eval_bytes = args.eval.read_bytes()
    declared, issues = parse_issues(check_bytes.decode(errors="replace"))
    classified: list[dict[str, str]] = []
    for index, issue in enumerate(issues, 1):
        message = issue.get("message", "")
        classified.append({
            "index": str(index),
            "severity": issue.get("severity", ""),
            "entity": issue.get("entity", ""),
            "message": message,
            "message_sha256": sha256_bytes(message.encode()),
            "family": classify(message),
        })

    family_counts = Counter(row["family"] for row in classified)
    unclassified = family_counts.get("unclassified", 0)
    parse_mismatch = int(declared is None or declared != len(classified))
    first_repair = next((family for family in REPAIR_PRIORITY if family_counts.get(family, 0)), None)
    eval_text = eval_bytes.decode(errors="replace")
    summary = {
        "schema_version": 1,
        "structural_issue_count": len(classified),
        "classified_issue_count": len(classified) - unclassified,
        "unclassified_issue_count": unclassified,
        "parse_mismatch_count": parse_mismatch,
        "families": dict(sorted(family_counts.items())),
        "semantic_drift_count": count_value(eval_text, "drift"),
        "semantic_needs_judgement_count": count_value(eval_text, "needs_judgement"),
        "first_repair_family": first_repair,
    }
    private = {
        "schema_version": 1,
        "recorded_at": datetime.now(timezone.utc).isoformat(),
        "check_sha256": sha256_bytes(check_bytes),
        "eval_sha256": sha256_bytes(eval_bytes),
        "declared_issue_count": declared,
        "issues": classified,
        "summary": summary,
    }
    args.private_output.parent.mkdir(parents=True, exist_ok=True)
    args.private_output.write_text(json.dumps(private, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(summary, sort_keys=True))
    return 0 if not unclassified and not parse_mismatch else 2


if __name__ == "__main__":
    sys.exit(main())
