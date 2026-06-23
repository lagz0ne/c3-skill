---
id: c3-0
c3-seal: dc67c917de1add2f05a1749f1beb9234471ff3be6a88875efe1a0dbf6a1e1265
title: c3-design
goal: 'Build and distribute C3 — a knowledge-graph architecture-docs tool that holds a codebase''s architecture as frozen, verifiable facts — shipped three ways: a Go CLI engine, an agent skill packaged for Claude and Codex, and an npm runtime manager.'
---

# c3-design

## Goal

Build and distribute C3 — a knowledge-graph architecture-docs tool that holds a codebase's architecture as frozen, verifiable facts — shipped three ways: a Go CLI engine, an agent skill packaged for Claude and Codex, and an npm runtime manager.

## Containers

| ID | Name | Boundary | Status | Responsibilities | Goal Contribution |
| --- | --- | --- | --- | --- | --- |
| c3-1 | Go CLI |  | active | Provide every c3x operation as a single cross-compiled Go binary — the engine that reads, writes, validates, and freezes the architecture graph. | Provide every c3x operation as a single cross-compiled Go binary — the engine that reads, writes, validates, and freezes the architecture graph. |
| c3-2 | Claude Skill |  | active | Teach an agent to operate C3 through shared skill instructions, Claude and Codex plugin packaging, and a wrapper that runs the selected C3 runtime. | Teach an agent to operate C3 through shared skill instructions, Claude and Codex plugin packaging, and a wrapper that runs the selected C3 runtime. |
| c3-3 | npm @c3x/cli |  | active | Install, manage, and run the c3x binary from npm — a thin client that serves local discovery commands, resolves verified GitHub Release runtimes, and forwards normal commands. | Install, manage, and run the c3x binary from npm — a thin client that serves local discovery commands, resolves verified GitHub Release runtimes, and forwards normal commands. |
| c3-4 | dev-tooling | service | active | Hold the standalone build/test programs that support the c3x CLI but ship separately from the binary — the search-ranking quality harness and the embedding-asset builder — so they are first-class facts with their own code surfaces rather than undescribed corners of the tree. | Hold the standalone build/test programs that support the c3x CLI but ship separately from the binary — the search-ranking quality harness and the embedding-asset builder — so they are first-class facts with their own code surfaces rather than undescribed corners of the tree. |

## Abstract Constraints

| Constraint | Rationale | Affected Containers |
| --- | --- | --- |
| Architecture facts are frozen and mutate only through a change-unit | A shared contract that silently drifts is worse than none; freezing makes divergence detectable by construction | c3-1 |
| The CLI binary is the single source of behavior; the skill and npm client only wrap it | One implementation to verify; distribution surfaces stay thin and replaceable | c3-1, c3-2, c3-3 |
| Releases are cut by CI from a version tag, never by hand | Reproducible, checksummed artifacts across the platform matrix | c3-1, c3-3 |
