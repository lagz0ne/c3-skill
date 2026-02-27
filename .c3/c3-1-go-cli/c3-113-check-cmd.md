---
id: c3-113
c3-version: 4
title: check-cmd
type: component
category: feature
parent: c3-1
goal: Validate structural integrity of .c3/ docs — required fields, numbering, wiring completeness
summary: Reports PASS/WARN/FAIL for each entity; skill uses this as the final gate in onboard and change operations
uses: [c3-101, c3-102, c3-104]
---

# check-cmd

## Goal

Validate structural integrity of `.c3/` docs — required fields, numbering, wiring completeness.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Entity graph | c3-102 |
| IN (uses) | Citation validation | c3-104 |
| OUT (provides) | PASS/WARN/FAIL report |  |

## Code References

| File | Purpose |
|------|---------|
| `cli/cmd/check_enhanced.go` | Enhanced check with structured output |
