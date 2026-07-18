#!/usr/bin/env python3
"""Probe issue-source entity kinds through local C3 without leaking identifiers."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import subprocess
import sys
from collections import Counter
from datetime import datetime, timezone
from pathlib import Path
from typing import Sequence


CHANGE_DOC_TYPES = {"adr", "prd", "atomic-design-change"}
TERMINAL_CHANGE_DOC_STATUSES = {"done", "superseded"}


def sha256_bytes(value: bytes) -> str:
    return hashlib.sha256(value).hexdigest()


def sha256_file(path: Path) -> str:
    return sha256_bytes(path.read_bytes())


def scalar(text: str, key: str) -> str:
    match = re.search(rf"(?m)^{re.escape(key)}:\s*(.+?)\s*$", text)
    if not match:
        return ""
    raw = match.group(1).strip()
    if raw.startswith('"'):
        try:
            value = json.loads(raw)
            return value if isinstance(value, str) else raw
        except json.JSONDecodeError:
            return raw
    return raw


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description=__doc__)
    result.add_argument("--private-manifest", type=Path, required=True)
    result.add_argument("--family", required=True)
    result.add_argument("--c3-dir", type=Path, required=True)
    result.add_argument("--wrapper", type=Path, required=True)
    result.add_argument("--local-binary", type=Path, required=True)
    result.add_argument("--local-version", required=True)
    result.add_argument("--private-output", type=Path, required=True)
    result.add_argument("--timeout", type=int, default=60)
    return result


def main(argv: Sequence[str] | None = None) -> int:
    args = parser().parse_args(argv)
    manifest = json.loads(args.private_manifest.read_text(encoding="utf-8"))
    selected_issues = [
        row
        for row in manifest.get("issues", [])
        if row.get("family") == args.family and row.get("entity")
    ]
    entities = sorted(
        {
            row.get("entity", "")
            for row in selected_issues
        }
    )
    private_output = args.private_output.resolve()
    raw_dir = private_output.parent / f"{private_output.stem}-raw"
    raw_dir.mkdir(parents=True, exist_ok=True)
    env = os.environ.copy()
    env.update(
        {
            "C3X_MODE": "agent",
            "C3X_LOCAL_BINARY": str(args.local_binary.resolve()),
            "C3X_LOCAL_VERSION": args.local_version,
        }
    )
    rows: list[dict[str, object]] = []
    read_failure_count = 0
    parse_failure_count = 0
    for index, entity in enumerate(entities, 1):
        try:
            result = subprocess.run(
                ["bash", str(args.wrapper.resolve()), "--c3-dir", str(args.c3_dir.resolve()), "read", entity],
                env=env,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                timeout=args.timeout,
                check=False,
            )
            stdout, stderr, exit_code = result.stdout, result.stderr, result.returncode
        except subprocess.TimeoutExpired as exc:
            stdout = exc.stdout or b""
            stderr = (exc.stderr or b"") + b"\ntimeout\n"
            exit_code = 124
        (raw_dir / f"{index:02d}.stdout.toon").write_bytes(stdout)
        (raw_dir / f"{index:02d}.stderr.txt").write_bytes(stderr)
        text = stdout.decode(errors="replace")
        entity_type = scalar(text, "type")
        status = scalar(text, "status")
        if exit_code != 0:
            read_failure_count += 1
        if exit_code == 0 and not entity_type:
            parse_failure_count += 1
        rows.append(
            {
                "entity": entity,
                "type": entity_type,
                "status": status,
                "exit_code": exit_code,
                "stdout_sha256": sha256_bytes(stdout),
                "stderr_sha256": sha256_bytes(stderr),
            }
        )

    type_counts = Counter(str(row["type"]) for row in rows if row["type"])
    status_counts = Counter(str(row["status"]) for row in rows if row["status"])
    frozen_fact_count = sum(row["type"] not in CHANGE_DOC_TYPES for row in rows if row["type"])
    mutable_change_doc_count = sum(
        row["type"] in CHANGE_DOC_TYPES and row["status"] not in TERMINAL_CHANGE_DOC_STATUSES for row in rows
    )
    terminal_change_doc_count = sum(
        row["type"] in CHANGE_DOC_TYPES and row["status"] in TERMINAL_CHANGE_DOC_STATUSES for row in rows
    )
    type_by_entity = {str(row["entity"]): str(row["type"]) for row in rows}
    status_by_entity = {str(row["entity"]): str(row["status"]) for row in rows}
    issue_status_counts = Counter(status_by_entity.get(str(issue["entity"]), "") for issue in selected_issues)
    issue_status_counts.pop("", None)
    mutable_issue_count = sum(
        type_by_entity.get(str(issue["entity"])) in CHANGE_DOC_TYPES
        and status_by_entity.get(str(issue["entity"])) not in TERMINAL_CHANGE_DOC_STATUSES
        for issue in selected_issues
    )
    terminal_issue_count = sum(
        type_by_entity.get(str(issue["entity"])) in CHANGE_DOC_TYPES
        and status_by_entity.get(str(issue["entity"])) in TERMINAL_CHANGE_DOC_STATUSES
        for issue in selected_issues
    )
    summary = {
        "schema_version": 1,
        "entity_count": len(entities),
        "type_counts": dict(sorted(type_counts.items())),
        "status_counts": dict(sorted(status_counts.items())),
        "issue_status_counts": dict(sorted(issue_status_counts.items())),
        "mutable_issue_count": mutable_issue_count,
        "terminal_issue_count": terminal_issue_count,
        "frozen_fact_count": frozen_fact_count,
        "mutable_change_doc_count": mutable_change_doc_count,
        "terminal_change_doc_count": terminal_change_doc_count,
        "read_failure_count": read_failure_count,
        "parse_failure_count": parse_failure_count,
        "runtime": {
            "version": args.local_version,
            "binary_sha256": sha256_file(args.local_binary.resolve()),
            "wrapper_sha256": sha256_file(args.wrapper.resolve()),
        },
    }
    private = {
        "recorded_at": datetime.now(timezone.utc).isoformat(),
        "family": args.family,
        "manifest_sha256": sha256_file(args.private_manifest.resolve()),
        "rows": rows,
        "summary": summary,
    }
    private_output.parent.mkdir(parents=True, exist_ok=True)
    private_output.write_text(json.dumps(private, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(summary, sort_keys=True))
    return 0 if entities and read_failure_count == 0 and parse_failure_count == 0 else 2


if __name__ == "__main__":
    sys.exit(main())
