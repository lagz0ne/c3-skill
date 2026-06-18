# Design: configurable wiring + the contextual change-unit overlay

Co-designed with Codex (two read-only high-effort passes) and verified against the
live code. This is the implementation spec for making C3's wiring "work as the use
case desires" and making graph operations work with the contextual change unit.

## The verified problem

Citations (a component citing the `ref-*`/`rule-*` that govern it) are **disconnected
from their display**. Verified on a fresh project:
- Authoring the ref in the **Governance reference column** → `c3 add` → **no `uses`
  edge** (graph empty; `check` says "in sync").
- `uses:` in the `--file` frontmatter → `c3 add` **strips it** → no edge.
- `c3 wire <component> <ref>` → **refused** (a component is frozen the moment it's
  created; `wire` is refused on frozen facts).
- Rebuild (`check`/import) → still no edge.

So `uses` edges (what `graph`/`check`/impact read) can today be set **only** through the
change-unit path; the reference column is **display-only**, and `check` never catches the
display≠edge mismatch. The schema already has column primitives (`reference`, `entity_id`,
`cite`, `edge<…>`) but nothing connects them to edge extraction. `c3 wire`'s `citeRow`
hardcodes `Governance/Reference` for components.

## Goal (desired behavior)

1. **One source of truth — the canvas-declared citation column.** Authoring it *is* the
   citation; it materializes the real `uses` edge. No separate hand-synced frontmatter
   field; no display that can diverge from the edge.
2. **Every authoring path wires consistently** — `add --file` stops stripping; `c3 wire`
   resolves the column from the canvas and, on a frozen fact, **stages a change-unit block
   patch** (no silent ADR); the change-unit lands column + edge together.
3. **`check` enforces it** — a citation in the column must resolve to a real entity and a
   real edge.
4. **Configurable & domain-portable** — the citation column is canvas-declared, so the same
   wiring works for any drawing.
5. **Graph works with the contextual change unit** — read/graph ops overlay the active
   unit's staged-but-unapplied patches, so you preview the post-change graph before apply.

## Core design (Codex pass 1) — column metadata drives the edge

Not name-matching. A column carries `type:` (how to parse the cell) and an `edge:` role
(what relationship to materialize) + `targets:` (allowed cited types):

```yaml
- name: Reference
  type: reference     # parse
  edge: uses          # wire this relationship
  targets: [ref, rule]
```

- **Body column = source of truth.** A shared `DeclaredEdges(def, body, store)` extractor +
  a central `syncCanvasOwnedRelationships(store, entity, def, body)` called by
  add / import / write / `change apply` (applyBlock/Insert/Whole). Frontmatter `uses:`
  demotes to legacy/migration input for canvases without a declared edge column.
- **`Governance.Reference` becomes `edge: uses`**; `Contract.Evidence` stays `type:
  reference` with no `edge:` (cites for resolution, wires nothing).
- **`c3 wire`** resolves the edge column from the source canvas (drop the hardcode);
  non-frozen → direct; frozen → require `--unit`, stage a **block patch** on the table node.
- **`check`** validates declared edge cells resolve + match DB relationships, only for
  canvases that declare an `edge:` column.
- **Migration** `c3 migrate citations --dry-run`: equal → strip frontmatter `uses:`;
  body-only → derive+export; mismatch → block with a repair plan.
- **Hardest risks:** seal churn (relationships live in frontmatter today, so suppressing
  body-owned `uses:` rewrites every seal — reseal intentionally); change-unit create-order
  (a unit creating a ref + a citer must sync edges after all body writes in the tx);
  half-migrated projects (gate enforcement on the project's resolved canvas).

## Overlay design (Codex pass 2) — the contextual unit as a read lens

- **Contextual unit is explicit, never inferred.** Resolution: `--unit <id>` (authoritative)
  → `c3x change use <id>` (disposable local state, excluded from canonical sync/check truth)
  → no fallback (reads use applied state). `change new` prints the next step; `wire --unit X`
  prints `preview: c3x graph <id> --unit X`.
- **Overlay = the real apply path in a rollback tx.** `store.WithPreviewTx(fn)` runs the
  **same** `changeset.Apply` used by commit (no separate graph interpreter — correctness
  depends on it), reads from the transient store, always rolls back. Snapshot the baseline
  **before** opening the tx (SQLite is single-conn). Loud failure on preflight drift/canvas —
  never silently fall back to applied state.
- **Lens-aware by default (with `--unit`/active unit):** `graph` (+ direction/mermaid),
  impact/reverse, `read`, `lookup`, `list` (relationship/recipe output).
- **Canonical-only unless explicit `--unit`:** `check` (CI truth; opt-in `check --unit X`),
  `index`, `sync/export`, apply/status.
- **`search` is NOT overlay-safe yet** — its semantic/index reads bypass the tx
  (`store/search.go`, `store/semantic.go`); refactor to tx `exec` first or label partial.
- **Every overlaid output says it's previewed** — text header `context: unit … (preview,
  not applied)`; JSON/TOON `context:{unit, overlay:preview, applied:false}` + a `delta`
  block; mermaid comment + styled staged edges. Compute staged markers by diffing
  baseline-snapshot vs preview. Block `read --cite` under overlay (preview hashes aren't
  canonical anchors).

## Build plan (test-gated, sequenced)

| Phase | Delivers | Depends |
|---|---|---|
| **A · Wiring core** | `ColumnDef.Edge`+`Targets` (+ ≤1-writer validation); `DeclaredEdges`; central `syncCanvasOwnedRelationships` (add/import/write/apply); mark `Governance.Reference` `edge: uses`; add/import derive from body | — |
| **B · wire + vehicle** | `wire` resolves the column; frozen → `--unit` stages a block patch; apply resyncs canvas-owned edges | A |
| **C · check + migration** | `check` enforces cell↔edge; `c3 migrate citations --dry-run`; intentional reseal | A, B |
| **D · contextual overlay** | `WithPreviewTx`/`WithUnitOverlay`; `--unit`/`change use`; lens-aware graph/read/lookup/list; preview markers; `read --cite`/`check` guarded | A |

## Status

- **Phase A started:** `ColumnDef.Edge`/`Targets` added (`internal/schema/schema.go`);
  the unambiguous-writer rule enforced at canvas parse (`internal/schema/canvas.go`
  `ValidateCanvas`); tests in `internal/schema/canvas_edge_test.go`; full `go test ./...`
  green. Next: `DeclaredEdges` + `syncCanvasOwnedRelationships` + the first red test
  (`add --file` Governance→edge), then wire it into add/import.
