---
id: ref-variant-system
c3-seal: b962283c3f13f7174346d881d5db79b45e198d9a2957bb8d15a71a506ca12680
title: Variant System (tailwind-variants)
type: ref
goal: Provide a single, type-safe styling API for all interactive UI elements. Every button, badge, input, and interactive component uses `tailwind-variants` (tv) — never raw Tailwind class strings composed inline.
---

# Variant System

## Goal

Provide a single, type-safe styling API for all interactive UI elements. Every button, badge, input, and interactive component uses `tailwind-variants` (tv) — never raw Tailwind class strings composed inline.

## Choice

- All interactive element styles defined as `tv()` variant functions in `variants.ts`
- Variant functions (e.g., `button()`, `badge()`) are called directly -- no wrapper components
- New variants are added to `variants.ts`, never composed inline

## Why

- Single source of truth per element type prevents visual drift across screens
- Type-safe variant API catches invalid prop combinations at compile time
- Centralizing in `variants.ts` makes the full design vocabulary discoverable in one file

## Convention

| Rule | Why |
| --- | --- |
| All interactive element styles defined as tv() variants | Single source of truth per element type |
| Use variant functions, not raw class strings | Type-safe, consistent, prevents drift |
| Add new variants to variants.ts, not inline | Discoverability, reuse |
| Prefer existing variants over one-off classes | Reduces visual inconsistency |

## Variant Catalog

### button

Primary interactive control. Used everywhere for actions.

| Variant | Usage |
| --- | --- |
| default / primary | Primary CTA (orange) |
| outline | Secondary actions (Edit, Cancel) |
| ghost | Tertiary/icon-only actions |
| destructive / error | Delete, remove, revoke |
| success | Positive actions (Restore, Approve) |
| warning | Caution actions |
| link | Inline text links |
| accent | Accent-colored actions |
| neutral | Neutral emphasis |

| Size | Usage |
| --- | --- |
| default / md | Standalone buttons |
| sm | Footer actions, inline actions |
| xs | List header actions, compact UI |
| icon / icon-sm / icon-xs | Icon-only buttons |

Loading state: `loading: true` adds spinner + disabled.

### badge

Status indicators and count labels.

| Variant | Usage |
| --- | --- |
| default | Primary emphasis (owner role) |
| secondary | Neutral labels (anyof mode, secondary roles) |
| outline | Counts, group overflow (+N) |
| success | Active, approved, sent, created |
| warning | Pending, imported, in-progress |
| info | Completed, informational |
| destructive / error | Inactive, failed, deleted, obsolete |
| neutral | Neutral emphasis |
| accent | Accent-colored labels |

| Size | Usage |
| --- | --- |
| default | Standard badges |
| sm | List group counts |
| xs | Inline role badges, count badges in list headers |

### listItem

List panel items in MasterDetailLayout.

| Prop | Values |
| --- | --- |
| selected | true/false — controls highlight + left border |
| status | none/inprogress/imported/completed/obselete — controls left border color when selected |

Note: the `obselete` spelling is a consistent typo across CSS and JS. Do not "fix" it — it would break class matching.

### checkbox

Form checkboxes with touch targets.

| Size | Usage |
| --- | --- |
| sm | Compact lists |
| md | Standard forms |
| lg | Touch-friendly |

Companion: `checkboxTouchTarget()` wraps checkbox with 44px min tap target.

### drawer

Slide-out panel for forms and detail views.

| Prop | Values |
| --- | --- |
| side | left/right |
| width | sm/md/lg/xl/full |

Slots: overlay, panel, header, title, subtitle, closeButton, body, actions.

### modal

Dialog overlay for confirmations and alerts.

| Prop | Values |
| --- | --- |
| size | sm/md/lg/xl/full |
| actionsAlign | left/center/right/between |

Slots: backdrop, box, header, headerTitle, title, subtitle, closeButton, body, actions.

### Other Variants

| Variant | Purpose |
| --- | --- |
| input | Text inputs with error state |
| select | Select dropdowns with error state |
| textarea | Multi-line inputs with error state |
| fileInput | File upload inputs |
| alert | Alert banners (success/error/warning/info) |
| confirmButton | Confirmation dialog buttons (danger/warning/info) |
| radioButton / radioGroup | Radio selections |
| formActions | Form footer layout (uses footer-bar CSS) |
| inputGroup | Grouped input layout (removes inner border-radii) |
| errorAlert | Error alert with type-specific styling (network/auth/validation/server) |
| passwordStrengthBar | Password strength indicator (levels 0-5) |
| toggleButton | Toggle button with active state |
| crudTable | Table slots (header/body/table/thead/th/tr/td etc.) |

## Applies To

Every screen. This is the foundation layer — all other UI refs build on these variants.

## Cited By

- `ref-form-patterns` (button variants in form actions)
- `ref-ui-patterns` (badge/button variants in UI patterns)
