---
id: c3-108
c3-seal: 80d6ddc5963a1b35b935224f21f740c950a17e2e9dfe2f6755c011a6ee8a9f8b
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
| Inputs | Accept argv flags such as --json, environment values such as C3X_MODE=agent, and command result structs or slices that need serialization. | c3-1 |
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
| argv/env parsing | IN | --json marks explicit JSON for non-agent callers; C3X_MODE=agent marks machine output without making JSON the wire format. | c3-1 boundary | go test ./...; cli/cmd/options_test.go. |
| preverify gate | IN | When canonical .c3/ docs exist, read-only commands must preverify before dispatch, but mutating commands must reach command-local validation and canonical export. | c3-108 boundary | cli/main.go; TestRun_MutatingCommandBypassesBrokenCanonicalPreverify. |
| output serializer | OUT | Under C3X_MODE=agent, generic structured output must serialize as TOON and must not start with JSON object/array delimiters. | c3-1 boundary | go test ./...; cli/cmd/output_test.go. |
| command wrappers | OUT | Commands that need wrapper-level structured results, including add, must call shared output format resolution instead of local json.NewEncoder paths. | c3-108 boundary | cli/main.go; cli/cmd/add.go; TestRun_AddAgentModeReturnsTOON. |
| command compatibility | OUT | Existing commands that call writeJSON for machine output inherit TOON in agent mode without per-command rewrites. | c3-1 boundary | go test ./...; agent smoke commands list, lookup, read, check, add. |
| explicit JSON | OUT | Outside agent mode, explicit --json remains JSON for scripts and tests that parse JSON. | c3-1 boundary | go test ./...; JSON tests in cli/cmd/*_test.go. |
## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Repair catch-22 | Dispatcher runs verify before write, set, add, or other repair mutations can execute. | Break a canonical seal and run both a read-only command and add adr against the same tree. | Run go test -count=1 . -run TestRun_MutatingCommandBypassesBrokenCanonicalPreverify. |
| JSON leak in agent mode | A command bypasses shared serializer or a test restores agent JSON assumptions. | Search for json.NewEncoder outside tests and smoke agent commands for leading { or [. | Run go test ./... and C3X_MODE=agent bash skills/c3/bin/c3x.sh add ref agent-output in a temp project. |
| Wrapper bypass | Main command dispatch builds JSON after command execution instead of asking the command package for structured output. | Inspect cli/main.go and command-specific result structs for direct encoders. | Run TestRun_AddAgentModeReturnsTOON and rg json.NewEncoder cli --glob '!**/*_test.go'. |
| Human JSON regression | Serializer changes break explicit --json outside agent mode. | JSON parse tests fail for non-agent command paths. | Run go test ./.... |
| TOON shape regression | Generic slice or object output becomes unreadable for agents. | cli/cmd/output_test.go checks object and slice TOON output. | Run go test ./cmd -run TestWriteJSON_Agent. |
## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/main.go | Contract preverify gate and command wrappers rows. | Dispatch can stay thin; it must not block mutating repair commands behind canonical preverify. | cli/main_test.go TestRun_MutatingCommandBypassesBrokenCanonicalPreverify; cli/main_test.go TestRun_AddAgentModeReturnsTOON. |
| cli/cmd/add.go | Contract command wrappers and output serializer rows. | Result fields may grow; agent mode must go through WriteObjectOutput or writeJSON-compatible helpers. | cli/main_test.go TestRun_AddAgentModeReturnsTOON. |
| cli/cmd/migrate_dryrun.go and cli/cmd/check_legacy.go | Contract command compatibility row. | Legacy commands may keep their result structs; JSON branches must call writeJSON. | cli/cmd/migrate_dryrun_test.go TestRunMigrateDryRun_AgentModeReturnsTOON; rg json.NewEncoder cli --glob '!**/*_test.go'. |
| cli/cmd/helpers.go | Contract output serializer row and Change Safety JSON leak row. | Implementation can evolve; agent mode must choose TOON before JSON. | go test ./.... |
| cli/cmd/output.go and cli/cmd/options.go | Contract argv/env parsing row. | Internal flag names may remain legacy; behavior must match the contract. | go test ./cmd -run TestResolveFormat. |
| cli/internal/toon/toon.go | Contract command compatibility row. | TOON formatting can improve; it must support structs, maps, slices, and scalars used by command output. | go test ./cmd -run TestWriteJSON_Agent. |
| skills/c3/SKILL.md and references | Governance and Contract rows. | Copy can be concise; it must not instruct agents to parse JSON from agent mode. | rg "agent JSON" skills/c3 cli. |
