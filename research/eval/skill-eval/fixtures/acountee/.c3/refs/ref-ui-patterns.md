---
id: ref-ui-patterns
c3-seal: cc23897aa0762e75fd962ffc47b901447136b96b4697507f35e45536eed8772b
title: UI Interaction Patterns
type: ref
goal: Standardize shared UI interaction conventions across all screens.
---

# UI Interaction Patterns

## Goal

Standardize shared UI interaction conventions across all screens.

## Choice

- Slide-out drawers for create/edit forms and fixed-position alerts for feedback are standardized; confirmation dialogs use `ConfirmModal` / `ConfirmDrawer` as the target standard (with documented legacy exceptions)
- Shared Radix `Tabs` component with underline indicator for all tabbed navigation (detail panes and page-level), no custom tab CSS
- Consistent formatting: VND via `Intl.NumberFormat`, status badges mapped by meaning (success/warning/destructive), monospace for currency and IDs

## Why

- One component per interaction type (drawer for forms, modal/drawer confirms, alert for feedback) reduces ambiguity and keeps screen behavior predictable
- A single tab component with two usage contexts (detail pane vs page-level) keeps tab behavior and styling identical across the app
- Standardized formatting and status badge mapping ensure financial data and workflow states look the same on every screen

## Current Compliance Snapshot (2026-03-04)

| Pattern | Status | Notes |
| --- | --- | --- |
| Alerts (fixed top-4 right-4 z-50 max-w-md) | PASS | Implemented across all screens |
| Search input (IconSearch + native input with h-8 pl-9) | PASS | Consistent in list/table screens |
| Tabs (TabsList / TabsTrigger / TabsContent) | PASS | Used in detail panes and page-level tabs |
| Workbench sticky period/action bars | PASS | Implemented in cleanup/export tabs with mobile bottom action bar |
| Confirmation dialog standard | PASS | All screens use ConfirmModal or ConfirmDrawer — window.confirm() migration complete |

## Drawer Dialog Pattern

| Rule | Why |
| --- | --- |
| Use slide-out drawers for create/edit forms | Keep users in context while editing |
| Validate required fields before submit | Prevent incomplete submissions |
| Show loading state during submit (Loader2 spinner + disabled button) | Clear feedback for async actions |
| Form wraps DrawerBody + DrawerActions | Consistent anatomy |
| Cancel (outline) + Submit (default) in DrawerActions | Predictable button placement |
| Width: md for simple forms, lg for complex (multi-section) forms | Appropriate space for content |

**Drawer anatomy:**

```tsx
<Drawer isOpen={open} onClose={close} title="Create Entity" width="md">
  <form onSubmit={handleSubmit}>
    <DrawerBody>
      <div className="space-y-5">
        <Field><FieldLabel>Name</FieldLabel><Input /></Field>
        <Separator />
        <Field><FieldLabel>Details</FieldLabel><Textarea /></Field>
      </div>
    </DrawerBody>
    <DrawerActions>
      <button className={button({ variant: 'outline', size: 'sm' })}>Cancel</button>
      <button className={button({ size: 'sm' })} disabled={isSubmitting}>
        {isSubmitting && <Loader2 className="h-4 w-4 animate-spin mr-2" />}
        Save
      </button>
    </DrawerActions>
  </form>
</Drawer>
```

Multi-section forms use `<Separator />` between field groups.

## Alert Feedback Convention

The `Alert` component itself uses `relative w-full` positioning. Each screen wraps it in a fixed container:

```tsx
<div className="fixed top-4 right-4 z-50 max-w-md">
  <Alert type={type} dismissible autoHide={type === 'success' ? 3000 : undefined} onDismiss={...}>
    {message}
  </Alert>
</div>
```

| Rule | Why |
| --- | --- |
| Wrap Alert in fixed top-4 right-4 z-50 max-w-md | Consistent visibility, no layout shift |
| autoHide={3000} for success alerts | Auto-dismiss success |
| No autoHide for error alerts | Errors persist until acknowledged |
| One alert at a time per screen | Avoid stacking |

## Search Input Pattern

All searchable lists use the same input structure:

```tsx
<div className="relative">
  <IconSearch size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground" />
  <input
    className="w-full h-8 pl-9 pr-3 text-sm rounded-md border border-input bg-background"
    placeholder="Search..."
  />
</div>
```

| Rule | Why |
| --- | --- |
| Icon positioned absolutely at left | Consistent visual, no layout complexity |
| Height h-8 | Compact, matches filter inputs |
| Native <input> (not Input component) | Simple search doesn't need error states |

## Loading Spinners

| Context | Spinner | Size |
| --- | --- | --- |
| Inline (button submit) | Loader2 from lucide-react | h-4 w-4 animate-spin mr-2 |
| Page-level (data loading) | Loader2 centered | h-8 w-8 animate-spin text-muted-foreground |
| Refresh button | IconRefresh with conditional spin | className={spinning ? 'animate-spin' : ''} |

## Status Badge Mapping Convention

Every screen with status badges defines a `getStatusBadgeVariant(status)` function that maps status strings to `badge()` variants.

| Status Meaning | Badge Variant |
| --- | --- |
| Active / Approved / Sent / Created | success |
| Pending / In Progress / Imported | warning |
| Completed / Info | info |
| Inactive / Failed / Deleted / Obsolete | destructive |
| Default / Unknown | secondary or outline |

## Formatting Conventions

| Data Type | Formatter | CSS |
| --- | --- | --- |
| Currency (VND) | Intl.NumberFormat('vi-VN', { style: 'currency', currency: 'VND' }) | font-mono |
| Dates | Intl.DateTimeFormat('vi-VN') or <ClientDate> component | Normal text |
| IDs / Account numbers | Raw string | font-mono text-xs |
| Amounts in lists | Formatted currency | .amount class (mono, tabular-nums, right-aligned) |

## Tabs

All tabbed navigation uses the shared Radix `Tabs` component.

**Components:** `Tabs`, `TabsList`, `TabsTrigger`, `TabsContent` (Radix `@radix-ui/react-tabs`)

**Visual style:** Underline indicator — triggers sit in a row with a bottom border, active tab shows a 2px primary-color underline. Text is muted when inactive, foreground when active.

| Rule | Why |
| --- | --- |
| Always use the Radix Tabs component | Single tab mechanism across the app |
| Never build custom tab CSS (pill tabs, segment controls) | Visual consistency |
| TabsContent default is mt-0 — parent containers handle spacing | Avoids double-spacing when tabs split across containers |
| Count badges go inside TabsTrigger as inline <span> | Keeps counts visually attached to the tab |

**Two usage contexts:**

| Context | Where tabs sit | Border handling | Example |
| --- | --- | --- | --- |
| Detail pane | TabsList inside DetailHeader, TabsContent inside DetailContent | TabsList gets border-b-0 (header already has border) | InvoiceScreen, PaymentRequestsScreen |
| Page-level | TabsList inside admin-page-header, TabsContent below | Default border from TabsList | OrganizationScreen (Users / Teams) |

**Detail pane structure:**

```tsx
<Tabs defaultValue="general" className="flex flex-col h-full">
  <DetailHeader>
    <TabsList className="w-full border-b-0">
      <TabsTrigger value="general">General</TabsTrigger>
      <TabsTrigger value="services" className="gap-1.5">
        Services
        <span className="...badge classes...">{count}</span>
      </TabsTrigger>
      <TabsTrigger value="audit">Audit</TabsTrigger>
    </TabsList>
  </DetailHeader>
  <DetailContent className="pt-0 px-3 pb-3 sm:px-4 sm:pb-4">
    <TabsContent value="general">...</TabsContent>
    <TabsContent value="audit" className="pt-3">...</TabsContent>
  </DetailContent>
  <DetailFooter>...</DetailFooter>
</Tabs>
```

Key points:

- `Tabs` wraps the entire detail (`DetailHeader` + `DetailContent` + `DetailFooter`)
- `TabsList` uses `border-b-0` since `DetailHeader` already provides the border
- `DetailContent` uses `pt-0` — no top padding, content sits right below the border
- Tabs whose content lacks its own top spacing (e.g. Audit) use `className="pt-3"`
- No `DetailHeader` identity content (name, status) — the facet grid inside the main tab provides this

**Count badge pattern** (when a tab shows item counts):

```tsx
<TabsTrigger value="users" className="gap-1.5">
  <IconUsers size={14} />
  <span>Users</span>
  <span className={cn(
    "ml-1 inline-flex items-center justify-center min-w-[1.25rem] h-5 px-1.5 text-[0.6875rem] font-semibold rounded-full",
    isActive ? "bg-primary/10 text-primary" : "bg-muted text-muted-foreground"
  )}>
    {count}
  </span>
</TabsTrigger>
```

## MonthPicker (Period Selection)

A `Select` dropdown combining year and month into a single period picker for screens that load data by month. Uses the shared Radix `Select` component.

| Rule | Why |
| --- | --- |
| Use Radix Select component (same as all other selects) | Single select mechanism across the app |
| Value format: ${year}-${month} | Parseable, sortable period key |
| Options: current year down to 5 years back | Reasonable historical range |
| Months capped to current month for current year | Cannot select future periods |
| Height h-8, width w-40 sm:w-48 | Compact, matches filter input sizing |

**Structure:**

```tsx
<Select value={`${year}-${month}`} onValueChange={(v) => { const [y, m] = v.split('-').map(Number); onChange(y, m); }}>
  <SelectTrigger className="w-40 sm:w-48 h-8 text-sm">
    <SelectValue />
  </SelectTrigger>
  <SelectContent>
    {periodOptions.map((opt) => (
      <SelectItem key={opt.value} value={opt.value}>{opt.label}</SelectItem>
    ))}
  </SelectContent>
</Select>
```

**Screens using it:**

- WorkbenchScreen — Invoice Cleanup tab, Export PRs tab

## Sticky Header Bar (Period + Actions)

A bar pinned at the top of scrollable tab content containing a period selector on the left and action buttons on the right. Actions are hidden on mobile and shown via a sticky bottom bar instead (see [ref-responsive-layout](ref-responsive-layout.md) § Sticky Bottom Action Bar).

| Rule | Why |
| --- | --- |
| Use sticky top-0 z-10 bg-background pb-3 | Controls stay visible while scrolling table content |
| Period selector left, actions right: flex items-center justify-between gap-3 | Consistent filter-left, action-right layout |
| Desktop actions: hidden sm:flex | Actions show inline on desktop only |
| Mobile actions via sticky bottom bar | Thumb-friendly action placement on mobile |

**Structure:**

```tsx
<div className="sticky top-0 z-10 bg-background pb-3 flex items-center justify-between gap-3">
  <MonthPicker year={year} month={month} onChange={handlePeriodChange} />
  <div className="hidden sm:flex">{actionBar}</div>
</div>
```

**Screens using it:**

- WorkbenchScreen — Invoice Cleanup tab, Export PRs tab

## Modal & Confirm Dialogs

### Modal

Centered overlay dialog. Used for confirmations, alerts, and non-form content.

**Exported:** `Modal`, `ModalHeader`, `ModalBody`, `ModalActions`.

| Prop | Purpose |
| --- | --- |
| isOpen | Visibility control |
| onClose | Close handler (ESC + backdrop click) |
| title | Header text |
| size | sm / md / lg / xl / full |

### ConfirmModal

Pre-built confirmation dialog with confirm/cancel buttons.

```tsx
<ConfirmModal
  isOpen={open}
  onClose={close}
  onConfirm={handleConfirm}
  title="Delete Item?"
  message="This action cannot be undone."
  confirmText="Delete"
  variant="danger"  // danger | warning | info
  isLoading={deleting}
/>
```

### ConfirmDrawer

Same confirmation pattern but as a slide-out drawer. Used when the confirm action relates to a detail panel context.

```tsx
<ConfirmDrawer
  isOpen={open}
  onClose={close}
  onConfirm={handleConfirm}
  title="Delete Payment Method"
  message="Are you sure?"
  confirmText="Delete"
  variant="danger"
  isLoading={deleting}
/>
```

Exported from the Drawer module.

### Confirmation Standard

All destructive actions use `ConfirmModal` (centered dialog) or `ConfirmDrawer` (slide-out) — no `window.confirm()` remains.

| Rule | Why |
| --- | --- |
| New delete/destructive confirmations MUST use ConfirmModal or ConfirmDrawer | Single confirmation UX across the app |
| PaymentRequestsScreen uses ConfirmDrawer for all destructive actions (delete, reject, recall, unapprove, unlink) | Context-appropriate slide-out confirmation |

## Applies To

- InvoiceScreen (detail tabs)
- PaymentRequestsScreen (detail tabs, ConfirmDrawer for destructive actions)
- OrganizationScreen (page-level tabs, ConfirmModal for delete)
- UserManagementScreen (ConfirmModal for delete)
- TeamManagementScreen (ConfirmModal for delete)
- ApprovalConfigScreen
- PaymentsScreen (search input, alerts, formatting)
- AuditLogScreen
- NotificationLogScreen
- WorkbenchScreen (tabs, MonthPicker, sticky header bar, alerts, formatting)
- All future screens

## Cited By

- `ref-detail-content-strategy` (tab content strategy)
- `ref-variant-system` (button/badge variants)
- `ref-form-patterns` (drawer dialog pattern)
