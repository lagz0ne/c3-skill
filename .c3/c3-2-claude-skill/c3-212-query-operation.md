---
id: c3-212
c3-seal: 2b8496381efc8ce7cd38097d40d68baa279a6194bb06e51b15da6bb795be5e97
title: query-operation
type: component
category: feature
parent: c3-2
goal: Demonstrate how the skill answers architecture questions by using C3 search, topology, lookup, read, graph, and impact commands before reasoning.
uses:
    - c3-201
    - c3-210
---

# query-operation

## Goal

Demonstrate how the skill answers architecture questions by using C3 search, topology, lookup, read, graph, and impact commands before reasoning.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own query-operation behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep query-operation decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for query-operation so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before query-operation behavior is changed. | c3-2 |
| Inputs | Accept only the files, commands, data, or calls that belong to query-operation ownership. | c3-2 |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | c3-2 |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | c3-2 |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks query-operation to deliver its documented responsibility. | c3-2 |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | c3-2 |
| Alternate paths | When a request falls outside query-operation ownership, hand it to the parent or sibling component. | c3-2 |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | c3-2 |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-2 | policy | Governs query-operation behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| query-operation input | IN | Callers must provide context that matches the component goal and parent fit. | c3-2 boundary | c3x lookup plus targeted tests or review. |
| query-operation output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-2 boundary | c3x check and project test suite. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |
| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| Code, docs, tests, prompts | Goal, Governance, Contract, and Change Safety sections. | Names and framework shape may vary; behavior and boundaries may not. | c3x check, c3x verify, and relevant tests. |
