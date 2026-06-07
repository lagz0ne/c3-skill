#!/usr/bin/env python3
"""Rubric eval: pose grounded Q&A about the CURRENT .c3 docs to an agent and
score answers mechanically against ground truth.

Distinct from agent_efficiency_eval.py (which scores canvas *authoring*). Here
the ground truth is this repo's real architecture (research/eval/rubric.jsonl),
and we measure whether an agent using the C3 skill answers correctly. Scoring is
deterministic — substring/id-set rules per item, no model-graded judgement.

Dry-run by default (plans the matrix, spends no tokens). Pass --run to execute
agents. JSONL output + summary + optional pass-rate gate.

Usage:
  python scripts/rubric_eval.py --dry-run
  python scripts/rubric_eval.py --run --agent claude --output runs/rubric.jsonl --summary s.json
  python scripts/rubric_eval.py --run --require-pass-rate 0.9
"""
from __future__ import annotations

import argparse
import json
import os
import re
import shlex
import subprocess
import sys
import time
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parent.parent
DEFAULT_RUBRIC = REPO_ROOT / "research" / "eval" / "rubric.jsonl"

# Entity id tokens. \d+ is greedy so "c3-114" extracts whole — "c3-1" is only
# matched when it is genuinely c3-1 (no trailing digit), which is what lets the
# id-set scorers stay reliable.
ID_RE = re.compile(
    r"\b(?:c3-\d+(?:\.\d+)?|ref-[a-z0-9-]+|rule-[a-z0-9-]+|adr-[0-9]{8}-[a-z0-9-]+|recipe-[a-z0-9-]+)\b"
)

_SCORER_KEYS = {"require", "require_any", "forbid", "ids_include", "id_set", "min_ids_from"}


def extract_ids(text: str) -> set[str]:
    return set(ID_RE.findall(text or ""))


def scorer_is_valid(scorer: dict[str, Any]) -> bool:
    if not isinstance(scorer, dict) or not scorer:
        return False
    keys = set(scorer.keys()) - {"id_set_exclude_subject"}
    return bool(keys) and keys.issubset(_SCORER_KEYS | {"id_set_exclude_subject"})


def score_answer(scorer: dict[str, Any], answer: str) -> tuple[bool, list[str]]:
    """Deterministically score a free-text answer against a scorer spec.

    All present conditions must hold. Substring checks are case-insensitive on
    whitespace-normalized text; id checks use boundary-aware id extraction.
    """
    reasons: list[str] = []
    norm = " ".join((answer or "").split()).lower()
    ids = extract_ids(answer)

    for need in scorer.get("require", []):
        if need.lower() not in norm:
            reasons.append(f"missing required: {need!r}")
    if "require_any" in scorer:
        opts = scorer["require_any"]
        if not any(o.lower() in norm for o in opts):
            reasons.append(f"none of require_any present: {opts}")
    for bad in scorer.get("forbid", []):
        if bad.lower() in norm:
            reasons.append(f"forbidden present: {bad!r}")
    for need_id in scorer.get("ids_include", []):
        if need_id not in ids:
            reasons.append(f"missing id: {need_id}")
    if "id_set" in scorer:
        subject = scorer.get("id_set_exclude_subject")
        expected = set(scorer["id_set"]) - ({subject} if subject else set())
        got = ids - ({subject} if subject else set())
        if got != expected:
            reasons.append(f"id_set mismatch: got {sorted(got)} want {sorted(expected)}")
    if "min_ids_from" in scorer:
        spec = scorer["min_ids_from"]
        hits = ids & set(spec["set"])
        if len(hits) < spec["min"]:
            reasons.append(f"need >={spec['min']} ids from set, got {len(hits)}")

    return (not reasons), reasons


def load_rubric(path: Path) -> list[dict[str, Any]]:
    items: list[dict[str, Any]] = []
    for n, line in enumerate(Path(path).read_text().splitlines(), 1):
        if not line.strip():
            continue
        try:
            item = json.loads(line)
        except json.JSONDecodeError as err:
            raise ValueError(f"{path}:{n}: invalid JSONL: {err}") from err
        items.append(item)
    return items


def default_agents() -> dict[str, list[str]]:
    return {
        "claude": shlex.split(os.environ.get("C3_EVAL_CLAUDE_CMD", "claude -p --dangerously-skip-permissions {prompt}")),
        "codex": shlex.split(os.environ.get(
            "C3_EVAL_CODEX_CMD",
            "codex --ask-for-approval never exec --json --sandbox workspace-write --skip-git-repo-check --ignore-user-config {prompt}",
        )),
    }


PROMPT_TEMPLATE = (
    "Follow the local C3 skill at skills/c3/SKILL.md, and use ONLY the local C3 CLI in this repo via "
    "`C3X_MODE=agent bash skills/c3/bin/c3x.sh <cmd>` (never bare c3x or a global skill). "
    "Answer this question about the CURRENT .c3 architecture concisely, naming exact entity ids "
    "(c3-NNN, ref-*, rule-*, adr-*). Question: {question}"
)


def run_agent(command_template: list[str], prompt: str) -> tuple[str, int]:
    cmd = [part.replace("{prompt}", prompt) for part in command_template]
    proc = subprocess.run(cmd, cwd=str(REPO_ROOT), capture_output=True, text=True)
    return proc.stdout + "\n" + proc.stderr, proc.returncode


def summarize(records: list[dict[str, Any]]) -> dict[str, Any]:
    scored = [r for r in records if not r.get("dry_run") and not r.get("agent_unavailable")]
    unavailable = [r for r in records if r.get("agent_unavailable")]
    passes = [r for r in scored if r.get("passed")]
    by_category: dict[str, dict[str, Any]] = {}
    for cat in sorted({str(r.get("category")) for r in scored}):
        crs = [r for r in scored if r.get("category") == cat]
        cps = [r for r in crs if r.get("passed")]
        by_category[cat] = {"records": len(crs), "passes": len(cps),
                            "pass_rate": round(len(cps) / len(crs), 4) if crs else 0.0}
    by_agent: dict[str, dict[str, Any]] = {}
    for ag in sorted({str(r.get("agent")) for r in scored if r.get("agent")}):
        ars = [r for r in scored if r.get("agent") == ag]
        aps = [r for r in ars if r.get("passed")]
        by_agent[ag] = {"records": len(ars), "passes": len(aps),
                        "pass_rate": round(len(aps) / len(ars), 4) if ars else 0.0}
    return {
        "record_count": len(scored),
        "unavailable_count": len(unavailable),
        "pass_count": len(passes),
        "pass_rate": round(len(passes) / len(scored), 4) if scored else 0.0,
        "by_category": by_category,
        "by_agent": by_agent,
    }


def main(argv: list[str] | None = None) -> int:
    p = argparse.ArgumentParser(description="Rubric Q&A eval against current .c3 docs")
    p.add_argument("--rubric", default=str(DEFAULT_RUBRIC))
    p.add_argument("--run", action="store_true", help="execute agents (spends tokens)")
    p.add_argument("--dry-run", action="store_true", help="plan only (default)")
    p.add_argument("--agent", choices=["claude", "codex"], action="append")
    p.add_argument("--output", default="rubric-results.jsonl")
    p.add_argument("--summary")
    p.add_argument("--require-pass-rate", type=float, help="exit nonzero below this overall pass-rate")
    p.add_argument("--timestamp", type=float)
    args = p.parse_args(argv)

    items = load_rubric(Path(args.rubric))
    agents = default_agents()
    selected = args.agent or list(agents)
    dry_run = not args.run
    ts = args.timestamp if args.timestamp is not None else time.time()

    out = Path(args.output)
    out.parent.mkdir(parents=True, exist_ok=True)
    records: list[dict[str, Any]] = []
    with out.open("w") as fh:
        for agent_id in selected:
            for item in items:
                rec: dict[str, Any] = {
                    "agent": agent_id, "id": item["id"], "category": item["category"],
                    "question": item["question"], "dry_run": dry_run, "timestamp": ts,
                }
                if not dry_run:
                    prompt = PROMPT_TEMPLATE.format(question=item["question"])
                    answer, code = run_agent(agents[agent_id], prompt)
                    passed, reasons = score_answer(item["scorer"], answer)
                    # A nonzero exit means the agent CLI failed to run (e.g. auth
                    # 401) — classify as unavailable and exclude from the pass-rate
                    # denominator rather than scoring its error text as a failure.
                    unavailable = code != 0
                    rec.update({"passed": passed, "exit_code": code, "agent_unavailable": unavailable,
                                "score_reasons": reasons, "answer_chars": len(answer),
                                "answer_snippet": " ".join((answer or "").split())[:500]})
                records.append(rec)
                fh.write(json.dumps(rec, sort_keys=True) + "\n")

    summary = summarize(records)
    if args.summary:
        Path(args.summary).write_text(json.dumps(summary, indent=2, sort_keys=True) + "\n")
    if dry_run:
        print(f"dry-run: planned {len(records)} rubric item(s) across {len(selected)} agent(s); pass --run to execute")
    else:
        print(f"rubric eval: {summary['pass_count']}/{summary['record_count']} passed ({summary['pass_rate']:.0%})")
        for cat, c in summary["by_category"].items():
            print(f"  {cat}: {c['passes']}/{c['records']} ({c['pass_rate']:.0%})")

    if args.require_pass_rate is not None and not dry_run:
        if summary["pass_rate"] < args.require_pass_rate:
            print(f"FAIL: pass-rate {summary['pass_rate']:.0%} < required {args.require_pass_rate:.0%}")
            return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
