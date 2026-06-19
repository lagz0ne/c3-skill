# Topic Prompt: Build a Traceable Design System (custom canvases + dense wiring)

You are the keeper of a product's design system. Capture it in C3 as a **traceable
knowledge graph**, not as architecture docs. C3 is a general knowledge-graph contract
tool: you DEFINE YOUR OWN doc types (canvases) for this domain, author facts of those
types, and WIRE each fact to the facts it derives from — so `c3 check` validates that
every citation resolves and the design system becomes a graph you can audit.

Do **not** model this with the builtin architecture canvases (`system`, `container`,
`component`). A design token is not a deployable container; a button is not a process.
Define canvases that fit the domain.

## The domain

A design system has, at minimum:

- **design tokens** — named, themeable decisions: a name, a value, and a category
  (color, spacing, typography, radius, shadow). Tokens are the leaves: the single
  source of truth for every concrete value.
- **UI components** — buttons, inputs, modals, etc. Each has an anatomy, interaction
  states, and accessibility behavior. A component must **consume tokens, never
  hardcode values**, and must **follow** an accessibility rule / interaction pattern.
- **accessibility rules / interaction patterns** — reusable "why": focus-visible
  contract, hit-target minimum, contrast floor, disclosure pattern, etc.
- **flows** — an ordered user journey that **sequences components** (e.g. checkout:
  product → cart → address → pay → confirmation).

## What you must do

### 1. Design 2–4 custom canvases for this domain

Define your own doc types with `C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh
canvas add <id> < schema.md`. A canvas schema is YAML: frontmatter (`id`,
`type: canvas`, `description`), then a body with `domain:` and `sections:`. Each
section has `name`, `content_type: text|table`, `required:`, `purpose:`, and for
tables a `columns:` list (each column a `name:` + a `type:`).

Before you write a single schema, study the format and the column vocabulary on the
builtin **non-architecture** canvases — they are your templates:

```
C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh canvas read user-story
C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh canvas read pm-requirement
C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh canvas read atomic-design-change
```

Column `type:` values include: `text`, `enum` (with a `values:` list), `date`,
`entity_id`, `cite`, `evidence`, `check`, and — the one that makes a graph —
`reference` / `edge<...>`. An **edge column** is how one fact CITES another:

```yaml
# explicit form: type reference + the relationship + allowed target types
- name: Token
  type: reference
  edge: uses
  targets:
    - design-token
```

```yaml
# inline shorthand for the same idea
- name: Component
  type: edge<ui-component>
```

A canvas may declare **at most one column per relationship name** (one unambiguous
place a `uses` edge lands), so pick distinct edge names when a canvas wires two
different relationships.

**Hard requirement: every custom canvas you define must carry at least one edge
column that wires to another of this domain's types.** Concretely, your canvases must
make these edges expressible:

- a UI component → an edge column (e.g. `uses-tokens`) citing the **design tokens** it
  consumes;
- a UI component → an edge column (e.g. `follows`) citing the **a11y rule / pattern**
  it obeys;
- a flow → an edge column citing the **UI components** it sequences, in order.

### 2. Author a lean first cut — then GROW it under a named pressure

Read the canvas you defined with `c3 schema <type>`, then author facts of your custom
types (`c3 add <type> <slug> --file body.md`, or a change-unit create-patch for a
frozen fact). Keep the first cut deliberately lean: a small token set, a couple of
components, one or two a11y rules, one flow. Do **not** pre-build later-stage
structure.

Then apply this growth pressure and make it a visible step, not a late cleanup:

> **Theming.** The product must ship a light theme and a dark theme (or two brands).
> A single flat token like `color.action.primary = #2563EB` can no longer be the
> source of truth, because its value now depends on the theme.

This forces a real **canvas climb / new type**: split tokens into **primitive** tokens
(raw palette values, e.g. `blue-600 = #2563EB`) and **semantic** tokens (theme-aware
roles, e.g. `action.primary`, which resolves to a primitive **per theme**). Raise the
token canvas (or add a semantic-token type) so a semantic token can **cite the
primitive(s) it resolves to per theme** via an edge column — and **migrate** affected
facts: components must now consume the **semantic** token, not the raw primitive.
Re-point the existing component cites and verify nothing dangles.

You decide the exact shape (raise one canvas vs. add a sibling type) — but the climb,
the new edge (semantic → primitive), and the migration of component cites must all be
real and visible.

### 3. Wire a dense traceability graph and verify it

The point of the edge columns is coverage. After the climb:

- **every UI component** traces to the token(s) it uses and the a11y rule it follows —
  no component carries a hardcoded value where a token cite belongs;
- **every (semantic) token** is used by **at least one** component — an orphan token
  (defined, cited by nobody) is a failure;
- at least one **end-to-end flow** sequences real components across the system.

Prove coverage with the reverse graph, not by assertion:

```
C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh graph <token-id> --direction reverse
```

A token's reverse graph must show the components that consume it. Run this for a token
that should be used and (ideally) for one you suspect is orphaned, and report what you
found.

### 4. Keep `c3 check` honest

Run `c3 check` and **report the exact result** — pass/fail and any warnings or errors,
named. Resolve real problems (a cite to a non-existent fact, a missing required
section). If a warning is structural noise you choose not to chase (e.g. the seeded
top-level context), say so explicitly rather than implying silence. Do not claim
"clean" if the tool printed warnings — name them and explain.

## Constraints

- Use local C3 only: `C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh`.
- Define and use **custom** canvases for the domain. Do **not** model the design
  system with the builtin `system` / `container` / `component` architecture canvases.
- Avoid codemap work — there is no implementation repo here; this is a docs-first
  knowledge graph. Do not fall into Derived-Materials / file-binding busywork.
- Prefer compact, concrete C3 artifacts (real ids, real cites, real edges) over long
  prose.
- Run verification and report the exact result, including the reverse-graph evidence
  and the literal `c3 check` outcome.
