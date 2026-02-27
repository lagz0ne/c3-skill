---
id: c3-112
c3-version: 4
title: list-cmd
type: component
category: feature
parent: c3-1
goal: Output the full .c3/ topology in human-readable or machine-readable formats
summary: Supports --flat, --compact, and --json output modes; used by skill as the precondition check
uses: [c3-101, c3-102]
---

# list-cmd

## Goal

Output the full `.c3/` topology in human-readable or machine-readable formats.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Full entity graph | c3-102 |
| OUT (provides) | Topology output (flat/compact/JSON) |  |

## Code References

| File | Purpose |
|------|---------|
| `cli/cmd/list.go` | List command |
