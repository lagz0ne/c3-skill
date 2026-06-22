---
target: rule-output-via-helpers
scope: whole
type: rule
title: output-via-helpers
---
# output-via-helpers

## Goal

Keep machine output uniform: one place decides TOON vs JSON and honors agent mode, so every command speaks the same serialization.

## Rule

A command's structured result serializes through the shared output helpers — `WriteTableOutput` for a tabular dataset, `WriteObjectOutput` for a single object — never a hand-rolled `fmt.Print` of the data. The helpers own the format switch: agent mode and the default both emit TOON, an explicit `--json` outside agent mode emits JSON, and help hints attach only in agent mode. `fmt.Fprint` is for the prose/human path and for emitting the helper's already-formatted output, not for re-implementing serialization at a call site.

## Golden Example

```go
// WriteObjectOutput writes a single object with optional help hints.
// JSON is explicit compatibility; every other structured format uses TOON.
func WriteObjectOutput(w io.Writer, data any, format OutputFormat, hints []HelpHint) error {
	switch format {
	case FormatJSON:
		if err := writeJSON(w, data); err != nil {
			return err
		}
		writeHints(w, hints)
		return nil
	default:
		out, err := toon.MarshalObject(data)
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
| `fmt.Fprintf(w, "%s: %s\n", row.ID, row.Name)` to emit a result set | `WriteTableOutput(w, "items", rows, fields, format, hints)` | A bespoke format ignores the TOON/JSON switch and agent mode, so two commands drift into two machine formats. |
| `json.NewEncoder(w).Encode(data)` directly in a handler | Route through `WriteObjectOutput` / `writeJSON`, which apply agent-mode TOON | Encoding inline bypasses agent-mode TOON and the help-hint attachment the helpers centralize. |
