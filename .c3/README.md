---
id: c3-0
c3-version: 4
c3-seal: ea10e3dba27354970d29c910a3e3e6f5900c8300efa6f27535b4d9fad084e5d9
title: c3-design
goal: Build and distribute the c3 Claude Code plugin — a CLI-driven architecture documentation system for large codebases.
summary: Pairs a cross-compiled Go CLI (c3x) with a Claude Code skill to create, navigate, and audit structured .c3/ architecture docs in any codebase
---

# c3-design

## Goal

Build and distribute the c3 Claude Code plugin — a CLI-driven architecture documentation system for large codebases.

## Abstract Constraints

| Constraint | Rationale | Affected Containers |
| --- | --- | --- |
| CLI must compile to 4 targets (linux/darwin × amd64/arm64) | Plugin users span platforms; no runtime deps allowed | Go CLI |
| Plugin distributed as a GitHub Releases zip; binaries bundled on main | Marketplace installs from a zip URL; binaries are gitignored on dev | Go CLI, Claude Skill |
| Skill text (description + triggers) ≤ 1024 chars per entity | Claude Code SDK limit for reliable skill triggering | Claude Skill |

## Containers

| ID | Name | Boundary | Status | Responsibilities | Goal Contribution |
| --- | --- | --- | --- | --- | --- |
| c3-1 | Go CLI | process | active | All c3x commands: init, add, list, check, lookup, codemap, coverage, wire | Provides the data-layer tools the skill uses to read/write .c3/ docs |
| c3-2 | Claude Skill | Claude Code session | active | Intent routing, workflow orchestration, AI reasoning over .c3/ docs | Surfaces c3x capabilities through natural language via Claude Code |
