# Rubric notes — evolve-scheduling-platform

A high-depth, high-diversity grow+evolve scenario. Unlike the pure-growth topics
(grow-todo-app, grow-warehouse), this one forces the **model itself to morph** three times
(turns 5, 9, 10) and plants an **emergent drift** the agent must detect unprompted — so it
is the topic that exercises the **evolve-unit / morph gate**, and it is scored on
**coherence the `check` command cannot see**, not just on `check` passing.

`check` validates structure (sections present, references resolve, membership synthesized).
The depth of this topic is the **invariants** below — semantic properties a reviewer
verifies by reading the `.c3/` graph. A run that is `check`-clean but violates an invariant
has missed the point of the topic. Score the invariants first; they are the spine.

## Invariants — the scoring spine (each with its falsifier)

For each, the falsifier is what a reviewer finds in the workspace `.c3/` to fail the run.

| Invariant | Holds at | Falsifier (find this → the invariant is broken) |
| --- | --- | --- |
| **INV-CONCURRENCY** | turn 3 → end | Concurrency Contract names no real slot-serialization mechanism ("we use transactions", "it locks" — *what* lock, on *what*?), OR the booking recipe doesn't cite the contract, OR no doc explains why two concurrent confirms can't both win. |
| **INV-NOSTRADDLE** | turns 5, 9, 10 | After a morph turn, any instance of the morphed type still carries an old-shape column or lacks a new-shape required one; OR the split is *cosmetic* (e.g. `ActorScope` is a verbatim copy of the old `scope`, or uniformly `N.A` across all policies); OR the morph and its instance-migration land in **two** units (reshape, then a late cleanup) instead of one. |
| **INV-ISOLATION** | turns 6 → 10 | A patient-data component with no wired tenant-isolation-policy edge; a cross-tenant read path (recipe/component) citing no tenant-assertion rule; OR after the re-root the data store is re-homed off the tenancy plane. |
| **INV-NOSTRAND** | turns 6, 10 | An orphan (child of the retired orchestrator with no live parent) or a dangle (ref/rule/recipe still citing the retired orchestrator) survives the unit; OR a reparent lands a fact under a structurally-valid but **wrong** owner (a scheduling component re-homed under the identity plane just to clear the orphan warning). |
| **INV-GOVERNANCE-LIVE** | turn 9 → end | A rule whose `subject` or named column references a fact/column an earlier morph removed — **specifically the turn-4 policy-`scope` rule surviving turn 5 unchanged** (see drift answer key); OR a ref citing a retired fact. |
| **INV-REROOT-COHERENCE** | turn 10 | A plane unreachable from the new portfolio root; a plane parented under **both** the old system and the new root (a topology straddle); OR the system fact left in its old single-product shape. |

## Reviewer runbook — how to surface each falsifier

Run these against the finished workspace (`<run>.workspace/`); `c3` is the throwaway
HEAD binary, `C3X_MODE=agent /tmp/c3x-score --c3-dir .c3 <cmd>`. The graph-aware commands
do the tracing; a plain `grep` over `.c3/` catches textual residue a morph should have
removed. By-eye, but systematic — any reviewer runs the same checks.

| Invariant | Commands → what to look for |
| --- | --- |
| **INV-CONCURRENCY** | `read <booking-component>` → the Concurrency Contract names a concrete mechanism, not "transactions". `graph <booking-recipe>` → it cites that contract. |
| **INV-NOSTRADDLE** | `canvas read policy` (and `rule`) → the live shape. `read <each policy>` → carries `DataScope`+`ActorScope`, never `scope`; the second column isn't a copy of the first. `change status` / `change view <morph-adr>` → the reshape AND the instance-migration are in ONE accepted unit, not two. `grep -rn 'scope' .c3/` → no policy body still on the old column. |
| **INV-ISOLATION** | `graph <each patient-data component>` → an edge to a tenant-isolation policy. `read <cross-tenant recipe/component>` → cites the tenant-assertion rule. After turn 10, `graph <data-store>` → still under the tenancy plane. |
| **INV-NOSTRAND** | `check` → zero orphan/dangle warnings. `list` → the retired orchestrator is gone; every former child has a live parent. `graph <each reparented component>` → its new parent actually owns its responsibility (semantic — not just warning-silenced). |
| **INV-GOVERNANCE-LIVE** | `grep -rn 'scope' .c3/changes .c3/*rule*` and `search "policy scope"` → no rule body still names the morphed-away `scope`. `read <each rule>` → `subject` cites a live fact (resolve it with `read`). |
| **INV-REROOT-COHERENCE** | `graph <portfolio root>` → every plane reachable from it. `list` → no plane appears under both the old system and the new root. `read <system fact>` → its shape was actually restructured, not left as-was. |

## The morph map (this topic has three non-additive morphs + one additive climb)

- **Turn 3 — additive climb** (NOT a morph): Concurrency Contract becomes a required section.
  Every component migrated up. A section is *added*; nothing is reshaped.
- **Turn 5 — `policy` data-shape morph:** `scope` → `DataScope` + `ActorScope`; `basis` → typed
  enum. Every policy instance migrated **and** every fact that *refers to* the `scope` column
  (the turn-4 rule) re-authored — in one unit.
- **Turn 9 — `rule` shape morph:** `Enforcement` (manual/automated/blocked-in-CI) and `subject`
  become required typed columns. Every rule migrated in one unit. This turn ALSO carries the
  drift reconciliation (below).
- **Turn 10 — re-root topology morph:** system → portfolio; shared planes re-homed; the system
  fact's own shape restructured. The whole graph migrated in coherent units.

A strong run keeps each turn's **primary** mechanism distinct (climb ≠ data-morph ≠
shape-morph ≠ re-root) — three different kinds of evolution, not "morph something" thrice.

## The adversarial drift answer key

The drift is **emergent, not injected** (the harness builds from a clean init): the
scenario's own sequence creates a latent incoherence the agent must notice.

- **Planted (turn 4):** the governing rule *"every patient-data field must cite a policy whose
  `scope` admits that field"* — it names the policy `scope` column.
- **Trigger (turn 5):** the `policy` morph removes `scope` (splits it into `DataScope` +
  `ActorScope`). The turn-4 rule now references a column that no longer exists.
- **Best (full credit):** the agent catches it **at turn 5** — treats the rule as an affected
  fact of the morph (a morph's affected facts include what *refers to* the shape, which the
  turn-5 prompt hints) and re-authors it inside the morph unit. INV-NOSTRADDLE in its full
  sense.
- **Acceptable (partial):** the agent misses it at 5 but catches it at **turn 9's**
  reconciliation ("find any rule whose cited subject or named column no longer exists").
- **Fail:** the stale rule survives to turn 10 still naming `scope`. INV-GOVERNANCE-LIVE
  broken. This is the dominant failure the turn is designed to expose: an agent whose notion
  of "migrate every affected fact" covers *instances of* the type but not facts that
  *reference* it.

## What good looks like, per turn (brief)

1. **Onboard** — lean rung-1; descend system → containers → components; one core recipe; a
   couple of refs/rules. `check` clean.
2. **Change-unit** — waitlist/cancellation as a change-unit (ADR + patches), wired into
   booking. Facts move, canvas unchanged.
3. **Additive climb** — Concurrency Contract required; EVERY component migrated up; the
   contract names a real mechanism and the booking recipe cites it (INV-CONCURRENCY).
4. **Custom canvas + wiring + rule** — a typed `policy` canvas defined in the flow; components
   wired to policies (grounded — the resolution-grounding fix should hold for `policy-…` ids);
   the turn-4 `scope`-naming rule authored (this plants the drift).
5. **Data-shape morph** — `policy` reshaped, every instance AND the turn-4 rule migrated in one
   unit; lands only when nothing is left in the old shape (INV-NOSTRADDLE + early drift catch).
6. **Retire + reparent** — tenancy plane added; execution components reparented to their true
   owners; the single-clinic orchestrator retired with every orphan/dangle healed in the SAME
   unit; data store under the tenancy plane (INV-NOSTRAND + INV-ISOLATION begins).
7. **Conflict** — a real block-overlap drift on the scheduling API Contract, resolved via
   `change rebase` (3-way), both units preserved.
8. **Recipe** — tenant onboarding as a cross-plane recipe owned by no single component; each
   handoff names the receiving component.
9. **Governance morph + drift reckoning** — `rule` canvas reshaped (typed `Enforcement` +
   `subject`), every rule migrated in-unit; refs/rules cited densely; the turn-4 stale rule
   caught and healed if not already (INV-GOVERNANCE-LIVE); stale rules retired.
10. **Re-root morph** — topology reshaped toward a portfolio; the whole graph migrated in
    coherent units; reachable from the new root; isolation survives the migration
    (INV-REROOT-COHERENCE + INV-ISOLATION holds through). The deepest stress.

## The evolve-unit — the trigger and its output (turns 5, 9, 10)
The evolve-unit is **wired** (11.2.0): a `canvas`-scope patch morphs a type, the morph gate
refuses unless every instance is valid against the new shape once the unit's migrations apply,
and the canvas file + the facts flip atomically. So the morph turns should now land **clean** —
and the test is whether the agent reaches them the obvious way, with the trigger connected to a
clear output:

- **The trigger** (obvious): a `canvas`-scope `*.patch.md` — target = the type, body = the new
  canvas def — in the SAME change-unit as the instance-migration patches. The skill teaches it
  (`change.md` §Morphing the model).
- **The output** (clear, stated by the turn): the morphed canvas on disk + every instance valid
  against the new shape, no straddle, in one flip.

Score the morph turns on whether the trigger met that output:
- **Full credit** — a `canvas`-scope patch + the migrations in ONE unit; `change apply` reports
  `morphed canvas <type>`, `change status` reads applied, the no-straddle gate held (every
  instance migrated, *including referrers* like the turn-4 rule). Obvious trigger, clean output.
- **Partial** — the right end state reached the wrong way: an unguarded `c3 canvas write` + the
  migrations as a separate/late step (not atomic, not gated). The output may be correct, but the
  agent did not use the evolve-unit — the request never connected to the gated path.
- **Fail** — an instance left straddling two shapes (INV-NOSTRADDLE) or a stale referrer
  (INV-GOVERNANCE-LIVE): the morph never landed its output.

A capable agent given a clear target shape (the turn states it) and the wired mechanism (the
skill teaches it) should find this straightforward. The eval measures whether the
request→output path is as obvious in practice as it is by design.

## Diversity / depth scoring
- **Diversity:** credit a run that genuinely uses a *different* mechanism each turn
  (change-unit ≠ climb ≠ data-morph ≠ retire ≠ conflict ≠ recipe ≠ shape-morph ≠ re-root), not
  ten variations of "add a component" or "morph something."
- **Depth:** each turn should be substantial — multiple facts, real invariants made
  reviewable — not a one-line token change. An invariant held through a hard turn is worth
  more than reaching turn 10. Partway-but-coherent-and-invariant-true beats racing-to-10-broken.

## Common gotchas (carried from other topics)
- Membership rows are synthesized — never hand-authored; leave parent tables header-only and
  set `parent:`.
- `--file` for any table/mermaid/code body; inline strings corrupt quoting.
- Author patch/canvas scratch OUTSIDE `.c3/` (it regenerates and drops stray files).
- A climb/morph must migrate EVERY affected fact in the unit — including facts that *reference*
  a reshaped column, not only instances of the reshaped type. No fact straddles two shapes.
- Run `c3 check` after every turn and report the exact `issues[...]` line — but remember
  `check`-clean is necessary, not sufficient: the invariants above are the bar.
