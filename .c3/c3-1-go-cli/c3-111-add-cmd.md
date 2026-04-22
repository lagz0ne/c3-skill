---
id: c3-111
c3-version: 4
c3-seal: 0dfea1e6263344d87672463270ce0b2eb5474001b2dde6b408419ceacb543c06
title: add-cmd
type: component
category: feature
parent: c3-1
goal: Create new containers, components, refs, rules, or ADRs with correct numbering and wired into the parent doc.
summary: Assigns IDs via the numbering library, creates stub docs from templates, and updates parent component tables
uses:
    - c3-101
    - c3-102
    - c3-103
    - c3-104
---

# add-cmd

## Goal

Create new containers, components, refs, rules, or ADRs with correct numbering and wired into the parent doc.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own add-cmd behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep add-cmd decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for add-cmd so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before add-cmd behavior is changed. | c3-1 |
| Inputs | Accept only the files, commands, data, or calls that belong to add-cmd ownership. | c3-1 |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | c3-1 |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | c3-1 |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks add-cmd to deliver its documented responsibility. | c3-1 |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | c3-1 |
| Alternate paths | When a request falls outside add-cmd ownership, hand it to the parent or sibling component. | c3-1 |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | c3-1 |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-1 | policy | Governs add-cmd behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| add body input | IN | c3x add requires stdin body content and validates it against the target entity schema before inserting anything into the local cache. | c3-111 command validation boundary. | cli/cmd/add.go; cli/cmd/add_test.go TestRunAdd_NilReaderFails and TestRunAdd_MissingSectionsFails. |
| ADR creation | OUT | c3x add adr creates the entire ADR entity from the provided body in one operation; it must not create a partial ADR entity or file when validation, content write, or canonical export fails. | c3-111 uses dispatcher rollback in c3-108 for canonical export failures. | cli/cmd/add.go compensation path; cli/main_test.go TestRun_AddADRRollsBackWhenCanonicalExportFails. |
| agent result | OUT | Agent-mode add output uses TOON and includes c3x-owned next-step hints for ADR schema, read, write, check, and verify. | c3-111 result with c3-108 presentation helpers. | cli/cmd/add_test.go TestRunAdd_AdrAgentHintsUseCLISchema; cli/main_test.go TestRun_AddAgentModeReturnsTOON. |
| ADR completeness gate | IN | c3x add adr rejects thin ADR bodies and requires all ADR schema sections with table rows before any entity insert. | c3-111 add validation boundary. | cli/cmd/add.go validateADRCreationBody; TestRunAdd_AdrRequiresCompleteBody. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Partial ADR creation | add adr inserts the entity, then content write or canonical export fails. | Check local cache for the ADR slug and canonical adr/*.md files after forced export failure. | go test -count=1 . -run TestRun_AddADRRollsBackWhenCanonicalExportFails. |
| Thin ADR creation | ADR body omits required schema sections or agents ignore CLI hints. | Run c3x schema adr and c3x check --include-adr after creation. | go test ./cmd -run TestRunAdd_AdrAgentHintsUseCLISchema; c3x check --include-adr. |
| Agent output regression | add returns JSON or loses help hints in C3X_MODE=agent. | Smoke add in a temp project and inspect output prefix and help[n]. | go test ./...; C3X_MODE=agent c3x add adr smoke-output. |
| Incremental ADR creation | Agent creates a Goal-only ADR and depends on a second write command for required decision-ledger sections. | Run add adr with only Goal and expect validation errors for missing Context, Decision, Underlay C3 Changes, and other ledger sections. | go test ./cmd -run TestRunAdd_AdrRequiresCompleteBody. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/cmd/add.go | Contract add body input row and Contract ADR creation row. | Entity builders may grow; add must validate before insert and compensate if content write fails. | go test ./cmd -run TestRunAdd. |
| cli/main.go | Contract ADR creation row and c3-108 mutating command rollback contract. | Rollback mechanism may be centralized; add adr must remain all-or-nothing across cache and canonical export. | go test -count=1 . -run TestRun_AddADRRollsBackWhenCanonicalExportFails. |
| cli/cmd/add_test.go and cli/main_test.go | Contract ADR creation row, Contract agent result row, and Change Safety rows. | Tests may cover more entity types; ADR rollback and TOON hints must remain covered. | go test ./... |
| cli/cmd/add.go validateADRCreationBody | Contract ADR completeness gate row and Change Safety incremental ADR creation risk. | Validation wording may change; complete ADR creation must reject missing sections, empty sections, empty tables, and missing table columns before insert. | go test ./cmd -run TestRunAdd_AdrRequiresCompleteBody. |
