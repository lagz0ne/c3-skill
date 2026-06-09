---
id: ref-bulk-operations
c3-seal: 32cf9f6df17edd9fecc9fb94a85ac3029d218aba18dc5382233d9f46318974c7
title: Bulk Operations Pattern
type: ref
goal: Allow users to select multiple list items and perform batch actions. The pattern overlays selection UI onto the existing list without changing the list layout.
---

# Bulk Operations Pattern

## Goal

Allow users to select multiple list items and perform batch actions. The pattern overlays selection UI onto the existing list without changing the list layout.

## Choice

- Bulk mode is an overlay: checkboxes appear inline in existing list items, action bar replaces the normal footer
- Activation via header button or keyboard shortcut (B); ESC exits and clears selection
- Non-actionable items shown at reduced opacity to communicate what can be acted on

## Why

- Overlaying selection avoids a separate "bulk view" -- users stay in context with the same list layout
- Keyboard shortcut enables power-user speed; ESC provides a safe exit
- Reduced opacity on non-actionable items prevents confusion about which items will be affected

## Convention

| Rule | Why |
| --- | --- |
| Toggle bulk mode via header button or keyboard shortcut (B) | Quick access without menu diving |
| Checkboxes appear inline in list items when active | Minimal UI change, familiar pattern |
| Select all/deselect all bar appears above list | Batch control |
| Bulk action bar replaces normal footer | Context-appropriate actions |
| Non-actionable items shown at reduced opacity | Clear which items can be acted on |

## Activation

| Trigger | Result |
| --- | --- |
| Press B key | Toggle bulk mode |
| Click bulk mode button in list header | Toggle bulk mode |
| ESC while in bulk mode | Exit bulk mode, clear selection |

## Selection UI

```
List Header
├── [Bulk Mode Toggle Button]
│
Select All Bar (visible when bulk mode active)
├── "Select All ({total})" / "Deselect All"
├── Selection count badge (primary/10, rounded-full)
│
List Items (with checkboxes)
├── [checkbox] Item content...  (opacity-100 if actionable)
├── [checkbox] Item content...  (opacity-50 if not actionable)
│
Bulk Action Bar (replaces normal footer)
├── Action buttons: "Approve (3)", "Obsolete (5)", etc.
├── "Clear" ghost button
```

## Selection State

| Property | Type | Purpose |
| --- | --- | --- |
| bulkMode | boolean | Whether bulk selection is active |
| selectedIds | Set<number> | Currently selected item IDs |
| actionableIds | number[] | Items that can receive the bulk action |

## Count Badge

Selection count uses: `text-xs font-medium text-primary bg-primary/10 px-2 py-0.5 rounded-full`

## Checkbox Variant

Uses `checkbox()` variant from the variant system (not raw `<input type="checkbox">`).

## Applies To

- InvoiceScreen (bulk obsolete/restore)
- PaymentRequestsScreen (bulk approve in approvals mode)

## Cited By

- c3-1-frontend (Bulk Operations)
