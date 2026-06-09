---
id: ref-org-tiles
c3-seal: 423dfb9516bd82afe3cc112f492a48bacfe6b35b01fda8dd14783212f87e9b20
title: Organization Tile Grid
type: ref
goal: Provide a card-based grid layout for browsing entities with moderate detail. Used when entities benefit from visual scanning rather than tabular rows.
---

# Organization Tile Grid

## Goal

Provide a card-based grid layout for browsing entities with moderate detail. Used when entities benefit from visual scanning rather than tabular rows.

## Choice

- Auto-fill CSS grid (`minmax(280px, 1fr)`) for responsive card layout without breakpoints
- Each tile follows a consistent avatar + body + meta anatomy with optional expandable member section
- Actions exposed via `DropdownMenu` on the tile header, not inline buttons

## Why

- Auto-fill grid adapts to any container width without breakpoint management
- Card anatomy (avatar + body + meta) enables visual scanning -- faster entity recognition than table rows
- DropdownMenu keeps the tile surface clean, surfacing actions only on intent

## Convention

| Rule | Why |
| --- | --- |
| Use org-tiles auto-fill grid | Responsive card layout without breakpoint management |
| Each tile follows avatar + body + meta structure | Consistent card anatomy |
| Actions via DropdownMenu on tile header | Clean, non-cluttered tile surface |
| Expandable tiles for nested content (e.g., team members) | Drill-down without leaving the grid |

## Structure

```
org-tiles (auto-fill grid, minmax 280px/300px)
└── org-tile (card: bg, border, radius, hover shadow)
    ├── org-tile-header (avatar + dropdown trigger)
    │   ├── org-tile-avatar (-team, -role, -system variants)
    │   └── DropdownMenu (Edit, Delete)
    ├── org-tile-body (title + desc + meta)
    │   ├── org-tile-title (0.875rem, 500 weight)
    │   ├── org-tile-desc (0.75rem, muted, 2-line clamp)
    │   └── org-tile-meta (flex wrap of meta-items)
    └── org-tile-members (expandable, border-top, bg-muted/30)
        ├── org-tile-member (icon + email)
        └── org-tile-members-empty
```

## Avatar Variants

| Variant | Color | Usage |
| --- | --- | --- |
| org-tile-avatar | Muted background | Users |
| org-tile-avatar-team | Info/10 background | Teams |
| org-tile-avatar-role | Warning/10 background | Roles |
| org-tile-avatar-system | Primary/10 background | System entities |

## Expandable Tiles

| State | Treatment |
| --- | --- |
| Collapsed | Standard tile, org-tile-body is clickable |
| Expanded | org-tile-expanded class adds primary border, shows org-tile-members section |
| Empty members | org-tile-members-empty centered muted text |

## Page Integration

Used inside `admin-page` + page-level `Tabs`:

```admin-page
├── admin-page-header
│   ├── TabsList (Users | Teams tabs with count badges)
│   ├── Search input
│   └── Add button (label changes per tab)
├── org-content (flex-1, overflow-y auto)
│   ├── TabsContent "users" → org-tiles with user tiles
│   └── TabsContent "teams" → org-tiles with team tiles
└── (no pagination — data pre-fetched)
```

## CSS Classes

| Class | Purpose |
| --- | --- |
| org-content | Scrollable content area |
| org-loading / org-empty | Loading/empty states (centered, h-12rem) |
| org-tiles | Auto-fill grid container |
| org-tile | Card (border, radius, hover effect) |
| org-tile-system | System entity variant (primary border tint) |
| org-tile-header | Top row (avatar + actions) |
| org-tile-actions | Action buttons container (flex, align-center, gap 0.25rem) |
| org-tile-avatar | 2.5rem circle, with -team, -role, -system variants |
| org-tile-body | Content area (title, desc, meta) |
| org-tile-title | Entity name |
| org-tile-desc | Description (2-line clamp) |
| org-tile-meta | Meta info row (flex wrap) |
| org-tile-meta-item | Individual meta piece (icon + text) |
| org-tile-expanded | Expanded state border |
| org-tile-members | Expandable member list |
| org-tile-member | Member row |
| org-tile-member-email | Member email (truncated) |
| org-tile-members-empty | Empty members message |
| org-tile-footer | Tile footer (flex wrap, gap 0.375rem, border-top, bg-muted/30) |

## Applies To

- OrganizationScreen (Users tab + Teams tab)

## Cited By

- `ref-list-view-patterns` (Tabbed Tile Layout pattern)
- `ref-ui-patterns` (page-level tabs)
