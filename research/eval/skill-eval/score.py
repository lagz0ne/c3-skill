#!/usr/bin/env python3
"""Deterministic text scorer for skill-eval answers.

Usage:
  python3 research/eval/skill-eval/score.py [--round N] CASE_ID ANSWER.md

The scorer is intentionally simple: it checks literal answer evidence against
rubric.jsonl plus the checkable universal criteria from rubric.md.
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parent
RUBRIC = ROOT / "rubric.jsonl"
FIXTURE_C3 = ROOT / "fixtures" / "acountee" / ".c3"

ID_RE = re.compile(
    r"\b(?:c3-\d+|ref-[a-z0-9-]+|rule-[a-z0-9-]+|adr-\d{8}-[a-z0-9-]+|recipe-[a-z0-9-]+)\b"
)
COMMAND_SUBCOMMAND_RE = re.compile(
    r"(?:(?:\bc3\b)|(?:c3x\.sh)|(?:\bc3x\b)).*?\b(search|list|lookup|read|graph|check|schema|canvas)\b"
)


def load_cases() -> dict[str, dict]:
    cases: dict[str, dict] = {}
    with RUBRIC.open("r", encoding="utf-8") as f:
        for line in f:
            if line.strip():
                row = json.loads(line)
                cases[row["id"]] = row
    return cases


def fixture_ids() -> set[str]:
    ids = {"c3-0"}
    if not FIXTURE_C3.exists():
        return ids

    for path in FIXTURE_C3.rglob("*.md"):
        stem = path.stem
        if stem == "README":
            parent = path.parent.name
            if re.fullmatch(r"c3-\d+-[a-z0-9-]+", parent):
                ids.add("-".join(parent.split("-")[:2]))
            continue

        if stem.startswith("c3-"):
            parts = stem.split("-")
            if len(parts) >= 2 and parts[1].isdigit():
                ids.add("-".join(parts[:2]))
        elif stem.startswith(("ref-", "rule-", "adr-", "recipe-")):
            ids.add(stem)
    return ids


def evidence_region(text: str) -> str:
    lower = text.lower()
    start = lower.find("evidence commands")
    if start == -1:
        return text
    end = lower.find("\n## answer", start)
    if end == -1:
        end = lower.find("\nanswer", start)
    return text[start:] if end == -1 else text[start:end]


def command_lines(evidence: str) -> list[str]:
    lines: list[str] = []
    for raw in evidence.splitlines():
        line = raw.strip()
        if not line or line.startswith("```"):
            continue
        if line.startswith(("$ ", "- ")):
            line = line[2:].strip()
        if line.startswith("c3()") or "c3() {" in line or line.startswith("alias c3="):
            continue
        if COMMAND_SUBCOMMAND_RE.search(line):
            lines.append(line)
    return lines


def add_point(result: dict, ok: bool, criterion: str, max_points: int = 1, points: int | None = None) -> None:
    result["max"] += max_points
    if ok:
        result["score"] += max_points if points is None else points
    else:
        result["failed"].append(criterion)


def score(case_id: str, answer_file: Path) -> dict:
    cases = load_cases()
    if case_id not in cases:
        raise SystemExit(f"unknown case id: {case_id}")

    case = cases[case_id]
    text = answer_file.read_text(encoding="utf-8")
    evidence = evidence_region(text)
    commands = command_lines(evidence)
    valid_ids = fixture_ids()
    scorer = case["scorer"]

    result = {
        "case": case_id,
        "score": 0,
        "max": 0,
        "failed": [],
    }

    # Universal checkable criteria.
    local_bound = (
        "C3X_MODE=agent bash skills/c3/bin/c3x.sh" in evidence
        or ("c3() {" in evidence and "skills/c3/bin/c3x.sh" in evidence)
    )
    bare_global = bool(re.search(r"(^|\s)c3x(\s|$)", evidence)) and "skills/c3/bin/c3x.sh" not in evidence
    add_point(result, local_bound and not bare_global, "U1 local-c3-only")

    first = commands[0] if commands else ""
    first_match = COMMAND_SUBCOMMAND_RE.search(first)
    add_point(result, bool(first_match and first_match.group(1) == "search"), "U2 search-first")

    required_ids = scorer.get("ids_include", [])
    present_ids = [entity_id for entity_id in required_ids if entity_id in text]
    if not required_ids or len(present_ids) == len(required_ids):
        u4_points = 3
    elif len(present_ids) == 0:
        u4_points = 0
    elif len(present_ids) >= max(1, len(required_ids) - 1):
        u4_points = 2
    else:
        u4_points = 1
    result["max"] += 3
    result["score"] += u4_points
    if u4_points < 3:
        missing = [entity_id for entity_id in required_ids if entity_id not in text]
        result["failed"].append(f"U4 exact-ids missing={missing}")

    seen_ids = set(ID_RE.findall(text))
    hallucinated = sorted(entity_id for entity_id in seen_ids if entity_id not in valid_ids)
    add_point(result, not hallucinated, f"U7 hallucinated-ids={hallucinated}")

    # Case-local scorer from rubric.jsonl.
    for required in scorer.get("require", []):
        add_point(result, required in text, f"require:{required}")

    any_terms = scorer.get("require_any", [])
    if any_terms:
        add_point(result, any(term in text for term in any_terms), f"require_any:{any_terms}")

    for entity_id in required_ids:
        add_point(result, entity_id in text, f"ids_include:{entity_id}")

    forbidden_terms = scorer.get("forbid", [])
    if forbidden_terms:
        found = [term for term in forbidden_terms if term in text]
        add_point(result, not found, f"forbid:{found}")

    return result


def main(argv: list[str]) -> int:
    round_no: int | None = None
    args = argv[1:]
    if len(args) >= 2 and args[0] == "--round":
        round_no = int(args[1])
        args = args[2:]
    if len(args) != 2:
        print("usage: score.py [--round N] CASE_ID ANSWER.md", file=sys.stderr)
        return 2
    result = score(args[0], Path(args[1]))
    if round_no is not None:
        result = {"round": round_no, **result}
    print(json.dumps(result, sort_keys=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
