---
id: c3-210
c3-version: 4
title: operation-refs
type: component
category: feature
parent: c3-2
goal: Provide step-by-step execution guidance for each of the six c3 operations
summary: Six reference docs loaded on demand by skill-router; each defines preconditions, stages, gates, and final checks for its operation
uses: [c3-201]
---

# operation-refs

## Goal

Provide step-by-step execution guidance for each of the six c3 operations.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Classified intent | c3-201 |
| OUT (provides) | Step-by-step operation guidance |  |

## Code References

| File | Purpose |
|------|---------|
| `skills/c3/references/onboard.md` | Onboard operation (init + discovery + docs + codemap) |
| `skills/c3/references/query.md` | Query operation (topology + doc navigation) |
| `skills/c3/references/audit.md` | Audit operation (check + semantic phases) |
| `skills/c3/references/change.md` | Change operation (impact + ADR + lookup + edit) |
| `skills/c3/references/ref.md` | Ref operation (add/update/list patterns) |
| `skills/c3/references/sweep.md` | Sweep operation (impact assessment) |
