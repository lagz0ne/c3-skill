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
        "summary, verified, artifacts, root_cause, design_change, adr_id as applicable. "
        "Keep output concise."
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
                "When done, write eval_result.json in repo root with keys summary, verified, artifacts, root_cause, design_change, adr_id as applicable. "
                "Keep output concise."
            ),
            accuracy_checks=("has_eval_result", "verified", "has_adr_id", "mentions_design_change"),
        ),
        EvalCase(
            id="adr_create",
            title="Create ADR Successfully",
            prompt=(
                "Token-budgeted C3 ADR task. Use local c3x only. Do not run broad searches. "
                "Create one valid ADR for reducing c3x token cost. Run at most: c3x schema adr; "
                "c3x add adr <slug> --file <body>; c3x check --include-adr --only <adr-id>. "
                + marker
            ),
            accuracy_checks=("has_eval_result", "verified", "has_adr_id"),
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
        r"(?m)(?:^|[;&|]\s*)"
        r"((?:C3X_MODE=agent\s+)?bash\s+skills/c3/bin/c3x\.sh\s+[^\n\"']+|"
        r"c3x\s+(?:lookup|read|list|check|schema|graph|add|write|set|wire|delete)\b[^\n\"']*)"
    )
    c3_commands = []
    for match in c3_pattern.finditer(text):
        command = match.group(1).strip()
        if command not in c3_commands:
            c3_commands.append(command)
    return {
        "json_event_count": sum(1 for line in text.splitlines() if line.strip().startswith("{")),
        "c3_command_count": len(c3_commands),
        "c3_command_sequence": c3_commands,
        "broad_search_count": len(re.findall(r"\brg\s+|grep\s+-R|find\s+\.", text)),
        "tool_output_bytes_total": len(text.encode()),
        "c3_output_bytes_total": sum(len(cmd.encode()) for cmd in c3_commands),
    }


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
        else:
            checks[check] = False
    return checks


def score_result(result: RunResult) -> dict[str, Any]:
    accuracy_values = list(result.accuracy_checks.values())
    accuracy_score = sum(1 for value in accuracy_values if value) / len(accuracy_values) if accuracy_values else 0.0
    turn_count = result.turn_count if result.turn_count is not None else extract_turn_count(result.stdout + "\n" + result.stderr)
    return {
        "agent": result.agent_id,
        "case": result.case_id,
        "dry_run": result.dry_run,
        "exit_code": result.exit_code,
        "elapsed_ms": result.elapsed_ms,
        "tokens_total": result.token_usage.get("total_tokens") if result.token_usage else None,
        "turn_count": turn_count,
        "accuracy_score": accuracy_score,
        "accuracy_checks": result.accuracy_checks,
        "output_dir": result.output_dir,
        "artifact_dir": result.artifact_dir,
        "command": result.command,
        "metric_basis": METRIC_BASIS,
        **result.trace_metrics,
    }


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
