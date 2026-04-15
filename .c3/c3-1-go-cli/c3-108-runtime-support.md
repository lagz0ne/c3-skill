---
id: c3-108
c3-seal: b44edc308456c7a06b0aae108f10c736ae67ad4a369381f350a314ffd0ed601b
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
| Alternate paths | Outside agent mode, explicit --json continues to produce JSON for scripts that ask for it. | c3-1 |
| Failure behavior | If serialization cannot represent a result, return the serializer error and let command handling fail visibly. | c3-1 |
## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-1 | policy | Container-level CLI runtime and distribution boundaries. | Parent container scope beats local helper convenience. | Runtime support stays inside Go CLI ownership. |
| adr-20260415-agent-mode-toon-only | adr | Decision that agent mode must not emit JSON through legacy machine-output paths. | This ADR overrides older agent JSON wording in tests and docs. | C3X_MODE=agent now means TOON even when internal command routing sets JSON=true. |
## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| argv/env parsing | IN | --json marks explicit JSON for non-agent callers; C3X_MODE=agent marks machine output without making JSON the wire format. | c3-1 boundary | go test ./...; cli/cmd/options_test.go. |
| output serializer | OUT | Under C3X_MODE=agent, generic structured output must serialize as TOON and must not start with JSON object/array delimiters. | c3-1 boundary | go test ./...; cli/cmd/output_test.go. |
| command compatibility | OUT | Existing commands that call writeJSON for machine output inherit TOON in agent mode without per-command rewrites. | c3-1 boundary | go test ./...; agent smoke commands list, lookup, read, check. |
| explicit JSON | OUT | Outside agent mode, explicit --json remains JSON for scripts and tests that parse JSON. | c3-1 boundary | go test ./...; JSON tests in cli/cmd/*_test.go. |
## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| JSON leak in agent mode | A command bypasses shared serializer or a test restores agent JSON assumptions. | Search for agent JSON wording and smoke agent commands for leading { or [. | Run go test ./... and C3X_MODE=agent bash skills/c3/bin/c3x.sh list --json. |
| Human JSON regression | Serializer changes break explicit --json outside agent mode. | JSON parse tests fail for non-agent command paths. | Run go test ./.... |
| TOON shape regression | Generic slice or object output becomes unreadable for agents. | cli/cmd/output_test.go checks object and slice TOON output. | Run go test ./cmd -run TestWriteJSON_Agent. |
## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/cmd/helpers.go | Contract output serializer row and Change Safety JSON leak row. | Implementation can evolve; agent mode must choose TOON before JSON. | go test ./.... |
| cli/cmd/output.go and cli/cmd/options.go | Contract argv/env parsing row. | Internal flag names may remain legacy; behavior must match the contract. | go test ./cmd -run TestResolveFormat. |
| cli/internal/toon/toon.go | Contract command compatibility row. | TOON formatting can improve; it must support structs, maps, slices, and scalars used by command output. | go test ./cmd -run TestWriteJSON_Agent. |
| skills/c3/SKILL.md and references | Governance and Contract rows. | Copy can be concise; it must not instruct agents to parse JSON from agent mode. | rg "agent JSON" skills/c3 cli. |
