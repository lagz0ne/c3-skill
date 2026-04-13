---
id: c3-113
c3-version: 4
c3-seal: 49e35a9778ecc2bbdf32d673da5c6004f4569e768158e2be6fcb93a40eb795ec
title: check-cmd
type: component
category: feature
parent: c3-1
goal: Validate structural integrity of `.c3/` docs, ref and rule compliance — required fields, numbering, wiring, scope cross-checks, origin validation.
summary: Reports PASS/WARN/FAIL for each entity; includes schema definitions, structural index building, and ref scope cross-checking
uses:
    - c3-101
    - c3-102
    - c3-104
---

# check-cmd
## Goal

Validate structural integrity of `.c3/` docs, ref and rule compliance — required fields, numbering, wiring, scope cross-checks, origin validation.

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN (uses) | Entity graph | c3-102 |
| IN (uses) | Citation validation | c3-104 |
| OUT (provides) | PASS/WARN/FAIL report |  |
## Code References

| File | Purpose |
| --- | --- |
| cli/cmd/check_enhanced.go | Enhanced check with structured output |
| cli/internal/schema/schema.go | Section schema definitions |
| cli/internal/index/index.go | Structural index builder |
