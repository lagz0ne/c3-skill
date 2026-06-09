---
id: c3-103
c3-version: 3
c3-seal: 6e6172bc379f08c6590e2f256d6c4f2f56c79eaf516eeb8b521a39a82a0ed5d0
title: UI Kit (shadcn/radix)
type: component
category: foundation
parent: c3-1
goal: Design system primitives - Button, Dialog, Select, Input, etc. built on Radix UI
uses:
    - ref-responsive-layout
    - ref-ui-patterns
    - ref-variant-system
---

# UI Kit (shadcn/radix)

## Goal

Design system primitives - Button, Dialog, Select, Input, etc. built on Radix UI

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own UI Kit (shadcn/radix) behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep UI Kit (shadcn/radix) decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for UI Kit (shadcn/radix) so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before UI Kit (shadcn/radix) behavior is changed. | ref-responsive-layout |
| Inputs | Accept only the files, commands, data, or calls that belong to UI Kit (shadcn/radix) ownership. | ref-responsive-layout |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-responsive-layout |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-responsive-layout |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks UI Kit (shadcn/radix) to deliver its documented responsibility. | ref-responsive-layout |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-responsive-layout |
| Alternate paths | When a request falls outside UI Kit (shadcn/radix) ownership, hand it to the parent or sibling component. | ref-responsive-layout |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-responsive-layout |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-responsive-layout | ref | Governs UI Kit (shadcn/radix) behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| UI Kit (shadcn/radix) input | IN | Callers must provide context that matches the component goal and parent fit. | c3-1 boundary | c3x lookup plus targeted tests or review. |
| UI Kit (shadcn/radix) output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-1 boundary | c3x check and project test suite. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |
| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| Code, docs, tests, prompts | Goal, Governance, Contract, and Change Safety sections. | Names and framework shape may vary; behavior and boundaries may not. | c3x check, c3x verify, and relevant tests. |

## Architecture Details

## Dependencies

- `@radix-ui/react-*` -- dialog, alert-dialog, select, popover, dropdown-menu, radio-group, separator, tooltip
- `cmdk` -- command palette
- `tailwind-variants` -- `tv()` for variant definitions
- `clsx` + `tailwind-merge` -- `cn()` utility for class merging

## Components

| Component | Radix Primitive | File |
| --- | --- | --- |
| Button | -- | button.tsx |
| Dialog / DialogContent | @radix-ui/react-dialog | dialog.tsx |
| AlertDialog | @radix-ui/react-alert-dialog | alert-dialog.tsx |
| Select | @radix-ui/react-select | select.tsx |
| Popover | @radix-ui/react-popover | popover.tsx |
| DropdownMenu | @radix-ui/react-dropdown-menu | dropdown-menu.tsx |
| Sheet | @radix-ui/react-dialog | sheet.tsx |
| Tooltip | @radix-ui/react-tooltip | tooltip.tsx |
| Command / Combobox | cmdk | command.tsx, combobox.tsx |
| RadioGroup | @radix-ui/react-radio-group | radio-group.tsx |
| Tabs | -- | tabs.tsx |
| Input | -- | input.tsx |
| Textarea | -- | textarea.tsx |
| Separator | @radix-ui/react-separator | separator.tsx |
| Skeleton | -- | skeleton.tsx |
| Sidebar | custom | sidebar.tsx |
| Toast / Toaster | custom hook | toast.tsx, toaster.tsx, use-toast.ts |
| ClientDate | -- | client-date.tsx |
| Field | -- | field.tsx |

## Variant System

All variants defined in `variants.ts` using `tv()` from `tailwind-variants`:

```typescript
import { tv } from 'tailwind-variants'

export const button = tv({
  base: 'inline-flex items-center justify-center rounded-md text-sm font-medium ...',
  variants: {
    variant: {
      default: 'bg-primary text-primary-foreground ...',
      destructive: 'bg-destructive text-destructive-foreground ...',
      outline: 'border border-input bg-background ...',
      ghost: 'hover:bg-accent ...',
      secondary: 'bg-secondary ...',
      // + success, warning, accent, neutral, link, error
    },
    size: { default: 'h-9 px-4 py-2', sm: '...', xs: '...', lg: '...', icon: '...', 'icon-sm': '...', 'icon-xs': '...' },
    loading: { true: 'opacity-50 pointer-events-none' },
  },
})
```

**Other variant definitions in `variants.ts`:** badge, input, inputGroup, select, textarea, fileInput, checkbox, alert, errorAlert, modal (slots), drawer (slots), listItem, confirmButton, radioButton, radioGroup, formActions, passwordStrengthBar, toggleButton, crudTable (slots).

## Class Merging

```typescript
// lib/utils.ts
import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

## Toast System

```typescript
import { useToast } from '@/components/ui/use-toast'

const { toast } = useToast()
toast({ title: 'Success', description: 'Done', variant: 'default' })
```
