---
id: c3-117
c3-seal: 54a42fc5ca76da168b5e53a35be48bf5d4f1048694f4a8ae3937bf67ed96d021
title: docs-state-cmds
type: component
category: feature
parent: c3-1
goal: Read, write, set, validate schema, and report status for canonical C3 documents.
---

# docs-state-cmds

## Goal

Read, write, set, validate schema, and report status for canonical C3 documents.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own docs-state-cmds behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep docs-state-cmds decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for docs-state-cmds so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before docs-state-cmds behavior is changed. | c3-1 |
| Inputs | Accept only the files, commands, data, or calls that belong to docs-state-cmds ownership. | c3-1 |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | c3-1 |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | c3-1 |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks docs-state-cmds to deliver its documented responsibility. | c3-1 |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | c3-1 |
| Alternate paths | When a request falls outside docs-state-cmds ownership, hand it to the parent or sibling component. | c3-1 |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | c3-1 |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-1 | policy | Governs docs-state-cmds behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| schema text output | OUT | c3x schema <type> must print section names, content type, required marker, purpose text, and table columns so the CLI tells agents what durable content belongs in each section. | c3-117 owns schema command presentation; schema definitions remain in c3-113. | cli/cmd/schema.go; cli/cmd/schema_test.go TestRunSchema_ADRIncludesDecisionLedger. |
| schema JSON output | OUT | Explicit --json keeps machine-readable schema sections, purposes, and columns for non-agent scripts; agent mode output policy remains owned by runtime support. | c3-117 boundary with c3-108 serializer behavior. | cli/cmd/schema_test.go TestRunSchema_JSON_ADRUnderlayColumns. |
| docs-state repair loop | OUT | Read, write, set, check, and verify remain the CLI path for canonical docs; skill references may point here but must not become a second enforcement checklist. | c3-1 canonical doc ownership. | C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr; c3x verify. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |
| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/cmd/schema.go | Contract schema text output row and Purpose section. | Formatting can evolve; purpose text must remain present in human schema output. | go test ./cmd -run TestRunSchema_ADRIncludesDecisionLedger. |
| cli/cmd/schema_test.go | Contract schema text output row, Contract schema JSON output row, and Change Safety contract drift risk. | Assertions may grow with schema sections; tests must preserve ADR underlay purpose and columns. | go test ./cmd -run TestRunSchema_ADR. |
| skills/c3/SKILL.md and references | Contract docs-state repair loop row and Governance c3-1 policy. | Reference prose may be shorter than CLI help; it must route detail and enforcement back to c3x. | rg "c3x schema adr" skills/c3 cli. |
