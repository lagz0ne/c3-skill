---
id: c3-1
c3-version: 4
c3-seal: dc19609b2b14a635e7c23646fa059951c8879bcef4b4e51d578019727a274604
title: Go CLI
type: container
boundary: process
parent: c3-0
goal: Provide all c3x commands as a single cross-compiled binary that reads and writes `.c3/` architecture docs.
summary: Cross-compiled Go binary exposing init, add, list, check, lookup, codemap, coverage, and wire commands
---

# Go CLI
## Goal

Provide all c3x commands as a single cross-compiled binary that reads and writes `.c3/` architecture docs.

## Responsibilities

- Own all file-system read/write for `.c3/` architecture docs
- Provide deterministic entity numbering for add operations
- Match source files to architecture components via codemap patterns
- Validate structural integrity of the doc tree
- Compile to a self-contained binary for all supported platforms
## Complexity Assessment

**Level:** moderate
**Why:** Multiple commands with shared library layer; codemap glob matching + coverage calculation is non-trivial; wiring/citation tracking spans multiple doc formats.

## Components

| ID | Name | Category | Status | Goal Contribution |
| --- | --- | --- | --- | --- |
| c3-101 | frontmatter | foundation | active | Parses YAML frontmatter from all .c3/ markdown files |
| c3-102 | walker | foundation | active | Discovers and traverses the .c3/ entity tree (containers, components, refs, rules, ADRs) |
| c3-103 | templates | foundation | active | Provides embedded scaffolding templates for new docs (incl. rules) |
| c3-104 | wiring | foundation | active | Tracks uses:/via: citations between entities |
| c3-105 | codemap-lib | foundation | active | Parses, matches, and validates code-map.yaml patterns for components and rules |
| c3-110 | init-cmd | feature | active | Scaffolds .c3/ directory from scratch |
| c3-111 | add-cmd | feature | active | Creates container/component/ref/rule/adr entities with numbering |
| c3-112 | list-cmd | feature | active | Outputs topology as flat/compact/JSON |
| c3-113 | check-cmd | feature | active | Validates structural integrity of .c3/ docs, ref and rule compliance |
| c3-114 | lookup-cmd | feature | active | Maps file paths/globs to components + governing refs and rules |
| c3-115 | codemap-cmd | feature | active | Scaffolds code-map.yaml stubs for all entities |
| c3-116 | coverage-cmd | feature | active | Reports code-map coverage and rule governance stats |
