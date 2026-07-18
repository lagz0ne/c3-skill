#!/usr/bin/env python3
"""Fail-closed worst-case cost admission for a bounded multi-turn eval arm."""

from __future__ import annotations

import argparse
import json
import math
from dataclasses import asdict, dataclass


@dataclass(frozen=True)
class Admission:
    base_input_tokens: int
    tool_result_tokens: int
    uncached_input_tokens: int
    cached_input_tokens: int
    output_tokens: int
    estimated_cost_usd: float
    ceiling_usd: float
    headroom_usd: float
    admitted: bool


def estimate(
    *,
    base_input_bytes: int,
    max_tool_result_bytes: int,
    tool_calls: int,
    max_output_tokens: int,
    bytes_per_token: float,
    input_per_million_usd: float,
    cached_input_per_million_usd: float,
    output_per_million_usd: float,
    ceiling_usd: float,
) -> Admission:
    if min(base_input_bytes, max_tool_result_bytes, tool_calls, max_output_tokens) < 0:
        raise ValueError("byte, call, and token limits must be non-negative")
    if bytes_per_token <= 0 or ceiling_usd < 0:
        raise ValueError("bytes_per_token must be positive and ceiling must be non-negative")
    if min(input_per_million_usd, cached_input_per_million_usd, output_per_million_usd) < 0:
        raise ValueError("prices must be non-negative")

    base_tokens = math.ceil(base_input_bytes / bytes_per_token)
    tool_tokens = math.ceil(max_tool_result_bytes / bytes_per_token)
    # Worst case: every allowed tool-result byte arrives after the first call,
    # so it is retained in every later request.
    uncached = base_tokens + tool_tokens
    cached = base_tokens * tool_calls + tool_tokens * max(tool_calls - 1, 0)
    cost = (
        uncached * input_per_million_usd
        + cached * cached_input_per_million_usd
        + max_output_tokens * output_per_million_usd
    ) / 1_000_000
    headroom = ceiling_usd - cost
    return Admission(
        base_input_tokens=base_tokens,
        tool_result_tokens=tool_tokens,
        uncached_input_tokens=uncached,
        cached_input_tokens=cached,
        output_tokens=max_output_tokens,
        estimated_cost_usd=round(cost, 9),
        ceiling_usd=ceiling_usd,
        headroom_usd=round(headroom, 9),
        admitted=headroom >= 0,
    )


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--base-input-bytes", type=int, required=True)
    parser.add_argument("--max-tool-result-bytes", type=int, required=True)
    parser.add_argument("--tool-calls", type=int, required=True)
    parser.add_argument("--max-output-tokens", type=int, required=True)
    parser.add_argument("--bytes-per-token", type=float, default=3.0)
    parser.add_argument("--input-per-million-usd", type=float, required=True)
    parser.add_argument("--cached-input-per-million-usd", type=float, required=True)
    parser.add_argument("--output-per-million-usd", type=float, required=True)
    parser.add_argument("--ceiling-usd", type=float, required=True)
    return parser.parse_args()


def main() -> int:
    result = estimate(**vars(parse_args()))
    print(json.dumps(asdict(result), sort_keys=True, separators=(",", ":")))
    return 0 if result.admitted else 2


if __name__ == "__main__":
    raise SystemExit(main())
