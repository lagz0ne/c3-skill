---
id: ref-form-patterns
title: Form Patterns
---

# Form Patterns

## Goal

Establish conventions for building forms with validation, state management, and UI components using a context-based Form component.

## Conventions

| Rule | Why |
|------|-----|
| Use Form wrapper with FormContext | Provides centralized state, validation, and submission handling |
| Define validationRules declaratively | Enables consistent validation across forms with reusable rules |
| Use FormField components (not raw inputs) | Ensures consistent styling, error display, and accessibility |
| Validation debounced at 300ms | Prevents excessive validation during typing |
| Errors cleared on field change | Immediate feedback when user corrects input |
| onValuesChange for dependent fields | Enables dynamic form behavior (e.g., show/hide based on type) |
| Submit button disabled while isSubmitting | Prevents double submission |

## Testing

| Convention | How to Test |
|------------|-------------|
| Validation rules trigger | Fill field, blur, verify error appears |
| Debounced validation | Fill rapidly, verify single validation call |
| Error clearing | Enter invalid, then valid, verify error clears |
| Submit prevention | Click submit, verify button disabled during async |
| Form context access | Mount child outside Form, expect error thrown |

## References

- `apps/start/src/components/Form*.tsx` - Form components
- `apps/start/src/lib/validation.ts` - Validation utilities
