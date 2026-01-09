---
id: ref-data-sync
title: Data Sync
---

# Data Sync

## Goal

Establish conventions for real-time data synchronization via WebSocket with delta updates and optimistic UI patterns.

## Conventions

| Rule | Why |
|------|-----|
| Connect WebSocket on authenticated user only | No sync for unauthenticated sessions |
| Process delta messages via applyDelta helper | Consistent merge of add/update/delete changesets |
| Track executionId for optimistic updates | Prevents duplicate UI updates when own action returns |
| Handle ack messages separately | Server confirms receipt, client can clear pending state |
| Cleanup connection on scope dispose | Prevents memory leaks and zombie connections |
| Use controller.update for atomic state changes | Ensures React re-renders correctly |

## Testing

| Convention | How to Test |
|------------|-------------|
| Delta processing | Send mock delta message, verify state updated |
| ExecutionId filtering | Send message with own executionId, verify no duplicate |
| Connection cleanup | Dispose scope, verify WebSocket closed |
| Reconnection | Simulate disconnect, verify auto-reconnect |
| Error handling | Send malformed message, verify logged not crashed |

## References

- `apps/start/src/lib/pumped/atoms/sync.ts` - Client sync atom
- `apps/start/src/server/resources/sync.ts` - Server sync types
