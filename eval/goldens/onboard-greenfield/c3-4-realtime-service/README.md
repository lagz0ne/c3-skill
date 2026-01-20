---
id: c3-4
c3-version: 3
title: Real-time Service
type: container
parent: c3-0
summary: Live cursor sync and collaborative editing via WebSocket
---

# Real-time Service

## Complexity Assessment

**Level:** moderate
**Why:** WebSocket connection management with presence state, but single-purpose service with clear boundaries. Horizontal scaling adds complexity.

## Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| Runtime | Bun | WebSocket native support |
| Protocol | WebSocket | Bidirectional real-time |
| Pub/Sub | Redis | Cross-instance message distribution |
| Serialization | MessagePack | Efficient binary encoding |

## Components

| ID | Name | Category | Responsibility | Status |
|----|------|----------|----------------|--------|
| c3-401 | Connection Manager | foundation | WebSocket lifecycle and authentication | Documented |
| c3-402 | Presence State | foundation | User cursor and activity tracking | Documented |
| c3-403 | Event Router | foundation | Message routing between clients | |
| c3-411 | Cursor Broadcast | feature | Live cursor position sharing | |
| c3-412 | Concept Sync | feature | Real-time concept updates | |

## Internal Structure

```mermaid
graph TB
    subgraph Foundation["Foundation"]
        C401[c3-401 Connection Mgr]
        C402[c3-402 Presence State]
        C403[c3-403 Event Router]
    end

    subgraph Feature["Feature"]
        C411[c3-411 Cursor Broadcast]
        C412[c3-412 Concept Sync]
    end

    subgraph External["External"]
        Client[c3-1 Frontend]
        Redis[(Redis Pub/Sub)]
    end

    Client <-->|WebSocket| C401
    C401 --> C402
    C401 --> C403
    C403 --> C411
    C403 --> C412
    C403 <--> Redis
```

## Message Flow

```mermaid
sequenceDiagram
    participant C1 as Client 1
    participant WS as Real-time Service
    participant R as Redis
    participant C2 as Client 2

    C1->>WS: Connect + Auth
    WS->>WS: Add to canvas room
    WS->>R: Publish presence join
    R-->>WS: (other instances)

    C1->>WS: Cursor move
    WS->>R: Publish cursor event
    R-->>WS: Broadcast
    WS-->>C2: Cursor update

    C1->>WS: Concept updated
    WS->>R: Publish concept event
    R-->>WS: Broadcast
    WS-->>C2: Concept update
```

## Fulfillment

| Linkage | Component | How |
|---------|-----------|-----|
| c3-1 <-> c3-4 | c3-401 | WebSocket with auth token |
| c3-4 -> c3-3 | c3-402 | Presence state persistence (optional) |
