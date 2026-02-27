---
id: c3-104
c3-version: 4
title: wiring
type: component
category: foundation
parent: c3-1
goal: Track and validate entity citations between components and refs
summary: Reads uses/via fields from frontmatter, resolves IDs, reports uncited refs and missing citations; used by check and wire commands
---

# wiring

## Goal

Track and validate entity citations between components and refs (`uses`/`via` frontmatter fields).

## Container Connection

check-cmd's semantic validation and wire-cmd's citation management both depend on this. Without wiring, there's no way to verify that ref usage is correctly documented.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Frontmatter parsing | c3-101 |
| IN (uses) | Entity discovery | c3-102 |
| OUT (provides) | Citation validation results |  |

## Code References

| File | Purpose |
|------|---------|
| `cli/internal/wiring/wiring.go` | Citation resolution and validation |
| `cli/cmd/wire.go` | Wire command using this library |
