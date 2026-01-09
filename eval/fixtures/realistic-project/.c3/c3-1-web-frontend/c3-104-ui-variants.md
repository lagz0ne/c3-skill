---
id: c3-104
c3-version: 3
title: UI Variants
type: component
category: foundation
parent: c3-1
summary: tailwind-variants definitions for consistent component styling with DaisyUI
---

# UI Variants

Provides type-safe styling definitions using tailwind-variants (tv) for modals, alerts, buttons, inputs, and other UI primitives. Integrates with DaisyUI component classes.

## Contract

| Provides | Expects |
|----------|---------|
| modal() with slots | size, actionsAlign variants |
| alert() | type (success, error, warning, info) |
| input(), textarea(), select() | error, disabled states |
| button() | variant, size, loading |
| table-related variants | Consistent table styling |

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Invalid variant value | TypeScript compile error |
| Missing slot | Slot function returns undefined |
| Compound variants | Multiple conditions combine classes |
| Default variants | Applied when not specified |

## Testing

| Scenario | Verifies |
|----------|----------|
| Variant application | modal({size: 'lg'}) includes max-w-2xl |
| Slot access | modal().box() returns correct classes |
| Type safety | Wrong variant name is compile error |
| Default handling | modal() uses defaultVariants |

## References

- `apps/start/src/components/ui/` - Variant definitions and UI primitives
