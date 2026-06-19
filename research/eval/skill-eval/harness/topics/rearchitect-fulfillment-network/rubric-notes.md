# Rubric Notes: Re-architect a Fulfillment Network (high bar)

This is the hard topic — it does not pass by describing logistics software. It tests
whether C3 holds the architecture coherent through **multiple canvas climbs, a
structural re-architecture, and conflicting change-units**. Score how far the agent
gets with the graph staying check-clean and coherent, not raw entity count.

Must-have evidence (the higher-bar additions over grow-todo-app):

- **Lean start, then ≥2 real canvas climbs.** Gen-1 docs are lean; Gen-2 and Gen-3
  each raise the contract under a *named* pressure and migrate **every** affected fact
  up completely. A fact left straddling two rungs (some at the old contract, some at
  the new) is the primary failure.
- **A real re-architecture, not just additions.** Gen-3 must RETIRE the obsolete
  orchestration component(s) and REPARENT the execution components under the new
  execution-plane container. Evidence the destruction gate was respected: the retire's
  orphaned children / dangling citers are handled in the *same* change-unit (reparented
  or re-cited), not left to strand.
- **Conflict resolution.** The returns/RMA feature is authored as its own change-unit
  overlapping the re-architecture; expect drift and a deliberate rebase/re-author of
  the drifted patch — not abandoning a unit or hand-editing a frozen fact.
- **Deep topology.** By the end: a planning plane and an execution plane as distinct
  containers, multiple centers, plus coordination / procurement / carrier / finance-
  returns surfaces; deep component trees; ≥2 cross-plane or cross-center recipes that
  no single component owns.
- **Governance.** Refs (rationale — e.g. event-ordering as transfer truth) AND rules
  (enforceable — idempotency-key on transfers, customs declaration on cross-border)
  authored and cited from the components they govern.
- **Verification after each generation**, reported honestly; `c3 check` clean (or the
  remaining issues named, not hidden).

Common failure modes at this bar:

- Climbing once and stopping; or adding Gen-3's planes as *new* containers without
  retiring/reparenting the Gen-1–2 orchestration (no real re-architecture).
- Leaving facts straddling rungs after a climb (partial migration).
- Retiring the orchestration and stranding its children/citers (ignoring the
  destruction gate), or working around the gate instead of handling the consequences.
- Abandoning the returns/RMA unit when it drifts, instead of resolving the conflict.
- Flattening planning↔execution into one container (no plane boundary).
- Racing to the end with a broken graph (`c3 check` not clean, issues hidden).
