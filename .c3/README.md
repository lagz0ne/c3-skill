---
id: c3-0
c3-seal: 6c834f84d71ab4d179027153a90ed2ad2ffd91193d3e78680d6df2b924e1973f
title: c3-design
goal: 'Build and distribute C3 — a knowledge-graph architecture-docs tool that holds a codebase''s architecture as frozen, verifiable facts — shipped three ways: a Go CLI engine, a Claude skill, and an npm installer.'
---

# c3-design

## Goal

Build and distribute C3 — a knowledge-graph architecture-docs tool that holds a codebase's architecture as frozen, verifiable facts — shipped three ways: a Go CLI engine, a Claude skill, and an npm installer.

## Containers

| ID | Name | Boundary | Status | Responsibilities | Goal Contribution |
| --- | --- | --- | --- | --- | --- |
| c3-1 | Go CLI |  | active | Provide every c3x operation as a single cross-compiled Go binary — the engine that reads, writes, validates, and freezes the architecture graph. | Provide every c3x operation as a single cross-compiled Go binary — the engine that reads, writes, validates, and freezes the architecture graph. |
| c3-2 | Claude Skill |  | active | Teach an agent to operate C3 — route intent to the right operation and run each one through the local CLI binary. | Teach an agent to operate C3 — route intent to the right operation and run each one through the local CLI binary. |
| c3-3 | npm @c3x/cli |  | active | Install and run the c3x binary from npm — a thin client that downloads the right platform build and forwards arguments. | Install and run the c3x binary from npm — a thin client that downloads the right platform build and forwards arguments. |

## Abstract Constraints

| Constraint | Rationale | Affected Containers |
| --- | --- | --- |
| Architecture facts are frozen and mutate only through a change-unit | A shared contract that silently drifts is worse than none; freezing makes divergence detectable by construction | c3-1 |
| The CLI binary is the single source of behavior; the skill and npm client only wrap it | One implementation to verify; distribution surfaces stay thin and replaceable | c3-1, c3-2, c3-3 |
| Releases are cut by CI from a version tag, never by hand | Reproducible, checksummed artifacts across the platform matrix | c3-1, c3-3 |
