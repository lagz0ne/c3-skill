---
id: recipe-modal-dialog
c3-seal: 28f9374b0fd246ade6dd881f2954d3bb0bf720e130d039f3f7de3a7bc937597b
title: Modal & Dialog Patterns
type: recipe
goal: Trace when and how to use each overlay type — drawers for forms, modals for confirmations, alerts for feedback.
sources:
    - c3-102
    - c3-103
    - ref-form-patterns
    - ref-ui-patterns
---

# Modal & Dialog Patterns

## Goal

Trace when and how to use each overlay type — drawers for forms, modals for confirmations, alerts for feedback.

## Narrative

Three overlay types, each with a single purpose:

**Drawers** — slide-out panels for create/edit forms. Anatomy:
`Drawer` > `form` > `DrawerBody` + `DrawerActions`. Width `md` for simple
forms, `lg` for multi-section. Cancel (outline) + Submit (default) buttons.
Loading state: `Loader2` spinner + disabled button. Multi-section forms
use `<Separator />` between field groups.

**Modals** — centered overlays for confirmations and non-form content.
Two variants:

- `ConfirmModal` — pre-built confirm/cancel with `variant` (danger/warning/info)
and `isLoading` prop. Used for destructive actions.
- `ConfirmDrawer` — same confirmation pattern but as slide-out, used when
the action relates to a detail panel context.

**Alerts** — fixed-position feedback. Positioned `fixed top-4 right-4 z-50
max-w-md`. Success alerts auto-hide at 3000ms. Error alerts persist until
dismissed. One alert per screen at a time.

## Migration Backlog

Three screens still use `window.confirm()`:

- OrganizationScreen (delete user/team)
- UserManagementScreen (delete user)
- TeamManagementScreen (delete team)

**Rule**: New delete confirmations MUST use `ConfirmModal` or `ConfirmDrawer`.
Existing legacy usage migrated incrementally.

## Decision Tree

```
Is it a form (create/edit)?
├── Yes → Drawer (md or lg width)
└── No → Is it a confirmation?
    ├── Yes → In detail panel context?
    │   ├── Yes → ConfirmDrawer
    │   └── No → ConfirmModal
    └── No → Is it feedback?
        └── Yes → Alert (fixed position)
```
