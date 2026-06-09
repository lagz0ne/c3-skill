---
id: recipe-approval-workflow
c3-seal: fb52a2244292ff5e3e8e81bc15f31ac42033db1a5539a2f96ce5095e50ce4e24
title: Approval Workflow
type: recipe
goal: Trace the end-to-end approval workflow from PR creation through multi-step approval to completion, including workbench bulk operations.
sources:
    - c3-0
    - c3-205
    - c3-212
    - ref-approval-chain
    - ref-audit-trail
    - ref-sync
---

# Approval Workflow

## Goal

Trace the end-to-end approval workflow from PR creation through multi-step approval to completion, including workbench bulk operations.

## Narrative

The approval workflow is the core business domain. Payment requests move
through `draft → pending → approved → completed`, governed by multi-step
approval chains (ref-approval-chain).

Each step has assigned approvers and a mode: `anyof` (one suffices) or
`allof` (every assigned user must approve). Mode validation is app-level
logic in `prService`, not DB constraints. Step advancement triggers
notification to next-step approvers via the notification system (c3-211).

PR flows (c3-205) own 15 operations covering the full lifecycle. The
workbench (c3-212) extends this with bulk operations: cleanup imported
invoices, export approved PRs with resolved bank info, import paid PRs
with payment references.

Two PR types exist: **direct** (bank info from linked invoices) and
**advanced** (bank info stored on PR itself). Workbench export resolves
this transparently.

## Cross-Cutting Contracts

- Every mutation emits a sync delta (ref-sync), then the flow acks
- Approval mutations are audit-captured via DB trigger on `pr` table —
do NOT also call `createAuditEntry` (ref-audit-trail)
- All operations run in transaction scope (c3-202 execution context)
- Notifications are fire-and-forget (error suppressed, logged)

## Risk

Approval chain logic spans c3-205 (flows) + prService + approvalQueries.
Changing step semantics has blast radius across all three layers plus the
notification dispatch path.
