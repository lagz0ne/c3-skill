# Rubric Notes: Build a Product Spec Hierarchy

The run should not pass merely by describing a product or listing OKRs in prose.
It must show that the agent used C3 as a general knowledge-graph contract tool:
defined its OWN doc types, wired a dense traceability graph between them, grew
the spec under pressure, and kept `c3 check` honest. This topic deliberately
exercises the two features the architecture topics barely touch — **custom
canvases** and **typed edge wiring** — so weight those heavily.

## Invariants — the scoring spine (with falsifiers)

`check`-clean is necessary, not sufficient: it proves citations resolve, not that the
ladder is **complete** or that the pivot re-wired instead of deleting. These are the bar.
The falsifier is what a reviewer finds in the finished workspace `.c3/` to fail the run.

| Invariant | Falsifier (find this → broken) |
| --- | --- |
| **INV-STORY-LADDERS** | A story missing a ladder edge (serves / refines / satisfies), or citing a retired/nonexistent fact. |
| **INV-OBJECTIVE-WORKED** | `graph <live objective> --direction reverse` is empty — a dead objective with no work under it. |
| **INV-PIVOT-REWIRED** | An epic/story still serving the retired objective; OR the retired objective deleted/dangling instead of superseded (terminal). |
| **INV-CLIMB-COMPLETE** | A post-climb story with no `satisfies` edge; OR the edge present from the very first cut (no climb + migration shown). |
| **INV-RELEASE-SPANS** | A release/roadmap recipe citing ≤1 epic — it spans nothing. |
| **INV-KR-TYPED** | After the grading morph: an objective still carrying prose key-results (not migrated to the typed Target/Actual/Status/Confidence table), OR a cosmetic morph — every objective's Status is a uniform placeholder, so the reshape carried no grading. The morph gate keeps `check` clean either way. |
| **INV-SUPERSEDE-LINKED** | A superseded objective with `status: superseded` but `graph <retired-objective> --direction reverse` shows NO `supersedes` edge from its successor — the agent hand-flipped status (a frontmatter patch) instead of `c3 supersede`, so the pivot's provenance is gone. `check` passes (status is a legal value). |

## Reviewer runbook — how to surface each falsifier
Run against `<run>.workspace/` with the HEAD binary (`C3X_MODE=agent /tmp/c3x-score --c3-dir .c3 <cmd>`).

| Invariant | Commands → what to look for |
| --- | --- |
| **INV-STORY-LADDERS** | `read <each story>` → serves/refines/satisfies filled; `graph <story>` → three edges to live facts. |
| **INV-OBJECTIVE-WORKED** | `graph <each live objective> --direction reverse` → ≥1 epic/story. |
| **INV-PIVOT-REWIRED** | `read <retired objective>` → status superseded/terminal; `graph <retired objective> --direction reverse` → empty (all re-pointed); the successor id appears in the re-pointed facts. |
| **INV-CLIMB-COMPLETE** | `change view`/`change status <climb-adr>` → the `satisfies` edge added in a climb unit; `read <each story>` → satisfies present. |
| **INV-RELEASE-SPANS** | `graph <release-recipe>` → edges to ≥2 epics. |
| **INV-KR-TYPED** | `canvas read objective` → the KR section is the typed table; `read <each live objective>` → typed KRs with a real Status each, none left prose; `change status <morph-adr>` → reshape + migration in ONE unit. |
| **INV-SUPERSEDE-LINKED** | `read <retired objective>` → status superseded; `graph <retired objective> --direction reverse` → a `supersedes` edge from the successor; if absent, it was a bare status flip. |

## Must-have evidence

- **Local C3 command evidence**, not bare/global `c3x` (the
  `/opt/c3/skills/c3/bin/c3x.sh` wrapper).
- **Custom canvases, not built-ins.** The agent ran `canvas add <id>` to define
  its OWN types (`objective`, `epic`, `story`, and ideally a `requirement` and a
  release/roadmap recipe). It did NOT solve the task with `system` / `container`
  / `component`, and it did NOT just reuse the built-in `pm-requirement` /
  `user-story` / `prd` canvases. Defining custom types is the central evidence —
  if the run used only built-ins, it fails the topic.
- **At least one edge column per custom canvas, wiring to another domain type.**
  Concretely: `epic` advances → `objective`; `story` serves → `objective`,
  refines → `epic`, satisfies → `requirement` — each as its own typed edge
  column (`edge<...>` or `reference` + `edge:` + `targets:`), one writer per
  relationship. A canvas with no edge column (a flat doc that cites nothing) does
  not count.
- **Dense wiring with every citation resolving.** Multiple objectives, epics
  under them, and several stories, each carrying real ids in its edge columns —
  not one token example. `c3 check` confirms every citation resolves (no orphan
  citation / unresolved-edge errors).
- **Traceability completeness, no orphans.** Every story ladders up to an
  objective (serves + refines + satisfies populated, pointing at facts that
  exist); every live objective has at least one story/epic under it.
  `c3 graph <objective-id> --direction reverse` shows the epics and stories that
  ladder up. A story that serves nothing, or a live objective with no work, is
  the failure the rubric is checking for.
- **A real growth / climb step with migration.** A lean first cut, then the
  named **Q3 strategy pivot**: an objective retired/superseded with a successor
  stood up, AND the `story` canvas raised to add the `satisfies` → `requirement`
  edge (introducing `requirement` as a first-class type). Affected epics/stories
  were re-pointed onto the new strategy and existing stories migrated to carry
  the new edge — and `c3 check` was re-run after migration. The retire and the
  climb should be distinct, visible steps.
- **Cross-cutting recipe.** A release/roadmap doc citing several epics that ship
  together (a fact that spans the hierarchy, not just laddering one chain).
- **Check-clean honesty.** `c3 check` run after the lean cut, after the
  migration, and at the end; the final exact result reported. If something is
  unresolved, it is named, not hidden behind a "looks good" claim.

## Common failure modes

- **Reused built-in canvases** (`component`, or the built-in `pm-requirement` /
  `user-story` / `prd`) instead of defining custom types — misses the whole
  point of the topic.
- **Sparse or fake wiring** — edge columns left blank, filled with prose instead
  of ids, or only one example wired while the rest dangle.
- **Orphan work** — a story that serves no objective, or a live objective with
  no story/epic laddering up to it; traceability claimed but `graph --direction
  reverse` shows gaps.
- **Citations that do not resolve** — an edge points at an id that was never
  created (or at the wrong target type), and `c3 check` is not clean (or was not
  re-run after the change).
- **No real growth** — builds the full hierarchy in one pass, or "mentions" the
  pivot without actually retiring/superseding an objective and re-pointing its
  epics/stories; or raises no canvas (the `satisfies` edge present from the
  start, so there is no climb + migration to show).
- **Silent deletion instead of retire/supersede** — the retired objective just
  vanishes rather than being marked terminal/superseded, leaving the history of
  the pivot untraceable.
- **Prose OKRs with no C3 trail** — objectives/epics/stories described in
  Markdown narrative rather than authored as wired C3 facts of custom types.
- **Skipping verification** — claims a clean spec without `c3 check` output, or
  runs check before the migration and calls the final state verified.
- **Codemap busywork** — chasing codemap/Derived-Materials when there is no code;
  it should be absent or explicitly deferred.
