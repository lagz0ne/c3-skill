---
id: ref-admin-page-layout
c3-seal: 316963a7a4da1bfe5de96020bcf930c512a755e1093531aa6fc17a5189d8451b
title: Admin Page Layout
type: ref
goal: Provide consistent structure for admin screens that use full-page tables (not MasterDetailLayout). These screens display large, paginated datasets with filters.
---

# Admin Page Layout

## Goal

Provide consistent structure for admin screens that use full-page tables (not MasterDetailLayout). These screens display large, paginated datasets with filters.

## Choice

- Four-zone layout: header + filter bar + scrollable table + pagination footer, all within an `admin-page` flex-column container
- Server-side pagination with `PAGE_SIZE = 25`, always-visible pagination footer (even with 0 results)
- Responsive table columns via `hidden md:table-cell` and `overflow-x-auto` wrappers
- Mobile behavior for log-style pages: collapsible filter section and sticky bottom pagination while filters are closed

## Why

- A predictable 4-zone structure means users always know where to find filters, data, and navigation regardless of which admin screen they're on
- Always-visible pagination gives users dataset size context at all times, preventing confusion about whether data is missing or simply on another page
- Progressive column hiding keeps essential data (timestamp, status) visible on mobile without horizontal scrolling
- Collapsible filters plus sticky pagination preserve mobile viewport space while keeping controls reachable

## Convention

| Rule | Why |
| --- | --- |
| Use admin-page class as root flex-column container | Consistent admin screen structure |
| Header + filters + scrollable table + pagination footer | Predictable 4-zone layout |
| Server-side pagination with PAGE_SIZE = 25 | Large datasets don't load all at once |
| Pagination footer always visible, even with 0 results | Users see count context at all times |
| On mobile, support filter collapse and sticky pagination | Better use of limited vertical space |

## Structure

```
admin-page (h-full, flex column, overflow hidden)
├── admin-page-header (flex-shrink-0, border-bottom)
│   ├── admin-page-title
│   ├── admin-page-subtitle (optional)
│   └── Action buttons (top-right)
├── Filter bar (flex-shrink-0, border-bottom, bg-muted/30)
│   └── Grid of filter fields (grid-cols-1 sm:grid-cols-2 md:grid-cols-5)
├── Scrollable table area (flex-1, min-h-0, overflow-auto)
│   └── overflow-x-auto wrapper
│       └── admin-table (w-full, border-collapse)
│           └── Low-priority columns: hidden md:table-cell
└── Pagination footer (flex-shrink-0, border-top, flex-col sm:flex-row)
    ├── "Showing X - Y of Z entries"
    └── Previous / Page N of M (hidden sm:inline) / Next
```

## Admin Table

| Rule | Why |
| --- | --- |
| Use admin-table class on <table> | Consistent table styling |
| Wrap table in overflow-x-auto | Prevent page-level horizontal scroll on mobile per ref-responsive-layout |
| Table headers: uppercase, letter-spacing, muted | Scannable column labels |
| Rows: hover highlight, cursor-pointer when expandable | Interactive feedback |
| Actions column: admin-table-actions right-aligned | Consistent action placement |
| Hide low-priority columns with hidden md:table-cell | Show essential data on mobile per ref-responsive-layout |
| Expandable rows: bg-muted/30 p-4 with colSpan={N} | Drill-down without leaving table |
| Expanded content uses overflow-x-auto | Long JSON/diff content scrolls within cell |

## Filter Bar

| Rule | Why |
| --- | --- |
| All filters in a single bar below the header | Visible, accessible |
| Log-style pages may collapse filters on mobile behind a toggle row (md:hidden) | Avoids large fixed filter blocks on small screens |
| Grid layout: grid-cols-1 sm:grid-cols-2 md:grid-cols-5 | Responsive filter layout per ref-responsive-layout |
| Select fields use "_all" sentinel for "All" option | Consistent empty-filter value |
| Search + Clear buttons at end of filter grid | Explicit filter application |
| All inputs use h-8 height | Compact, consistent sizing |

## Pagination

| Rule | Why |
| --- | --- |
| Always show footer with count + navigation | Users always know dataset size |
| Stack on mobile: flex flex-col sm:flex-row | Avoid cramped layout per ref-responsive-layout |
| On mobile log pages, footer can be sticky bottom-0 while filter panel is closed | Keeps page navigation reachable without scrolling |
| "Showing X - Y of Z entries" on top/left | Context for current page |
| Previous/Next with page info on bottom/right | Standard pagination controls |
| Hide "Page N of M" on mobile: hidden sm:inline | Save horizontal space |
| Show "0 - 0 of 0" when empty | Never hide pagination |
| PAGE_SIZE = 25 | Consistent page size |

## Empty State

Uses the project-wide `empty-page-*` classes (defined in `styles.css`):

| Class | Purpose |
| --- | --- |
| empty-page | Centered container (h-full, flex column, p-3rem) |
| empty-page-icon | 3rem icon, muted, 50% opacity, stroke-width 1.5 |
| empty-page-title | 1rem, 500 weight |
| empty-page-desc | 0.8125rem, muted, max-w 24rem |
| empty-page-action | CTA button wrapper (margin-top 1rem) |

## CSS Classes

| Class | Purpose |
| --- | --- |
| admin-page | Root (h-full, flex column, overflow hidden) |
| admin-page-header | Header (flex-shrink-0, p 1.25rem 1.5rem, border-bottom) |
| admin-page-title | Title (1.125rem, 600 weight) |
| admin-page-subtitle | Subtitle (0.8125rem, muted) |
| admin-page-content | Scrollable content area (flex 1, flex column, overflow auto, p 1.5rem) — used when content is not a table. Flex-col ensures children (e.g. TabsContent) can stretch to fill height for empty states. |
| admin-table | Table (w-full, border-collapse, 0.8125rem) |
| admin-table-actions | Actions cell (flex, gap, justify-end) |
| empty-page | Empty state container (project-wide) |
| empty-page-icon / empty-page-title / empty-page-desc / empty-page-action | Empty state parts |

## Applies To

- AuditLogScreen (paginated audit entries with expandable diff)
- NotificationLogScreen (paginated notifications with retry)
- PaymentsScreen (simple table variant — no filters, no pagination)
- OrganizationScreen (`admin-page` root structure, without table)
- WorkbenchScreen (tabbed operational table variant — `admin-page` + `Tabs` in header, `admin-table` per tab, no filters/pagination)

## Cited By

- `ref-list-view-patterns` (pattern selection)
- `ref-responsive-layout` (breakpoint definitions)
