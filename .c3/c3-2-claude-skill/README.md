---
id: c3-2
c3-version: 4
c3-seal: dcca17b0c864fe5818f8c3096fe799275ec0221238a3f49af88af54f9dce49b6
title: Claude Skill
type: container
boundary: app
parent: c3-0
goal: Expose c3 architecture workflows through natural language by routing user intent to the right operation and executing it via c3x.
summary: SKILL.md intent router + per-operation reference docs that orchestrate AI-driven c3x workflows inside Claude Code sessions
---

# Claude Skill
## Goal

Expose c3 architecture workflows through natural language by routing user intent to the right operation and executing it via c3x.

## Responsibilities

- Classify natural language intent into supported C3 operations.
- Load and execute the appropriate operation workflow component.
- Call c3x commands at each step of an operation.
- Apply ASSUMPTION_MODE when the user declines to answer clarifying questions.
## Complexity Assessment

**Level:** moderate
**Why:** Intent classification must be reliable across varied phrasings; each operation is a multi-step workflow; skill descriptions are hard-constrained to 1024 chars.

## Components

| ID | Name | Category | Status | Goal Contribution |
| --- | --- | --- | --- | --- |
| c3-201 | skill-router | foundation | active | Classifies intent and dispatches to the matching workflow component. |
| c3-210 | operation-workflow-index | feature | active | Defines the shared operation-reference contract followed by each operation component. |
| c3-211 | onboard-operation | feature | active | Demonstrates project adoption and initial C3 topology creation. |
| c3-212 | query-operation | feature | active | Demonstrates architecture question answering through search, lookup, read, graph, and impact commands. |
| c3-213 | audit-operation | feature | active | Demonstrates structural and semantic C3 documentation quality review. |
| c3-214 | change-operation | feature | active | Demonstrates ADR-first architecture changes with lookup, parent delta, implementation, and verification. |
| c3-215 | migrate-operation | feature | active | Demonstrates C3 upgrade and cache recovery flows without treating c3.db as submitted truth. |
| c3-216 | ref-operation | feature | active | Demonstrates reusable architectural ref creation, update, listing, and explanation. |
| c3-217 | rule-operation | feature | active | Demonstrates enforceable coding rule creation, update, adoption, and evaluation. |
| c3-218 | sweep-operation | feature | active | Demonstrates transitive impact assessment before high-risk architecture changes. |
