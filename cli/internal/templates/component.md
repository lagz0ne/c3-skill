---
id: c3-${N}${NN}
c3-version: 4
title: ${COMPONENT_NAME}
type: component
category: ${CATEGORY}
parent: c3-${N}
goal: ${GOAL}
summary: ${SUMMARY}
---

# ${COMPONENT_NAME}

## Goal

${GOAL}

## Parent Fit

| Field | Value |
|-------|-------|
| Parent | c3-${N} |
| Role | Explain the exact responsibility this component owns for its parent container. |
| Boundary | Name what stays inside this component and what must remain outside it. |
| Collaboration | Name the adjacent component, ref, rule, or caller this component must coordinate with. |

## Purpose

Explain the durable reason this component exists, the behavior it owns, and at least one non-goal future agents must not move into it.

## Foundational Flow

| Aspect | Detail | Reference |
|--------|--------|-----------|
| Preconditions | State what must already be true before this component can run or be used. | N.A - replace when a governing ref, rule, or component exists |
| Inputs | Name the concrete data, files, commands, events, or calls consumed. | N.A - replace when a governing ref, rule, or component exists |
| State / data | Describe what state is read, written, derived, cached, or deliberately not stored. | N.A - replace when a governing ref, rule, or component exists |
| Shared dependencies | Name shared helpers, lower-layer contracts, or platform constraints. | N.A - replace when a governing ref, rule, or component exists |

## Business Flow

| Aspect | Detail | Reference |
|--------|--------|-----------|
| Actor / caller | Name who or what asks this component to do work. | N.A - replace when a governing ref, rule, or component exists |
| Primary path | Describe the intended successful workflow in operational terms. | N.A - replace when a governing ref, rule, or component exists |
| Alternate paths | Describe valid alternate behavior and why it is allowed. | N.A - replace when a governing ref, rule, or component exists |
| Failure behavior | Describe how failure is surfaced, contained, logged, or retried. | N.A - replace when a governing ref, rule, or component exists |

## Governance

| Reference | Type | Governs | Precedence | Notes |
|-----------|------|---------|------------|-------|
| N.A - no governing reference chosen yet | N.A - no reference type yet | State what future ref, rule, ADR, spec, policy, or example should govern. | Parent container contract wins until replaced. | Replace before review when a governing source exists. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
|---------|-----------|----------|----------|----------|
| Input surface | IN | State what callers must provide and what this component may assume. | Name the boundary crossed by the input. | Name the test, command, file, or review evidence. |
| Output surface | OUT | State what this component returns, writes, emits, or guarantees. | Name the boundary crossed by the output. | Name the test, command, file, or review evidence. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
|------|---------|-----------|-----------------------|
| Contract drift | Inputs, outputs, or boundaries change. | Compare Contract and Parent Fit against consumers. | Run the narrow tests or C3 lookup proving consumers still match. |
| Governance drift | A referenced rule, ref, ADR, spec, or policy changes. | Re-read Governance rows and cited documents. | Run c3x check plus the component-specific verification. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
|----------|------------------|------------------|----------|
| Code / docs / tests | Goal, Contract, Change Safety, and Governance. | Names and framework shape may vary; behavior and boundaries may not. | Name the command, file, or review artifact proving derivation. |
