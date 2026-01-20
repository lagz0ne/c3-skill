---
id: ref-state-management
title: State Management Patterns
---

# State Management Patterns

## Goal

Establish consistent patterns for frontend state management using atomic state (Jotai/Zustand style), ensuring predictable data flow, efficient updates, and proper synchronization with real-time and API sources.

## Overview

Conventions for frontend state using atomic state management (Jotai/Zustand style).

## State Categories

| Category | Persistence | Sync |
|----------|-------------|------|
| Canvas | Session | Real-time |
| Concepts | API | Real-time |
| Selection | Session | None |
| UI | LocalStorage | None |
| Presence | Memory | Real-time |

## Atom Principles

### Granularity

| Good | Bad |
|------|-----|
| `viewportAtom` | `canvasStateAtom` (too broad) |
| `selectedIdsAtom` | `uiAtom` (too broad) |
| `conceptByIdAtom(id)` | `allConceptsAtom` (for single access) |

### Derived State

| Pattern | Use |
|---------|-----|
| Selector atoms | Computed from base atoms |
| Family atoms | Parameterized (e.g., by ID) |
| Async atoms | Data fetching with suspense |

## Update Patterns

### Optimistic Updates

| Step | Action |
|------|--------|
| 1 | Update local state immediately |
| 2 | Send API request |
| 3 | On success: no-op (already updated) |
| 4 | On failure: rollback + show error |

### Batch Updates

| Scenario | Approach |
|----------|----------|
| Multiple position changes | Single transaction atom |
| Multiple selections | Set operation, not individual |
| Real-time bursts | Debounce + batch |

## Real-time Integration

### Inbound Events

| Event | State Update |
|-------|--------------|
| concept_created | Add to concepts |
| concept_updated | Merge into existing |
| concept_deleted | Remove from concepts |
| cursor_moved | Update presence |

### Conflict Resolution

| Conflict | Resolution |
|----------|------------|
| Local edit + remote edit | Remote wins, show indicator |
| Local delete + remote update | Delete wins |
| Concurrent moves | Last timestamp wins |

## Persistence

| Store | Data | Expiry |
|-------|------|--------|
| LocalStorage | UI preferences | Never |
| SessionStorage | Transient state | Tab close |
| IndexedDB | Offline cache | Manual |

## Cited By

- c3-104 (State Atoms)
- c3-113 (AI Chat Panel)
