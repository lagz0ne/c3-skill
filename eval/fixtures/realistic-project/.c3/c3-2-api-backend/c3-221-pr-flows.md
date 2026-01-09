---
id: c3-221
c3-version: 3
title: PR Flows
type: component
category: feature
parent: c3-2
summary: Payment request business flows for CRUD operations, approvals, and status management
---

# PR Flows

Provides the core business logic for payment request operations including creation with approval flow setup, updates, approval/unapproval, request for approvals, and completion.

## Uses

| Category | Component | For |
|----------|-----------|-----|
| Ref | ref-flow-patterns | Flow structure with deps/parse/factory |
| Ref | ref-query-patterns | prQueries, approvalQueries |
| Ref | ref-realtime-sync | Emit changes after mutations |
| Foundation | c3-205 Middleware | currentUserTag access |

## Behavior

| Trigger | Result |
|---------|--------|
| newPaymentRequest | Creates PR + approval record + steps + step users |
| updatePr | Updates PR fields, emits change |
| approvePr | Records approval, advances step if mode satisfied |
| unapprovePr | Removes approval record |
| requestApprovals | Sets PR status to 'requested' |
| completePr | Sets PR status to 'complete' |
| listPrs | Returns PRs filtered by user permissions |

## Testing

| Scenario | Verifies |
|----------|----------|
| Create with approval | PR and approval records created in transaction |
| Approval anyof mode | Single approval advances step |
| Approval allof mode | All users must approve to advance |
| Unauthorized approve | Returns failure, not throws |
| Sync emission | Change event sent after mutation |
| User filtering | Users only see accessible PRs |

## References

- `apps/start/src/server/flows/pr.ts` - Flow definitions
- `apps/start/src/server/services/pr.ts` - Business logic
- `apps/start/src/server/dbs/queries/pr.ts` - Queries
