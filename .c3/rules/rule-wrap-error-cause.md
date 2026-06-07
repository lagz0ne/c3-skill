---
id: rule-wrap-error-cause
c3-seal: d1b598183f616c01a0937ebdd31914bb3e5b511af892b366556c5ef08501e5aa
title: wrap-error-cause
type: rule
goal: Every returned error in the Go CLI preserves its cause and context so failures stay traceable across the dispatcher, store, and command layers.
---

# wrap-error-cause

## Goal

Every returned error in the Go CLI preserves its cause and context so failures stay traceable across the dispatcher, store, and command layers.

## Rule

All returned errors must wrap the cause with `fmt.Errorf("<context>: %w", err)` — never drop the cause or format a returned error with `%v`.

## Golden Example

```go
// cli/main.go — dispatcher wrapping a cache-refresh failure
if err := cmd.EnsureLocalCache(c3Dir, opts.IncludeADR, opts.Only, io.Discard); err != nil {
    return fmt.Errorf("error: refresh cache before %q: %w", opts.Command, err) // REQUIRED: %w preserves the cause
}
```

## Not This

| Anti-Pattern | Correct | Why Wrong Here |
| --- | --- | --- |
| return fmt.Errorf("failed: %v", err) | return fmt.Errorf("failed: %w", err) | %v flattens the cause, so callers lose errors.Is/errors.As and the unwrap chain. |
| return err at a layer boundary | return fmt.Errorf("<context>: %w", err) | A bare return drops the context needed to locate where the failure crossed layers. |

## Scope

**Applies to:**

- All Go components in the c3-1 container that return errors across function or layer boundaries.

**Does NOT apply to:**

- Test helpers asserting on sentinel errors with `errors.Is`.

## Override

Deviate only when returning a sentinel error meant for `errors.Is` matching; document the sentinel at its definition.
