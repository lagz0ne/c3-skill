---
id: c3-116
c3-version: 4
title: coverage-cmd
type: component
category: feature
parent: c3-1
goal: Report what percentage of source files are mapped in code-map.yaml
summary: JSON output showing mapped, excluded, and unmapped file counts per component
uses: [c3-102, c3-105]
---

# coverage-cmd

## Goal

Report what percentage of source files are mapped in `code-map.yaml`.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Codemap matching | c3-105 |
| OUT (provides) | Coverage stats JSON |  |

## Code References

| File | Purpose |
|------|---------|
| `cli/cmd/coverage.go` | Coverage command |
