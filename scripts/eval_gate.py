#!/usr/bin/env python3
"""Quality gate for the C3 continuous research+eval loop.

The eval harness (`agent_efficiency_eval.py`) ships a *canvas-only* summary.
This gate reads the raw JSONL records and computes a **whole-matrix quality
pass-rate** (accuracy + ADR quality >= floor + canvas quality >= floor across
every case), compares it to a committed baseline, and emits a keep/discard
verdict. Token efficiency is a secondary guardrail, never the primary signal.

Decision rule (quality-first, per the project's chosen target metric):
  * no baseline yet      -> "establish" (exit 0); writes baseline with --update-baseline
  * pass-rate held/up    -> "keep"      (exit 0); bumps baseline with --update-baseline
  * pass-rate regressed  -> "discard"   (exit 1)
  * any record >= no_go tokens -> "discard" (exit 1)  [secondary guardrail]

Every invocation appends one line to the history JSONL so progress is visible
run-over-run. This script never executes agents and never spends tokens.

Usage:
  python scripts/eval_gate.py --candidate run.jsonl --label "hyp-12 tighten adr ref"
  python scripts/eval_gate.py --candidate run.jsonl --update-baseline   # accept as new best
  python scripts/eval_gate.py --candidate run.jsonl --json              # machine-readable verdict
"""
from __future__ import annotations

import argparse
import json
import sys
import time
from pathlib import Path
from typing import Any

sys.path.insert(0, str(Path(__file__).resolve().parent))

import agent_efficiency_eval as ev

REPO_ROOT = Path(__file__).resolve().parent.parent
DEFAULT_BASELINE = REPO_ROOT / "research" / "eval" / "baseline.json"
DEFAULT_HISTORY = REPO_ROOT / "research" / "eval" / "history.jsonl"

# Candidate pass-rate must be at least this far below baseline before we call it
# a regression. Guards against float noise; pass-rates are exact ratios so this
# is effectively "no regression tolerated".
REGRESSION_EPSILON = 1e-9


def _is_adr_case(case_id: str) -> bool:
    """ADR-bearing cases carry a quality floor on the ADR artifact."""
    return "adr" in case_id and case_id not in ev.CANVAS_EXPECTATIONS


def record_passes_quality(record: dict[str, Any]) -> bool:
    """Whole-matrix quality rule, layered by case type.

    Base: the run completed cleanly and the deterministic accuracy checks all
    passed. Canvas cases additionally require the canvas quality floor; ADR
    cases additionally require the ADR quality floor.
    """
    if record.get("exit_code") != 0:
        return False
    if (record.get("accuracy_score") or 0.0) < 1.0:
        return False
    case_id = record.get("case")
    if case_id in ev.CANVAS_EXPECTATIONS:
        return bool(record.get("canvas_quality_passed"))
    if _is_adr_case(str(case_id)):
        return (record.get("adr_quality_score") or 0.0) >= ev.ADR_QUALITY_FLOOR
    return True


def compute_quality(records: list[dict[str, Any]]) -> dict[str, Any]:
    """Reduce raw eval records to the quality scorecard the gate compares on."""
    scored = [
        r
        for r in records
        if not r.get("dry_run") and not r.get("agent_unavailable")
    ]
    passes = [r for r in scored if record_passes_quality(r)]
    pass_rate = round(len(passes) / len(scored), 4) if scored else 0.0

    by_case: dict[str, dict[str, Any]] = {}
    for case_id in sorted({str(r.get("case")) for r in scored if r.get("case")}):
        case_records = [r for r in scored if r.get("case") == case_id]
        case_passes = [r for r in case_records if record_passes_quality(r)]
        by_case[case_id] = {
            "records": len(case_records),
            "passes": len(case_passes),
            "pass_rate": round(len(case_passes) / len(case_records), 4) if case_records else 0.0,
        }

    by_agent: dict[str, dict[str, Any]] = {}
    for agent_id in sorted({str(r.get("agent")) for r in scored if r.get("agent")}):
        agent_records = [r for r in scored if r.get("agent") == agent_id]
        agent_passes = [r for r in agent_records if record_passes_quality(r)]
        by_agent[agent_id] = {
            "records": len(agent_records),
            "passes": len(agent_passes),
            "pass_rate": round(len(agent_passes) / len(agent_records), 4) if agent_records else 0.0,
        }

    tokens = [int(r.get("tokens_total") or 0) for r in scored]
    eff = [int(r["effective_tokens_total"]) for r in scored if r.get("effective_tokens_total") is not None]
    # Secondary token guardrail reuses the harness's OWN pressure verdict, which
    # is computed on EFFECTIVE tokens (input - cached + output) — not the
    # cache-inflated `tokens_total`. A record breaches only when the harness
    # itself rated it no_go; that keeps the gate honest about real spend and
    # avoids a magic absolute ceiling that would reject every cached run.
    no_go_records = [r for r in scored if r.get("threshold_status") == "no_go"]
    return {
        "established": True,
        "record_count": len(scored),
        "pass_count": len(passes),
        "pass_rate": pass_rate,
        "by_case": by_case,
        "by_agent": by_agent,
        "tokens_mean": round(sum(tokens) / len(tokens), 1) if tokens else 0.0,
        "tokens_max": max(tokens) if tokens else 0,
        "effective_tokens_mean": round(sum(eff) / len(eff), 1) if eff else None,
        "token_no_go_count": len(no_go_records),
        "adr_quality_floor": ev.ADR_QUALITY_FLOOR,
        "canvas_quality_floor": ev.CANVAS_QUALITY_FLOOR,
    }


def load_baseline(path: Path) -> dict[str, Any]:
    if not path.exists():
        return {"established": False}
    data = json.loads(path.read_text())
    if not isinstance(data, dict):
        raise ValueError(f"{path}: baseline must be a JSON object")
    return data


def decide(candidate: dict[str, Any], baseline: dict[str, Any]) -> dict[str, Any]:
    """keep / discard / establish, with the reasons that drove it."""
    reasons: list[str] = []
    token_breach = candidate["token_no_go_count"] > 0
    if token_breach:
        reasons.append(
            f"{candidate['token_no_go_count']} record(s) rated no_go on effective "
            f"tokens by the harness"
        )

    if not baseline.get("established"):
        decision = "establish" if not token_breach else "discard"
        if token_breach:
            reasons.append("baseline not established; token guardrail still binds")
        else:
            reasons.append("no baseline yet; establishing from this run")
        return {
            "decision": decision,
            "reasons": reasons,
            "candidate_pass_rate": candidate["pass_rate"],
            "baseline_pass_rate": None,
            "pass_rate_delta": None,
            "token_guardrail_breached": token_breach,
        }

    base_rate = float(baseline.get("pass_rate", 0.0))
    delta = round(candidate["pass_rate"] - base_rate, 4)
    regressed = candidate["pass_rate"] < base_rate - REGRESSION_EPSILON

    if regressed:
        reasons.append(f"quality pass-rate regressed {base_rate:.2%} -> {candidate['pass_rate']:.2%}")
        decision = "discard"
    elif token_breach:
        decision = "discard"
    else:
        if delta > 0:
            reasons.append(f"quality pass-rate improved {base_rate:.2%} -> {candidate['pass_rate']:.2%}")
        else:
            reasons.append(f"quality pass-rate held at {candidate['pass_rate']:.2%}")
        decision = "keep"

    # Surface per-case regressions even when overall holds — useful research signal.
    base_cases = baseline.get("by_case", {})
    for case_id, cand in candidate["by_case"].items():
        prev = base_cases.get(case_id)
        if prev and cand["pass_rate"] < prev.get("pass_rate", 0.0):
            reasons.append(
                f"case {case_id} dropped {prev['pass_rate']:.2%} -> {cand['pass_rate']:.2%}"
            )

    return {
        "decision": decision,
        "reasons": reasons,
        "candidate_pass_rate": candidate["pass_rate"],
        "baseline_pass_rate": base_rate,
        "pass_rate_delta": delta,
        "token_guardrail_breached": token_breach,
    }


def append_history(path: Path, entry: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a") as fh:
        fh.write(json.dumps(entry, sort_keys=True) + "\n")


def write_baseline(path: Path, candidate: dict[str, Any], label: str, timestamp: float) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    payload = dict(candidate)
    payload["label"] = label
    payload["timestamp"] = timestamp
    path.write_text(json.dumps(payload, indent=2, sort_keys=True) + "\n")


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Quality gate for the C3 research+eval loop")
    parser.add_argument("--candidate", required=True, help="candidate eval JSONL (from --run)")
    parser.add_argument("--baseline", default=str(DEFAULT_BASELINE), help="committed baseline JSON")
    parser.add_argument("--history", default=str(DEFAULT_HISTORY), help="append-only verdict log")
    parser.add_argument("--label", default="", help="short description of the change under test")
    parser.add_argument("--update-baseline", action="store_true", help="on keep/establish, write candidate as new baseline")
    parser.add_argument("--no-history", action="store_true", help="do not append to the history log")
    parser.add_argument("--timestamp", type=float, help="override timestamp (defaults to wall clock)")
    parser.add_argument("--json", action="store_true", help="print machine-readable verdict to stdout")
    args = parser.parse_args(argv)

    candidate_path = Path(args.candidate)
    if not candidate_path.exists():
        parser.error(f"candidate not found: {candidate_path}")

    records = ev.load_jsonl_records(candidate_path)
    candidate = compute_quality(records)
    baseline = load_baseline(Path(args.baseline))
    verdict = decide(candidate, baseline)
    timestamp = args.timestamp if args.timestamp is not None else time.time()

    accepted = verdict["decision"] in ("keep", "establish")
    if accepted and args.update_baseline:
        write_baseline(Path(args.baseline), candidate, args.label, timestamp)

    history_entry = {
        "timestamp": timestamp,
        "label": args.label,
        "decision": verdict["decision"],
        "candidate_pass_rate": verdict["candidate_pass_rate"],
        "baseline_pass_rate": verdict["baseline_pass_rate"],
        "pass_rate_delta": verdict["pass_rate_delta"],
        "tokens_mean": candidate["tokens_mean"],
        "tokens_max": candidate["tokens_max"],
        "token_guardrail_breached": verdict["token_guardrail_breached"],
        "record_count": candidate["record_count"],
        "baseline_updated": bool(accepted and args.update_baseline),
    }
    if not args.no_history:
        append_history(Path(args.history), history_entry)

    if args.json:
        print(json.dumps({"verdict": verdict, "candidate": candidate, "history": history_entry}, indent=2, sort_keys=True))
    else:
        rate = candidate["pass_rate"]
        base = verdict["baseline_pass_rate"]
        base_str = f"{base:.2%}" if base is not None else "n/a"
        print(f"[{verdict['decision'].upper()}] quality {rate:.2%} (baseline {base_str}) "
              f"over {candidate['record_count']} record(s)")
        for reason in verdict["reasons"]:
            print(f"  - {reason}")
        if accepted and args.update_baseline:
            print(f"  -> baseline updated: {args.baseline}")

    # exit 0 on keep/establish, 1 on discard — drives the loop's revert decision.
    return 0 if accepted else 1


if __name__ == "__main__":
    raise SystemExit(main())
