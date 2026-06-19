# Topic Prompt: Grow & Evolve a Multi-Tenant Scheduling Platform

Start with a small C3 project for a single-clinic appointment scheduler, then grow **and
evolve** it as the product reshapes. This is a high-depth, high-diversity test: ten turns,
each introducing ONE distinct primary pressure — and unlike a pure-growth scenario, some
turns force the **model itself to morph** (reshape a canvas, re-root the topology), not just
add. Keep the first rung lean; raise — or reshape — the contract only when a real pressure
demands it, and migrate every affected fact **completely, in the same change-unit**.

The platform evolves through ten turns:

1. **Onboard (lean).** A single-clinic appointment scheduler. Descend the abstraction:
   system → containers (booking web, scheduling API, store) → the components that own
   request → confirm → remind, plus the one cross-container recipe the clinic is *for*.
   Add the governing refs/rules. Keep rung-1 lean; do not pre-build later structure.

2. **Change-unit — a feature.** Add waitlist + cancellation. New components, wired into
   booking. Facts move; the model holds.

3. **Additive climb.** Double-booking must never happen — a real invariant. Raise the
   component canvas to require a *Concurrency Contract* section, and migrate every
   component up to the new rung completely.

4. **Custom canvas + wiring.** Compliance arrives. Define a custom `policy` fact-type in
   the flow (its own canvas: typed columns for scope, basis, enforcement), and wire every
   component that touches patient data to the policies it must satisfy (edge columns).

5. **Non-additive MORPH — evolve the model.** The `policy` shape was wrong: the single
   `Scope` column must split into `DataScope` + `ActorScope`, and `basis` must become a
   typed enum. This is **not** an addition — it *reshapes* the canvas. Morph the `policy`
   canvas AND migrate every existing policy instance to the new shape **in the same unit**;
   the change must not land if any policy is left in the old shape.

6. **Multi-tenant re-architecture (retire + reparent).** The clinic becomes a platform of
   many clinics. Introduce a tenancy plane; reparent the execution components under it,
   retire the now-obsolete single-clinic orchestrator, and handle every consequence
   (orphaned children, dangling citers) in the same change-unit.

7. **Concurrent conflict.** While the tenancy split is in flight, a separate
   billing-integration feature edits the **same block** of a shared fact (the scheduling
   API's Contract). Applying one unit drifts the other's cited anchor — resolve it with
   `change rebase` (3-way), not by abandoning a unit.

8. **Cross-cutting recipe.** Tenant onboarding (provision → seed schedule → invite staff →
   activate) crosses the tenancy, scheduling, and identity planes and is owned by no single
   component. Author it as a recipe.

9. **Governance deepening.** Add the refs (rationale: why tenant isolation is row-level)
   and rules (enforceable: every cross-tenant read carries a tenant assertion; every
   patient-data field cites a policy). Cite them densely; retire any rule the
   re-architecture made stale.

10. **Deep re-root MORPH.** The model has outgrown "system = one product": the business now
    runs a portfolio of scheduling products sharing the tenancy + identity planes. Reshape
    the top of the topology toward that — re-home the planes, restructure the system fact's
    own shape — migrating the whole graph in coherent change-units. The deepest model morph.

## Pressures the docs must make reviewable
Double-booking prevention (concurrency); who may read/write across tenants (authority +
isolation); how a reshaped `policy` never leaves an instance in the old shape; how the
re-architecture never strands a child or citer; how a private patient record is never
exposed across a tenancy or migration boundary; how the model can re-root without losing
coherence.

## Your task
1. Initialize, then grow **and evolve** C3 docs in the isolated project, turn by turn.
2. Keep rung-1 lean; raise OR reshape the contract only under real pressure, and when you
   do, migrate every affected fact **completely in the same change-unit** — additive
   (climb) or non-additive (morph). No fact may straddle two shapes.
3. Make turns 5 and 10 real **model morphs** (reshape a canvas / re-root the topology),
   each with its full instance-migration in-unit — not additions, not late cleanups.
4. Make turn 6 a real structural re-architecture; turn 7 a real block-overlap conflict
   resolved by rebase.
5. End deep: tenancy, scheduling, identity, billing surfaces; deep component trees;
   cross-plane recipes; governing refs + rules; a `policy` custom canvas wired across the
   graph.
6. Run verification after each turn and report the exact result.

## Constraints
- Use local C3 only: `C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh`.
- Avoid codemap work unless required; it is not the focus of this eval.
- Prefer compact, concrete C3 artifacts over long prose.
- The bar is high and the turns are deep — staying check-clean and coherent partway is
  worth more than racing to turn 10 with a broken graph.
