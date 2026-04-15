---
id: c3-1
c3-version: 4
c3-seal: de0a6bcfd0f72b635fd9387ef12c2baf29763d42ac4d3e51f1ded987bd33607a
title: Go CLI
type: container
boundary: process
parent: c3-0
goal: Provide all c3x commands as a single cross-compiled binary that reads and writes `.c3/` architecture docs.
summary: Cross-compiled Go binary exposing init, add, list, check, lookup, codemap, coverage, and wire commands
---

## Goal

Provide all c3x commands as a single cross-compiled binary that reads and writes `.c3/` architecture docs.

## Responsibilities

- Own all file-system read/write for `.c3/` architecture docs
- Provide deterministic entity numbering for add operations
- Match source files to architecture components via codemap patterns
- Validate structural integrity of the doc tree
- Compile to a self-contained binary for all supported platforms
- Provide npm wrapper delegation for human/default CLI access without taking over skill-owned agent mode
## Complexity Assessment

**Level:** moderate
**Why:** Multiple commands with shared library layer; codemap glob matching + coverage calculation is non-trivial; wiring/citation tracking spans multiple doc formats.

## Components

| ID | Name | Category | Status | Goal Contribution |
| --- | --- | --- | --- | --- |
| c3-101 | frontmatter | foundation | active | Parses YAML frontmatter from all .c3/ markdown files |
| c3-102 | walker | foundation | active | Discovers and traverses the .c3/ entity tree |
| c3-103 | templates | foundation | active | Provides embedded scaffolding templates for new docs |
| c3-104 | wiring | foundation | active | Tracks citations between entities |
| c3-105 | codemap-lib | foundation | active | Parses, matches, and validates code-map.yaml patterns |
| c3-106 | content-lib | foundation | active | Parses, renders, and bridges structured markdown content |
| c3-107 | store-lib | foundation | active | Persists entities, relationships, changelog, codemap, hashes, nodes, and versions |
| c3-108 | runtime-support | foundation | active | Provides bootstrap, option parsing, output, config, and agent presentation helpers |
| c3-109 | npm-cli-wrapper | feature | active | Provides the npm @c3x/cli shim for discovery and human/default delegation |
| c3-110 | init-cmd | feature | active | Scaffolds .c3/ directory from scratch |
| c3-111 | add-cmd | feature | active | Creates container/component/ref/rule/adr entities with numbering |
| c3-112 | list-cmd | feature | active | Outputs topology as flat/compact/JSON |
| c3-113 | check-cmd | feature | active | Validates structural integrity, layer integration, refs, rules, and schema |
| c3-114 | lookup-cmd | feature | active | Maps file paths/globs to components plus governing refs and rules |
| c3-115 | codemap-cmd | feature | active | Scaffolds code-map.yaml stubs for all entities |
| c3-116 | coverage-cmd | feature | active | Reports code-map coverage and governance stats |
| c3-117 | docs-state-cmds | feature | active | Reads, writes, sets, validates schema, and reports status for C3 docs |
| c3-118 | analysis-cmds | feature | active | Queries, graphs, diffs, and impacts architecture state |
| c3-119 | sync-lifecycle-cmds | feature | active | Handles import, export, sync, repair, migrate, delete, and git guardrails |
| c3-120 | history-marketplace-cmds | feature | active | Handles versions, hash, nodes, prune, and marketplace commands |
