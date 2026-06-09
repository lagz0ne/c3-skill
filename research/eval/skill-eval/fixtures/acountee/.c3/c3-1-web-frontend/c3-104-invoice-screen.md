---
id: c3-104
c3-version: 3
c3-seal: fc8185b7bfea3c1212a8cf325ea04720a14a77dc3ea11e82c132805f938e2a95
title: InvoiceScreen
type: component
category: feature
parent: c3-1
goal: Invoice management - upload, view, filter, link to PRs, bulk operations
uses:
    - ref-audit-timeline
    - ref-detail-content-strategy
    - ref-filter-footer
    - ref-form-patterns
    - ref-list-view-patterns
    - ref-master-detail-layout
    - ref-responsive-layout
    - ref-sft-behavioral-spec
    - ref-ui-patterns
    - ref-variant-system
---

# InvoiceScreen

## Goal

Invoice management - upload, view, filter, link to PRs, bulk operations

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own InvoiceScreen behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep InvoiceScreen decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for InvoiceScreen so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before InvoiceScreen behavior is changed. | ref-audit-timeline |
| Inputs | Accept only the files, commands, data, or calls that belong to InvoiceScreen ownership. | ref-audit-timeline |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-audit-timeline |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-audit-timeline |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks InvoiceScreen to deliver its documented responsibility. | ref-audit-timeline |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-audit-timeline |
| Alternate paths | When a request falls outside InvoiceScreen ownership, hand it to the parent or sibling component. | ref-audit-timeline |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-audit-timeline |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-audit-timeline | ref | Governs InvoiceScreen behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| InvoiceScreen input | IN | Callers must provide context that matches the component goal and parent fit. | c3-1 boundary | c3x lookup plus targeted tests or review. |
| InvoiceScreen output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-1 boundary | c3x check and project test suite. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |
| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| Code, docs, tests, prompts | Goal, Governance, Contract, and Change Safety sections. | Names and framework shape may vary; behavior and boundaries may not. | c3x check, c3x verify, and relevant tests. |

## Architecture Details

## Business Purpose

Central hub for managing supplier invoices. Finance users upload XML invoice files, review invoice details, link invoices to payment requests, and clean up stale records. Supports the full invoice lifecycle from import through to PR linkage.

## What Users Can Do

- **Import invoices** -- Upload XML files via drag-and-drop dialog (Ctrl+I shortcut). Parsed server-side via `importFiles`.
- **Browse and filter** -- Virtualized list with month-group headers. Filter by status, date range, supplier, search term. Toggle between active/archived/all views.
- **View invoice detail** -- Tabbed detail pane (General, Services with count badge, Audit). Shows supplier info, amount (VND), dates, payment details, linked PR, and line-item services.
- **Link/unlink to PR** -- Associate an invoice with a payment request via `linkPaymentRequest`, switch PR with `changePrToOther`, or detach with `unlinkPaymentRequest`.
- **Edit payment details** -- Update bank account info on the invoice via `updatePaymentDetail` dialog.
- **Mark obsolete / restore** -- Single or bulk operations. `markInvoiceAsRedundant` moves to archived; restore brings back.
- **Bulk operations** -- Toggle bulk mode (B shortcut), select multiple invoices via checkboxes, apply batch obsolete/restore.

## Data Flow

```
SSR loader -> invoices atom (pumped)
           -> prs atom (for PR linking)
           -> payments atom (for payment detail options)

Client-side filtering of prefetched atom data.
Real-time updates via NATS sync.

Server functions (invoice.ts):
  importFiles, linkPaymentRequest, unlinkPaymentRequest,
  changePrToOther, updatePaymentDetail,
  markInvoiceAsRedundant, markInvoiceAsImported
```

## Layout

Master-detail with FilterFooter. Virtualized list (sticky month-group headers) on left, tabbed detail pane on right. Keyboard shortcuts: Ctrl+I (import), B (bulk mode), Escape, Enter.

## Key Wiring

- **Atoms**: `invoices`, `prs`, `payments` from `@/lib/pumped`
- **Server functions**: `@/server/functions/invoice` (all 7 functions)
- **Components**: MasterDetailLayout, FilterFooter, InvoiceFilterContent, AuditLogPanel, Drawer (import + payment edit), Combobox (PR picker)
- **Helpers**: `InvoiceService` for parsing services JSON, `prHelpers` for PR display

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN (uses) | See component wiring and code sections in this document. | c3-1 |
| OUT (provides) | Exposes behavior described by this component to dependent callers. | c3-1 |
