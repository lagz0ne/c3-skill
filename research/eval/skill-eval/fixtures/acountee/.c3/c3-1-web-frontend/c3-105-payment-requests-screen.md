---
id: c3-105
c3-version: 3
c3-seal: 34bd37ba4203cdbc59f1d2d828f453f16a5bd768cc9e4f1aa9db2316a90e2b93
title: PaymentRequestsScreen
type: component
category: feature
parent: c3-1
goal: PR management - create, approve, reject, complete, with dual-mode PR/approvals view
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

# PaymentRequestsScreen

## Goal

PR management - create, approve, reject, complete, with dual-mode PR/approvals view

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own PaymentRequestsScreen behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep PaymentRequestsScreen decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for PaymentRequestsScreen so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before PaymentRequestsScreen behavior is changed. | ref-audit-timeline |
| Inputs | Accept only the files, commands, data, or calls that belong to PaymentRequestsScreen ownership. | ref-audit-timeline |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-audit-timeline |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-audit-timeline |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks PaymentRequestsScreen to deliver its documented responsibility. | ref-audit-timeline |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-audit-timeline |
| Alternate paths | When a request falls outside PaymentRequestsScreen ownership, hand it to the parent or sibling component. | ref-audit-timeline |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-audit-timeline |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-audit-timeline | ref | Governs PaymentRequestsScreen behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| PaymentRequestsScreen input | IN | Callers must provide context that matches the component goal and parent fit. | c3-1 boundary | c3x lookup plus targeted tests or review. |
| PaymentRequestsScreen output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-1 boundary | c3x check and project test suite. |

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

Core workflow screen for payment request lifecycle. Users create PRs, route them through multi-step approval chains, and mark them complete after payment. Operates in two modes: **PR mode** (full CRUD for PR owners) and **Approvals mode** (focused approval queue for approvers with bulk approve).

## What Users Can Do

- **Create PR** -- Form with name, supplier, amount, type (normal/advanced), payment method, linked invoices, file attachments. Calls `createPr`.
- **Edit PR** -- Update draft PRs via `updatePr`. Manage file attachments via `updatePrAttachments`.
- **Approval workflow** -- `requestApprovals` submits for review. Approvers see pending items and can `approvePr`, `rejectPr`, or bulk `approveAll`. PR creator can `recallPr` to pull back. Approvers can `unapprovePr` to retract.
- **Complete / Uncomplete** -- After all approvals pass, `completePr` marks the PR done. `uncompletePr` reopens.
- **Delete draft** -- `removePr` deletes PRs still in draft state.
- **Export bank transfer** -- Generate export file for a single approved PR via ExportBankTransferDialog.
- **View detail** -- Tabbed pane (Details, Audit) showing PR metadata, approval chain status, linked invoices, attachments.
- **Auto-mark notifications** -- Selecting a PR auto-marks matching notifications as read.

## Dual Mode

| Mode | List Shows | Available Actions |
| --- | --- | --- |
| PRs | All user-visible PRs, grouped by status or flat list | Full CRUD, request approval, complete |
| Approvals | Only PRs pending current user's approval | Approve, reject, bulk approve |

Mode is passed as a prop (`mode: 'prs' | 'approvals'`).

## Data Flow

```
SSR loader -> prs atom, invoices atom, payments atom, user atom, approvalFlow atom

Client-side filtering via usePaymentRequestFilters hook.
Filters: status, date range, amount range, creator, sort, view mode (grouped/list), archived toggle.
Real-time updates via NATS sync.

Server functions (pr.ts):
  createPr, updatePr, approvePr, unapprovePr, requestApprovals,
  completePr, uncompletePr, rejectPr, recallPr, removePr,
  approveAll, updatePrAttachments, removePrAttachment
```

## Approval Chain Display

Shows step-by-step approval progress: pending steps with required approvers, completed steps with who approved and when, and whether the current user can act on the current step.

## Key Wiring

- **Atoms**: `prs`, `invoices`, `payments`, `user`, `approvalFlow` from `@/lib/pumped`
- **Server functions**: `@/server/functions/pr` (13 functions)
- **Hooks**: `usePaymentRequestFilters`, `usePaymentRequestSelection`, `useBulkOperations`, `useKeyboardShortcuts`, `usePRActions`, `useAutoMarkNotificationsRead` from `paymentRequestHooks.ts`
- **Components**: MasterDetailLayout, FilterFooter, PRFilterContent, AuditLogPanel, Drawer (create/edit PR), ExportBankTransferDialog
- **Shared**: `prHelpers` for status logic, `approvalFlowSchema` from `@acountee/shared`

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN (uses) | See component wiring and code sections in this document. | c3-1 |
| OUT (provides) | Exposes behavior described by this component to dependent callers. | c3-1 |
