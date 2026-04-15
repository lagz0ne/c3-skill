---
id: c3-118
c3-seal: e94cd4d2c9d09da96f58ef5f845923d4248457f08e514a8dbb49b75990bec956
title: analysis-cmds
type: component
category: feature
parent: c3-1
goal: Query, graph, diff, and impact commands for understanding architecture state and change consequences.
---

# analysis-cmds
## Goal

Query, graph, diff, and impact commands for understanding architecture state and change consequences.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own analysis-cmds behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep analysis-cmds decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |
## Purpose

Provide durable agent-ready documentation for analysis-cmds so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before analysis-cmds behavior is changed. | c3-1 |
| Inputs | Accept only the files, commands, data, or calls that belong to analysis-cmds ownership. | c3-1 |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | c3-1 |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | c3-1 |
## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks analysis-cmds to deliver its documented responsibility. | c3-1 |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | c3-1 |
| Alternate paths | When a request falls outside analysis-cmds ownership, hand it to the parent or sibling component. | c3-1 |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | c3-1 |
## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-1 | policy | Governs analysis-cmds behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |
## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| analysis-cmds input | IN | Callers must provide context that matches the component goal and parent fit. | c3-1 boundary | c3x lookup plus targeted tests or review. |
| analysis-cmds output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-1 boundary | c3x check and project test suite. |
## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |
| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |
## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| Code, docs, tests, prompts | Goal, Governance, Contract, and Change Safety sections. | Names and framework shape may vary; behavior and boundaries may not. | c3x check, c3x verify, and relevant tests. |
