# Rubric Notes: Grow Warehouse Inventory System

The run should not pass merely by describing warehouse software. It must show
that the agent can use C3 to grow architectural documentation.

Must-have evidence:

- Local C3 command evidence, not bare/global `c3x`.
- A system with at least frontend, backend, integration, and database
  containers.
- Components under multiple containers, with ownership boundaries.
- Cross-container flows or recipes for at least two complex operations, such as
  reservation-to-pick, receiving-to-putaway, cycle count adjustment, return
  processing, or outbound shipment.
- Domain concepts such as inventory state, lot/location, reservation,
  correction/audit, integration event, or stock movement.
- Explicit growth step: lean initial docs, pressure for a richer rung, raised
  contract/canvas or richer required sections, migration of affected docs, and
  verification after migration.
- Migration/document gardening work for adding features, adjusting sections,
  retiring stale assumptions, and fixing failed checks.
- Codemap is absent or explicitly deferred.

Common failure modes:

- One-container architecture with frontend/backend/database only as component
  labels.
- Generic implementation plan with no C3 command trail.
- Mentions "grow as you go" but does not raise a contract or migrate docs.
- Adds features but not migrations or gardening.
- Runs verification before growth and claims the later state is verified.
