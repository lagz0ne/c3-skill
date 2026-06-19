# Rubric Notes: Build a Test-Coverage Knowledge Graph

The run should not pass merely by describing tests for a feature, or by writing a
test plan in prose. It must show that the agent can use C3 as a **general
knowledge-graph contract tool** — defining its own document types and wiring a
dense, traceable QA graph — not as an architecture-diagram generator. This topic
deliberately exercises the two features the architecture topics barely touch:
**custom canvases** and **typed-citation wiring**.

## Must-have evidence

- **Local C3 command evidence**, not bare/global `c3x` (commands run through
  `/opt/c3/skills/c3/bin/c3x.sh`).

- **Custom canvases, not built-ins.** The agent ran `canvas add` to define its own
  QA types (`requirement`, `risk`, `test-case`, and a `test-plan`/`test-suite`),
  visible in `canvas list` with a non-architecture `domain` and `source` of project
  canvas files. The feature is modelled with these types — **not** with
  `system`/`container`/`component`/`recipe`. Reusing the architecture canvases to
  describe the feature is a fail for this topic even if otherwise tidy.

- **≥1 edge column per custom canvas, wiring to another domain type.** Each canvas
  has at least one table column of `type: reference` with `edge: <relationship>` and
  `targets:` naming another of the agent's types (or the inline `edge<a|b>` form).
  Concretely: `test-case` carries **two** distinct edges — `verifies → requirement`
  and `covers → risk`; `test-plan` wires (e.g. `governs`) to the requirements/risks
  in scope. `c3x schema <type>` shows the edge metadata. A canvas with no edge column
  owns no wiring and does not count.

- **Dense wiring, every citation resolving.** Multiple test-cases (target a dozen-ish,
  not two), and **every** test-case puts real fact ids in both edge columns — none
  left blank or `N.A`. The test-plan wires to its governed facts. Edges must be real
  graph edges: `c3x graph <requirement-id> --direction reverse` and
  `c3x graph <risk-id> --direction reverse` show actual test-cases pointing back.
  (Sanity check on mechanism: only `reference`/edge columns create the reverse-graph
  edge — a plain `entity_id` column resolves in `check` but produces **no** edge, so
  reverse-graph coverage is the proof that the agent wired, not just referenced.)

- **Traceability completeness — no orphan.** Every stated requirement and every
  stated risk has ≥1 covering test-case (shown via reverse-graph or an explicit
  coverage pass). An untested requirement or an uncovered risk that the agent leaves
  silently in place is the central failure; if a gap exists it must be surfaced and
  closed (or named as an explicit, justified `coverage-gap` fact), not hidden.

- **A real growth/climb step with migration.** A **named** pressure (a new sub-flow:
  discount-code / partial-refund, or password-reset / lockout) forces the contract up
  — a required section/column the first rung lacked (e.g. **Negative Cases**,
  **Evidence**) and/or a new type. The lean first cut is visible *before* the climb;
  the climb is a deliberate canvas raise; affected facts are **migrated completely**
  to satisfy the new bar (ideally via the change-unit climb: `change scaffold` →
  author patches → `change apply`), and the new sub-flow's own requirements/risks are
  added **and** covered. Verification is re-run *after* migration.

- **Check honesty.** The agent ran `c3x check` and reported the **exact** outcome,
  distinguishing errors from warnings. A truthful "0 errors, N warnings (ungrounded
  reference on custom-type edges X/Y)" is acceptable and expected; a bare "check is
  clean" that buries real warnings is not. Bonus integrity: filling the `c3-0` system
  context so its skeleton warnings clear, and using proper `cite` handles (or `N.A -
  <reason>`) rather than bare file paths in cite/evidence columns.

## Common failure modes

- **Built-in escape hatch:** modelling the feature with `container`/`component`/
  `recipe` (or shoehorning ADRs) instead of defining QA-specific canvases — dodges
  the whole point of the topic.
- **Referenced but not wired:** using `entity_id` (or prose / plain text) columns so
  `check` looks clean, but `graph --direction reverse` shows no test-cases pointing
  back — the traceability graph never actually formed.
- **Sparse or one-directional wiring:** test-cases that cite a requirement but no risk
  (or vice-versa), or only a couple of cases total, or a test-plan that governs
  nothing.
- **Orphan left standing:** a requirement or risk with zero covering test-cases that
  the agent never surfaces — traceability incomplete.
- **No real climb:** "grow as you go" mentioned but the canvas is never raised, or the
  richer test-case shape is built up-front (no lean first rung), or new structure is
  added without migrating the existing facts to satisfy it.
- **Migration as one late cleanup** instead of: lean cut → named pressure → raised
  canvas → complete migration → re-verify, with the new sub-flow's requirements/risks
  added and covered.
- **Check theater:** claiming clean while warnings/errors remain, or running `check`
  before the growth/migration and reporting that earlier state as the final result.
- **Codemap busywork:** chasing file-mapping/Derived-Materials when the feature has no
  source — wiring here is edges and cite handles, not codemap.
- **Bare/global `c3x`** instead of the local skill binary.
