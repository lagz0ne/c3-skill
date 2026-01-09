---
id: c3-103
c3-version: 3
title: State Atoms
type: component
category: foundation
parent: c3-1
summary: @pumped-fn/lite atoms for reactive client-side state management
---

# State Atoms

Provides reactive state containers for user, invoices, payment requests, payments, and approval flow using @pumped-fn/lite atoms with controller-based updates.

## Contract

| Provides | Expects |
|----------|---------|
| user atom | User object with permissions |
| invoices atom | ListInvoiceRow[] array |
| prs atom | ListPrRow[] array |
| payments atom | ListPaymentsRow[] array |
| approvalFlow atom | Record<string, ApprovalFlow> |
| controller.update() | Function to modify state |

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Empty initial state | Empty arrays, null user |
| Concurrent updates | Last write wins |
| SSR hydration | preset() initializes from loader |
| Scope dispose | Atoms cleaned up |

## Testing

| Scenario | Verifies |
|----------|----------|
| Initial preset | useAtom returns SSR data |
| Controller update | Modify state, component re-renders |
| Delta merge | applyDelta correctly adds/updates/deletes |
| Type safety | Wrong shape rejected at compile time |

## References

- `apps/start/src/lib/pumped/` - All atom definitions and utilities
