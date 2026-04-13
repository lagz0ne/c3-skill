---
id: adr-00000000-c3-adoption
c3-version: 4
c3-seal: e20ee5f281e643a5890dba6ff2d4c759b9ce334334d61ba596bac988265ab729
title: C3 Architecture Documentation Adoption
type: adr
goal: Adopt C3 methodology for c3-design.
status: in-progress
date: "2026-02-27"
affects:
    - c3-0
---

# C3 Architecture Documentation Adoption
## Goal

Adopt C3 methodology for c3-design.

## Workflow

```mermaid
flowchart TD
    GOAL([Goal]) --> S0

    subgraph S0["Stage 0: Inventory"]
        S0_DISCOVER[Discover codebase] --> S0_ASK{Gaps?}
        S0_ASK -->|Yes| S0_SOCRATIC[Socratic] --> S0_DISCOVER
        S0_ASK -->|No| S0_LIST[List items + diagram]
    end

    S0_LIST --> G0{Inventory complete?}
    G0 -->|No| S0_DISCOVER
    G0 -->|Yes| S1

    subgraph S1["Stage 1: Details"]
        S1_CONTAINER[Per container] --> S1_INT[Internal comp]
        S1_CONTAINER --> S1_LINK[Linkage comp]
        S1_INT --> S1_REF[Extract refs]
        S1_LINK --> S1_REF
        S1_REF --> S1_ASK{Questions?}
        S1_ASK -->|Yes| S1_SOCRATIC[Socratic] --> S1_CONTAINER
        S1_ASK -->|No| S1_NEXT{More?}
        S1_NEXT -->|Yes| S1_CONTAINER
    end

    S1_NEXT -->|No| G1{Fix inventory?}
    G1 -->|Yes| S0_DISCOVER
    G1 -->|No| S2

    subgraph S2["Stage 2: Finalize"]
        S2_CHECK[Integrity checks]
    end

    S2_CHECK --> G2{Issues?}
    G2 -->|Inventory| S0_DISCOVER
    G2 -->|Detail| S1_CONTAINER
    G2 -->|None| DONE([Implemented])
```
## Stage 0: Inventory
### Context Discovery

| Arg | Value |
| --- | --- |
| PROJECT |  |
| GOAL |  |
| SUMMARY |  |
### Abstract Constraints

| Constraint | Rationale | Affected Containers |
| --- | --- | --- |
|  |  |  |
### Container Discovery

| N | CONTAINER_NAME | BOUNDARY | GOAL | SUMMARY |
| --- | --- | --- | --- | --- |
| 1 |  |  |  |  |
| 2 |  |  |  |  |
### Component Discovery (Brief)

| N | NN | COMPONENT_NAME | CATEGORY | GOAL | SUMMARY |
| --- | --- | --- | --- | --- | --- |
|  |  |  | foundation (01-09) |  |  |
|  |  |  | feature (10+) |  |  |
### Ref Discovery

| SLUG | TITLE | GOAL | Scope | Applies To |
| --- | --- | --- | --- | --- |
|  |  |  |  |  |
### Overview Diagram

```mermaid
graph TD
    %% Fill after discovery
```
### Gate 0

- [ ] Context args filled
- [ ] Abstract Constraints identified
- [ ] All containers identified with args (including BOUNDARY)
- [ ] All components identified (brief) with args and category
- [ ] Cross-cutting refs identified
- [ ] Overview diagram generated
## Stage 1: Details
### Container: c3-1

**Created:** [ ] `.c3/c3-1-{slug}/README.md`

| Type | Component ID | Name | Category | Doc Created |
| --- | --- | --- | --- | --- |
| Internal |  |  |  | [ ] |
| Linkage |  |  |  | [ ] |
### Container: c3-N

_(repeat per container from Stage 0)_

### Refs Created

| Ref ID | Pattern | Doc Created |
| --- | --- | --- |
|  |  | [ ] |
### Gate 1

- [ ] All container README.md created
- [ ] All component docs created
- [ ] All refs documented
- [ ] No new items discovered (else -> Gate 0)
## Stage 2: Finalize
### Integrity Checks

| Check | Status |
| --- | --- |
| Context <-> Container (all containers listed in c3-0) | [ ] |
| Container <-> Component (all components listed in container README) | [ ] |
| Component <-> Component (linkages documented) | [ ] |
| * <-> Refs (refs cited correctly, Cited By updated) | [ ] |
### Gate 2

- [ ] All integrity checks pass
- [ ] Run audit
## Conflict Resolution

If later stage reveals earlier errors:

| Conflict | Found In | Affects | Resolution |
| --- | --- | --- | --- |
|  |  |  |  |
## Exit

When Gate 2 complete -> change frontmatter status to `implemented`

## Audit Record

| Phase | Date | Notes |
| --- | --- | --- |
| Adopted | 20260227 | Initial C3 structure created |
