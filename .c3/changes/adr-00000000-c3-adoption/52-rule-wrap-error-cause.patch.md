---
target: rule-wrap-error-cause
scope: whole
type: rule
title: wrap-error-cause
---
# wrap-error-cause

## Goal

Preserve the failure chain: an error that crosses a function boundary should say what this layer was doing and still carry the underlying cause.

## Rule

When a function returns an error it received from a lower layer, it wraps that error with context using `fmt.Errorf("<what this layer was doing>: %w", err)`. The `%w` verb (not `%v` or `%s`) is required so the original error stays unwrapped-reachable for `errors.Is` / `errors.As` and so the full chain is visible. The context names the operation and, where it helps, the subject (the file, entity, or stage) — e.g. `apply %s: %w`, `read inspect %s: %w`. A sentinel returned unchanged needs no wrap; a freshly constructed leaf error (no cause to carry) is the exception.

## Golden Example

```go
func ReadInspectDir(dir string) ([]InspectCarrier, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read change folder %s: %w", dir, err)
	}
	...
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("read inspect %s: %w", name, err)
		}
		...
	}
}
```

## Not This

| Anti-Pattern | Correct | Why Wrong Here |
| --- | --- | --- |
| `return fmt.Errorf("read inspect %s: %v", name, err)` | `return fmt.Errorf("read inspect %s: %w", name, err)` | `%v` flattens the cause to text, so `errors.Is`/`errors.As` can no longer reach the underlying error. |
| `return err` straight up through several layers | `return fmt.Errorf("apply %s: %w", p.Source, err)` | An unwrapped bubble loses which stage and which file failed — the saga error becomes an opaque leaf with no context. |
