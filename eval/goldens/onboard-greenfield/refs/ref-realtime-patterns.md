---
id: ref-realtime-patterns
title: Real-time Communication Patterns
---

# Real-time Communication Patterns

## Goal

Establish WebSocket communication patterns for real-time collaboration including connection lifecycle, message protocols, room management, and presence synchronization.

## Overview

Conventions for WebSocket-based real-time features.

## Connection Lifecycle

### States

| State | Description | Next |
|-------|-------------|------|
| Disconnected | No connection | Connecting |
| Connecting | Handshake in progress | Connected, Failed |
| Connected | Ready for messages | Disconnected |
| Reconnecting | Auto-retry after disconnect | Connected, Failed |
| Failed | Max retries exceeded | Manual reconnect |

### Reconnection Strategy

| Attempt | Delay |
|---------|-------|
| 1 | 1s |
| 2 | 2s |
| 3 | 4s |
| 4 | 8s |
| 5+ | 15s (cap) |

## Message Protocol

### Envelope

| Field | Type | Description |
|-------|------|-------------|
| type | string | Event type |
| payload | object | Event data |
| timestamp | number | Server timestamp |
| correlationId | string | Request tracking |

### Event Types

| Type | Direction | Purpose |
|------|-----------|---------|
| cursor_move | Bi | Cursor position update |
| concept_update | Server | Concept changed |
| presence_join | Server | User joined canvas |
| presence_leave | Server | User left canvas |
| ping | Client | Keep-alive |
| pong | Server | Keep-alive response |

## Room Management

### Room Lifecycle

| Event | Action |
|-------|--------|
| Join | Subscribe to canvas events |
| Leave | Unsubscribe, cleanup |
| Kick | Server-initiated leave |

### Room Capacity

| Scenario | Behavior |
|----------|----------|
| Under limit | Join succeeds |
| At limit | Queue or reject |
| Over limit (scaling) | Load balance across instances |

## Presence

### Presence Data

| Field | Update Frequency |
|-------|------------------|
| cursor position | 20Hz max |
| selection | On change |
| viewport | On change |
| activity status | On idle transition |

### Cursor Optimization

| Technique | Description |
|-----------|-------------|
| Throttle | Max 20 updates/second |
| Batch | Combine multiple users |
| Interpolate | Client-side smoothing |
| Cull | Skip off-screen cursors |

## Error Handling

| Error | Recovery |
|-------|----------|
| Connection lost | Auto-reconnect |
| Auth expired | Re-authenticate, reconnect |
| Room not found | Redirect to canvas list |
| Rate limited | Back off, queue messages |

## Cited By

- c3-401 (Connection Manager)
- c3-402 (Presence State)
