#!/usr/bin/env python3
"""Condition-canonical scoring for independently blinded paired judgments."""

from __future__ import annotations

from collections import defaultdict
from typing import Any, Mapping, Sequence


CONDITIONS = {"with_c3", "without_c3"}
ALIASES = {"A", "B"}
VALID_LABELS = {"gold": set("SCN"), "forbidden": set("SN")}


class ScoreInputError(ValueError):
    """Raised when scoring input cannot be resolved without guessing."""


def _mapping_index(
    rows: Sequence[Mapping[str, Any]],
) -> dict[tuple[str, str], dict[str, str]]:
    output: dict[tuple[str, str], dict[str, str]] = {}
    for row in rows:
        judge_id = str(row.get("judge_id", ""))
        case_id = str(row.get("case_id", ""))
        key = (judge_id, case_id)
        if not judge_id or not case_id:
            raise ScoreInputError("mapping requires stable judge_id and case_id")
        if key in output:
            raise ScoreInputError(f"duplicate mapping for {judge_id}/{case_id}")
        aliases = row.get("aliases")
        if not isinstance(aliases, Mapping):
            raise ScoreInputError(f"mapping aliases missing for {judge_id}/{case_id}")
        normalized = {str(alias): str(condition) for alias, condition in aliases.items()}
        if set(normalized) != ALIASES or set(normalized.values()) != CONDITIONS:
            raise ScoreInputError(
                f"alias mapping must be a bijection for {judge_id}/{case_id}"
            )
        output[key] = normalized
    return output


def _validate_counts(
    proposition_counts: Mapping[str, Mapping[str, int]],
) -> dict[str, dict[str, int]]:
    if not proposition_counts:
        raise ScoreInputError("proposition counts are required")
    output: dict[str, dict[str, int]] = {}
    for case_id, counts in proposition_counts.items():
        stable_case_id = str(case_id)
        if not stable_case_id:
            raise ScoreInputError("stable case_id is required")
        if set(counts) != {"gold", "forbidden"}:
            raise ScoreInputError(f"invalid proposition counts for {stable_case_id}")
        normalized = {kind: int(counts[kind]) for kind in ("gold", "forbidden")}
        if normalized["gold"] <= 0 or normalized["forbidden"] < 0:
            raise ScoreInputError(f"invalid proposition counts for {stable_case_id}")
        output[stable_case_id] = normalized
    return output


def _canonical_labels(
    judgments: Sequence[Mapping[str, Any]],
    mappings: dict[tuple[str, str], dict[str, str]],
    counts: dict[str, dict[str, int]],
) -> tuple[dict[tuple[str, str, str, int], dict[str, str]], set[str]]:
    seen: set[tuple[str, str]] = set()
    cells: dict[tuple[str, str, str, int], dict[str, str]] = defaultdict(dict)
    judges: set[str] = set()

    for row in judgments:
        judge_id = str(row.get("judge_id", ""))
        case_id = str(row.get("case_id", ""))
        key = (judge_id, case_id)
        if key in seen:
            raise ScoreInputError(f"duplicate judgment for {judge_id}/{case_id}")
        seen.add(key)
        judges.add(judge_id)
        if key not in mappings:
            raise ScoreInputError(f"judgment has no mapping for {judge_id}/{case_id}")
        if case_id not in counts:
            raise ScoreInputError(f"judgment has unknown stable case_id {case_id}")

        answers = row.get("answers")
        if not isinstance(answers, Mapping) or set(answers) != ALIASES:
            raise ScoreInputError(f"judgment aliases must be A and B for {judge_id}/{case_id}")

        for alias in sorted(ALIASES):
            answer = answers[alias]
            if not isinstance(answer, Mapping) or set(answer) != {"gold", "forbidden"}:
                raise ScoreInputError(
                    f"judgment cells incomplete for {judge_id}/{case_id}/{alias}"
                )
            condition = mappings[key][alias]
            for kind in ("gold", "forbidden"):
                labels = answer[kind]
                if not isinstance(labels, str):
                    raise ScoreInputError(
                        f"{kind} labels must be a string for {judge_id}/{case_id}/{alias}"
                    )
                expected = counts[case_id][kind]
                if len(labels) != expected:
                    raise ScoreInputError(
                        f"{kind} count mismatch for {judge_id}/{case_id}/{alias}: "
                        f"expected {expected}, got {len(labels)}"
                    )
                for index, label in enumerate(labels, 1):
                    if label not in VALID_LABELS[kind]:
                        raise ScoreInputError(
                            f"invalid {kind} label {label!r} for "
                            f"{judge_id}/{case_id}/{alias}"
                        )
                    cell = (case_id, condition, kind, index)
                    if judge_id in cells[cell]:
                        raise ScoreInputError(f"duplicate canonical cell {cell}/{judge_id}")
                    cells[cell][judge_id] = label

    if set(mappings) != seen:
        missing = sorted(set(mappings) - seen)
        raise ScoreInputError(f"missing judgment for mapping {missing[0]}")
    if len(judges) < 2:
        raise ScoreInputError("at least two independent judges are required")

    expected_keys = {
        (judge_id, case_id)
        for judge_id in judges
        for case_id in counts
    }
    if seen != expected_keys:
        missing = sorted(expected_keys - seen)
        extra = sorted(seen - expected_keys)
        detail = missing[0] if missing else extra[0]
        raise ScoreInputError(f"missing judgment for judge/case cell {detail}")
    for cell, labels in cells.items():
        if set(labels) != judges:
            raise ScoreInputError(f"missing judge label for canonical cell {cell}")
    return dict(cells), judges


def score_condition_blind(
    judgments: Sequence[Mapping[str, Any]],
    alias_mappings: Sequence[Mapping[str, Any]],
    proposition_counts: Mapping[str, Mapping[str, int]],
) -> dict[str, Any]:
    """Score paired answers after mapping every judge-local alias to a condition."""

    counts = _validate_counts(proposition_counts)
    mappings = _mapping_index(alias_mappings)
    cells, _judges = _canonical_labels(judgments, mappings, counts)

    resolved: dict[tuple[str, str, str, int], str] = {}
    agreement_count = 0
    for cell, by_judge in cells.items():
        labels = set(by_judge.values())
        if len(labels) == 1:
            resolved[cell] = next(iter(labels))
            agreement_count += 1
        else:
            resolved[cell] = "N" if cell[2] == "gold" else "S"

    metrics: dict[tuple[str, str], dict[str, int | float]] = {
        (case_id, condition): {
            "gold_supported": 0,
            "gold_count": counts[case_id]["gold"],
            "blocking_false": 0,
        }
        for case_id in counts
        for condition in sorted(CONDITIONS)
    }
    for (case_id, condition, kind, _index), label in resolved.items():
        metric = metrics[(case_id, condition)]
        if kind == "gold":
            metric["gold_supported"] += int(label == "S")
            metric["blocking_false"] += int(label == "C")
        else:
            metric["blocking_false"] += int(label == "S")

    pairs = []
    for case_id in sorted(counts):
        pair: dict[str, Any] = {"case_id": case_id}
        for condition in ("with_c3", "without_c3"):
            metric = metrics[(case_id, condition)]
            pair[condition] = {
                **metric,
                "coverage": metric["gold_supported"] / metric["gold_count"],
            }
        pair["coverage_lift"] = (
            pair["with_c3"]["coverage"] - pair["without_c3"]["coverage"]
        )
        pair["blocking_false_lift"] = (
            pair["with_c3"]["blocking_false"]
            - pair["without_c3"]["blocking_false"]
        )
        pairs.append(pair)

    return {
        "claim_label_agreement": agreement_count / len(resolved),
        "claim_label_count": len(resolved),
        "pairs": pairs,
        "required_surface_lift": (
            sum(pair["coverage_lift"] for pair in pairs) / len(pairs)
        ),
        "blocking_false_claim_lift": (
            sum(pair["blocking_false_lift"] for pair in pairs) / len(pairs)
        ),
    }
