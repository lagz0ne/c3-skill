# Topic Prompt: Grow & Evolve a Multi-Tenant Scheduling Platform

Start with a small C3 project for a single-clinic appointment scheduler, then grow **and
evolve** it as the product reshapes. This is a high-depth, high-diversity test: ten turns,
each introducing ONE distinct primary pressure — and unlike a pure-growth scenario, some
turns force the **model itself to morph** (reshape a canvas, re-root the topology), not just
add. Keep the first rung lean; raise — or reshape — the contract only when a real pressure
demands it, and migrate every affected fact **completely, in the same change-unit**.

This topic is scored on more than `c3 check` passing. `check` proves the docs are
*structurally* valid — sections present, references resolve, membership synthesized. It does
NOT prove the docs are *coherent*: that a reshaped canvas left no instance in the old shape,
that one tenant's patient data is unreachable from another, that a reparented component
landed under its true owner, that a rule still names a fact that exists. Those are the
**invariants** below, and they are the real bar. Make them reviewable in the docs, and keep
them true as the model evolves.

The platform evolves through ten turns:

1. **Onboard (lean).** A single-clinic appointment scheduler. Descend the abstraction:
   system → containers (booking web, scheduling API, store) → the components that own
   request → confirm → remind, plus the one cross-container recipe the clinic is *for*.
   Add the governing refs/rules. Keep rung-1 lean; do not pre-build later structure.

2. **Change-unit — a feature.** Add waitlist + cancellation. New components, wired into
   booking. Facts move; the model holds.

3. **Additive climb — a real invariant.** Double-booking must never happen. Raise the
   component canvas to require a *Concurrency Contract* section, and migrate every
   component up to the new rung completely. The Concurrency Contract must name the actual
   mechanism that serializes a slot (a lock, a version, a single-writer) — not "we use
   transactions" — and the booking recipe must cite it. A reader must be able to see, from
   the docs alone, *why* two concurrent requests for the same slot cannot both confirm.

4. **Custom canvas + wiring + a governing rule.** Compliance arrives. Define a custom
   `policy` fact-type in the flow (its own canvas: typed columns for **scope**, basis,
   enforcement), and wire every component that touches patient data to the policies it must
   satisfy (edge columns). Author the governing rule that makes the obligation enforceable:
   *every patient-data field must cite a policy whose `scope` admits that field*. The rule
   names the policy's `scope`.

5. **Non-additive MORPH — evolve the model.** The `policy` shape was wrong: the single
   `scope` column conflates two independent things and must split into `DataScope` +
   `ActorScope`, and `basis` must become a typed enum. This is **not** an addition — it
   *reshapes* the canvas. Morph the `policy` canvas AND migrate every existing policy
   instance to the new shape **in the same unit**; the change must not land if any policy is
   left in the old shape. Remember that a morph's "affected facts" are not only the
   *instances* of the type — anything that *refers to* the old shape is affected too.

6. **Multi-tenant re-architecture (retire + reparent).** The clinic becomes a platform of
   many clinics. Introduce a tenancy plane; reparent the execution components under it,
   retire the now-obsolete single-clinic orchestrator, and handle every consequence
   (orphaned children, dangling citers) in the same change-unit. Reparent each fact under
   the owner that *actually* owns it now — not under whatever parent clears the orphan
   warning fastest. From this turn on, the data store lives under the tenancy plane: no
   documented path may let one tenant read another tenant's patient data.

7. **Concurrent conflict.** While the tenancy split is in flight, a separate
   billing-integration feature edits the **same block** of a shared fact (the scheduling
   API's Contract). Applying one unit drifts the other's cited anchor — resolve it with
   `change rebase` (3-way), not by abandoning a unit. Both features must survive.

8. **Cross-cutting recipe.** Tenant onboarding (provision → seed schedule → invite staff →
   activate) crosses the tenancy, scheduling, and identity planes and is owned by no single
   component. Author it as a recipe; each handoff must name the receiving component.

9. **Governance deepening — and a reckoning with drift.** Add the refs (rationale: why
   tenant isolation is row-level, why slot-locking is pessimistic) and rules (enforceable:
   every cross-tenant read carries a tenant assertion; every patient-data field cites a
   policy). "Enforceable" has been prose — reshape the `rule` canvas so it is *typed*: a
   required `Enforcement` column (manual / automated / blocked-in-CI) and a `subject` column
   that must cite the live fact the rule governs. Migrate every existing rule to the new
   shape in the same unit. As you do, **reconcile every rule against today's reality**: an
   earlier model morph may have moved the ground under a rule written against an old shape.
   Find any rule whose cited subject or named column no longer exists, and heal it. Retire
   any rule the re-architecture made stale.

10. **Deep re-root MORPH.** The model has outgrown "system = one product": the business now
    runs a **portfolio** of scheduling products sharing the tenancy + identity planes. The
    target is concrete: the genesis `system` fact becomes the **portfolio**, whose members are
    the shared **tenancy** and **identity** planes plus the scheduling **products** — the
    original single-clinic scheduler is now one product among them, keeping its own
    booking/scheduling components but drawing on the shared planes. Migrate the whole graph to
    that shape in coherent change-units (you choose the sequence). The whole graph must stay
    reachable from the new root, every plane re-homed exactly once, and tenant-isolation must
    survive the migration. The deepest model morph.

## Invariants — your docs must make these reviewable AND keep them true

These hold beyond `c3 check`. State each so a reviewer can verify it from the `.c3/` graph,
and keep each true as the model evolves through the turns where it applies.

- **INV-CONCURRENCY** (turn 3 on) — double-booking is *mechanistically* prevented: the
  booking component's Concurrency Contract names the real serialization mechanism for a
  slot, and the booking recipe cites it. The docs explain why two concurrent confirms cannot
  both win.
- **INV-NOSTRADDLE** (turns 5, 9, 10) — after any model morph, NO fact of the morphed type
  (and nothing that *refers to* its shape) is left in the old shape. The morph and its full
  instance-migration are ONE change-unit, not a reshape followed by a late cleanup.
- **INV-ISOLATION** (turns 6→10) — no documented path lets one tenant's patient data be read
  by another. Every patient-data component is wired to a tenant-isolation policy; every
  cross-tenant read path carries a tenant assertion. This survives the re-architecture (6)
  and the re-root (10).
- **INV-NOSTRAND** (turns 6, 10) — every retire/reparent heals every orphaned child and
  dangling citer **in the same unit**, and every reparent re-homes the fact under its true
  owner, not under whatever silences the warning.
- **INV-GOVERNANCE-LIVE** (turn 9 on) — every governing rule/ref cites only live facts and
  live shapes. A rule whose subject or named column was reshaped by an earlier morph is
  caught and re-authored — not left dangling against a column that no longer exists.
- **INV-REROOT-COHERENCE** (turn 10) — after the re-root, the whole graph is reachable from
  the new top, and every plane is re-homed exactly once (no plane both old-rooted and
  new-rooted).

## Your task
1. Initialize, then grow **and evolve** C3 docs in the isolated project, turn by turn.
2. Keep rung-1 lean; raise OR reshape the contract only under real pressure, and when you
   do, migrate every affected fact **completely in the same change-unit** — additive
   (climb) or non-additive (morph). No fact may straddle two shapes.
3. Make turns 5, 9, and 10 real **model morphs** (reshape the `policy` canvas / reshape the
   `rule` canvas / re-root the topology), each with its full instance-migration in-unit —
   not additions, not late cleanups.
4. Make turn 6 a real structural re-architecture; turn 7 a real block-overlap conflict
   resolved by rebase.
5. Hold every invariant above through the turns where it applies. State each so a reviewer
   can check it from the graph.
6. Run verification after each turn and report the exact result.

## Constraints
- Use local C3 only: `C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh`.
- Avoid codemap work unless required; it is not the focus of this eval.
- Prefer compact, concrete C3 artifacts over long prose.
- The bar is high and the turns are deep — staying check-clean, coherent, and
  invariant-true partway is worth far more than racing to turn 10 with a broken graph or a
  stranded invariant.
