---
id: c3-102
c3-version: 4
title: walker
type: component
category: foundation
parent: c3-1
goal: Traverse the .c3/ directory tree to discover all containers, components, refs, and ADRs
summary: Canonical entity discovery used by list, check, lookup, and codemap commands
uses: [c3-101]
---

# walker

## Goal

Traverse the `.c3/` directory tree to discover all containers, components, refs, and ADRs.

## Container Connection

Any command that needs to know "what exists" depends on this. Without walker, list/check/lookup/codemap cannot enumerate entities.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | .c3/ directory path |  |
| IN (uses) | Frontmatter parsing | c3-101 |
| OUT (provides) | Ordered entity list with metadata |  |

## Code References

| File | Purpose |
|------|---------|
| `cli/internal/walker/walker.go` | Recursive .c3/ traversal |
