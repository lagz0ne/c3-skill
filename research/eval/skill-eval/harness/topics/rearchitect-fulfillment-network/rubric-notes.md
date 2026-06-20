# Rubric Notes: Re-architect a Fulfillment Network (high bar)

This is the hard topic — it does not pass by describing logistics software. It tests
whether C3 holds the architecture coherent through **multiple canvas climbs, a
structural re-architecture, and conflicting change-units**. Score how far the agent
gets with the graph staying check-clean and coherent, not raw entity count.

## Invariants — the scoring spine (with falsifiers)

`check`-clean is necessary, not sufficient: it validates structure, not whether the
re-architecture actually divided responsibilities, wired its recipes, or bound its rules.
These are the bar. The falsifier is what a reviewer finds in the finished `.c3/` to fail it.

| Invariant | Falsifier (find this → broken) |
| --- | --- |
| **INV-PLANE-PURE** | An execution-plane component whose Responsibilities/Goal still *decides* (forecast / allocate / plan waves), or a planning-plane component that *executes* (pick / pack / ship) — the monolith's mixed responsibilities were copied wholesale under a new parent instead of split. Reparented to a valid parent, so `check` is clean. |
| **INV-RECIPE-WIRED** | A cross-plane recipe whose handoffs are prose component *names* (or a `text`/`entity_id` column, not a `reference`/edge column), so `graph <recipe>` is empty — the recipe traces nothing. The seed recipe requires only `Goal`, so `check` passes. |
| **INV-RULE-GOVERNS** | A rule whose `graph <rule> --direction reverse` is empty (an orphan policy nobody is bound by), OR the transfer-idempotency rule cited from an unrelated component and not from the inter-center transfer component it names. A rule with no inbound `uses` is structurally fine, so `check` is clean. |

## Reviewer runbook — how to surface each falsifier
Run against `<run>.workspace/` with the HEAD binary (`C3X_MODE=agent /tmp/c3x-score --c3-dir .c3 <cmd>`).

| Invariant | Commands → what to look for |
| --- | --- |
| **INV-PLANE-PURE** | `read <each execution component>` → no planning verbs (forecast / allocate / plan / decide-what); `read <each planning component>` → no execution verbs (pick / pack / ship). |
| **INV-RECIPE-WIRED** | `graph <each cross-plane recipe>` → real component edges spanning both planes; empty while the prose lists a chain = decorative. |
| **INV-RULE-GOVERNS** | `graph <each rule> --direction reverse` → the governed component(s); confirm the transfer-idempotency rule is reached from the transfer/inventory component, not a carrier/other surface. |

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
