---
id: ref-form-patterns
c3-seal: e7cd331d53efe340a81065b17bb2f88afe3aca50fe349d84be53d09018b7e114
title: Form Patterns
type: ref
goal: Standardize how forms are built across all screens. Every form — whether a simple 2-field CRUD dialog or a complex multi-section editor — follows the same structural conventions, state management approach, validation strategy, and error display pattern.
---

# Form Patterns

## Goal

Standardize how forms are built across all screens. Every form — whether a simple 2-field CRUD dialog or a complex multi-section editor — follows the same structural conventions, state management approach, validation strategy, and error display pattern.

## Choice

- Forms use `<Drawer>` as the primary container and controlled inputs with explicit `isSubmitting` handling
- Validation baseline in current FE is HTML/native required checks + server validation; Zod is a target convention for new/complex forms
- Create/edit sharing and shared form extraction are preferred patterns, with known legacy duplication still present in admin screens

## Why

- Uniform drawer structure and controlled inputs keep forms predictable to build and review
- Explicitly distinguishing current baseline from target architecture keeps docs truthful while preserving migration direction
- Shared create/edit components reduce drift once extracted

## Current Implementation Snapshot (2026-02-26)

| Pattern Area | Status | Notes |
| --- | --- | --- |
| Drawer container + form semantics | PASS | Screen forms consistently use Drawer + <form onSubmit> |
| Submit loading state (Loader2, disable controls) | PASS | Implemented across admin/payment/workbench forms |
| Zod validation in screen forms | WARN | No zod usage in apps/start/src/screens; validation mostly native + server-side |
| Shared user/team CRUD form extraction | WARN | Dialog implementations are duplicated across Organization/UserManagement/TeamManagement |

## Container: Always a Drawer

All forms live inside a `<Drawer>`. No inline forms, no modals-as-forms.

| Rule | Why |
| --- | --- |
| Use <Drawer> as the form container | Consistent slide-out UX, keeps user in context |
| width="md" for simple forms (1-4 fields) | Appropriate space |
| width="lg" for complex forms (sections, dynamic lists, file uploads) | Needs the room |
| Title describes the action: "Create {Entity}" / "Edit {Entity}" | Clear intent |

## Form Element

Always wrap content in a `<form>` tag with `onSubmit`.

```tsx
<Drawer isOpen={open} onClose={close} title="Create Team" width="md">
  <form onSubmit={handleSubmit}>
    <DrawerBody>
      {/* fields */}
    </DrawerBody>
    <DrawerActions>
      {/* buttons */}
    </DrawerActions>
  </form>
</Drawer>
```

| Rule | Why |
| --- | --- |
| Always use <form onSubmit>, never button onClick for submission | Native form semantics, Enter key support, accessibility |
| e.preventDefault() in handler | Prevent page reload |
| Never skip the <form> tag (even for "non-form" drawers with actions) | Consistency |

## Inputs: Controlled

All inputs are controlled — `value` + `onChange` bound to state.

| Rule | Why |
| --- | --- |
| Use value and onChange on all inputs | Predictable state, easy reset, enables validation feedback |
| Never use defaultValue for form inputs | Avoids split between state and DOM, prevents stale data bugs |
| Disable all inputs during submission: disabled={isSubmitting} | Prevent double-submit |

## State: Single Form Object

Form state lives in a single object, not scattered `useState` calls.

```tsx
const [form, setForm] = useState({
  name: '',
  description: '',
  capabilities: [] as string[],
})
const [isSubmitting, setIsSubmitting] = useState(false)

const update = <K extends keyof typeof form>(key: K, value: typeof form[K]) =>
  setForm(prev => ({ ...prev, [key]: value }))
```

| Rule | Why |
| --- | --- |
| One useState for the form data object | Single source of truth, easy to reset or pre-fill |
| Separate isSubmitting boolean | Submission state is not form data |
| update() helper for field changes | Avoids verbose spread syntax in every onChange |
| Pre-fill for edit mode: useState(editEntity ?? defaults) | One initialization path |

**Reset on close:** Set form back to defaults when drawer closes:

```tsx
onClose={() => {
  setForm(defaults)
  setErrors({})
  close()
}}
```

## Validation Strategy (Current + Target)

Current FE behavior:

- Screen forms primarily use required fields and server-returned errors.
- Zod schemas are not yet used in the screen form layer.

Target pattern for new/refactored forms:

- Validate with a Zod schema on submit for deterministic field-level error mapping.

```tsx
const schema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string().optional(),
})

const [errors, setErrors] = useState<Record<string, string>>({})

const handleSubmit = (e: FormEvent) => {
  e.preventDefault()
  const result = schema.safeParse(form)
  if (!result.success) {
    setErrors(Object.fromEntries(
      result.error.issues.map(i => [i.path[0], i.message])
    ))
    return
  }
  setErrors({})
  setIsSubmitting(true)
  // submit result.data
}
```

| Rule | Why |
| --- | --- |
| Current baseline: required fields + server validation is acceptable for existing forms | Matches implemented behavior, avoids false documentation claims |
| Target: define a Zod schema for new/refactored forms | Single validation source, type-safe |
| Validate on submit (not on every keystroke) | Less noise, simpler |
| Clear errors on successful validation | Don't show stale errors |
| Clear individual field error on change: update(key, value) + setErrors(prev => { const next = {...prev}; delete next[key]; return next }) | Immediate feedback when user fixes a field |
| Reuse flow schemas when the form maps to a server flow | DRY, consistent with backend |

### Field-level error display

Show errors below the field using a consistent pattern:

```tsx
<Field className="gap-2">
  <FieldLabel htmlFor="name">Name</FieldLabel>
  <Input
    id="name"
    value={form.name}
    onChange={e => update('name', e.target.value)}
    aria-invalid={!!errors.name}
  />
  {errors.name && (
    <p className="text-xs text-destructive">{errors.name}</p>
  )}
</Field>
```

| Rule | Why |
| --- | --- |
| Error text: text-xs text-destructive | Visible but not overwhelming |
| aria-invalid on the input | Accessibility |
| Error below the field, not above | Standard placement |

## Field Layout

### Simple forms (1 section)

```tsx
<DrawerBody>
  <div className="space-y-4">
    <Field className="gap-2">...</Field>
    <Field className="gap-2">...</Field>
  </div>
</DrawerBody>
```

### Multi-section forms (2+ logical groups)

```tsx
<DrawerBody>
  <div className="space-y-4">
    {/* Section 1 */}
    <Field className="gap-2">...</Field>
    <Field className="gap-2">...</Field>
  </div>
  <Separator className="my-6" />
  <div className="space-y-4">
    {/* Section 2 */}
    <Field className="gap-2">...</Field>
  </div>
</DrawerBody>
```

| Rule | Why |
| --- | --- |
| space-y-4 between fields within a section | Consistent vertical rhythm |
| <Separator className="my-6" /> between sections | Clear visual break |
| <Field className="gap-2"> wraps every labeled input | Consistent label-to-input spacing |
| <FieldLabel htmlFor={id}> on every field | Accessibility, clickable labels |
| <FieldDescription> for optional hints only | Don't over-describe obvious fields |

### Dynamic lists (e.g., approval steps)

```tsx
<div className="space-y-4">
  {items.map((item, i) => (
    <div key={i} className="border rounded-lg p-4 bg-muted/20 space-y-3">
      {/* fields for this item */}
      <div className="flex justify-end gap-2">
        {/* reorder / remove buttons */}
      </div>
    </div>
  ))}
  <button type="button" className={button({ variant: 'outline', size: 'sm' })}>
    Add Item
  </button>
</div>
```

## Actions (Submit / Cancel)

```tsx
<DrawerActions>
  <button
    type="button"
    className={button({ variant: 'outline', size: 'sm' })}
    onClick={close}
  >
    Cancel
  </button>
  <button
    type="submit"
    className={button({ size: 'sm' })}
    disabled={isSubmitting}
  >
    {isSubmitting && <Loader2 className="h-4 w-4 animate-spin mr-2" />}
    {isSubmitting ? 'Saving...' : isEditMode ? 'Save' : 'Create'}
  </button>
</DrawerActions>
```

| Rule | Why |
| --- | --- |
| Cancel is type="button" (never submit) | Prevent accidental submission |
| Submit is type="submit" | Native form submission, Enter key works |
| Use button() variant function, not <Button> component | One API for buttons (see ref-variant-system) |
| Cancel: variant: 'outline', Submit: default | Consistent visual hierarchy |
| Both size: 'sm' | Compact drawer footer |
| Disabled when isSubmitting | Prevent double-submit |
| Loading: Loader2 spinner + "Saving..." text | Clear feedback |

### Submit button text

| Mode | Idle | Loading |
| --- | --- | --- |
| Create | "Create" | "Creating..." |
| Edit | "Save" | "Saving..." |
| Special action | Verb ("Import", "Export") | "Importing...", "Exporting..." |

## Error Handling

Two levels of error display:

### 1. Field-level (validation)

Shown inline below the field (see Validation section above). Cleared on field change.

### 2. Submission-level (server errors)

Shown as an `Alert` at the top of `DrawerBody`:

```tsx
<DrawerBody>
  {alert && (
    <Alert
      type={alert.type}
      dismissible
      autoHide={alert.type === 'success' ? 3000 : undefined}
      onDismiss={() => setAlert(null)}
    >
      {alert.message}
    </Alert>
  )}
  {/* fields */}
</DrawerBody>
```

| Rule | Why |
| --- | --- |
| Server errors show as Alert inside DrawerBody | Visible without leaving the form |
| Success alerts auto-hide (3000ms) | Don't block the UI |
| Error alerts persist until dismissed | Ensure user acknowledges |
| Never silently swallow errors (no bare console.error) | Users must see failures |

## Success Handling

| Rule | Why |
| --- | --- |
| On success: call onSuccess?.() callback, then close drawer | Parent refreshes data |
| If server returns a sync handle: await result.wait() before closing | Ensure real-time sync completes |
| Reset form state on close | Clean slate for next open |

## Edit vs Create

A single form component handles both modes:

```tsx
type Props = {
  entity?: Entity       // undefined = create, defined = edit
  onClose: () => void
  onSuccess?: () => void
}

const isEditMode = !!entity
const [form, setForm] = useState(entity ?? { name: '', description: '' })
```

| Rule | Why |
| --- | --- |
| Presence of entity prop determines mode | No separate mode prop needed |
| Pre-fill from entity using useState(entity ?? defaults) | One initialization path |
| Disable immutable fields in edit mode (e.g., email) with <FieldDescription> explaining why | Clear UX |
| Title changes: "Create {Entity}" vs "Edit {Entity}" | Context |

## Shared Forms

When two screens need the same form (e.g., OrganizationScreen and UserManagementScreen both create/edit users):

| Rule | Why |
| --- | --- |
| Extract the form to a shared component | Single source of truth |
| Form component accepts entity?, onClose, onSuccess props | Reusable interface |
| Screen-specific behavior goes in the parent's onSuccess callback | Form stays generic |
| Avoid duplicate forms across screens; existing duplicates are migration candidates | Reduces drift while acknowledging current code state |

## Checkbox Lists (capabilities, permissions)

```tsx
<Field className="gap-2">
  <FieldLabel>Capabilities</FieldLabel>
  <div className="space-y-2">
    {ALL_CAPABILITIES.map(cap => (
      <label key={cap.id} className="flex items-start gap-2 cursor-pointer">
        <input
          type="checkbox"
          className="mt-1"
          checked={form.capabilities.includes(cap.id)}
          onChange={e => {
            const next = e.target.checked
              ? [...form.capabilities, cap.id]
              : form.capabilities.filter(c => c !== cap.id)
            update('capabilities', next)
          }}
        />
        <div>
          <div className="text-sm font-medium">{cap.label}</div>
          <div className="text-xs text-muted-foreground">{cap.description}</div>
        </div>
      </label>
    ))}
  </div>
</Field>
```

## File Upload Fields

For forms with file attachments (ImportDialog, PaymentRequestDialog):

| Rule | Why |
| --- | --- |
| Hidden <input type="file"> triggered by a styled drop zone | Custom drag-and-drop UX |
| File list in state: files: File[] | Controlled, can remove individual files |
| Show file name + size + remove button per file | User can manage selection |
| Accept attribute restricts file types | Prevent wrong uploads |

## Applies To

Every screen with create/edit/configure functionality:

- InvoiceScreen (ImportDialog, PaymentDialog)
- PaymentRequestsScreen (PaymentRequestDialog, ExportBankTransferDialog)
- PaymentsScreen (PaymentForm)
- UserManagementScreen (Create/Edit User, Role Assignment)
- TeamManagementScreen (Create/Edit Team)
- ApprovalConfigScreen (Edit Flow)
- OrganizationScreen (shared User/Team forms)

## Cited By

- `ref-variant-system` (button/input variants used in forms)
- `ref-ui-patterns` (Drawer dialog pattern)
