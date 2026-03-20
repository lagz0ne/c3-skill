---
id: c3-114
c3-version: 4
title: lookup-cmd
type: component
category: feature
parent: c3-1
goal: Map a file path or glob to the component(s), refs, and rules that govern it
summary: The primary "file context" tool — skill calls this before reading or editing any file
uses: [c3-101, c3-102, c3-105]
---

# lookup-cmd

## Goal

Map a file path or glob to the component(s), refs, and rules that govern it.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Entity graph | c3-102 |
| IN (uses) | Codemap matching | c3-105 |
| OUT (provides) | Component + ref mapping for a file |  |

## Code References

| File | Purpose |
|------|---------|
| `cli/cmd/lookup.go` | Lookup command |
