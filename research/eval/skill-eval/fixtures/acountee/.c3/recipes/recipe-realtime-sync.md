---
id: recipe-realtime-sync
c3-seal: a9c83e34549a93f1e03500ad6d7c696bb2f40488e0faecbfd68327dedd637961
title: Real-time Sync
type: recipe
goal: Trace the real-time sync path from server mutation to client state update via NATS pub/sub.
sources:
    - c3-0
    - c3-1
    - c3-2
    - c3-202
    - c3-211
    - ref-sync
---

# Real-time Sync

## Goal

Trace the real-time sync path from server mutation to client state update via NATS pub/sub.

## Narrative

Real-time sync uses a two-layer NATS pattern: services emit **deltas**
(entity-level changesets), flows emit **acks** (operation complete).

The `executionId` is the linchpin — a string value threaded from HTTP
response → execution context tag → sync.emit/ack → client tracker.
The originating client calls `result.wait()` which resolves when the
matching ack arrives via NATS broadcast. All other clients apply the
delta immediately.

NATS subjects:

- `{prefix}.broadcast` — deltas + acks to all connected users
- `{prefix}.user.{escaped_email}` — per-user notifications

Delta payloads carry full records in `ChangeSet<T>` with `add`, `update`,
`delete` arrays. Client-side `applyDelta` processes in order:
delete → update → add. Partial records are an anti-pattern.

The notification system (c3-211) uses a separate NATS JetStream stream
(`NOTIFICATIONS`) with workqueue semantics. It operates independently
from broadcast sync — notifications are durable, sync is ephemeral.

## Critical Rules

- Services own `sync.emit()` (after DB write)
- Flows own `sync.ack()` (at end, guarded: `if (executionId)`)
- Never call `sync.ack()` inside a service
- Keep executionId as string end-to-end (no numeric coercion)
- Always send full records in deltas

## Risk

Sync and notifications share NATS but are architecturally separate.
Confusing the two (e.g., trying to ack a notification) would break both.
