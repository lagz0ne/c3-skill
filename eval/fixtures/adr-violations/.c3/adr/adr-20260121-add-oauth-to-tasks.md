---
id: adr-20260121-add-oauth-to-tasks
title: Add OAuth verification to task routes
status: proposed
date: 2026-01-21
affects: [c3-103]
approved-files:
  - src/routes/tasks.ts
---

# Add OAuth verification to task routes

## Status

**Proposed** - 2026-01-21

## Problem

Some tasks need OAuth-based verification for third-party integrations.

## Decision

Add OAuth token verification directly in task routes (c3-103) to handle OAuth callbacks.

## Rationale

Task routes need to verify OAuth tokens for integration features.

| Considered | Rejected Because |
|------------|------------------|
| Add to auth middleware | Would affect all routes |

## Affected Layers

| Layer | Document | Change |
|-------|----------|--------|
| Component | c3-103 | Add OAuth verification |

## Approved Files

```yaml
approved-files:
  - src/routes/tasks.ts
```

## Verification

- [ ] OAuth verification works in task routes
- [ ] Other routes unaffected
