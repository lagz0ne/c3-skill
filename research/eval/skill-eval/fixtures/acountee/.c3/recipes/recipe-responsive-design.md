---
id: recipe-responsive-design
c3-seal: f07a0fb4b9dd28e0ca6b2f5555d2cd1a11f5d89e3ad51f24ec4b17f04cf7a623
title: Responsive Design
type: recipe
goal: Trace how the app adapts across mobile/tablet/desktop — breakpoints, layout shifts, sticky controls, and column hiding rules.
sources:
    - c3-109
    - ref-admin-page-layout
    - ref-master-detail-layout
    - ref-responsive-layout
---

# Responsive Design

## Goal

Trace how the app adapts across mobile/tablet/desktop — breakpoints, layout shifts, sticky controls, and column hiding rules.

## Narrative

Three breakpoints, mobile-first only (ref-responsive-layout):

| Tier | Range | Tailwind |
| --- | --- | --- |
| Mobile | <768px | base |
| Tablet | 768-1023px | md: |
| Desktop | >=1024px | lg: |

**Never use `max-width` queries.** `useIsMobile()` hook uses 768px.

**MasterDetailLayout** adapts per tier: stacked with slide transition
on mobile, side-by-side with 256px list on tablet, 320px list on desktop.

**Admin tables** hide low-priority columns with `hidden md:table-cell`.
Priority: always show timestamp + status; show entity type + user at `md:`;
show secondary IDs at `lg:`.

**Filter bars** use `grid grid-cols-1 sm:grid-cols-2 md:grid-cols-5`.
All inputs `h-8`. Log-style pages collapse filters on mobile behind a toggle.

**Pagination** stacks vertically on mobile (`flex flex-col sm:flex-row`),
hides "Page N of M" on mobile (`hidden sm:inline`). Always visible, even
with 0 results.

**Sticky controls** split by viewport:

- Desktop: sticky header bar (`sticky top-0 z-10`) with period selector
left, action buttons right (`hidden sm:flex`)
- Mobile: sticky bottom bar (`sm:hidden sticky bottom-0`) with count +
actions at thumb reach

## Critical Rules

- Base styles = mobile. Add complexity with `md:` and `lg:` only
- Primary identifier + status visible at ALL viewports
- Wrap tables in `overflow-x-auto` on mobile
- Desktop actions in header, mobile actions in bottom bar — never both visible
- Tile grids: `grid-cols-1 sm:grid-cols-2 lg:grid-cols-3`, min 280px per tile
