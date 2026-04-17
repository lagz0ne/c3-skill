---
id: c3-108
c3-seal: 0f42684567c2c405e7250e3d1d7e601f9bb973117cf79f956b740c55d61865b6
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
| Primary path | Parse flags, detect C3X_MODE=agent, route agent machine output through TOON, and keep human output readable by default. | adr-20260415-agent-mode-toon-only |
| Alternate paths | When canonical docs fail preverification, read-only commands stay blocked, while mutating commands bypass the dispatcher gate and rely on their own payload validation plus canonical export. | adr-20260415-mutation-preverify-repair-bypass |
| Failure behavior | If serialization or command-local validation cannot represent a safe result, return the error and include the next repair step where the command can prove it. | c3-1 |
## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-1 | policy | Container-level CLI runtime and distribution boundaries. | Parent container scope beats local helper convenience. | Runtime support stays inside Go CLI ownership. |
| adr-20260415-agent-mode-toon-only | adr | Decision that agent mode must not emit JSON through legacy machine-output paths. | This ADR overrides older agent JSON wording in tests and docs. | C3X_MODE=agent means TOON even when internal command routing sets JSON=true. |
| adr-20260415-force-agent-toon-output | adr | Systematic sweep of command wrappers and direct encoders that bypassed writeJSON. | Shared output helpers beat command-local JSON encoding. | add, migrate dry-run, and legacy check now use TOON-aware structured output in agent mode. |
| adr-20260415-mutation-preverify-repair-bypass | adr | Dispatcher preverify gate for broken canonical docs. | Mutating commands need command-local validation and export path access; read-only commands stay gated. | Prevents repair catch-22 for add, write, set, wire, delete, codemap, and migrate. |
## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| agent output serializer | OUT | Under C3X_MODE=agent, structured command output must be TOON and contextual hints render as help[n] rows, not JSON-only prose. | c3-108 runtime presentation boundary. | cli/cmd/output.go; cli/cmd/add_test.go TestRunAdd_AdrAgentHintsUseCLISchema. |
| ADR creation hints | OUT | Adding or reading an ADR must route agents to c3x schema adr, c3x read <adr> --full, c3x write <adr> < adr.md, and c3x check --include-adr && c3x verify. | c3-108 owns hint rendering; ADR schema content belongs to c3-113 and c3-117. | cli/cmd/cascade_hints.go; TestRunAdd_AdrAgentHintsUseCLISchema. |
| add help | OUT | c3x add help must teach the ADR workflow through CLI commands and must not advertise unsupported ADR --goal shortcuts. | c3-108 help surface. | cli/cmd/help.go; cli/cmd/help_test.go TestShowHelp_AddADRWorkflowPointsAtSchema. |
| explicit JSON | OUT | Outside agent mode, explicit --json remains available where commands support it; agent-facing defaults remain TOON. | c3-108 compatibility boundary. | go test ./cmd. |
| mutating command rollback | OUT | For mutating commands, dispatcher snapshots the .c3 cache and canonical tree before command execution; if command handling or canonical export returns an error, it restores the pre-command state before returning the failure. | c3-108 dispatcher boundary; command-specific validation still belongs to each command component. | cli/main.go mutationSnapshot; cli/main_test.go TestRun_AddADRRollsBackWhenCanonicalExportFails. |
## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Repair catch-22 | Dispatcher runs verify before write, set, add, or other repair mutations can execute. | Break a canonical seal and run both a read-only command and add adr against the same tree. | Run go test -count=1 . -run TestRun_MutatingCommandBypassesBrokenCanonicalPreverify. |
| JSON leak in agent mode | A command bypasses shared serializer or a test restores agent JSON assumptions. | Search for json.NewEncoder outside tests and smoke agent commands for leading { or [. | Run go test ./... and C3X_MODE=agent bash skills/c3/bin/c3x.sh add ref agent-output in a temp project. |
| Wrapper bypass | Main command dispatch builds JSON after command execution instead of asking the command package for structured output. | Inspect cli/main.go and command-specific result structs for direct encoders. | Run TestRun_AddAgentModeReturnsTOON and rg json.NewEncoder cli --glob '!**/*_test.go'. |
| Human JSON regression | Serializer changes break explicit --json outside agent mode. | JSON parse tests fail for non-agent command paths. | Run go test ./.... |
| TOON shape regression | Generic slice or object output becomes unreadable for agents. | cli/cmd/output_test.go checks object and slice TOON output. | Run go test ./cmd -run TestWriteJSON_Agent. |
| Partial mutation | A mutating command updates local cache but canonical export fails afterward. | Simulate export failure during add adr and inspect both canonical files and c3.db for rollback. | go test -count=1 . -run TestRun_AddADRRollsBackWhenCanonicalExportFails. |
## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/cmd/cascade_hints.go | Contract ADR creation hints row and Purpose section. | Hint wording may be compact; commands must remain c3x-native and actionable. | go test ./cmd -run TestRunAdd_AdrAgentHintsUseCLISchema. |
| cli/cmd/help.go | Contract add help row and Change Safety human JSON regression risk. | Examples may change; ADR path must point to c3x schema adr and stdin body creation. | go test ./cmd -run TestShowHelp_AddADRWorkflowPointsAtSchema. |
| cli/cmd/add_test.go and cli/cmd/help_test.go | Contract agent output serializer, ADR creation hints, and add help rows. | Tests may assert additional CLI hints as workflow grows. | go test ./cmd. |
| skills/c3/SKILL.md | Contract ADR creation hints row and Governance c3-1 policy. | Skill is reference-only; it must route enforcement to c3x output. | rg "Enforcement source" skills/c3/SKILL.md. |
| cli/main.go | Contract mutating command rollback row and Change Safety partial mutation risk. | Snapshot implementation may change; failed mutating commands must leave cache and canonical docs at pre-command state. | go test -count=1 . -run TestRun_AddADRRollsBackWhenCanonicalExportFails. |
