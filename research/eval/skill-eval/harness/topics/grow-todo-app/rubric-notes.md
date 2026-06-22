# Rubric Notes: Grow a Collaborative TODO Web App

The run should not pass merely by describing TODO-app software. It must show that
the agent can use C3 to grow architectural documentation as adjacent product
complexity accrues.

Must-have evidence:

- Local C3 command evidence, not bare/global `c3x`.
- A system with at least a web frontend, an API/backend, and a realtime-sync
  surface; an auth/identity boundary and persistence as the product warrants.
- Components under multiple containers, with ownership boundaries.
- Cross-container flows or recipes for at least two complex operations, such as
  realtime sync reconciliation, reconnect/replay de-duplication, shared-list
  invite-to-edit authority, concurrent-edit resolution, or per-user theme precedence.
- Domain concepts such as sync truth/reconciliation, list membership and authority,
  replay idempotency, task ownership/privacy, or theme precedence.
- Explicit growth step: lean initial docs, a real pressure for a richer rung, a
  raised contract/canvas or richer required sections, complete migration of affected
  facts, and verification after migration.
- Migration/document gardening as single-user → shared-list, private → membership,
  single-owner → concurrent-edit, and system-theme → per-user assumptions break —
  each break a visible step, plus retiring stale facts and fixing failed checks.
- Codemap is absent or explicitly deferred (the product has no code yet — docs-first).

Common failure modes:

- Building the full multi-stage system at once instead of growing it under pressure.
- Treating migration as one late cleanup rather than per-assumption-break steps.
- Frontend-state/collaboration pressure left implicit (no doc surface for realtime
  truth, membership authority, replay, or theme precedence).
- Skipping verification, or reporting a self-claimed result not backed by `c3 check`.
- Pre-building later-stage sections in the first rung (over-heavy initial canvas).
- Bare/global `c3x` instead of the local skill binary.
- Falling into codemap/Derived-Materials busywork when there is no code yet.
