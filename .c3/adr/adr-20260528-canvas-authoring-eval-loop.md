---
id: adr-20260528-canvas-authoring-eval-loop
c3-seal: e657630b5b5a493689f0b0553b6573e52ceb55e60024ee5376a0bd2d6c91f525
title: canvas-authoring-eval-loop
type: adr
goal: Define the first canvas authoring eval loop so C3 can measure whether agents can write files matching canvas expectations and use a 90 percent mechanical success gate.
status: proposed
date: "2026-05-28"
template: implementation-change
---

## Goal

Define the first canvas authoring eval loop so C3 can measure whether agents can write files matching canvas expectations and use a 90 percent mechanical success gate.

## Context

The generic canvas CLI now exposes built-in canvases for software ADRs, atomic design changes, PM requirements, PRDs, and user stories. That proves discoverability, but not authoring success. The active goal requires an eval loop that can later run agents and prove they can produce matching files at a high success rate. The existing agent efficiency harness measured C3 sessions and ADR quality; it did not score canvas-shaped artifacts.

## Decision

Extend `scripts/agent_efficiency_eval.py` with five canvas authoring cases, a mechanical canvas scorer, and an aggregate 90 percent pass-rate gate. Each case instructs the agent to read the canvas definition first, write one named markdown artifact, and report it through `eval_result.json`. The scorer checks required sections, required table columns, table rows, cite/check/edge/enum cell semantics, placeholder avoidance, and matching canvas id. Canvas cases pass only when every mechanical canvas check passes; the run-level gate passes when at least 90 percent of available live canvas records pass. Agent authentication/tool availability failures are reported separately and excluded from the canvas concept denominator.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-108 | component | Owns the agent/runtime support surface where the eval harness and its tests are now mapped and scored. | c3-108#n1633@v1:sha256:ae80704ae7172ccccc82f6ba7b67f4fe434e41a8e3571e164a7b2165e4e4f06b "Provide CLI bootstrap, option parsing, output shaping, config resolution, and human/agent presentation helpers." | Review eval prompts, scoring metrics, code-map ownership, and test coverage. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| N.A - no existing ref governs canvas eval scoring. | N.A - this slice creates the first scorer contract for canvas authoring. | N.A - no existing ref governs this slice. | N.A - no ref update in this slice. |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| N.A - no existing rule governs canvas eval scoring. | N.A - enforcement is by script tests and the eval scorer itself. | N.A - no existing rule governs this slice. | N.A - no rule update in this slice. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Eval cases | Add five canvas authoring cases for c3-adr, atomic design change, PM requirement, PRD, and user story. | python scripts/test_agent_efficiency_eval.py |
| Mechanical scorer | Add CanvasExpectation and evaluate_canvas_quality checks for sections, tables, cite/check/edge/enum cells, and placeholder avoidance. | python scripts/test_agent_efficiency_eval.py |
| Score output | Include canvas_quality_score, canvas_quality_checks, and canvas_artifact in scored eval records, and copy declared artifacts before workspace cleanup. | python scripts/test_agent_efficiency_eval.py |
| Aggregate gate | Summarize live canvas records with pass count, unavailable count, per-case rate, and --require-canvas-90 nonzero exit on failure. | python scripts/test_agent_efficiency_eval.py |
| Live proof | Run five Codex canvas cases across c3-adr, atomic design change, PM requirement, PRD, and user story; all available records passed. Claude rows were reported as unavailable due API 401 auth failure, not counted as canvas concept failures. | python scripts/agent_efficiency_eval.py --run --case canvas_c3_adr --case canvas_atomic_design --case canvas_pm_requirement --case canvas_prd --case canvas_user_story --output /tmp/canvas-eval-live-matrix-current.jsonl --summary /tmp/canvas-eval-live-matrix-current-summary.json --require-canvas-90 |
| Provider availability proof | Add a separate strict provider-availability gate so skipped providers cannot be hidden behind the available-agent quality rate. | python scripts/agent_efficiency_eval.py --run --agent claude --case canvas_prd --output /tmp/canvas-eval-claude-availability.jsonl --summary /tmp/canvas-eval-claude-availability-summary.json --require-canvas-agent-availability |
| Required provider proof | Add named required-provider gates so preserved evidence can prove that specific agents are present, or fail when a required provider is absent from the result set. | python scripts/agent_efficiency_eval.py --from-results /tmp/canvas-eval-codex-repeat2-v3.jsonl --summary /tmp/canvas-eval-codex-repeat2-v3-required-both-summary.json --require-canvas-90 --min-canvas-records 10 --min-canvas-records-per-case 2 --require-canvas-agent-availability --require-canvas-agent codex --require-canvas-agent claude |
| Replayable proof | Existing live JSONL can be summarized and gated without rerunning expensive agents, so stricter gates can be applied to preserved evidence. | python scripts/agent_efficiency_eval.py --from-results /tmp/canvas-eval-live-matrix-current.jsonl --summary /tmp/canvas-eval-live-matrix-current-strict-summary.json --require-canvas-90 --require-canvas-agent-availability |
| Repeated proof | Repeat support runs each selected canvas more than once and can require both total available records and per-canvas available records; current Codex proof is 10/10 across two trials of each built-in canvas. | python scripts/agent_efficiency_eval.py --run --agent codex --case canvas_c3_adr --case canvas_atomic_design --case canvas_pm_requirement --case canvas_prd --case canvas_user_story --repeat 2 --output /tmp/canvas-eval-codex-repeat2-v3.jsonl --summary /tmp/canvas-eval-codex-repeat2-v3-summary.json --require-canvas-90 --min-canvas-records 10 --require-canvas-agent-availability |
| Code map | Map eval harness scripts to c3-108 ownership. | c3x lookup scripts/agent_efficiency_eval.py |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| scripts/agent_efficiency_eval.py | Adds canvas case matrix, quality floor, artifact discovery/copying, markdown table parsing, canvas quality scoring, live pass-rate summary, replay-from-JSONL mode, repeated trials, total and per-canvas minimum record gates, named required-agent gates, per-agent availability summary, unavailable-agent classification, and strict canvas pass logic. | python scripts/test_agent_efficiency_eval.py |
| scripts/test_agent_efficiency_eval.py | Adds tests for canvas cases, trace metrics, scoring pass/fail behavior, artifact persistence, score output fields, aggregate gate math, unavailable agents, provider availability gating, replay mode, repeated trials, total and per-canvas minimum record gates, named required-agent gates, current citation handles, and edge ids. | python scripts/test_agent_efficiency_eval.py |
| scripts/build.sh | Disables Go VCS stamping so copied eval workspaces without .git can rebuild the local C3 wrapper. | bash scripts/build.sh --version dev |
| .c3/code-map.yaml | Maps eval harness scripts to c3-108. | c3x lookup scripts/test_agent_efficiency_eval.py |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| scripts/agent_efficiency_eval.py --dry-run | Matrix includes canvas authoring cases without spending agent tokens. | python scripts/test_agent_efficiency_eval.py |
| score_result | Emits canvas_quality_score and related check details for every eval record. | test_score_result_includes_canvas_quality_metric |
| evaluate_accuracy | Canvas cases require an artifact and every mechanical canvas check to pass. | test_canvas_cases_require_canvas_score |
| summarize_records | Live canvas matrix fails below 90 percent available-record pass rate and reports unavailable agents separately by provider. | test_summarize_records_requires_canvas_90_percent_pass_rate |
| --require-canvas-agent-availability | Full provider-coverage runs fail when any requested canvas agent is unavailable. | test_require_canvas_agent_availability_fails_on_unavailable_provider |
| --require-canvas-agent | Full provider-coverage proof can require specific agents to appear with available records instead of accepting a single-agent result set. | test_from_results_can_require_named_canvas_agents |
| --from-results | Preserved JSONL eval evidence can be replayed through current summary and gate policy without rerunning agents. | test_from_results_replays_summary_and_gates_without_running_agents |
| --repeat / --min-canvas-records / --min-canvas-records-per-case | The harness can demand repeated available records for the whole canvas set and for each built-in canvas case, rather than accepting one lucky pass or a lopsided sample. | test_repeat_expands_matrix_with_trial_numbers; test_from_results_can_require_minimum_canvas_records; test_from_results_can_require_minimum_records_per_canvas_case |
| trace metrics | c3x canvas commands count as C3 commands. | test_extract_trace_metrics_finds_c3_commands |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Wait for live agent runs before adding scorer. | Without a deterministic scorer, live runs cannot prove the 90 percent target. |
| Score by transcript text only. | Transcript mentions do not prove the artifact has required sections, table columns, or primitives. |
| Add full YAML or markdown AST dependency. | A small markdown parser is enough for this eval slice and keeps the harness dependency-free. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Scorer overfits current built-ins. | Expectations are explicit per built-in and can be extended when canvas grammar matures. | Positive and negative PRD/user-story scorer tests. |
| Agents pass shape but produce shallow text. | Scorer rejects placeholders and requires rows plus grounded primitive cells; deeper semantic scoring can be next. | test_evaluate_canvas_quality_penalizes_missing_primitives |
| Claude CLI auth is unavailable in this environment. | Classify authentication/tool failures separately from canvas concept failures; use --require-canvas-agent-availability when the proof must include every requested provider, and rerun Claude rows when credentials are fixed. | Live matrix reports canvas_unavailable_count: 5 and available Codex canvas pass rate 5/5; Claude-only availability gate exits 1. |
| Passing numeric score can hide one failed primitive. | Canvas pass now requires all mechanical checks, while the numeric score remains diagnostic. | test_score_result_includes_canvas_quality_metric |

## Verification

| Check | Result |
| --- | --- |
| python scripts/test_agent_efficiency_eval.py | pass |
| python scripts/agent_efficiency_eval.py --dry-run --output /tmp/canvas-eval.jsonl | pass |
| python scripts/agent_efficiency_eval.py --run --case canvas_c3_adr --case canvas_atomic_design --case canvas_pm_requirement --case canvas_prd --case canvas_user_story --output /tmp/canvas-eval-live-matrix-current.jsonl --summary /tmp/canvas-eval-live-matrix-current-summary.json --require-canvas-90 | pass: 5/5 available canvas records, 5 Claude records unavailable due API 401 |
| python scripts/agent_efficiency_eval.py --run --agent claude --case canvas_prd --output /tmp/canvas-eval-claude-availability.jsonl --summary /tmp/canvas-eval-claude-availability-summary.json --require-canvas-agent-availability | expected fail: Claude unavailable due API 401, proves skipped providers are visible |
| python scripts/agent_efficiency_eval.py --from-results /tmp/canvas-eval-live-matrix-current.jsonl --summary /tmp/canvas-eval-live-matrix-current-replay-summary.json --require-canvas-90 | pass: replayed preserved live evidence at 5/5 available records |
| python scripts/agent_efficiency_eval.py --from-results /tmp/canvas-eval-live-matrix-current.jsonl --summary /tmp/canvas-eval-live-matrix-current-strict-summary.json --require-canvas-90 --require-canvas-agent-availability | expected fail: replayed strict provider coverage detects 5 unavailable Claude records |
| python scripts/agent_efficiency_eval.py --run --agent codex --case canvas_c3_adr --case canvas_atomic_design --case canvas_pm_requirement --case canvas_prd --case canvas_user_story --repeat 2 --output /tmp/canvas-eval-codex-repeat2-v3.jsonl --summary /tmp/canvas-eval-codex-repeat2-v3-summary.json --require-canvas-90 --min-canvas-records 10 --require-canvas-agent-availability | pass: 10/10 Codex canvas records, 100 percent, no unavailable records |
| python scripts/agent_efficiency_eval.py --from-results /tmp/canvas-eval-codex-repeat2-v3.jsonl --summary /tmp/canvas-eval-codex-repeat2-v3-coverage-summary.json --require-canvas-90 --min-canvas-records 10 --min-canvas-records-per-case 2 --require-canvas-agent-availability | pass: replayed Codex proof with at least two available records per canvas case |
| python scripts/agent_efficiency_eval.py --from-results /tmp/canvas-eval-codex-repeat2-v3.jsonl --summary /tmp/canvas-eval-codex-repeat2-v3-required-codex-summary.json --require-canvas-90 --min-canvas-records 10 --min-canvas-records-per-case 2 --require-canvas-agent-availability --require-canvas-agent codex | pass: Codex-required proof passes on preserved repeated evidence |
| python scripts/agent_efficiency_eval.py --from-results /tmp/canvas-eval-codex-repeat2-v3.jsonl --summary /tmp/canvas-eval-codex-repeat2-v3-required-both-summary.json --require-canvas-90 --min-canvas-records 10 --min-canvas-records-per-case 2 --require-canvas-agent-availability --require-canvas-agent codex --require-canvas-agent claude | expected fail: Claude-required proof fails because preserved repeated evidence contains no available Claude records |
| bash scripts/build.sh --version dev | pass |
| c3x lookup scripts/agent_efficiency_eval.py | pass |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --only adr-20260528-canvas-authoring-eval-loop --include-adr | pass |
