---
target: act-3-grow-the-canvas
scope: whole
type: design-act
title: "Act 3 — The canvas grows with the need"
---
# Act 3 — The canvas grows with the need

## Thesis

The canvas grows with the need.

## Move

When the work outgrows the model, raise the canvas a rung and migrate every fact up to it; completeness is never relaxed — every fact stays complete to the new bar.

## Tool Guarantee

| Guarantee | Mechanism | Source |
|-----------|-----------|--------|
| Raising the bar migrates every fact through the same atomic saga, and the merged result must satisfy the raised canvas | The climb is a change-unit: its `insert` carriers append the new section to each fact and `apply` runs the canvas gate over the merged body inside one all-or-nothing transaction | `cli/cmd/change.go`, `cli/internal/changeset/apply.go` |
| A fact carried up that derives code re-attests against its raised contract before the climb can land | The inspection gate fires on the climb's contract-touching `insert` carriers, refusing the unit until each obligated fact's inspection points inside its resolved territory | `cli/cmd/inspect.go` |

## Why This Shape

Growth is climbing to a higher rung rather than loosening the current one, so completeness is never relaxed; reusing the change-unit saga for the migration — instead of a bespoke "schema upgrade" path — means the raised canvas is enforced the instant the climb lands, with the same atomic, gated guarantees every other change already gets.

## Surfaces

| Surface | Owns |
|---------|------|
| `skills/c3/references/canvas.md` | The canvas as a rung: why a fact is always complete to its current rung and what raising the bar means |
| `skills/c3/references/change.md` | Climbing operationally: scaffolding the per-fact `insert` carriers and landing the migration atomically |
