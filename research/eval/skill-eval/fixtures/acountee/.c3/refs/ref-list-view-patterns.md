---
id: ref-list-view-patterns
c3-seal: f3e46a8d0051724ac8402823d7c26387009d44022bdefaf4bd8d4303f68e44f3
title: List View Patterns
type: ref
goal: Define when to use each list/table layout pattern. Every screen that displays a collection of items should use one of these patterns — not invent a new one.
---

# List View Patterns

## Goal

Define when to use each list/table layout pattern. Every screen that displays a collection of items should use one of these patterns — not invent a new one.

## Choice

- Five layout patterns: Master-Detail, Full-Page Table, Simple Table, Tabbed Tile Layout, and Tabbed Operational Table — each with clear selection criteria
- Decision tree based on entity richness, dataset size, operational tool count, and entity type count
- Feature screens use `@tanstack/react-virtual` with sticky group headers for performant list rendering

## Why

- A fixed set of patterns with explicit "when to use / when NOT to use" criteria prevents ad-hoc layout invention and keeps the UI consistent
- The decision tree encodes the selection logic so new screens can pick the right pattern without guessing
- Virtualization with sticky headers handles large lists (invoices, PRs) without performance degradation while maintaining group context

## Patterns

### 1. Master-Detail (MasterDetailLayout)

Side-by-side list + detail pane. Desktop: list 320px fixed, detail fills remaining. Mobile: stacked with slide transition.

**When to use:**

- Entity has rich detail that benefits from a persistent detail pane
- User workflow involves scanning the list and inspecting/acting on individual items
- Entity has detail tabs, approval workflows, or multi-section content

**When NOT to use:**

- Entity is simple enough that all info fits in a table row
- No meaningful "detail" beyond what the list shows
- Primary action is bulk operations, not per-item inspection

**Screens using it:** InvoiceScreen, PaymentRequestsScreen, UserManagementScreen, TeamManagementScreen, ApprovalConfigScreen

### 2. Full-Page Table (admin-page + admin-table)

Full-width table with header, optional filters, scrollable body, and pagination footer. The table fills the content area.

**When to use:**

- Large datasets that need server-side pagination
- Content is best read as rows (log entries, audit trail, notification history)
- Filters are important for narrowing results
- Row expansion for drill-down (e.g., diff view)

**When NOT to use:**

- Dataset is small and fits in memory (use Simple Table instead)
- User needs to see detail alongside the list (use Master-Detail)

**CSS classes:** `admin-page`, `admin-page-header`, `admin-table`, `empty-page`

**Screens using it:**

- AuditLogScreen — paginated audit entries with filters and expandable diff
- NotificationLogScreen — paginated notification logs with retry

### 3. Simple Table (admin-page + admin-table, no filters/pagination)

Same visual style as Full-Page Table but without filters or pagination. For small, manageable datasets where all items fit in a single scrollable list.

**When to use:**

- Small reference data (typically < 100 items)
- Minimal interaction — just view, add, edit, delete
- No need for filtering or pagination
- Data is pre-fetched, not server-paginated

**When NOT to use:**

- Dataset can grow large (use Full-Page Table with pagination)
- Entity needs rich detail pane (use Master-Detail)

**CSS classes:** Same as Full-Page Table: `admin-page`, `admin-page-header`, `admin-table`

**Screens using it:**

- PaymentsScreen — payment methods (small reference list)

### 4. Tabbed Tile Layout

Tabbed interface with search + grid of cards. Each tab shows a different entity type in a tile/card layout. Uses the shared Radix `Tabs` component (see `ref-ui-patterns` § Tabs).

**When to use:**

- Multiple related entity types viewed together (users + teams)
- Tile layout provides better scanning than table rows
- Actions are per-card via dropdown menus

**When NOT to use:**

- Single entity type (use Master-Detail or Table)
- Data has many columns (use Table)

**Screens using it:**

- OrganizationScreen — combined users/teams with tabs

### 5. Tabbed Operational Table (admin-page + Tabs + admin-table)

Tabbed interface where each tab contains an operational table with bulk selection and actions. Tabs live inside `admin-page-header`, and each `TabsContent` holds its own table, period selector, and action bar. Uses the shared Radix `Tabs` component (see `ref-ui-patterns` § Tabs).

**When to use:**

- Multiple operational tools sharing the same page, each with its own dataset
- Each tool has a table with bulk selection and per-row or bulk actions
- Data is period-scoped (month picker) rather than filtered or paginated
- Tools are related by role/workflow but operate on different entity types

**When NOT to use:**

- Single table view (use Simple Table or Full-Page Table)
- Tabs show the same entity type in different views (use Tabbed Tile Layout)
- Entity needs rich detail inspection (use Master-Detail)

**CSS classes:** `admin-page`, `admin-page-header` (with `Tabs` and `TabsList` inside), `admin-page-content`, `admin-table`

**Screens using it:**

- WorkbenchScreen — 3 tabs: Invoice Cleanup, Export PRs, Import Paid PRs

## Virtualized List with Sticky Group Headers

Feature screens (Invoice, PaymentRequests) use `@tanstack/react-virtual` for performant list rendering with month/status group headers.

| Rule | Why |
| --- | --- |
| Use useVirtualizer with custom rangeExtractor for sticky headers | Group headers stay visible while scrolling through group |
| Groups defined by month (InvoiceScreen) or status (PR grouped view) | Natural data organization |
| Group headers show: label left, count badge right | Scannable group sizes |
| Group headers in PR screen have color-coded left border (border-l-[3px]) | Visual status identification |
| activeStickyIndexRef tracks current sticky header | Single sticky header at a time |

**Two list modes (PaymentRequestsScreen):**

| Mode | Behavior |
| --- | --- |
| list | Flat virtualized list, no groups |
| grouped | Virtualized with sticky group headers |

Toggle via `viewMode` in filter state.

## Decision Flow

```
Does the entity have rich detail content?
├── Yes → Master-Detail
└── No → Is the dataset large (100+ items, paginated)?
    ├── Yes → Full-Page Table (with filters + pagination)
    └── No → Are there multiple operational tools on one page?
        ├── Yes → Tabbed Operational Table (tabs + period selector + bulk actions)
        └── No → Is there multiple entity types to browse?
            ├── Yes → Tabbed Tile Layout
            └── No → Simple Table (no filters, no pagination)
```

## Cited By

- `ref-master-detail-layout`
- `ref-detail-content-strategy`
- `ref-admin-page-layout`
- `ref-org-tiles`
