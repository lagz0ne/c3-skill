# Plan: Migrate UI from DaisyUI to shadcn/ui

**ADR:** adr-20260107-shadcn-ui-migration
**Status:** Ready to execute
**Estimated Effort:** 4-5 developer-days
**Risk Level:** Medium (reduced scope, keep tv())

---

## Pre-Execution Checklist

- [ ] Create git worktree: `git worktree add ../acountee-shadcn-migration feature/shadcn-ui-migration`
- [ ] Verify current E2E tests pass: `bun run --cwd apps/e2e test`
- [ ] Take screenshots for before/after comparison

---

## Key Simplifications

| Original Plan | Simplified |
|---------------|------------|
| Convert lofi OKLCH theme to shadcn HSL | Use default Zinc theme as-is |
| Replace tailwind-variants with cva | Keep tv() for compositions, cva only in shadcn primitives |
| Install all shadcn components | Only install Radix where accessibility matters |
| Recreate all components | Simple components recreated with tv() |

---

## Component Strategy

### Use Radix Primitives (Accessibility Required)

These components have complex accessibility requirements (focus trap, keyboard nav, ARIA):

| Component | Radix Package | Why Radix |
|-----------|---------------|-----------|
| Dialog/Modal | @radix-ui/react-dialog | Focus trap, escape, portal, aria-modal |
| Sheet/Drawer | @radix-ui/react-dialog | Same as Dialog, side animation |
| AlertDialog | @radix-ui/react-alert-dialog | Focus trap + alert semantics |
| Toast | @radix-ui/react-toast | Focus, stacking, auto-dismiss, viewport |
| Select | @radix-ui/react-select | Keyboard nav, positioning, typeahead |

### Recreate with tv() (Styling Only)

These components are primarily styling with minimal behavior:

| Component | Approach |
|-----------|----------|
| Button | tv() with variants (already exists, update styling) |
| Badge | tv() with variants (already exists, update styling) |
| Input | tv() with error state |
| Label | Simple styled component |
| Alert | tv() with type variants (already exists) |
| Fieldset/Legend | Simple styled components |

### Keep Existing (Just Update Classes)

| Component | Change |
|-----------|--------|
| Form.tsx | Update field styling |
| FormComponents.tsx | Update DaisyUI classes to Tailwind |
| ErrorBoundary.tsx | Update button/alert classes |

---

## Audit Summary

### DaisyUI Class Usage by File

| File | Class Count | Key Patterns |
|------|-------------|--------------|
| `_authed.tsx` | 28 | `btn`, `btn-ghost`, `btn-square`, `loading`, `tooltip` |
| `InvoiceScreen.tsx` | 47 | `btn-*`, `badge-*`, `fieldset`, `file-input`, `select`, `alert` |
| `PaymentRequestsScreen.tsx` | 52 | `btn-*`, `badge-*`, `fieldset`, `file-input` |
| `AdminScreen.tsx` | 18 | `btn-*`, `badge-*`, `input`, `loading`, `alert` |
| `PaymentsScreen.tsx` | 3 | `btn-ghost`, `btn-sm` |
| `FormComponents.tsx` | 22 | `fieldset`, `floating-label`, `join`, `select` |
| `Modal.tsx` | 8 | `btn`, `btn-ghost`, `alert-*` |
| `Drawer.tsx` | 6 | `btn-ghost`, `btn-primary` |
| `Toast.tsx` | 6 | `alert-*` |
| `ErrorBoundary.tsx` | 5 | `alert-warning`, `btn-*` |
| **Total** | **~200** | |

---

## Design Decisions

### DD-1: Use Default Zinc Theme

**Decision:** Use shadcn's default Zinc theme without customization.

**Rationale:** Goal is visual refresh to shadcn aesthetic, not preserving current look.

### DD-2: Keep tailwind-variants for ALL Our Components

**Decision:**
- **Our components (Button, Badge, Input, Alert, etc.) use tv()** - update styling, keep the library
- **Radix wrapper components (Dialog, Toast, Select) use cva internally** - copy-pasted from shadcn/ui, we don't modify these

**Boundary is clear:**
| Component Type | Library | Wrapped with tv()? |
|----------------|---------|-------------------|
| Button, Badge, Input, Alert, etc. | tv() | N/A - these ARE tv() |
| Dialog, Sheet, Toast, Select, Tooltip | cva (inside shadcn) | NO - use directly as-is |

**Key clarification:** We do NOT wrap shadcn Radix components with additional tv(). We use them directly:

```tsx
// ❌ WRONG - don't wrap shadcn components
const dialogStyles = tv({ ... })
<Dialog className={dialogStyles()} />

// ✅ CORRECT - use shadcn components directly
<Dialog>
  <DialogContent size="lg">...</DialogContent>
</Dialog>

// ✅ CORRECT - our variants remain tv()
<button className={button({ variant: 'ghost', size: 'sm' })}>
```

### DD-3: Minimal Radix Installation

**Decision:** Only install Radix packages for components requiring accessibility primitives.

```bash
# Only these Radix packages
@radix-ui/react-dialog      # Modal, Sheet
@radix-ui/react-alert-dialog # ConfirmModal
@radix-ui/react-toast       # Toast system
@radix-ui/react-select      # Select dropdowns
@radix-ui/react-slot        # Component composition
```

### DD-4: Loading Spinner

**Decision:** Replace DaisyUI `loading loading-spinner` with Lucide `Loader2` + animation.

```tsx
// Before
<span className="loading loading-spinner loading-sm"></span>

// After
<Loader2 className="h-4 w-4 animate-spin" />
```

### DD-5: Toast API Compatibility

**Decision:** Preserve existing `toast.success()` API while using Radix internally.

### DD-6: Input Group (join pattern)

**Decision:** Custom CSS utility class replaces DaisyUI join/join-item.

```css
.input-group {
  display: flex;
}
.input-group > *:not(:first-child):not(:last-child) { border-radius: 0; }
.input-group > *:first-child { border-top-right-radius: 0; border-bottom-right-radius: 0; }
.input-group > *:last-child { border-top-left-radius: 0; border-bottom-left-radius: 0; }
.input-group > *:not(:last-child) { border-right: none; }
```

### DD-7: Floating Label → Standard Label

**Decision:** Replace floating labels with label-above-input (shadcn standard).

### DD-8: Tooltip Strategy

**Decision:** Install Radix Tooltip and implement properly.

**Rationale:**
- Native `title` is unreliable and has accessibility issues
- `_authed.tsx` uses tooltips for navigation hints - these should be accessible
- Adding one more Radix package is minimal overhead

**Add to dependencies:**
```bash
@radix-ui/react-tooltip
```

**Implementation:**
```tsx
// Before
<button className="tooltip tooltip-right" data-tip="Settings">

// After
<Tooltip>
  <TooltipTrigger asChild>
    <button>...</button>
  </TooltipTrigger>
  <TooltipContent side="right">Settings</TooltipContent>
</Tooltip>
```

**File:** Create `components/ui/tooltip.tsx` (copy from shadcn/ui).

### DD-9: Radio Button Groups (join pattern)

**Decision:** Update existing `radioButton` and `radioGroup` tv() variants in `components/ui/variants.ts`.

**Current location:** `apps/start/src/components/ui/variants.ts` (lines 277-305)

**Current code uses:**
- `btn join-item` → DaisyUI class
- `join join-horizontal` → DaisyUI class

**Update to:**
```typescript
// In components/ui/variants.ts (update existing)
export const radioGroup = tv({
  base: 'inline-flex rounded-md border border-input',
  variants: {
    orientation: {
      horizontal: 'flex-row',
      vertical: 'flex-col',
    },
  },
});

export const radioButton = tv({
  base: 'px-3 py-2 text-sm font-medium transition-colors first:rounded-l-md last:rounded-r-md border-r border-input last:border-r-0',
  variants: {
    selected: {
      true: 'bg-primary text-primary-foreground',
      false: 'bg-background hover:bg-muted',
    },
  },
});
```

**Used in:** Search for `radioGroup` and `radioButton` imports to find usage locations.

---

## Phase 1: Foundation Setup (Day 1)

### 1.1 Create Worktree

```bash
git worktree add ../acountee-shadcn-migration feature/shadcn-ui-migration
cd ../acountee-shadcn-migration/apps/start
```

### 1.2 Update Dependencies

```bash
# Remove DaisyUI only (keep tailwind-variants!)
bun remove daisyui

# Add Radix + cva for shadcn primitives
bun add @radix-ui/react-dialog @radix-ui/react-alert-dialog \
  @radix-ui/react-toast @radix-ui/react-select @radix-ui/react-slot \
  @radix-ui/react-tooltip \
  class-variance-authority clsx
```

### 1.3 Create `cn()` Utility

**File:** `apps/start/src/lib/utils.ts`
```typescript
import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

### 1.4 Update CSS with shadcn Theme Variables

**File:** `apps/start/src/styles.css`

Replace DaisyUI theme blocks with shadcn Zinc theme:

```css
@import "tailwindcss";

/* === SHADCN ZINC THEME === */
:root {
  --background: 0 0% 100%;
  --foreground: 240 10% 3.9%;
  --card: 0 0% 100%;
  --card-foreground: 240 10% 3.9%;
  --popover: 0 0% 100%;
  --popover-foreground: 240 10% 3.9%;
  --primary: 240 5.9% 10%;
  --primary-foreground: 0 0% 98%;
  --secondary: 240 4.8% 95.9%;
  --secondary-foreground: 240 5.9% 10%;
  --muted: 240 4.8% 95.9%;
  --muted-foreground: 240 3.8% 46.1%;
  --accent: 240 4.8% 95.9%;
  --accent-foreground: 240 5.9% 10%;
  --destructive: 0 84.2% 60.2%;
  --destructive-foreground: 0 0% 98%;
  --success: 142 76% 36%;
  --success-foreground: 0 0% 98%;
  --warning: 38 92% 50%;
  --warning-foreground: 0 0% 98%;
  --info: 199 89% 48%;
  --info-foreground: 0 0% 98%;
  --border: 240 5.9% 90%;
  --input: 240 5.9% 90%;
  --ring: 240 5.9% 10%;
  --radius: 0.5rem;
}

.dark {
  --background: 240 10% 3.9%;
  --foreground: 0 0% 98%;
  --card: 240 10% 3.9%;
  --card-foreground: 0 0% 98%;
  --popover: 240 10% 3.9%;
  --popover-foreground: 0 0% 98%;
  --primary: 0 0% 98%;
  --primary-foreground: 240 5.9% 10%;
  --secondary: 240 3.7% 15.9%;
  --secondary-foreground: 0 0% 98%;
  --muted: 240 3.7% 15.9%;
  --muted-foreground: 240 5% 64.9%;
  --accent: 240 3.7% 15.9%;
  --accent-foreground: 0 0% 98%;
  --destructive: 0 62.8% 30.6%;
  --destructive-foreground: 0 0% 98%;
  --success: 142 76% 36%;
  --success-foreground: 0 0% 98%;
  --warning: 38 92% 50%;
  --warning-foreground: 0 0% 98%;
  --info: 199 89% 48%;
  --info-foreground: 0 0% 98%;
  --border: 240 3.7% 15.9%;
  --input: 240 3.7% 15.9%;
  --ring: 240 4.9% 83.9%;
}

/* Base layer */
* { border-color: hsl(var(--border)); }
body {
  background-color: hsl(var(--background));
  color: hsl(var(--foreground));
}
```

### 1.5 Keep Custom CSS Systems

Preserve but update variable references:

```css
/* List item system - update variable names */
.list-item-hover:hover {
  background-color: hsl(var(--muted));
}
.list-item-selected {
  background-color: hsl(var(--muted));
  border-left-color: hsl(var(--border));
}
.list-item-selected.list-item-status-inprogress {
  border-left-color: hsl(var(--success));
}
/* ... etc */

/* Detail section system - update variable names */
.detail-section {
  border-bottom: 2px solid hsl(var(--border));
}
/* ... etc */

/* Input group (replaces join/join-item) */
.input-group { display: flex; }
.input-group > *:not(:first-child):not(:last-child) { border-radius: 0; }
.input-group > *:first-child { border-top-right-radius: 0; border-bottom-right-radius: 0; }
.input-group > *:last-child { border-top-left-radius: 0; border-bottom-left-radius: 0; }
.input-group > *:not(:last-child) { border-right: none; }
```

### 1.6 Verification Checkpoint

```bash
bun run dev  # Should start (will have broken styles)
bunx @typescript/native-preview --noEmit  # No type errors
```

---

## Phase 2: Update tv() Variants (Day 1)

### 2.1 Update `components/ui/variants.ts`

Update existing tv() definitions to use shadcn styling.

**Note:** Some variants need to be CREATED (not updated):
- `badge` - NEW (currently uses inline DaisyUI classes)
- `input` - UPDATE existing
- `alert` - UPDATE existing
- `button` - UPDATE existing
- `modal` - UPDATE existing
- `drawer` - UPDATE existing

```typescript
import { tv, type VariantProps } from 'tailwind-variants';

// Button - update to shadcn styling
export const button = tv({
  base: 'inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50',
  variants: {
    variant: {
      default: 'bg-primary text-primary-foreground shadow hover:bg-primary/90',
      destructive: 'bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/90',
      outline: 'border border-input bg-background shadow-sm hover:bg-accent hover:text-accent-foreground',
      secondary: 'bg-secondary text-secondary-foreground shadow-sm hover:bg-secondary/80',
      ghost: 'hover:bg-accent hover:text-accent-foreground',
      link: 'text-primary underline-offset-4 hover:underline',
      success: 'bg-success text-success-foreground shadow-sm hover:bg-success/90',
      warning: 'bg-warning text-warning-foreground shadow-sm hover:bg-warning/90',
    },
    size: {
      default: 'h-9 px-4 py-2',
      sm: 'h-8 rounded-md px-3 text-xs',
      xs: 'h-7 rounded-md px-2 text-xs',
      lg: 'h-10 rounded-md px-8',
      icon: 'h-9 w-9',
      'icon-sm': 'h-8 w-8',
      'icon-xs': 'h-7 w-7',
    },
  },
  defaultVariants: {
    variant: 'default',
    size: 'default',
  },
});

// Badge - update to shadcn styling
export const badge = tv({
  base: 'inline-flex items-center rounded-md border px-2.5 py-0.5 text-xs font-semibold transition-colors',
  variants: {
    variant: {
      default: 'border-transparent bg-primary text-primary-foreground',
      secondary: 'border-transparent bg-secondary text-secondary-foreground',
      destructive: 'border-transparent bg-destructive text-destructive-foreground',
      outline: 'text-foreground',
      success: 'border-transparent bg-success text-success-foreground',
      warning: 'border-transparent bg-warning text-warning-foreground',
      info: 'border-transparent bg-info text-info-foreground',
      neutral: 'border-transparent bg-muted text-muted-foreground',
    },
    size: {
      default: 'px-2.5 py-0.5 text-xs',
      sm: 'px-2 py-0.5 text-[0.65rem]',
      xs: 'px-1.5 py-0 text-[0.6rem]',
    },
  },
  defaultVariants: {
    variant: 'default',
    size: 'default',
  },
});

// Input - update to shadcn styling
export const input = tv({
  base: 'flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
  variants: {
    hasError: {
      true: 'border-destructive focus-visible:ring-destructive',
      false: '',
    },
  },
  defaultVariants: {
    hasError: false,
  },
});

// Alert - update to shadcn styling
export const alert = tv({
  base: 'relative w-full rounded-lg border px-4 py-3 text-sm',
  variants: {
    type: {
      default: 'bg-background text-foreground',
      success: 'border-success/50 text-success bg-success/10',
      error: 'border-destructive/50 text-destructive bg-destructive/10',
      warning: 'border-warning/50 text-warning bg-warning/10',
      info: 'border-info/50 text-info bg-info/10',
    },
  },
  defaultVariants: {
    type: 'default',
  },
});

// Keep modal/drawer slots for structure, update styling
export const modal = tv({
  slots: {
    backdrop: 'fixed inset-0 z-50 bg-black/80 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
    box: 'fixed left-[50%] top-[50%] z-50 grid w-full translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-6 shadow-lg duration-200 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%] data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%] sm:rounded-lg',
    header: 'flex flex-col space-y-1.5 text-center sm:text-left',
    title: 'text-lg font-semibold leading-none tracking-tight',
    subtitle: 'text-sm text-muted-foreground',
    closeButton: 'absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none',
    body: 'py-4',
    actions: 'flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2',
  },
  variants: {
    size: {
      sm: { box: 'max-w-md' },
      md: { box: 'max-w-lg' },
      lg: { box: 'max-w-2xl' },
      xl: { box: 'max-w-5xl' },
      full: { box: 'max-w-7xl' },
    },
  },
  defaultVariants: {
    size: 'lg',
  },
});

// Drawer/Sheet - update with shadcn sheet styling
export const drawer = tv({
  slots: {
    overlay: 'fixed inset-0 z-50 bg-black/80 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
    panel: 'fixed z-50 gap-4 bg-background p-6 shadow-lg transition ease-in-out data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:duration-300 data-[state=open]:duration-500',
    header: 'flex flex-col space-y-2',
    title: 'text-lg font-semibold text-foreground',
    subtitle: 'text-sm text-muted-foreground',
    closeButton: 'absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100',
    body: 'py-4',
    actions: 'flex gap-3',
  },
  variants: {
    side: {
      left: { panel: 'inset-y-0 left-0 h-full border-r data-[state=closed]:slide-out-to-left data-[state=open]:slide-in-from-left' },
      right: { panel: 'inset-y-0 right-0 h-full border-l data-[state=closed]:slide-out-to-right data-[state=open]:slide-in-from-right' },
    },
    width: {
      sm: { panel: 'w-3/4 sm:max-w-sm' },
      md: { panel: 'w-3/4 sm:max-w-md' },
      lg: { panel: 'w-3/4 sm:max-w-lg' },
      xl: { panel: 'w-3/4 sm:max-w-xl' },
      full: { panel: 'w-full' },
    },
  },
  defaultVariants: {
    side: 'right',
    width: 'md',
  },
});

// ... update other existing tv() variants similarly
```

---

## Phase 3: Create Radix-based Primitives (Day 2)

Create these in `apps/start/src/components/ui/`. Copy from shadcn/ui and adjust.

### 3.1 Dialog (`dialog.tsx`)

```typescript
// Copy from: https://ui.shadcn.com/docs/components/dialog
// Exports: Dialog, DialogTrigger, DialogContent, DialogHeader,
//          DialogFooter, DialogTitle, DialogDescription, DialogClose
import * as DialogPrimitive from "@radix-ui/react-dialog"
// ... standard shadcn dialog implementation with cva
```

### 3.2 Sheet (`sheet.tsx`)

```typescript
// Copy from: https://ui.shadcn.com/docs/components/sheet
// Exports: Sheet, SheetTrigger, SheetContent, SheetHeader,
//          SheetFooter, SheetTitle, SheetDescription, SheetClose
// Uses DialogPrimitive with side variants (left, right, top, bottom)
```

### 3.3 AlertDialog (`alert-dialog.tsx`)

```typescript
// Copy from: https://ui.shadcn.com/docs/components/alert-dialog
// Exports: AlertDialog, AlertDialogTrigger, AlertDialogContent,
//          AlertDialogHeader, AlertDialogFooter, AlertDialogTitle,
//          AlertDialogDescription, AlertDialogAction, AlertDialogCancel
```

### 3.4 Toast (`toast.tsx`, `toaster.tsx`, `use-toast.ts`)

```typescript
// toast.tsx - Copy from shadcn
// toaster.tsx - Copy from shadcn
// use-toast.ts - CUSTOMIZE to preserve existing API:

export function useToast() {
  const { toast: radixToast } = useRadixToast()

  return {
    toast: {
      success: (message: string) => radixToast({ description: message, variant: "success" }),
      error: (message: string) => radixToast({ description: message, variant: "destructive" }),
      warning: (message: string) => radixToast({ description: message, variant: "warning" }),
      info: (message: string) => radixToast({ description: message, variant: "default" }),
    }
  }
}
```

### 3.5 Select (`select.tsx`)

```typescript
// Copy from: https://ui.shadcn.com/docs/components/select
// Exports: Select, SelectTrigger, SelectValue, SelectContent,
//          SelectItem, SelectGroup, SelectLabel
```

### 3.6 Tooltip (`tooltip.tsx`)

```typescript
// Copy from: https://ui.shadcn.com/docs/components/tooltip
// Exports: Tooltip, TooltipTrigger, TooltipContent, TooltipProvider
```

**Important:** Radix Tooltip requires wrapping the app in `<TooltipProvider>`.

Add to `_authed.tsx` layout:
```tsx
import { TooltipProvider } from '@/components/ui/tooltip'

export function AuthedLayout({ children }) {
  return (
    <TooltipProvider>
      {/* existing layout */}
    </TooltipProvider>
  )
}
```

### 3.7 Verification Checkpoint

```bash
# After creating all primitives, verify imports work:
bunx @typescript/native-preview --noEmit
```

---

## Phase 4: Migrate Components (Day 2-3)

### 4.1 Modal.tsx → Dialog

Replace custom Modal with Radix Dialog, use modal tv() for additional styling.

### 4.2 Drawer.tsx → Sheet

Replace with Radix Sheet, use drawer tv() for styling.

### 4.3 Toast.tsx

Replace with Radix Toast, preserve API.

### 4.4 FormComponents.tsx

- Replace `floating-label` with standard Label above Input
- Replace `join`/`join-item` with `.input-group`
- Use updated input tv()
- Keep `fieldset` element with updated styling

### 4.5 prHelpers.ts

Update `getStatusBadgeClass` to return badge variant:

```typescript
// Before
case 'draft': return 'badge-neutral';

// After
case 'draft': return 'neutral' as const;
```

---

## Phase 5: Migrate Screens (Day 3-4)

### 5.1 DaisyUI → Tailwind/tv() Mapping

| DaisyUI | Replacement |
|---------|-------------|
| `btn btn-primary` | `className={button()}` |
| `btn btn-ghost` | `className={button({ variant: 'ghost' })}` |
| `btn btn-ghost btn-sm` | `className={button({ variant: 'ghost', size: 'sm' })}` |
| `btn btn-ghost btn-square` | `className={button({ variant: 'ghost', size: 'icon' })}` |
| `btn btn-error` | `className={button({ variant: 'destructive' })}` |
| `btn btn-success` | `className={button({ variant: 'success' })}` |
| `badge badge-success` | `className={badge({ variant: 'success' })}` |
| `badge badge-xs` | `className={badge({ size: 'xs' })}` |
| `loading loading-spinner` | `<Loader2 className="animate-spin" />` |
| `alert alert-warning` | `className={alert({ type: 'warning' })}` |
| `input input-bordered` | `className={input()}` |
| `select select-bordered` | `<Select>` component |
| `fieldset` | `<fieldset className="space-y-2">` |
| `fieldset-legend` | `<legend className="text-sm font-medium mb-2">` |
| `tooltip tooltip-right data-tip="X"` | `<Tooltip><TooltipTrigger>...</TooltipTrigger><TooltipContent side="right">X</TooltipContent></Tooltip>` |
| `join join-horizontal` | `className={radioGroup()}` |
| `btn join-item` | `className={radioButton({ selected })}` |

**Conditional tooltip pattern** (used in `_authed.tsx` sidebar):

```tsx
// Before (DaisyUI)
<button
  className={`${!expanded && !isMobile ? 'tooltip tooltip-right' : ''}`}
  data-tip={!expanded && !isMobile ? "Invoices" : undefined}
>
  <Icon />
</button>

// After (Radix) - extract tooltip wrapper
const NavButton = ({ label, expanded, isMobile, children }) => {
  if (!expanded && !isMobile) {
    return (
      <Tooltip>
        <TooltipTrigger asChild>{children}</TooltipTrigger>
        <TooltipContent side="right">{label}</TooltipContent>
      </Tooltip>
    );
  }
  return children;
};

// Usage
<NavButton label="Invoices" expanded={expanded} isMobile={isMobile}>
  <button><Icon /></button>
</NavButton>
```

### 5.2 Files to Update

| File | Changes |
|------|---------|
| `_authed.tsx` | Buttons, loading spinners |
| `InvoiceScreen.tsx` | Buttons, badges, fieldsets, selects |
| `PaymentRequestsScreen.tsx` | Buttons, badges, fieldsets |
| `AdminScreen.tsx` | Buttons, badges, input, loading |
| `PaymentsScreen.tsx` | Buttons |
| `ErrorBoundary.tsx` | Alert, buttons |

---

## Phase 6: Cleanup & Verification (Day 4)

### 6.1 Search for Remaining DaisyUI

```bash
grep -r "btn-\|modal-\|alert-\|badge-\|input-\|select-\|loading " apps/start/src --include="*.tsx"
# Should return 0 matches
```

### 6.2 Remove DaisyUI Plugin References

Remove from styles.css:
```css
@plugin "daisyui" { ... }
@plugin "daisyui/theme" { ... }
```

### 6.3 TypeScript Check

```bash
bunx @typescript/native-preview --noEmit
```

### 6.4 Manual Testing Checklist

- [ ] Login screen renders
- [ ] Sidebar navigation works
- [ ] Theme toggle (light/dark) works
- [ ] Invoice list displays
- [ ] Invoice drawer opens
- [ ] Invoice upload modal works
- [ ] PR list displays
- [ ] PR creation form works
- [ ] Toast notifications work
- [ ] Loading spinners appear
- [ ] Mobile responsive

### 6.5 E2E Tests

**Known selector issues:** E2E tests use some class-based selectors that need updating:

| File | Current Selector | Fix |
|------|------------------|-----|
| `smoke.spec.ts:139` | `.badge` | Change to `[data-testid="status-badge"]` or add `data-testid` |
| `admin.spec.ts:109` | `.text-error` | Change to `[class*="destructive"]` or add `data-testid="error-text"` |

**Pre-migration step:**
```bash
# Search for class-based selectors in E2E tests
grep -r "\.badge\|\.btn\|\.alert\|\.text-error" apps/e2e --include="*.ts"
# Update any matches to use data-testid or class-contains selectors
```

**Post-migration:**
```bash
# Run E2E tests
bun run --cwd apps/e2e test
```

**If tests fail:**
1. Check if failure is selector-related (update to data-testid)
2. Visual regression is expected - update screenshots after migration
3. Functional failures need debugging - check component behavior

**Test files:** `apps/e2e/tests/*.spec.ts`

---

## Phase 7: Documentation (Day 4-5)

### 7.1 Update C3 Docs

- `c3-111-form-patterns.md` - Note shadcn primitives + tv() for compositions
- `c3-113-error-handling.md` - Update Toast reference
- `c3-115-toast-system.md` - Document Radix implementation

### 7.2 Update ADR Status

```yaml
status: implemented
```

---

## Rollback Strategy

### Checkpoint Commits

1. `feat(ui): phase 1 - foundation and theme`
2. `feat(ui): phase 2 - update tv() variants`
3. `feat(ui): phase 3 - radix primitives`
4. `feat(ui): phase 4-5 - component and screen migration`
5. `feat(ui): phase 6-7 - cleanup and docs`

### Emergency Rollback

```bash
git checkout main -- apps/start/src/
bun add daisyui
```

---

## Success Criteria

1. ✅ No DaisyUI classes in codebase
2. ✅ All E2E tests pass
3. ✅ TypeScript compiles without errors
4. ✅ shadcn Zinc theme applied
5. ✅ Light/dark mode works
6. ✅ tv() used for compositions
7. ✅ Radix used only where needed
8. ✅ All interactive components functional
