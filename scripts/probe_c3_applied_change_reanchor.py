#!/usr/bin/env python3
"""Build an exact citation re-anchor plan from an already-applied C3 change carrier."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import subprocess
import sys
from collections import defaultdict
from datetime import datetime, timezone
from pathlib import Path
from typing import Sequence

from probe_c3_citation_repair import CITATION_RE, sections, sha256_bytes, sha256_file, source_section, toon_scalar


MESSAGE_RES = (
    re.compile(
        r"citation to (?P<target>[A-Za-z0-9][A-Za-z0-9._-]*) cites node "
        r"(?P<node>[0-9]+) from (?P<owner>[A-Za-z0-9][A-Za-z0-9._-]*)"
    ),
    re.compile(
        r"Evidence for .+? row (?P<target>[A-Za-z0-9][A-Za-z0-9._-]*) cites node "
        r"(?P<node>[0-9]+) from (?P<owner>[A-Za-z0-9][A-Za-z0-9._-]*)"
    ),
)
TARGET_ONLY_RES = {
    "stale_citation": re.compile(
        r"Evidence for .+? row (?P<target>[A-Za-z0-9][A-Za-z0-9._-]*) has a stale cite"
    ),
    "empty_citation_snippet": re.compile(
        r"citation to (?P<target>[A-Za-z0-9][A-Za-z0-9._-]*) has empty snippet"
    ),
}


def parse_issue(message: str, family: str) -> tuple[str, int | None, str] | None:
    target_only = TARGET_ONLY_RES.get(family)
    if target_only:
        match = target_only.search(message)
        return (match.group("target"), None, "") if match else None
    for pattern in MESSAGE_RES:
        match = pattern.search(message)
        if match:
            return match.group("target"), int(match.group("node")), match.group("owner")
    return None


def parse_carrier(path: Path) -> dict[str, str] | None:
    text = path.read_text(encoding="utf-8")
    match = re.fullmatch(r"---\n(?P<header>.*?)\n---\n(?P<body>.*)", text, re.DOTALL)
    if not match:
        return None
    header: dict[str, str] = {}
    for line in match.group("header").splitlines():
        key, separator, value = line.partition(":")
        if not separator:
            return None
        header[key.strip()] = value.strip()
    if not all(header.get(key) for key in ("target", "scope", "base")):
        return None
    return {**header, "body": match.group("body").rstrip("\n")}


def citation_core(match: re.Match[str]) -> str:
    return (
        f'{match.group("entity")}#n{match.group("node")}@v{match.group("version")}:'
        f'sha256:{match.group("hash").lower()}'
    )


def section_content(body: str, heading: str) -> str:
    match = re.search(
        rf"(?ms)^##\s+{re.escape(heading)}\s*\n+(?P<content>.*?)(?=^##\s+|\Z)",
        body,
    )
    return match.group("content").rstrip() if match else ""


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description=__doc__)
    result.add_argument("--private-manifest", type=Path, required=True)
    result.add_argument("--entity-probe", type=Path, required=True)
    result.add_argument("--c3-dir", type=Path, required=True)
    result.add_argument("--wrapper", type=Path, required=True)
    result.add_argument("--local-binary", type=Path, required=True)
    result.add_argument("--local-version", required=True)
    result.add_argument("--private-output", type=Path, required=True)
    result.add_argument(
        "--family",
        action="append",
        choices=["citation_entity_mismatch", "stale_citation", "empty_citation_snippet"],
        help="issue family to reconcile; defaults to citation_entity_mismatch",
    )
    result.add_argument("--timeout", type=int, default=60)
    return result


def main(argv: Sequence[str] | None = None) -> int:
    args = parser().parse_args(argv)
    manifest = json.loads(args.private_manifest.read_text(encoding="utf-8"))
    entity_probe = json.loads(args.entity_probe.read_text(encoding="utf-8"))
    mutable = {
        row.get("entity")
        for row in entity_probe.get("rows", [])
        if row.get("type") in {"adr", "prd", "atomic-design-change"}
        and row.get("status") not in {"done", "superseded", "implemented", "provisioned"}
        and row.get("exit_code") == 0
    }
    selected_families = set(args.family or ["citation_entity_mismatch"])
    issue_rows = [
        row
        for row in manifest.get("issues", [])
        if row.get("family") in selected_families and row.get("entity") in mutable
    ]
    grouped: dict[tuple[str, str, int | None, str], list[dict[str, object]]] = defaultdict(list)
    parse_failure_count = 0
    for row in issue_rows:
        parsed = parse_issue(str(row.get("message", "")), str(row.get("family", "")))
        if parsed is None:
            parse_failure_count += 1
            continue
        target, node, owner = parsed
        grouped[(str(row["entity"]), target, node, owner)].append(row)

    private_output = args.private_output.resolve()
    raw_dir = private_output.parent / f"{private_output.stem}-raw"
    raw_dir.mkdir(parents=True, exist_ok=True)
    wrapper = args.wrapper.resolve()
    binary = args.local_binary.resolve()
    c3_dir = args.c3_dir.resolve()
    env = os.environ.copy()
    env.update(
        {
            "C3X_MODE": "agent",
            "C3X_LOCAL_BINARY": str(binary),
            "C3X_LOCAL_VERSION": args.local_version,
        }
    )
    command_count = 0
    read_failure_count = 0

    def read(entity: str, command_args: list[str]) -> tuple[str, dict[str, object]]:
        nonlocal command_count, read_failure_count
        command_count += 1
        try:
            result = subprocess.run(
                ["bash", str(wrapper), "--c3-dir", str(c3_dir), "read", entity, *command_args],
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
    reconciled_issue_row_count = 0
    missing_carrier_count = 0
    ambiguous_carrier_count = 0
    carrier_parse_failure_count = 0
    missing_old_citation_count = 0
    ambiguous_old_citation_count = 0
    missing_current_match_count = 0
    ambiguous_current_match_count = 0
    source_section_missing_count = 0

    for (source_id, target_id, old_node, wrong_owner), rows in grouped.items():
        source_text, source_read = read(source_id, ["--full"])
        source_body = toon_scalar(source_text, "body")
        old_matches = [
            match
            for match in CITATION_RE.finditer(source_body)
            if match.group("entity") == target_id
            and (old_node is None or int(match.group("node")) == old_node)
        ]
        if len(old_matches) != 1:
            if old_matches:
                ambiguous_old_citation_count += 1
            else:
                missing_old_citation_count += 1
            continue
        old = old_matches[0]
        heading = source_section(source_body, old.start())
        old_section = section_content(source_body, heading) if heading else ""
        if not heading or not old_section:
            source_section_missing_count += 1
            continue

        carrier_paths = sorted((c3_dir / "changes" / source_id).glob("*.patch.md"))
        parsed_carriers: list[tuple[Path, dict[str, str]]] = []
        for path in carrier_paths:
            raw_carrier = path.read_text(encoding="utf-8")
            header_match = re.match(r"---\n(?P<header>.*?)\n---\n", raw_carrier, re.DOTALL)
            if not header_match or not re.search(
                rf"(?m)^target:\s*{re.escape(target_id)}\s*$", header_match.group("header")
            ):
                continue
            carrier = parse_carrier(path)
            if carrier is None:
                carrier_parse_failure_count += 1
                continue
            if (
                carrier["target"] == target_id
                and carrier["scope"] == "block"
                and carrier["base"] == citation_core(old)
            ):
                parsed_carriers.append((path, carrier))
        if len(parsed_carriers) != 1:
            if parsed_carriers:
                ambiguous_carrier_count += 1
            else:
                missing_carrier_count += 1
            continue
        carrier_path, carrier = parsed_carriers[0]
        content_hash = hashlib.sha256(carrier["body"].encode()).hexdigest()

        target_text, target_read = read(target_id, ["--full"])
        target_body = toon_scalar(target_text, "body")
        cite_reads: list[dict[str, object]] = []
        current_matches: dict[str, re.Match[str]] = {}
        for target_section in sections(target_body):
            cite_text, cite_read = read(target_id, ["--section", target_section, "--cite"])
            cite_reads.append(cite_read)
            for candidate in CITATION_RE.finditer(cite_text):
                if candidate.group("entity") == target_id and candidate.group("hash").lower() == content_hash:
                    current_matches[citation_core(candidate)] = candidate
        if len(current_matches) != 1:
            if current_matches:
                ambiguous_current_match_count += 1
            else:
                missing_current_match_count += 1
            continue
        new_handle = next(iter(current_matches))
        old_handle = old.group(0)
        if old_section.count(old_handle) != 1:
            ambiguous_old_citation_count += 1
            continue
        new_section = old_section.replace(old_handle, new_handle, 1)
        plans.append(
            {
                "source": source_id,
                "section": heading,
                "old_section": old_section,
                "new_section": new_section,
                "target_entity": target_id,
                "wrong_owner": wrong_owner,
                "issue_families": sorted({str(row.get("family", "")) for row in rows}),
                "old_handle": old_handle,
                "new_handle": new_handle,
                "carrier_sha256": sha256_file(carrier_path),
                "carrier_content_sha256": content_hash,
                "issue_row_count": len(rows),
                "source_read": source_read,
                "target_read": target_read,
                "cite_reads": cite_reads,
            }
        )
        reconciled_issue_row_count += len(rows)

    summary = {
        "schema_version": 1,
        "issue_row_count": len(issue_rows),
        "source_group_count": len(grouped),
        "safe_reanchor_count": len(plans),
        "reconciled_issue_row_count": reconciled_issue_row_count,
        "missing_carrier_count": missing_carrier_count,
        "ambiguous_carrier_count": ambiguous_carrier_count,
        "carrier_parse_failure_count": carrier_parse_failure_count,
        "missing_old_citation_count": missing_old_citation_count,
        "ambiguous_old_citation_count": ambiguous_old_citation_count,
        "missing_current_match_count": missing_current_match_count,
        "ambiguous_current_match_count": ambiguous_current_match_count,
        "source_section_missing_count": source_section_missing_count,
        "parse_failure_count": parse_failure_count,
        "read_failure_count": read_failure_count,
        "wrapper_read_count": command_count,
        "runtime": {
            "version": args.local_version,
            "binary_sha256": sha256_file(binary),
            "wrapper_sha256": sha256_file(wrapper),
        },
    }
    private = {
        "recorded_at": datetime.now(timezone.utc).isoformat(),
        "families": sorted(selected_families),
        "manifest_sha256": sha256_file(args.private_manifest.resolve()),
        "entity_probe_sha256": sha256_file(args.entity_probe.resolve()),
        "plans": plans,
        "summary": summary,
    }
    private_output.parent.mkdir(parents=True, exist_ok=True)
    private_output.write_text(json.dumps(private, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(summary, sort_keys=True))
    clean = (
        len(issue_rows) > 0
        and len(grouped) > 0
        and len(plans) == len(grouped)
        and reconciled_issue_row_count == len(issue_rows)
        and parse_failure_count == 0
        and read_failure_count == 0
        and carrier_parse_failure_count == 0
    )
    return 0 if clean else 2


if __name__ == "__main__":
    sys.exit(main())
