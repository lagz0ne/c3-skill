#!/usr/bin/env python3
"""Classify missing compliance rows without inventing their semantic content."""

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


CHANGE_DOC_TYPES = {"adr", "prd", "atomic-design-change"}
TERMINAL_STATUSES = {"done", "superseded", "implemented", "provisioned"}
MISSING_REF_RE = re.compile(
    r"ADR missing compliance ref (?P<target>[A-Za-z0-9][A-Za-z0-9._-]*) \((?P<reason>.+)\)$"
)
CITATION_RE = re.compile(
    r"[A-Za-z0-9][A-Za-z0-9._-]*#n[0-9]+@v[0-9]+:sha256:[0-9a-fA-F]{64}(?:\s+\"[^\"]*\")?"
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


def sections(body: str) -> list[str]:
    return [match.group(1).strip() for match in re.finditer(r"(?m)^##\s+(.+?)\s*$", body)]


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
    parsed_rows: list[dict[str, str]] = []
    parse_failure_count = 0
    for issue in selected:
        match = MISSING_REF_RE.search(str(issue.get("message", "")))
        if not match:
            parse_failure_count += 1
            continue
        parsed_rows.append(
            {
                "source": str(issue["entity"]),
                "target": match.group("target"),
                "reason": match.group("reason"),
            }
        )

    sources = sorted({row["source"] for row in parsed_rows})
    targets = sorted({row["target"] for row in parsed_rows})
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

    def read(entity: str, extra: list[str]) -> tuple[str, dict[str, object]]:
        nonlocal command_count
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
        prefix = raw_dir / f"{command_count:02d}"
        prefix.with_suffix(".stdout.toon").write_bytes(stdout)
        prefix.with_suffix(".stderr.txt").write_bytes(stderr)
        return stdout.decode(errors="replace"), {
            "exit_code": exit_code,
            "stdout_sha256": sha256_bytes(stdout),
            "stderr_sha256": sha256_bytes(stderr),
        }

    source_records: list[dict[str, object]] = []
    source_read_failure_count = 0
    source_with_compliance_section_count = 0
    for source in sources:
        text, receipt = read(source, ["--full"])
        body = toon_scalar(text, "body")
        has_section = bool(re.search(r"(?m)^##\s+Compliance Refs\s*$", body))
        if receipt["exit_code"] != 0:
            source_read_failure_count += 1
        if has_section:
            source_with_compliance_section_count += 1
        source_records.append({"entity": source, "has_compliance_section": has_section, "read": receipt})

    target_records: list[dict[str, object]] = []
    target_read_failure_count = 0
    target_with_citation_count = 0
    for target in targets:
        full_text, full_receipt = read(target, ["--full"])
        target_sections = sections(toon_scalar(full_text, "body")) if full_receipt["exit_code"] == 0 else []
        handles: list[str] = []
        cite_receipts: list[dict[str, object]] = []
        cite_failed = False
        for section in target_sections:
            cite_text, cite_receipt = read(target, ["--section", section, "--cite"])
            cite_receipts.append(cite_receipt)
            if cite_receipt["exit_code"] != 0:
                cite_failed = True
                continue
            handles.extend(CITATION_RE.findall(cite_text))
        if full_receipt["exit_code"] != 0 or cite_failed or not target_sections:
            target_read_failure_count += 1
        if handles:
            target_with_citation_count += 1
        target_records.append(
            {
                "entity": target,
                "sections": target_sections,
                "citation_handles": handles,
                "full_read": full_receipt,
                "cite_reads": cite_receipts,
            }
        )

    # Existence and current cite material do not decide the required Why/Action
    # fields. Those rows remain semantic work until reviewed against each change.
    summary = {
        "schema_version": 1,
        "mutable_issue_count": len(selected),
        "source_entity_count": len(sources),
        "target_entity_count": len(targets),
        "source_with_compliance_section_count": source_with_compliance_section_count,
        "target_with_citation_count": target_with_citation_count,
        "semantic_judgement_row_count": len(parsed_rows),
        "mechanical_repair_candidate_count": 0,
        "parse_failure_count": parse_failure_count,
        "source_read_failure_count": source_read_failure_count,
        "target_read_failure_count": target_read_failure_count,
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
        "rows": parsed_rows,
        "source_records": source_records,
        "target_records": target_records,
        "summary": summary,
    }
    private_output.parent.mkdir(parents=True, exist_ok=True)
    private_output.write_text(json.dumps(private, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(summary, sort_keys=True))
    complete = (
        len(selected) > 0
        and len(parsed_rows) == len(selected)
        and source_with_compliance_section_count == len(sources)
        and target_with_citation_count == len(targets)
        and parse_failure_count == 0
        and source_read_failure_count == 0
        and target_read_failure_count == 0
    )
    return 0 if complete else 2


if __name__ == "__main__":
    sys.exit(main())
