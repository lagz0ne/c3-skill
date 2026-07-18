#!/usr/bin/env python3
"""Build an exact section repair for an Affected Topology Evidence-column cascade."""

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


TARGET_RE = re.compile(r"Affected Topology row for ([A-Za-z0-9][A-Za-z0-9._-]*) must include Evidence citation")
CITATION_RE = re.compile(
    r"([A-Za-z0-9][A-Za-z0-9._-]*)#n[0-9]+@v[0-9]+:sha256:[0-9a-fA-F]{64}(?:\s+\"[^\"]*\")?"
)


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


def section(body: str, name: str) -> str:
    match = re.search(rf"(?ms)^##\s+{re.escape(name)}\s*$\n(.*?)(?=^##\s+|\Z)", body)
    return match.group(1).strip() if match else ""


def parse_table(value: str) -> tuple[list[str], list[list[str]]] | None:
    lines = [line.strip() for line in value.splitlines() if line.strip()]
    if len(lines) < 3 or not all(line.startswith("|") and line.endswith("|") for line in lines):
        return None
    headers = [cell.strip() for cell in lines[0].strip("|").split("|")]
    rows = [[cell.strip() for cell in line.strip("|").split("|")] for line in lines[2:]]
    if any(len(row) != len(headers) for row in rows):
        return None
    return headers, rows


def render_table(headers: list[str], rows: list[list[str]]) -> str:
    header = "| " + " | ".join(headers) + " |"
    separator = "|" + "|".join("---" for _ in headers) + "|"
    body = ["| " + " | ".join(row) + " |" for row in rows]
    return "\n".join([header, separator, *body])


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description=__doc__)
    result.add_argument("--private-manifest", type=Path, required=True)
    result.add_argument("--entity-probe", type=Path, required=True)
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
    status_by_entity = {
        str(row.get("entity")): str(row.get("status"))
        for row in entity_probe.get("rows", [])
        if row.get("type") == "adr" and row.get("exit_code") == 0
    }
    column_issues = [row for row in manifest.get("issues", []) if row.get("family") == "missing_required_column"]
    evidence_issues = [row for row in manifest.get("issues", []) if row.get("family") == "missing_topology_evidence"]
    parse_failure_count = 0
    candidates: list[dict[str, str]] = []
    for issue in evidence_issues:
        match = TARGET_RE.search(str(issue.get("message", "")))
        if not match:
            parse_failure_count += 1
            continue
        source = str(issue.get("entity", ""))
        matching_columns = [
            row
            for row in column_issues
            if str(row.get("entity", "")) == source
            and 'missing required column "Evidence" in table: Affected Topology' in str(row.get("message", ""))
        ]
        if len(matching_columns) != 1:
            parse_failure_count += 1
            continue
        candidates.append({"source": source, "target": match.group(1)})

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
    read_failure_count = 0

    def read(entity: str, extra: list[str]) -> tuple[str, dict[str, object]]:
        nonlocal command_count, read_failure_count
        command_count += 1
        try:
            result = subprocess.run(
                ["bash", str(wrapper), "--c3-dir", str(c3_dir), "read", entity, *extra],
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
        if exit_code != 0:
            read_failure_count += 1
        prefix = raw_dir / f"{command_count:02d}"
        prefix.with_suffix(".stdout.toon").write_bytes(stdout)
        prefix.with_suffix(".stderr.txt").write_bytes(stderr)
        return stdout.decode(errors="replace"), {
            "exit_code": exit_code,
            "stdout_sha256": sha256_bytes(stdout),
            "stderr_sha256": sha256_bytes(stderr),
        }

    plans: list[dict[str, object]] = []
    shape_failure_count = 0
    for candidate in candidates:
        source_text, source_receipt = read(candidate["source"], ["--full"])
        old_section = section(toon_scalar(source_text, "body"), "Affected Topology")
        parsed = parse_table(old_section)
        if not parsed:
            shape_failure_count += 1
            continue
        headers, rows = parsed
        if "Evidence" in headers or headers[:3] != ["Entity", "Type", "Why affected"] or not rows:
            shape_failure_count += 1
            continue
        target_rows = [row for row in rows if row[0] == candidate["target"]]
        other_rows = [row for row in rows if row[0] != candidate["target"]]
        if len(target_rows) != 1 or any(not row[0].startswith("N.A -") for row in other_rows):
            shape_failure_count += 1
            continue
        target_text, target_receipt = read(candidate["target"], ["--section", "Goal", "--cite"])
        handles = [match.group(0) for match in CITATION_RE.finditer(target_text) if match.group(1) == candidate["target"]]
        if len(handles) != 1:
            shape_failure_count += 1
            continue
        insert_at = headers.index("Governance review") if "Governance review" in headers else len(headers)
        new_headers = [*headers[:insert_at], "Evidence", *headers[insert_at:]]
        new_rows = [
            [*row[:insert_at], handles[0] if row[0] == candidate["target"] else row[0], *row[insert_at:]]
            for row in rows
        ]
        plans.append(
            {
                **candidate,
                "source_status": status_by_entity.get(candidate["source"], ""),
                "section": "Affected Topology",
                "old_section": old_section,
                "new_section": render_table(new_headers, new_rows),
                "handle": handles[0],
                "source_read": source_receipt,
                "target_read": target_receipt,
            }
        )

    status_counts = Counter(str(plan["source_status"]) for plan in plans if plan["source_status"])
    summary = {
        "schema_version": 1,
        "candidate_cluster_count": len(candidates),
        "safe_section_repair_count": len(plans),
        "expected_issue_reduction": len(plans) * 2,
        "source_status_counts": dict(sorted(status_counts.items())),
        "parse_failure_count": parse_failure_count,
        "shape_failure_count": shape_failure_count,
        "read_failure_count": read_failure_count,
        "wrapper_read_count": command_count,
        "runtime": {
            "version": args.local_version,
            "binary_sha256": sha256_file(args.local_binary.resolve()),
            "wrapper_sha256": sha256_file(wrapper),
        },
    }
    private = {
        "recorded_at": datetime.now(timezone.utc).isoformat(),
        "manifest_sha256": sha256_file(args.private_manifest.resolve()),
        "entity_probe_sha256": sha256_file(args.entity_probe.resolve()),
        "plans": plans,
        "summary": summary,
    }
    private_output.parent.mkdir(parents=True, exist_ok=True)
    private_output.write_text(json.dumps(private, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(summary, sort_keys=True))
    clean = (
        len(candidates) == 1
        and len(plans) == 1
        and status_counts == {"open": 1}
        and parse_failure_count == 0
        and shape_failure_count == 0
        and read_failure_count == 0
    )
    return 0 if clean else 2


if __name__ == "__main__":
    sys.exit(main())
