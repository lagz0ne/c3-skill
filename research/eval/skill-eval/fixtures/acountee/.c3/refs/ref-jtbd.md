---
id: ref-jtbd
c3-seal: 327816b3f7dc41aed8ddcb789782ec3ab60ce5eb9d2e2edc593b57a03a4fee9a
title: Jobs-To-Be-Done (JTBD) Map
type: ref
goal: Map every user job in Acountee to its implementation (C3 component, SFT screen, flow) — the single source of truth for what users hire this product to do and where each job lives.
---

# Jobs-To-Be-Done Map

## Goal

Map every user job in Acountee to its implementation (C3 component, SFT screen, flow) — the single source of truth for what users hire this product to do and where each job lives.

## Choice

JTBD framework with actor-job-implementation traceability. Each job maps to: actor, C3 component (backend/frontend), SFT screen/region, and SFT flow. Gaps are surfaced explicitly.

## Why

- Makes product intent visible — every feature exists to serve a job
- Surfaces spec gaps — jobs without SFT coverage are behavioral blind spots
- Guides prioritization — jobs cluster around actors, revealing who the product serves most
- Bridges C3 (architecture) and SFT (behavior) — the missing "why" layer

## How

### Actors

| Actor | Role | Primary Screens |
| --- | --- | --- |
| Finance Operator | Day-to-day: import invoices, create PRs, manage payments, period-end ops | Invoices, PaymentRequests, Payments, Workbench |
| Approver | Reviews and approves/rejects PRs in multi-step chains | PaymentRequests (Approvals mode), NotificationBell |
| Owner/Admin | Manages users, teams, roles, approval configs, monitors system | Admin* screens |

### State Machines

Two entity lifecycle state machines govern the core domain:

**PR Lifecycle** (`PRDetail` region — the crown jewel):

```
draft → pending → approved → completed
  ↑       ↑↓          ↑↓
  │    unapprove    uncomplete
  │    (self-loop)
  ├── reject (back to draft)
  ├── recall (back to draft)
  └── delete (terminal)
```

Diagram: https://diashort.apps.quickable.co/d/608160b3

**Invoice Lifecycle** (`InvoiceDetail` region):

```
imported ⇄ archived  (mark_obsolete / restore)
```

Diagram: https://diashort.apps.quickable.co/d/89018b6d

**App Shell** (AuthGuard, ErrorBoundary, Sidebar, Theme):

- AuthGuard: checking → authenticated | unauthenticated (→ Login redirect)
- AdminGuard: verifying → owner | denied (→ Invoices redirect)
- ErrorBoundary: healthy → error → retry (max 3) | reload | continue
- Sidebar: expanded ⇄ collapsed
- Theme: light ⇄ dark

Diagram: https://diashort.apps.quickable.co/d/9cda2ac2

### Diagrams

| Diagram | URL |
| --- | --- |
| PR Lifecycle State Machine | https://diashort.apps.quickable.co/d/608160b3 |
| Invoice Lifecycle State Machine | https://diashort.apps.quickable.co/d/89018b6d |
| App Shell States | https://diashort.apps.quickable.co/d/9cda2ac2 |
| Navigation Graph | https://diashort.apps.quickable.co/d/a31cd3b6 |
| All 27 Flows (grouped) | https://diashort.apps.quickable.co/d/3f24cbd1 |
| Main Screen Interactions | https://diashort.apps.quickable.co/d/576315d4 |
| Admin Screen Interactions | https://diashort.apps.quickable.co/d/12e46e7a |

### Access & Identity

| ID | Job Statement | Actor | C3 | SFT | SFT Flow |
| --- | --- | --- | --- | --- | --- |
| J1 | When I need to access Acountee, I want to log in with Google OAuth, so I can start working | All | c3-213 | @s11 Login, @r32 LoginForm | @f22 LoginFlow |
| J2 | When I'm done working, I want to log out securely | All | c3-213 | @r35 AppSidebar | @f26 LogoutFlow |
| J3 | When my session expires, the auth guard redirects me to login | All | c3-213 | @r42 AuthGuard | @f28 AuthGuardFlow |

### Invoice Management

| ID | Job Statement | Actor | C3 | SFT | SFT Flow |
| --- | --- | --- | --- | --- | --- |
| J4 | When supplier invoices arrive as XML, I want to bulk-import them | Finance | c3-206 | @s1, @r4 ImportDialog | @f3 InvoiceToPRFlow |
| J5 | When invoices are imported, I want to filter/search/browse them | Finance | c3-104 | @s1, @r1 InvoiceList, @r3 | — |
| J6 | When an invoice needs payment, I want to link it to a PR | Finance | c3-206 | @s1, @r2 InvoiceDetail | @f3, @f14 |
| J7 | When invoices become irrelevant, I want to mark obsolete or restore | Finance | c3-206 | @s1, @r2 | @f15 InvoiceCleanupFlow |
| J8 | When reviewing a PR, I want to preview the PDF inline | Finance | c3-105 | @s2, @r34 PDFPreviewDialog | — |

### Payment Request Lifecycle

| ID | Job Statement | Actor | C3 | SFT | SFT Flow |
| --- | --- | --- | --- | --- | --- |
| J9 | When I need to initiate a payment, I want to create a PR | Finance | c3-205 | @s2, @r8 CreatePRDrawer | @f1 PRApprovalFlow |
| J10 | When a draft PR needs changes, I want to edit it before submission | Finance | c3-205 | @s2, @r9 EditPRDrawer | @f7 PREditFlow |
| J11 | When a PR needs evidence, I want to attach/manage files | Finance | c3-205 | @s2, @r6 PRDetail | @f8 PRAttachmentFlow |
| J12 | When my PR is ready, I want to submit for approval | Finance | c3-205 | @s2, @r6 | @f1 |
| J13 | When I'm an assigned approver, I want to review and approve | Approver | c3-205 | @s2, @r6, @r10 ApprovalChainDisplay | @f1 |
| J14 | When a PR doesn't meet requirements, I want to reject it | Approver | c3-205 | @s2, @r6, @r37 ConfirmRejectDrawer | @f2 PRRejectionFlow |
| J15 | When I submitted a PR prematurely, I want to recall it | Finance | c3-205 | @s2, @r6, @r36 ConfirmRecallDrawer | @f6 PRRecallFlow |
| J16 | When I approved in error, I want to retract my approval | Approver | c3-205 | @s2, @r6, @r38 ConfirmUnapproveDrawer | @f23 UnapproveFlow |
| J17 | When a PR is fully approved, I want to mark it complete | Finance | c3-205 | @s2, @r6 | @f1 |
| J18 | When I have many pending approvals, I want to bulk-approve them | Approver | c3-205 | @s2, @r5 PRList | @f5 BulkApproveFlow |
| J19 | When an approved PR needs a bank transfer file, I want to export it | Finance | c3-205 | @s2, @r11 ExportBankTransferDialog | @f19 BankTransferExportFlow |
| J20 | When a draft PR is no longer needed, I want to delete it | Finance | c3-205 | @s2, @r6, @r39 ConfirmDeletePRDrawer | @f25 DeleteDraftFlow |
| J21 | When a completed PR needs correction, I want to reopen it | Finance | c3-205 | @s2, @r6 | @f24 UncompleteFlow |

### Payment Method Management

| ID | Job Statement | Actor | C3 | SFT | SFT Flow |
| --- | --- | --- | --- | --- | --- |
| J22 | When we onboard suppliers or update bank details, I want to CRUD payment methods | Finance | c3-207 | @s3, @r13, @r15 | @f13 PaymentCRUDFlow |

### Period-End Operations

| ID | Job Statement | Actor | C3 | SFT | SFT Flow |
| --- | --- | --- | --- | --- | --- |
| J23 | When a period closes, I want to bulk-mark stale invoices obsolete by month | Finance | c3-212 | @s4, @r16 InvoiceCleanupTab | @f4 EndOfPeriodFlow |
| J24 | When approved PRs need processing, I want to export them as CSV by month | Finance | c3-212 | @s4, @r17 ExportApprovedTab | @f4 |
| J25 | When payments are processed externally, I want to import payment refs via CSV | Finance | c3-212 | @s4, @r18 ImportPaidTab | @f4 |

### Administration

| ID | Job Statement | Actor | C3 | SFT | SFT Flow |
| --- | --- | --- | --- | --- | --- |
| J26 | When staff changes, I want to create/edit/deactivate users | Owner | c3-210 | @s5, @s7, @r40 ConfirmDeleteUserModal | @f9, @f20 |
| J27 | When responsibilities change, I want to assign/revoke roles | Owner | c3-210 | @s5, @r20, @r46 RoleAssignmentDialog | @f21 RoleAssignmentFlow |
| J28 | When org structure changes, I want to manage teams | Owner | c3-210 | @s6, @s7, @r41 ConfirmDeleteTeamModal | @f10 TeamManagementFlow |
| J29 | When approval policies change, I want to configure workflows | Owner | c3-210 | @s8, @r44 FlowFormDrawer | @f12, @f27 ApprovalStepEditFlow |
| J30 | When compliance checks are needed, I want to filter/inspect the audit trail | Owner | c3-208 | @s9, @r29 AuditTable | @f16 AuditTrailFlow |
| J31 | When notification delivery fails, I want to view history and retry dispatches | Owner | c3-211 | @s10, @r30 NotificationTable | @f17 NotificationTriageFlow |

### Cross-Cutting

| ID | Job Statement | Actor | C3 | SFT | SFT Flow |
| --- | --- | --- | --- | --- | --- |
| J32 | When my team works simultaneously, I want real-time sync | All | c3-101 (natsSync) | — (infra) | — |
| J33 | When approval actions happen, I want in-app notifications | Approver | c3-214 | @r31 NotificationBell | @f18 NotificationApprovalFlow |
| J34 | When I click a notification, I want to navigate to the PR | Approver | c3-214 | @r31 | @f18 |
| J35 | When I'm in Slack, I want to receive and interact with approval notifications | Approver | c3-215 | N/A (external) | — |
| J36 | When I navigate the app, I want a sidebar with permission-based sections | All | c3-103 | @r35 AppSidebar | — |
| J37 | When an error occurs, I want to recover (retry/reload/continue) | All | c3-103 | @r47 ErrorBoundary | — |

### Coverage

36/37 jobs have full C3+SFT traceability. J32 (real-time sync) is infrastructure. J35 (Slack) is external UI.

### SFT Stats (as of 2026-03-23)

| Metric | Count |
| --- | --- |
| Screens | 11 |
| Regions | 47 |
| Events | 124 |
| Flows | 27 |
| Tags | 24 |
| Validation errors | 0 |
