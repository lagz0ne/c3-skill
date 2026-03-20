---
id: c3-116
c3-version: 4
title: coverage-cmd
type: component
category: feature
parent: c3-1
goal: Report code-map coverage, ref governance, and rule governance metrics
summary: JSON output showing mapped, excluded, and unmapped file counts per component plus ref and rule governance percentages
uses: [c3-102, c3-105, c3-113]
---

# coverage-cmd

## Goal

Report code-map coverage, ref governance, and rule governance metrics.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Codemap matching | c3-105 |
| IN (uses) | Structural index / ref governance | c3-113 |
| OUT (provides) | Coverage stats JSON |  |

## Code References

| File | Purpose |
|------|---------|
| `cli/cmd/coverage.go` | Coverage command |
