# Topic Prompt: Re-architect a Fulfillment Network Across Three Generations

Grow C3 docs for an order-fulfillment platform that is **re-architected twice** as it
scales from one warehouse to a planning-driven multi-center network. This is a
high-complexity test: the docs must stay coherent and check-clean across multiple
canvas climbs, a structural re-architecture that retires and re-homes facts, and
overlapping change-units that conflict. Keep the first rung lean; raise the contract
only when a real pressure demands it.

The platform evolves through three generations:

## Generation 1 — single fulfillment center (lean rung-1)
One warehouse. The core flow: receive → store → pick → pack → ship. Model it lean —
a couple of containers, the components that own each step, and the one or two
cross-container operations the center is *for*. Do NOT pre-build later-generation
structure.

## Generation 2 — multi-center network (first climb + new topology)
The business opens more centers. New pressures the docs must make reviewable:
- **Cross-center inventory truth** — which center owns the authoritative on-hand when
  stock moves between centers.
- **Inter-center transfer reconciliation** — a transfer must never double-count or
  lose stock in flight.
- **Carrier integration and procurement** — inbound supply and outbound carriers.
This is a real rise in complexity *level*: climb the relevant canvas (add the required
section(s) the new invariants need — e.g. a cross-center flow / safety contract) and
**migrate every affected fact up to the new rung, completely**. Add the new containers
(network coordination, procurement, carrier) and their components.

## Generation 3 — planning/execution split (re-architecture: retire + reparent + a conflicting change)
Scale forces a control-plane / data-plane split. Introduce a **planning plane**
(demand forecasting, allocation, wave planning) that decides *what* each center does;
the centers become a pure **execution plane** that does it. This re-architecture must:
- **Retire** the now-obsolete monolithic orchestration component(s) from Gen 1–2 that
  mixed planning and execution — and handle every consequence the retire surfaces
  (orphaned children, dangling citers) **in the same change-unit**.
- **Reparent** the execution components under the new execution-plane container, and
  re-home any cross-cutting concern to the plane that now owns it.
- Add the planning-plane container and its components, with a cross-plane recipe
  (forecast → allocation → wave → execution) that no single component owns.
- **Concurrency:** while the re-architecture is in flight, a *separate* feature —
  **returns/RMA** (a returned item flows back into inventory) — is also being authored
  as its own change-unit. The two units touch overlapping facts; expect drift and
  resolve the conflict (re-author the drifted patch against the moved fact) rather than
  abandoning either unit.

## Throughout — governance and gardening
As generations land, add the **refs** (rationale: e.g. why event-ordering is the
transfer-truth source) and **rules** (enforceable: e.g. "every inter-center transfer
carries an idempotency key", "every cross-border shipment has a customs declaration")
that govern the components, and cite them. Retire stale facts as assumptions break;
keep `c3 check` clean after each generation.

## Your task
1. Initialize and grow C3 docs in the isolated project, generation by generation.
2. Keep Gen-1 lean and complete; do not pre-build later generations.
3. Climb the canvas only under real pressure, and when you climb, **migrate every
   affected fact completely** — no fact may straddle two rungs.
4. Make the Gen-3 re-architecture a real structural change: retire the obsolete
   orchestration, reparent execution components, add the planning plane — each
   consequence handled in its change-unit so the graph never strands.
5. Author the returns/RMA feature as a concurrent change-unit and resolve the conflict
   it creates with the re-architecture.
6. End with a deep network: at least planning, execution (multiple centers),
   coordination, procurement, carrier, and finance/returns surfaces; deep component
   trees; cross-plane and cross-center recipes; governing refs and rules.
7. Run verification after each generation and report the exact result.

## Constraints
- Use local C3 only: `C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh`.
- Avoid codemap work unless absolutely required; it is not the focus of this eval.
- Prefer compact, concrete C3 artifacts over long prose.
- The bar is high — getting *partway* with the docs staying check-clean and coherent
  is more valuable than racing to the end with a broken graph.
