#!/usr/bin/env python3
"""Classify unknown compliance references without choosing replacement semantics."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Sequence


TARGET_RE = re.compile(r"Compliance Refs references unknown ref:\s*([A-Za-z0-9][A-Za-z0-9._-]*)")
CHANGE_DOC_TYPES = {"adr", "prd", "atomic-design-change"}
TERMINAL_STATUSES = {"done", "superseded", "implemented", "provisioned"}


def sha256_bytes(value: bytes) -> str:
    return hashlib.sha256(value).hexdigest()


def sha256_file(path: Path) -> str:
    return sha256_bytes(path.read_bytes())


def toon_scalar(text: str, key: str) -> str:
    match = re.search(rf"(?m)^{re.escape(key)}:\s*(.+?)\s*$", text)
    if not match:
        return ""
    raw = match.group(1).strip()
    if raw.startswith('"'):
        try:
            value = json.loads(raw)
            return value if isinstance(value, str) else ""
        except json.JSONDecodeError:
            return ""
    return raw


def compliance_action(body: str, target: str) -> str:
    match = re.search(r"(?ms)^##\s+Compliance Refs\s*$\n(.*?)(?=^##\s+|\Z)", body)
    if not match:
        return ""
    lines = [line.strip() for line in match.group(1).splitlines() if line.strip().startswith("|")]
    if len(lines) < 3:
        return ""
    headers = [cell.strip() for cell in lines[0].strip("|").split("|")]
    try:
        ref_index = headers.index("Ref")
        action_index = headers.index("Action")
    except ValueError:
        return ""
    for line in lines[2:]:
        cells = [cell.strip() for cell in line.strip("|").split("|")]
        if len(cells) > max(ref_index, action_index) and cells[ref_index] == target:
            return cells[action_index]
    return ""


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description=__doc__)
    result.add_argument("--private-manifest", type=Path, required=True)
    result.add_argument("--entity-probe", type=Path, required=True)
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
    entity_probe = json.loads(args.entity_probe.read_text(encoding="utf-8"))
    mutable = {
        str(row.get("entity"))
        for row in entity_probe.get("rows", [])
        if row.get("type") in CHANGE_DOC_TYPES
        and row.get("status") not in TERMINAL_STATUSES
        and row.get("exit_code") == 0
    }
    selected = [
        row
        for row in manifest.get("issues", [])
        if row.get("family") == args.family and str(row.get("entity")) in mutable
    ]
    parsed: list[dict[str, str]] = []
    parse_failure_count = 0
    for issue in selected:
        match = TARGET_RE.search(str(issue.get("message", "")))
        if not match:
            parse_failure_count += 1
            continue
        parsed.append({"source": str(issue["entity"]), "target": match.group(1)})

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
    wrapper = args.wrapper.resolve()
    c3_dir = args.c3_dir.resolve()
    command_count = 0

    def read(entity: str) -> tuple[str, dict[str, object]]:
        nonlocal command_count
        command_count += 1
        try:
            result = subprocess.run(
                ["bash", str(wrapper), "--c3-dir", str(c3_dir), "read", entity, "--full"],
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
        prefix = raw_dir / f"{command_count:02d}"
        prefix.with_suffix(".stdout.toon").write_bytes(stdout)
        prefix.with_suffix(".stderr.txt").write_bytes(stderr)
        return stdout.decode(errors="replace"), {
            "exit_code": exit_code,
            "stdout_sha256": sha256_bytes(stdout),
            "stderr_sha256": sha256_bytes(stderr),
        }

    source_bodies: dict[str, str] = {}
    source_records: dict[str, dict[str, object]] = {}
    source_read_failure_count = 0
    for source in sorted({row["source"] for row in parsed}):
        text, receipt = read(source)
        source_records[source] = receipt
        source_bodies[source] = toon_scalar(text, "body")
        if receipt["exit_code"] != 0:
            source_read_failure_count += 1

    target_records: dict[str, dict[str, object]] = {}
    target_exists: dict[str, bool] = {}
    for target in sorted({row["target"] for row in parsed}):
        _text, receipt = read(target)
        target_records[target] = receipt
        target_exists[target] = receipt["exit_code"] == 0

    rows: list[dict[str, object]] = []
    action_parse_failure_count = 0
    explicit_create_action_count = 0
    for row in parsed:
        action = compliance_action(source_bodies.get(row["source"], ""), row["target"])
        if not action:
            action_parse_failure_count += 1
        if "create" in action.lower():
            explicit_create_action_count += 1
        rows.append(
            {
                **row,
                "action": action,
                "target_exists": target_exists.get(row["target"], False),
                "source_read": source_records.get(row["source"]),
                "target_read": target_records.get(row["target"]),
            }
        )

    target_exists_count = sum(1 for value in target_exists.values() if value)
    target_missing_count = len(target_exists) - target_exists_count
    semantic_judgement_row_count = len(rows)
    summary = {
        "schema_version": 1,
        "selected_issue_count": len(selected),
        "source_entity_count": len(source_bodies),
        "target_entity_count": len(target_exists),
        "target_exists_count": target_exists_count,
        "target_missing_count": target_missing_count,
        "explicit_create_action_count": explicit_create_action_count,
        "semantic_judgement_row_count": semantic_judgement_row_count,
        "mechanical_repair_candidate_count": 0,
        "parse_failure_count": parse_failure_count,
        "action_parse_failure_count": action_parse_failure_count,
        "source_read_failure_count": source_read_failure_count,
        "wrapper_read_count": command_count,
        "runtime": {
            "version": args.local_version,
            "binary_sha256": sha256_file(args.local_binary.resolve()),
            "wrapper_sha256": sha256_file(wrapper),
        },
    }
    private = {
        "recorded_at": datetime.now(timezone.utc).isoformat(),
        "family": args.family,
        "manifest_sha256": sha256_file(args.private_manifest.resolve()),
        "entity_probe_sha256": sha256_file(args.entity_probe.resolve()),
        "rows": rows,
        "summary": summary,
    }
    private_output.parent.mkdir(parents=True, exist_ok=True)
    private_output.write_text(json.dumps(private, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(summary, sort_keys=True))
    complete = (
        len(selected) > 0
        and len(rows) == len(selected)
        and parse_failure_count == 0
        and action_parse_failure_count == 0
        and source_read_failure_count == 0
    )
    return 0 if complete else 2


if __name__ == "__main__":
    sys.exit(main())
