#!/usr/bin/env python3
"""Controlled C3 agent-efficiency eval harness.

Default mode is dry-run: no agent CLI is executed. Use --run to spend tokens.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import shlex
import shutil
import subprocess
import tempfile
import time
from dataclasses import dataclass
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[1]
METRIC_BASIS = "tokens_total, turn_count, accuracy_score, elapsed_ms"
TOKEN_THRESHOLDS = {"soft": 120000, "upper": 180000, "no_go": 250000}
ADR_QUALITY_FLOOR = 0.8
ADR_QUALITY_CHECKS = (
    "owner_correct",
    "root_cause_specific",
    "decision_concrete",
    "pressure_response_specific",
    "component_delta_specific",
    "scope_bounded",
    "verification_executable",
    "implementation_recoverable",
    "alternatives_real",
    "risk_testable",
    "no_boilerplate_na",
    "stop_condition_met",
)


@dataclass(frozen=True)
class EvalCase:
    id: str
    title: str
    prompt: str
    accuracy_checks: tuple[str, ...]


@dataclass(frozen=True)
class AgentSpec:
    id: str
    command_template: tuple[str, ...]


@dataclass(frozen=True)
class PlanItem:
    agent_id: str
    case_id: str
    command: list[str]
    prompt: str
    dry_run: bool


@dataclass
class RunResult:
    agent_id: str
    case_id: str
    command: list[str]
    dry_run: bool
    exit_code: int
    elapsed_ms: int
    stdout: str
    stderr: str
    output_dir: str
    artifact_dir: str
    accuracy_checks: dict[str, bool]
    token_usage: dict[str, int] | None
    turn_count: int | None
    trace_metrics: dict[str, Any]


def default_cases() -> list[EvalCase]:
    marker = (
        "When done, write eval_result.json in repo root with keys "
        "summary, verified, artifacts, root_cause, design_change, pressure_response, adr_id as applicable. "
        "Keep output concise."
    )
    pressure_marker = (
        "Before creating an ADR, inspect Up Cap pressure and decide whether to decompose/split a component, "
        "move the concern upward, or explicitly justify staying additive; record this as pressure_response. "
        "Also record component_delta naming the new/split/extracted component or a no-delta justification. "
    )
    return [
        EvalCase(
            id="skill_task_session",
            title="C3 Skill Task Session",
            prompt=(
                "Use the local C3 project instructions and local C3 skill at skills/c3/SKILL.md. "
                "Start a C3 session to identify the owner of cli/cmd/list.go, summarize the relevant constraints, "
                "and propose one smallest next action. Do not edit source files. "
                + marker
            ),
            accuracy_checks=("has_eval_result", "verified", "mentions_artifacts"),
        ),
        EvalCase(
            id="skill_content_limit_adr",
            title="C3 Skill Change Session For List Content Limit",
            prompt=(
                "Use the local C3 project instructions and local C3 skill at skills/c3/SKILL.md. "
                "Handle this user request: make c3x list show limited content instead of showing all content. "
                "Trace the affected C3 entities and code ownership, create a valid ADR/design work order for the change, "
                "and stop before implementing source-code changes. Use the local built c3x only. "
                + pressure_marker
                + marker
            ),
            accuracy_checks=("has_eval_result", "verified", "has_adr_id", "mentions_design_change", "mentions_pressure_response", "mentions_component_delta"),
        ),
        EvalCase(
            id="adr_create",
            title="Create ADR Successfully",
            prompt=(
                "Token-budgeted C3 ADR task. Use local c3x only. Do not run broad searches. "
                "Create one valid ADR for reducing c3x token cost. Run at most: c3x schema adr; "
                "c3x add adr <slug> --file <body>; c3x check --include-adr --only <adr-id>. "
                + pressure_marker
                + marker
            ),
            accuracy_checks=("has_eval_result", "verified", "has_adr_id", "mentions_pressure_response", "mentions_component_delta"),
        ),
        EvalCase(
            id="task_session",
            title="Start Session To Accomplish Task",
            prompt=(
                "Token-budgeted C3 session. Do not edit source files. Do not run broad searches. "
                "Use only local c3x. Run at most: c3x lookup cli/cmd/list.go; "
                "c3x read c3-1 --section Goal; c3x read c3-112 --section Goal. "
                "Then identify owner, summarize constraints from those outputs, and propose one smallest next action. "
                + marker
            ),
            accuracy_checks=("has_eval_result", "verified", "mentions_artifacts"),
        ),
        EvalCase(
            id="debug_session",
            title="Start Session To Debug",
            prompt=(
                "Token-budgeted C3 debug session. Do not edit source files. Do not run broad searches. "
                "Find likely root cause for high c3x token usage. Use at most: c3x list; "
                "c3x lookup scripts/agent_efficiency_eval.py; c3x read c3-108 --section Goal; "
                "inspect scripts/agent_efficiency_eval.py only if needed. "
                + marker
            ),
            accuracy_checks=("has_eval_result", "verified", "mentions_root_cause"),
        ),
        EvalCase(
            id="system_design_change",
            title="Start Session To Make System Design Change",
            prompt=(
                "Token-budgeted C3 system-design session. Do not edit source files. Do not run broad searches. "
                "Design budgeted output modes for c3x. Use at most: c3x list; c3x read c3-108 --section Goal; "
                "c3x read c3-112 --section Goal. Produce design direction and verification plan. "
                + marker
            ),
            accuracy_checks=("has_eval_result", "verified", "mentions_design_change"),
        ),
    ]


def default_agents() -> list[AgentSpec]:
    return [
        AgentSpec("claude", tuple(_env_command("C3_EVAL_CLAUDE_CMD", "claude -p {prompt}"))),
        AgentSpec(
            "codex",
            tuple(
                _env_command(
                    "C3_EVAL_CODEX_CMD",
                    "codex --ask-for-approval never exec --json --sandbox workspace-write --skip-git-repo-check --ignore-user-config {prompt}",
                )
            ),
        ),
    ]


def _env_command(name: str, fallback: str) -> list[str]:
    return shlex.split(os.environ.get(name, fallback))


def build_plan(cases: list[EvalCase], agents: list[AgentSpec], dry_run: bool) -> list[PlanItem]:
    plan: list[PlanItem] = []
    for agent in agents:
        for case in cases:
            command = [part.format(prompt=case.prompt) for part in agent.command_template]
            plan.append(PlanItem(agent.id, case.id, command, case.prompt, dry_run))
    return plan


def run_plan_item(item: PlanItem, case: EvalCase, keep_workspace: bool) -> RunResult:
    started = time.monotonic()
    workspace = ""
    stdout = ""
    stderr = ""
    exit_code = 0
    tmp: tempfile.TemporaryDirectory[str] | None = None
    artifact_dir = tempfile.mkdtemp(prefix=f"c3-eval-artifacts-{item.agent_id}-{item.case_id}-")

    if item.dry_run:
        stdout = "dry-run: " + " ".join(shlex.quote(part) for part in item.command)
    else:
        tmp = tempfile.TemporaryDirectory(prefix=f"c3-eval-{item.agent_id}-{item.case_id}-")
        workspace = tmp.name
        _copy_controlled_workspace(Path(workspace))
        proc = subprocess.run(
            item.command,
            cwd=workspace,
            text=True,
            capture_output=True,
            timeout=int(os.environ.get("C3_EVAL_TIMEOUT_SECONDS", "900")),
            check=False,
        )
        stdout = proc.stdout
        stderr = proc.stderr
        exit_code = proc.returncode

    write_artifacts(Path(artifact_dir), item, stdout, stderr, workspace)
    elapsed_ms = int((time.monotonic() - started) * 1000)
    token_usage = extract_token_usage(stdout + "\n" + stderr)
    turn_count = extract_turn_count(stdout + "\n" + stderr)
    trace_metrics = extract_trace_metrics(stdout + "\n" + stderr)
    accuracy = (
        {check: False for check in case.accuracy_checks}
        if item.dry_run
        else evaluate_accuracy(Path(workspace) if workspace else None, case, stdout, stderr)
    )
    result = RunResult(
        agent_id=item.agent_id,
        case_id=item.case_id,
        command=item.command,
        dry_run=item.dry_run,
        exit_code=exit_code,
        elapsed_ms=elapsed_ms,
        stdout=stdout,
        stderr=stderr,
        output_dir=workspace,
        artifact_dir=artifact_dir,
        accuracy_checks=accuracy,
        token_usage=token_usage,
        turn_count=turn_count,
        trace_metrics=trace_metrics,
    )
    if tmp is not None and not keep_workspace:
        tmp.cleanup()
    elif tmp is not None:
        tmp.cleanup = lambda: None  # type: ignore[method-assign]
    return result


def write_artifacts(path: Path, item: PlanItem, stdout: str, stderr: str, workspace: str) -> None:
    path.mkdir(parents=True, exist_ok=True)
    (path / "stdout.jsonl").write_text(stdout)
    (path / "stderr.log").write_text(stderr)
    (path / "prompt.txt").write_text(item.prompt)
    (path / "command.json").write_text(json.dumps(item.command, indent=2))
    if workspace:
        result_path = Path(workspace) / "eval_result.json"
        if result_path.exists():
            shutil.copy2(result_path, path / "eval_result.json")
        diff = subprocess.run(
            ["git", "diff", "--no-ext-diff"],
            cwd=workspace,
            text=True,
            capture_output=True,
            check=False,
        )
        (path / "workspace_diff.patch").write_text(diff.stdout)


def _copy_controlled_workspace(dst: Path) -> None:
    ignore = shutil.ignore_patterns(
        ".git",
        ".cache",
        "skills/c3/bin/c3x-*-*-*",
        "__pycache__",
        "*.pyc",
        "node_modules",
    )
    for name in ["AGENTS.md", "CLAUDE.md", "README.md", "CHANGELOG.md", ".c3", "cli", "skills", "scripts"]:
        src = ROOT / name
        target = dst / name
        if src.is_dir():
            shutil.copytree(src, target, ignore=ignore)
        elif src.exists():
            shutil.copy2(src, target)
    subprocess.run(["bash", "scripts/build.sh"], cwd=dst, check=True, stdout=subprocess.DEVNULL)


def extract_token_usage(text: str) -> dict[str, int] | None:
    usage: dict[str, int] = {}
    for line in text.splitlines():
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            continue
        candidate = obj.get("usage") if isinstance(obj, dict) else None
        if not isinstance(candidate, dict):
            candidate = obj if isinstance(obj, dict) else None
        if not isinstance(candidate, dict):
            continue
        for key in (
            "input_tokens",
            "cached_input_tokens",
            "output_tokens",
            "reasoning_output_tokens",
            "cache_creation_input_tokens",
            "cache_read_input_tokens",
        ):
            value = candidate.get(key)
            if isinstance(value, int):
                usage[key] = usage.get(key, 0) + value
        total = candidate.get("total_tokens")
        if isinstance(total, int):
            usage["total_tokens"] = usage.get("total_tokens", 0) + total
    if not usage:
        return None
    if "total_tokens" not in usage:
        usage["total_tokens"] = usage.get("input_tokens", 0) + usage.get("output_tokens", 0)
    return usage


def extract_turn_count(text: str) -> int | None:
    patterns = [
        re.compile(r'"type"\s*:\s*"turn\.completed"'),
        re.compile(r"^\s*turn\s+\d+\b", re.IGNORECASE),
        re.compile(r'"type"\s*:\s*"(assistant|user|message)"'),
        re.compile(r"^\s*(assistant|user)\s*:", re.IGNORECASE),
    ]
    count = 0
    for line in text.splitlines():
        if any(pattern.search(line) for pattern in patterns):
            count += 1
    return count or None


def extract_trace_metrics(text: str) -> dict[str, Any]:
    c3_pattern = re.compile(
        r"^((?:C3X_MODE=agent\s+)?bash\s+skills/c3/bin/c3x\.sh\s+[^\n\"']+|"
        r"c3x\s+(?:lookup|read|list|check|schema|graph|add|write|set|wire|delete)\b[^\n\"']*)$"
    )
    broad_search_pattern = re.compile(r"^(?:rg\b|grep\s+-R\b|find\s+\.)")
    c3_commands = []
    broad_search_commands = []
    for command in trace_command_candidates(text):
        if command.startswith("|"):
            continue
        match = c3_pattern.match(command)
        if broad_search_pattern.match(command) and command not in broad_search_commands:
            broad_search_commands.append(command)
        if not match:
            continue
        command = match.group(1).strip()
        if command not in c3_commands:
            c3_commands.append(command)
    return {
        "json_event_count": sum(1 for line in text.splitlines() if line.strip().startswith("{")),
        "c3_command_count": len(c3_commands),
        "c3_command_sequence": c3_commands,
        "broad_search_count": len(broad_search_commands),
        "tool_output_bytes_total": len(text.encode()),
        "c3_output_bytes_total": sum(len(cmd.encode()) for cmd in c3_commands),
    }


def trace_command_candidates(text: str) -> list[str]:
    candidates: list[str] = []
    for line in text.splitlines():
        stripped = line.strip()
        if not stripped:
            continue
        try:
            obj = json.loads(stripped)
        except json.JSONDecodeError:
            candidates.append(stripped)
            continue
        item = obj.get("item") if isinstance(obj, dict) else None
        command = item.get("command") if isinstance(item, dict) else None
        if isinstance(command, str):
            candidates.append(normalize_shell_command(command))
    return candidates


def normalize_shell_command(command: str) -> str:
    try:
        parts = shlex.split(command)
    except ValueError:
        return command.strip()
    if len(parts) >= 3 and parts[0].endswith("bash") and parts[1] == "-lc":
        return parts[2].strip()
    return command.strip()


def evaluate_accuracy(workspace: Path | None, case: EvalCase, stdout: str, stderr: str) -> dict[str, bool]:
    text = (stdout + "\n" + stderr).lower()
    result: dict[str, Any] = {}
    if workspace:
        result_path = workspace / "eval_result.json"
        if result_path.exists():
            try:
                result = json.loads(result_path.read_text())
            except json.JSONDecodeError:
                result = {}

    checks: dict[str, bool] = {}
    for check in case.accuracy_checks:
        if check == "has_eval_result":
            checks[check] = bool(result)
        elif check == "verified":
            checks[check] = bool(result.get("verified")) or "verified" in text
        elif check == "has_adr_id":
            checks[check] = bool(result.get("adr_id")) or "adr-" in text
        elif check == "mentions_artifacts":
            checks[check] = bool(result.get("artifacts")) or "artifact" in text
        elif check == "mentions_root_cause":
            checks[check] = bool(result.get("root_cause")) or "root cause" in text
        elif check == "mentions_design_change":
            checks[check] = bool(result.get("design_change")) or "design change" in text
        elif check == "mentions_pressure_response":
            pressure_response = result.get("pressure_response")
            checks[check] = _specific_pressure_response(pressure_response) or _has_any(
                text,
                ("pressure_response", "up cap", "decompose", "split", "staying additive", "stay additive"),
            )
        elif check == "mentions_component_delta":
            component_delta = result.get("component_delta")
            checks[check] = _specific_component_delta(component_delta) or _has_any(
                text,
                ("component_delta", "new component", "split component", "extract component", "no-delta"),
            )
        else:
            checks[check] = False
    return checks


def evaluate_adr_quality(workspace: Path | None, transcript_text: str) -> dict[str, Any]:
    result = _read_eval_result(workspace)
    adr_text = _read_declared_adr_text(workspace, result)
    if not result and not adr_text:
        return {"score": 0.0, "checks": {check: False for check in ADR_QUALITY_CHECKS}}
    text = "\n".join(
        str(part)
        for part in (
            transcript_text,
            json.dumps(result, sort_keys=True),
            adr_text,
        )
    ).lower()

    checks = {
        "owner_correct": _has_any(text, ("cli/cmd/", "scripts/", "skills/", "c3-", "ref-", "rule-")),
        "root_cause_specific": _specific_text(result.get("root_cause")) or _has_any(text, ("root cause", "runlist", "liststructured", "agent-mode", "token", "output")),
        "decision_concrete": _specific_text(result.get("design_change")) and not _has_any(str(result.get("design_change", "")).lower(), ("best practices", "improve quality")),
        "pressure_response_specific": _specific_pressure_response(result.get("pressure_response")) or _has_any(text, ("up cap", "decompose", "split component", "split ownership", "stay additive", "staying additive")),
        "component_delta_specific": _specific_component_delta(result.get("component_delta")) or _has_any(text, ("new component", "split component", "extract component", "no-delta")),
        "scope_bounded": _has_any(text, ("stop before", "before implementation", "do not edit", "no source", "not change", "preserve")),
        "verification_executable": _has_any(text, ("c3x check", "go test", "python ", "pytest", "npm test", "bunx", "smoke")),
        "implementation_recoverable": _has_any(text, ("cli/cmd/", "scripts/", "skills/", "test", "update ", "add ", "change ")),
        "alternatives_real": _has_any(text, ("alternative", "rejected because", "instead of", "preserve")),
        "risk_testable": _has_any(text, ("risk", "failure", "regression", "detection", "mitigation")),
        "no_boilerplate_na": text.count("n.a -") <= 3 and not _has_any(text, ("best practices", "improve quality", "optimize performance")),
        "stop_condition_met": _has_any(text, ("stopped before", "stop before", "before source-code", "do not edit source")),
    }
    score = sum(1 for value in checks.values() if value) / len(checks)
    return {"score": round(score, 4), "checks": checks}


def _read_eval_result(workspace: Path | None) -> dict[str, Any]:
    if not workspace:
        return {}
    result_path = workspace / "eval_result.json"
    if not result_path.exists():
        return {}
    try:
        data = json.loads(result_path.read_text())
    except json.JSONDecodeError:
        return {}
    return data if isinstance(data, dict) else {}


def _read_declared_adr_text(workspace: Path | None, result: dict[str, Any]) -> str:
    if not workspace:
        return ""
    candidates: list[Path] = []
    for artifact in result.get("artifacts") or []:
        if isinstance(artifact, str):
            candidates.append(workspace / artifact)
            if artifact.startswith("adr-"):
                adr_name = artifact if artifact.endswith(".md") else f"{artifact}.md"
                candidates.extend(workspace.glob(f"**/{adr_name}"))
    adr_id = result.get("adr_id")
    if isinstance(adr_id, str) and adr_id:
        candidates.extend(workspace.glob(f"**/{adr_id}.md"))
    for candidate in candidates:
        if candidate.is_file():
            return candidate.read_text(errors="replace")
    return ""


def _has_any(text: str, needles: tuple[str, ...]) -> bool:
    return any(needle in text for needle in needles)


def _specific_text(value: Any) -> bool:
    if not isinstance(value, str):
        return False
    words = re.findall(r"[A-Za-z0-9_./:-]+", value)
    if len(words) < 6:
        return False
    generic = {"improve", "quality", "best", "practices", "optimize", "better"}
    return len({word.lower() for word in words} - generic) >= 4


def _specific_pressure_response(value: Any) -> bool:
    if not _specific_text(value):
        return False
    text = str(value).lower()
    has_pressure = _has_any(text, ("up cap", "pressure", "cap", "current load", "token"))
    has_decision = _has_any(text, ("decompose", "split", "extract", "move upward", "stay additive", "staying additive", "additive"))
    generic = _has_any(text, ("as needed", "best practices", "improve quality"))
    return has_pressure and has_decision and not generic


def _specific_component_delta(value: Any) -> bool:
    if not _specific_text(value):
        return False
    text = str(value).lower()
    has_component = _has_any(text, ("c3-", "component", "owner"))
    has_delta = _has_any(text, ("new", "split", "extract", "move", "no new", "no-delta", "keep"))
    generic = _has_any(text, ("as needed", "best practices", "improve quality", "change components as needed"))
    return has_component and has_delta and not generic


def evaluate_threshold_pressure(
    case_id: str,
    tokens_total: int | None,
    accuracy_score: float,
    adr_quality_score: float,
    trace_metrics: dict[str, Any],
) -> dict[str, Any]:
    if tokens_total is None:
        return {
            "status": "unknown",
            "action": "measure",
            "reasons": [],
            "target_tokens": None,
            "potential_savings": None,
        }

    reasons: list[str] = []
    if tokens_total >= TOKEN_THRESHOLDS["no_go"]:
        status = "no_go"
    elif tokens_total >= TOKEN_THRESHOLDS["upper"]:
        status = "upper"
    elif tokens_total >= TOKEN_THRESHOLDS["soft"]:
        status = "soft"
    else:
        status = "ok"

    if status != "ok":
        reasons.append(f"tokens_{status}")
    if (trace_metrics.get("broad_search_count") or 0) > 0:
        reasons.append("broad_search")
    if (trace_metrics.get("tool_output_bytes_total") or 0) >= 50000:
        reasons.append("tool_output_pressure")
    if accuracy_score < 0.66:
        reasons.append("accuracy_below_guard")
    if "adr" in case_id and tokens_total >= TOKEN_THRESHOLDS["soft"] and adr_quality_score < ADR_QUALITY_FLOOR:
        reasons.append("adr_quality_below_floor")

    if accuracy_score < 0.66:
        action = "fail_accuracy"
        target = tokens_total
    elif status == "no_go":
        action = "stop_or_split"
        target = TOKEN_THRESHOLDS["no_go"]
    elif "adr_quality_below_floor" in reasons:
        action = "fail_quality_or_split"
        target = TOKEN_THRESHOLDS["soft"]
    elif status == "upper":
        action = "require_cap_justification"
        target = TOKEN_THRESHOLDS["upper"]
    elif status == "soft":
        action = "warn_and_bound_scope"
        target = TOKEN_THRESHOLDS["soft"]
    else:
        action = "accept"
        target = tokens_total

    return {
        "status": status,
        "action": action,
        "reasons": reasons,
        "target_tokens": target,
        "potential_savings": max(0, tokens_total - target),
    }


def score_result(result: RunResult) -> dict[str, Any]:
    accuracy_values = list(result.accuracy_checks.values())
    accuracy_score = sum(1 for value in accuracy_values if value) / len(accuracy_values) if accuracy_values else 0.0
    turn_count = result.turn_count if result.turn_count is not None else extract_turn_count(result.stdout + "\n" + result.stderr)
    quality_workspace = Path(result.output_dir) if result.output_dir else None
    if quality_workspace is not None and not quality_workspace.exists():
        quality_workspace = Path(result.artifact_dir) if result.artifact_dir else None
    adr_quality = evaluate_adr_quality(quality_workspace, result.stdout + "\n" + result.stderr)
    tokens_total = result.token_usage.get("total_tokens") if result.token_usage else None
    pressure = evaluate_threshold_pressure(
        result.case_id,
        tokens_total,
        accuracy_score,
        adr_quality["score"],
        result.trace_metrics,
    )
    scored = {
        "agent": result.agent_id,
        "case": result.case_id,
        "dry_run": result.dry_run,
        "exit_code": result.exit_code,
        "elapsed_ms": result.elapsed_ms,
        "tokens_total": tokens_total,
        "turn_count": turn_count,
        "accuracy_score": accuracy_score,
        "accuracy_checks": result.accuracy_checks,
        "output_dir": result.output_dir,
        "artifact_dir": result.artifact_dir,
        "command": result.command,
        "metric_basis": METRIC_BASIS,
        "adr_quality_score": adr_quality["score"],
        "adr_quality_checks": adr_quality["checks"],
        "threshold_status": pressure["status"],
        "threshold_action": pressure["action"],
        "threshold_reasons": pressure["reasons"],
        "threshold_target_tokens": pressure["target_tokens"],
        "threshold_potential_savings": pressure["potential_savings"],
        **result.trace_metrics,
    }
    return scored


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--run", action="store_true", help="execute agent CLIs and spend tokens")
    parser.add_argument("--dry-run", action="store_true", help="print/write planned matrix only")
    parser.add_argument("--agent", choices=["claude", "codex"], action="append", help="limit agent")
    parser.add_argument("--case", choices=[case.id for case in default_cases()], action="append", help="limit case")
    parser.add_argument("--output", default="agent-efficiency-results.jsonl", help="JSONL output path")
    parser.add_argument("--keep-workspace", action="store_true", help="keep live temp workspaces")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    dry_run = not args.run
    cases = [case for case in default_cases() if not args.case or case.id in args.case]
    agents = [agent for agent in default_agents() if not args.agent or agent.id in args.agent]
    plan = build_plan(cases, agents, dry_run=dry_run)
    cases_by_id = {case.id: case for case in cases}
    output = Path(args.output)
    output.parent.mkdir(parents=True, exist_ok=True)

    with output.open("w") as fh:
        for item in plan:
            result = run_plan_item(item, cases_by_id[item.case_id], keep_workspace=args.keep_workspace)
            fh.write(json.dumps(score_result(result), sort_keys=True) + "\n")

    print(f"wrote {len(plan)} eval record(s) to {output}")
    if dry_run:
        print("dry-run only; pass --run to execute agent CLIs and spend tokens")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
