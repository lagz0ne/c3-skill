# Handoff — Evolve-unit + deep eval — 2026-06-19

## TL;DR
Hardened the C3 core (retire gate, grounding, edge/hierarchy clarity), reshaped onboarding
to an emergent-canvas walkthrough, and **prototyped the evolve-unit's one new mechanic (the
morph gate)**. Next: finish wiring the evolve-unit, then run the new deep/diverse eval topic
`evolve-scheduling-platform` to prove it through the loop.

## Shipped this session (branch `changedoc-impl`, all committed, NOT released)
- `7c0428e` retire gate overlay-aware — validated by both rearchitect laps (codex + claude, issues[0], real retire/reparent/conflict).
- `fa5e34e` ground references by resolution, drop the id-shape regex — proven 8→0 on design-system.
- `78fdfb7` reserve "edge" for wiring — parentage ≠ graph edge.
- `8fe4fa0` onboarding as emergent-canvas walkthrough — validated (warehouse `issues[0]` with a real climb; qa-coverage custom-canvas/wiring clean).
- `7d3e24b` morph-gate prototype — unit-tested, **NOT wired**.

Release deferred — **the user owns release timing; do not cut a release unprompted.**

## The evolve-unit (design pinned — memory: `evolve-unit-morph-the-model`)
Progress = two units: a **change-unit** morphs FACTS within the model; an **evolve-unit**
morphs the MODEL (schema + instructions), migrating facts to fit — any direction, not just
expand. "Similar to the change-unit": the evolve-unit IS the change-unit machine with the
**canvas admitted as a patch target**. Mostly reuse; the only genuinely new mechanic is the
morph gate, and it is proven.

### Done
- `ScopeCanvas` patch (canvas as a patch target) — `cli/internal/changeset/patch.go`.
- `morphGate` (`cli/cmd/morph.go`) — overlay-aware; refuses a canvas morph unless every
  instance of the type validates against the new shape; mirrors `retireGate`. Tested
  (`cli/cmd/morph_test.go`), full suite green.

### Remaining to make it usable (all reuse — no new unknowns)
1. Wire `morphGate` into `RunChangeApply` (`cli/cmd/change.go`) — one line beside `retireGate`.
2. Apply a `canvas`-scope patch: write `.c3/canvases/<type>.md` atomically with the unit
   (see `runCanvasWrite` in `cli/cmd/canvas.go` for the write path; `applyOne` in
   `cli/internal/changeset/apply.go` for the per-patch hook).
3. `ParsePatch` accepts a `canvas`-scope patch file (Target = type, Content = new def);
   `validScopes` already includes `ScopeCanvas`.
4. **Instructions-in-canvas seam** — extend the canvas `description`/`reject_if` to carry
   how-to-fill/use, so a schema-patch morphs the guidance atomically with the shape. This is
   what keeps userland instructions natural THROUGH evolution (memory:
   `onboarding-emergent-canvas`).
5. Skill teaching — add the evolve-unit to `change.md` / SKILL.md (it reframes Act 3 from
   "climb a rung" to "morph the model"); keep "patching" at a high note (memory:
   `critical-concepts-keep-high-note`).

## The deep eval ("go further — 10 turns, diversity, depth")
New topic: `research/eval/skill-eval/harness/topics/evolve-scheduling-platform/`
- 10 diverse, deep turns (onboard / change / climb / custom-canvas / **morph** /
  retire+reparent / conflict / recipe / governance / **re-root morph**) — deeper and more
  diverse than grow-todo-app, and the only topic that exercises the **non-additive morph**
  (turns 5, 10).
- Turns 5/10 depend on the evolve-unit wiring above. Until wired they expose the gap
  (unguarded `canvas write` + late cleanup, instances transiently invalid). Re-run after
  wiring to confirm the morph turns go clean and atomic.

## Eval infra
- Harness: `research/eval/skill-eval/harness/bin/run-blindbox.sh --agent <claude|codex>
  --topic <t> --auth session --run-timeout 1800 --label <l>` (run in background; it notifies
  on completion).
- Providers: **claude + codex work** (OAuth / `~/.codex/auth.json` staged into bwrap).
  **kilo is OUT OF CREDITS** (HTTP 402 `usage_limit_exceeded`, negative balance) — paid kilo
  models (deepseek/qwen/kimi/gemini/...) unusable until the user tops up app.kilo.ai/profile;
  free kilo models are too small for these topics.
- Scoring: build a throwaway binary from HEAD and check the workspace with it —
  `cd cli && go build -tags embedmodel -buildvcs=false -o /tmp/c3x-score .` then
  `(cd <run>.workspace && C3X_MODE=agent /tmp/c3x-score --c3-dir .c3 check)`. Agent-mode
  `check` emits TOON/YAML (`total:`, `issues[K]:`), not JSON. Run `/tmp/c3x-score` directly
  (it is the binary, not the `c3x.sh` wrapper).
- Local C3 only (CLAUDE.md): never bare/global `c3x`; after a source edit, `rm` the
  gitignored `skills/c3/bin/c3x-<ver>-linux-amd64` so `c3x.sh` rebuilds. Do NOT rebuild the
  mounted binary while an eval lap is running.

## Open finding — the deeper "without fixed setup"
qa-coverage surfaced it: the mandatory genesis root (`c3-0` = system, container→component
descent) is itself architecture-shaped and doesn't fit non-architecture domains (QA/PM left
`c3-0` bodyless → 3 warnings). The full "morph toward what's needed" eventually makes the
genesis root + descent an evolve target too. Out of this round's scope (lean seed + emergent
custom); noted for later.

## Memory (cross-session context already persisted)
`critical-concepts-keep-high-note`, `onboarding-emergent-canvas`, `evolve-unit-morph-the-model`,
`grounding-by-resolution-not-shape`, `integrity-on-the-tool`, `three-act-story-canonical`.
