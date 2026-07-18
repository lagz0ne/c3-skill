#!/usr/bin/env python3
"""Plan and retain budgeted paired skill evaluations without leaking project data.

The default and currently supported command is ``plan``. It performs no agent
calls. Live execution is added through an isolated runner adapter; this module
owns pairing, model selection, budgets, and the public result boundary.
"""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import os
import platform
import random
import re
import shlex
import shutil
import subprocess
import tempfile
import time
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Any

try:
    from scripts import agent_efficiency_eval as efficiency
except ModuleNotFoundError:  # direct `python3 scripts/paired_skill_eval.py`
    import agent_efficiency_eval as efficiency


CONDITIONS = ("without_c3", "with_c3")
CASE_FAMILIES = ("blast_radius", "pre_initiative_change_unit")
ROOT = Path(__file__).resolve().parents[1]
BASELINE_INSTRUCTIONS = (
    ROOT / "research" / "eval" / "skill-eval" / "harness" / "prompts" / "neutral-repository-baseline.md"
)
C3_TREATMENT_INSTRUCTIONS = (
    ROOT / "research" / "eval" / "skill-eval" / "harness" / "prompts" / "c3-impact-treatment.md"
)
C3_BOOTSTRAP = (
    ROOT / "research" / "eval" / "skill-eval" / "harness" / "bin" / "c3-impact-bootstrap.sh"
)
OPAQUE_CASE_ID = re.compile(r"^(BR|PI)-\d{3,6}$")
OPAQUE_STUDY_ID = re.compile(r"^STUDY-\d{3,6}$")
SHA256 = re.compile(r"^[0-9a-f]{64}$")


class BudgetExceeded(ValueError):
    """Raised before scheduling when a proposed run set crosses a frozen wall."""


@dataclass(frozen=True)
class EvalCase:
    case_id: str
    family: str
    prompt: str
    registered_gold_lanes: tuple[str, ...] = ()


def _normalized_words(value: str) -> tuple[str, ...]:
    return tuple(re.findall(r"[a-z0-9]+", value.lower()))


def audit_prompt_leakage(prompt: str, registered_gold_lanes: list[str]) -> dict[str, Any]:
    """Detect prompts that enumerate the answer lanes they are meant to discover."""
    prompt_words = _normalized_words(prompt)
    prompt_text = " ".join(prompt_words)
    matched: list[str] = []
    for lane in registered_gold_lanes:
        normalized_lane = " ".join(_normalized_words(lane))
        if normalized_lane and normalized_lane in prompt_text:
            matched.append(lane)
    registered_count = len(registered_gold_lanes)
    ratio = len(matched) / registered_count if registered_count else 0.0
    # One incidental overlap is allowed. A prompt fails when it supplies a
    # majority of a multi-lane rubric or enumerates three or more lanes.
    passed = len(matched) < 3 and ratio < 0.5
    return {
        "passed": passed,
        "registered_lane_count": registered_count,
        "matched_lane_count": len(matched),
        "matched_lane_ratio": ratio,
        "matched_lanes": matched,
    }


@dataclass(frozen=True)
class BudgetPolicy:
    max_runs: int = 24
    max_total_cost_usd: float = 5.0
    max_cost_per_run_usd: float = 0.5
    max_tokens_per_run: int = 250_000
    timeout_seconds: int = 900
    max_tool_calls: int = 6
    max_tool_result_bytes: int = 131_072
    max_output_bytes: int = 524_288
    observe_cost_only: bool = False


@dataclass(frozen=True)
class SelectedModel:
    model_id: str
    provider: str
    price_version: str
    input_per_million_usd: float
    cached_input_per_million_usd: float
    output_per_million_usd: float
    estimated_cost_usd: float


@dataclass(frozen=True)
class CommonRunManifest:
    case_id: str
    family: str
    trial: int
    provider: str
    model: str
    repo_snapshot_sha256: str
    task_sha256: str
    timeout_seconds: int
    max_tokens: int
    reasoning_effort: str = "medium"
    baseline_instruction_sha256: str = ""
    max_tool_calls: int = 6
    max_tool_result_bytes: int = 131_072
    max_output_bytes: int = 524_288


GENERIC_RESULT_FIELDS = frozenset(
    {
        "study_id",
        "case_id",
        "case_family",
        "condition",
        "agent",
        "model",
        "reasoning_effort",
        "trial",
        "status",
        "exit_code",
        "timed_out",
        "quality_score",
        "correctness_score",
        "trace_completeness_score",
        "reasoning_depth_score",
        "grounding_score",
        "no_hallucination_score",
        "change_usefulness_score",
        "passed",
        "input_tokens",
        "cached_input_tokens",
        "output_tokens",
        "reasoning_output_tokens",
        "total_tokens",
        "effective_tokens",
        "model_turns",
        "tool_calls",
        "runtime_guard_reason",
        "tool_result_bytes",
        "elapsed_ms",
        "setup_elapsed_ms",
        "cost_usd",
        "scoring_cost_usd",
        "independent_review_count",
        "deterministic_evidence_count",
        "review_receipt_sha256",
        "deterministic_evidence_sha256",
        "total_cost_usd",
        "cost_policy",
        "protocol_deviation_codes",
        "price_version",
        "runner_version",
        "skill_sha256",
        "baseline_instruction_sha256",
        "treatment_instruction_sha256",
        "treatment_runtime_layer",
        "c3_invocation_count",
        "c3_route_command_count",
        "c3_impact_command_count",
        "c3_evidence_command_count",
        "c3_success_count",
        "c3_route_success_count",
        "c3_impact_success_count",
        "c3_evidence_success_count",
        "max_tool_calls",
        "max_tool_result_bytes",
        "max_output_bytes",
        "max_answer_words",
    }
)


def _money(value: float) -> float:
    return round(value, 8)


def select_cheapest_model(
    pricing: dict[str, Any],
    *,
    provider: str,
    estimated_input_tokens: int,
    estimated_output_tokens: int,
) -> SelectedModel:
    version = str(pricing.get("version") or "").strip()
    if not version:
        raise ValueError("pricing needs a non-empty version")
    if pricing.get("currency") != "USD":
        raise ValueError("pricing currency must be USD")
    candidates: list[SelectedModel] = []
    for row in pricing.get("models") or []:
        if not isinstance(row, dict):
            continue
        if row.get("provider") != provider or row.get("eligible") is not True:
            continue
        if row.get("live_allowed") is not True:
            continue
        model_id = str(row.get("id") or "").strip()
        if not model_id:
            continue
        input_price = float(row["input_per_million_usd"])
        cached_price = float(row["cached_input_per_million_usd"])
        output_price = float(row["output_per_million_usd"])
        if min(input_price, cached_price, output_price) < 0:
            raise ValueError(f"negative price for model {model_id}")
        estimate = (
            estimated_input_tokens * input_price
            + estimated_output_tokens * output_price
        ) / 1_000_000
        candidates.append(
            SelectedModel(
                model_id=model_id,
                provider=provider,
                price_version=version,
                input_per_million_usd=input_price,
                cached_input_per_million_usd=cached_price,
                output_per_million_usd=output_price,
                estimated_cost_usd=_money(estimate),
            )
        )
    if not candidates:
        raise ValueError(f"no live-allowed eligible models for provider {provider}")
    return min(candidates, key=lambda item: (item.estimated_cost_usd, item.model_id))


def select_model(
    pricing: dict[str, Any],
    *,
    provider: str,
    model_id: str,
    estimated_input_tokens: int,
    estimated_output_tokens: int,
) -> SelectedModel:
    version = str(pricing.get("version") or "").strip()
    if not version or pricing.get("currency") != "USD":
        raise ValueError("pricing needs a USD version")
    for row in pricing.get("models") or []:
        if not isinstance(row, dict) or row.get("provider") != provider or row.get("id") != model_id:
            continue
        if row.get("eligible") is not True or row.get("live_allowed") is not True:
            raise ValueError(f"requested model {model_id} is not live-allowed")
        input_price = float(row["input_per_million_usd"])
        cached_price = float(row["cached_input_per_million_usd"])
        output_price = float(row["output_per_million_usd"])
        if min(input_price, cached_price, output_price) < 0:
            raise ValueError(f"negative price for model {model_id}")
        estimate = (
            estimated_input_tokens * input_price
            + estimated_output_tokens * output_price
        ) / 1_000_000
        return SelectedModel(
            model_id=model_id,
            provider=provider,
            price_version=version,
            input_per_million_usd=input_price,
            cached_input_per_million_usd=cached_price,
            output_per_million_usd=output_price,
            estimated_cost_usd=_money(estimate),
        )
    raise ValueError(f"requested model {model_id} is not listed for provider {provider}")


def check_budget(
    policy: BudgetPolicy,
    *,
    planned_runs: int,
    estimated_cost_per_run: float,
    spent_cost_usd: float = 0.0,
) -> None:
    if planned_runs > policy.max_runs:
        raise BudgetExceeded(
            f"planned runs {planned_runs} exceed max_runs {policy.max_runs}"
        )
    if policy.observe_cost_only:
        return
    if estimated_cost_per_run > policy.max_cost_per_run_usd:
        raise BudgetExceeded(
            "estimated per-run cost "
            f"${estimated_cost_per_run:.6f} exceeds ${policy.max_cost_per_run_usd:.6f}"
        )
    estimated_total = spent_cost_usd + planned_runs * estimated_cost_per_run
    if estimated_total > policy.max_total_cost_usd:
        raise BudgetExceeded(
            f"estimated total ${estimated_total:.6f} exceeds "
            f"${policy.max_total_cost_usd:.6f}"
        )


def build_arm_manifest(
    common: CommonRunManifest,
    condition: str,
    *,
    skill_sha256: str | None,
) -> dict[str, Any]:
    if condition not in CONDITIONS:
        raise ValueError(f"unknown condition: {condition}")
    if not SHA256.fullmatch(common.repo_snapshot_sha256):
        raise ValueError("repo snapshot must be sha256")
    if not SHA256.fullmatch(common.task_sha256):
        raise ValueError("task must be sha256")
    if not SHA256.fullmatch(common.baseline_instruction_sha256):
        raise ValueError("baseline instructions must be sha256")
    if condition == "with_c3" and not (skill_sha256 and SHA256.fullmatch(skill_sha256)):
        raise ValueError("with_c3 needs a skill sha256")
    if condition == "without_c3" and skill_sha256 is not None:
        raise ValueError("without_c3 cannot carry a skill sha256")
    return {
        "condition": condition,
        "common": asdict(common),
        "treatment": {
            "c3_available": condition == "with_c3",
            "c3_dir_present": condition == "with_c3",
            "skill_sha256": skill_sha256,
            "forced_instruction_sha256": (
                hashlib.sha256(C3_TREATMENT_INSTRUCTIONS.read_bytes()).hexdigest()
                if condition == "with_c3"
                else None
            ),
            "bootstrap_entrypoint": (
                "/opt/c3/bin/c3-impact-bootstrap" if condition == "with_c3" else None
            ),
            "bootstrap_sha256": (
                hashlib.sha256(C3_BOOTSTRAP.read_bytes()).hexdigest()
                if condition == "with_c3"
                else None
            ),
            "bootstrap_internal_c3_call_count": (
                {"minimum": 4, "search_miss_fallback": 5}
                if condition == "with_c3"
                else {"minimum": 0, "search_miss_fallback": 0}
            ),
            "estimand": "c3_treatment_package" if condition == "with_c3" else "neutral_baseline",
            "uptake_acceptance": (
                "supervisor_transcript_exact_first_command_exit_zero"
                if condition == "with_c3"
                else None
            ),
            "tool_call_accounting": (
                "bootstrap_is_one_provider_tool_call_and_four_or_five_c3_subcommands"
                if condition == "with_c3"
                else "provider_tool_calls_only"
            ),
            "provider_tool_set_parity": True,
        },
    }


def non_treatment_differences(
    control: dict[str, Any], treatment: dict[str, Any]
) -> dict[str, list[Any]]:
    left = control.get("common") or {}
    right = treatment.get("common") or {}
    keys = sorted(set(left) | set(right))
    return {key: [left.get(key), right.get(key)] for key in keys if left.get(key) != right.get(key)}


def empty_generic_result(
    *,
    study_id: str,
    case_id: str,
    family: str,
    condition: str,
    agent: str,
    model: str,
    trial: int,
    price_version: str,
    reasoning_effort: str = "medium",
) -> dict[str, Any]:
    return {
        "study_id": study_id,
        "case_id": case_id,
        "case_family": family,
        "condition": condition,
        "agent": agent,
        "model": model,
        "reasoning_effort": reasoning_effort,
        "trial": trial,
        "status": "planned",
        "exit_code": None,
        "timed_out": False,
        "quality_score": None,
        "correctness_score": None,
        "trace_completeness_score": None,
        "reasoning_depth_score": None,
        "grounding_score": None,
        "no_hallucination_score": None,
        "change_usefulness_score": None,
        "passed": None,
        "input_tokens": None,
        "cached_input_tokens": None,
        "output_tokens": None,
        "reasoning_output_tokens": None,
        "total_tokens": None,
        "effective_tokens": None,
        "model_turns": None,
        "tool_calls": None,
        "runtime_guard_reason": None,
        "tool_result_bytes": None,
        "elapsed_ms": None,
        "setup_elapsed_ms": None,
        "cost_usd": None,
        "scoring_cost_usd": None,
        "independent_review_count": None,
        "deterministic_evidence_count": None,
        "review_receipt_sha256": [],
        "deterministic_evidence_sha256": [],
        "total_cost_usd": None,
        "cost_policy": "validity_wall",
        "protocol_deviation_codes": [],
        "price_version": price_version,
        "runner_version": "paired-skill-eval-v1",
        "skill_sha256": None,
        "baseline_instruction_sha256": hashlib.sha256(BASELINE_INSTRUCTIONS.read_bytes()).hexdigest(),
        "treatment_instruction_sha256": (
            hashlib.sha256(C3_TREATMENT_INSTRUCTIONS.read_bytes()).hexdigest()
            if condition == "with_c3"
            else None
        ),
        "treatment_runtime_layer": (
            {
                "codex": "codex_developer_instructions",
                "claude": "claude_append_system_prompt",
                "kilo": "repository_instructions",
            }[agent]
            if condition == "with_c3"
            else None
        ),
        "c3_invocation_count": 0,
        "c3_route_command_count": 0,
        "c3_impact_command_count": 0,
        "c3_evidence_command_count": 0,
        "c3_success_count": 0,
        "c3_route_success_count": 0,
        "c3_impact_success_count": 0,
        "c3_evidence_success_count": 0,
        "max_tool_calls": 6,
        "max_tool_result_bytes": 131_072,
        "max_output_bytes": 524_288,
        "max_answer_words": 250,
    }


def answer_word_limit_deviations(answer_path: Path, max_answer_words: int) -> list[str]:
    if max_answer_words < 1:
        raise ValueError("max_answer_words must be a positive integer")
    answer = answer_path.read_text(encoding="utf-8", errors="replace")
    return (
        ["ANSWER_WORD_LIMIT_BREACH"]
        if len(answer.split()) > max_answer_words
        else []
    )


def validate_generic_result(record: dict[str, Any]) -> None:
    unknown = sorted(set(record) - GENERIC_RESULT_FIELDS)
    missing = sorted(GENERIC_RESULT_FIELDS - set(record))
    if unknown:
        raise ValueError(f"unknown retained fields: {', '.join(unknown)}")
    if missing:
        raise ValueError(f"missing retained fields: {', '.join(missing)}")
    if not OPAQUE_STUDY_ID.fullmatch(str(record["study_id"])):
        raise ValueError("study_id must be opaque, like STUDY-001")
    if not OPAQUE_CASE_ID.fullmatch(str(record["case_id"])):
        raise ValueError("case_id must be opaque, like BR-001 or PI-001")
    if record["case_family"] not in CASE_FAMILIES:
        raise ValueError("unknown case family")
    if record["condition"] not in CONDITIONS:
        raise ValueError("unknown condition")
    if record["agent"] not in {"codex", "claude", "kilo"}:
        raise ValueError("unknown agent")
    if record["status"] not in {"planned", "collected", "scored", "invalid"}:
        raise ValueError("unknown status")
    if not re.fullmatch(r"[A-Za-z0-9._/-]{1,120}", str(record["model"])):
        raise ValueError("model id contains disallowed characters")
    if record["reasoning_effort"] not in {"low", "medium", "high", "xhigh", "max"}:
        raise ValueError("unknown reasoning effort")
    if not re.fullmatch(r"[A-Za-z0-9._-]{1,80}", str(record["price_version"])):
        raise ValueError("price version contains disallowed characters")
    if not isinstance(record["trial"], int) or record["trial"] < 1:
        raise ValueError("trial must be a positive integer")
    if not isinstance(record["max_answer_words"], int) or record["max_answer_words"] < 1:
        raise ValueError("max_answer_words must be a positive integer")
    deviations = record["protocol_deviation_codes"]
    if not isinstance(deviations, list) or not all(
        isinstance(value, str) and re.fullmatch(r"[A-Z0-9_]+", value)
        for value in deviations
    ):
        raise ValueError("protocol deviation codes must be uppercase enums")
    skill_sha = record["skill_sha256"]
    if record["condition"] == "with_c3" and not (
        isinstance(skill_sha, str) and SHA256.fullmatch(skill_sha)
    ):
        raise ValueError("with_c3 retained rows need skill_sha256")
    if record["condition"] == "without_c3" and skill_sha is not None:
        raise ValueError("without_c3 retained rows cannot include skill_sha256")
    if not SHA256.fullmatch(str(record["baseline_instruction_sha256"])):
        raise ValueError("baseline instructions must be sha256")
    treatment_sha = record["treatment_instruction_sha256"]
    if record["condition"] == "with_c3" and not (
        isinstance(treatment_sha, str) and SHA256.fullmatch(treatment_sha)
    ):
        raise ValueError("with_c3 retained rows need treatment_instruction_sha256")
    if record["condition"] == "without_c3" and treatment_sha is not None:
        raise ValueError("without_c3 retained rows cannot include treatment instructions")
    expected_layer = {
        "codex": "codex_developer_instructions",
        "claude": "claude_append_system_prompt",
        "kilo": "repository_instructions",
    }[record["agent"]]
    treatment_layer = record["treatment_runtime_layer"]
    if record["condition"] == "with_c3" and treatment_layer != expected_layer:
        raise ValueError("with_c3 retained rows need the provider treatment layer")
    if record["condition"] == "without_c3" and treatment_layer is not None:
        raise ValueError("without_c3 retained rows cannot include a treatment layer")
    for field in (
        "c3_invocation_count",
        "c3_route_command_count",
        "c3_impact_command_count",
        "c3_evidence_command_count",
        "c3_success_count",
        "c3_route_success_count",
        "c3_impact_success_count",
        "c3_evidence_success_count",
    ):
        if not isinstance(record[field], int) or record[field] < 0:
            raise ValueError(f"{field} must be a non-negative integer")
    review_count = record["independent_review_count"]
    evidence_count = record["deterministic_evidence_count"]
    review_receipts = record["review_receipt_sha256"]
    evidence_receipts = record["deterministic_evidence_sha256"]
    if not isinstance(review_receipts, list) or not all(
        isinstance(value, str) and SHA256.fullmatch(value) for value in review_receipts
    ):
        raise ValueError("review receipts must be sha256 values")
    if not isinstance(evidence_receipts, list) or not all(
        isinstance(value, str) and SHA256.fullmatch(value) for value in evidence_receipts
    ):
        raise ValueError("deterministic evidence receipts must be sha256 values")
    if record["status"] == "scored":
        if not isinstance(review_count, int) or review_count < 1:
            raise ValueError("scored rows need one independent review")
        if not isinstance(evidence_count, int) or evidence_count < 1:
            raise ValueError("scored rows need deterministic evidence")
        if len(review_receipts) != review_count or len(set(review_receipts)) != review_count:
            raise ValueError("scored rows need distinct review receipt hashes")
        if len(evidence_receipts) != evidence_count or len(set(evidence_receipts)) != evidence_count:
            raise ValueError("scored rows need deterministic evidence receipt hashes")
    if not isinstance(record["max_tool_calls"], int) or record["max_tool_calls"] < 1:
        raise ValueError("max_tool_calls must be positive")
    if not isinstance(record["max_tool_result_bytes"], int) or record["max_tool_result_bytes"] < 1:
        raise ValueError("max_tool_result_bytes must be positive")
    if not isinstance(record["max_output_bytes"], int) or record["max_output_bytes"] < 1:
        raise ValueError("max_output_bytes must be positive")


def successful_c3_bootstrap_count(transcript: str) -> int:
    commands: list[dict[str, Any]] = []
    for line in transcript.splitlines():
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            continue
        if event.get("type") != "item.completed":
            continue
        item = event.get("item")
        if not isinstance(item, dict) or item.get("type") != "command_execution":
            continue
        commands.append(item)
    if not commands:
        return 0
    first = commands[0]
    command = first.get("command")
    if not isinstance(command, str) or first.get("exit_code") != 0:
        return 0
    shell_command = command.strip()
    shell_prefix = "/usr/bin/zsh -lc "
    if shell_command.startswith(shell_prefix):
        shell_command = shell_command[len(shell_prefix) :].strip()
        if (
            len(shell_command) >= 2
            and shell_command[0] == shell_command[-1]
            and shell_command[0] in {'"', "'"}
        ):
            shell_command = shell_command[1:-1]
    if re.search(r"[;&|$`()<>\\\n\r]", shell_command):
        return 0
    try:
        argv = shlex.split(shell_command)
    except ValueError:
        return 0
    if len(argv) < 3 or argv[:2] != ["bash", "/opt/c3/bin/c3-impact-bootstrap"]:
        return 0
    if "C3_BOOTSTRAP_WRAPPER" in command:
        return 0
    return 1


def c3_uptake_deviations(
    condition: str, meta: dict[str, str], transcript: str
) -> list[str]:
    fields = (
        "c3_invocation_count",
        "c3_route_command_count",
        "c3_impact_command_count",
        "c3_evidence_command_count",
        "c3_success_count",
        "c3_route_success_count",
        "c3_impact_success_count",
        "c3_evidence_success_count",
    )
    counts = {
        field: int(meta.get(field, "0")) if meta.get(field, "0").isdigit() else 0
        for field in fields
    }
    bootstrap_successes = successful_c3_bootstrap_count(transcript)
    if condition == "without_c3":
        return ["C3_CONTROL_CONTAMINATION"] if any(counts.values()) or bootstrap_successes else []
    complete = bootstrap_successes >= 1
    return [] if complete else ["C3_UPTAKE_MISSING"]


def build_plan(
    cases: list[EvalCase],
    *,
    pricing: dict[str, Any],
    provider: str,
    repeats: int,
    policy: BudgetPolicy,
    estimated_input_tokens: int,
    estimated_output_tokens: int,
    model_id: str | None = None,
    reasoning_effort: str = "medium",
) -> dict[str, Any]:
    if repeats < 1:
        raise ValueError("repeat must be at least 1")
    if reasoning_effort not in {"low", "medium", "high", "xhigh", "max"}:
        raise ValueError("unknown reasoning effort")
    if model_id:
        selected = select_model(
            pricing,
            provider=provider,
            model_id=model_id,
            estimated_input_tokens=estimated_input_tokens,
            estimated_output_tokens=estimated_output_tokens,
        )
    else:
        selected = select_cheapest_model(
            pricing,
            provider=provider,
            estimated_input_tokens=estimated_input_tokens,
            estimated_output_tokens=estimated_output_tokens,
        )
    planned_runs = len(cases) * len(CONDITIONS) * repeats
    check_budget(
        policy,
        planned_runs=planned_runs,
        estimated_cost_per_run=selected.estimated_cost_usd,
    )
    runs = [
        {
            "case_id": case.case_id,
            "case_family": case.family,
            "trial": trial,
            "condition": condition,
            "provider": provider,
            "model": selected.model_id,
            "reasoning_effort": reasoning_effort,
            "timeout_seconds": policy.timeout_seconds,
            "max_tokens": policy.max_tokens_per_run,
            "max_tool_calls": policy.max_tool_calls,
            "max_tool_result_bytes": policy.max_tool_result_bytes,
            "max_output_bytes": policy.max_output_bytes,
            "estimated_cost_usd": selected.estimated_cost_usd,
        }
        for case in cases
        for trial in range(1, repeats + 1)
        for condition in CONDITIONS
    ]
    return {
        "dry_run": True,
        "case_count": len(cases),
        "planned_runs": planned_runs,
        "selected_model": selected.model_id,
        "reasoning_effort": reasoning_effort,
        "price_version": selected.price_version,
        "estimated_cost_per_run_usd": selected.estimated_cost_usd,
        "estimated_total_cost_usd": _money(planned_runs * selected.estimated_cost_usd),
        "budget": asdict(policy),
        "runs": runs,
    }


def load_cases(path: Path) -> list[EvalCase]:
    cases: list[EvalCase] = []
    seen: set[str] = set()
    for lineno, line in enumerate(path.read_text(encoding="utf-8").splitlines(), start=1):
        if not line.strip():
            continue
        row = json.loads(line)
        case = EvalCase(
            case_id=str(row.get("case_id") or ""),
            family=str(row.get("family") or ""),
            prompt=str(row.get("prompt") or ""),
            registered_gold_lanes=tuple(str(value) for value in row.get("registered_gold_lanes", [])),
        )
        if not OPAQUE_CASE_ID.fullmatch(case.case_id):
            raise ValueError(f"line {lineno}: case_id must be opaque")
        if case.case_id in seen:
            raise ValueError(f"line {lineno}: duplicate case_id")
        if case.family not in CASE_FAMILIES:
            raise ValueError(f"line {lineno}: unknown case family")
        if not case.prompt.strip():
            raise ValueError(f"line {lineno}: prompt is required")
        if not all(lane.strip() for lane in case.registered_gold_lanes):
            raise ValueError(f"line {lineno}: registered gold lanes must be non-empty strings")
        leakage = audit_prompt_leakage(case.prompt, list(case.registered_gold_lanes))
        if not leakage["passed"]:
            raise ValueError(
                f"line {lineno}: prompt leaks {leakage['matched_lane_count']} of "
                f"{leakage['registered_lane_count']} registered gold lanes"
            )
        expected_prefix = "BR-" if case.family == "blast_radius" else "PI-"
        if not case.case_id.startswith(expected_prefix):
            raise ValueError(f"line {lineno}: case_id prefix does not match family")
        seen.add(case.case_id)
        cases.append(case)
    if not cases:
        raise ValueError("cases file is empty")
    return cases


def load_pricing(path: Path) -> dict[str, Any]:
    value = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(value, dict):
        raise ValueError("pricing must be a JSON object")
    return value


def validate_live_pricing(pricing: dict[str, Any], *, now: dt.datetime | None = None) -> None:
    raw = str(pricing.get("observed_at") or "").strip()
    if not raw:
        raise ValueError("live pricing needs observed_at")
    try:
        observed = dt.datetime.fromisoformat(raw.replace("Z", "+00:00"))
    except ValueError as exc:
        raise ValueError("live pricing observed_at must be ISO-8601") from exc
    if observed.tzinfo is None:
        raise ValueError("live pricing observed_at needs a timezone")
    current = now or dt.datetime.now(dt.timezone.utc)
    age = current - observed.astimezone(dt.timezone.utc)
    if age > dt.timedelta(days=30):
        raise ValueError("live pricing is stale beyond P30D")
    if age < -dt.timedelta(days=1):
        raise ValueError("live pricing observed_at is in the future")


def repo_state(repo: Path) -> str:
    head = subprocess.run(
        ["git", "-C", str(repo), "rev-parse", "HEAD"],
        text=False,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )
    status = subprocess.run(
        ["git", "-C", str(repo), "status", "--porcelain=v1", "-z", "--untracked-files=all"],
        text=False,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )
    if head.returncode != 0 or status.returncode != 0:
        raise ValueError("repo state is unreadable")
    return hashlib.sha256(head.stdout + b"\0" + status.stdout).hexdigest()


def tree_sha256(root: Path) -> str:
    digest = hashlib.sha256()
    for path in sorted(item for item in root.rglob("*") if item.is_file()):
        relative = path.relative_to(root)
        if relative.parts[0] == "bin":
            continue
        digest.update(relative.as_posix().encode("utf-8"))
        digest.update(b"\0")
        digest.update(hashlib.sha256(path.read_bytes()).hexdigest().encode("ascii"))
        digest.update(b"\n")
    return digest.hexdigest()


def runner_freeze_environment(repo: Path, prompt: Path | None = None) -> dict[str, str]:
    skill_root = ROOT / "skills" / "c3"
    version_path = skill_root / "bin" / "VERSION"
    wrapper_path = skill_root / "bin" / "c3x.sh"
    version = version_path.read_text(encoding="utf-8").strip()
    system = platform.system().lower()
    machine = {"x86_64": "amd64", "aarch64": "arm64", "arm64": "arm64"}.get(
        platform.machine(), platform.machine()
    )
    binary_path = skill_root / "bin" / f"c3x-{version}-{system}-{machine}"
    if os.environ.get("C3_FROZEN_BINARY"):
        binary_path = Path(os.environ["C3_FROZEN_BINARY"]).resolve()
    if not binary_path.is_file():
        raise ValueError("frozen local C3 binary is missing")
    head = subprocess.run(
        ["git", "-C", str(repo), "rev-parse", "HEAD"],
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )
    if head.returncode != 0:
        raise ValueError("repo must be a Git checkout with a readable HEAD")
    values = {
        "C3_EXPECT_SEED_HEAD_SHA256": hashlib.sha256(
            head.stdout.strip().encode("utf-8")
        ).hexdigest(),
        "C3_EXPECT_SELECTED_BINARY_SHA256": hashlib.sha256(binary_path.read_bytes()).hexdigest(),
        "C3_EXPECT_SKILL_MD_SHA256": hashlib.sha256((skill_root / "SKILL.md").read_bytes()).hexdigest(),
        "C3_EXPECT_SKILL_TREE_SHA256": tree_sha256(skill_root),
        "C3_EXPECT_WRAPPER_SHA256": hashlib.sha256(wrapper_path.read_bytes()).hexdigest(),
        "C3_EXPECT_VERSION_SHA256": hashlib.sha256(version_path.read_bytes()).hexdigest(),
    }
    if prompt is not None:
        values["C3_EXPECT_PROMPT_SHA256"] = hashlib.sha256(prompt.read_bytes()).hexdigest()
    return values


def calculate_cost_usd(usage: dict[str, int], model: SelectedModel) -> float:
    input_tokens = max(0, int(usage.get("input_tokens", 0)))
    cached = min(input_tokens, max(0, int(usage.get("cached_input_tokens", 0))))
    uncached = input_tokens - cached
    output = max(0, int(usage.get("output_tokens", 0)))
    return _money(
        (
            uncached * model.input_per_million_usd
            + cached * model.cached_input_per_million_usd
            + output * model.output_per_million_usd
        )
        / 1_000_000
    )


def observed_cost_usd(
    usage: dict[str, int],
    model: SelectedModel,
    *,
    reported_cost_usd: float | None,
) -> float:
    calculated = calculate_cost_usd(usage, model)
    if reported_cost_usd is None:
        return calculated
    return _money(max(calculated, reported_cost_usd))


def extract_tool_call_count(text: str) -> int:
    count = 0
    for line in text.splitlines():
        try:
            value = json.loads(line)
        except json.JSONDecodeError:
            continue
        if not isinstance(value, dict):
            continue
        event_type = str(value.get("type") or "")
        item = value.get("item") if isinstance(value.get("item"), dict) else {}
        item_type = str(item.get("type") or "")
        if "tool" in event_type.lower() or "tool" in item_type.lower():
            count += 1
    return count


def read_meta(path: Path) -> dict[str, str]:
    values: dict[str, str] = {}
    if not path.exists():
        return values
    for line in path.read_text(encoding="utf-8", errors="replace").splitlines():
        key, separator, value = line.partition("=")
        if separator and re.fullmatch(r"[A-Za-z0-9_.-]+", key):
            values[key] = value
    return values


def validate_score(score: dict[str, Any], *, minimum_independent_reviews: int = 1) -> None:
    numeric_fields = (
        "quality_score",
        "correctness_score",
        "trace_completeness_score",
        "reasoning_depth_score",
        "grounding_score",
        "no_hallucination_score",
        "change_usefulness_score",
    )
    for field in numeric_fields:
        value = score.get(field)
        if not isinstance(value, (int, float)) or isinstance(value, bool) or not 1 <= float(value) <= 5:
            raise ValueError(f"score field {field} must be numeric from 1 to 5")
    if not isinstance(score.get("passed"), bool):
        raise ValueError("score passed must be boolean")
    independent_review_count = int(score.get("independent_review_count") or 0)
    deterministic_evidence_count = int(score.get("deterministic_evidence_count") or 0)
    if minimum_independent_reviews < 1:
        raise ValueError("minimum independent reviews must be positive")
    if independent_review_count < minimum_independent_reviews:
        raise ValueError(f"score needs at least {minimum_independent_reviews} independent review")
    if deterministic_evidence_count < 1:
        raise ValueError("score needs deterministic evidence")
    review_receipts = score.get("review_receipt_sha256")
    if not isinstance(review_receipts, list) or not all(
        isinstance(value, str) and SHA256.fullmatch(value) for value in review_receipts
    ):
        raise ValueError("score needs review receipt hashes")
    if len(review_receipts) != independent_review_count or len(set(review_receipts)) != independent_review_count:
        raise ValueError("score needs distinct review receipt hashes")
    evidence_receipts = score.get("deterministic_evidence_sha256")
    if not isinstance(evidence_receipts, list) or not all(
        isinstance(value, str) and SHA256.fullmatch(value) for value in evidence_receipts
    ):
        raise ValueError("score needs deterministic evidence receipt hashes")
    if len(evidence_receipts) != deterministic_evidence_count or len(set(evidence_receipts)) != deterministic_evidence_count:
        raise ValueError("score needs one hash per deterministic evidence receipt")
    scoring_cost = score.get("scoring_cost_usd")
    if (
        not isinstance(scoring_cost, (int, float))
        or isinstance(scoring_cost, bool)
        or float(scoring_cost) < 0
    ):
        raise ValueError("score needs a non-negative scoring_cost_usd")


def render_prompt(case: EvalCase, condition: str, *, max_tool_calls: int = 6) -> str:
    common = f"""# Paired architecture evaluation

Work only in the isolated repository snapshot. Do not use memories, global
skills, plugins, or host files. Answer the task as an engineer preparing a safe
change.

Return one compact JSON object with keys in this exact order: current_truth, change_impact, proposed_invariants, and unknowns. Establish current_truth before
designing the change. Each current-truth claim must include a precise source anchor
and say what the repository does now. change_impact must include owners, the causal
path, affected direct and indirect surfaces, failure boundaries, and executable
verification where evidence supports them. proposed_invariants must be explicitly
future-facing. unknowns must name the smallest next proof. Separate proved facts
from inference. Do not present a proposal as current behavior.
Target at most 220 words so the complete JSON stays below the hard 250-word
validity limit.

Use at most {max_tool_calls} discovery tool calls. This is both the research
authority and the runtime ceiling. Keep each tool result below 40 lines and
roughly 1 KB by using focused patterns, narrow ranges, and explicit `head`
bounds. Prefer targeted lookup and narrow source reads; stop
when the causal path, risks, and verification are supported. Do not implement
changes or run broad test suites. In the final answer, do not name, quote, or
describe any analysis tool, skill, condition, treatment, or control arm.

## Task

{case.prompt.strip()}
"""
    if condition == "with_c3":
        treatment = """

## Treatment

Directly invoke the C3 impact workflow for this task with exactly this entrypoint:

`bash /opt/c3/bin/c3-impact-bootstrap "<short behavior or domain>"`

Run that bootstrap once before ordinary repository tools. It performs the required
skill load, route, graph, and evidence reads through the frozen local C3 wrapper.
Do not separately read `SKILL.md` before the bootstrap. Follow the returned route
and references for pre-change blast-radius analysis. Spend the remaining tool-call
budget on narrow source proof across distinct owner, consumer/state, persistence or
event, compatibility or isolation, and test lanes. Do not replace
the bootstrap with manual repeated searches, bare `c3x`, `$c3`, `/c3`, or any
installed/global C3 skill.
"""
    else:
        treatment = """

## Control

Inspect the repository's ordinary code and documentation with the standard
local tools available in the isolated workspace.
"""
    return common + treatment


def run_score_command(template: str, *, answer: Path, cases: Path, case_id: str) -> dict[str, Any]:
    command = [
        part.replace("{answer}", str(answer)).replace("{cases}", str(cases)).replace("{case_id}", case_id)
        for part in shlex.split(template)
    ]
    result = subprocess.run(
        command,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )
    if result.returncode != 0:
        raise RuntimeError(f"blind scorer failed for case {case_id}")
    try:
        score = json.loads(result.stdout)
    except json.JSONDecodeError as exc:
        raise RuntimeError(f"blind scorer returned invalid JSON for case {case_id}") from exc
    if not isinstance(score, dict):
        raise RuntimeError(f"blind scorer returned a non-object for case {case_id}")
    validate_score(score)
    return score


def append_generic_result(path: Path, record: dict[str, Any]) -> None:
    validate_generic_result(record)
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a", encoding="utf-8") as handle:
        handle.write(json.dumps(record, sort_keys=True, separators=(",", ":")) + "\n")


def run_study(args: argparse.Namespace) -> dict[str, Any]:
    if not args.run:
        raise ValueError("live execution requires --run")
    required = ("repo", "cases", "pricing", "runner", "results", "study_id")
    if not args.defer_scoring:
        required += ("score_command",)
    missing = [name for name in required if not getattr(args, name, None)]
    if missing:
        raise ValueError("live execution missing: " + ", ".join(missing))
    if not OPAQUE_STUDY_ID.fullmatch(args.study_id):
        raise ValueError("study_id must be opaque, like STUDY-001")

    repo = Path(args.repo)
    cases_path = Path(args.cases)
    pricing = load_pricing(Path(args.pricing))
    validate_live_pricing(pricing)
    cases = load_cases(cases_path)
    policy = BudgetPolicy(
        max_runs=args.max_runs,
        max_total_cost_usd=args.budget_usd,
        max_cost_per_run_usd=args.max_cost_per_run_usd,
        max_tokens_per_run=args.max_tokens_per_run,
        timeout_seconds=args.timeout_seconds,
        max_tool_calls=args.max_tool_calls,
        max_tool_result_bytes=args.max_tool_result_bytes,
        max_output_bytes=args.max_output_bytes,
        observe_cost_only=args.observe_cost_only,
    )
    plan = build_plan(
        cases,
        pricing=pricing,
        provider=args.provider,
        repeats=args.repeat,
        policy=policy,
        estimated_input_tokens=args.estimated_input_tokens,
        estimated_output_tokens=args.estimated_output_tokens,
        model_id=args.model,
        reasoning_effort=args.effort,
    )
    if args.model:
        selected = select_model(
            pricing,
            provider=args.provider,
            model_id=args.model,
            estimated_input_tokens=args.estimated_input_tokens,
            estimated_output_tokens=args.estimated_output_tokens,
        )
    else:
        selected = select_cheapest_model(
            pricing,
            provider=args.provider,
            estimated_input_tokens=args.estimated_input_tokens,
            estimated_output_tokens=args.estimated_output_tokens,
        )
    source_state = repo_state(repo)
    snapshot_sha = repo_snapshot_sha256(repo)
    skill_path = ROOT / "skills" / "c3" / "SKILL.md"
    skill_sha = hashlib.sha256(skill_path.read_bytes()).hexdigest()
    baseline_sha = hashlib.sha256(BASELINE_INSTRUCTIONS.read_bytes()).hexdigest()
    treatment_sha = hashlib.sha256(C3_TREATMENT_INSTRUCTIONS.read_bytes()).hexdigest()
    frozen_runner_env = runner_freeze_environment(repo)
    results_path = Path(args.results)
    private_manifest_path = Path(args.private_answer_manifest) if args.private_answer_manifest else None
    private_answer_dir = Path(args.private_answer_dir) if args.private_answer_dir else None
    if bool(private_manifest_path) != bool(private_answer_dir):
        raise ValueError("--private-answer-manifest and --private-answer-dir must be used together")
    if private_manifest_path and private_manifest_path.exists() and private_manifest_path.stat().st_size:
        raise ValueError("private answer manifest must not already contain records")
    if private_answer_dir:
        private_answer_dir.mkdir(parents=True, exist_ok=True)
    if results_path.exists() and results_path.stat().st_size:
        raise ValueError("results file must not already contain records")

    schedule = list(plan["runs"])
    random.Random(args.random_seed).shuffle(schedule)
    spent = 0.0
    completed = 0
    pair_manifests: dict[tuple[str, int], dict[str, dict[str, Any]]] = {}

    with tempfile.TemporaryDirectory(prefix="paired-skill-eval-raw-") as raw_name:
        raw_root = Path(raw_name)
        for index, row in enumerate(schedule, start=1):
            remaining = policy.max_total_cost_usd - spent
            if not policy.observe_cost_only and remaining < selected.estimated_cost_usd:
                raise BudgetExceeded("remaining budget cannot admit the next run")
            case = next(item for item in cases if item.case_id == row["case_id"])
            common = CommonRunManifest(
                case_id=case.case_id,
                family=case.family,
                trial=int(row["trial"]),
                provider=args.provider,
                model=selected.model_id,
                repo_snapshot_sha256=snapshot_sha,
                task_sha256=hashlib.sha256(case.prompt.encode("utf-8")).hexdigest(),
                timeout_seconds=policy.timeout_seconds,
                max_tokens=policy.max_tokens_per_run,
                reasoning_effort=args.effort,
                baseline_instruction_sha256=baseline_sha,
                max_tool_calls=policy.max_tool_calls,
                max_tool_result_bytes=policy.max_tool_result_bytes,
                max_output_bytes=policy.max_output_bytes,
            )
            condition = str(row["condition"])
            manifest = build_arm_manifest(
                common,
                condition,
                skill_sha256=skill_sha if condition == "with_c3" else None,
            )
            pair_key = (case.case_id, int(row["trial"]))
            pair_manifests.setdefault(pair_key, {})[condition] = manifest
            arms = pair_manifests[pair_key]
            if set(arms) == set(CONDITIONS):
                differences = non_treatment_differences(arms["without_c3"], arms["with_c3"])
                if differences:
                    raise RuntimeError("paired arm manifests differ outside the treatment")

            opaque_label = f"R{index:04d}"
            run_dir = raw_root / opaque_label
            prompt = raw_root / f"{opaque_label}.prompt.md"
            prompt.write_text(
                render_prompt(case, condition, max_tool_calls=policy.max_tool_calls),
                encoding="utf-8",
            )
            command = [
                str(args.runner),
                "--agent",
                args.provider,
                "--prompt-file",
                str(prompt),
                "--seed-repo",
                str(repo),
                "--condition",
                condition,
                "--run-dir",
                str(run_dir),
                "--label",
                opaque_label,
                "--auth",
                args.auth,
                "--model",
                selected.model_id,
                "--effort",
                args.effort,
                "--run-timeout",
                str(policy.timeout_seconds),
                "--max-tool-calls",
                str(policy.max_tool_calls),
                "--max-tool-result-bytes",
                str(policy.max_tool_result_bytes),
                "--max-output-bytes",
                str(policy.max_output_bytes),
            ]
            if args.provider == "claude":
                command.extend(["--max-budget-usd", str(policy.max_cost_per_run_usd)])
            started = time.monotonic()
            run_env = os.environ.copy()
            run_env.update(frozen_runner_env)
            run_env["C3_EXPECT_PROMPT_SHA256"] = hashlib.sha256(prompt.read_bytes()).hexdigest()
            run = subprocess.run(
                command,
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
                env=run_env,
            )
            elapsed_ms = int((time.monotonic() - started) * 1000)
            if repo_state(repo) != source_state:
                raise RuntimeError("source repository changed during an isolated run")
            stdout_path = run_dir / f"{opaque_label}.stdout.txt"
            stderr_path = run_dir / f"{opaque_label}.stderr.txt"
            answer_path = run_dir / f"{opaque_label}.md"
            meta = read_meta(run_dir / f"{opaque_label}.meta.txt")
            agent_elapsed_ms = int(meta.get("agent_elapsed_ms", elapsed_ms))
            setup_elapsed_ms = int(meta["setup_elapsed_ms"]) if meta.get("setup_elapsed_ms", "").isdigit() else None
            transcript = "\n".join(
                [
                    stdout_path.read_text(encoding="utf-8", errors="replace") if stdout_path.exists() else "",
                    stderr_path.read_text(encoding="utf-8", errors="replace") if stderr_path.exists() else "",
                ]
            )
            usage = efficiency.extract_token_usage(transcript)
            turns = efficiency.extract_turn_count(transcript)
            deviations: list[str] = []
            guard_status = meta.get("runtime_guard_status")
            guard_reason = meta.get("runtime_guard_reason")
            if run.returncode == 86 or guard_status == "budget_killed":
                deviations.append("RUNTIME_BUDGET_KILLED")
            elif run.returncode != 0:
                deviations.append("RUNNER_FAILED")
            if meta.get("baseline_instruction_sha256") != baseline_sha:
                deviations.append("BASELINE_PARITY_BREACH")
            if condition == "with_c3":
                if meta.get("treatment_instruction_sha256") != treatment_sha:
                    deviations.append("TREATMENT_INSTRUCTION_BREACH")
                expected_treatment_layer = {
                    "codex": "codex_developer_instructions",
                    "claude": "claude_append_system_prompt",
                    "kilo": "repository_instructions",
                }[args.provider]
                if meta.get("treatment_runtime_layer") != expected_treatment_layer:
                    deviations.append("TREATMENT_RUNTIME_LAYER_BREACH")
            elif meta.get("treatment_instruction_sha256") != "unmounted":
                deviations.append("CONTROL_INSTRUCTION_BREACH")
            deviations.extend(c3_uptake_deviations(condition, meta, transcript))
            if usage is None:
                deviations.append("USAGE_MISSING")
                usage = {}
            if turns is None:
                deviations.append("TURN_COUNT_MISSING")
            total_tokens = int(usage.get("total_tokens", 0))
            if not policy.observe_cost_only and total_tokens > policy.max_tokens_per_run:
                deviations.append("PER_RUN_TOKEN_BUDGET_BREACH")
            actual_cost = observed_cost_usd(
                usage,
                selected,
                reported_cost_usd=efficiency.extract_reported_cost_usd(transcript),
            )
            if not policy.observe_cost_only and actual_cost > policy.max_cost_per_run_usd:
                deviations.append("PER_RUN_COST_BUDGET_BREACH")
            if agent_elapsed_ms > policy.timeout_seconds * 1000:
                deviations.append("PER_RUN_TIME_BUDGET_BREACH")
            if not answer_path.exists():
                deviations.append("ANSWER_MISSING")
            else:
                deviations.extend(
                    answer_word_limit_deviations(answer_path, args.max_answer_words)
                )

            record = empty_generic_result(
                study_id=args.study_id,
                case_id=case.case_id,
                family=case.family,
                condition=condition,
                agent=args.provider,
                model=selected.model_id,
                trial=int(row["trial"]),
                price_version=selected.price_version,
                reasoning_effort=args.effort,
            )
            record.update(
                {
                    "status": "invalid" if deviations else (
                        "collected" if args.defer_scoring else "scored"
                    ),
                    "exit_code": run.returncode,
                    "timed_out": guard_reason == "max_seconds" or run.returncode == 124,
                    "input_tokens": usage.get("input_tokens"),
                    "cached_input_tokens": usage.get("cached_input_tokens"),
                    "output_tokens": usage.get("output_tokens"),
                    "reasoning_output_tokens": usage.get("reasoning_output_tokens"),
                    "total_tokens": usage.get("total_tokens"),
                    "effective_tokens": usage.get("effective_tokens"),
                    "model_turns": turns,
                    "tool_calls": int(meta["runtime_guard_tool_calls"])
                    if meta.get("runtime_guard_tool_calls", "").isdigit()
                    else extract_tool_call_count(transcript),
                    "runtime_guard_reason": guard_reason,
                    "tool_result_bytes": int(meta["runtime_guard_tool_result_bytes"])
                    if meta.get("runtime_guard_tool_result_bytes", "").isdigit()
                    else None,
                    "elapsed_ms": agent_elapsed_ms,
                    "setup_elapsed_ms": setup_elapsed_ms,
                    "cost_usd": actual_cost,
                    "scoring_cost_usd": None,
                    "total_cost_usd": actual_cost,
                    "cost_policy": "observed_only" if policy.observe_cost_only else "validity_wall",
                    "protocol_deviation_codes": deviations,
                    "skill_sha256": skill_sha if condition == "with_c3" else None,
                    "baseline_instruction_sha256": baseline_sha,
                    "treatment_instruction_sha256": treatment_sha
                    if condition == "with_c3"
                    else None,
                    "treatment_runtime_layer": expected_treatment_layer
                    if condition == "with_c3"
                    else None,
                    "c3_invocation_count": int(meta.get("c3_invocation_count", "0")),
                    "c3_route_command_count": int(meta.get("c3_route_command_count", "0")),
                    "c3_impact_command_count": int(meta.get("c3_impact_command_count", "0")),
                    "c3_evidence_command_count": int(meta.get("c3_evidence_command_count", "0")),
                    "c3_success_count": int(meta.get("c3_success_count", "0")),
                    "c3_route_success_count": int(meta.get("c3_route_success_count", "0")),
                    "c3_impact_success_count": int(meta.get("c3_impact_success_count", "0")),
                    "c3_evidence_success_count": int(meta.get("c3_evidence_success_count", "0")),
                    "max_tool_calls": policy.max_tool_calls,
                    "max_tool_result_bytes": policy.max_tool_result_bytes,
                    "max_output_bytes": policy.max_output_bytes,
                    "max_answer_words": args.max_answer_words,
                }
            )
            if not deviations:
                if private_manifest_path and private_answer_dir:
                    answer_sha = hashlib.sha256(answer_path.read_bytes()).hexdigest()
                    retained_answer = private_answer_dir / case.case_id / f"{answer_sha}.md"
                    retained_answer.parent.mkdir(parents=True, exist_ok=True)
                    if not retained_answer.exists():
                        shutil.copyfile(answer_path, retained_answer)
                    private_manifest_path.parent.mkdir(parents=True, exist_ok=True)
                    with private_manifest_path.open("a", encoding="utf-8") as private_handle:
                        private_handle.write(json.dumps({
                            "study_id": args.study_id,
                            "case_id": case.case_id,
                            "trial": int(row["trial"]),
                            "condition": condition,
                            "answer_sha256": answer_sha,
                            "answer_path": str(retained_answer),
                            "prompt_sha256": hashlib.sha256(prompt.read_bytes()).hexdigest(),
                            "baseline_instruction_sha256": baseline_sha,
                            "base_result": record,
                        }, sort_keys=True) + "\n")
                if not args.defer_scoring:
                    score = run_score_command(
                        args.score_command,
                        answer=answer_path,
                        cases=cases_path,
                        case_id=case.case_id,
                    )
                    for field in (
                        "quality_score",
                        "correctness_score",
                        "trace_completeness_score",
                        "reasoning_depth_score",
                        "grounding_score",
                        "no_hallucination_score",
                        "change_usefulness_score",
                        "passed",
                        "independent_review_count",
                        "deterministic_evidence_count",
                        "review_receipt_sha256",
                        "deterministic_evidence_sha256",
                    ):
                        record[field] = score[field]
                    scoring_cost = _money(float(score["scoring_cost_usd"]))
                    total_cost = _money(actual_cost + scoring_cost)
                    record["scoring_cost_usd"] = scoring_cost
                    record["total_cost_usd"] = total_cost
                    if not policy.observe_cost_only and total_cost > policy.max_cost_per_run_usd:
                        deviations.append("PER_RUN_COST_BUDGET_BREACH")
            spent = _money(spent + float(record["total_cost_usd"] or 0.0))
            if not policy.observe_cost_only and spent > policy.max_total_cost_usd:
                deviations.append("TOTAL_COST_BUDGET_BREACH")
            if deviations:
                record["status"] = "invalid"
            record["protocol_deviation_codes"] = deviations
            append_generic_result(results_path, record)
            completed += 1
            if deviations or (not policy.observe_cost_only and spent > policy.max_total_cost_usd):
                break

    return {
        "study_id": args.study_id,
        "planned_runs": plan["planned_runs"],
        "completed_runs": completed,
        "selected_model": selected.model_id,
        "spent_cost_usd": spent,
        "dry_run": False,
    }


def repo_snapshot_sha256(repo: Path) -> str:
    result = subprocess.run(
        ["git", "-C", str(repo), "rev-parse", "HEAD"],
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )
    if result.returncode != 0:
        raise ValueError("repo must be a Git checkout with a readable HEAD")
    return hashlib.sha256(result.stdout.strip().encode("utf-8")).hexdigest()


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest="command", required=True)
    plan = sub.add_parser("plan", help="build a no-spend paired run plan")
    plan.add_argument("--repo", required=True)
    plan.add_argument("--cases", required=True)
    plan.add_argument("--pricing", required=True)
    plan.add_argument("--provider", default="codex")
    plan.add_argument("--model")
    plan.add_argument("--effort", choices=("low", "medium", "high", "xhigh", "max"), default="medium")
    plan.add_argument("--repeat", type=int, default=2)
    plan.add_argument("--max-runs", type=int, default=24)
    plan.add_argument("--budget-usd", type=float, default=5.0)
    plan.add_argument("--max-cost-per-run-usd", type=float, default=0.5)
    plan.add_argument("--max-tokens-per-run", type=int, default=250_000)
    plan.add_argument("--timeout-seconds", type=int, default=900)
    plan.add_argument("--max-tool-calls", type=int, default=6)
    plan.add_argument("--max-tool-result-bytes", type=int, default=131_072)
    plan.add_argument("--max-output-bytes", type=int, default=524_288)
    plan.add_argument("--estimated-input-tokens", type=int, default=100_000)
    plan.add_argument("--estimated-output-tokens", type=int, default=20_000)
    run = sub.add_parser("run", help="execute a budgeted paired study through an isolated runner")
    run.add_argument("--run", action="store_true", help="required live-spend acknowledgement")
    run.add_argument("--repo")
    run.add_argument("--cases")
    run.add_argument("--pricing")
    run.add_argument(
        "--runner",
        default=str(ROOT / "research" / "eval" / "skill-eval" / "harness" / "bin" / "run-blindbox.sh"),
    )
    run.add_argument("--score-command")
    run.add_argument(
        "--defer-scoring",
        action="store_true",
        help="collect valid private answers without per-arm scoring",
    )
    run.add_argument("--results")
    run.add_argument("--private-answer-manifest")
    run.add_argument("--private-answer-dir")
    run.add_argument("--study-id")
    run.add_argument("--provider", default="codex")
    run.add_argument("--model")
    run.add_argument("--effort", choices=("low", "medium", "high", "xhigh", "max"), default="medium")
    run.add_argument("--auth", choices=("env", "session"), default="session")
    run.add_argument("--repeat", type=int, default=2)
    run.add_argument("--random-seed", type=int, default=1)
    run.add_argument("--max-runs", type=int, default=24)
    run.add_argument("--budget-usd", type=float, default=5.0)
    run.add_argument(
        "--observe-cost-only",
        action="store_true",
        help="record token and dollar use without using either as an admission or validity wall",
    )
    run.add_argument("--max-cost-per-run-usd", type=float, default=0.5)
    run.add_argument("--max-tokens-per-run", type=int, default=250_000)
    run.add_argument("--timeout-seconds", type=int, default=900)
    run.add_argument("--max-tool-calls", type=int, default=6)
    run.add_argument("--max-tool-result-bytes", type=int, default=131_072)
    run.add_argument("--max-output-bytes", type=int, default=524_288)
    run.add_argument("--max-answer-words", type=int, default=250)
    run.add_argument("--estimated-input-tokens", type=int, default=100_000)
    run.add_argument("--estimated-output-tokens", type=int, default=20_000)
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    if args.command == "run" and not args.run:
        print("live execution requires --run", file=__import__("sys").stderr)
        return 2
    try:
        if args.command == "run":
            summary = run_study(args)
            print(json.dumps(summary, sort_keys=True, indent=2))
            return 0
        repo_snapshot_sha256(Path(args.repo))
        cases = load_cases(Path(args.cases))
        pricing = load_pricing(Path(args.pricing))
        policy = BudgetPolicy(
            max_runs=args.max_runs,
            max_total_cost_usd=args.budget_usd,
            max_cost_per_run_usd=args.max_cost_per_run_usd,
            max_tokens_per_run=args.max_tokens_per_run,
            timeout_seconds=args.timeout_seconds,
            max_tool_calls=args.max_tool_calls,
            max_tool_result_bytes=args.max_tool_result_bytes,
            max_output_bytes=args.max_output_bytes,
        )
        plan = build_plan(
            cases,
            pricing=pricing,
            provider=args.provider,
            repeats=args.repeat,
            policy=policy,
            estimated_input_tokens=args.estimated_input_tokens,
            estimated_output_tokens=args.estimated_output_tokens,
            model_id=args.model,
            reasoning_effort=args.effort,
        )
    except (OSError, RuntimeError, ValueError, json.JSONDecodeError) as exc:
        print(str(exc), file=__import__("sys").stderr)
        return 2
    print(json.dumps(plan, sort_keys=True, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
