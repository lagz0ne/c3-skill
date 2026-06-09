---
id: adr-20260211-responsive-admin-layout
c3-seal: dfbcfa6bfba21616ad68ed0e4dcb2a95f4278b6b7c873ac399c6fa513501ba84
title: Responsive Admin Layout and Mobile Navigation
type: adr
goal: Define and complete a responsive layout strategy so admin and core screens remain usable across mobile, tablet, and desktop breakpoints.
status: implemented
date: "2026-02-11"
affects:
    - c3-1
    - c3-103
    - c3-104
    - c3-105
    - c3-107
approved-files:
    - .c3/refs/ref-responsive-layout.md
    - apps/start/src/routes/_authed/admin.tsx
    - apps/start/src/components/admin/AdminSidebar.tsx
    - apps/start/src/styles.css
    - apps/start/src/screens/admin/UserManagementScreen.tsx
    - apps/start/src/screens/admin/TeamManagementScreen.tsx
    - apps/start/src/screens/admin/ApprovalConfigScreen.tsx
    - apps/start/src/screens/admin/AuditLogScreen.tsx
    - apps/start/src/screens/admin/NotificationLogScreen.tsx
    - apps/start/src/screens/admin/OrganizationScreen.tsx
    - apps/start/src/screens/InvoiceScreen.tsx
    - apps/start/src/screens/PaymentRequestsScreen.tsx
    - apps/start/src/components/MasterDetailLayout.tsx
base-commit: e7dd79709d6f8d294c31841bea9e3fb7c9b64140
---

# Responsive Admin Layout and Mobile Navigation

## Goal

Define and complete a responsive layout strategy so admin and core screens remain usable across mobile, tablet, and desktop breakpoints.

## Problem

The admin area uses a fixed 16rem CSS sidebar (`.admin-sidebar`) that has zero responsive behavior. On mobile devices, the sidebar permanently consumes screen width, leaving the content area unusable. The main app screens (InvoiceScreen, PaymentRequestsScreen) have basic mobile support via MasterDetailLayout but lack tablet optimization and have responsive gaps in filter bars, tables, and detail panes. There is no documented convention for responsive breakpoints or mobile navigation patterns.

## Decision

1. **Create `ref-responsive-layout`** -- a single new ref defining mobile-first breakpoints, responsive container rules, and the admin mobile navigation pattern (hamburger + drawer). All existing layout refs (`ref-admin-page-layout`, `ref-master-detail-layout`) will reference this ref for breakpoint definitions.
**Create `ref-responsive-layout`** -- a single new ref defining mobile-first breakpoints, responsive container rules, and the admin mobile navigation pattern (hamburger + drawer). All existing layout refs (`ref-admin-page-layout`, `ref-master-detail-layout`) will reference this ref for breakpoint definitions.
2. **Redesign the admin sidebar** -- replace the fixed-width CSS sidebar with a responsive pattern:
**Redesign the admin sidebar** -- replace the fixed-width CSS sidebar with a responsive pattern:

- Mobile (<768px): hidden sidebar, hamburger button in a sticky header, sidebar opens as a slide-out drawer (Sheet)
- Tablet (768px-1023px): collapsible icon-only sidebar
- Desktop (>=1024px): full sidebar with labels

1. **Update all admin screens** for mobile/tablet responsiveness using Tailwind responsive prefixes (sm:, md:, lg:).
**Update all admin screens** for mobile/tablet responsiveness using Tailwind responsive prefixes (sm:, md:, lg:).
2. **Refine main app screens** (InvoiceScreen, PaymentRequestsScreen) for better mobile/tablet experience in filter bars, detail content, and action areas.
**Refine main app screens** (InvoiceScreen, PaymentRequestsScreen) for better mobile/tablet experience in filter bars, detail content, and action areas.

## Rationale

- The app-level `AppSidebar` already uses shadcn/ui's `Sidebar` component with Sheet-based mobile drawer, proving the pattern works in this codebase. The admin sidebar should follow the same approach rather than inventing a different one.
- Tailwind responsive prefixes (sm:, md:, lg:) are already the styling system in use. No new CSS framework or runtime JS media queries needed beyond what exists.
- A single `ref-responsive-layout` ref avoids scattering responsive rules across multiple refs and gives all screens a single source of truth for breakpoints.

## Work Breakdown

### Task 1: Create ref-responsive-layout (no dependencies)

Create `.c3/refs/ref-responsive-layout.md` defining:

- Breakpoint definitions (mobile <768px, tablet 768-1023px, desktop >=1024px)
- Mobile-first principle (base styles = mobile, add complexity with sm:/md:/lg:)
- Admin mobile navigation pattern (hamburger + Sheet drawer)
- Table responsive rules (horizontal scroll wrapper, priority column hiding)
- Filter bar responsive rules (stack on mobile, grid on desktop)
- Container max-width and padding rules per breakpoint

Files: `.c3/refs/ref-responsive-layout.md`

### Task 2: Redesign admin layout and sidebar (no dependencies, parallel with Task 1)

Replace fixed `.admin-sidebar` with responsive navigation:

- Update `AdminSidebar.tsx`: responsive sidebar using Sheet for mobile drawer, collapsible for tablet, full for desktop
- Update `admin.tsx` route: add mobile header with hamburger trigger for admin area
- Update `styles.css`: replace fixed `.admin-sidebar` CSS with responsive classes

Files: `apps/start/src/routes/_authed/admin.tsx`, `apps/start/src/components/admin/AdminSidebar.tsx`, `apps/start/src/styles.css`

### Task 3: Update admin MasterDetail screens (depends on Task 2)

Update UserManagementScreen, TeamManagementScreen, ApprovalConfigScreen for responsive:

- Ensure detail content grids stack on mobile (grid-cols-1 on mobile, grid-cols-2 on desktop)
- Action buttons stack or wrap on mobile
- Drawer forms respect mobile viewport

Files: `apps/start/src/screens/admin/UserManagementScreen.tsx`, `apps/start/src/screens/admin/TeamManagementScreen.tsx`, `apps/start/src/screens/admin/ApprovalConfigScreen.tsx`

### Task 4: Update admin table screens (depends on Task 2)

Update AuditLogScreen, NotificationLogScreen for responsive:

- Filter bars: stack to 1-2 columns on mobile
- Tables: horizontal scroll wrapper, hide low-priority columns on mobile via responsive display classes
- Pagination: compact layout on mobile
- Expanded rows: full-width on mobile

Files: `apps/start/src/screens/admin/AuditLogScreen.tsx`, `apps/start/src/screens/admin/NotificationLogScreen.tsx`

### Task 5: Update OrganizationScreen (depends on Task 2)

Update tile grid for responsive:

- Grid columns: 1 on mobile, 2 on tablet, 3 on desktop
- Tab headers: adjust spacing for mobile
- Search bar: full width on mobile

Files: `apps/start/src/screens/admin/OrganizationScreen.tsx`

### Task 6: Refine main app screens (no admin dependency, parallel)

Update InvoiceScreen, PaymentRequestsScreen for mobile/tablet:

- Filter panels: responsive grid layout
- Detail pane content: responsive padding and grid adjustments
- Action footers: responsive button layout

Files: `apps/start/src/screens/InvoiceScreen.tsx`, `apps/start/src/screens/PaymentRequestsScreen.tsx`

### Task 7: Update MasterDetailLayout component (parallel with Task 6)

Refine the shared layout component:

- Add tablet breakpoint behavior (list panel narrower at md, full at lg)
- Ensure transitions work across all three breakpoints

Files: `apps/start/src/components/MasterDetailLayout.tsx`

### Task 8: Update C3 docs (depends on all above)

- Update `ref-admin-page-layout` to reference `ref-responsive-layout` for breakpoints
- Update `ref-master-detail-layout` to reference `ref-responsive-layout`
- Update `c3-107` (Admin Screens) doc with responsive behavior notes
- Update `c3-1` (Web Frontend) container to list new ref

## Verification

1. `bunx @typescript/native-preview --noEmit` from `apps/start/` -- no type errors
2. Visual verification at mobile (375px), tablet (768px), and desktop (1280px) viewports:

- Admin sidebar: hamburger+drawer on mobile, collapsible on tablet, full on desktop
- Admin tables: horizontally scrollable on mobile, filters stack appropriately
- MasterDetail screens: list/detail stacking works at all breakpoints
- Organization tiles: grid adjusts per breakpoint

1. No regressions in existing desktop behavior

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
