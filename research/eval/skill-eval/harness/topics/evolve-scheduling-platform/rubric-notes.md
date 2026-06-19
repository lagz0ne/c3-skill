# Rubric notes — evolve-scheduling-platform

A high-depth, high-diversity grow+evolve scenario. Unlike the pure-growth topics
(grow-todo-app, grow-warehouse), this one forces the **model itself to morph** twice
(turns 5, 10) — so it is the topic that exercises the **evolve-unit / morph gate**.

## What good looks like, per turn
1. **Onboard** — lean rung-1; descend system → containers → components; one core recipe;
   a couple of refs/rules. `c3 check` clean. No pre-built later structure.
2. **Change-unit** — waitlist/cancellation as a change-unit (ADR + patches), wired into
   booking. Facts move, canvas unchanged.
3. **Additive climb** — component canvas raised (Concurrency Contract required); EVERY
   component migrated up in the unit; no component left below the bar.
4. **Custom canvas + wiring** — a real `policy` canvas defined in the flow (typed columns,
   not prose); components wired to policies via edge/reference columns; grounded (no
   ungrounded-reference warnings — the resolution-grounding fix should hold for `policy-…`
   ids).
5. **Non-additive MORPH (the signature turn)** — the `policy` canvas is *reshaped* (Scope →
   DataScope + ActorScope; basis → enum), and EVERY policy instance migrated to the new
   shape **in the same unit**. The win condition: the unit does not land while any policy is
   still in the old shape, and after it lands `check` is clean. A weak run does this as an
   unguarded `canvas write` + a late, separate cleanup (leaving instances transiently
   invalid) — that is the failure the morph gate exists to prevent.
6. **Retire + reparent** — tenancy plane added; execution components reparented; the
   single-clinic orchestrator retired with every orphan/dangle healed in the SAME unit
   (the overlay-aware retire gate allows reparent-then-retire in one unit).
7. **Conflict** — a real block-overlap drift on a shared fact's block, resolved via
   `change rebase` (3-way), both units preserved.
8. **Recipe** — tenant onboarding as a cross-plane recipe owned by no single component.
9. **Governance** — refs (rationale) + rules (enforceable) cited densely; stale rules
   retired.
10. **Deep re-root MORPH** — the top of the topology is reshaped toward a portfolio; the
    whole graph migrated in coherent units; `check` clean at the end. The deepest stress of
    the evolve-unit.

## Dependency — turns 5 and 10 need the evolve-unit wired
As of the morph-gate prototype (commit 7d3e24b) the gate exists and is unit-tested, but it
is **not yet wired into `change apply`** and there is **no `canvas`-scope patch file path**.
So today an agent can only morph a canvas via the unguarded `c3 canvas write` and carry the
instance-migration as ordinary patches — the morph is **not atomic with the migration** and
**not gated**. That is exactly the gap turns 5/10 are designed to expose. Score those turns
two ways: (a) did the agent keep the docs coherent through the morph at all; (b) did it have
to leave instances transiently invalid / do a late cleanup — evidence the wired evolve-unit
is needed. Re-run after wiring `morphGate` + the canvas-scope carrier to confirm the morph
turns go clean and atomic.

## Diversity / depth scoring
- **Diversity:** credit a run that genuinely uses a *different* mechanism each turn
  (change-unit ≠ climb ≠ morph ≠ retire ≠ conflict ≠ recipe), not ten variations of "add a
  component."
- **Depth:** each turn should be substantial — multiple facts, real invariants made
  reviewable — not a one-line token change. Partway-but-coherent beats racing-to-10-broken.

## Common gotchas (carried from other topics)
- Membership rows are synthesized — never hand-authored; leave parent tables header-only and
  set `parent:`.
- `--file` for any table/mermaid/code body; inline strings corrupt quoting.
- Author patch/canvas scratch OUTSIDE `.c3/` (it regenerates and drops stray files).
- A climb/morph must migrate EVERY affected fact in the unit; no fact straddles two shapes.
- Run `c3 check` after every turn and report the exact `issues[...]` line.
