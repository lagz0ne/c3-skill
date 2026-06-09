---
id: c3-1
c3-version: 3
c3-seal: 0acf5a21257520fdde0b2f20c61526bb31c35cecec0b25ccc3de40688b351893
title: Web Frontend
type: container
parent: c3-0
goal: React SPA for approval workflows—Invoices, PRs, Payments, and Finance Workbench. TanStack Router with pumped-fn atoms for state, real-time NATS sync.
---

# Web Frontend

## Goal

React SPA for approval workflows—Invoices, PRs, Payments, and Finance Workbench. TanStack Router with pumped-fn atoms for state, real-time NATS sync.

## Responsibilities

- Render approval workflow UIs for finance, admin, and BOD users.
- Maintain client state consistency through HTTP actions and NATS-driven sync updates.
- Enforce capability-aware UX boundaries for privileged/admin features.

## Components

| ID | Name | Category | Status | Goal Contribution |
| --- | --- | --- | --- | --- |
| c3-101 | State Management | foundation | active | Shared client state and execution tracking for all screens |
| c3-102 | Form System | foundation | active | Validated, reusable form interactions |
| c3-103 | UI Kit (shadcn/radix) | foundation | active | Consistent design primitives for application screens |
| c3-104 | InvoiceScreen | feature | active | Invoice lifecycle operations |
| c3-105 | PaymentRequestsScreen | feature | active | PR workflow creation and approvals |
| c3-106 | PaymentsScreen | feature | active | Payment method and payment flow management |
| c3-107 | Admin Screens | feature | active | Admin-only management and configuration |
| c3-108 | DevTools (DevShell) | foundation | active | Developer diagnostics and runtime tooling |
| c3-109 | Workbench Screen | feature | active | Finance operations tooling (cleanup/export/import) |

## Wiring

| To | Protocol | What |
| --- | --- | --- |
| c3-2 (API Backend) | HTTP | /act endpoint for mutations, data fetch |
| c3-4 (NATS) | WSS | Real-time sync via natsSync atom |

## Foundation

- **State**: pumped-fn/lite atoms (`api`, `user`, stores for prs/invoices/payments, `natsSync`)
- **Forms**: Zod validation, Drawer forms, FormComponents
- **UI Kit**: shadcn/ui (Radix), Tailwind 4 + daisyUI
- **Real-time**: nats.ws for NATS subscriptions

## Screens

| Screen | Responsibility |
| --- | --- |
| InvoiceScreen | Invoice management with approval workflow, filter/list/detail/actions |
| PaymentRequestsScreen | PR management with approval chain |
| PaymentsScreen | Payment tracking table |
| WorkbenchScreen | Finance ops: invoice cleanup, PR export, paid PR import |
| Admin Screens | Users, teams, roles, audit logs, approval config (owner role required) |

## Tech Stack

| Layer | Technology |
| --- | --- |
| Framework | TanStack Start (SSR + file-based routing) |
| UI | React 19 |
| State | @pumped-fn/lite (atoms, DI) |
| Styling | Tailwind 4 + daisyUI |
| Components | shadcn/ui (Radix) |
| Real-time | nats.ws |
