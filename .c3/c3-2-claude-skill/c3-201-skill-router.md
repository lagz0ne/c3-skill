---
id: c3-201
c3-version: 4
c3-seal: f68d317e96b60e6aa5b4419502085a1df329fc47534f0e69f3c3c8a668c75ce6
title: skill-router
type: component
category: foundation
parent: c3-2
goal: Classify user intent into a supported C3 operation and dispatch to the matching workflow component.
summary: SKILL.md entry point — the only file Claude Code loads; must fit triggering constraints (≤1024 chars description)
---

# skill-router

## Goal

Classify user intent into a supported C3 operation and dispatch to the matching workflow component.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own skill-router behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep skill-router decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for skill-router so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before skill-router behavior is changed. | c3-2 |
| Inputs | Accept only the files, commands, data, or calls that belong to skill-router ownership. | c3-2 |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | c3-2 |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | c3-2 |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks skill-router to deliver its documented responsibility. | c3-2 |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | c3-2 |
| Alternate paths | When a request falls outside skill-router ownership, hand it to the parent or sibling component. | c3-2 |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | c3-2 |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-2 | policy | Governs skill-router behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| intent classification | IN | Skill-router maps user intent to an operation reference and then lets c3x commands provide enforcement, hints, schemas, and failure guidance. | c3-201 router boundary; operation detail belongs to c3-210 through c3-218. | skills/c3/SKILL.md Dispatch and Enforcement source sections. |
| operation dispatch | OUT | Dispatch loads the matching reference workflow, but reference prose must route actionable steps through c3x rather than becoming a second policy engine. | c3-2 skill container boundary. | skills/c3/SKILL.md; skills/c3/references/change.md. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |
| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| skills/c3/SKILL.md | Contract intent classification row, Contract operation dispatch row, and Governance c3-2 policy. | Instruction copy can stay concise; it must state c3x is the enforcement source and skill is reference routing. | rg "Enforcement source" skills/c3/SKILL.md. |
| skills/c3/references/*.md | Contract operation dispatch row and Change Safety contract drift risk. | Each operation can have local sequence hints; failures, schemas, and repair steps must come from c3x where available. | rg "c3x" skills/c3/references. |
