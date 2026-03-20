---
id: adr-20260320-add-rule-entity-type
title: Add rule as first-class entity type
type: adr
status: implemented
date: 2026-03-20
affects: [c3-101, c3-102, c3-103, c3-105, c3-110, c3-111, c3-112, c3-113, c3-114, c3-115, c3-116, c3-201, c3-210]
---

# Add rule as first-class entity type

## Goal

Update all C3 architecture docs to reflect the new `rule` entity type — enforceable coding standards separate from architectural decisions (refs). Code is already implemented; docs need to catch up.

## Work Breakdown

1. Update component goals/descriptions that now handle rules (c3-101 through c3-116)
2. Update container README Components tables
3. Update context README to mention rules
4. Update c3-201 (skill-router) for 7 operations (was 6)
5. Update c3-210 (operation-refs) to include rule operation
6. Run c3x check to verify structural integrity

## Risks

None — docs-only change to reflect already-implemented code.
