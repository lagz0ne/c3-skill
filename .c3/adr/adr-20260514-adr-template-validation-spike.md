---
id: adr-20260514-adr-template-validation-spike
c3-seal: 4a1d54052bfa18aed51ed07f9def5d622ba415112710568244afc32d743affd0
title: adr-template-validation-spike
type: adr
goal: Extract the current hard-coded ADR validation into a small template-oriented seam so the CLI can prove ADR templates can drive schema/add/check behavior without changing current ADR user behavior yet.
status: proposed
date: "2026-05-14"
---

## Goal

Extract the current hard-coded ADR validation into a small template-oriented seam so the CLI can prove ADR templates can drive schema/add/check behavior without changing current ADR user behavior yet.

## Context

Current ADR sections and rejection rules live in Go schema definitions, while add/check/linkage validators separately assume fixed ADR section names. This makes the ADR model strict but not user-shaped. The spike should keep existing strict behavior and test only the concept that an ADR template can be the source for schema output and validation inputs.

## Decision

Introduce a minimal internal ADR template concept that wraps the current ADR section definitions and reject rules, then route ADR schema and creation validation through that template accessor. Do not add custom project template loading, versioning, lifecycle policy, or new CLI surface in this spike.

## Affected Topology

| Entity | Type | Why affected | Governance review |
| --- | --- | --- | --- |
| c3-111 | component | add-cmd owns ADR creation validation and must keep rejecting incomplete ADRs before insert. | Review add-cmd ADR completeness and rollback contract. |
| c3-113 | component | check-cmd owns schema definitions and validation consumers that should now read ADR sections through the template seam. | Review schema registry and ADR validation contract. |
| c3-117 | component | docs-state-cmds owns schema command presentation that should be able to advertise template-derived help later. | Review schema output behavior stays compatible. |

## Compliance Refs

| Ref | Why required | Action |
| --- | --- | --- |
| N.A - no governing ref found in lookup for these command files. | N.A - current component docs cite parent policy only. | N.A - no ref action in this spike. |

## Compliance Rules

| Rule | Why required | Action |
| --- | --- | --- |
| N.A - no governing rule found in lookup for these command files. | N.A - current component docs cite parent policy only. | N.A - no rule action in this spike. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| cli/internal/schema | Add minimal ADR template accessors around existing ADR schema/reject rules. | go test ./cmd ./internal/schema. |
| cli/cmd | Route ADR creation/schema validation through template accessor while preserving current output. | focused add/schema tests. |
| tests | Add tests proving template accessor owns current ADR sections and add validation uses it. | go test ./cmd ./internal/schema. |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| schema registry | Add ADR template seam without removing current SectionDef data model. | go test ./internal/schema. |
| add validator | Use template-derived ADR sections for creation validation. | go test ./cmd -run TestRunAdd_Adr. |
| schema command | Keep existing schema output compatible while sourcing ADR reject rules from template seam. | go test ./cmd -run TestRunSchema. |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| c3x add adr | Still rejects missing/empty ADR sections and tables before insert. | go test ./cmd -run TestRunAdd_Adr. |
| c3x schema adr | Still prints current ADR sections and reject contract. | go test ./cmd -run TestRunSchema. |
| c3x check --include-adr | Existing ADR linkage validation remains unchanged in this spike. | C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr --only adr-20260514-adr-template-validation-spike. |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Build full custom template loader now. | Too broad for the current proof; project template parsing can come after the seam proves current behavior. |
| Add new CLI commands first. | Command surface before validator extraction would create API churn without proof. |
| Loosen current ADR validation immediately. | The current need is to preserve quality while making the strictness template-owned. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Behavior drift in current ADR validation | Keep current sections/reject rules as the default built-in template and run existing tests. | go test ./cmd ./internal/schema. |
| Over-abstraction | Add only accessors/data wrapper required by tests, no loader/DSL. | Review diff for unused generality. |
| Uncharted ADR linkage file ownership | Keep adr_linkage behavior unchanged and note codemap gap as no-delta. | c3x lookup cli/cmd/adr_linkage.go evidence. |

## Verification

| Check | Result |
| --- | --- |
| go test ./cmd ./internal/schema | Passed. |
| go test ./... | Passed. |
| go run . schema adr | Passed; output shows Template: implementation-change and current REJECT IF contract. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr --only adr-20260514-adr-template-validation-spike | Passed. |
