---
id: adr-20260107-shadcn-ui-migration
title: Migrate UI from DaisyUI to shadcn/ui
status: implemented
date: 2026-01-07
affects: [c3-1, c3-111, c3-113, c3-115]
---

# Migrate UI from DaisyUI to shadcn/ui

## Status

**Implemented** - 2026-01-07

## Problem

The current UI uses DaisyUI which provides a specific visual style. A visual refresh is desired to adopt the shadcn/ui aesthetic - cleaner, more minimal, neutral tones. This requires replacing the underlying component library and styling approach.

## Decision

Perform a migration from DaisyUI to shadcn/ui with a hybrid styling approach:

1. **Replace DaisyUI with Radix UI primitives** - Only for accessibility-critical components (Dialog, Toast, Select)
2. **Keep tailwind-variants (tv()) for compositions** - Our higher-level components continue using tv()
3. **Use cva only inside shadcn primitives** - Standard shadcn approach for primitives
4. **Adopt Default Zinc theme** - shadcn's default neutral gray tones
5. **Maintain light/dark mode support** - using CSS variables

### Dependencies

**Add:**
- `@radix-ui/react-dialog` - Modal/Drawer (focus trap, a11y)
- `@radix-ui/react-alert-dialog` - ConfirmModal
- `@radix-ui/react-select` - Select dropdowns (keyboard nav, a11y)
- `@radix-ui/react-toast` - Toast notifications (stacking, auto-dismiss)
- `@radix-ui/react-tooltip` - Tooltips (a11y, positioning)
- `@radix-ui/react-slot` - Component composition
- `class-variance-authority` - For shadcn primitives only
- `clsx` - Class name utility

**Remove:**
- `daisyui`

**Keep:**
- `tailwind-variants` - For compositions and business logic variants
- `lucide-react` - shadcn uses this
- `tailwind-merge` - used in `cn()` helper

### Component Strategy

**Use Radix (accessibility required):**
- Dialog/Modal - focus trap, escape, portal
- Sheet/Drawer - side animation, a11y
- AlertDialog - ConfirmModal
- Toast - stacking, auto-dismiss
- Select - keyboard nav, positioning

**Recreate with tv() (styling only):**
- Button - update styling, keep tv()
- Badge - update styling, keep tv()
- Input - update styling, keep tv()
- Alert - update styling, keep tv()
- Fieldset/Legend - simple styled components

### Component Migration

| Current | New |
|---------|-----|
| Modal (custom focus trap) | Dialog (Radix) + modal tv() |
| Drawer (checkbox-based) | Sheet (Radix Dialog variant) + drawer tv() |
| Alert (DaisyUI classes) | Alert with tv() variants |
| Toast (custom) | Toast (Radix) preserving API |
| Button (tv) | Button (tv) with shadcn styling |
| Badge (tv) | Badge (tv) with shadcn styling |

## Rationale

| Considered | Rejected Because |
|------------|------------------|
| Replace all tv() with cva | tv() slots have no cva equivalent; massive rewrite for no benefit |
| Style adoption only (keep DaisyUI) | Doesn't provide Radix accessibility primitives |
| Full cva migration | Loses compound variants power actively used in listItem, passwordStrengthBar |

**Why this hybrid approach:**
- Radix primitives only where accessibility matters (Dialog, Toast, Select)
- Keep tv() for compositions - already works, has slots, compound variants
- Minimal scope change - update styling, not rewrite everything
- Pragmatic engineering over dogmatic adherence to one library

## Affected Layers

| Layer | Document | Change |
|-------|----------|--------|
| Container | c3-1 README | Update dependencies |
| Component | c3-111 Form Patterns | Note shadcn primitives + tv() |
| Component | c3-113 Error Handling | Update Toast reference |
| Component | c3-115 Toast System | Document Radix implementation |

## Code Changes

| File/Directory | Change |
|----------------|--------|
| `package.json` | Add Radix packages, cva, clsx; remove daisyui only |
| `styles.css` | Replace DaisyUI theme with shadcn Zinc CSS variables |
| `components/ui/` | New Radix-based shadcn primitives (Dialog, Toast, Select) |
| `components/ui/variants.ts` | Update tv() styling to shadcn aesthetic |
| `components/Modal.tsx` | Rewrite using Radix Dialog |
| `components/Drawer.tsx` | Rewrite using Radix Sheet |
| `components/Toast.tsx` | Rewrite using Radix Toast (preserve API) |
| `screens/*.tsx` | Replace DaisyUI classes with tv() calls |

## Verification

- [ ] All existing screens render without errors
- [ ] Light/dark mode toggle works
- [ ] Form validation displays correctly
- [ ] Modal/Drawer open/close with focus trap
- [ ] Toast notifications appear and auto-dismiss
- [ ] E2E tests pass (c3-3)
- [ ] No DaisyUI classes remain in codebase
- [ ] TypeScript compiles without errors
