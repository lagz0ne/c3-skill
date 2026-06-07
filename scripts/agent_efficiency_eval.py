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
CANVAS_QUALITY_FLOOR = 0.9


@dataclass(frozen=True)
class CanvasExpectation:
    canvas_id: str
    artifact_name: str
    required_sections: tuple[str, ...]
    table_columns: dict[str, tuple[str, ...]]
    enum_columns: dict[tuple[str, str], tuple[str, ...]]
    cite_columns: tuple[tuple[str, str], ...]
    check_columns: tuple[tuple[str, str], ...]
    edge_columns: tuple[tuple[str, str], ...]


CANVAS_EXPECTATIONS: dict[str, CanvasExpectation] = {
    "canvas_c3_adr": CanvasExpectation(
        canvas_id="c3-adr",
        artifact_name="canvas-c3-adr.md",
        required_sections=("Goal", "Context", "Decision", "Affected Topology", "Compliance Refs", "Compliance Rules", "Work Breakdown", "Underlay C3 Changes", "Enforcement Surfaces", "Alternatives Considered", "Risks", "Verification"),
        table_columns={
            "Affected Topology": ("Entity", "Type", "Why affected", "Evidence", "Governance review"),
            "Compliance Refs": ("Ref", "Why required", "Evidence", "Action"),
            "Compliance Rules": ("Rule", "Why required", "Evidence", "Action"),
            "Verification": ("Check", "Result"),
        },
        enum_columns={("Affected Topology", "Type"): ("system", "container", "component", "N.A - <reason>")},
        cite_columns=(("Affected Topology", "Evidence"), ("Compliance Refs", "Evidence"), ("Compliance Rules", "Evidence")),
        check_columns=(),
        edge_columns=(),
    ),
    "canvas_atomic_design": CanvasExpectation(
        canvas_id="atomic-design-change",
        artifact_name="canvas-atomic-design-change.md",
        required_sections=("Goal", "Affected Units", "Change Record"),
        table_columns={
            "Affected Units": ("Unit", "Level", "Why affected", "Evidence"),
            "Change Record": ("Change", "Break risk", "Result", "Evidence"),
        },
        enum_columns={("Affected Units", "Level"): ("atom", "molecule", "organism", "template", "page", "N.A - <reason>")},
        cite_columns=(("Affected Units", "Evidence"), ("Change Record", "Evidence")),
        check_columns=(("Change Record", "Result"),),
        edge_columns=(),
    ),
    "canvas_pm_requirement": CanvasExpectation(
        canvas_id="pm-requirement",
        artifact_name="canvas-pm-requirement.md",
        required_sections=("Need", "Facts", "Acceptance"),
        table_columns={
            "Facts": ("Fact", "Evidence"),
            "Acceptance": ("Scenario", "Result", "Trace"),
        },
        enum_columns={},
        cite_columns=(("Facts", "Evidence"),),
        check_columns=(("Acceptance", "Result"),),
        edge_columns=(("Acceptance", "Trace"),),
    ),
    "canvas_prd": CanvasExpectation(
        canvas_id="prd",
        artifact_name="canvas-prd.md",
        required_sections=("Goal", "Requirements", "Story Traces"),
        table_columns={
            "Requirements": ("Requirement", "Priority", "Evidence"),
            "Story Traces": ("Story", "Status", "Evidence"),
        },
        enum_columns={("Requirements", "Priority"): ("must", "should", "could", "wont")},
        cite_columns=(("Requirements", "Evidence"), ("Story Traces", "Evidence")),
        check_columns=(("Story Traces", "Status"),),
        edge_columns=(("Story Traces", "Story"),),
    ),
    "canvas_user_story": CanvasExpectation(
        canvas_id="user-story",
        artifact_name="canvas-user-story.md",
        required_sections=("Story", "Acceptance", "Trace"),
        table_columns={
            "Acceptance": ("Criterion", "Result", "Evidence"),
            "Trace": ("Source", "Why derived", "Evidence"),
        },
        enum_columns={},
        cite_columns=(("Acceptance", "Evidence"), ("Trace", "Evidence")),
        check_columns=(("Acceptance", "Result"),),
        edge_columns=(("Trace", "Source"),),
    ),
}


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
    trial: int = 1


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
    trial: int = 1


def default_cases() -> list[EvalCase]:
    marker = (
        "When done, write eval_result.json in repo root with keys "
        "summary, verified, artifacts, root_cause, design_change, pressure_response, adr_id as applicable. "
        "Keep output concise."
    )
    pressure_marker = (
        "Before creating an ADR, inspect graph/governance pressure and decide whether to decompose/split a component, "
        "move the concern upward, or explicitly justify staying additive; record this as pressure_response. "
        "Also record component_delta naming the new/split/extracted component or a no-delta justification. "
    )
    canvas_marker = (
        "Use local C3 only. Run `C3X_MODE=agent bash skills/c3/bin/c3x.sh canvas read {canvas_id}` first. "
        "Write one markdown artifact at {artifact_name} that matches the canvas sections and table columns. "
        "Use `## <section name>` headings for every canvas section exactly as named by `canvas read`; do not use `#` for section headings. "
        "Use concrete non-placeholder content. For cite cells, paste an exact C3 citation handle from `c3x read --cite` "
        "or `c3x read <id> --section <section> --cite`; the handle must include `@v` and `:sha256:`. "
        "Use `N.A - <reason>` only when the source does not exist yet. For check cells use pass, fail, blocked, skipped, pending, or n_a. "
        "For edge cells use `relation:target-id`, `source->target`, a C3 id, or `N.A - <reason>`. "
        "When done, write eval_result.json with keys summary, verified, canvas_id, canvas_artifact, artifacts. "
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
                "Create one valid ADR for reducing c3x token cost. Before drafting, write a terse Discovery Brief "
                "from the task goal and targeted C3 reads: owner, governing material, and stop condition. "
                "This is ADR-only; stop before implementing source-code changes. Run at most: c3x schema adr; "
                "c3x lookup cli/cmd/list.go; c3x graph c3-112 --depth 1; c3x add adr <slug> --file <body>; "
                "c3x check --include-adr --only <adr-id>. Do not use rg/find/sed/cat or raw .c3 file reads. "
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
        EvalCase(
            id="canvas_c3_adr",
            title="Author C3 ADR Canvas File",
            prompt=canvas_marker.format(canvas_id="c3-adr", artifact_name="canvas-c3-adr.md")
            + "Theme: decide whether canvas validation belongs in check-cmd or docs-state-cmds.",
            accuracy_checks=("has_eval_result", "verified", "has_canvas_artifact", "canvas_score_90"),
        ),
        EvalCase(
            id="canvas_atomic_design",
            title="Author Atomic Design Change Canvas File",
            prompt=canvas_marker.format(canvas_id="atomic-design-change", artifact_name="canvas-atomic-design-change.md")
            + "Theme: change a button atom color token and trace impact to molecule, organism, template, and page levels.",
            accuracy_checks=("has_eval_result", "verified", "has_canvas_artifact", "canvas_score_90"),
        ),
        EvalCase(
            id="canvas_pm_requirement",
            title="Author PM Requirement Canvas File",
            prompt=canvas_marker.format(canvas_id="pm-requirement", artifact_name="canvas-pm-requirement.md")
            + "Theme: require saved filter sharing in a project dashboard.",
            accuracy_checks=("has_eval_result", "verified", "has_canvas_artifact", "canvas_score_90"),
        ),
        EvalCase(
            id="canvas_prd",
            title="Author PRD Canvas File",
            prompt=canvas_marker.format(canvas_id="prd", artifact_name="canvas-prd.md")
            + "Theme: define PRD requirements for workspace notification preferences.",
            accuracy_checks=("has_eval_result", "verified", "has_canvas_artifact", "canvas_score_90"),
        ),
        EvalCase(
            id="canvas_user_story",
            title="Author User Story Canvas File",
            prompt=canvas_marker.format(canvas_id="user-story", artifact_name="canvas-user-story.md")
            + "Theme: user can mute a noisy workspace channel from notification settings.",
            accuracy_checks=("has_eval_result", "verified", "has_canvas_artifact", "canvas_score_90"),
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


def build_plan(cases: list[EvalCase], agents: list[AgentSpec], dry_run: bool, repeat: int = 1) -> list[PlanItem]:
    plan: list[PlanItem] = []
    for agent in agents:
        for case in cases:
            for trial in range(1, repeat + 1):
                command = [part.format(prompt=case.prompt) for part in agent.command_template]
                plan.append(PlanItem(agent.id, case.id, command, case.prompt, dry_run, trial))
    return plan


def run_plan_item(item: PlanItem, case: EvalCase, keep_workspace: bool) -> RunResult:
    started = time.monotonic()
    workspace = ""
    stdout = ""
    stderr = ""
    exit_code = 0
    tmp: tempfile.TemporaryDirectory[str] | None = None
    artifact_dir = tempfile.mkdtemp(prefix=f"c3-eval-artifacts-{item.agent_id}-{item.case_id}-t{item.trial}-")

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
        trial=item.trial,
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
            copy_declared_artifacts(Path(workspace), path, result_path)
        diff = subprocess.run(
            ["git", "diff", "--no-ext-diff"],
            cwd=workspace,
            text=True,
            capture_output=True,
            check=False,
        )
        (path / "workspace_diff.patch").write_text(diff.stdout)


def copy_declared_artifacts(workspace: Path, artifact_dir: Path, result_path: Path) -> None:
    try:
        result = json.loads(result_path.read_text())
    except json.JSONDecodeError:
        return
    candidates: list[str] = []
    canvas_artifact = result.get("canvas_artifact")
    if isinstance(canvas_artifact, str):
        candidates.append(canvas_artifact)
    for artifact in result.get("artifacts") or []:
        if isinstance(artifact, str):
            candidates.append(artifact)
    seen: set[str] = set()
    for rel in candidates:
        rel_path = Path(rel)
        if rel_path.is_absolute() or ".." in rel_path.parts or str(rel_path) in seen:
            continue
        seen.add(str(rel_path))
        src = workspace / rel_path
        if not src.is_file():
            continue
        dst = artifact_dir / rel_path
        dst.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(src, dst)


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
    usage["effective_tokens"] = max(0, usage.get("input_tokens", 0) - usage.get("cached_input_tokens", 0)) + usage.get("output_tokens", 0)
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
        r"c3x\s+(?:lookup|read|list|check|schema|graph|add|write|set|wire|delete|canvas)\b[^\n\"']*)$"
    )
    broad_search_pattern = re.compile(r"^(?:rg\b|grep\s+-R\b|find\s+\.)")
    c3_commands = []
    broad_search_commands = []
    tool_output_bytes_total = 0
    c3_output_bytes_total = 0
    for event in trace_command_events(text):
        command = event["command"]
        if command.startswith("|"):
            continue
        match = c3_pattern.match(command)
        if broad_search_pattern.match(command) and command not in broad_search_commands:
            broad_search_commands.append(command)
        output = event.get("output")
        if output is not None:
            output_bytes = len(output.encode())
            tool_output_bytes_total += output_bytes
            if match:
                c3_output_bytes_total += output_bytes
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
        "tool_output_bytes_total": tool_output_bytes_total,
        "c3_output_bytes_total": c3_output_bytes_total,
        "transcript_bytes_total": len(text.encode()),
        "c3_command_bytes_total": sum(len(cmd.encode()) for cmd in c3_commands),
        "agent_unavailable": _agent_unavailable(text),
    }


def trace_command_candidates(text: str) -> list[str]:
    return [event["command"] for event in trace_command_events(text)]


def trace_command_events(text: str) -> list[dict[str, str | None]]:
    events: list[dict[str, str | None]] = []
    for line in text.splitlines():
        stripped = line.strip()
        if not stripped:
            continue
        try:
            obj = json.loads(stripped)
        except json.JSONDecodeError:
            events.append({"command": stripped, "output": None})
            continue
        item = obj.get("item") if isinstance(obj, dict) else None
        command = item.get("command") if isinstance(item, dict) else None
        if isinstance(command, str):
            output = item.get("aggregated_output")
            events.append({
                "command": normalize_shell_command(command),
                "output": output if isinstance(output, str) else "",
            })
    return events


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
            checks[check] = _specific_pressure_response(result.get("pressure_response"))
        elif check == "mentions_component_delta":
            component_delta = result.get("component_delta")
            checks[check] = _specific_component_delta(component_delta) or _has_any(
                text,
                ("component_delta", "new component", "split component", "extract component", "no-delta"),
            )
        elif check == "has_canvas_artifact":
            checks[check] = _canvas_artifact_path(workspace, result, case.id) is not None
        elif check == "canvas_score_90":
            checks[check] = canvas_quality_passes(evaluate_canvas_quality(workspace, result, case.id))
        else:
            checks[check] = False
    return checks


def evaluate_canvas_quality(workspace: Path | None, result: dict[str, Any], case_id: str) -> dict[str, Any]:
    expectation = CANVAS_EXPECTATIONS.get(case_id)
    checks = {
        "has_canvas_artifact": False,
        "canvas_id_matches": False,
        "required_sections_present": False,
        "tables_have_required_columns": False,
        "tables_have_rows": False,
        "cite_cells_grounded": False,
        "check_cells_valid": False,
        "edge_cells_nonempty": False,
        "enum_cells_valid": False,
        "no_placeholder_text": False,
    }
    if expectation is None:
        return {"score": 0.0, "checks": checks, "artifact": None}

    artifact = _canvas_artifact_path(workspace, result, case_id)
    if artifact is None:
        return {"score": 0.0, "checks": checks, "artifact": None}
    text = artifact.read_text(errors="replace")
    sections = parse_markdown_sections(text)
    tables = {name: parse_first_markdown_table(body) for name, body in sections.items()}

    checks["has_canvas_artifact"] = True
    checks["canvas_id_matches"] = result.get("canvas_id") == expectation.canvas_id or expectation.canvas_id in text
    checks["required_sections_present"] = all(section in sections and sections[section].strip() for section in expectation.required_sections)
    checks["tables_have_required_columns"] = all(
        table is not None and all(column in table["headers"] for column in columns)
        for section, columns in expectation.table_columns.items()
        for table in [tables.get(section)]
    )
    checks["tables_have_rows"] = all(
        table is not None and len(table["rows"]) > 0
        for section, table in tables.items()
        if section in expectation.table_columns
    )
    checks["cite_cells_grounded"] = _canvas_cells_pass(expectation.cite_columns, tables, _grounded_cite)
    checks["check_cells_valid"] = _canvas_cells_pass(expectation.check_columns, tables, _valid_check_result)
    checks["edge_cells_nonempty"] = _canvas_cells_pass(expectation.edge_columns, tables, _valid_edge)
    checks["enum_cells_valid"] = all(
        _canvas_cells_pass(((section, column),), tables, lambda value, allowed=allowed: _valid_enum(value, allowed))
        for (section, column), allowed in expectation.enum_columns.items()
    )
    checks["no_placeholder_text"] = not _has_any(text.lower(), ("tbd", "todo", "lorem", "as needed", "best practices"))

    score = sum(1 for value in checks.values() if value) / len(checks)
    return {"score": round(score, 4), "checks": checks, "artifact": str(artifact)}


def canvas_quality_passes(quality: dict[str, Any]) -> bool:
    checks = quality.get("checks") if isinstance(quality, dict) else None
    return (
        isinstance(checks, dict)
        and (quality.get("score") or 0.0) >= CANVAS_QUALITY_FLOOR
        and all(bool(value) for value in checks.values())
    )


def _canvas_artifact_path(workspace: Path | None, result: dict[str, Any], case_id: str) -> Path | None:
    expectation = CANVAS_EXPECTATIONS.get(case_id)
    if workspace is None or expectation is None:
        return None
    candidates: list[Path] = []
    declared = result.get("canvas_artifact")
    if isinstance(declared, str) and declared:
        candidates.append(workspace / declared)
    for artifact in result.get("artifacts") or []:
        if isinstance(artifact, str):
            candidates.append(workspace / artifact)
    candidates.append(workspace / expectation.artifact_name)
    for candidate in candidates:
        if candidate.is_file():
            return candidate
    return None


def parse_markdown_sections(text: str) -> dict[str, str]:
    sections: dict[str, list[str]] = {}
    current: str | None = None
    for line in text.splitlines():
        if line.startswith("## ") and not line.startswith("### "):
            current = line[3:].strip()
            sections[current] = []
            continue
        if current is not None:
            sections[current].append(line)
    return {name: "\n".join(lines).strip() for name, lines in sections.items()}


def parse_first_markdown_table(text: str) -> dict[str, Any] | None:
    lines = [line.strip() for line in text.splitlines() if line.strip().startswith("|")]
    if len(lines) < 3:
        return None
    headers = _split_markdown_row(lines[0])
    separator = _split_markdown_row(lines[1])
    if not headers or not all(set(cell.replace(":", "").strip()) <= {"-"} for cell in separator):
        return None
    rows = []
    for line in lines[2:]:
        cells = _split_markdown_row(line)
        if len(cells) != len(headers):
            continue
        rows.append(dict(zip(headers, cells)))
    return {"headers": headers, "rows": rows}


def _split_markdown_row(line: str) -> list[str]:
    line = line.strip()
    if line.startswith("|"):
        line = line[1:]
    if line.endswith("|"):
        line = line[:-1]
    return [cell.strip() for cell in line.split("|")]


def _canvas_cells_pass(columns: tuple[tuple[str, str], ...], tables: dict[str, dict[str, Any] | None], predicate: Any) -> bool:
    if not columns:
        return True
    for section, column in columns:
        table = tables.get(section)
        if table is None or not table["rows"]:
            return False
        for row in table["rows"]:
            if not predicate(row.get(column, "")):
                return False
    return True


def _grounded_cite(value: str) -> bool:
    value = value.strip().strip('"')
    if value.startswith("N.A -") and len(value.split()) >= 4:
        return True
    return "@v" in value and ":sha256:" in value


def _valid_check_result(value: str) -> bool:
    return value.strip().lower() in {"pass", "fail", "blocked", "skipped", "pending", "n_a", "n.a - not applicable"}


def _valid_edge(value: str) -> bool:
    value = value.strip()
    if value == "" or value.startswith("TBD"):
        return False
    return (
        ":" in value
        or "->" in value
        or "#" in value
        or value.startswith("N.A -")
        or re.fullmatch(r"[A-Za-z][A-Za-z0-9_.-]*-[A-Za-z0-9_.-]+", value) is not None
    )


def _valid_enum(value: str, allowed: tuple[str, ...]) -> bool:
    value = value.strip()
    return value in allowed or (value.startswith("N.A -") and any(item.startswith("N.A -") for item in allowed))


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
        "pressure_response_specific": _specific_pressure_response(result.get("pressure_response")),
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
    has_pressure = _has_any(text, ("governance pressure", "graph pressure", "pressure", "current load", "token"))
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


def _agent_unavailable(text: str) -> bool:
    return _has_any(
        text.lower(),
        (
            "failed to authenticate",
            "invalid authentication credentials",
            "api error: 401",
            "command not found",
        ),
    )


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
    eval_result = _read_eval_result(quality_workspace)
    canvas_quality = evaluate_canvas_quality(quality_workspace, eval_result, result.case_id)
    tokens_total = result.token_usage.get("total_tokens") if result.token_usage else None
    effective_tokens = result.token_usage.get("effective_tokens") if result.token_usage else None
    pressure = evaluate_threshold_pressure(
        result.case_id,
        effective_tokens,
        accuracy_score,
        adr_quality["score"],
        result.trace_metrics,
    )
    scored = {
        "agent": result.agent_id,
        "case": result.case_id,
        "trial": result.trial,
        "dry_run": result.dry_run,
        "exit_code": result.exit_code,
        "elapsed_ms": result.elapsed_ms,
        "tokens_total": tokens_total,
        "effective_tokens_total": effective_tokens,
        "turn_count": turn_count,
        "accuracy_score": accuracy_score,
        "accuracy_checks": result.accuracy_checks,
        "output_dir": result.output_dir,
        "artifact_dir": result.artifact_dir,
        "command": result.command,
        "metric_basis": METRIC_BASIS,
        "adr_quality_score": adr_quality["score"],
        "adr_quality_checks": adr_quality["checks"],
        "canvas_quality_score": canvas_quality["score"],
        "canvas_quality_passed": canvas_quality_passes(canvas_quality),
        "canvas_quality_checks": canvas_quality["checks"],
        "canvas_artifact": canvas_quality["artifact"],
        "threshold_status": pressure["status"],
        "threshold_action": pressure["action"],
        "threshold_reasons": pressure["reasons"],
        "threshold_target_tokens": pressure["target_tokens"],
        "threshold_potential_savings": pressure["potential_savings"],
        **result.trace_metrics,
    }
    return scored


def summarize_records(records: list[dict[str, Any]]) -> dict[str, Any]:
    canvas_records = [
        record
        for record in records
        if record.get("case") in CANVAS_EXPECTATIONS and not record.get("dry_run")
    ]
    available_canvas_records = [record for record in canvas_records if not record.get("agent_unavailable")]
    unavailable_canvas_records = [record for record in canvas_records if record.get("agent_unavailable")]
    canvas_passes = [
        record
        for record in available_canvas_records
        if record.get("exit_code") == 0
        and bool(record.get("canvas_quality_passed"))
        and (record.get("accuracy_score") or 0.0) >= 1.0
    ]
    pass_rate = len(canvas_passes) / len(available_canvas_records) if available_canvas_records else 0.0
    by_case: dict[str, dict[str, Any]] = {}
    for case_id in CANVAS_EXPECTATIONS:
        case_records = [record for record in available_canvas_records if record.get("case") == case_id]
        case_passes = [record for record in canvas_passes if record.get("case") == case_id]
        by_case[case_id] = {
            "records": len(case_records),
            "passes": len(case_passes),
            "pass_rate": round(len(case_passes) / len(case_records), 4) if case_records else 0.0,
        }
    by_agent: dict[str, dict[str, Any]] = {}
    for agent_id in sorted({str(record.get("agent")) for record in canvas_records if record.get("agent")}):
        agent_records = [record for record in canvas_records if record.get("agent") == agent_id]
        available_agent_records = [record for record in agent_records if not record.get("agent_unavailable")]
        agent_passes = [record for record in canvas_passes if record.get("agent") == agent_id]
        by_agent[agent_id] = {
            "records": len(available_agent_records),
            "unavailable": len(agent_records) - len(available_agent_records),
            "passes": len(agent_passes),
            "pass_rate": round(len(agent_passes) / len(available_agent_records), 4) if available_agent_records else 0.0,
        }
    return {
        "record_count": len(records),
        "canvas_record_count": len(available_canvas_records),
        "canvas_unavailable_count": len(unavailable_canvas_records),
        "canvas_pass_count": len(canvas_passes),
        "canvas_pass_rate": round(pass_rate, 4),
        "canvas_target": CANVAS_QUALITY_FLOOR,
        "canvas_gate_passed": bool(available_canvas_records) and pass_rate >= CANVAS_QUALITY_FLOOR,
        "canvas_all_agents_available": len(unavailable_canvas_records) == 0,
        "canvas_by_case": by_case,
        "canvas_by_agent": by_agent,
    }


def load_jsonl_records(path: Path) -> list[dict[str, Any]]:
    records: list[dict[str, Any]] = []
    for line_number, line in enumerate(path.read_text().splitlines(), start=1):
        if not line.strip():
            continue
        try:
            record = json.loads(line)
        except json.JSONDecodeError as err:
            raise ValueError(f"{path}:{line_number}: invalid JSONL record: {err}") from err
        if not isinstance(record, dict):
            raise ValueError(f"{path}:{line_number}: JSONL record must be an object")
        records.append(record)
    return records


def write_summary_if_requested(summary: dict[str, Any], path: str | None) -> None:
    if not path:
        return
    summary_path = Path(path)
    summary_path.parent.mkdir(parents=True, exist_ok=True)
    summary_path.write_text(json.dumps(summary, indent=2, sort_keys=True) + "\n")


def print_canvas_summary(summary: dict[str, Any], output: Path | None = None) -> None:
    if output is not None:
        print(f"wrote {summary['record_count']} eval record(s) to {output}")
    print(
        "canvas pass rate: "
        f"{summary['canvas_pass_count']}/{summary['canvas_record_count']} "
        f"({summary['canvas_pass_rate']:.0%}); target {summary['canvas_target']:.0%}"
    )
    if summary["canvas_unavailable_count"]:
        print(f"canvas unavailable records: {summary['canvas_unavailable_count']}")


def gate_exit_code(
    summary: dict[str, Any],
    require_canvas_90: bool,
    require_canvas_agent_availability: bool,
    min_canvas_records: int = 0,
    min_canvas_records_per_case: int = 0,
    require_canvas_agents: list[str] | None = None,
) -> int:
    if require_canvas_90 and not summary["canvas_gate_passed"]:
        return 1
    if require_canvas_agent_availability and not summary["canvas_all_agents_available"]:
        return 1
    if min_canvas_records and summary["canvas_record_count"] < min_canvas_records:
        return 1
    if min_canvas_records_per_case:
        for case_summary in summary["canvas_by_case"].values():
            if case_summary["records"] < min_canvas_records_per_case:
                return 1
    for agent_id in require_canvas_agents or []:
        agent_summary = summary["canvas_by_agent"].get(agent_id)
        if agent_summary is None or agent_summary["records"] == 0 or agent_summary["unavailable"] > 0:
            return 1
    return 0


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--run", action="store_true", help="execute agent CLIs and spend tokens")
    parser.add_argument("--dry-run", action="store_true", help="print/write planned matrix only")
    parser.add_argument("--from-results", help="summarize and gate an existing JSONL result file without running agents")
    parser.add_argument("--repeat", type=int, default=1, help="repeat each selected agent/case this many times")
    parser.add_argument("--agent", choices=["claude", "codex"], action="append", help="limit agent")
    parser.add_argument("--case", choices=[case.id for case in default_cases()], action="append", help="limit case")
    parser.add_argument("--output", default="agent-efficiency-results.jsonl", help="JSONL output path")
    parser.add_argument("--summary", help="optional JSON summary output path")
    parser.add_argument("--require-canvas-90", action="store_true", help="exit nonzero unless live canvas records pass the 90 percent gate")
    parser.add_argument("--require-canvas-agent-availability", action="store_true", help="exit nonzero if any requested live canvas agent is unavailable")
    parser.add_argument("--require-canvas-agent", choices=["claude", "codex"], action="append", help="exit nonzero unless this agent has available live canvas records and no unavailable rows")
    parser.add_argument("--min-canvas-records", type=int, default=0, help="exit nonzero unless at least this many available live canvas records are present")
    parser.add_argument("--min-canvas-records-per-case", type=int, default=0, help="exit nonzero unless every built-in canvas case has this many available live records")
    parser.add_argument("--keep-workspace", action="store_true", help="keep live temp workspaces")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    if args.from_results:
        records = load_jsonl_records(Path(args.from_results))
        summary = summarize_records(records)
        write_summary_if_requested(summary, args.summary)
        print_canvas_summary(summary)
        return gate_exit_code(
            summary,
            args.require_canvas_90,
            args.require_canvas_agent_availability,
            args.min_canvas_records,
            args.min_canvas_records_per_case,
            args.require_canvas_agent,
        )

    dry_run = not args.run
    if args.repeat < 1:
        raise ValueError("--repeat must be >= 1")
    if args.min_canvas_records < 0:
        raise ValueError("--min-canvas-records must be >= 0")
    if args.min_canvas_records_per_case < 0:
        raise ValueError("--min-canvas-records-per-case must be >= 0")
    cases = [case for case in default_cases() if not args.case or case.id in args.case]
    agents = [agent for agent in default_agents() if not args.agent or agent.id in args.agent]
    plan = build_plan(cases, agents, dry_run=dry_run, repeat=args.repeat)
    cases_by_id = {case.id: case for case in cases}
    output = Path(args.output)
    output.parent.mkdir(parents=True, exist_ok=True)

    scored_records = []
    with output.open("w") as fh:
        for item in plan:
            result = run_plan_item(item, cases_by_id[item.case_id], keep_workspace=args.keep_workspace)
            scored = score_result(result)
            scored_records.append(scored)
            fh.write(json.dumps(scored, sort_keys=True) + "\n")

    summary = summarize_records(scored_records)
    write_summary_if_requested(summary, args.summary)
    print_canvas_summary(summary, output)
    if dry_run:
        print("dry-run only; pass --run to execute agent CLIs and spend tokens")
    return gate_exit_code(
        summary,
        args.require_canvas_90,
        args.require_canvas_agent_availability,
        args.min_canvas_records,
        args.min_canvas_records_per_case,
        args.require_canvas_agent,
    )


if __name__ == "__main__":
    raise SystemExit(main())
