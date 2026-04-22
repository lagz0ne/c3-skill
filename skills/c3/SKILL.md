---
name: c3
description: >
  Triggers on /c3 or architecture questions in projects with .c3/ directory.
  Phrases: "adopt C3", "onboard", "where is X", "audit architecture", "check docs",
  "add component", "implement feature", "what breaks if I change X", "add ref",
  "coding standard". Ops: onboard, query, audit, change, ref, rule, sweep.
  Classifies intent, loads ref, executes.
---

# C3

CLI: `C3X_MODE=agent bash <skill-dir>/bin/c3x.sh <command> [args]`

- **Enforce:** `c3x` = single source. Skill = router: classify -> run c3x -> follow output. No duplicate CLI checklists.
- **Agent mode:** `C3X_MODE=agent` -> TOON output (~40% fewer tokens). Commands append `help[]` hints.
- **Entry point:** `c3x list` for topology + coverage, `c3x check` for validation. `--help` for help.

| Command | Purpose |
|---------|---------|
| `list` | Topology with counts + coverage (`--flat`, `--compact`) |
| `check` | Validate docs match code (`--fix`, `--only`, `--include-adr`, `--only-touched`) |
| `add <type> <slug>` | Create entity; body via stdin or `--file <path>` (`--container`, `--feature`) |
| `write <id>` | Rewrite body or a section (`--section <name>`, `--file <path>`, stdin) |
| `set <id> <field> <val>` | Update frontmatter field (goal, status, boundary, category, title, date, codemap patterns) |
| `wire <src> <tgt>` | Link entities (`--remove` unlinks) |
| `schema <type>` | Section defs for entity type |
| `read <id>` | Entity content; agent truncates 1500 chars (`--full` bypasses) |
| `lookup <file-or-glob>` | File/glob -> component + refs |
| `delete <id>` | Remove entity + clean refs (`--dry-run`) |
| `graph <id>` | Relationship graph (`--depth`, `--direction forward|reverse`, `--format mermaid`) |

Types for `add`: `container`, `component`, `ref`, `rule`, `adr`, `recipe`

**Authoring:** body with mermaid, code fences, or tables -> use `--file <path>`, not inline strings. Single-sentence text edits -> `echo "..." | c3x write <id> --section <name>`. Whole-body rewrite -> `c3x write <id> --file body.md`. Frontmatter fields -> `c3x set <id> <field> <value>`.

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
| impact, "what breaks", assess, sweep, "is this safe" | **sweep** | `references/sweep.md` |
| recipe, "trace end-to-end", "cross-cutting flow" | **query**(read) / **change**(create) | `references/query.md` / `references/change.md` |

## Dispatch

1. Classify op (ambiguous -> `AskUserQuestion` with options)
2. Load `references/<op>.md`
3. Execute (Task tool for parallelism)

## Precondition

Before every op except onboard:
```bash
bash <skill-dir>/bin/c3x.sh list
bash <skill-dir>/bin/c3x.sh check
```
Missing `.c3/` -> **onboard**. Follow `help[]` hints in output.

## CoT Harness

C3 context loaded. Before touching anything, ask C3 first:

`c3x lookup <file-or-glob>` -> component docs, refs, constraints. They supersede assumptions.

Context shifts mid-task -> look up again. Topology = source of truth.

## ASSUMPTION_MODE

First `AskUserQuestion` denial -> `ASSUMPTION_MODE = true` for session.
- Never call `AskUserQuestion` again
- High-impact: state assumption, mark `[ASSUMED]`
- Low-impact: auto-proceed

---

## Shared Rules

**HARD RULE -- .c3/ is CLI-only. NEVER Read/Glob/Edit/Write `.c3/` files.**

`.c3/c3.db` = disposable cache, not submitted state. Raw file access bypasses CLI contract -> stale/misleading. ALL access via c3x:

| Op | Commands |
|----|----------|
| Create | `add` |
| Read | `read <id>`, `list`, `lookup`, `graph` |
| Update | `write <id>`, `set`, `wire` (`--remove` unwires) |
| Delete | `delete` |
| Validate | `check`, `schema` |

Missing c3x operation -> STOP, tell user. No file-tool workarounds.

**Search strategy:** code->entity via `c3x lookup <file-or-glob>`; topology via `c3x list`; full-text over doc bodies -> grep over `.c3/`.

**`c3x check` after every mutation** (`add`, `write`, `set`, `wire`, `delete`). Errors = blockers.

**ADR-first for changes:**
Every **change** starts `c3x add adr <slug>` before implementation.
(Exception: **ref-add** creates ADR at completion -- `references/ref.md`.)
ADR = work order. Use `c3x schema adr`, `c3x read <adr> --full`, `help[]` for required detail. Rejects thin creation; complete up front, `N.A - <reason>` for inapplicable rows. `list`/`check` exclude ADRs by default; `--include-adr` only when working on specific ADR. ADR content historical -- verify against current docs.

**Stop if:**
- No ADR for change -> `c3x add adr <slug>` NOW
- Guessing intent -> `AskUserQuestion` (skip if ASSUMPTION_MODE)
- Jumping to component -> start Context down
- Updating docs without code check

## File Context -- MANDATORY before reading/altering any file

**1. Lookup:**
```bash
bash <skill-dir>/bin/c3x.sh lookup <file-path>
bash <skill-dir>/bin/c3x.sh lookup 'src/auth/**'
```
Run when any file path surfaces. No match = uncharted.

**2. Load rules:** For every `rule-*` from lookup:
```bash
bash <skill-dir>/bin/c3x.sh read <rule-id>
```
Extract `## Rule`, `## Golden Example`, `## Not This`. Code MUST match golden pattern. Deviations need Override or new ADR.

**3. Parent context:** Read upward first:
```bash
bash <skill-dir>/bin/c3x.sh read <parent-container-id>
bash <skill-dir>/bin/c3x.sh read <component-id>
```
Parent responsibilities + membership = integration contract.

**4. Graph context:**
```bash
bash <skill-dir>/bin/c3x.sh graph <parent-container-id> --depth 1
bash <skill-dir>/bin/c3x.sh graph <component-id> --depth 1
```
Change affects interface -> check consumers first.

**Result:** Refs + rules + parent + graph = full constraint set. All honored.

**New component:** Top-down. Container `Components` + `Responsibilities` first, then component. Not done until parent has Parent Delta decision.

**Navigation:** Context -> Container -> Component

## Graph Output

Include mermaid when relationships matter:
```bash
bash <skill-dir>/bin/c3x.sh graph <entity-id> --format mermaid
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
No `.c3/` or re-onboard. Scaffold -> discovery -> inject CLAUDE.md -> show capabilities.
`references/onboard.md`

### query
`c3x list` for topology, `c3x lookup <file-or-glob>` for code->entity, `c3x graph <id> --direction reverse` for dependents. Full-text over bodies -> grep over `.c3/`.
`references/query.md`

### audit
`c3x check` -> `c3x list` -> semantic phases -> PASS/WARN/FAIL table.
`references/audit.md`

### change
ADR first -> `c3x list` -> affected entities -> `c3x lookup` files -> fill ADR -> approve -> execute -> `c3x check`. Provision gate: implement or `status: provisioned`.
`references/change.md`

### ref
Modes: Add / Update / List / Usage.
`references/ref.md`

### rule
Modes: Add / Update / List / Usage.
`references/rule.md`

### sweep
`c3x graph <id> --direction reverse` -> transitive deps -> parallel assessment -> synthesize. Advisory only.
`references/sweep.md`
