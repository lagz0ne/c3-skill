---
id: c3-108
c3-seal: ab80f21c52a4138da854eff84e6adae35554c5f7b269eda903c461d004762ae2
title: runtime-support
type: component
category: foundation
parent: c3-1
goal: Provide CLI bootstrap, option parsing, output shaping, config resolution, and human/agent presentation helpers.
---

# runtime-support

## Goal

Provide CLI bootstrap, option parsing, output shaping, config resolution, and human/agent presentation helpers.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own runtime-support behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep runtime-support decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Own CLI runtime behavior shared across commands: argument parsing, environment mode detection, serialization selection, help-hint rendering, and human versus agent presentation. It does not own command-specific business logic or C3 document mutation semantics.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before runtime-support behavior is changed. | c3-1 |
| Inputs | Accept argv flags such as --json and --continue, environment values such as C3X_MODE=agent, and command result structs or slices that need serialization. | c3-1 |
| State / data | Preserve mode selection as process-local runtime state; do not persist output-mode decisions into .c3/ docs. | c3-1 |
| Shared dependencies | Use cli/internal/toon for agent serialization and Go JSON encoding only for non-agent explicit JSON output. | c3-1 |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Human shell, npm shim, skill wrapper, or automated workflow invokes c3x and expects output shaped for that caller. | c3-1 |
| Primary path | Parse flags, detect C3X_MODE=agent, verify canonical state, auto-repair recoverable cache/seal drift, route agent machine output through TOON, and keep human output readable by default. | adr-20260421-self-healing-preflight |
| Alternate paths | Explicit lifecycle commands such as migrate and sync export keep command-owned control of repair/export sequencing; normal read and mutation commands no longer require a separate repair handoff for recoverable drift. | adr-20260421-self-healing-preflight |
| Failure behavior | If automatic repair cannot prove canonical truth, stop before dispatch with the auto-repair failure and original verification error so the user sees the real blocker instead of a generic repair instruction. | adr-20260421-self-healing-preflight |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-1 | policy | Container-level CLI runtime and distribution boundaries. | Parent container scope beats local helper convenience. | Runtime support stays inside Go CLI ownership. |
| adr-20260415-agent-mode-toon-only | adr | Decision that agent mode must not emit JSON through legacy machine-output paths. | This ADR overrides older agent JSON wording in tests and docs. | C3X_MODE=agent means TOON even when internal command routing sets JSON=true. |
| adr-20260415-force-agent-toon-output | adr | Systematic sweep of command wrappers and direct encoders that bypassed writeJSON. | Shared output helpers beat command-local JSON encoding. | add, migrate dry-run, and legacy check now use TOON-aware structured output in agent mode. |
| adr-20260415-mutation-preverify-repair-bypass | adr | Dispatcher preverify gate for broken canonical docs. | Mutating commands need command-local validation and export path access; read-only commands stay gated. | Prevents repair catch-22 for add, write, set, wire, delete, codemap, and migrate. |
| adr-20260421-self-healing-preflight | adr | Dispatcher preflight self-heals recoverable canonical/cache drift before normal command dispatch. | New self-healing policy supersedes older read-only broken-canonical blocking wording. | repair remains explicit, but normal operations attempt it automatically before failing. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| agent output serializer | OUT | Under C3X_MODE=agent, structured command output must be TOON and contextual hints render as help[n] rows, not JSON-only prose. | c3-108 runtime presentation boundary. | cli/cmd/output.go; cli/cmd/add_test.go TestRunAdd_AdrAgentHintsUseCLISchema. |
| ADR creation hints | OUT | Adding or reading an ADR must route agents to c3x schema adr, c3x read <adr> --full, c3x write <adr> < adr.md, focused c3x verify --only <adr> --include-adr, and full c3x check --include-adr plus c3x verify --include-adr before final handoff. | c3-108 owns hint rendering; ADR schema content belongs to c3-113 and c3-117. | cli/cmd/cascade_hints.go; TestRunAdd_AdrAgentHintsUseCLISchema. |
| add help | OUT | c3x add help must teach the ADR workflow through CLI commands and must not advertise unsupported ADR --goal shortcuts. | c3-108 help surface. | cli/cmd/help.go; cli/cmd/help_test.go TestShowHelp_AddADRWorkflowPointsAtSchema. |
| cascade review hints | OUT | Check, diff, graph, and other agent hints must show scoped verify --only <id> as the branch-safe proof path while unrelated docs are still in progress. | c3-108 hint surface; c3-119 owns verify behavior. | cli/cmd/cascade_hints.go; TestRunCheck_AgentTOONIncludesCascadeReviewHint. |
| explicit JSON | OUT | Outside agent mode, explicit --json remains available where commands support it; agent-facing defaults remain TOON. | c3-108 compatibility boundary. | go test ./cmd. |
| mutating command rollback | OUT | For mutating commands, dispatcher snapshots the .c3 cache and canonical tree before command execution; if command handling or canonical export returns an error, it restores the pre-command state before returning the failure. | c3-108 dispatcher boundary; command-specific validation still belongs to each command component. | cli/main.go mutationSnapshot; cli/main_test.go TestRun_AddADRRollsBackWhenCanonicalExportFails. |
| self-healing preflight | IN/OUT | Normal store-backed commands verify canonical .c3 state and automatically run repair for recoverable cache or seal drift before dispatch; only unrecoverable repair failure stops the command. | c3-108 owns dispatch timing; c3-119 owns repair mechanics. | cli/main.go preflight branch; cli/main_test.go TestRun_ListSelfHealsBrokenSeal and TestRun_CommandsSelfHealBrokenCanonicalPreverify. |
| verify include-adr dispatch | IN/OUT | The existing --include-adr parser flag is passed into verify, repair, and preflight verify without adding a second mode or command-specific parser path. | c3-108 runtime dispatch boundary; c3-119 owns lifecycle verification behavior. | cli/main.go VerifyOptions and RepairOptions construction; go test . -run TestRun_Verify. |
| verify only dispatch | IN/OUT | The repeatable --only parser flag is passed into verify, repair, and preflight verify so command dispatch can select a focused canonical doc set without a separate command. | c3-108 runtime parsing and dispatch boundary; c3-119 owns lifecycle sync behavior and c3-113 owns scoped schema validation. | cli/cmd/options.go; cli/main.go; cli/cmd/options_test.go TestParseArgs. |
| short-lived write coordinator | IN/OUT | Mutating commands may forward argv, stdin, cwd, and C3X_MODE to a short-lived per-.c3 Unix socket leader; the leader serializes requests through the internal command path with coordinator forwarding disabled. | c3-108 owns runtime dispatch and request forwarding; command-specific mutation semantics remain with sibling command components. | cli/main.go runThroughCoordinator and runQueuedRequest; cli/internal/coord/**; cli/main_test.go TestRun_ConcurrentMutationsUseShortLivedCoordinator and TestRun_CoordinatorForwardsPipedInput. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Repair catch-22 | Dispatcher runs verify before write, set, add, or other repair mutations can execute. | Break a canonical seal and run both a read-only command and add adr against the same tree. | Run go test -count=1 . -run TestRun_MutatingCommandBypassesBrokenCanonicalPreverify. |
| JSON leak in agent mode | A command bypasses shared serializer or a test restores agent JSON assumptions. | Search for json.NewEncoder outside tests and smoke agent commands for leading { or [. | Run go test ./... and C3X_MODE=agent bash skills/c3/bin/c3x.sh add ref agent-output in a temp project. |
| Wrapper bypass | Main command dispatch builds JSON after command execution instead of asking the command package for structured output. | Inspect cli/main.go and command-specific result structs for direct encoders. | Run TestRun_AddAgentModeReturnsTOON and rg json.NewEncoder cli --glob '!**/*_test.go'. |
| Human JSON regression | Serializer changes break explicit --json outside agent mode. | JSON parse tests fail for non-agent command paths. | Run go test ./.... |
| TOON shape regression | Generic slice or object output becomes unreadable for agents. | cli/cmd/output_test.go checks object and slice TOON output. | Run go test ./cmd -run TestWriteJSON_Agent. |
| Partial mutation | A mutating command updates local cache but canonical export fails afterward. | Simulate export failure during add adr and inspect both canonical files and c3.db for rollback. | go test -count=1 . -run TestRun_AddADRRollsBackWhenCanonicalExportFails. |
| Auto-repair hides real corruption | A normal command sees broken canonical seals or stale cache and repair cannot prove the tree safe. | Command returns auto-repair failed with the original verification error instead of continuing. | go test -count=1 . -run TestRun_ListSelfHealsBrokenSeal; go test ./.... |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/cmd/cascade_hints.go | Contract ADR creation hints row and Purpose section. | Hint wording may be compact; commands must remain c3x-native and actionable. | go test ./cmd -run TestRunAdd_AdrAgentHintsUseCLISchema. |
| cli/cmd/help.go | Contract add help row and Change Safety human JSON regression risk. | Examples may change; ADR path must point to c3x schema adr and stdin body creation. | go test ./cmd -run TestShowHelp_AddADRWorkflowPointsAtSchema. |
| cli/cmd/add_test.go and cli/cmd/help_test.go | Contract agent output serializer, ADR creation hints, and add help rows. | Tests may assert additional CLI hints as workflow grows. | go test ./cmd. |
| skills/c3/SKILL.md | Contract ADR creation hints row and Governance c3-1 policy. | Skill is reference-only; it must route enforcement to c3x output. | rg "Enforcement source" skills/c3/SKILL.md. |
| cli/main.go | Contract mutating command rollback row and Change Safety partial mutation risk. | Snapshot implementation may change; failed mutating commands must leave cache and canonical docs at pre-command state. | go test -count=1 . -run TestRun_AddADRRollsBackWhenCanonicalExportFails. |
