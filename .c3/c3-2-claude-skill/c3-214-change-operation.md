---
id: c3-214
c3-seal: 1ffd795a6e8eaea2c594525f84d2bbbeda1f69fbd04dc61bca3ba2dbc25c6467
title: change-operation
type: component
category: feature
parent: c3-2
goal: Demonstrate how the skill makes architecture-affecting changes ADR-first, with lookup context, parent-delta evidence, implementation, and verification.
uses:
    - c3-201
    - c3-210
---

# change-operation

## Goal

Demonstrate how the skill makes architecture-affecting changes ADR-first, with lookup context, parent-delta evidence, implementation, and verification.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own change-operation behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep change-operation decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Ensure architecture-affecting work starts with an ADR that carries enough detail to survive beyond the chat: context, decision, enforcement surfaces, alternatives, risks, verification, and parent-delta evidence. The ADR is the work order and audit trail that prevents strict workflow details from being lost after implementation.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before change-operation behavior is changed. | c3-2 |
| Inputs | Accept only the files, commands, data, or calls that belong to change-operation ownership. | c3-2 |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | c3-2 |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | c3-2 |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks change-operation to deliver its documented responsibility. | c3-2 |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | c3-2 |
| Alternate paths | When a request falls outside change-operation ownership, hand it to the parent or sibling component. | c3-2 |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | c3-2 |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-2 | policy | Governs change-operation behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| ADR creation | OUT | Change workflow starts by invoking c3x add adr <slug>; the CLI creates the work order and emits next-step hints. | c3-214 skill workflow boundary; enforcement is c3x output. | skills/c3/references/change.md Phase 1; cli/cmd/cascade_hints.go ADR hints. |
| ADR detail fill | OUT | ADR structure and section meaning come from c3x schema adr and c3x read <adr> --full, not duplicated skill checklists. | c3-214 with c3-117 schema command and c3-113 schema registry. | skills/c3/references/change.md Phase 2; cli/cmd/schema.go; cli/internal/schema/schema.go. |
| Context gate | IN | Touched files or globs are mapped with c3x lookup, then parent/component and applicable refs/rules are loaded before editing. | c3-214 boundary. | skills/c3/references/change.md Phase 3; c3x lookup for touched files. |
| Parent delta | OUT | Component changes record whether parent container/context changed or why no parent delta was needed, following CLI lookup/read/graph hints. | c3-214 boundary. | c3x help[] from lookup/read/set/write; c3x check --include-adr. |
| Verification close | OUT | Change is not done until relevant tests, c3x check --include-adr, and c3x verify pass or the blocker is reported. | c3-214 boundary. | skills/c3/references/change.md Phase 4. |
| Complete ADR creation | OUT | Change workflow must create a complete ADR work order through c3x add adr; thin ADR creation followed by incremental fill is not allowed. | c3-214 workflow boundary; c3-111 enforces add-time completeness. | skills/c3/references/change.md Phase 1; cli/cmd/add.go validateADRCreationBody; TestRunAdd_AdrRequiresCompleteBody. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Thin ADR | ADR only restates the goal and misses decision detail. | Read ADR before execution; missing Context, Decision, Enforcement Surfaces, Risks, or Verification for strict changes is a blocker. | c3x read <adr> --full; c3x check --include-adr. |
| Context bypass | Agent edits a file without lookup and parent context. | Compare touched files against lookup evidence in ADR or final report. | c3x lookup <file> for every touched file. |
| Parent drift | Component doc or ownership changes without parent decision. | Contract Cascade Gate has blank parent row or no-delta reason. | Record Parent Delta in ADR and run c3x check --include-adr. |
| Verification gap | Task list is done but tests or C3 validation are missing. | Final report lacks concrete command evidence. | Run relevant tests plus c3x verify. |
| Incremental ADR workflow | Workflow creates an ADR before the decision ledger is complete and expects follow-up writes to supply core context. | Compare change reference Phase 1 with c3x add adr validation behavior. | go test ./cmd -run TestRunAdd_AdrRequiresCompleteBody; c3x check --include-adr. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| skills/c3/references/change.md | Contract ADR creation row, Contract ADR detail fill row, and Change Safety thin ADR risk. | Reference prose can stay compact; it must route detail and failure recovery to c3x schema/help/hints. | rg "The CLI is the source of truth" skills/c3/references/change.md. |
| skills/c3/SKILL.md | Contract ADR detail fill row and Governance c3-2 policy. | Top-level skill can mention workflow order; enforcement detail must stay in c3x. | rg "Enforcement source" skills/c3/SKILL.md. |
| ADRs created by change workflow | Contract ADR detail fill row and Contract Verification close row. | Small changes may stay concise; underlay C3 changes need schema-guided sections and evidence. | c3x schema adr; c3x check --include-adr. |
| Agent final reports | Contract Verification close row and Change Safety verification gap risk. | Format may vary; must include concrete command evidence or explicit blocker. | c3x verify; relevant test command. |
| skills/c3/references/change.md | Contract Complete ADR creation row and Change Safety incremental ADR workflow risk. | Reference prose may stay concise; Phase 1 must require complete ADR creation through c3x add adr. | rg "complete ADR" skills/c3/references/change.md. |
