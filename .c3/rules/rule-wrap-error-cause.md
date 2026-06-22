---
id: rule-wrap-error-cause
c3-seal: 20b666f882d8ac06aa87aebbbef1b4b3c58c2efa2bbade95145e03afaafeaa0d
title: wrap-error-cause
type: rule
goal: 'Preserve the failure chain: an error that crosses a function boundary should say what this layer was doing and still carry the underlying cause.'
---

# wrap-error-cause

## Goal

Preserve the failure chain: an error that crosses a function boundary should say what this layer was doing and still carry the underlying cause.

## Rule

When a function returns an error it received from a lower layer, it wraps that error with context using `fmt.Errorf("<what this layer was doing>: %w", err)`. The `%w` verb (not `%v` or `%s`) is required so the original error stays unwrapped-reachable for `errors.Is` / `errors.As` and so the full chain is visible. The context names the operation and, where it helps, the subject (the file, entity, or stage) — e.g. `apply %s: %w`, `read inspect %s: %w`. A sentinel returned unchanged needs no wrap; a freshly constructed leaf error (no cause to carry) is the exception.

## Golden Example

`````go
func LoadEvalSpecs(c3Dir string) ([]eval.Spec, error) {
	entries, err := os.ReadDir(filepath.Join(c3Dir, "eval"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read eval dir: %w", err)
	}
	var specs []eval.Spec
	for _, ent := range entries {
		b, err := os.ReadFile(filepath.Join(c3Dir, "eval", ent.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", ent.Name(), err)
		}
		var sp eval.Spec
		if err := yaml.Unmarshal(b, &sp); err != nil {
			return nil, fmt.Errorf("parse %s: %w", ent.Name(), err)
		}
		specs = append(specs, sp)
	}
	return specs, nil
}
```
````

## Not This

| Anti-Pattern | Correct | Why Wrong Here |
| --- | --- | --- |
| return fmt.Errorf("read inspect %s: %v", name, err) | return fmt.Errorf("read inspect %s: %w", name, err) | %v flattens the cause to text, so errors.Is/errors.As can no longer reach the underlying error. |
| return err straight up through several layers | return fmt.Errorf("apply %s: %w", p.Source, err) | An unwrapped bubble loses which stage and which file failed — the saga error becomes an opaque leaf with no context. |
`````
