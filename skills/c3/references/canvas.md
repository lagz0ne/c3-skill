# canvas — your architecture's own vocabulary, as a contract

This is the **model** surface (Act 1, the shaping; Act 3, why you climb). It owns
what a canvas *is* and why you *raise* one. The climb's steps live in change.md; the
frozen-fact rule and rung *definition* are the shared contract in SKILL.md — cited
here, not re-taught.

## A canvas is the shape of a fact-type

A **canvas** is your project's vocabulary made enforceable: for one fact-type
(`system`, `container`, `component`, `ref`, `rule`, `recipe`, the change-docs, or any
type you define), it declares the sections each fact carries and each table's typed
columns. The shape is **data, not code**: `c3 init` seeds lean canvases, materializes
them to `.c3/canvases/<type>.md`, and from then on **you own them**. `c3 check`
validates facts against *that file* — so editing a definition changes what is
enforced, with no second hardcoded copy. Read the live shape, never memory:
`c3 schema <type>` (rendered, leads with REJECT IF) and `c3 canvas read <type>` (the
owned source) are the same contract from two angles.

**Membership is part of the shape, and the tool fills it.** A parent fact-type carries
a membership table, but no author ever writes a row into it: every parentage-changing
path synthesizes it from the children's `parent:` edges (`c3 add`, `change apply`,
`check --fix`). Set a child's `parent:`, the row appears; the column is a *consequence*
of the shape, never hand-authored truth. (Mechanics: change.md.)

## Define your own fact-type — and wire it

C3 is a general knowledge-graph tool, not only an architecture one. When the project
needs a doc kind the builtins don't cover — a `test-case`, a `design-token`, a
`pm-objective` — **define its canvas** and the type is first-class (authored, frozen,
checked, graphed like any other): `c3 canvas add <id> < schema.md`, where the schema is
a `type: canvas` doc with `domain:`, `sections:` (each `name` / `content_type: text|table`
/ `required` / `purpose`, and for tables `columns:` with a `type:`), and `reject_if:`.
Copy a builtin for the exact shape — `c3 canvas read user-story`.

**Wiring is a column.** A table column becomes a graph edge when its `type:` is
`reference` (with `edge: <rel>` + `targets: <type>,…`) or `edge<typeA|typeB>`: put a
cited fact's id in the cell, and `c3 check` materializes an edge of that relationship
and verifies the citation resolves — to **any** fact, builtin or your own custom type.
This is how a `test-case` **verifies** a `requirement`, a `ui-component` **uses** a
`design-token`, a `story` **serves** an `objective`. Dense wiring turns separate docs
into one traceable graph; `c3 graph <id> --direction reverse` shows who cites a fact —
its coverage. A cell that names no resolvable fact reads "ungrounded"; `N.A - <reason>`
is the explicit "no link, on purpose."

## A canvas is a rung — why you raise it

A canvas is a **rung**: a complete contract for one complexity *level*, sized to the
project **now** (rung-1 = the lean `init` default). A fact is always complete to its
current rung — deeper sections are a higher rung, not a hole to backfill (the rung
definition is SKILL.md's contract). So growth is never "fill in the thin parts"; it is
a **climb**: when the work outgrows the model, you *raise the canvas* — make a section
required, or author a richer one — and then **every fact below the new bar migrates up,
completely**. That migration is *why* the climb exists: integrity forbids a fact
straddling two rungs, so raising the bar obligates bringing all facts up to it. The
climb runs as a change-unit (`change scaffold` → fill → `change apply`, gated so it
refuses while any required section is still empty) — the steps are change.md's.

## Worked outcome

- **"Make our components carry a `## Threat Model`"** → `c3 canvas read component`,
  add the section, `c3 canvas write component`, then `c3 check`. The section is now
  required and check enforces it across every existing component (a rung climb).
- **"Track test coverage"** → `c3 canvas add test-case` with a `verifies` column
  (`type: reference, edge: verifies, targets: requirement`); author cases that put the
  requirement id in that column; `c3 graph <req-id> --direction reverse` then shows every
  covering case — and an untested requirement shows none (a custom fact-type + wiring).

## Anti-goals

- **Don't enumerate a fixed set of sections, types, or columns in prose.** The shape
  is the user's data — read it (`c3 canvas list`, `c3 schema <type>`), don't recite it.
- **Don't treat `adr` as special.** ADR is just the `adr` canvas; its shape is editable
  like any other fact-type's.
