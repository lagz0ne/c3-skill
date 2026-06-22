---
id: rule-dispatcher-error-hint
c3-seal: 47d8ec4bcd3c546ba7657e0689b3d9ccabd2c39352132ba19cf65687d033253c
title: dispatcher-error-hint
type: rule
goal: 'Make every user-facing CLI failure recoverable: a wrong invocation should tell the user what to do next, not just that something was wrong.'
---

# dispatcher-error-hint

## Goal

Make every user-facing CLI failure recoverable: a wrong invocation should tell the user what to do next, not just that something was wrong.

## Rule

A user-facing error returned from the top-level dispatcher (a usage error, a missing argument, an unknown command, an unmet precondition) carries an actionable hint. The hint rides in the error string on its own line, prefixed `hint:`, after an `error:` summary line. `main` prints the returned error verbatim to stderr and exits non-zero, so the hint reaches the user with no extra wiring. Hints name a concrete next command (e.g. `c3x init`, `c3x change <sub> <id>`, `c3x --help`).

## Golden Example

```go
case "lookup":
    if len(opts.Args) < 1 {
        return fmt.Errorf("error: lookup requires a <file-path> argument\nhint: run 'c3x lookup --help' for usage")
    }
```

```go
default:
    return fmt.Errorf("error: unknown command '%s'\nhint: run 'c3x --help' to see available commands", opts.Command)
```

## Not This

| Anti-Pattern | Correct | Why Wrong Here |
| --- | --- | --- |
| return fmt.Errorf("invalid command") | return fmt.Errorf("error: unknown command '%s'\nhint: run 'c3x --help' to see available commands", opts.Command) | A bare failure tells the user nothing about how to recover; the dispatcher is the one place that knows the valid next step. |
| Printing the hint with fmt.Fprintln(os.Stderr, ...) inside the handler | Return it in the error string; main already prints the error to stderr and exits 1 | Side-channel printing splits the message from the non-zero exit and bypasses the testable error return. |
