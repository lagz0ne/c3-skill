---
id: c3-109
c3-version: 3
c3-seal: 066fa43b0c0b4e5bade29dbf6514ef11c9fe6059780713c54855396f18835ca0
title: Workbench Screen
type: component
category: feature
parent: c3-1
goal: Finance team operational tools - invoice cleanup, approved PR export, paid PR import
uses:
    - ref-bulk-operations
    - ref-list-view-patterns
    - ref-responsive-layout
    - ref-sft-behavioral-spec
    - ref-ui-patterns
    - ref-variant-system
---

# Workbench Screen

## Goal

Finance team operational tools - invoice cleanup, approved PR export, paid PR import

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own Workbench Screen behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep Workbench Screen decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for Workbench Screen so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before Workbench Screen behavior is changed. | ref-bulk-operations |
| Inputs | Accept only the files, commands, data, or calls that belong to Workbench Screen ownership. | ref-bulk-operations |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-bulk-operations |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-bulk-operations |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks Workbench Screen to deliver its documented responsibility. | ref-bulk-operations |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-bulk-operations |
| Alternate paths | When a request falls outside Workbench Screen ownership, hand it to the parent or sibling component. | ref-bulk-operations |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-bulk-operations |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-bulk-operations | ref | Governs Workbench Screen behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| Workbench Screen input | IN | Callers must provide context that matches the component goal and parent fit. | c3-1 boundary | c3x lookup plus targeted tests or review. |
| Workbench Screen output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-1 boundary | c3x check and project test suite. |

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

Month-to-month operational toolset for finance teams. Three tabs handle the end-of-period workflow: clean up stale invoices, export approved PRs for bank transfers, and import payment references to close the loop. Accessible to users whose team has the `workbench` capability (initially `finance` and `bod` teams).

## Tabs

### Invoice Cleanup

Select a month, see all `imported`-status invoices for that period. Bulk-select and mark as obsolete to clean up invoices that won't be processed.

- MonthPicker selects year-month period
- Checkbox table with invoice number, supplier, amount, date
- Select all / individual selection, then "Mark as Obsolete" bulk action
- Calls `workbenchListInvoicesForCleanup` (load) and `workbenchBulkMarkObsolete` (action)

### Export Approved PRs

Select a month, see all approved PRs with bank/payment info. Select which PRs to include and export a CSV for manual bank transfers.

- MonthPicker selects period, pre-selects all PRs on load
- Table: PR ID, name, supplier, amount, bank name, bank account, type
- Client-side CSV generation with browser download
- CSV columns: PR ID, PR Name, Supplier, Amount, Bank Name, Bank Account, Account Name, Type
- Calls `workbenchListApprovedPrsForExport` (load)

### Import Paid PRs

Upload a CSV of `pr_id,payment_reference` pairs to mark PRs as complete after bank payment.

- Drag-and-drop or file picker CSV upload
- Parses and validates rows (checks valid PR ID + non-empty reference)
- Preview table with per-row validation status
- Import button calls `workbenchImportPaidPrs`, shows per-row success/failure results
- Clear button resets to upload state

## Data Flow

```
Each tab manages its own local state (no shared atoms).
Data loaded on-demand per month selection.

Server functions (workbench.ts):
  workbenchListInvoicesForCleanup, workbenchBulkMarkObsolete,
  workbenchListApprovedPrsForExport, workbenchImportPaidPrs
```

## Key Wiring

- **Server functions**: `@/server/functions/workbench` (4 functions)
- **Auth**: `workbench` capability required via `hasCapability()`
- **Components**: Tabs (3-tab interface), MonthPicker (Select-based period picker), admin-page/admin-table layout, Alert (feedback)
- **Responsive**: Sticky header bar with MonthPicker. Mobile: short tab labels, hidden columns (date, bank, account, type), sticky bottom action bar. Desktop: full labels, all columns, static count.

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN (uses) | See component wiring and code sections in this document. | c3-1 |
| OUT (provides) | Exposes behavior described by this component to dependent callers. | c3-1 |
