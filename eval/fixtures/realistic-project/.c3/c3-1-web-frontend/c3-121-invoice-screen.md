---
id: c3-121
c3-version: 3
title: Invoice Screen
type: component
category: feature
parent: c3-1
summary: Invoice listing with filtering, detail view, and status tracking
---

# Invoice Screen

Provides the interface for viewing and filtering invoices imported into the system. Shows invoice list with search/filter, detail drawer with line items, and status indicators.

## Uses

| Category | Component | For |
|----------|-----------|-----|
| Foundation | c3-103 State Atoms | invoices state |
| Foundation | c3-104 UI Variants | Table, drawer, filter styling |
| Ref | ref-data-sync | Real-time invoice updates |
| Ref | ref-error-handling | Error display |

## Behavior

| Trigger | Result |
|---------|--------|
| Page load | Fetches and displays invoice list |
| Select invoice | Opens detail drawer with line items |
| Apply filter | List shows matching invoices |
| Search by number/supplier | Filters list in real-time |
| Status change (from sync) | UI updates automatically |

## Testing

| Scenario | Verifies |
|----------|----------|
| List rendering | All invoices displayed with key fields |
| Detail view | Line items shown correctly |
| Filter by status | Only matching status shown |
| Search | Partial match on invoice number/supplier |
| Real-time update | Invoice change from server appears |

## References

- `apps/start/src/screens/InvoiceScreen.tsx` - Screen component
- `apps/start/src/routes/_authed/invoices.tsx` - Route
