# Rubric Notes: Build a Traceable Design System

This topic does not test "can the agent describe a design system." It tests the two
C3 capabilities the architecture topics barely exercise: **defining custom canvases
with edge columns**, and **wiring a dense, every-citation-resolves traceability
graph** — then growing it under a real pressure. Score on the evidence below, not on
prose quality.

## Invariants — the scoring spine (with falsifiers)

`check`-clean is necessary, not sufficient: it proves citations resolve, not that the
graph is **complete**. These are the bar. The falsifier is what a reviewer finds in the
finished workspace `.c3/` to fail the run.

| Invariant | Falsifier (find this → broken) |
| --- | --- |
| **INV-COMPONENT-WIRED** | A UI component with an empty `uses` (token) or `follows` (a11y rule) edge column. |
| **INV-NO-HARDCODE** | A raw value (hex / px / font literal) in a component section where a token cite belongs — the component hardcodes instead of citing. |
| **INV-TOKEN-USED** | `graph <semantic-token> --direction reverse` is empty for a defined token — an orphan nobody consumes. |
| **INV-SEMANTIC-RESOLVES** | After the climb: a semantic token with no per-theme primitive edge, OR a component still wired to a raw primitive (migration incomplete). |
| **INV-FLOW-SEQUENCES** | A flow whose steps are prose component *names*, not edge cites; `graph <flow>` shows no component edges. |
| **INV-FORMAT-TYPED** | After the morph: a token still on the old free-text `Value` (not migrated to typed `Format` + `Value`), OR a `Format` that contradicts its value (a token typed `color` holding `8px`) — a mechanical retype that didn't actually sort formats. The morph gate keeps `check` clean. |

## Reviewer runbook — how to surface each falsifier
Run against `<run>.workspace/` with the HEAD binary (`C3X_MODE=agent /tmp/c3x-score --c3-dir .c3 <cmd>`); `grep` over `.c3/` catches hardcoded literals.

| Invariant | Commands → what to look for |
| --- | --- |
| **INV-COMPONENT-WIRED** | `read <each ui-component>` → `uses` + `follows` filled with real ids; `graph <component>` shows both edges. |
| **INV-NO-HARDCODE** | `grep -rnE '#[0-9a-fA-F]{3,6}' .c3/` (hex; also scan for px/font literals) → a raw value in a *component* body where a token cite belongs (a token's own value section is fine; a component's is not). |
| **INV-TOKEN-USED** | `graph <each semantic token> --direction reverse` → ≥1 consuming component; empty = orphan. |
| **INV-SEMANTIC-RESOLVES** | `read <each semantic token>` → per-theme primitive edges; `grep` component cites for primitive ids → none should remain after migration. |
| **INV-FLOW-SEQUENCES** | `read <flow>` + `graph <flow>` → ordered component edges, not text names. |
| **INV-FORMAT-TYPED** | `canvas read <token-type>` → the typed `Format` column; `read <each token>` → Format present and consistent with the value; `change status <morph-adr>` → reshape + migration in ONE unit. |

## Must-have evidence

### A. Custom canvases — NOT the builtins

- The agent ran `canvas add <id> < schema.md` to define **its own** doc types for the
  domain (tokens, UI components, a11y rules/patterns, flows) — at least 2, up to 4.
- It did **not** model the design system with `system` / `container` / `component`.
  A design token modeled as a `container`, or a button as a `component` under a
  container, is the central failure of this topic — fail the custom-canvas dimension.
- Schemas are well-formed: frontmatter (`id`, `type: canvas`, `description`) + body
  with `domain:` and `sections:` (each `name` / `content_type` / `required` /
  `purpose`, table sections with typed `columns:`). Bonus for using `reject_if:` to
  encode a domain invariant (e.g. "component hardcodes a value instead of citing a
  token").
- Evidence: `canvas add` output ("Created canvas ..."), `canvas list` showing the new
  types with `source` other than `built-in`, and/or `c3 schema <type>` output.

### B. At least one edge column per custom canvas, wiring to another domain type

- Each defined canvas carries **≥1 edge column** — `type: reference` with
  `edge:`/`targets:`, or the inline `edge<type>` form — pointing at another of the
  domain's types. A canvas with only `text`/`enum` columns and no edge does not count.
- The three load-bearing edges exist in the schemas:
  - UI component → tokens it **uses** (e.g. a `uses` edge, targets the token type);
  - UI component → the a11y rule / pattern it **follows**;
  - flow → the UI components it **sequences**.
- `targets:` restricts the edge to the right type (e.g. a `uses-tokens` column targets
  the token canvas, not arbitrary entities). Reasonable target restriction is a plus.

### C. Dense wiring — every citation resolves, real coverage

- Facts of the custom types were authored (`c3 add <type> <slug>`, sections as `##`
  headings; or change-unit create-patches) and the edge columns are **actually filled
  with real fact ids**, not left blank or stubbed `N.A`.
- Wiring is **dense, not token**: most components cite multiple tokens + a rule; the
  flow sequences several components in order. A graph with one lonely edge fails this.
- **Coverage proven by reverse graph, not assertion.** The agent ran
  `c3 graph <token-id> --direction reverse` and the output shows the consuming
  component(s) wiring a `uses` edge to that token. Strong runs also probe a
  suspected-orphan token and report the (empty/non-empty) result.
- **Traceability completeness:** every component traces to ≥1 token + the rule it
  follows; every (semantic) token is used by ≥1 component. An **orphan token**
  (defined, cited by nobody) or a component carrying a **hardcoded value** where a
  token cite belongs is a named failure — and a strong run finds and fixes (or
  explicitly flags) it.

### D. A real growth / climb step with migration

- The first cut is genuinely **lean** (small token set, a couple of components, a rule
  or two, one flow) — later-stage theming structure is **not** pre-built.
- The **theming pressure** is applied as a visible step and forces a **canvas climb /
  new type**: the flat token splits into **primitive** vs **semantic** tokens (or an
  equivalent raising of the token canvas), with a **new edge** letting a semantic token
  cite the primitive(s) it resolves to **per theme**.
- **Migration is complete:** component cites are re-pointed from raw/primitive values
  to the **semantic** token, and the agent re-verified nothing dangles after the
  re-point. The climb, the new semantic→primitive edge, and the migration of component
  cites must all be real — not described, not deferred to "later."

### E. check-clean honesty

- The agent ran `c3 check` and **reported the literal outcome** — pass/fail plus any
  warnings/errors, named.
- **Known wrinkle (do not penalize, reward honesty about it):** `c3 check`'s
  `reference`-column grounding only auto-recognizes the builtin id prefixes
  (`c3-N`, `ref-*`, `rule-*`, `adr-*`, `recipe-*`). A cite to a **custom-type** id
  (e.g. `design-token-...`, `ui-component-...`) now grounds cleanly — `check` resolves
  it by entity lookup and the edge materializes in the graph, so a well-wired system is
  `issues[0]`. The honest, correct report is: structural pass / zero errors, with these
  expected warnings named and explained — NOT a bare "check is clean" that hides them,
  and NOT a panic that the wiring is broken (the reverse graph proves it is not).
- Real errors (a cite to a fact that does not exist, a missing required section, a
  malformed schema) must be **fixed**, not narrated. Note also that a bare `init`
  seeds a top-level context that emits a few "missing required section" warnings;
  acknowledging or filling those is a plus, ignoring-while-claiming-clean is not.

## Common failure modes

- **Reusing builtin architecture canvases** (system/container/component) instead of
  defining custom token / component / flow canvases — defeats the whole topic.
- **No edge columns** — canvases are flat tables of text/enum, so nothing wires and
  the reverse graph is empty; "traceability" is asserted in prose only.
- **Sparse wiring** — one or two edges total; components hardcode values instead of
  citing tokens; the flow lists component *names as text* rather than edge cites.
- **Orphan tokens / hardcoded values** left unaddressed and unreported.
- **No real climb** — theming handled by editing a value in place or adding a `dark`
  text note, with no primitive/semantic split, no new edge, and no component migration
  (or migration deferred to a single late cleanup).
- **Dishonest check report** — claiming "clean" while warnings printed, OR conversely
  treating the expected custom-id `reference` warnings as a fatal break and thrashing
  to "fix" working wiring.
- **Coverage by assertion** — claiming every token is used without ever running
  `graph ... --direction reverse` to prove it.
- **Codemap / Derived-Materials busywork** — there is no code repo here; binding facts
  to files is off-task.
- **Bare/global `c3x`** instead of the local `/opt/c3/skills/c3/bin/c3x.sh` binary.
