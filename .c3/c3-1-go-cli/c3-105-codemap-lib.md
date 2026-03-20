---
id: c3-105
c3-version: 4
title: codemap-lib
type: component
category: foundation
parent: c3-1
goal: Parse code-map.yaml, match files to components and rules via glob patterns, validate map completeness
summary: Core library for all codemap operations — used by lookup, codemap-cmd, and coverage-cmd
uses: [c3-102]
---

# codemap-lib

## Goal

Parse `code-map.yaml`, match files to components and rules via glob patterns, validate map completeness.

## Container Connection

lookup-cmd, codemap-cmd, and coverage-cmd all reduce to "what does this file map to?" — that answer comes from this component.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Entity discovery | c3-102 |
| OUT (provides) | File-to-component mapping |  |

## Code References

| File | Purpose |
|------|---------|
| `cli/internal/codemap/codemap.go` | Parse + match logic |
| `cli/internal/codemap/coverage.go` | Coverage calculation |
| `cli/internal/codemap/validate.go` | Map validation |
