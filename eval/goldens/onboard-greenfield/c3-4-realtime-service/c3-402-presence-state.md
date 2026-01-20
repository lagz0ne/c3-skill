---
id: c3-402
c3-version: 3
title: Presence State
type: component
parent: c3-4
category: foundation
summary: User cursor positions and activity tracking
---

# Presence State

## Goal

Track and broadcast user presence including cursor positions, selection state, and activity indicators for collaborative awareness.

## Contract

From c3-4 (Real-time Service): "User cursor and activity tracking"

## Interface Diagram

```mermaid
flowchart LR
    subgraph IN["Receives"]
        I1[Cursor position]
        I2[Selection state]
        I3[Activity events]
    end

    subgraph PROCESS["Owns"]
        P1[Position aggregation]
        P2[State merge]
        P3[Stale detection]
    end

    subgraph OUT["Provides"]
        O1[Presence updates]
        O2[User list]
        O3[Activity indicators]
    end

    IN --> PROCESS --> OUT
```

## Hand-offs

| Direction | What | To/From |
|-----------|------|---------|
| IN | Cursor position updates | c3-1 Frontend |
| IN | Selection changes | c3-1 Frontend |
| IN | Connect/disconnect | c3-401 Connection Manager |
| OUT | Presence broadcasts | c3-403 Event Router |
| OUT | Initial presence state | c3-401 (on room join) |

## Presence Data Model

| Field | Type | Update Frequency |
|-------|------|------------------|
| userId | UUID | Static |
| displayName | string | Static |
| avatarUrl | string | Static |
| cursorX | float | High (throttled) |
| cursorY | float | High (throttled) |
| selectedIds | UUID[] | Medium |
| lastActivity | timestamp | On any action |
| viewportBounds | {x,y,w,h} | Low |

## State Flow

```mermaid
sequenceDiagram
    participant U1 as User 1
    participant PS as Presence State
    participant R as Redis
    participant U2 as User 2

    U1->>PS: Join room
    PS->>PS: Add to local state
    PS->>R: Publish presence_join
    PS-->>U1: Current room presence
    R-->>PS: (broadcast)
    PS-->>U2: User 1 joined

    loop Cursor updates
        U1->>PS: Cursor move (throttled)
        PS->>PS: Update local
        PS->>R: Publish cursor batch
        R-->>PS: (broadcast)
        PS-->>U2: Cursor update
    end

    U1->>PS: Disconnect
    PS->>PS: Remove from local
    PS->>R: Publish presence_leave
    PS-->>U2: User 1 left
```

## Conventions

| Rule | Why |
|------|-----|
| Throttle cursor to 20Hz | Network efficiency |
| Batch cursor updates | Reduce message count |
| Stale after 60s inactivity | Show idle indicator |
| Remove after 5min stale | Clean up |

## Cursor Interpolation

| Client Behavior | Server Support |
|-----------------|----------------|
| Interpolate between updates | Send velocity hint |
| Smooth animation | Include timestamp |
| Predict next position | Linear extrapolation |

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Very fast cursor movement | Downsample, keep latest |
| Same user in two tabs | Show both cursors |
| User goes idle | Mark as idle after 60s |
| Redis unavailable | Local-only mode, warn |

## References

- Presence handler: `src/ws/presence.ts`
- Cursor batching: `src/ws/cursor-batch.ts`
- Cites: ref-realtime-patterns, ref-collaborative-editing
