---
id: adr-20260121-add-archive
title: Add task archiving feature
status: proposed
date: 2026-01-21
affects: [c3-103]
approved-files:
  - src/routes/tasks.ts
---

# Add task archiving feature

## Status

**Proposed** - 2026-01-21

## Problem

Users want to archive completed tasks instead of deleting them.

## Decision

Add archive/unarchive endpoints to task routes (c3-103).

## Rationale

Archiving is a task operation that fits within c3-103's responsibility of task CRUD.

| Considered | Rejected Because |
|------------|------------------|
| Soft delete | Less explicit |

## Affected Layers

| Layer | Document | Change |
|-------|----------|--------|
| Component | c3-103 | Add archive endpoints |

## Approved Files

```yaml
approved-files:
  - src/routes/tasks.ts
```

## Verification

- [ ] Archive endpoint works
- [ ] Unarchive endpoint works
- [ ] Archived tasks excluded from list by default
