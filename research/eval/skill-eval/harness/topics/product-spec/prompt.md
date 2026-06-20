# Topic Prompt: Build a Product Spec Hierarchy

Invent a product, then build its **product-spec hierarchy** as a traceable C3
knowledge graph. This is a PM domain, not an architecture domain: there is no
frontend/backend/database here. C3 is a general knowledge-graph contract tool —
prove it by **defining your own doc types** for product strategy and wiring a
dense traceability graph between them.

You name the product (a SaaS, a consumer app, a fintech tool — your choice). The
hierarchy you must end with:

- **Objectives** — a strategic outcome plus measurable key results (an OKR).
- **Epics** — a chunk of work that advances an objective.
- **Stories** — as-a / I-want / so-that, with acceptance criteria, that ladder
  UP: each story serves an objective, refines an epic, and satisfies a
  requirement.
- **Requirements** — a captured product/user need a story must satisfy.

The spine of this eval is **wiring**: every story ladders up. A story that
serves nothing, an objective with no work under it, or a citation that points at
a fact that does not exist is the failure. Traceability means every story
reaches an objective and every objective has at least one story — no orphan
work, no dead objective.

The spec grows under one named pressure: the **"Q3 strategy pivot."** Leadership
retires one objective and stands up a successor; the epics and stories that
laddered up to the retired objective must be re-pointed (re-wired) onto the
surviving strategy — exercise retire/supersede + re-wire, not silent deletion.

## Your task

1. **Define your own canvases** with `canvas add <id> < schema.md`. Do NOT reuse
   the built-in architecture canvases (`system`, `container`, `component`) and
   do NOT lean on the built-in `pm-requirement` / `user-story` / `prd` canvases —
   define **2–4 custom canvas types of your own** for this domain (at minimum
   `objective`, `epic`, `story`). A canvas is sealed C3 markdown: frontmatter
   (`id`, `type: canvas`, `description`) then a body with `domain:`, `sections:`
   (each `name`, `content_type: text|table`, `required:`, `purpose:`, and for
   tables `columns:` with each column's `name:` + `type:`), `reject_if:`, and
   `workorder:`. Before you write one, read a built-in non-architecture canvas to
   get the exact shape right:
   `C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh canvas read user-story`
   and `canvas read pm-requirement`. Column types you can use:
   `text`, `enum` (+ `values:`), `date`, `entity_id`, `cite`, `evidence`,
   `check`, and — the one that wires the graph — a typed **edge column**
   (`type: edge<typeA|typeB>`, or `type: reference` with `edge: <rel>` and
   `targets: typeA,typeB`). One edge column owns one relationship name.

2. **Every canvas must carry at least one edge column** that wires UP to another
   of your domain's types. At minimum:
   - `epic` advances → an `objective` (an edge column targeting `objective`).
   - `story` serves → an `objective`, refines → an `epic`, satisfies → a
     `requirement` (three separate edge columns, one per relationship).
   This is what makes the spec a graph instead of a pile of docs.

3. **Author a lean first cut, then grow it.** Start small and complete — a
   couple of objectives, the epics under them, and the stories under those,
   each wired up. Do NOT pre-build every section you can imagine. Then apply the
   **Q3 strategy pivot**: retire/supersede one objective, stand up its
   successor, and re-point the affected epics and stories onto the new strategy.
   The pivot must also force a real **canvas climb**: the lean `story` canvas
   ships without the `satisfies` → `requirement` edge; the pivot is where
   `requirement` becomes a first-class type and the `story` canvas is raised to
   add the `satisfies` edge column — then **migrate the existing stories** to
   carry the new wiring. Make the retire and the climb visible as their own
   steps, not one late cleanup.

4. **Evolve the model — grade the OKRs.** The spec defines an objective as an OKR
   with *measurable* key results, yet the lean `objective` canvas holds them as
   prose, recording no target-vs-actual — so nothing can be graded. The quarter
   closes and leadership must grade each objective. **Morph** the `objective`
   canvas: reshape the prose key-results into a typed table — `Target`, `Actual`,
   `Status` (on-track / at-risk / missed), `Confidence` — and **migrate every live
   objective** (including the Q3 successor) to the new shape **in the same unit**.
   This is a non-additive reshape (a morph, not a climb): it restructures the KR
   section, it does not just add a column, and no objective may straddle the old
   and new shapes. It is also what *justifies* the pivot — an objective graded
   *missed* is why leadership re-strategized.

5. **Add one cross-cutting recipe** — a release / roadmap doc that spans several
   epics (cites the epics shipping together). Define a custom canvas for it too
   if a built-in does not fit.

6. **Wire densely and keep it traceable.** Author facts of your custom types via
   `c3 add <type> <slug> --file` (read the canvas you defined with
   `c3 schema <type>` first) or via a change-unit. Then prove coverage:
   `c3 graph <objective-id> --direction reverse` should show real epics and
   stories laddering up to each objective, and **every story must reach an
   objective**. No orphan stories; no childless objectives (except the retired
   one, which should be terminal/superseded, not dangling).

7. **Keep `c3 check` clean.** Every citation must resolve to a real fact. Run
   `c3 check` after the lean cut, after the pivot/migration, and at the end.
   Report the exact final result; if anything is unresolved, name it rather than
   hiding it.

## Invariants — your graph must make the holes visible

These hold beyond `c3 check`. `check` proves every citation resolves; it does NOT prove
the ladder is complete — that no story serves nothing, no live objective sits without
work, that the pivot re-wired rather than deleted. State each so a reviewer can verify it
from the graph, and close every hole.

- **INV-STORY-LADDERS** — every story serves → an objective, refines → an epic, and
  (after the climb) satisfies → a requirement; all populated, all pointing at live facts.
- **INV-OBJECTIVE-WORKED** — every live objective has ≥1 epic/story laddering up to it.
  No dead objective.
- **INV-PIVOT-REWIRED** — after the Q3 pivot the retired objective is superseded
  (terminal), not deleted or left dangling, and every epic/story that laddered to it is
  re-pointed onto the successor.
- **INV-CLIMB-COMPLETE** — the `satisfies → requirement` edge was absent at rung-1 and
  added in the climb; every existing story was migrated to carry it.
- **INV-RELEASE-SPANS** — the release/roadmap recipe cites several epics shipping
  together — a fact that spans the hierarchy, not one chain.
- **INV-KR-TYPED** — after the grading morph, every live objective carries the typed
  key-result table (Target / Actual / Status / Confidence); none is left with prose KRs,
  and the values are real — every objective has a genuine Status, not a uniform placeholder.
- **INV-SUPERSEDE-LINKED** — the superseded objective carries its successor's `supersedes`
  backlink (use `c3 supersede`, which writes it), not a hand-flipped `superseded` status
  with no link — the pivot's provenance (what replaced what) must be traceable.

## Constraints

- Use local C3 only: `C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh`.
- Avoid codemap work unless absolutely required; this is a docs-only spec with
  no code, so codemap is not the focus of this eval.
- Prefer compact, concrete C3 artifacts (canvases, wired facts, graph output)
  over long prose.
- Run verification and report the exact result.
