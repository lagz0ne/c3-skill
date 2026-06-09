---
id: ref-master-detail-layout
c3-seal: 8cbf41ab4522aae853739356c42be59efdadceba2e2b9d6d5489d7e41ee21b60
title: Master-Detail Layout Pattern
type: ref
goal: Provide a consistent layout pattern for screens displaying list + detail views, with automatic responsive behavior.
---

# Master-Detail Layout Pattern

## Goal

Provide a consistent layout pattern for screens displaying list + detail views, with automatic responsive behavior.

## Choice

- Single `MasterDetailLayout` component with named slot children (`listHeader`, `listContent`, `emptyState`, `detailContent`) for type-safe structure
- Three-tier responsive: side-by-side on desktop (320px list) and tablet (256px list), stacked with slide transition on mobile
- Internal `MobileBackContext` handles back navigation automatically — screens don't wire mobile nav manually

## Why

- Named slots enforce consistent structure across all list+detail screens while keeping each screen's content flexible
- Automatic responsive behavior with mobile slide transition means no per-screen mobile view implementations
- Centralizing mobile back navigation in the layout eliminates duplicated wiring across every screen

## Convention

| Rule | Why |
| --- | --- |
| Use for all list+detail screens | Consistent UX |
| Pass children as named slots | Type-safe structure |
| Let layout handle mobile nav | No custom mobile views |

## Structure

```typescript
<MasterDetailLayout
  selectedItem={selected}
  onBackToList={() => setSelected(null)}
>
  {{
    listHeader: <FilterPanel />,
    listContent: <ItemList />,
    emptyState: <EmptyMessage />,
    detailContent: <DetailView />
  }}
</MasterDetailLayout>
```

## Behavior

| Breakpoint | Layout |
| --- | --- |
| Desktop (>=1024px) | Side-by-side: list 320px (lg:w-80 lg:min-w-80), detail fills remaining (flex-1) |
| Tablet (768-1023px) | Side-by-side: list 256px (md:w-64 md:min-w-64), detail fills remaining (flex-1) |
| Mobile (<768px) | Stacked: detail slides over list (translate-x + opacity, 150ms ease-out) |

See [ref-responsive-layout](ref-responsive-layout.md) for breakpoint definitions.

## Exported Sub-components

| Component | Purpose |
| --- | --- |
| ListHeader | Flex row (justify-between items-center) for list panel title + actions |
| ListItem | Uses listItem() variant. Props: isSelected, status, onClick |
| EmptyState | Applies empty-page CSS class — centered container for empty lists. Children use empty-page-icon, empty-page-title, empty-page-desc, empty-page-action |
| DetailHeader | Header bar with mobile back button. min-h-[50px], border-b, p-3 |
| DetailContent | Scrollable content area (flex-1 overflow-auto). Accepts className |
| DetailFooter | Footer with footer-bar footer-bar-detail CSS classes |

## Slots

```typescript
children: {
  listHeader: ReactNode
  listContent: ReactNode
  listFooter?: ReactNode    // optional, renders in footer-bar-list
  emptyState: ReactNode
  detailContent?: ReactNode
}
```

Additional props: `selectedItem`, `onBackToList`, `isLoading`, `loadingText`, `detailEmptyState: { title, message }`, `className`.

## Detail Header

`DetailHeader` provides the mobile back button and a bottom border separator. For screens with tabs, `TabsList` goes inside `DetailHeader` — no entity identity content (name, ID, status badge). The facet grid inside the main tab handles identity. See `ref-ui-patterns` § Tabs for the full detail pane structure.

For admin screens without tabs (UserManagement, TeamManagement, ApprovalConfig), `DetailHeader` still holds entity identity content (name + status badge).

## Mobile Back Navigation

`MobileBackContext` is an internal context that passes `onBack` and `isMobile` to `DetailHeader`. This enables the back chevron button on mobile without screens needing to wire it manually.

- `MobileBackContext.Provider` wraps the entire layout with `{ onBack, isMobile }` from `onBackToList` prop
- `DetailHeader` reads context via `useContext(MobileBackContext)` to conditionally render `IconChevronLeft` button
- Screens never interact with this context directly — it's an internal mechanism of the layout component

## List Panel Empty State

The list content area uses `flex flex-col` + `flex-1` so `EmptyState` components center vertically. Screens render their empty states inside `listContent` (not the `emptyState` prop). See `ref-detail-content-strategy` for empty state styling rules.

## Applies To

- InvoiceScreen
- PaymentRequestsScreen
- UserManagementScreen
- TeamManagementScreen
- ApprovalConfigScreen

## Cited By

- `ref-list-view-patterns` (pattern selection)
- `ref-responsive-layout` (breakpoint definitions)
