---
id: c3-102
c3-version: 3
c3-seal: 1249a97db46fb8d38c19092242731240c25ad90c27a5f64249f81f27a4a7c218
title: Form System
type: component
category: foundation
parent: c3-1
goal: Form, FormComponents, Validator for data entry with validation
uses:
    - ref-form-patterns
---

# Form System

## Goal

Form, FormComponents, Validator for data entry with validation

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own Form System behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep Form System decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for Form System so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before Form System behavior is changed. | ref-form-patterns |
| Inputs | Accept only the files, commands, data, or calls that belong to Form System ownership. | ref-form-patterns |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-form-patterns |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-form-patterns |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks Form System to deliver its documented responsibility. | ref-form-patterns |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-form-patterns |
| Alternate paths | When a request falls outside Form System ownership, hand it to the parent or sibling component. | ref-form-patterns |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-form-patterns |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-form-patterns | ref | Governs Form System behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| Form System input | IN | Callers must provide context that matches the component goal and parent fit. | c3-1 boundary | c3x lookup plus targeted tests or review. |
| Form System output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-1 boundary | c3x check and project test suite. |

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

- `tailwind-variants` -- styling via variants from `ui/variants.ts`
- `@tabler/icons-react` -- icons in specialized inputs
- `lucide-react` -- Loader2 spinner for submit buttons

## Architecture

### Form (Primary)

Context-based form with `FormContext`. Manages values, errors, touched state, and debounced field validation (300ms).

**Components in `Form.tsx`:**

| Component | Purpose |
| --- | --- |
| Form | Wrapper. Takes initialValues, validationRules, onSubmit |
| Input | Text/password/email with floating label, prefix/suffix, left/right icon support |
| Textarea | Multi-line input |
| Select | Dropdown with safe option handling |
| FileInput | Drag-and-drop file upload with size validation |
| RadioGroup | Segmented button group (horizontal/vertical) |
| SubmitButton | Auto-disables during submission, shows spinner |
| Button | General purpose button (non-form) |
| FormActions | Footer bar with alignment variants |

**Components in `FormComponents.tsx`:**

| Component | Purpose |
| --- | --- |
| FloatingInput | Minimal fieldset-based input |
| InputGroup | Input with icons/prefix/suffix in a combined border |
| PasswordInput | Visibility toggle + optional strength indicator |
| SearchInput | Search with clear button |
| DateRangePicker | Start/end date pair |
| PhoneInput | Country code selector + phone field |
| CurrencyInput | Currency symbol prefix + decimal formatting |

### Validator (Alternative)

Standalone `ValidationProvider` with richer rule types (conditional validation, dependencies, built-in presets for email/password/phone/currency/url).

| Component | Purpose |
| --- | --- |
| ValidationProvider | Wraps children with validation context |
| ValidatorInput | Self-contained input with debounced validation |
| ValidatorTextarea | Self-contained textarea |
| ValidatorSelect | Self-contained select |
| ValidationPresets | Pre-built rules: email, password, phone, currency, url |
| ValidationUtils | Helpers: isFormValid, getFirstError, combineRules |

## Validation

**Form approach** uses `lib/validation.ts`: simple `ValidationRule` with required/minLength/maxLength/pattern/custom. Applied via `validateForm()`.

**Validator approach** uses richer rules: conditional `when()`, field `dependencies`, typed validators (email, phone, currency, url), and custom functions with access to all form values.

Both use debounced validation (300ms default) and show errors only after field is touched.

## Usage

```tsx
<Form
  initialValues={{ name: "", amount: 0 }}
  validationRules={{
    name: { required: true, minLength: 2 },
    amount: { required: true, custom: v => v <= 0 ? "Must be positive" : null },
  }}
  onSubmit={async (values) => {
    await actions.act("createPr", values)
  }}
>
  <Input name="name" label="Name" required />
  <Input name="amount" label="Amount" type="number" required />
  <FormActions>
    <SubmitButton>Create</SubmitButton>
  </FormActions>
</Form>
```

## Wiring to Server

Forms submit via the `actions` atom (`api.act()`). The action sends FormData to `/act` endpoint. Server validation errors surface through the ActionResult, typically shown via toast.
