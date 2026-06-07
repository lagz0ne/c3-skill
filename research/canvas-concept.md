# Canvas — finished concept (draft for review)

> Status: **draft**. Drafted while the baseline eval runs (no live skill/CLI/`.c3`
> mutation). To be promoted to an ADR (itself a `c3-adr` canvas instance) and
> decomposed into gated implementation slices **after** the baseline lands.

## The concept in one line

**A canvas definition is C3's universal shape language: every C3 entity type —
component, container, system, ref, rule, ADR, PRD, user-story — is described by a
canvas definition, and `c3 check` is the single validator that enforces every
entity against its definition.** Canvas is not an authoring feature on top of C3;
it is the schema substrate underneath all of it.

## Settled decisions (this session)

1. **Canvas def = universal shape contract.** The mechanically-verifiable shape
   of *every* C3 entity (components and containers included) is expressed as a
   canvas definition.
2. **Instance = first-class sealed C3 entity.** A filled artifact (ADR, PRD, …)
   is a managed C3 doc: sealed, rooted under `.c3/`, in the graph.
3. **Validation = `c3 check`.** One validator, driven by canvas definitions, over
   all entities. No per-type bespoke validator, no separate `canvas check`.
4. **Canvas subsumes ADR.** An ADR is the canvas instance of type `c3-adr` — the
   most visible case of "every entity is a canvas instance," not a special path.

## The proof this is real, not aspirational

C3 today already has **three parallel shape mechanisms built from identical
primitives** — they just live in separate registries:

| Mechanism | Location | Covers | Built from |
| --- | --- | --- | --- |
| `schema.Registry` | `cli/internal/schema/schema.go:106` | component, container, context(system), ref, rule | `SectionDef` + `ColumnDef` |
| canvas built-ins | `cli/internal/schema/canvas.go:36` | c3-adr, atomic-design-change, pm-requirement, prd, user-story | `SectionDef` + `ColumnDef` |
| `ADRTemplate` | `cli/internal/schema/schema.go:46` | ADR shape | bespoke struct |

The first two are the **same data structure** keyed differently. `c3 check`'s
`validateColumn` (`check_enhanced.go:690`) already validates entities generically
against `ColumnDef`. So "make every entity shape a canvas def" is not a new
abstraction — it is **merging two registries that already speak the same
language**, and retiring the third.

## The model

```
ONE registry of CANVAS DEFINITIONS (the shape language)
  structural types:  component · container · system · ref · rule
  document types:    c3-adr · prd · user-story · atomic-design-change · pm-requirement · <project-defined>
  each = sections (text/table) + typed columns (text/date/enum/cite/check/edge/entity_id)

every C3 ENTITY is an INSTANCE of its definition
  rooted under .c3/, sealed, in the graph
  authored/scaffolded → filled (cite cells grounded via read --cite) → checked → sealed

c3 check = the ONE validator
  shape (vs definition) + evidence (citations resolve) + governance (status rules)
```

ADR is the reference implementation: it is already a sealed, checked, graphed,
status-tracked instance. The concept lifts that treatment from one type to all.

## Evidence grounding is the spine

What makes an instance *governed* rather than merely structured: every `cite`
cell asserting something about the system carries a real C3 evidence handle
(`@v<ver>:sha256:<root>` from `c3x read --cite`) or an explicit `N.A - <reason>`,
and `c3 check` verifies it resolves. A canvas instance is a *claim grounded in
the graph*. (This is the through-line from the read-cite-evidence-handles ADR.)

## Governance per type (O1 — resolved)

Everything goes through `check`; the *mechanically verifiable* shape of every
type is its canvas definition. Governance vocabulary is **per definition**: a
definition declares the status enum (and terminal states) that fit its type. ADR
keeps its rich `proposed → implemented → provisioned` set with terminal freezing;
structural entities (component/container) keep their current check semantics;
document types declare what fits. The unifying invariant is not one status list —
it is *one validator over one shape language*.

## What this dissolves

| Loose end | Resolution |
| --- | --- |
| `schema.Registry` vs canvas registry (two registries, one structure) | one canvas-definition registry |
| `ADRTemplate` bespoke shape mechanism | ADR shape = the `c3-adr` canvas definition |
| `adr-20260528-managed-adr-templates` (proposed) | subsumed — templates are canvas definitions |
| Python `CANVAS_EXPECTATIONS` duplicates the Go registry | eval scores by running `c3 check` on the instance |
| eval artifact-path heuristic | instances live at `.c3/<type>/<id>.md` |

## Open sub-decisions (leans adopted unless overridden)

- **O2 — instance home.** `.c3/<type>/<id>.md`, generalizing the existing
  `.c3/adr/`. *Adopted.*
- **O3 — instance id scheme.** Generalize `adr-YYYYMMDD-slug` →
  `<type>-YYYYMMDD-slug` for document types; structural entities keep `c3-<n>` ids.
  *Adopted.*
- **O4 — subsumption depth in code.** Thin generic canvas-def core; ADR (and
  structural-entity) semantics layered as specializations so governance never
  regresses; generalize incrementally. *Adopted; revisit when planning slice 1.*

## Implementation path (post-baseline, each slice gated by the eval loop)

**North star:** one canvas-definition registry + `c3 check` as the one validator
over all entities. **Wedge order** (smallest safe steps first):

1. **Unify the registries.** Fold `schema.Registry` entries into the canvas
   definition registry; structural entities become built-in canvas defs. No
   behavior change — prove `c3 check` output is identical before/after.
2. **Route check through canvas defs.** Make `check`'s entity validation read the
   unified registry; retire `ADRTemplate`.
3. **First-class doc instances.** Give document types (adr/prd/…) the same
   rooted-storage + seal + graph treatment ADR has; `.c3/<type>/<id>.md`.
4. **Authoring command.** Scaffold an instance from a definition (generalizes ADR
   creation).
5. **Eval refactor.** Score canvas cases via `c3 check`; retire the Python
   `CANVAS_EXPECTATIONS` duplicate.
6. **Skill wiring (move A).** Document canvas as the authoring operation
   (scaffold → ground evidence → check → seal), ADR as the canonical example;
   add `references/canvas.md`.
7. **Finalize ADRs (move C).** Promote canvas-authoring-eval-loop with real
   claude+codex baseline data; supersede managed-adr-templates.

Each slice is one change measured against the baseline by `c3-research-eval` —
keep on quality hold/improve, revert on regression.

## Implementation progress

Work happens on branch `canvas-schema-substrate` (worktree at `../c3-canvas`,
mirroring main's current working tree as base commit `eb49cdf`). Behavior oracle:
`c3 check` → total 73, no issues; all `go test` green.

Baseline established (`86b6a32`): **90% quality pass-rate** (18/20). Weak cells:
`claude × canvas_c3_adr` (canvas quality) and `skill_content_limit_adr`
(accuracy). `effective_tokens_mean` 66.8k; `token_no_go_count` 0. This is the
bar the remaining slices are gated against.

- **Slice 2 — DONE** (`0e75503`). `DefinitionFor(entityType) (Canvas, bool)` —
  one definition source keyed by entity type (O5). Any entity type's shape is
  now expressible as a canvas definition; `ForType` routes through it.
  **Verified:** `c3 check` byte-identical, go test green.
- **Slice 1 — DONE** (`f96c239`). `strict_component_docs.go` re-encoded
  component's section order, table headers, min-rows, and column enums in 4
  hardcoded maps that `Registry["component"]` already declares. Replaced with
  `deriveStrictRules(defs)`; the strict validator (`validateStrictDoc`) is now
  generic over any definition. Removed the redundant component branch in
  `write.go`'s `allowedSectionNames`. **Verified:** `c3 check` byte-identical to
  oracle, go test green.

- **Slice 5 — DONE** (`41b8f28`). Type-driven semantic validation. Component
  `Reference`/`Evidence` columns retyped to semantic types (`reference`,
  `evidence`) reusing the existing `Type` field (no new fields). Strict
  grounding checks now key off column **type** from the definition, not
  hardcoded section/column names. Removed `isReferenceColumn`,
  `isEvidenceColumn`, `requiresGroundedReference`. **Verified** byte-identical.
- **Slice 6 — DONE** (`5cbc422`). Linkage classifies citations by real entity
  type (`citationType`), not the `ref-`/`rule-` id prefix (prefix kept only as
  dangling-id fallback). **Verified** byte-identical.

## Done: the generic-first base (in main working tree + branch `canvas-schema-substrate`)

c3x now expresses today's content through canvas definitions, generic-first:
shape via one entity-type-keyed registry (`DefinitionFor`); validation
(structural + semantic) derived from / keyed off the definition; linkage
classified by real type. All verified `c3 check` byte-identical + go test green.

## The "no-overspecify" boundary — what intentionally STAYS typed

Generic-first does NOT mean zero type-literals. Only **shape + validation**
belongs in definitions. These stay type-specific *by design* (forcing them into
schema fields would be overspecifying):

| Kept typed | Why |
| --- | --- |
| layer traversal system→container→component (`adr_linkage`, `layerSection`) | C3's structural model, already mirrored by Components/Containers entity_id tables |
| ADR coverage + rule-origin governance (`check_enhanced` adr/rule branches) | genuine per-type governance semantics |
| export paths, graph shapes, cascade hints (`export`/`graph`/`cascade_hints`) | presentation / filesystem layout, not shape |
| `ClassifyDoc` frontmatter→type dispatch | the lookup that *finds* the definition |
| strict-validation gate (`if type == component`) | only component declares strict rules today; generalizes when doc-types become first-class (future) |

## Open sub-decision surfaced during slice 1

- **O5 — namespace reconciliation (BLOCKS registry unification).** Two registries
  use different key spaces: `schema.Registry` is keyed by **entity type**
  (`component`, `container`, `context`, `adr`, `ref`, `rule`, `recipe`); the
  canvas registry is keyed by **canvas id** (`c3-adr`, `prd`, `user-story`,
  `atomic-design-change`, `pm-requirement`). `adr`≈`c3-adr` but not identical;
  structural types have no canvas; doc types have no entity type. Unifying into
  one canvas-definition registry requires deciding how these namespaces map.
  Options: (a) one registry keyed by entity type, doc-canvases become entity
  types (`prd`, `user-story` are entity types too); (b) one registry keyed by
  canvas id, entity types map to a canvas id (`component`→`c3-component`); (c)
  keep two key spaces, add a bridge resolver. This is the gate before slice 2+.

## Empirical hook (from the in-flight baseline)

Claude's weakest case is `canvas_c3_adr` (the only canvas type it fails), while
`adr_create` passes — i.e. the *c3-adr canvas* is harder than the *legacy ADR
path* for the same artifact. That gap is direct evidence for slices 1–4: unifying
the shape + giving instances real check/grounding should close it. Codex data
pending at baseline completion.
