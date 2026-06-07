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
| `check` | Validate docs against their canvas definitions (`--fix`, `--only`, `--include-adr`, `--only-touched`) |
| `canvas <list\|read\|add\|write>` | Manage canvas definitions (the shape of each entity type); user-owned at `.c3/canvases/` |
| `schema <type>` | Render the canvas definition for `<type>` (sections, columns, REJECT IF). Source of truth = `c3 canvas read <type>` |
| `add <type> <slug>` | Create entity; body via stdin or `--file <path>` (`--container`, `--feature`). Valid types = `c3 canvas list` |
| `write <id>` | Rewrite body or a section (`--section <name>`, `--file <path>`, stdin) |
| `set <id> <field> <val>` | Update frontmatter field (goal, status, boundary, category, title, date, codemap patterns) |
| `wire <src> <tgt>` | Link entities (`--remove` unlinks) |
| `read <id>` | Entity content; agent truncates 1500 chars (`--full` bypasses) |
| `lookup <file-or-glob>` | File/glob -> component + refs |
| `delete <id>` | Remove entity + clean refs (`--dry-run`) |
| `graph <id>` | Relationship graph (`--depth`, `--direction forward|reverse`, `--format mermaid`) |

**Don't assume a fixed type set** — confirm any type's shape with `c3 schema <type>` (works for every entity type); `c3 canvas list` shows the document/definition canvases.

**Authoring:** read the type's definition first (`c3 schema <type>`), author to it. Body with mermaid, code fences, or tables -> use `--file <path>`, not inline strings. Single-sentence text edits -> `echo "..." | c3 write <id> --section <name>`. Whole-body rewrite -> `c3 write <id> --file body.md`. Frontmatter fields -> `c3 set <id> <field> <value>`.

---

## Intent Classification

| Keywords | Op | Ref |
|----------|----|-----|
| adopt, init, scaffold, bootstrap, onboard, "create .c3" | **onboard** | `references/onboard.md` |
| where, explain, how, diagram, trace, "show me", "what is" | **query** | `references/query.md` |
| audit, validate, "check docs", drift | **audit** | `references/audit.md` |
| add, change, fix, implement, refactor, remove, provision, design | **change** | `references/change.md` |
| pattern, convention, "create ref", "update ref", standardize | **ref** | `references/ref.md` |
| "coding rule", "coding standard", "coding convention", "split ref into rule" | **rule** | `references/rule.md` |
| marketplace, "browse rules", "adopt rule", "install rule from" | **rule** (Adopt) | `references/rule.md` |
| "edit the canvas", "change the shape", "what sections does X have", "add a doc type", "customize the ADR/component definition", "add prd/user-story" | **canvas** | `references/canvas.md` |
| impact, "what breaks", assess, sweep, "is this safe" | **sweep** | `references/sweep.md` |
| recipe, "trace end-to-end", "cross-cutting flow" | **query**(read) / **change**(create) | `references/query.md` / `references/change.md` |

## Dispatch

1. Classify op (ambiguous -> `AskUserQuestion` with options)
2. Load `references/<op>.md`
3. Execute (Task tool for parallelism)

## Precondition

**Read-only fast path:** for file-owner, "where is", summarize constraints, or smallest-next-action queries that do not mutate docs/code, start with the narrowest `c3 lookup <file>` or `c3 read <id> --section <name>`. Skip `list` and `check` unless lookup misses, drift is suspected, topology-wide inventory is required, or the user explicitly asks for validation/audit. Prefer section reads. Skip graph unless relationship/dependent impact is part of the answer.

Before every op except onboard:
```bash
c3 list
c3 check
```
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

**Canvas DEFINITIONS at `.c3/canvases/<type>.md` are the exception — they are user-owned markdown.** Read them (`c3 canvas read <type>` / directly) and edit them (`c3 canvas write` or by hand); they are meant to be customized. `c3 check` validates instances against them.

`.c3/c3.db` = disposable cache, not submitted state. Raw access to *instance* files bypasses the CLI contract -> stale/misleading. Instance access via the `c3` command handle:

| Op | Commands |
|----|----------|
| Create | `add` |
| Read | `read <id>`, `list`, `lookup`, `graph` |
| Update | `write <id>`, `set`, `wire` (`--remove` unwires) |
| Delete | `delete` |
| Validate | `check` |
| Definitions | `canvas <list\|read\|add\|write>`, `schema <type>` |

Missing packaged CLI operation -> STOP, tell user. No file-tool workarounds.

**Search strategy:** code->entity via `c3 lookup <file-or-glob>`; topology via `c3 list`; doc bodies via targeted `c3 read <id> --section <name>` or `c3 read <id> --full`.

**`c3 check` after every mutation** (`add`, `write`, `set`, `wire`, `delete`, `canvas`). Errors = blockers.

**Read the definition before authoring or validating any entity.** Required sections / columns / reject rules come from `c3 schema <type>` (the project's canvas definition), never from memory or this skill's prose. If a user edited the definition, that edit is the contract.

**ADR-first for changes:**
Every **change** starts `c3 add adr <slug>` before implementation.
(Exception: **ref-add** creates ADR at completion -- `references/ref.md`.)
ADR = work order, and ADR = the `adr` canvas. Start with `c3 schema adr` (or `c3 canvas read adr` for the owned source) BEFORE drafting any ADR body. The schema output leads with a **REJECT IF** block — read it first; the bullets ARE the rejection contract. Per-section `fill:` and `rejected when:` lines apply the same gate at section level. Same shape for any `c3 schema <type>` — REJECT IF first, fill the body to the contract, never draft freehand and reconcile later. Use `c3 read <adr> --full` and `help[]` to verify the body still matches the contract. ADR creation is all-or-nothing; thin sections fail at creation, `N.A - <reason>` for inapplicable rows. `list`/`check` exclude ADRs by default; `--include-adr` only when working on specific ADR. **Terminal-state ADRs (`status: implemented` and `status: provisioned`) are exempt from `c3 check` validation** — historical, content frozen. ADRs cannot be created as `implemented`; transition `proposed → accepted → implemented` (direct `proposed → implemented` is blocked). ADR content historical -- verify against current docs.

**Stop if:**
- No ADR for change -> `c3 add adr <slug>` NOW
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

**New component:** Top-down. Container `Components` + `Responsibilities` first, then component. Not done until parent has Parent Delta decision.

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
`c3 list` for topology, `c3 lookup <file-or-glob>` for code->entity, `c3 graph <id> --direction reverse` for dependents. For body text, use targeted `c3 read` output.
`references/query.md`

### audit
`c3 check` -> `c3 list` -> semantic phases (each entity validated against its canvas definition) -> PASS/WARN/FAIL table.
`references/audit.md`

### change
ADR first -> `c3 list` -> affected entities -> `c3 lookup` files -> fill ADR (to the `adr` canvas) -> approve -> execute -> `c3 check`. Provision gate: implement or `status: provisioned`.
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
