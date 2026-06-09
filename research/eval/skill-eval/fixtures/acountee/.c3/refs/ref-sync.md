---
id: ref-sync
c3-seal: 051a9c2b006ec69ea73737dce2e9c7d5559911e61a602d24400a48a21df87d1c
title: Real-time Sync Pattern
type: ref
goal: Propagate server-side mutations to all connected clients in real-time via NATS pub/sub. Clients receive granular deltas (not full state) and resolve optimistic UI updates deterministically via execution ID matching.
---

# Real-time Sync Pattern

## Goal

Propagate server-side mutations to all connected clients in real-time via NATS pub/sub. Clients receive granular deltas (not full state) and resolve optimistic UI updates deterministically via execution ID matching.

## Choice

Adopt a two-layer sync contract where services emit entity deltas and flows emit completion acknowledgements (`ack`) using a shared `executionId`.

## Why

Separating data propagation from completion signaling gives deterministic client reconciliation, supports concurrent updates, and avoids over-fetching full state snapshots.

## Architecture

Two-layer sync: **services emit deltas**, **flows send acks**. Both use an `executionId` so the originating client knows when its mutation landed.

```
  UI action
    |
    v
  Flow (orchestrates mutation)
    |--- calls service method via ctx.exec()
    |       |
    |       v
    |     Service (DB write)
    |       |--- sync.emit({ entity, type, id, data }, executionId)
    |       |       |
    |       |       v
    |       |     NATS broadcast --> all clients apply delta
    |       |
    |       v
    |     returns result
    |
    |--- sync.ack(executionId)  // after all service calls complete
    |       |
    |       v
    |     NATS broadcast --> originating client resolves wait()
    |
    v
  Return { success, executionId } to UI
```

Key insight: `sync.emit()` carries the data change. `sync.ack()` is the "I'm done" signal. The client's `executionTracker` resolves on whichever arrives first with that `executionId`.

## Execution ID Contract

`executionId` is the correlation key across HTTP, flow/service execution, NATS messages, and client wait/notify.

- Canonical type: **string** at all boundaries.
- The value returned by `/act` must exactly match the value used in `executionIdTag`, `sync.emit`, and `sync.ack`.
- Client `executionTracker` keys are string-based; type mismatch can prevent `notify()` from resolving `wait()`.
- `result.wait()` is a UX optimization, not correctness-critical; timeout fallback (2s) prevents permanent hangs.

## NATS Subjects

| Subject | Purpose | Publisher |
| --- | --- | --- |
| {prefix}.broadcast (default: sync.broadcast) | Deltas + acks to all users | publisher.publishToAll() |
| {prefix}.user.{escaped_email} (default: sync.user.{escaped_email}) | Notifications to specific user | publisher.publishToUser() |

Email escaping: `@` and `.` replaced with `_`.

## Subject Prefix Contract

- Server-side subjects are prefix-driven (`natsConfig.subjectPrefix`).
- Current default prefix is `sync` (`NATS_SUBJECT_PREFIX`).
- Frontend subscriptions currently use `sync.broadcast` and `sync.user.{escaped_email}` directly.
- If prefix changes from `sync`, frontend subscription wiring must change in lockstep.

## Delta Payload Shape

```typescript
// Wire format for a single entity type's changes
type ChangeSet<T> = {
  add: T[]       // new records
  update: T[]    // full replacement of changed records
  delete: number[] // IDs to remove
}

// Delta message broadcast via NATS
interface DeltaMessage {
  type: 'delta'
  changes: {
    invoices?: ChangeSet<InvoiceRecord>
    prs?: ChangeSet<PaymentRequestRecord>
    payments?: ChangeSet<PaymentMethodRecord>
  }
  executionId?: string
}

// Ack message — no data, just "executionId is done"
interface AckMessage {
  type: 'ack'
  executionId: string
}

// Notification message — targeted to specific user
interface NotificationMessage {
  type: 'notification'
  notification: {
    id: string
    notification_type: string
    payload: unknown
    created_at: string
  }
  executionId?: string
}
```

## Golden Example: Service Emitting a Delta

Services own the DB write and emit the delta with the changed record. The `executionId` comes from the execution context tag set by middleware.

```typescript
// Inside a service method (e.g., prService.approve)
const updatedPr = await db.update(prTable)
  .set({ status: 'approved' })
  .where(eq(prTable.id, prId))
  .returning()
  .then(rows => rows[0]);

// Emit delta — broadcasts full updated record to all clients
const executionId = execCtx.data.seekTag(executionIdTag);
await sync.emit(
  { entity: 'pr', type: 'update', id: prId, data: updatedPr },
  executionId
);
```

`sync.emit()` builds the `DeltaMessage` and calls `publisher.publishToAll()`. Each entity type maps to its key in `changes`:

- `entity: 'invoice'` → `changes.invoices`
- `entity: 'pr'` → `changes.prs`
- `entity: 'payment'` → `changes.payments`

## Golden Example: Flow Calling sync.ack()

Flows orchestrate one or more service calls, then ack. The ack tells the originating client "your mutation is fully complete."

```typescript
// Inside a flow handler
const { sync, executionId } = ctx.deps;

// Service calls (each one emits its own delta internally)
const result = await ctx.exec({
  fn: prService.approve,
  params: [ctx.input.prId, currentUser.email]
});
if (!result.success) return result;

// After ALL service work is done, ack the execution
if (executionId) {
  await sync.ack(executionId);
}
```

Every flow follows this pattern: do work, then `sync.ack(executionId)` at the end. The ack is always conditional on `executionId` existing (it may not in non-interactive contexts).

## Golden Example: Client-Side Subscription

The frontend subscribes to NATS via WebSocket. When a delta arrives, it applies changes to atoms using `applyDelta`. When an ack/delta carries an `executionId`, it resolves the matching `executionTracker` promise.

```typescript
// natsSync atom — subscribes to broadcast subject
const prefix = 'sync'; // from NATS_SUBJECT_PREFIX
const sub = nc.subscribe(`${prefix}.broadcast`);
for await (const msg of sub) {
  const syncMsg = decode(msg.data) as SyncMessage;

  // Resolve execution tracker (works for both ack and delta)
  if (syncMsg.executionId) {
    tracker.notify(syncMsg.executionId);
  }

  // Ack has no data — skip delta processing
  if (syncMsg.type === 'ack') continue;

  // Apply delta to reactive atoms
  if (syncMsg.type === 'delta' && syncMsg.changes) {
    if (syncMsg.changes.prs) {
      prsCtrl.update(prev => applyDelta(prev, syncMsg.changes.prs!));
    }
    if (syncMsg.changes.invoices) {
      invoicesCtrl.update(prev => applyDelta(prev, syncMsg.changes.invoices!));
    }
    if (syncMsg.changes.payments) {
      paymentsCtrl.update(prev => applyDelta(prev, syncMsg.changes.payments!));
    }
  }
}
```

## Golden Example: UI Waiting for Sync

The API layer wraps `executionTracker.wait()` into the action result. UI code awaits it before proceeding (e.g., closing a modal) so the user sees fresh data.

```typescript
// In a UI handler
const result = await actions.act('approvePr', { prId });
if (result.success) {
  // Wait for NATS to deliver the delta/ack (2s timeout auto-resolves)
  if (result.wait) await result.wait();
  closeModal();
}
```

`executionTracker.wait(id, timeout)` returns a promise that resolves when `tracker.notify(id)` is called (from the NATS subscription), or after `timeout` ms (default 2000) as a safety fallback.

## applyDelta Merge Logic

```typescript
function applyDelta<T extends { id: number }>(
  current: T[],
  changes: ChangeSet<T>
): T[] {
  let result = [...current];
  // 1. Remove deleted
  if (changes.delete.length > 0) {
    const deleteSet = new Set(changes.delete);
    result = result.filter(item => !deleteSet.has(item.id));
  }
  // 2. Replace updated (full record replacement, not field merge)
  if (changes.update.length > 0) {
    const updateMap = new Map(changes.update.map(item => [item.id, item]));
    result = result.map(item => updateMap.get(item.id) ?? item);
  }
  // 3. Append added
  if (changes.add.length > 0) {
    result = [...result, ...changes.add];
  }
  return result;
}
```

Order matters: delete first, then update, then add. Updates are full record replacement (not partial merge).

## Convention

| Rule | Why |
| --- | --- |
| Services call sync.emit() after DB write | Delta carries the actual data change |
| Flows call sync.ack(executionId) at the end | Signals "mutation complete" to originating client |
| Always guard ack: if (executionId) | executionId may not exist in non-interactive contexts |
| Keep executionId as string end-to-end | executionTracker correlation depends on exact key equality |
| Broadcast to all, filter on client | Simplifies server; RBAC filtering happens in UI atoms |
| Send full records in deltas, not partial fields | applyDelta does full replacement — partial would lose data |

## Anti-patterns

| Don't | Why |
| --- | --- |
| Call sync.ack() inside a service | Services emit data deltas; flows own the ack lifecycle |
| Emit a delta without data on add/update | sync.emit silently drops the message (logged as warning) |
| Await result.wait() without the guard | result.wait is only present when executionId was returned |
| Skip sync.ack() in a flow | Client's executionTracker falls back to timeout, causing sluggish UI |
| Return numeric executionId from HTTP while NATS sends string | wait() may resolve only via timeout due key mismatch |
| Send partial records in delta updates | applyDelta replaces the entire record — partial data = data loss |

## Cited By

- c3-104 (Invoice Screen)
- c3-105 (Payment Requests Screen)
- c3-205 (PR Flows)
- c3-206 (Invoice Flows)
- c3-207 (Payment Flows)
- c3-212 (Workbench Flows)
- c3-209 (NATS Credential Generator)
- c3-211 (Notification System)
