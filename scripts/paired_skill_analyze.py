#!/usr/bin/env python3
"""Analyze generic scored paired-eval rows with a deterministic bootstrap."""

from __future__ import annotations

import argparse
import json
import math
import random
import statistics
import sys
from collections import defaultdict
from pathlib import Path
from typing import Any

ROOT = Path(__file__).resolve().parents[1]
if str(ROOT) not in sys.path:
    sys.path.insert(0, str(ROOT))
from scripts import paired_skill_eval as paired


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--input", required=True, type=Path)
    parser.add_argument("--study-id", required=True)
    parser.add_argument("--seed", type=int, default=1)
    parser.add_argument("--bootstrap", type=int, default=10000)
    return parser.parse_args()


def mean(values: list[float]) -> float:
    return round(statistics.mean(values), 8)


def bootstrap_interval(values: list[float], *, seed: int, iterations: int) -> dict[str, Any]:
    if not values:
        raise ValueError("bootstrap requires at least one paired value")
    if iterations < 100:
        raise ValueError("bootstrap iterations must be at least 100")
    rng = random.Random(seed)
    samples = [
        statistics.mean(rng.choice(values) for _ in values)
        for _ in range(iterations)
    ]
    samples.sort()

    def percentile(p: float) -> float:
        position = (len(samples) - 1) * p
        lower = int(position)
        upper = min(lower + 1, len(samples) - 1)
        fraction = position - lower
        return samples[lower] + (samples[upper] - samples[lower]) * fraction

    low = percentile(0.025)
    high = percentile(0.975)
    return {
        "seed": seed,
        "iterations": iterations,
        "low": round(low, 8),
        "high": round(high, 8),
        "half_width": round((high - low) / 2, 8),
    }


def load_rows(path: Path, study_id: str) -> list[dict[str, Any]]:
    if not path.is_file():
        raise ValueError("input result ledger is missing")
    rows: list[dict[str, Any]] = []
    for line_number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        if not line.strip():
            continue
        try:
            row = json.loads(line)
        except json.JSONDecodeError as exc:
            raise ValueError(f"line {line_number}: invalid JSON") from exc
        paired.validate_generic_result(row)
        if row["study_id"] != study_id:
            raise ValueError("input contains a different study id")
        if row["status"] != "scored":
            raise ValueError("all rows must be scored before analysis")
        if row["protocol_deviation_codes"]:
            raise ValueError("analysis refuses rows with protocol deviations")
        score_fields = (
            "quality_score", "correctness_score", "trace_completeness_score",
            "reasoning_depth_score", "grounding_score", "no_hallucination_score",
            "change_usefulness_score",
        )
        for field in score_fields:
            value = row[field]
            if isinstance(value, bool) or not isinstance(value, (int, float)) or not math.isfinite(float(value)) or not 1 <= float(value) <= 5:
                raise ValueError(f"scored rows need finite {field} values from 1 to 5")
        if not isinstance(row["passed"], bool):
            raise ValueError("scored rows need a boolean passed value")
        if (row["independent_review_count"] or 0) < 1 or (row["deterministic_evidence_count"] or 0) < 1:
            raise ValueError("all scored rows need one qualifying review and evidence receipts")
        c3_counts = (
            "c3_invocation_count",
            "c3_route_command_count",
            "c3_impact_command_count",
            "c3_evidence_command_count",
            "c3_success_count",
            "c3_route_success_count",
            "c3_impact_success_count",
            "c3_evidence_success_count",
        )
        if row["condition"] == "with_c3" and not all(row[field] > 0 for field in c3_counts):
            raise ValueError("with_c3 rows need successful route, impact, and evidence uptake")
        if row["condition"] == "without_c3" and any(row[field] > 0 for field in c3_counts):
            raise ValueError("without_c3 rows cannot contain C3 uptake")
        for success, attempts in (
            ("c3_success_count", "c3_invocation_count"),
            ("c3_route_success_count", "c3_route_command_count"),
            ("c3_impact_success_count", "c3_impact_command_count"),
            ("c3_evidence_success_count", "c3_evidence_command_count"),
        ):
            if row[success] > row[attempts]:
                raise ValueError(f"{success} cannot exceed {attempts}")
        for field in ("total_tokens", "cost_usd", "total_cost_usd", "elapsed_ms", "model_turns", "tool_calls"):
            value = row[field]
            if isinstance(value, bool) or not isinstance(value, (int, float)) or not math.isfinite(float(value)) or float(value) < 0:
                raise ValueError(f"scored rows need finite non-negative {field} values")
        rows.append(row)
    if not rows:
        raise ValueError("input result ledger is empty")
    return rows


def pair_rows(rows: list[dict[str, Any]]) -> list[tuple[dict[str, Any], dict[str, Any]]]:
    grouped: dict[tuple[str, int], dict[str, dict[str, Any]]] = defaultdict(dict)
    for row in rows:
        key = (row["case_id"], row["trial"])
        if row["condition"] in grouped[key]:
            raise ValueError("duplicate arm in pair")
        grouped[key][row["condition"]] = row
    pairs: list[tuple[dict[str, Any], dict[str, Any]]] = []
    for key in sorted(grouped):
        arms = grouped[key]
        if set(arms) != {"with_c3", "without_c3"}:
            raise ValueError("every case/trial must have both paired arms")
        treatment, control = arms["with_c3"], arms["without_c3"]
        if treatment["case_family"] != control["case_family"]:
            raise ValueError("paired arms differ in case_family")
        for field in (
            "agent", "model", "reasoning_effort", "price_version",
            "baseline_instruction_sha256", "max_tool_calls", "max_tool_result_bytes",
            "max_output_bytes", "max_answer_words", "runner_version", "cost_policy",
            "skill_sha256", "treatment_instruction_sha256", "treatment_runtime_layer",
        ):
            if treatment[field] != control[field]:
                if field in {"skill_sha256", "treatment_instruction_sha256", "treatment_runtime_layer"}:
                    continue
                raise ValueError(f"paired arms differ in {field}")
        pairs.append((treatment, control))
    return pairs


def enforce_global_parity(pairs: list[tuple[dict[str, Any], dict[str, Any]]]) -> None:
    common_fields = (
        "agent", "model", "reasoning_effort", "price_version",
        "baseline_instruction_sha256", "max_tool_calls", "max_tool_result_bytes",
        "max_output_bytes", "max_answer_words", "runner_version", "cost_policy",
    )
    signatures = {
        tuple(row[field] for field in common_fields)
        for pair in pairs for row in pair
    }
    if len(signatures) != 1:
        raise ValueError("paired study rows differ in frozen run settings")
    treatment_fields = ("skill_sha256", "treatment_instruction_sha256", "treatment_runtime_layer")
    treatment_signatures = {tuple(treatment[field] for field in treatment_fields) for treatment, _ in pairs}
    if len(treatment_signatures) != 1:
        raise ValueError("treatment rows differ in frozen C3 settings")


def efficiency(rows: list[dict[str, Any]], condition: str) -> dict[str, float]:
    selected = [row for row in rows if row["condition"] == condition]
    fields = ("total_tokens", "cost_usd", "total_cost_usd", "elapsed_ms", "model_turns", "tool_calls")
    result: dict[str, float] = {}
    for field in fields:
        values = [row[field] for row in selected]
        if not all(isinstance(value, (int, float)) for value in values):
            raise ValueError(f"missing numeric {field} for {condition}")
        result[field] = mean([float(value) for value in values])
    return result


def main() -> int:
    args = parse_args()
    try:
        rows = load_rows(args.input, args.study_id)
        pairs = pair_rows(rows)
        if len(pairs) < 2:
            raise ValueError("at least two complete pairs are required")
        enforce_global_parity(pairs)
        unique_case_count = len({treatment["case_id"] for treatment, _ in pairs})
        observed_families = sorted({treatment["case_family"] for treatment, _ in pairs})
        required_families = sorted(paired.CASE_FAMILIES)
        deltas = [float(treatment["quality_score"]) - float(control["quality_score"]) for treatment, control in pairs]
        family_deltas: dict[str, list[float]] = defaultdict(list)
        for (treatment, control), delta in zip(pairs, deltas):
            family_deltas[treatment["case_family"]].append(delta)
        bootstrap = bootstrap_interval(deltas, seed=args.seed, iterations=args.bootstrap)
        minimum_cases_met = unique_case_count >= 20
        families_met = observed_families == required_families
        ci_target = 0.25
        ci_target_met = bootstrap["half_width"] <= ci_target
        quality_claim_eligible = minimum_cases_met and families_met and ci_target_met
        if not minimum_cases_met:
            validity_status = "below_confirmatory_minimum"
        elif not families_met:
            validity_status = "missing_required_case_family"
        elif not ci_target_met:
            validity_status = "ci_above_target"
        else:
            validity_status = "eligible_for_confirmatory_quality_gate"
        report = {
            "study_id": args.study_id,
            "pair_count": len(pairs),
            "arm_count": len(rows),
            "quality": {
                "mean_paired_delta": mean(deltas),
                "bootstrap": bootstrap,
                "family_effects": {
                    family: {"pair_count": len(values), "mean_paired_delta": mean(values)}
                    for family, values in sorted(family_deltas.items())
                },
                "passed_rate": {
                    condition: mean([1.0 if row["passed"] else 0.0 for row in rows if row["condition"] == condition])
                    for condition in ("with_c3", "without_c3")
                },
            },
            "study_validity": {
                "status": validity_status,
                "unique_case_count": unique_case_count,
                "minimum_cases": 20,
                "required_case_families": required_families,
                "observed_case_families": observed_families,
                "ci_half_width_target": ci_target,
                "ci_half_width_target_met": ci_target_met,
                "quality_claim_eligible": quality_claim_eligible,
            },
            "efficiency": {
                condition: efficiency(rows, condition)
                for condition in ("with_c3", "without_c3")
            },
            "generic_only": True,
        }
    except (OSError, ValueError, KeyError, TypeError, json.JSONDecodeError) as exc:
        print(str(exc), file=sys.stderr)
        return 2
    print(json.dumps(report, sort_keys=True, separators=(",", ":")))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
