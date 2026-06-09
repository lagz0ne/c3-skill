---
id: adr-20260212-workbench-feature
c3-seal: 507a99970db93591c88a4d4ef9fd1e01c07d71d76c1cc023f54de5505de2264d
title: Workbench - Finance Team Operational Toolset
type: adr
goal: Provide a dedicated finance operations surface for month-based cleanup, export, and import workflows that are currently manual or scattered.
status: implemented
date: "2026-02-12"
affects:
    - c3-1
    - c3-2
    - c3-204
    - c3-205
    - c3-206
approved-files:
    - .c3/c3-1-web-frontend/c3-109-workbench-screen.md
    - .c3/c3-2-api-backend/c3-212-workbench-flows.md
    - apps/start/src/server/dbs/schema.ts
    - apps/start/src/server/dbs/migrations/0009_workbench_payment_reference.sql
    - apps/start/src/server/flows/workbench.ts
    - apps/start/src/server/functions/workbench.ts
    - apps/start/src/routes/_authed/workbench.tsx
    - apps/start/src/screens/WorkbenchScreen.tsx
    - apps/start/src/components/AppSidebar.tsx
    - apps/start/src/lib/capabilities.ts
    - apps/start/src/server/dbs/migrations/0010_workbench_capability_seed.sql
new-components:
    - c3-109: Workbench Screen (frontend)
    - c3-212: Workbench Flows (backend)
---

# Workbench - Finance Team Operational Toolset

## Goal

Provide a dedicated finance operations surface for month-based cleanup, export, and import workflows that are currently manual or scattered.

## Problem

The finance team performs repetitive month-to-month operational tasks that are currently either manual or scattered across different screens:

1. **Invoice cleanup** -- Invoices arrive without filtering; some are wrongly released by partners or handled outside the system. These pile up and clutter the invoice list. Finance needs a focused tool to review and bulk-mark them as obsolete by month.
**Invoice cleanup** -- Invoices arrive without filtering; some are wrongly released by partners or handled outside the system. These pile up and clutter the invoice list. Finance needs a focused tool to review and bulk-mark them as obsolete by month.
2. **Exporting approved PRs** -- Once PRs are approved, finance must manually extract details (PR ID, name, supplier, amount, bank info) to perform bank transfers in a separate banking app. No batch export exists.
**Exporting approved PRs** -- Once PRs are approved, finance must manually extract details (PR ID, name, supplier, amount, bank info) to perform bank transfers in a separate banking app. No batch export exists.
3. **Importing paid PRs** -- After bank transfers are complete, PRs need to be marked as finished. Currently this is done one-by-one. Finance needs to import a CSV of PR IDs with payment references to bulk-complete them.
**Importing paid PRs** -- After bank transfers are complete, PRs need to be marked as finished. Currently this is done one-by-one. Finance needs to import a CSV of PR IDs with payment references to bulk-complete them.

## Decision

### 1. New top-level Workbench section

Add a **Workbench** nav item in the main sidebar, between Payment Methods and Admin. Visibility controlled by a new team capability `workbench` (initially assigned to `finance` and `bod` teams).

### 2. Tabbed interface with three tools

The Workbench screen uses a **tabbed layout** with three tabs:

#### Tab 1: Invoice Cleanup

- **Month selector** (year-month picker) to focus on a specific period
- **List of invoices** for that month in `imported` status (never linked to a PR, sitting idle)
- **Bulk select** with checkboxes + select-all
- **"Mark as Obsolete" button** -- calls existing `markInvoiceAsRedundant` flow in bulk
- Shows count of cleaned vs remaining invoices for the month

#### Tab 2: Export Approved PRs

- **Month selector** to pick a billing period
- **List of approved PRs** for that month (status = `approved`)
- **Select PRs** to export (default: all)
- **"Export CSV" button** -- generates and downloads a CSV file with columns:
PR ID, PR Name, Supplier, Amount, Bank Name, Bank Account, Account Name, Type (direct/advanced)
- Uses existing PR query data; no new backend query needed beyond a filtered list

#### Tab 3: Import Paid PRs

- **CSV upload area** (drag & drop or file picker)
- **CSV format**: `pr_id,payment_reference` (two columns, header row)
- **Preview table** showing parsed rows with validation status
- **"Import" button** -- for each valid row:
Find PR by ID, verify status = `approved`

Set `payment_reference` on the PR record

Call `completePr` flow to transition status to `complete`

- **Results summary** showing success/failure per row

### 3. Schema change: payment_reference column on pr table

Add a nullable `text` column `payment_reference` to the `pr` table. This stores the bank payment trace/reference number after import.

```sql
ALTER TABLE "pr" ADD COLUMN "payment_reference" text;
```

Update the Drizzle schema to include `paymentReference: text("payment_reference")`.

### 4. New capability: workbench

Add to `CAPABILITIES` constant:

```typescript
WORKBENCH: 'workbench'
```

Seed migration assigns capability to `finance` and `bod` teams.

### 5. Backend flows (c3-212)

New flow file `workbench.ts` with:

- **`bulkMarkObsolete`** -- accepts `{ invoiceIds: number[] }`, calls `markInvoiceAsRedundant` for each, returns per-item results
- **`listInvoicesForCleanup`** -- accepts `{ year: number, month: number }`, returns invoices in `imported` status for that month
- **`listApprovedPrsForExport`** -- accepts `{ year: number, month: number }`, returns approved PRs with payment details for that month
- **`importPaidPrs`** -- accepts `{ items: { prId: number, paymentReference: string }[] }`, validates, sets reference, completes each PR, returns per-item results

### 6. Server functions

New file `functions/workbench.ts` wrapping flows with `transactionMiddleware`.

## Rationale

- **Separate screen** rather than adding to existing Invoice/PR screens because workbench operations are cross-cutting (spans invoices AND PRs) and month-focused rather than item-focused.
- **Capability-based access** reuses the existing team capabilities system rather than inventing a new RBAC mechanism.
- **Bulk operations with per-item results** matches the existing `importFiles` pattern (returns success/failure per item).
- **CSV for import** is the simplest format that finance teams already work with; no need for Excel parsing.
- **`payment_reference` on PR table** directly (not a separate table) because it's a 1:1 relationship -- each PR has at most one payment reference.

## Impact

| Component | Change |
| --- | --- |
| c3-204 (Drizzle) | New column payment_reference on pr table |
| c3-205 (PR Flows) | completePr may receive payment_reference; existing flow signature unchanged |
| c3-206 (Invoice Flows) | Reuses markInvoiceAsRedundant; no changes needed |
| c3-1 (Frontend) | New route, screen, sidebar item |
| Capabilities | New workbench capability |
| Migrations | Two new: schema change + capability seed |

## Context

N.A - historical ADR; original context is captured in the git log around the ADR date and in the current code that implements the decision.

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| N.A - historical | Shipped via git commits; the c3 topology and code-map reflect the resulting structure. | c3x list --include-adr and git log around the ADR date |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| N.A - historical | Current .c3 entities, refs, and code-map are the post-change state. | c3x verify and c3x check |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| N.A - historical | Enforcement is implicit in the currently linked components and refs. | c3x graph and cited ref ids on the relevant components |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| N.A - historical | Alternatives were considered at decision time; rationale is preserved in the original commit message or branch discussion. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| N.A - historical | Risks were assessed pre-merge; the decision has since shipped without outstanding incidents tied to this ADR. | git log and project test suite |

## Verification

| Check | Result |
| --- | --- |
| Merged and running in production | PASS - see git log for the merge commit |
