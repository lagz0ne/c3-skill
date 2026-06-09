---
id: ref-responsive-layout
c3-seal: dfcae5fdf1e3f491772230aaa4061bff4d98798c980462d9587d305857a0ed30
title: Responsive Layout Patterns
type: ref
goal: Define consistent breakpoints, responsive navigation behavior, and mobile-first layout rules that all screens follow. This ref is the single source of truth for responsive design decisions.
---

# Responsive Layout Patterns

## Goal

Define consistent breakpoints, responsive navigation behavior, and mobile-first layout rules that all screens follow. This ref is the single source of truth for responsive design decisions.

## Choice

- Three breakpoints: mobile (<768px), tablet (768-1023px), desktop (>=1024px) using Tailwind's mobile-first `min-width` prefixes only
- Navigation adapts per tier: hidden sidebar with sheet drawer on mobile, icon-only sidebar on tablet, full sidebar on desktop
- Sticky header bar for period selectors on desktop, sticky bottom action bar for mobile — actions always within reach

## Why

- Three breakpoints with mobile-first progression keep responsive rules simple and avoid conflicting `max-width` queries
- Tier-specific navigation ensures the sidebar never crowds content on small screens while remaining fully accessible via the sheet drawer
- Splitting actions between sticky header (desktop) and sticky bottom bar (mobile) keeps controls thumb-reachable on each form factor

## Breakpoints

| Name | Range | Tailwind Prefix | CSS Media Query |
| --- | --- | --- | --- |
| Mobile | <768px | (base) | default |
| Tablet | 768px-1023px | md: | @media (min-width: 768px) |
| Desktop | >=1024px | lg: | @media (min-width: 1024px) |

The `useIsMobile()` hook uses 768px as its breakpoint.

## Mobile-First Principle

| Rule | Why |
| --- | --- |
| Base styles target mobile | Most constrained viewport first |
| Add complexity with md: and lg: | Progressive enhancement |
| Never use max-width media queries in Tailwind | Stick to mobile-first min-width prefixes |

## App-Level Navigation

The app-level sidebar (`AppSidebar`) uses the shadcn/ui `Sidebar` component which handles responsive behavior automatically:

| Breakpoint | Behavior |
| --- | --- |
| Mobile | Hidden; hamburger button in MobileHeader opens Sheet drawer |
| Desktop | Collapsible sidebar with icon/expanded states |

## Admin Navigation

The admin sidebar (`AdminSidebar`) follows a three-tier responsive pattern:

| Breakpoint | Behavior |
| --- | --- |
| Mobile (<768px) | Sidebar hidden. Sticky header with hamburger button. Tap opens Sheet drawer overlay with full nav items. |
| Tablet (768-1023px) | Narrow icon-only sidebar visible. Icons with tooltips, no text labels. |
| Desktop (>=1024px) | Full sidebar with icons, titles, and descriptions. |

## MasterDetailLayout

| Breakpoint | Layout |
| --- | --- |
| Mobile (<768px) | Stacked: list or detail, slide transition between views |
| Tablet (768-1023px) | Side-by-side: list w-64, detail flex-1 |
| Desktop (>=1024px) | Side-by-side: list w-80, detail flex-1 |

## Table Responsive Rules

| Rule | Why |
| --- | --- |
| Wrap tables in overflow-x-auto on mobile | Prevent page-level horizontal scroll |
| Hide low-priority columns with hidden md:table-cell | Show essential data on mobile |
| Keep primary identifier + status visible at all viewports | Users can always identify the row |
| Expanded rows use overflow-x-auto on pre/code blocks | Long JSON/text content scrolls within cell |

Priority column guidelines for admin tables:

- **Always visible**: timestamp, primary action/status
- **Show at md+**: entity type, user
- **Show at lg+**: secondary identifiers, record IDs

## Filter Bar Responsive Rules

| Rule | Why |
| --- | --- |
| Use grid grid-cols-1 sm:grid-cols-2 md:grid-cols-5 | Stack on mobile, expand on desktop |
| All filter inputs use h-8 | Compact, consistent sizing at all breakpoints |
| Search/Clear buttons always visible | Users can always apply or reset filters |

## Pagination Responsive Rules

| Rule | Why |
| --- | --- |
| Stack count above controls on mobile: flex flex-col sm:flex-row | Avoid cramped layout |
| "Showing X-Y of Z" on top/left | Context always visible |
| Prev/Next buttons on bottom/right | Standard placement |
| Hide "Page N of M" text on mobile with hidden sm:inline | Save horizontal space |

## Tile Grid Responsive Rules

| Rule | Why |
| --- | --- |
| Use grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 | Progressive column count |
| Tile minimum width: 280px | Readable content on all viewports |

## Detail Content Responsive Rules

| Rule | Why |
| --- | --- |
| admin-detail-grid stacks to single column on mobile | Fields readable without horizontal scroll |
| At sm: and above, 2-column grid with even items padded | Original behavior preserved |
| Detail footer buttons: flex flex-wrap gap-2 | Buttons wrap on mobile if needed |

## Sticky Header Bar (Period Selector + Actions)

For screens with period-scoped data, a sticky header bar pins the period selector and action buttons at the top of scrollable content so they remain accessible while scrolling through the table.

| Rule | Why |
| --- | --- |
| Use sticky top-0 z-10 bg-background on the header bar | Keeps controls visible while scrolling table content |
| Period selector on the left, action buttons on the right | Consistent layout: filter left, actions right |
| Action buttons use hidden sm:flex on desktop | Desktop shows actions inline in the header bar |
| Add pb-3 to the sticky bar | Visual separation from table content below |

**CSS convention:**

```
sticky top-0 z-10 bg-background pb-3 flex items-center justify-between gap-3
```

## Sticky Bottom Action Bar (Mobile)

On mobile, when the sticky header bar hides its action buttons (`hidden sm:flex`), a sticky bottom bar provides the same actions at thumb-reach. This pattern avoids forcing users to scroll back up to act on their selection.

| Rule | Why |
| --- | --- |
| Show only on mobile: sm:hidden sticky bottom-0 | Desktop uses inline header actions instead |
| Include item count on the left, action buttons on the right | Context + action in one bar |
| Use bg-background border-t py-3 | Visual separation from table, consistent with page chrome |
| Pair with hidden sm:block summary text for desktop | Desktop shows count as static text below the table |

**CSS convention:**

```
sm:hidden sticky bottom-0 bg-background border-t py-3 flex items-center justify-between
```

**Screens using these patterns:**

- WorkbenchScreen — MonthPicker + bulk actions in sticky header, mobile action bar in each tab

## Applies To

- All screens in the application
- `ref-admin-page-layout` (references this ref for breakpoints)
- `ref-master-detail-layout` (references this ref for breakpoints)
- AdminSidebar component
- MasterDetailLayout component

## Cited By

- `ref-admin-page-layout` (references breakpoints)
- `ref-master-detail-layout` (references breakpoints)
