---
id: c3-101
c3-version: 4
c3-seal: 20da22a5af33dd3d4840e061396aadd647e7f7e6fdaea5b5945208aa6b4597b2
title: frontmatter
type: component
category: foundation
parent: c3-1
goal: Parse and write YAML frontmatter embedded in `.c3/` markdown files.
summary: Provides Get/Set access to frontmatter fields; used by every command that reads entity metadata
uses:
    - ref-frontmatter-docs
---

# frontmatter

## Goal

Parse and write YAML frontmatter embedded in `.c3/` markdown files.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own frontmatter behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep frontmatter decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for frontmatter so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before frontmatter behavior is changed. | c3-1 |
| Inputs | Accept only the files, commands, data, or calls that belong to frontmatter ownership. | c3-1 |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | c3-1 |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | c3-1 |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks frontmatter to deliver its documented responsibility. | c3-1 |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | c3-1 |
| Alternate paths | When a request falls outside frontmatter ownership, hand it to the parent or sibling component. | c3-1 |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | c3-1 |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-1 | policy | Governs frontmatter behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |
| ref-frontmatter-docs | ref | frontmatter.go ParseFrontmatter is the canonical implementation: enforces the leading --- line and \n---\n terminator that define every .c3 doc, and owns the c3-seal field. | Cited ref contract beats uncited local prose. | Changes to the frontmatter delimiter or seal placement must comply with ref-frontmatter-docs. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| frontmatter input | IN | Callers must provide context that matches the component goal and parent fit. | c3-1 boundary | c3x lookup plus targeted tests or review. |
| frontmatter output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-1 boundary | c3x check and project test suite. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |
| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| Code, docs, tests, prompts | Goal, Governance, Contract, and Change Safety sections. | Names and framework shape may vary; behavior and boundaries may not. | c3x check, c3x verify, and relevant tests. |
