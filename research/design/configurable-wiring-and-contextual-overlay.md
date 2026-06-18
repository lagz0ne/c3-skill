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

**Core goal achieved — both halves functional, tested, committed, dogfood-safe:**

- **Phase A (done, `2a28a77`):** `ColumnDef.Edge`/`Targets` + ≤1-writer rule; `content.DeclaredEdges`
  + `SyncCanvasOwnedRelationships` (the one seam) wired into `add` + `import`; seed
  `Governance.Reference` marked `edge: uses` (targets ref/rule). Authoring the column wires a
  real `uses` edge; survives rebuild; existing projects (no edge-column) untouched.
- **Phase B core (done, `2a28a77`):** `changeset.Apply` threads an in-tx `syncEdges` hook;
  `change apply` re-derives body-owned edges atomically.
- **Phase C (done, `2747377`):** orphan citation → clean `X cites Y in Section.Column which does
  not exist`. display==edge is structural (the column IS the edge).
- **Phase D core (done, `1842e5a`):** `store.WithPreviewTx` (rolled-back tx); `WithUnitOverlay`
  applies a unit's patches via the *real* `changeset.Apply` in preview (with the edge sync) and
  rolls back; `graph <id> --unit <adr>` previews staged edges under a "preview — staged, not
  applied" header (all modes: text / mermaid / reverse / json); missing unit fails loud.
- **Embeddable bodies (done, `9251c48`):** parse/render round-trip preserves mermaid/code,
  tables, images, raw HTML + `<iframe>`/embed blocks, dividers, indented code (previously dropped).
- **Honesty finalization (done, `f3b4643`+`f05db90`):** agent-mode TOON now renders nested
  struct/map slices as proper indented blocks (was a `%v` dump for graph nodes/edges, check/
  lookup/read help, marketplace, schema sections/columns); `schema` text tags an edge column
  `→ edge: <rel> (targets: …)` so "run schema, find the edge column" is real. Skill rewritten
  to the column-IS-edge model (with a legacy fallback when no `→ edge:` tag is present),
  `graph --unit`, and embeddable bodies.

**Remaining (polish / extension, not blocking the stated goal):**
- `c3 wire` front door (resolve column from canvas; frozen → `--unit` stages a block patch) —
  today you author the column directly or via a change-unit block patch.
- `change use <id>` active-unit (so `--unit` need not repeat); lens-aware `read`/`lookup`/`list`
  overlay; staged/applied delta markers on json/mermaid; block `read --cite` under overlay;
  `search` overlay (needs tx-safe store reads first).
- `c3 migrate citations` + **c3-design's own canvas adopting the `edge:` column** (dogfood):
  c3-design is still on the legacy frontmatter-`uses:` path, so `schema component` here shows no
  `→ edge:` tag. Adopting means a deliberate reseal — a product decision to confirm, not a silent change.

## Codex post-implementation review — disposition

Codex (read-only, high effort) reviewed the implementation. **Fixed:**
- **#2 (high) machine-output corruption** — `graph --unit`'s `context:` header broke JSON/TOON
  (incl. agent mode) and mermaid. Now: JSON/TOON get no prefix, mermaid gets a `%%` comment,
  human text gets the header.
- **#3/#4 (high/med) targets + custom types** — prefix-guessing collapsed all `c3-*` to one
  type and dropped custom `<type>-<slug>` ids. Now id extraction is general and target-type
  filtering uses the **actual stored type** in the sync (container-cited-as-policy correctly
  wires nothing; custom types extract).
- **#8 (med) preview-tx nesting** — `WithPreviewTx` now refuses to run inside an open tx
  (would deadlock under `MaxOpenConns(1)`).

**Deferred (noted, low reachability today):**
- **#1 (high) overlay preflight** — the overlay runs `changeset.Apply` (drift + apply errors)
  but not the full `RunChangeApply` preflight (canvas gate, codemap dup/target). A preview of
  an *invalid* unit is therefore optimistic. The common case (valid staged unit) previews
  correctly. Fix = extract the preflight and run it in `WithUnitOverlay`.
- **#5 `edge: via` double-wire**, **#6 frontmatter-`uses`-patch vs body-owned edge**,
  **#7 `write` desync** — all latent: no seed canvas declares `edge: via`; the skill cites via
  the column (not frontmatter `uses` patches); `write` is refused on frozen edge-owning facts
  (components). Fix when a non-frozen edge-owning type or a custom `edge:` rel appears.
