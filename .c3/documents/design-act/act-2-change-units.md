---
id: act-2-change-units
c3-seal: b2e3bb087a37b2fa415282b0efc39626b55fcc59439ca7469f770b5d2c76f978
title: Act 2 — Change-units drive the work
type: design-act
---

# Act 2 — Change-units drive the work

## Thesis

Change-units drive the work.

## Move

Every change to a frozen fact rides a change-unit — an atomic, all-or-nothing saga: you declare the intent, the tool keeps the result integral across every gate, and you flip it in.

## Tool Guarantee

| Guarantee | Mechanism | Source |
| --- | --- | --- |
| The whole unit lands or nothing does, and only after the merged result still satisfies its canvas | apply runs the drift and canvas gates over every carrier, then writes inside one transaction that rolls back completely on any failure | cli/cmd/change.go, cli/internal/changeset/apply.go |
| A unit cannot destroy a fact and leave a child orphaned or a citer dangling | The destruction gate refuses a retire whose consequences the same unit does not also resolve | cli/cmd/destruction.go |
| Changing a fact that derives code forces a fresh, territory-grounded attestation that the code was inspected against the new body | The inspection gate refuses the unit unless each contract-touched fact with obligations carries an inspection whose evidence names a file inside the fact's resolved code-map territory | cli/cmd/inspect.go |

## Why This Shape

A frozen fact is shared truth, so it cannot be edited casually; routing every edit through one atomic saga makes integrity the tool's job rather than the author's discipline, which is why the alternative — direct edits guarded by review — was rejected: review catches drift by opinion, the saga catches it by construction.

## Surfaces

| Surface | Owns |
| --- | --- |
| skills/c3/references/change.md | The change-unit saga end-to-end: authoring the carriers, the gates apply runs, and flipping the unit in |
