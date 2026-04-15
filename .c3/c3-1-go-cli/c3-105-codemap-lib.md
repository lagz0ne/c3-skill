---
id: c3-105
c3-version: 4
c3-seal: c5d11167000eb07da3f209a0bcb3888a56eb41cee695c2939780f2f82fe433af
title: codemap-lib
type: component
category: foundation
parent: c3-1
goal: Parse `code-map.yaml`, match files to components and rules via glob patterns, validate map completeness.
summary: Core library for all codemap operations — used by lookup, codemap-cmd, and coverage-cmd
uses:
    - c3-102
---

# codemap-lib
## Goal

Parse `code-map.yaml`, match files to components and rules via glob patterns, validate map completeness.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own codemap-lib behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep codemap-lib decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |
## Purpose

Provide durable agent-ready documentation for codemap-lib so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before codemap-lib behavior is changed. | c3-1 |
| Inputs | Accept only the files, commands, data, or calls that belong to codemap-lib ownership. | c3-1 |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | c3-1 |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | c3-1 |
## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks codemap-lib to deliver its documented responsibility. | c3-1 |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | c3-1 |
| Alternate paths | When a request falls outside codemap-lib ownership, hand it to the parent or sibling component. | c3-1 |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | c3-1 |
## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-1 | policy | Governs codemap-lib behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |
## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| codemap-lib input | IN | Callers must provide context that matches the component goal and parent fit. | c3-1 boundary | c3x lookup plus targeted tests or review. |
| codemap-lib output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-1 boundary | c3x check and project test suite. |
## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |
| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |
## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| Code, docs, tests, prompts | Goal, Governance, Contract, and Change Safety sections. | Names and framework shape may vary; behavior and boundaries may not. | c3x check, c3x verify, and relevant tests. |
