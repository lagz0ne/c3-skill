---
id: c3-106
c3-version: 3
c3-seal: e8cdf9e13120d16b89e579667ac4eec012d9fb0a7f18d9cf5cf9c80aa4c22051
title: PaymentsScreen
type: component
category: feature
parent: c3-1
goal: Payment method management - CRUD for bank accounts used in PRs and invoices
uses:
    - ref-form-patterns
    - ref-list-view-patterns
    - ref-responsive-layout
    - ref-sft-behavioral-spec
    - ref-ui-patterns
    - ref-variant-system
---

# PaymentsScreen

## Goal

Payment method management - CRUD for bank accounts used in PRs and invoices

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own PaymentsScreen behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep PaymentsScreen decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for PaymentsScreen so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before PaymentsScreen behavior is changed. | ref-form-patterns |
| Inputs | Accept only the files, commands, data, or calls that belong to PaymentsScreen ownership. | ref-form-patterns |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-form-patterns |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-form-patterns |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks PaymentsScreen to deliver its documented responsibility. | ref-form-patterns |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-form-patterns |
| Alternate paths | When a request falls outside PaymentsScreen ownership, hand it to the parent or sibling component. | ref-form-patterns |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-form-patterns |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-form-patterns | ref | Governs PaymentsScreen behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| PaymentsScreen input | IN | Callers must provide context that matches the component goal and parent fit. | c3-1 boundary | c3x lookup plus targeted tests or review. |
| PaymentsScreen output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-1 boundary | c3x check and project test suite. |

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

Manages the organization's bank account / payment method registry. These payment methods are referenced when creating payment requests and editing invoice payment details. Simple reference data CRUD -- not transaction records.

## What Users Can Do

- **Add payment method** -- Opens PaymentForm drawer. Fields: name, bank name, bank account (monospace), account holder name, notes.
- **Edit** -- Click edit icon on any row to modify via same PaymentForm.
- **Delete** -- Click delete icon, confirm via ConfirmDrawer, calls `deletePayment`.
- **Search** -- Client-side filter across name, bank, account, and notes.

## Data Flow

```
SSR loader -> payments atom (pumped)

Client-side search filtering of prefetched atom data.

Server functions (payment.ts):
  createPayment, updatePayment, deletePayment
```

## Layout

Admin-page table layout (not master-detail). Header with title + "Add Payment" button + search. Plain HTML table with columns: Name, Bank, Account (font-mono), Note (truncated), Actions. Footer shows count. Empty state with CTA when no payment methods exist.

## Key Wiring

- **Atom**: `payments` from `@/lib/pumped`
- **Server functions**: `@/server/functions/payment` (3 functions)
- **Components**: PaymentForm (shared drawer for create/edit), ConfirmDrawer (delete confirmation), Alert (success/error feedback)

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN (uses) | See component wiring and code sections in this document. | c3-1 |
| OUT (provides) | Exposes behavior described by this component to dependent callers. | c3-1 |
