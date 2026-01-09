---
id: ref-realtime-sync
title: Real-time Sync
---

# Real-time Sync

## Goal

Establish conventions for server-side WebSocket sync broadcasting data changes to connected clients with delta updates and execution acknowledgments.

## Conventions

| Rule | Why |
|------|-----|
| Use sync.emit() after mutations | Notify connected clients of changes |
| Include executionId when available | Client can filter own actions |
| Structure as ChangeEvent with entity/type/id/data | Consistent message format |
| Support delta messages for efficiency | Only changed data transmitted |
| Support full sync for reconnection | Client can recover state |
| Send ack messages for execution confirmation | Client knows server processed action |
| Filter PRs by user on sync | Users only see PRs they can access |

## Testing

| Convention | How to Test |
|------------|-------------|
| Delta emission | Mutate, verify delta message sent |
| ExecutionId passthrough | Include ID, verify in outgoing message |
| Full sync content | Request full sync, verify all entities included |
| Client filtering | Connect as user, verify only accessible PRs |
| Connection cleanup | Disconnect client, verify no memory leak |

## References

- `apps/start/src/server/resources/sync.ts` - Sync types
- `apps/start/src/server/resources/wsServer.ts` - WebSocket server
