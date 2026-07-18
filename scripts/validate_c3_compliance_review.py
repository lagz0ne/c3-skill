#!/usr/bin/env python3
"""Validate a private semantic compliance review against its exact issue plan."""

from __future__ import annotations

import argparse
import json
import re
import sys
from collections import Counter
from pathlib import Path
from typing import Sequence


ALLOWED_ACTIONS = {"comply", "review", "update-ref", "create-ref"}


def parser() -> argparse.ArgumentParser:
    result = argparse.ArgumentParser(description=__doc__)
    result.add_argument("--private-plan", type=Path, required=True)
    result.add_argument("--response", type=Path, required=True)
    result.add_argument("--private-output", type=Path, required=True)
    result.add_argument("--max-cost-usd", type=float, required=True)
    result.add_argument("--max-tool-calls", type=int, required=True)
    result.add_argument("--max-output-bytes", type=int, required=True)
    return result


def parse_result(value: object) -> dict[str, object] | None:
    if isinstance(value, dict):
        return value
    if not isinstance(value, str):
        return None
    text = value.strip()
    fenced = re.fullmatch(r"```(?:json)?\s*(.*?)\s*```", text, re.DOTALL | re.IGNORECASE)
    if fenced:
        text = fenced.group(1)
    try:
        parsed = json.loads(text)
    except json.JSONDecodeError:
        return None
    return parsed if isinstance(parsed, dict) else None


def token_usage(envelope: dict[str, object]) -> int:
    model_usage = envelope.get("modelUsage")
    if isinstance(model_usage, dict) and model_usage:
        keys = ("inputTokens", "outputTokens", "cacheReadInputTokens", "cacheCreationInputTokens")
        return sum(
            int(record.get(key, 0) or 0)
            for record in model_usage.values()
            if isinstance(record, dict)
            for key in keys
        )
    usage = envelope.get("usage")
    if not isinstance(usage, dict):
        return 0
    keys = ("input_tokens", "output_tokens", "cache_read_input_tokens", "cache_creation_input_tokens")
    return sum(int(usage.get(key, 0) or 0) for key in keys)


def main(argv: Sequence[str] | None = None) -> int:
    args = parser().parse_args(argv)
    plan = json.loads(args.private_plan.read_text(encoding="utf-8"))
    envelope = json.loads(args.response.read_text(encoding="utf-8"))
    review = parse_result(envelope.get("result"))

    selected_pairs = [(str(row.get("source", "")), str(row.get("target", ""))) for row in plan.get("rows", [])]
    selected_counter = Counter(selected_pairs)
    handles_by_target = {
        str(record.get("entity", "")): set(str(value) for value in record.get("citation_handles", []))
        for record in plan.get("target_records", [])
    }

    proposed = review.get("rows", []) if isinstance(review, dict) and isinstance(review.get("rows"), list) else []
    unresolved = (
        review.get("unresolved", [])
        if isinstance(review, dict) and isinstance(review.get("unresolved"), list)
        else []
    )
    proposed_pairs = [(str(row.get("source", "")), str(row.get("target", ""))) for row in proposed if isinstance(row, dict)]
    unresolved_pairs = [
        (str(row.get("source", "")), str(row.get("target", ""))) for row in unresolved if isinstance(row, dict)
    ]
    delivered_counter = Counter(proposed_pairs + unresolved_pairs)
    missing_row_count = sum((selected_counter - delivered_counter).values())
    extra_row_count = sum((delivered_counter - selected_counter).values())
    duplicate_row_count = sum(count - 1 for count in delivered_counter.values() if count > 1)

    invalid_evidence_count = 0
    invalid_action_count = 0
    invalid_field_count = 0
    for row in proposed:
        if not isinstance(row, dict):
            invalid_field_count += 1
            continue
        target = str(row.get("target", ""))
        if str(row.get("evidence", "")) not in handles_by_target.get(target, set()):
            invalid_evidence_count += 1
        action = str(row.get("action", "")).strip()
        if action not in ALLOWED_ACTIONS and not action.startswith("N.A - "):
            invalid_action_count += 1
        confidence = row.get("confidence")
        source_evidence = row.get("source_evidence")
        if (
            not str(row.get("why_required", "")).strip()
            or not isinstance(source_evidence, list)
            or not source_evidence
            or not isinstance(confidence, (int, float))
            or isinstance(confidence, bool)
            or not 0 <= float(confidence) <= 1
        ):
            invalid_field_count += 1

    total_cost_usd = float(envelope.get("total_cost_usd", 0) or 0)
    duration_seconds = float(envelope.get("duration_ms", 0) or 0) / 1000
    num_turns = int(envelope.get("num_turns", 0) or 0)
    tool_call_upper_bound = max(0, num_turns - 1)
    output_bytes = args.response.stat().st_size
    permission_denials = envelope.get("permission_denials", [])
    permission_denial_count = len(permission_denials) if isinstance(permission_denials, list) else 1
    unsupported_count = int(review.get("unsupported_count", 0) or 0) if isinstance(review, dict) else 0
    complete = bool(
        review
        and not envelope.get("is_error", False)
        and review.get("review_status") == "complete"
        and len(selected_pairs) > 0
        and len(proposed) == len(selected_pairs)
        and len(unresolved) == 0
        and missing_row_count == 0
        and extra_row_count == 0
        and duplicate_row_count == 0
        and invalid_evidence_count == 0
        and invalid_action_count == 0
        and invalid_field_count == 0
        and unsupported_count == 0
        and total_cost_usd <= args.max_cost_usd
        and tool_call_upper_bound <= args.max_tool_calls
        and output_bytes < args.max_output_bytes
        and permission_denial_count == 0
    )
    summary = {
        "schema_version": 1,
        "complete": complete,
        "selected_row_count": len(selected_pairs),
        "proposed_row_count": len(proposed),
        "unresolved_row_count": len(unresolved),
        "missing_row_count": missing_row_count,
        "extra_row_count": extra_row_count,
        "duplicate_row_count": duplicate_row_count,
        "invalid_evidence_count": invalid_evidence_count,
        "invalid_action_count": invalid_action_count,
        "invalid_field_count": invalid_field_count,
        "unsupported_count": unsupported_count,
        "total_cost_usd": total_cost_usd,
        "duration_seconds": duration_seconds,
        "total_tokens": token_usage(envelope),
        "tool_call_upper_bound": tool_call_upper_bound,
        "output_bytes": output_bytes,
        "permission_denial_count": permission_denial_count,
        "provider_error": bool(envelope.get("is_error", False)),
    }
    private = {"summary": summary, "review": review, "envelope_type": envelope.get("type")}
    args.private_output.parent.mkdir(parents=True, exist_ok=True)
    args.private_output.write_text(json.dumps(private, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(summary, sort_keys=True))
    return 0 if complete else 2


if __name__ == "__main__":
    sys.exit(main())
