#!/usr/bin/env python3
"""Build a private exact-handle citation repair plan using local C3 reads only."""

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


CITATION_RE = re.compile(
    r"(?P<entity>[A-Za-z0-9][A-Za-z0-9._-]*)#n(?P<node>[0-9]+)@v(?P<version>[0-9]+):"
    r"sha256:(?P<hash>[0-9a-fA-F]{64})(?:\s+\"(?P<snippet>[^\"]*)\")?"
)
TARGET_FROM_MESSAGE_RES = (
    re.compile(r"citation to ([A-Za-z0-9][A-Za-z0-9._-]*) (?:has empty snippet|has stale node hash or snippet|cites version)"),
    re.compile(r"Evidence for .+? row ([A-Za-z0-9][A-Za-z0-9._-]*) cites version"),
    re.compile(r"Evidence for .+? has a stale cite \(no node of ([A-Za-z0-9][A-Za-z0-9._-]*) seals to that hash\)"),
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


def source_section(body: str, offset: int) -> str:
    headings = list(re.finditer(r"(?m)^##\s+(.+?)\s*$", body[:offset]))
    return headings[-1].group(1).strip() if headings else ""


def sections(body: str) -> list[str]:
    return [match.group(1).strip() for match in re.finditer(r"(?m)^##\s+(.+?)\s*$", body)]


def target_from_message(message: str) -> str:
    for pattern in TARGET_FROM_MESSAGE_RES:
        match = pattern.search(message)
        if match:
            return match.group(1)
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
        row.get("entity")
        for row in entity_probe.get("rows", [])
        if row.get("type") in {"adr", "prd", "atomic-design-change"}
        and row.get("status") not in {"done", "superseded", "implemented", "provisioned"}
        and row.get("exit_code") == 0
    }
    issues = [
        row
        for row in manifest.get("issues", [])
        if row.get("family") == args.family and row.get("entity") in mutable
    ]
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
        record = {
            "exit_code": exit_code,
            "stdout_sha256": sha256_bytes(stdout),
            "stderr_sha256": sha256_bytes(stderr),
        }
        return stdout.decode(errors="replace"), record

    plans: list[dict[str, object]] = []
    ambiguous_replacement_count = 0
    missing_replacement_count = 0
    source_section_missing_count = 0
    parse_failure_count = 0
    same_node_match_count = 0
    exact_snippet_match_count = 0
    for issue in issues:
        source_id = str(issue["entity"])
        target_id = target_from_message(str(issue.get("message", "")))
        if not target_id:
            parse_failure_count += 1
            continue
        source_text, source_record = read(source_id, ["--full"])
        source_body = toon_scalar(source_text, "body")
        old_matches = [
            match
            for match in CITATION_RE.finditer(source_body)
            if match.group("entity") == target_id
            and (args.family != "empty_citation_snippet" or not match.group("snippet"))
        ]
        if len(old_matches) != 1:
            if len(old_matches) > 1:
                ambiguous_replacement_count += 1
            else:
                missing_replacement_count += 1
            continue
        old = old_matches[0]
        section = source_section(source_body, old.start())
        if not section:
            source_section_missing_count += 1
            continue
        target_text, target_record = read(target_id, ["--full"])
        target_body = toon_scalar(target_text, "body")
        target_sections = sections(target_body)
        if not target_sections:
            missing_replacement_count += 1
            continue
        all_candidates: dict[str, re.Match[str]] = {}
        cite_records: list[dict[str, object]] = []
        for target_section in target_sections:
            cite_text, cite_record = read(target_id, ["--section", target_section, "--cite"])
            cite_records.append(cite_record)
            for candidate in CITATION_RE.finditer(cite_text):
                if candidate.group("entity") == target_id:
                    all_candidates[candidate.group(0)] = candidate
        candidates = {
            handle: candidate
            for handle, candidate in all_candidates.items()
            if candidate.group("node") == old.group("node")
        }
        match_strategy = "same_node"
        if not candidates and old.group("snippet"):
            candidates = {
                handle: candidate
                for handle, candidate in all_candidates.items()
                if candidate.group("snippet") == old.group("snippet")
            }
            match_strategy = "exact_snippet"
        if len(candidates) != 1:
            if len(candidates) > 1:
                ambiguous_replacement_count += 1
            else:
                missing_replacement_count += 1
            continue
        new_handle, new = next(iter(candidates.items()))
        if new.group("version") == old.group("version") and new.group("hash").lower() == old.group("hash").lower():
            missing_replacement_count += 1
            continue
        if match_strategy == "same_node":
            same_node_match_count += 1
        else:
            exact_snippet_match_count += 1
        plans.append(
            {
                "source_entity": source_id,
                "target_entity": target_id,
                "source_section": section,
                "old_handle": old.group(0),
                "new_handle": new_handle,
                "node_id": int(old.group("node")),
                "match_strategy": match_strategy,
                "source_read": source_record,
                "target_read": target_record,
                "cite_reads": cite_records,
            }
        )

    safe_replacement_count = len(plans)
    summary = {
        "schema_version": 1,
        "mutable_row_count": len(issues),
        "safe_replacement_count": safe_replacement_count,
        "ambiguous_replacement_count": ambiguous_replacement_count,
        "missing_replacement_count": missing_replacement_count,
        "source_section_missing_count": source_section_missing_count,
        "same_node_match_count": same_node_match_count,
        "exact_snippet_match_count": exact_snippet_match_count,
        "parse_failure_count": parse_failure_count,
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
        "family": args.family,
        "manifest_sha256": sha256_file(args.private_manifest.resolve()),
        "entity_probe_sha256": sha256_file(args.entity_probe.resolve()),
        "plans": plans,
        "summary": summary,
    }
    private_output.parent.mkdir(parents=True, exist_ok=True)
    private_output.write_text(json.dumps(private, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(summary, sort_keys=True))
    clean = (
        len(issues) > 0
        and safe_replacement_count == len(issues)
        and ambiguous_replacement_count == 0
        and missing_replacement_count == 0
        and source_section_missing_count == 0
        and parse_failure_count == 0
        and read_failure_count == 0
    )
    return 0 if clean else 2


if __name__ == "__main__":
    sys.exit(main())
