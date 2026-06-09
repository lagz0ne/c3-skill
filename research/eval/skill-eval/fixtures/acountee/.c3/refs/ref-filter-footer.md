---
id: ref-filter-footer
c3-seal: 04580b81998f1f872c964a35120a11cbbf414e2e2aae9a42020d2e81e353495a
title: Filter Footer Pattern
type: ref
goal: Provide a consistent, space-efficient filtering UI that lives in the list panel footer. The filter panel slides up from the footer bar, keeping filters accessible without consuming permanent screen space.
---

# Filter Footer Pattern

## Goal

Provide a consistent, space-efficient filtering UI that lives in the list panel footer. The filter panel slides up from the footer bar, keeping filters accessible without consuming permanent screen space.

## Choice

- Filters live in a slide-up panel from the footer bar, not in the header or a sidebar
- `FilterFooter` compound component with context provider manages open/close state and active filter count
- `CompactCount` shows filtered vs total count for immediate feedback on filter impact

## Why

- Footer placement preserves list viewport height -- filters don't shrink the visible list
- Slide-up panel keeps filters accessible without consuming permanent screen space
- Active filter count badge ensures users always know when filters are applied

## Convention

| Rule | Why |
| --- | --- |
| Use FilterFooter compound component for all list filtering | Consistent filter UX |
| Filters live in the footer, not the header | Preserves list viewport |
| Show active filter count badge on toggle | Users know when filters are active |
| ESC or click-outside closes the panel | Standard dismissal |
| CompactCount shows filtered vs total | Quick feedback on filter impact |

## Component API

Compound component with context provider:

```tsx
<FilterFooter
  hasActiveFilters={hasActiveFilters}
  activeFilterCount={count}
  stats={<CompactCount total={total} filtered={filtered} />}
  actions={<button onClick={onClear}>Clear</button>}
>
  {/* Filter form fields */}
  <FilterFields />
</FilterFooter>
```

### Sub-components (advanced usage)

| Component | Purpose |
| --- | --- |
| FilterFooter.Root | Context provider, ESC handler |
| FilterFooter.Panel | Slide-up drawer (absolute, max-h 50vh, z-10) |
| FilterFooter.Bar | Always-visible footer bar |
| FilterFooter.Stats | Static stats text display (non-interactive) |
| FilterFooter.Toggle | Filter icon button — variant changes: ghost (no filters), secondary (active filters), default (panel open) |
| FilterFooter.StatsToggle | Clickable stats that opens filter panel (combines Stats + Toggle) |
| FilterFooter.CompactCount | Shows {filtered} or {total} count |
| FilterFooter.Actions | Right-aligned action buttons |

## Behavior

| Trigger | Result |
| --- | --- |
| Click stats/toggle | Panel slides up from footer |
| ESC key | Panel closes |
| Click outside panel | Panel closes |
| Filter change | Count updates, CompactCount reflects filtering |
| Clear filters | Reset all, close panel |

## CSS Classes

| Class | Purpose |
| --- | --- |
| filter-footer | Root container, safe-area bottom |
| filter-footer-bar | Always-visible bar (min-h 2.25rem) |
| filter-footer-panel | Slide-up drawer with opacity/transform transition |
| filter-footer-panel-inner | Inner flex column (min-h 0, flex 1) — wraps header + content |
| filter-footer-panel-header | Panel title + close button |
| filter-footer-panel-content | Scrollable filter content |
| filter-footer-stats | Stats text display |
| filter-footer-stats-toggle | Clickable stats button (flex-shrink 0) |
| filter-footer-toggle | Filter icon button |
| filter-footer-actions | Right-aligned actions (ml-auto, flex, gap 0.5rem) |

## Applies To

- InvoiceScreen (invoice filters: status, date range, search, archived)
- PaymentRequestsScreen (PR filters: statuses, sort, view mode, amount range, date range, creator)

## Cited By

- `ref-master-detail-layout` (list footer slot)
