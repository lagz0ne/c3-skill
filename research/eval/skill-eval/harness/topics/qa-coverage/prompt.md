# Topic Prompt: Build a Test-Coverage Knowledge Graph

C3 is a general knowledge-graph contract tool, not only an architecture tool. In
this topic you will use it to build a **QA traceability graph**: the requirements
and risks of a non-trivial feature, and the test-cases that verify them — wired so
that every test cites what it proves, and every requirement/risk has a test proving
it. An untested requirement is a hole the graph must make visible.

Do **not** model software topology here. Do not reuse the built-in architecture
canvases (`system`, `container`, `component`, `recipe`). You must **define your own
document types** for the QA domain and author facts against them.

## The feature

Invent one non-trivial feature with real failure modes — for example a **checkout
flow** (cart → payment → confirmation) or a **sign-in/auth flow** (credentials →
session → recovery). Pick one and commit to it. It must have enough behavior to
yield several distinct requirements, several distinct risks, and a dozen-ish
test-cases — including money-loss / security / data-integrity risks that a single
happy-path test cannot cover.

## Your task

1. **Design 2–4 custom canvases** for this QA domain via
   `C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh canvas add <id> < schema.md`.
   At minimum design:
   - a **`requirement`** type (a testable behavior the feature must satisfy),
   - a **`risk`** type (a failure mode the test strategy must mitigate),
   - a **`test-case`** type in **Given/When/Then** form, with a **priority**, and a
     **Traceability** table that carries **two edge columns**: one that wires to the
     requirement it `verifies`, and one that wires to the risk it `covers`,
   - and a **`test-plan`** (or `test-suite`) type that states scope + risk strategy
     and **governs** the requirements/risks in scope — a cross-cutting fact that no
     single test-case owns.

   **At least one edge column per canvas must wire to another of your domain types.**
   An edge column is a table column of `type: reference` with `edge: <relationship>`
   and `targets: <typeA>,<typeB>` — this is how one fact CITES another and how the
   set becomes a traceable graph. Confirm the exact YAML by reading a built-in that
   already uses edge columns: `canvas read component` (its **Governance** table has a
   `type: reference` / `edge: uses` / `targets: ref,rule` column), and the inline
   `edge<a|b>` shorthand in `canvas read pm-requirement` and `canvas read user-story`.
   Read your own canvas back with `c3x schema <type>` before authoring facts.

2. **Author a lean first cut.** Start small and complete: a handful of requirements,
   a handful of risks, one test-plan, and an initial set of test-cases — each one in
   Given/When/Then form, each citing the requirement it verifies **and** the risk it
   covers. Do **not** pre-build the richer sections that a later sub-flow will force.

3. **Wire a dense traceability graph.** Every test-case must put real fact ids in its
   edge columns (the requirement it verifies, the risk it covers). The test-plan must
   wire to the requirements/risks it governs. Then prove coverage from the *target*
   side: `c3x graph <requirement-id> --direction reverse` and
   `c3x graph <risk-id> --direction reverse` must show real test-cases pointing back.
   **Traceability is complete only when every stated requirement and every stated risk
   has ≥1 covering test-case** — an orphan requirement or an uncovered risk is the
   failure this topic is built to expose. Find and close them.

4. **Grow under a named pressure — climb the canvas.** The feature gains a new
   sub-flow (e.g. checkout adds **discount-code / partial-refund** handling, or auth
   adds **password-reset / lockout**). This sub-flow has failure modes the lean
   test-case shape can't express, so **raise the contract**: add required structure
   the first rung lacked — for example a required **Negative Cases** column/section
   and a required **Evidence** column on `test-case` (and, if warranted, a new type
   such as `negative-case` or a `coverage-gap` fact). State the pressure by name,
   then **migrate the affected facts completely** so they satisfy the raised canvas —
   use the change-unit climb path (`c3x change scaffold <id>` stages one insert patch
   per fact below the new bar; author the patches; `c3x change apply <id>`). Migration
   is not one late cleanup: the new sub-flow's requirements and risks must be added
   and covered too.

5. **Keep `c3 check` honest.** Run `c3x check` and report the **exact** result —
   distinguish errors from warnings and name any that remain. An edge column
   targeting a custom (non-architecture) type grounds CLEANLY — `check` resolves the
   cite by entity lookup (any id shape, builtin or custom-type), so a well-wired graph
   is `issues[0]`.
   Likewise a `cite`/`evidence` value must be a real handle, not a bare file path. Do
   not hide these behind a "clean" claim — report errors vs warnings truthfully and
   say which facts they fall on.

## Constraints

- Use local C3 only: `C3X_MODE=agent bash /opt/c3/skills/c3/bin/c3x.sh`.
- Define your own QA canvases. Do **not** model the feature with the built-in
  architecture canvases (`system`/`container`/`component`/`recipe`) — those describe
  software topology, not test coverage.
- Avoid codemap work; it is not the focus of this eval (the feature need not have
  real source). Use edge columns and `cite`/`evidence` for wiring, not file mapping.
- Prefer compact, concrete C3 artifacts (canvases, facts, the reverse-graph output)
  over long prose.
- Run verification and report the exact result — including the reverse-graph coverage
  for at least one requirement and one risk, and the final `c3x check` summary.
