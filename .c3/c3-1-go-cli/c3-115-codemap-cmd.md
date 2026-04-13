---
id: c3-115
c3-version: 4
c3-seal: 4080f4d09294a9bb127a974260ee935b0034148be3b6da06868b65e647ad9b57
title: codemap-cmd
type: component
category: feature
parent: c3-1
goal: Scaffold or update `.c3/code-map.yaml` with empty stubs for all entities.
summary: Idempotent — safe to re-run; preserves existing patterns while adding missing stubs
uses:
    - c3-101
    - c3-102
    - c3-105
---

# codemap-cmd
## Goal

Scaffold or update `.c3/code-map.yaml` with empty stubs for all entities.

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN (uses) | Entity graph | c3-102 |
| IN (uses) | Codemap library | c3-105 |
| OUT (provides) | Updated code-map.yaml stubs |  |
## Code References

| File | Purpose |
| --- | --- |
| cli/cmd/codemap.go | Codemap scaffold command |
