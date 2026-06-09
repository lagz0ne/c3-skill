---
id: adr-20260212-unify-workbench-ui-patterns
c3-seal: 50e0546636bc57cbb8cbe77b4994acc1b1594c2f32b64f88ddb23eb1c01fc12f
title: Unify UI Patterns Between Workbench and Admin Views
type: adr
goal: Align Workbench UI behavior with established admin and responsive layout refs while documenting any newly introduced reusable patterns.
status: implemented
date: "2026-02-12"
affects:
    - c3-109
    - ref-admin-page-layout
    - ref-list-view-patterns
    - ref-responsive-layout
    - ref-ui-patterns
approved-files:
    - .c3/refs/ref-list-view-patterns.md
    - .c3/refs/ref-admin-page-layout.md
    - .c3/refs/ref-responsive-layout.md
    - .c3/refs/ref-ui-patterns.md
    - apps/start/src/screens/WorkbenchScreen.tsx
    - .c3/c3-1-web-frontend/c3-109-workbench-screen.md
base-commit: 0df454143c863e545b63d9f410fa7872ab6f50c9
---

# Unify UI Patterns Between Workbench and Admin Views

## Goal

Align Workbench UI behavior with established admin and responsive layout refs while documenting any newly introduced reusable patterns.

## Problem

The workbench screen (c3-109) was built rapidly and introduced several UI patterns -- tabbed operational tables, sticky bottom action bars, MonthPicker sticky headers -- that work well but are not documented in any ref. Meanwhile, its tables lack the responsive column hiding (`hidden md:table-cell`) that ref-responsive-layout mandates for admin tables. The result is undocumented patterns that future screens cannot discover, and mobile table usability that falls below the standard set by AuditLogScreen and NotificationLogScreen.

## Decision

1. Document the new patterns in refs so they become sanctioned and discoverable.
2. Align WorkbenchScreen.tsx code with existing responsive conventions.
3. Update component doc c3-109 to reflect the alignment.

No routing or auth model changes. Workbench stays at `/_authed/workbench` -- not under `/admin/*` -- because finance team users need workbench access without owner privileges.

## Rationale

Documenting patterns before fixing code ensures the code changes have a spec to point to. This is the same ref-first approach used in ADR-20260211 (responsive admin layout). The alternative -- fixing code without updating refs -- would leave the patterns undiscoverable and risk future screens re-inventing them.

## Work Breakdown

### Task 1: Update ref-list-view-patterns (doc only)

- File: `.c3/refs/ref-list-view-patterns.md`
- Add "Tabbed Operational Table" as pattern #5 describing the workbench's tabbed interface where each tab contains a table with bulk selection and actions
- Add WorkbenchScreen to the Decision Flow

### Task 2: Update ref-admin-page-layout (doc only)

- File: `.c3/refs/ref-admin-page-layout.md`
- Add WorkbenchScreen to "Applies To" with note about tabbed variant

### Task 3: Update ref-responsive-layout (doc only)

- File: `.c3/refs/ref-responsive-layout.md`
- Document "Sticky Bottom Action Bar" as sanctioned mobile pattern
- Document "Sticky Header Bar" (MonthPicker + actions) pattern

### Task 4: Update ref-ui-patterns (doc only)

- File: `.c3/refs/ref-ui-patterns.md`
- Document MonthPicker as a period selection convention
- Document sticky header bar (period selector + action buttons) pattern

### Task 5: Align WorkbenchScreen.tsx responsive columns (code)

- File: `apps/start/src/screens/WorkbenchScreen.tsx`
- Export PRs table: add `hidden md:table-cell` to Bank, Account, Type column headers and cells
- Invoice Cleanup table: add `hidden md:table-cell` to Date column header and cells
- Standardize tab content wrapper spacing to use `admin-page-content` conventions
- Depends on: Tasks 1-4 (refs must be updated first so code can reference them)

### Task 6: Update c3-109 component doc

- File: `.c3/c3-1-web-frontend/c3-109-workbench-screen.md`
- Add ref-responsive-layout to Uses table
- Document responsive column hiding behavior
- Depends on: Task 5

## Verification

- `bunx @typescript/native-preview --noEmit` passes from `apps/start/` (no type errors introduced)
- All ref docs have valid markdown structure
- WorkbenchScreen.tsx tables use `hidden md:table-cell` on low-priority columns
- c3-109 doc accurately reflects the code

## Context

N.A - historical ADR; original context is captured in the git log around the ADR date and in the current code that implements the decision.

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
