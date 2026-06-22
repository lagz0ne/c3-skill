# Design: switch-gated double-V

The model the user articulated (2026-06-18) for code↔doc consistency. This is the
spec to co-design with Codex and then build. See memory `switch-gated-double-v-model`.

## The problem (validated, concretely)
Code↔doc consistency cannot be enforced where C3 was betting on it:
- **Edit-time enforcement fails.** The original codemap bet: touch code under a glob → the
  doc is flagged. An LLM editing code just ignores that.
- **Drift-tracking is futile.** Hashing code blocks / markers in source has very limited
  success; the drift is unwinnable continuously.
- **`check` cannot see it.** Proven this session: deleting `wire.go`, `marketplace.go`,
  `internal/marketplace/**` left c3-design's `c3-104`/`c3-120` docs + code-map globs stale,
  and `c3 check --strict-codemap` stayed **fully green** (check validates docs-vs-seal, not
  docs-vs-code). The drift is invisible until a deliberate inspection.

## The model: gate the inspection at the switch
- **Plan upfront in the change unit** (the rich layer — "richer toward the running change").
- **Reads focus on the frozen** (graph/read/list answer the committed truth; `--unit` previews).
- **Apply = merge** (diff + conflict against frozen — the existing drift gate already is this).
- **The switch is the mandatory gate.** At `change apply` (staged → frozen) C3 **shows the
  content to derive** and **forces the inspection** — the LLM must do it because the switch is
  the only way to land a fact change; there is no edit-time path to ignore.
- **Codemap stays "territory"** — the cheap glob map of *where* a fact's derived code lives =
  *what to show* at the switch. Not drift-tracked, not hashed.

## What already exists (lean into it, don't reinvent)
- **Two-arm change unit** (`changeset`): `*.patch.md` = the INTERNAL (left-V) doc arm — sealed,
  drift-frozen by block cite-anchors, gated by `CheckDrift` at apply (`change.go:51`).
  `*.codemap.md` = the EXTERNAL (right-V) code arm — *not* sealed/owned ("code C3 cannot own"),
  full-replaces the fact's globs, with an optional `Base` recording pre-change globs "for the
  **derive→match view**" (`codemap_carrier.go:18-29`).
- **The switch** `RunChangeApply` (`change.go:33`): preflight gates (drift + canvas-valid +
  codemap-target-resolves + dup-carrier), then atomic apply. **No forced derivation inspection.**
- **The doc already declares the down-V** (component canvas): **Derived Materials**
  (Material / Must derive from / Allowed variance / Evidence) and **Change Safety**
  (Risk / Trigger / Detection / **Required Verification** [evidence]).
- **Overlay** `graph --unit` previews the staged unit.

## The gap = the forced up-V
The unit declares what should derive (Derived Materials) and what must be verified (Change
Safety Required Verification), and the codemap arm declares *where* the code is. **Nothing forces
the inspection** that the bound code actually derives from the *new* doc. The switch runs
mechanical gates only.

## Resolved design (Codex pass, high-effort)

1. **A third carrier: `*.inspect.md` (the up-V).** Keep it a carrier alongside `*.patch.md`
   (internal) and `*.codemap.md` (external) — do **not** fold the attestation into Change Safety /
   Derived Materials, because those are *declarations on the frozen fact* (requirements), and an
   execution record on the requirement makes the fact both requirement and proof. Shape:
   ```md
   ---
   target: c3-101
   covers:
     patches:    [{source: 01-goal.patch.md, result: sha256:…}, …]   # the doc landing it was inspected against
     codemaps:   [{source: 03.codemap.md, globs: [cli/cmd/change*.go, …]}]
   verdict: matches            # matches | updated
   inspector: agent
   ---
   ## Inspections
   | Obligation | Territory | Verdict | Evidence | Notes |
   |---|---|---|---|---|
   | Derived Materials row 1 | cli/cmd/change.go | matches | `rg RunChangeApply cli/cmd`; change.go:33 | inspected vs post-change doc |
   ```
2. **Hard-refuse at the switch.** Inspection joins the existing apply preflight (after drift /
   canvas / codemap-target, before dry-run output) — a loud warning doesn't match "no other way
   around." Frozen facts already mutate *only* through this switch.
3. **"Fresh" = fresh against the staged DOC post-state — never against code.** The key move: the
   inspect carrier must cover the unit's current patch **sources + their `result` hashes**
   (these already exist as the doc-landing check, `apply.go:334`) and the codemap **sources +
   globs**. If a covered patch/codemap changes after inspection → **stale → reject**. If *code*
   changes after inspection, the tool cannot know without hashing code — **that is the honest
   boundary** (and the boundary the model already accepts). Obligations are computed from a
   preview overlay (`WithUnitOverlay`) so they reflect the post-change fact.
4. **Anti-gaming: grounded evidence inside the resolved territory.** Reuse `isGroundedEvidence`
   as the floor, then strengthen: ≥1 evidence item is a command or path; ≥1 cited path **exists
   and falls inside the fact's resolved codemap territory**; C3 handles may use the current-handle
   validator (`check_enhanced.go:989`); bare `N.A` does not satisfy an inspection row unless the
   obligation is explicitly non-code / out-of-scope.
5. **Coverage scope = touched facts with actionable obligations.** Require inspection only when a
   touched fact (by patch or codemap carrier) has non-`N.A` rows in **Derived Materials** or
   **Change Safety.Required Verification**. A touched fact declaring code-derived material but with
   **no codemap territory** fails with a coverage-gap repair (add a `*.codemap.md` or mark the row
   out-of-scope). Laddering: untouched legacy gaps are grandfathered.
6. **Folds in #40.** view/status/rebase collapse into one `change view` that IS the inspection
   surface: patches (drift) + codemap (derive→match territory) + obligations-to-inspect +
   attestation coverage. (Or a sibling `change inspect <id>` that renders obligations + territory.)

## What this is — and isn't (the honest boundary)
Self-attestation by the same agent **prevents omission, not lying** — it is an **audit/workflow
gate, not a truth oracle.** It does not mechanically prove semantic code↔doc consistency. What it
*does*, that the rejected approaches did not: forces the inspection at the one unavoidable
chokepoint (the switch); makes the claim **explicit, grounded in real territory, falsifiable, and
reviewable** (a bare checkbox is rejected — evidence must point at code that exists in the fact's
globs); and leaves a record a human or a later second-agent pass can audit. Edit-time enforcement
was ignorable; drift-tracking was invisible; this is neither. A higher-assurance second-attestation
(independent reviewer agent) is a later opt-in policy, not v1.

**The LLM may lie — and that is fine; that is where the human in the loop matters** (maintainer's
call). The tool enforces that a fresh, territory-grounded inspection *record exists* before the
switch; it does **not** adjudicate the record's truth. The human judges truth — at `change accept`
(the one stored human judgment), reviewing the obligations + territory + the agent's attestation
that `change view`/`change inspect` surfaces. So the division is: **tool guarantees the record;
human reviews the claim.** The gate's value is converting silent, invisible drift into a loud,
grounded, reviewable artifact at the moment of the flip.

## Relationship to the `uses` fork (#42)
**Orthogonal.** This gates *changes to* the relationship model but cannot decide the ontology
(governance-citation vs component-dependency). #42 stays a separate decision.

## Build plan (Codex sequence)
1. Red tests: missing / valid / stale `*.inspect.md` carriers.
2. `changeset/inspect_carrier.go` — parser + validator (frontmatter `covers` + Inspections table).
3. `change view`/`change inspect`: compute post-state obligations via `WithUnitOverlay`; render
   obligations + resolved territory.
4. Apply preflight: inspection gate after drift/canvas/codemap-target, before dry-run output.
5. Evidence: reuse `isGroundedEvidence` floor + territory-aware inspection checks.
6. Help/dispatch; agent output stays `writeJSON` (TOON in `C3X_MODE=agent`).
7. Verify: focused change tests, `go test ./...`, local `c3local check`.

## Status — v1 shipped
**Built, tested, dogfooded (commits `cb82540`, `867bca0`, `8838000`):**
- `changeset/inspect_carrier.go` — `*.inspect.md` parser + `CoversFresh` freshness (anchored to
  the doc material hash, never code). Unit-tested (valid / rejects / fresh-stale / hash).
- `cmd/inspect.go` — `inspectionGate` (obligations from overlay → require fresh + grounded +
  territory-citing attestation; coverage-gap repair) + `RunChangeInspect` surface. Wired into
  `RunChangeApply` after the mechanical gates; `change inspect` dispatched + helped.
- Integration tests: gate **refuses without inspection**, **passes when grounded**, **refuses
  stale**, **refuses evidence outside territory** (anti-rubber-stamp).
- Dogfooded: `change inspect` on real `c3-104` computed its 3 obligations (Derived Materials +
  2 Change Safety), resolved territory, stamped material hashes — clean TOON.

**Deferred / follow-ups:**
- `change view`/`status`/`rebase` consolidation (#40) — fold the inspection coverage into the
  unified view (v1 ships `change inspect` as a sibling surface).
- `marshalMap` struct-value `%v` gap (latent; structs avoid it).
- Optional second-reviewer (independent agent) attestation policy.
- Dogfood the gate end-to-end on the c3-104/c3-120 self-doc drift (author the real change-units).
