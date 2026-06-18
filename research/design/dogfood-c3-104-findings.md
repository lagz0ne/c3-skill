# Dogfood findings — fixing c3-104 through the real change-unit flow

Driving an actual fix for `c3-104` ("wiring", stale after the `wire` command was
removed) end-to-end surfaced concrete friction in the change-unit flow. Findings are
the deliverable; two are fixed, the rest are open.

## Fixed this session
- **#3 ADR weight** → laddered the adr canvas (`1cbaca4`). Lean required core
  (Goal, Context, Decision, Affected Topology, Verification); the 7 work-order
  sections are optional and climb in for weighty decisions. A lean ADR now validates.
- **#2 frontmatter fields uneditable** → frontmatter patches gained boundary/category/date
  (`a85c385`), parity with `set` (which is refused on frozen facts).

## Open findings
1. **Code↔doc drift is invisible to `check`.** `wire.go`/`marketplace.go`/`internal/marketplace/**`
   were deleted; c3-104/c3-120 docs + code-map globs went stale and `check --strict-codemap`
   stayed green. This is exactly what the switch-gated double-V addresses (catch it *at the change*).
2. **`summary` has no edit path.** It lives in the entity Metadata JSON; no `set` case and no
   frontmatter-patch field. So a frozen fact's stale `summary` cannot be corrected through any
   flow. Either make it a first-class editable field (set + frontmatter patch) or derive/deprecate it.
3. **Retiring a *referenced* component is a multi-fact change-unit.** `applyRetire` drops the
   fact's *outgoing* edges, not incoming ones. c3-104 was referenced by c3-111 + c3-113 (`uses`)
   and listed in the c3-1 README — so a clean retire needs: retire patch + a frontmatter re-edge
   per referencer + a parent-README block patch. The references don't auto-clean.
4. **Affected Topology Evidence demands a full cite-with-snippet.** Each affected-entity row's
   Evidence must be `<entity>#n<node>@v<ver>:sha256:<hash> "exact snippet"` — heavyweight to
   author for a multi-entity change, and impractical when the relevant section is a large table
   (a container's Components table). Consider accepting a bare cite handle (no snippet) here.
5. **Table-row removal from a frozen fact via a block patch is unclear/broken.** A `block` patch
   on the c3-1 README Components section (full table minus one row, columns matching the canvas)
   was rejected at apply: `merged c3-1 violates its canvas: invalid required table: Components`.
   Root cause not fully diagnosed — either the block-patch-on-table path doesn't reconstruct the
   table from markdown content, or the cited node / content shape was wrong. **This blocked the
   c3-104 retire** (the other 3 patches — two frontmatter re-edges + the retire — passed preflight).
   Needs focused debugging; editing a table row is a common, core need.

## Status of the c3-104 retire — DONE (`da7aeae`)
Retired end-to-end (lean ADR + 4 patches: retire c3-104, re-edge c3-111/c3-113 uses→c3-106, delete the c3-1 README row). Required three mechanism fixes (`baedf0a`): empty-block-deletes-node (was unimplemented), MergedBody mirrors applyBlock, and the inspection gate triggers only on contract changes (not frontmatter re-edges). New edge cases noted: a retire cascades to ALL referencers (recipe-validation-system's sources auto-dropped c3-104), and a retire ADR self-cites the retired entity so its auto-done latch can't fire.

## (original blocker analysis)
Staged and dry-run only (nothing applied; c3-design facts untouched, then cleaned up). The lean
ADR + the 2 re-edge patches + the retire patch are correct and pass preflight; only the parent-
README table-row block patch (finding #5) blocks it. Resume by resolving #5, then re-author the
4-patch unit (retire c3-104; re-edge c3-111/c3-113 `uses` c3-104→c3-106; drop the c3-1 README row).
