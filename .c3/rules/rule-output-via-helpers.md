---
id: rule-output-via-helpers
c3-version: 4
c3-seal: 913627f9b1ef2613fc80ecd2e04bac23864e1724d09a139579851c49f4214c26
title: output-via-helpers
type: rule
goal: All command results serialize through one output layer so agent mode always yields TOON and human/JSON formats stay consistent across commands.
origin:
    - adr-20260415-agent-mode-toon-only
    - adr-20260415-force-agent-toon-output
---

# output-via-helpers

## Goal

All command results serialize through one output layer so agent mode always yields TOON and human/JSON formats stay consistent across commands.

## Rule

Commands must emit results via `WriteTableOutput`/`WriteObjectOutput` with a format from `ResolveFormat` — never call `json.Marshal` or `fmt.Fprintf` to serialize a result directly.

## Golden Example

```go
// cli/cmd/output.go - format resolution + the shared table writer
func ResolveFormat(jsonExplicit bool, agent bool) OutputFormat {
	if agent {
		return FormatTOON // REQUIRED: agent mode is always TOON
	}
	if jsonExplicit {
		return FormatJSON // REQUIRED: explicit JSON remains non-agent compatibility
	}
	return FormatTOON // REQUIRED: default structured output is TOON
}

// WriteTableOutput writes a tabular dataset with optional help hints.
// JSON is explicit compatibility; every other structured format uses TOON.
func WriteTableOutput(w io.Writer, label string, data any, fields []string, format OutputFormat, hints []HelpHint) error {
	switch format {
	case FormatJSON:
		if err := writeJSON(w, data); err != nil {
			return err
		}
		writeHints(w, hints)
		return nil
	default:
		out, err := toon.MarshalTable(label, data, fields) // REQUIRED: TOON path goes through the shared marshaller
		if err != nil {
			return err
		}
		fmt.Fprint(w, out)
		writeHints(w, hints)
		return nil
	}
}
```

## Not This

| Anti-Pattern | Correct | Why Wrong Here |
| --- | --- | --- |
| json.NewEncoder(w).Encode(result) in a command | WriteObjectOutput(w, result, format, hints) | A command-local encoder bypasses ResolveFormat, so C3X_MODE=agent no longer yields TOON — the exact regression adr-20260415-agent-mode-toon-only exists to prevent. |

## Scope

**Applies to:**

- c3-108 (runtime-support) and every command component that emits results to stdout.

**Does NOT apply to:**

- Diagnostic `fmt.Fprintln(stderr, ...)` warnings, which are not result payloads.

## Override

A command needing a non-TOON/JSON artifact (e.g. mermaid from `graph`) may write its own format, but must still route structured result data through the shared helpers.
