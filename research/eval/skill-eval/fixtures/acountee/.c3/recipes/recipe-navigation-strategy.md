---
id: recipe-navigation-strategy
c3-seal: 61c4e0b1d5ebbb837311c7de290e37b2387e2765da26755bf3bf0f5bd91e48d1
title: Navigation Strategy
type: recipe
goal: Trace how users navigate the app — routing, sidebars, mobile drawers, and within-screen navigation patterns.
sources:
    - c3-1
    - c3-107
    - c3-109
    - ref-admin-page-layout
    - ref-list-view-patterns
    - ref-master-detail-layout
    - ref-responsive-layout
---

# Navigation Strategy

## Goal

Trace how users navigate the app — routing, sidebars, mobile drawers, and within-screen navigation patterns.

## Narrative

**Routing**: TanStack Start with file-based routing. SSR + client navigation.
Routes map to screens: invoices, payment-requests, payments, workbench,
and admin/* sub-routes.

**App sidebar** (AppSidebar): shadcn/ui `Sidebar` component with automatic
responsive behavior:

- Desktop: collapsible sidebar with icon/expanded states
- Mobile: hidden; hamburger in `MobileHeader` opens Sheet drawer

**Admin sidebar** (AdminSidebar): three-tier responsive pattern:

- Mobile (<768px): hidden, sticky header with hamburger → Sheet drawer overlay
- Tablet (768-1023px): narrow icon-only sidebar with tooltips
- Desktop (>=1024px): full sidebar with icons, titles, descriptions

**Within-screen navigation** depends on the list view pattern
(ref-list-view-patterns decision tree):

| Pattern | Navigation |
| --- | --- |
| Master-Detail | List panel + detail pane. Mobile: stacked with slide transition, MobileBackContext handles back chevron automatically |
| Full-Page Table | Expandable rows for drill-down, no separate detail view |
| Simple Table | Flat list, inline actions only |
| Tabbed Tile Layout | Tab switching between entity types (users/teams) |
| Tabbed Operational Table | Tab switching between tools, each with own table |

**Tabs**: shared Radix `Tabs` component everywhere. Two contexts:

- Detail pane: `TabsList` inside `DetailHeader` (with `border-b-0`)
- Page-level: `TabsList` inside `admin-page-header`

## Critical Rules

- Mobile sidebar is always a Sheet drawer — never visible inline
- `MobileBackContext` is internal to MasterDetailLayout — screens never wire back nav manually
- All tabs use the shared Radix component — no custom tab CSS or pill tabs
- Admin routes require owner role check before rendering
