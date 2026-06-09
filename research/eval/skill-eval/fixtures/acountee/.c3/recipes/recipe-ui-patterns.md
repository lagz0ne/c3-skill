---
id: recipe-ui-patterns
c3-seal: bfbab3da5f31d2ed4f400ab886db7a1ad844917cf5525758ae9cbdc661e80bde
title: UI Foundations
type: recipe
goal: Trace the frontend foundation stack — state management, forms, component variants, and formatting conventions.
sources:
    - c3-1
    - c3-101
    - c3-102
    - c3-103
    - ref-form-patterns
    - ref-ui-patterns
    - ref-variant-system
---

# UI Foundations

## Goal

Trace the frontend foundation stack — state management, forms, component variants, and formatting conventions.

## Narrative

**State** (c3-101): pumped-fn atoms for reactive stores. Controllers
expose mutable handles. NATS sync applies server deltas to atom state.
All UI state flows through the sync layer — mutations go to server,
server emits delta, delta updates atoms, React re-renders. No local
optimistic state that diverges from server truth.

**Forms** (c3-102): Zod-validated forms rendered in Drawer overlays.
Form patterns (ref-form-patterns) standardize field layout, validation
display, and submission flow. `Field` > `FieldLabel` > input component.
Multi-section forms use `<Separator />`.

**Components** (c3-103): shadcn/ui + Radix primitives. Styling via
Tailwind 4 + daisyUI. The variant system (ref-variant-system) uses
`tailwind-variants` for type-safe component variants — `button()`,
`badge()`, `listItem()`, etc.

**Formatting**: VND currency via `Intl.NumberFormat('vi-VN')` with
`font-mono`. Dates via `Intl.DateTimeFormat('vi-VN')` or `<ClientDate>`.
IDs and account numbers in `font-mono text-xs`.

**Status badges**: `getStatusBadgeVariant(status)` maps to badge variants:
success (active/approved), warning (pending/imported), info (completed),
destructive (failed/deleted), secondary (default).

**Loading**: `Loader2` from lucide — `h-4 w-4` inline, `h-8 w-8` page-level.

**Search**: All searchable lists use `IconSearch` + native `input` with
`h-8 pl-9`. No error states needed for simple search.

## See Also

- `recipe-responsive-design` — breakpoints, layout adaptation, sticky bars
- `recipe-modal-dialog` — drawers, modals, confirms, alerts
- `recipe-navigation-strategy` — sidebars, routing, tabs, mobile nav
