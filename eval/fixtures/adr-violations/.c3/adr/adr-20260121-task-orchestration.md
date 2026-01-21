---
id: adr-20260121-task-orchestration
title: Add request orchestration in task routes
status: proposed
date: 2026-01-21
affects: [c3-103]
approved-files:
  - src/routes/tasks.ts
---

# Add request orchestration in task routes

## Status

**Proposed** - 2026-01-21

## Problem

Complex task operations need to coordinate with user data.

## Decision

Task routes (c3-103) will coordinate requests between task-routes and user-routes, managing the flow based on operation type.

## Rationale

Some operations need data from both tasks and users.

| Considered | Rejected Because |
|------------|------------------|
| API composition at client | Client complexity |

## Affected Layers

| Layer | Document | Change |
|-------|----------|--------|
| Component | c3-103 | Add coordination logic |

## Approved Files

```yaml
approved-files:
  - src/routes/tasks.ts
```

## Verification

- [ ] Task routes can coordinate with user routes
- [ ] Flow is managed correctly
