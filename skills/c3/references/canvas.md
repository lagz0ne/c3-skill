# canvas — inspect and edit the shape of entities

A **canvas definition** is the shape of an entity type: its sections (text or
table) and each table's typed columns. Every entity type has one — `context`,
`container`, `component`, `ref`, `rule`, `adr`, and document types (`prd`,
`user-story`, `atomic-design-change`, `pm-requirement`, plus any project-defined
type). c3x ships embedded defaults as **seeds**; on onboard they are materialized
into `.c3/canvases/<type>.md` as sealed markdown, and from then on **the user
owns them**.

The point: definitions are not baked into the tool. A team edits its definitions
to fit how *they* want their architecture docs shaped — c3x facilitates the wiring
(scaffold / validate / check); it does not dictate the shape.

## A canvas is a rung (the ladder)

A canvas is a **rung**: a complete contract for one complexity *level*. It starts at
the complexity that fits the project **now** — `c3 init` seeds **lean** rung-1
canvases — and is **raised deliberately** as the architecture earns it. Integrity is
the invariant: a fact is *always* complete to its canvas's current bar, never thin or
"filled in later." What grows is the level, not the completeness.

**Raising a canvas is a climb, and it migrates the facts.** When you make a section
required (or author a richer canvas with `c3 canvas write`), every existing fact that
now sits *below* the new bar must be brought up to it — completely. That migration is
the same change-unit mechanism, now with a name:

```bash
c3x change scaffold <unit-id>   # one insert-patch per below-bar fact, each with the
                                #   missing required sections as EMPTY templates
# ... fill every templated section with real content ...
c3x change apply <unit-id>      # lands the climb atomically — REFUSES to apply while
                                #   any templated section is still empty (canvas gate)
```

Scaffold stages the climb; apply lands it **only once filled** — the empty-section
gate is what guarantees the climb is a real one, not a rename. Each rung is solid on
its own and is **not** responsible for future rungs: size to now, climb when the
architecture warrants it, never pre-author a higher rung's sections.

Canvases stay **user-owned and editable** throughout — the ladder is a named mechanism
over the same `canvas write` + change-unit moves, not a new authority over the shape.

## Commands

```bash
c3 canvas list                 # every entity-type definition + domain + source
c3 canvas read <type>          # the canonical definition (user-owned markdown)
c3 schema <type>               # render the definition: sections, columns, REJECT IF
c3 canvas write <type> --file  # replace a project's definition (customize the shape)
c3 canvas add <type> --file    # define a brand-new project-specific entity type
```

`c3 schema <type>` and `c3 canvas read <type>` are the same contract from two
angles — `schema` is the rendered view, `canvas read` is the owned source.

## Outcomes this enables

- **"What sections does X require?"** -> `c3 schema <type>`. Answer from the
  project's definition, never from memory.
- **"Make our component docs carry a `## Threat Model` section"** -> `c3 canvas
  read component` > edit > `c3 canvas write component --file ...`, then `c3 check`.
  The new section is now required; check enforces it.
- **"Add a `decision-log` doc type"** -> author a definition, `c3 canvas add
  decision-log --file ...`. Now `c3 add decision-log <slug>` works.

## How it behaves

- **Validation reads the definition.** A user's edit changes what `c3 check`
  enforces — there is no second, hardcoded copy of the shape. If you edit a
  definition, re-run `c3 check` to see the new contract applied.
- **Frozen / user-owned.** A c3x upgrade ships new embedded seeds but **never
  overwrites** an existing `.c3/canvases/<type>.md` (write-if-absent). New seeds
  reach a project only on fresh onboard or an explicit re-materialize.
- **Column primitives (lean, mechanically checkable):** `text`, `date`, `enum`
  (with values), `cite`, `check`, `entity_id`, `reference`, `evidence`,
  `edge<a|b|...>`. A definition that uses an unsupported primitive is rejected.

## Anti-goals

- Don't enumerate a fixed set of sections or types in prose — read `c3 canvas list`
  / `c3 schema <type>`. The shape is data, and it is the user's.
- Don't treat `adr` as special. ADR is the `adr` canvas; its shape is editable like
  any other.
- Don't validate an entity against remembered sections — validate against its
  definition (`c3 check` already does this).
- Don't hand-edit entity *instances* under `.c3/` — those stay CLI-only. Only
  *definitions* under `.c3/canvases/` are user-editable.
