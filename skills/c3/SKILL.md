---
name: c3
description: >
  Triggers on /c3 or architecture questions in projects with .c3/ directory.
  Phrases: "adopt C3", "onboard", "where is X", "audit architecture", "check docs",
  "add component", "implement feature", "what breaks if I change X", "add ref",
  "coding standard", "edit the canvas", "add prd/user-story".
  Ops: onboard, query, audit, change, ref, rule, canvas, sweep. Canvas definitions
  are the shape of every entity type — user-owned markdown c3x validates against.
  Classifies intent, loads ref, executes.
---

# C3

CLI binary: packaged with this skill at `<skill-dir>/bin/c3x.sh`.

- **Agent command handle:** create a session-local alias/function once, then use it for every command:
  ```bash
  c3() { C3X_MODE=agent bash <skill-dir>/bin/c3x.sh "$@"; }
  ```
- **Enforce:** packaged `c3x.sh` = single source. Skill = router: classify -> run `c3` -> follow output. No duplicate CLI checklists.
- **Agent mode:** `C3X_MODE=agent` -> TOON output (~40% fewer tokens). Commands append `help[]` hints.
- **Entry point:** `c3 list` for topology + coverage, `c3 check` for validation. `--help` for help.

## Mental model — definitions are user-owned canvases

Every entity TYPE — `context`, `container`, `component`, `ref`, `rule`, `adr`, and
document types (`prd`, `user-story`, `atomic-design-change`, `pm-requirement`, and
any project-defined type) — has a **canvas definition**: the sections and typed
columns its docs must carry. Definitions are **user-owned markdown** materialized
at `.c3/canvases/<type>.md`; `c3 canvas` manages them and the user may edit them.

- **The project's definition is the contract — not anything written in this skill.**
  Before authoring or validating any entity, read its definition (`c3 schema <type>`
  renders it; `c3 canvas read <type>` is the owned source). Required sections, table
  columns, and reject rules come from there, never from a fixed list in this file.
- **`c3 check` validates against the definition**, so a user's edit to a definition
  changes what is enforced. c3x **facilitates the wiring** (scaffold / validate /
  check); it does not dictate shape.
- **ADR is not special — it is the `adr` canvas.** Read `c3 schema adr` for its shape.

| Command | Purpose |
|---------|---------|
| `list` | Topology with counts + coverage (`--flat`, `--compact`) |
| `check` | Validate docs against their canvas definitions (`--fix`, `--only`, `--include-adr`, `--only-touched`). Also verifies an accepted unit's declared code bindings resolve (WARN; `--strict-codemap` to gate) |
| `canvas <list\|read\|add\|write>` | Manage canvas definitions (the shape of each entity type); user-owned at `.c3/canvases/` |
| `schema <type>` | Render the canvas definition for `<type>` (sections, columns, REJECT IF). Source of truth = `c3 canvas read <type>` |
| `add <type> <slug>` | **Create** entity; body via stdin or `--file <path>` (`--container`, `--feature`). Unguarded — the create path. Valid types = `c3 canvas list` |
| `write <id>` | Rewrite body or a section (`--section <name>`, `--file <path>`, stdin). **Refused on a frozen fact** → use the change-unit flow. Direct only for change-docs (adr/prd) + canvas bodies |
| `set <id> <field> <val>` | Update a frontmatter field (goal, status, boundary, category, title, date). **Refused on a frozen fact** → change-unit flow. **Exception: `set <id> codemap '<glob>'`** — the external code binding is verified, not frozen, so this one field is editable on a frozen fact (live code-map maintenance) |
| `read <id>` | Entity content; agent truncates 1500 chars (`--full` bypasses). `--section <name> --cite` emits the patch base anchor |
| `search <query>` | Natural-language or conceptual query -> candidate entities by semantic + keyword + graph signals |
| `lookup <file-or-glob>` | File/glob -> component + refs |
| `delete <id>` | Remove entity + clean refs (`--dry-run`). **Refused on a frozen fact** → change-unit flow (`retire` patch) |
| `change <new\|view\|status\|accept\|apply\|rebase\|scaffold>` | The change-unit flow — the **only** way to edit a frozen fact. Author patches (and optional `.codemap.md` carriers for code re-bindings) in `.c3/changes/<unit-id>/`, then `change apply` lands them atomically. `view`/`status` show both arms — internal patches + external code bindings (see `references/change.md`) |
| `change scaffold <unit-id>` | Stage a rung-climb: scans every fact, finds those below their canvas's current required bar, and writes one `insert` patch per fact with an **empty** required-section template (heading + column headers, no rows). The emptiness gates the climb — `change apply` refuses an empty required section, so each template must be filled first (see `references/change.md` §Climbing a rung) |
| `graph <id>` | Relationship graph (`--depth`, `--direction forward|reverse`, `--format mermaid`) |

**Don't assume a fixed type set** — confirm any type's shape with `c3 schema <type>` (works for every entity type); `c3 canvas list` shows the document/definition canvases.

**Authoring — two surfaces, one boundary.** Read the type's definition first (`c3 schema <type>`), author to it. A body round-trips embeddable content verbatim — mermaid/code fences, tables, images, raw HTML and `<iframe>`/embed blocks, dividers, indented code — so diagrams and embeds survive store→render. Author such bodies via `--file <path>`, not inline strings.

- **FROZEN FACTS** — `system`, `container`, `component`, `ref`, `rule`, `recipe`, `pm-requirement`, `user-story` (every type whose canvas declares no `status:` set). **You cannot `write`/`set`/`delete` an existing one.** Editing a fact happens ONLY through the change-unit flow: `c3 read <id> --section <name> --cite` -> `c3 change new <unit-id>` -> author `<seq>-<slug>.patch.md` -> `c3 change apply <unit-id>`. See `references/change.md`. (**Creating** a new fact is exempt — `c3 add <type> <slug>` is unguarded.)
- **CHANGE-DOCS** (`adr`, `prd`, `atomic-design-change`) **and CANVAS bodies** — these declare `status:` / are user-owned, so they edit directly: single-sentence text -> `echo "..." | c3 write <id> --section <name>`; whole-body rewrite -> `c3 write <id> --file body.md`; frontmatter -> `c3 set <id> <field> <value>`.

The guard keys on the FIRST arg only, and an unknown/typo'd type is treated as frozen. The exact refusal names the fix: *"<id> is a fact — facts are frozen and change only through a change-unit."*

---

## Intent Classification

| Keywords | Op | Ref |
|----------|----|-----|
| adopt, init, scaffold, bootstrap, onboard, "create .c3" | **onboard** | `references/onboard.md` |
| where, explain, how, diagram, trace, "show me", "what is", "what handles" | **query** | `references/query.md` |
| audit, validate, "check docs", drift | **audit** | `references/audit.md` |
| add, change, fix, implement, refactor, remove, provision, design | **change** (the change-unit flow: ADR = unit; fact-edits ride as patches, land via `change apply`) | `references/change.md` |
| pattern, convention, "create ref", "update ref", standardize | **ref** | `references/ref.md` |
| "coding rule", "coding standard", "coding convention", "split ref into rule" | **rule** | `references/rule.md` |
| "adopt rule", "adapt an external rule", "codify a standard" | **rule** (Adopt) | `references/rule.md` |
| "edit the canvas", "change the shape", "what sections does X have", "add a doc type", "customize the ADR/component definition", "add prd/user-story" | **canvas** | `references/canvas.md` |
| impact, "what breaks", assess, sweep, "is this safe" | **sweep** | `references/sweep.md` |
| recipe, "trace end-to-end", "cross-cutting flow" | **query**(read) / **change**(create) | `references/query.md` / `references/change.md` |

## Dispatch

1. Classify op (ambiguous -> `AskUserQuestion` with options)
2. Load `references/<op>.md`
3. Execute (Task tool for parallelism)

## Precondition

**Read-only fast path:** for conceptual or natural-language discovery ("where is X", "what handles Y", paraphrases), start with `c3 search "<question>"`, then read the best candidates. For known files/globs, use `c3 lookup <file>`. For known IDs/sections, use `c3 read <id> --section <name>`. Do not run `c3 list` or `c3 check` before the first conceptual `c3 search`; use them only after search/lookup misses, drift is suspected, topology-wide inventory is required, mutation verification is needed, or the user explicitly asks for validation/audit. Prefer section reads. Skip graph unless relationship/dependent impact is part of the answer.

Missing `.c3/` -> **onboard**. Follow `help[]` hints in output.

## CoT Harness

C3 context loaded. Before touching anything, ask C3 first:

`c3 lookup <file-or-glob>` -> component docs, refs, constraints. They supersede assumptions.

Context shifts mid-task -> look up again. Topology = source of truth.

## ASSUMPTION_MODE

First `AskUserQuestion` denial -> `ASSUMPTION_MODE = true` for session.
- Never call `AskUserQuestion` again
- High-impact: state assumption, mark `[ASSUMED]`
- Low-impact: auto-proceed

---

## Shared Rules

**HARD RULE -- entity instances are CLI-only. NEVER Read/Glob/Edit/Write entity docs under `.c3/` (containers, components, refs, rules, adrs, recipes, document instances).** Mutate them only through the `c3` command handle below.

**HARD RULE -- FACTS ARE FROZEN.** A *fact* = any entity whose canvas declares no `status:` set (`system`, `container`, `component`, `ref`, `rule`, `recipe`, `pm-requirement`, `user-story`). `c3 write`/`set`/`delete` on an existing fact is **refused** — edit it ONLY by authoring patches in a change-unit and running `c3 change apply`. **Exempt (still direct):** `c3 add` (create), `c3 canvas write` (canvas bodies are user-owned), editing a *change-doc* (`adr`/`prd`/`atomic-design-change` — they declare `status:`), and **`c3 set <id> codemap`** (the external code binding is verified, not frozen — code churns, so the map is live-maintenance; deliberate re-bindings ride a change-unit as a `.codemap.md` carrier, see `references/change.md`). The guard checks the FIRST arg only; unknown types are treated as frozen.

**HARD RULE -- A FACT IS ALWAYS COMPLETE TO ITS RUNG.** The canvas is a **rung**: a complete contract for one complexity *level*, not a target to fill in over time. A fresh init's canvas is deliberately lean (rung-1) — every required section is present, but the deeper sections a complex project needs are a *higher* rung, not a hole. Completeness is never relaxed: a fact is always complete to its current rung, never thin or partial. What grows is the complexity **level**, and you grow it by **climbing a rung** — raise the canvas (make an optional section required, or author a richer one via `c3 canvas write`), then **migrate** every affected fact up to the new contract, completely. Migration is the mechanism of climbing: integrity forbids facts straddling two rungs. Each rung stands on its own and is not responsible for future rungs — solve now completely, do not over-engineer rung-1 with sections a later rung would carry; climb when the architecture earns it. The climb flow (`change scaffold` → fill → gated `change apply`) is in `references/change.md` §Climbing a rung.

**Canvas DEFINITIONS at `.c3/canvases/<type>.md` are the exception — they are user-owned markdown.** Read them (`c3 canvas read <type>` / directly) and edit them (`c3 canvas write` or by hand); they are meant to be customized. `c3 check` validates instances against them.

`.c3/c3.db` = disposable cache, not submitted state. Raw access to *instance* files bypasses the CLI contract -> stale/misleading. Instance access via the `c3` command handle:

| Op | Commands |
|----|----------|
| Create (unguarded) | `add` |
| Read | `read <id>` (`--cite` for a patch base), `list`, `lookup`, `graph` |
| Edit a FROZEN FACT | `change new` → author patch → `change apply` (the ONLY path; `write`/`set`/`delete` are refused on facts) |
| Edit a change-doc / retire a fact | `write <id>`, `set` on `adr`/`prd` only; fact frontmatter/retire ride as patches via `change apply` |
| Validate | `check` |
| Definitions (user-owned) | `canvas <list\|read\|add\|write>`, `schema <type>` |

Missing packaged CLI operation -> STOP, tell user. No file-tool workarounds.

**Search strategy:** concept->entity via `c3 search "<question>"`; code->entity via `c3 lookup <file-or-glob>`; topology via `c3 list`; doc bodies via targeted `c3 read <id> --section <name>` or `c3 read <id> --full`.

**Embeddings self-heal — no manual re-index after a change.** `change apply` does not touch embeddings; the next `c3 search` reconciles the changed fact incrementally. `c3 index` = full rebuild (maintenance only), never a correctness step.

**`c3 check` after every mutation** — after `change apply` (the landing for any fact edit), after `add` (create), and after `canvas write`. Errors = blockers. (`change apply` runs drift + canvas gates first, but `check` is still the post-land confirmation.)

**Read the definition before authoring or validating any entity.** Required sections / columns / reject rules come from `c3 schema <type>` (the project's canvas definition), never from memory or this skill's prose. If a user edited the definition, that edit is the contract.

**ADR-first — the ADR is the change-unit:**
Every **change** starts `c3 add adr <slug>` before implementation.
(Exception: **ref-add** creates ADR at completion -- `references/ref.md`.)
ADR = work order, and the ADR **is the change-unit**: same id, folder `.c3/changes/<adr-id>/`. You author the ADR *freely* (`add adr` / `write` / `set` — it is a change-doc, not a frozen fact); any edit it makes to a frozen fact rides as a `<seq>-<slug>.patch.md` in that folder and lands when you run `c3 change apply <adr-id>`. Author the reasoning, carry the fact-edits as patches, land them together.

Two distinct freezes — do not blur them:
- **Frozen FACT** (structural): a fact's *body* can only change through a change-unit. Applies the moment the fact exists.
- **Content-frozen change-doc** (historical): a *terminal-state* ADR is exempt from `c3 check` — its prose is a frozen record of a past decision.

ADR schema discipline is unchanged: start with `c3 schema adr` (or `c3 canvas read adr`) BEFORE drafting any ADR body. The schema output leads with a **REJECT IF** block — read it first; the bullets ARE the rejection contract. Per-section `fill:` and `rejected when:` lines apply the same gate at section level. Same shape for any `c3 schema <type>` — REJECT IF first, fill the body to the contract, never draft freehand and reconcile later. Use `c3 read <adr> --full` and `help[]` to verify the body still matches the contract. The adr canvas is **laddered**: a lean required core (Goal, Context, Decision, Affected Topology, Verification) covers a small change; the work-order sections (Compliance Refs/Rules, Work Breakdown, Underlay C3 Changes, Enforcement Surfaces, Alternatives, Risks) are optional and climb in for weightier decisions. Any section you DO include must be substantive — thin included sections fail; `N.A - <reason>` for inapplicable rows. `list`/`check` exclude ADRs by default; `--include-adr` only when working on a specific ADR. Change-doc status set is `[open, accepted, done, superseded]` (`c3 add adr` stamps `proposed` at creation — the unmigrated synonym for `open`); `accepted` auto-latches to `done` when its After-cites resolve fresh. **Terminal-state change-docs (`status: done`, `status: superseded`) are exempt from `c3 check`** — historical, content-frozen. ADR content is historical — verify against current docs.

**Stop if:**
- No ADR for change -> `c3 add adr <slug>` NOW (the ADR is the change-unit; fact-edits ride in `.c3/changes/<adr-id>/`)
- Guessing intent -> `AskUserQuestion` (skip if ASSUMPTION_MODE)
- Jumping to component -> start Context down
- Authoring/validating against remembered sections instead of `c3 schema <type>`
- Updating docs without code check

## File Context -- MANDATORY before reading/altering any file

**1. Lookup:**
```bash
c3 lookup <file-path>
c3 lookup 'src/auth/**'
```
Run when any file path surfaces. No match = uncharted.

**2. Load rules:** For every `rule-*` from lookup:
```bash
c3 read <rule-id>
```
Read its `## Rule` / `## Golden Example` / `## Not This` (or whatever sections its definition declares). Code MUST match golden pattern. Deviations need Override or new ADR.

**3. Parent context:** Read upward first:
```bash
c3 read <parent-container-id>
c3 read <component-id>
```
Parent responsibilities + membership = integration contract.

**4. Graph context:**
```bash
c3 graph <parent-container-id> --depth 1
c3 graph <component-id> --depth 1
```
Change affects interface -> check consumers first.

**Result:** Refs + rules + parent + graph = full constraint set. All honored.

**New component:** set its parent (`c3 add --container c3-N`, or `parent:` in a create-patch) — the container's `Components` row is synthesized by the tool on either path (it owns membership; never hand-edit the row). The Parent Delta decision is now only: does the parent's `Responsibilities` (or the member's `Goal Contribution` framing) change?

**Navigation:** Context -> Container -> Component

## Graph Output

Include mermaid when relationships matter:
```bash
c3 graph <entity-id> --format mermaid
```
Root selection > depth:
- **container**: components + refs (query, audit, onboard)
- **component**: constraints + siblings (query, change, sweep)
- **ref/rule**: citation graph (Usage, sweep)
- `--depth 1` default. `--depth 2` for cross-container only
- `--direction forward` = impact. `--direction reverse` = dependents
- Never graph `c3-0` -- one node, no signal
- `--unit <adr-id>` = **contextual preview**: overlays that change-unit's staged-but-unapplied patches (the real apply path, rolled back) so you see the post-change graph before `change apply`. Output is marked preview; all modes (text/mermaid/reverse/json). Pairs with the change flow: cite/stage → `graph --unit` → apply

## File Structure
```
.c3/
├── README.md                    # Context (c3-0)
├── canvases/<type>.md           # Canvas definitions (user-owned shape of each type)
├── adr/adr-YYYYMMDD-slug.md
├── refs/ref-slug.md
├── rules/rule-slug.md
├── recipes/recipe-slug.md
└── c3-N-name/
    ├── README.md                # Container
    └── c3-NNN-component.md
```

---

## Operations

### onboard
No `.c3/` or re-onboard. Scaffold (incl. materialize canvas definitions) -> discovery -> inject CLAUDE.md -> show capabilities.
`references/onboard.md`

### query
`c3 search <question>` first for fuzzy/conceptual discovery, `c3 lookup <file-or-glob>` for code->entity, `c3 graph <id> --direction reverse` for dependents. For body text, use targeted `c3 read` output. Deliver to the **Answer Depth Contract** in the reference — claims bound to reads, causal chain over entity lists, failure boundary stated.
`references/query.md`

### audit
`c3 check` -> `c3 list` -> semantic phases (each entity validated against its canvas definition) -> PASS/WARN/FAIL table.
`references/audit.md`

### change
ADR first (= the change-unit) -> `c3 list` -> affected entities -> `c3 lookup` files -> fill ADR to the `adr` canvas. **Editing a frozen fact:** `c3 read <id> --section <name> --cite` -> `c3 change new <adr-id>` -> author patches -> `c3 change view`/`status`/`accept` -> `c3 change apply` (the only mutation) -> `c3 check`. **Creating** a fact stays direct (`c3 add`). Provision gate: implement or design-only.
`references/change.md`

### ref
Modes: Add / Update / List / Usage. Author to the `ref` canvas (`c3 schema ref`).
`references/ref.md`

### rule
Modes: Add / Update / List / Usage. Author to the `rule` canvas (`c3 schema rule`).
`references/rule.md`

### canvas
Inspect or edit the definitions that shape entities: `c3 canvas list/read`, `c3 canvas write` to customize a project's shape, `c3 canvas add` for a new project-defined type. Definitions are user-owned at `.c3/canvases/`; editing one changes what `c3 check` enforces.
`references/canvas.md`

### sweep
`c3 graph <id> --direction reverse` -> transitive deps -> parallel assessment -> synthesize. Advisory only.
`references/sweep.md`
