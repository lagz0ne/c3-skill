---
id: adr-20260226-ui-pattern-review-gap-closure
c3-seal: adb595fc64f2b56a20c6d21454bda7baec500d08580233f2759d15070f3b9f3e
title: Reconcile UI Pattern Refs with Frontend Implementation
type: adr
goal: Align the current UI pattern refs with actual frontend implementation in `apps/start/src/screens` so architecture docs are accurate, and capture the remaining migration work as an explicit backlog.
status: implemented
date: "2026-02-26"
affects:
    - c3-1
    - c3-107
    - c3-109
    - ref-admin-page-layout
    - ref-form-patterns
    - ref-ui-patterns
approved-files:
    - .c3/refs/ref-ui-patterns.md
    - .c3/refs/ref-admin-page-layout.md
    - .c3/refs/ref-form-patterns.md
    - .c3/c3-1-web-frontend/c3-107-admin-screens.md
base-commit: 81cf760
---

# Reconcile UI Pattern Refs with Frontend Implementation

## Goal

Align the current UI pattern refs with actual frontend implementation in `apps/start/src/screens` so architecture docs are accurate, and capture the remaining migration work as an explicit backlog.

## Problem

Recent `ref-*` documentation describes standardized UI patterns, but a few areas in frontend screens still diverge:

- Deletion confirmation uses native `window.confirm()` in some admin screens.
- Admin full-page table docs did not describe current mobile filter collapse + sticky pagination behavior.
- Form docs described a strict Zod-first model that is not yet adopted in screen-level forms.

## Decision

Update pattern references to reflect current FE behavior as the source of truth, while preserving target standards and explicitly documenting remaining migration items.

## Work Breakdown

1. Audit UI pattern refs against implemented screens (Workbench, admin screens, master-detail screens, Payments).
2. Update `ref-ui-patterns` with compliance snapshot and explicit confirm-dialog migration gap.
3. Update `ref-admin-page-layout` to include implemented mobile behavior for filters and pagination.
4. Update `ref-form-patterns` to distinguish current baseline from target conventions.
5. Update `c3-107-admin-screens` to remove stale claim that all deletes use `ConfirmDrawer`.

## Findings Addressed

| Gap | Impact | Resolution |
| --- | --- | --- |
| Native confirm() still used in admin delete flows | Pattern drift and inconsistent UX | Documented as explicit migration gap in ref-ui-patterns and c3-107 |
| Admin log screens have mobile behavior not represented in ref | Docs incompletely describe FE behavior | Added mobile filter toggle + sticky mobile pagination guidance to ref-admin-page-layout |
| Form ref assumes Zod everywhere | Docs overstate current implementation | Updated ref-form-patterns to distinguish current baseline vs target adoption |

## Remaining Improvement Backlog

1. Replace native `window.confirm()` in `OrganizationScreen`, `UserManagementScreen`, `TeamManagementScreen` with `ConfirmModal`/`ConfirmDrawer`.
2. Extract duplicated user/team create-edit dialog implementations into shared form components.
3. Introduce Zod schemas for high-risk forms first (user/team CRUD and workbench import) and standardize inline field error rendering.
4. Add lint/check rule to flag new `window.confirm()` usage in frontend screens.

## Audit Record

| Phase | Date | Notes |
| --- | --- | --- |
| Audit | 2026-02-26 | Compared UI refs against FE screens |
| Ref update | 2026-02-26 | Updated UI/admin/form refs to reflect implementation and migration gaps |

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

## Verification

| Check | Result |
| --- | --- |
| Merged and running in production | PASS - see git log for the merge commit |
