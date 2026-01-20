---
id: c3-104
c3-version: 3
title: State Atoms
type: component
parent: c3-1
category: foundation
summary: Global state management for canvas, concepts, and UI state
---

# State Atoms

## Goal

Provide reactive, granular state management that enables efficient re-renders and clear data flow between canvas components.

## Contract

From c3-1 (Web Frontend): "Global state management for canvas and UI"

## Interface Diagram

```mermaid
flowchart LR
    subgraph IN["Receives"]
        I1[User actions]
        I2[API responses]
        I3[Real-time events]
    end

    subgraph PROCESS["Owns"]
        P1[Canvas state]
        P2[Concept state]
        P3[UI state]
        P4[Derived selectors]
    end

    subgraph OUT["Provides"]
        O1[Reactive values]
        O2[Dispatch actions]
        O3[Subscriptions]
    end

    IN --> PROCESS --> OUT
```

## Hand-offs

| Direction | What | To/From |
|-----------|------|---------|
| IN | Concept CRUD results | c3-2 API Backend |
| IN | Position/presence updates | c3-4 Real-time |
| IN | User interactions | Canvas components |
| OUT | Canvas viewport state | c3-101 Canvas Engine |
| OUT | Concept data | c3-102 Concept Node |
| OUT | Selection state | All feature components |

## State Domains

| Domain | Contents | Persistence |
|--------|----------|-------------|
| Canvas | viewport, zoom, pan position | Session only |
| Concepts | concept entities, positions | API-synced |
| Selection | selected IDs, multi-select | Session only |
| Presence | collaborator cursors, names | Real-time |
| UI | panels open, modal state | LocalStorage |

## Conventions

| Rule | Why |
|------|-----|
| Atoms are read-only externally | Mutations through actions only |
| Derived state via selectors | Single source of truth |
| Optimistic updates for positions | Smooth UX during collaboration |
| Batch real-time updates | Prevent render thrashing |

## State Flow

```mermaid
stateDiagram-v2
    [*] --> Loading: Initial load
    Loading --> Hydrated: API + localStorage
    Hydrated --> Syncing: Real-time connected
    Syncing --> Syncing: Collaborative edits
    Syncing --> Reconnecting: Connection lost
    Reconnecting --> Syncing: Reconnected
    Reconnecting --> Offline: Timeout
    Offline --> Syncing: Reconnected
```

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Conflicting position updates | Last-write-wins with server timestamp |
| Large concept sets (>500) | Pagination in concept atom |
| Stale real-time connection | Auto-reconnect with exponential backoff |
| LocalStorage quota exceeded | Graceful degradation, log warning |

## References

- Atom definitions: `src/atoms/index.ts`
- Canvas atoms: `src/atoms/canvas.ts`
- Concept atoms: `src/atoms/concepts.ts`
- Cites: ref-state-management
