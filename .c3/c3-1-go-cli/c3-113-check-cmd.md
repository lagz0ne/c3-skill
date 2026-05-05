---
id: c3-113
c3-version: 4
c3-seal: 539a2ac1b7138a3896cb718923306534f35bf5b8b69cf5e3db3c37eb9ccb9b3e
title: check-cmd
type: component
category: feature
parent: c3-1
goal: Validate structural integrity of `.c3/` docs, ref and rule compliance — required fields, numbering, wiring, scope cross-checks, origin validation.
summary: Reports PASS/WARN/FAIL for each entity; includes schema definitions, structural index building, and ref scope cross-checking
uses:
    - c3-101
    - c3-102
    - c3-104
---

# check-cmd

## Goal

Validate structural integrity of `.c3/` docs, ref and rule compliance — required fields, numbering, wiring, scope cross-checks, origin validation.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own check-cmd behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep check-cmd decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for check-cmd so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before check-cmd behavior is changed. | c3-1 |
| Inputs | Accept only the files, commands, data, or calls that belong to check-cmd ownership. | c3-1 |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | c3-1 |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | c3-1 |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks check-cmd to deliver its documented responsibility. | c3-1 |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | c3-1 |
| Alternate paths | When a request falls outside check-cmd ownership, hand it to the parent or sibling component. | c3-1 |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | c3-1 |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-1 | policy | Governs check-cmd behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| schema registry | OUT | Entity schemas define ordered sections, required markers, purpose hints, and typed table columns used by add, set, write, check, and schema commands. | c3-113 owns validation/schema definitions; c3-117 owns schema command presentation. | cli/internal/schema/schema.go; cli/cmd/schema_test.go. |
| ADR schema | OUT | ADR schema must preserve decision-ledger sections for Context, Decision, Work Breakdown, Underlay C3 Changes, Enforcement Surfaces, Alternatives Considered, Risks, and Verification. | c3-113 validation boundary. | cli/internal/schema/schema.go adr registry; TestRunSchema_ADRIncludesDecisionLedger; TestRunSchema_JSON_ADRUnderlayColumns. |
| validation consumers | OUT | Validation paths must reject missing required sections and expose schema issues through c3x check/add/write/set rather than skill-local enforcement. | c3-113 with command-specific mutation handlers. | cli/cmd/check_enhanced.go; cli/cmd/add.go; cli/cmd/write.go; cli/cmd/set.go; go test ./cmd. |
| scoped verify check | IN | When verify supplies --only selectors, check validation runs only on selected entities and filters global relationship/layer warnings to the selected docs so unrelated docs do not block focused verification. | c3-113 check validation boundary; c3-119 owns seal and sync path filtering. | cli/cmd/check_enhanced.go; cli/main_test.go TestRun_VerifyOnlySkipsUnselectedComponentDrift. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |
| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/internal/schema/schema.go | Contract schema registry row and Contract ADR schema row. | Section purposes and columns may grow; ADR underlay and enforcement sections must remain CLI-owned. | go test ./cmd -run TestRunSchema_ADR. |
| cli/cmd/check_enhanced.go and mutation validators | Contract validation consumers row and Change Safety contract drift risk. | Command-specific validation can be stricter; required-section enforcement stays schema-driven. | go test ./cmd. |
| skills/c3/references/change.md | Contract ADR schema row and Governance c3-1 policy. | Reference copy may point to c3x instead of repeating sections. | rg "The CLI is the source of truth" skills/c3/references/change.md. |
