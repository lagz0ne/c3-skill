---
id: adr-20260502-agent-efficiency-eval
c3-seal: a86d45fab143346684854f2fd1563c028c8dc0aab3317373341dd887b76da2de
title: agent-efficiency-eval
type: adr
goal: Set up a controlled agent-efficiency eval harness for C3 tasks so Claude CLI and Codex CLI can be compared on token efficiency, turn count, and final task accuracy across repeatable scenarios.
status: proposed
date: "2026-05-02"
---

## Goal

Set up a controlled agent-efficiency eval harness for C3 tasks so Claude CLI and Codex CLI can be compared on token efficiency, turn count, and final task accuracy across repeatable scenarios.

## Context

Users report that C3 workflows can consume many tokens, but current tests mostly validate CLI authoring surfaces and do not measure full agent-session efficiency. The affected topology is the Go CLI repository support surface and uncharted eval/test tooling around `cli/cmd/llm_eval_test.go` and scripts. The harness must avoid accidental live agent spend by default while still supporting real Claude and Codex CLI execution when explicitly requested.

## Decision

Add a repository-local eval harness with declarative cases and agent definitions, dry-run planning by default, isolated temp workspaces for live runs, and JSONL result output. The initial supported agents are `claude` and `codex`; initial cases cover successful ADR creation, task session startup, debug session startup, and system design change session startup. Metrics are token efficiency, turn count, runtime, exit status, and accuracy scores from deterministic artifact checks.

## Affected Topology

| Entity | Type | Why affected | Governance review |
| --- | --- | --- | --- |
| c3-1 | container | Repository-level CLI/support tooling gains an eval harness that runs against C3 workflows and local c3x binaries. | Review parent responsibilities and record no parent-doc delta if harness remains support tooling. |
| c3-108 | component | Runtime-support owns agent/human presentation and CLI-mode behavior relevant to measuring agent CLI sessions. | Ensure harness does not change c3x runtime contract; use as adjacent ownership evidence. |
| N.A - uncharted eval files | N.A - <reason> | cli/cmd/llm_eval_test.go and new eval scripts are not mapped by codemap yet. | Keep harness files explicit in ADR and verification; do not claim component ownership beyond support scope. |

## Compliance Refs

| Ref | Why required | Action |
| --- | --- | --- |
| ref-cross-compiled-binary | Harness should use local built skills/c3/bin/c3x.sh and not assume global c3x binary behavior. | comply |
| N.A - no additional refs | No existing ref directly governs eval harness design. | N.A - no matching ref exists |

## Compliance Rules

| Rule | Why required | Action |
| --- | --- | --- |
| N.A - no rules present | Current topology has no rule-* entities governing script/test harness style. | N.A - no rules present |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Eval config | Add declarative cases and agent command templates for Claude/Codex. | New files under controlled eval path. |
| Eval runner | Add dry-run default and explicit live-run mode that creates isolated temp workspaces. | Python unittest plus dry-run command output. |
| Metrics | Capture tokens when present, turns from transcript shape, runtime, exit code, and deterministic accuracy checks. | JSONL schema assertions in tests. |
| Docs | Add concise usage docs for running dry-run and live agent comparisons. | README or script help output. |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| CLI files | No c3x command behavior changes planned. | go test ./... from cli/ remains green. |
| Eval scripts | Add repository-local harness that invokes local skills/c3/bin/c3x.sh in isolated workspaces. | Python unittest and dry-run output. |
| C3 docs | Add this ADR to document uncharted eval support work. | c3local check --include-adr --only <adr> then c3local check. |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| Python unittest | Verifies cases, agents, dry-run results, and scorer behavior. | python -m unittest ... |
| Eval dry-run | Lists planned Claude/Codex runs without invoking agent CLIs. | python scripts/agent_efficiency_eval.py --dry-run |
| Go tests | Ensures no existing CLI behavior regresses. | go test ./... from cli/ |
| C3 check | Validates ADR and canonical C3 state. | C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Hand-run Claude/Codex sessions and compare notes | Not controlled, not repeatable, and cannot separate prompt/case variance from agent behavior. |
| Build eval as a new c3x command now | Larger public API surface before the eval shape is proven; script harness is enough for first controlled loop. |
| Run live agents by default in tests | Would spend tokens and depend on authentication/network state, making CI and local verification unstable. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Live eval accidentally burns tokens | Default to dry-run; require explicit --run for live agents. | Unit test asserts default plan does not execute commands. |
| Accuracy scoring becomes subjective | Start with deterministic artifact checks per case and store raw transcript/result for later human review. | Unit test validates accuracy check names and score bounds. |
| Agent CLIs emit different metadata formats | Store raw stdout/stderr plus best-effort token/turn extraction; unknown token fields become null, not failure. | Unit test covers missing usage metadata. |
| Temp workspace leaks | Use temp directories and optional keep flag. | Unit test uses tempdir; live runner cleanup path inspected by dry-run. |

## Verification

| Check | Result |
| --- | --- |
| python -m unittest scripts.test_agent_efficiency_eval | PASS - 8 tests cover matrix shape, dry-run JSONL, scoring fallback, token/turn extraction, trace metrics, artifacts, and fake-live eval_result.json scoring before cleanup. |
| python scripts/agent_efficiency_eval.py --dry-run --output /tmp/c3-agent-eval-dryrun.jsonl | PASS - writes 12 records for 2 agents x 6 cases without executing agent CLIs. |
| python -m py_compile scripts/agent_efficiency_eval.py scripts/test_agent_efficiency_eval.py | PASS - Python syntax/import compilation succeeded. |
| cd cli && go test ./... | PASS - all Go packages passed. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr --only adr-20260502-agent-efficiency-eval | PASS - focused ADR check returned no issues. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check | PASS - structural check returned no issues. |
| git diff --check | PASS - no whitespace errors. |
