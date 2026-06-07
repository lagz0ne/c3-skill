---
id: rule-dispatcher-error-hint
c3-seal: 6986cbfc4235bffc8830e15ba72af62fcf572084baa503e95397e61d7f8fdd38
title: dispatcher-error-hint
type: rule
goal: User-facing CLI errors from the top-level dispatcher guide the user to a next step, so a failure is actionable rather than a bare message.
---

# dispatcher-error-hint

## Goal

User-facing CLI errors from the top-level dispatcher guide the user to a next step, so a failure is actionable rather than a bare message.

## Rule

User-facing dispatcher errors must carry an `error:` prefix and, when the failure is recoverable, a `hint:` line naming the next step.

## Golden Example

```go
// cli/main.go — missing local cache, recoverable
if !hasDB {
    return fmt.Errorf("error: local C3 cache unavailable at %s\nhint: run 'c3x check' to rebuild from canonical .c3/, or 'c3x init' if this project is not onboarded", dbPath) // REQUIRED: error: prefix + hint: next step
}
```

## Not This

| Anti-Pattern | Correct | Why Wrong Here |
| --- | --- | --- |
| return fmt.Errorf("cache not found") | return fmt.Errorf("error: ...\nhint: ...") | A bare lowercase message at the user boundary gives no recovery path, so the user is stuck guessing the next command. |

## Scope

**Applies to:**

- c3-108 (runtime-support) — the top-level dispatcher in `cli/main.go` that surfaces errors to the CLI user.

**Does NOT apply to:**

- Internal library errors that are wrapped and re-surfaced by the dispatcher (those follow rule-wrap-error-cause instead).

## Override

Non-recoverable invariant violations may omit `hint:` when no user action can resolve them; keep the `error:` prefix.
