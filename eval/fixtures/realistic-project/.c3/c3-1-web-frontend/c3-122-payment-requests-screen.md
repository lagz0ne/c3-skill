---
id: c3-122
c3-version: 3
title: Payment Requests Screen
type: component
category: feature
parent: c3-1
summary: Main screen for creating, viewing, editing, and managing payment request workflows
---

# Payment Requests Screen

Provides the primary interface for payment request management including list view with filtering, detail panel with approval status, PR creation/editing forms, and workflow actions (request approval, approve, complete).

## Uses

| Category | Component | For |
|----------|-----------|-----|
| Foundation | c3-103 State Atoms | prs, invoices, approvalFlow state |
| Foundation | c3-104 UI Variants | Modal, form, table styling |
| Ref | ref-form-patterns | PR creation/edit forms |
| Ref | ref-data-sync | Real-time PR updates |
| Ref | ref-error-handling | Action error display |

## Behavior

| Trigger | Result |
|---------|--------|
| Click Create PR button | Opens creation drawer with form |
| Select PR from list | Shows detail panel with status, invoices, actions |
| Submit PR form | Creates PR via server function, syncs to list |
| Click Request Approval | Transitions PR to 'requested' status |
| Click Approve (as approver) | Records approval, advances step if conditions met |
| Click Complete | Marks PR as 'complete' |
| Apply filter | List shows matching PRs only |

## Testing

| Scenario | Verifies |
|----------|----------|
| PR creation | Form submits, PR appears in list |
| Invoice linking | Selected invoices attach to PR |
| Approval flow display | Correct steps and approvers shown |
| Status transition | Draft -> Requested -> Approved -> Complete |
| Filter by status | Only matching PRs visible |
| Real-time update | Other user's change appears without refresh |

## References

- `apps/start/src/screens/PaymentRequests*.tsx` - Screen and hooks
- `apps/start/src/routes/_authed/prs.tsx` - Route
