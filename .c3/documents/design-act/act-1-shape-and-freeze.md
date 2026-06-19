---
id: act-1-shape-and-freeze
c3-seal: 80884b2a1ee3f2ac77160b58c13b1688c03ea922be1b87477cde05d49a7efb80
title: Act 1 — Shape the model, then freeze the facts
type: design-act
---

# Act 1 — Shape the model, then freeze the facts

## Thesis

Shape the model, then freeze the facts.

## Move

Build the canvas — your architecture's own vocabulary — then onboard the facts the work needs, draft the first work, and flip the gate. At that flip the facts freeze: they become shared truth and are never hand-edited again.

## Tool Guarantee

| Guarantee | Mechanism | Source |
| --- | --- | --- |
| A fact joins its parent the instant you declare parentage — membership is never hand-authored | c3 add synthesizes the child's row into the parent on create, and check re-derives any missing row, so membership exists by construction | cli/cmd/add.go, cli/cmd/check_enhanced.go, cli/internal/changeset/membership.go |
| Once a fact carries a body it can no longer be edited in place — it is frozen | The freeze rule refuses every direct write/set/delete on a bodied fact and names the change-unit as the only legal mutation path | cli/cmd/freeze.go |

## Why This Shape

The model is built before the facts so the vocabulary stays the architecture's own rather than the tool's defaults; freezing is deferred to a gate flip — not imposed at first keystroke — so onboarding and the first draft stay fluid, and the hard guarantee lands exactly when the fact becomes shared truth that others will rely on.

## Surfaces

| Surface | Owns |
| --- | --- |
| skills/c3/references/onboard.md | Adopting a project: seeding the initial facts the first work needs and flipping the gate that freezes them |
| skills/c3/references/canvas.md | Defining the canvas — the architecture's own vocabulary a fact must satisfy |
