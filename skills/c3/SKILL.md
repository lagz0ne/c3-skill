---
name: c3
description: >
  This skill should be used when the user invokes /c3 or asks architecture questions
  about a project with a .c3/ directory. Trigger phrases: "adopt C3", "onboard this
  project", "where is X", "audit the architecture", "check docs", "add a component",
  "implement feature", "what breaks if I change X", "add a ref", "add a coding rule",
  "coding standard". Handles operations: onboard, query, audit, change, ref, rule, sweep.
  Classifies intent, loads ref, executes.
---

# C3

CLI: `C3X_MODE=agent bash <skill-dir>/bin/c3x.sh <command> [args]`

**Agent output mode:** With `C3X_MODE=agent`, c3x outputs TOON (Token-Optimized Object Notation) by default — ~40% fewer tokens than JSON. All commands append `help[]` contextual hints suggesting next steps.

**Content-first:** Running c3x with no arguments outputs a project dashboard (entity counts, coverage, pending ADRs) — not help text. Use `--help` for help.

| Command | Purpose |
|---------|---------|
| (no args) / `status` | Project dashboard: entity counts, coverage %, pending ADRs, warnings |
| `init` | Scaffold `.c3/` |
| `list` | Topology with files (`--flat`, `--compact`) |
| `check` | Structural validation (`--fix`) |
| `add <type> <slug>` | Create entity with body via stdin (`--container`, `--feature`) |
| `set <id> <field> <val>` | Update frontmatter field |
| `set <id> --section <name>` | Update section content (text or JSON table) |
| `wire <src> <tgt>` | Link component to ref (`--remove` to unlink) |
| `schema <type>` | Section definitions for entity type |
| `codemap` | Scaffold code-map entries for all components, refs + rules |
| `read <id>` | Full entity content; agent mode truncates body to 1500 chars (`--full` to bypass) |
| `lookup <file-or-glob>` | File or glob → component + refs |
| `coverage` | Code-map coverage stats |
| `delete <id>` | Remove entity + clean all references (`--dry-run`) |
| `query <terms>` | Full-text search across all entities (`--type`, `--limit`) |
| `impact <id>` | Transitive impact analysis — who depends on this? (`--depth`) |
| `diff` | Show uncommitted entity changes (`--mark <hash>`) |
| `export [dir]` | Dump DB to markdown files (escape hatch) |
| `migrate` | Populate content node tree (v7→v8, requires DB) |
| `migrate-legacy` | Import .c3/ markdown files into database (v6→v7, no DB needed) |
| `marketplace add <url>` | Register marketplace rule source (shallow clone) |
| `marketplace list` | Browse available rules (`--source`, `--tag`) |
| `marketplace show <rule-id>` | Preview marketplace rule content |
| `marketplace update` | Pull latest from registered sources |
| `marketplace remove <name>` | Unregister source + delete cache |

Types for `add`: `container`, `component`, `ref`, `rule`, `adr`, `recipe`

---

## Intent Classification

| Keywords | Op | Reference |
|----------|----|-----------|
| adopt, init, scaffold, bootstrap, onboard, "create .c3", "set up architecture" | **onboard** | `references/onboard.md` |
| where, explain, how, diagram, trace, "show me", "what is", "list components" | **query** | `references/query.md` |
| audit, validate, "check docs", drift, "docs up to date", "verify docs" | **audit** | `references/audit.md` |
| add, change, fix, implement, refactor, remove, provision, design | **change** | `references/change.md` |
| "migrate to db", "convert to database", "upgrade c3", "migrate .c3" | **migrate** | `references/migrate.md` |
| pattern, convention, "create ref", "update ref", "list refs", standardize | **ref** | `references/ref.md` |
| "add/create a coding rule", "document a rule", "coding standard", "coding convention", "migrate refs to rules", "split ref into rule" | **rule** | `references/rule.md` |
| marketplace, "browse rules", "adopt rule", "install rule from", "available rules" | **rule** (Adopt mode) | `references/rule.md` |
| impact, "what breaks", assess, sweep, "is this safe" | **sweep** | `references/sweep.md` |
| recipe, "trace end-to-end", "cross-cutting flow", "how does X flow" | **query** (read) / **change** (create) | `references/query.md` / `references/change.md` |

---

## Dispatch

1. Classify op (ambiguous → `AskUserQuestion` with 7 options)
2. Load `references/<op>.md`
3. Execute (use Task tool for parallelism)

---

## Precondition

Before every op except onboard and migrate:
```bash
bash <skill-dir>/bin/c3x.sh
```
Returns project dashboard (TOON). If error about missing `.c3/` → route to **onboard**. If mentions "markdown files but no database" → route to **migrate**. Otherwise, dashboard gives immediate orientation: entity counts, coverage, pending ADRs. Follow the `help[]` hints for next steps.

---

## CoT Harness

C3 context is already loaded. Every file, every directory has a place in the topology. Before touching anything, let C3 tell you what it knows — then follow its lead.

`c3x lookup <file-or-glob>` → component docs, refs, constraints. Read them. They supersede assumptions.

When context shifts to a new component mid-task, look it up again. The topology is the source of truth — just follow it.

---

## ASSUMPTION_MODE

First `AskUserQuestion` denial → `ASSUMPTION_MODE = true` for session.
- Never call `AskUserQuestion` again
- High-impact: state assumption, mark `[ASSUMED]`
- Low-impact: auto-proceed

---

## Shared Rules

**HARD RULE — .c3/ is CLI-only. NEVER use Read, Glob, Edit, or Write tools on `.c3/` files.**

Architecture data lives in `.c3/c3.db` (SQLite). Raw file access bypasses the database and returns stale or empty content. ALL access goes through c3x:

- Create:   `c3x add`, `c3x init`, `c3x codemap`
- Read:     `c3x read <id>`, `c3x list`, `c3x query`, `c3x lookup`, `c3x graph`, `c3x impact`
- Update:   `c3x write <id>`, `c3x set`, `c3x wire` (`--remove` to unwire)
- Delete:   `c3x delete`
- Validate: `c3x check`, `c3x coverage`, `c3x schema`

If c3x lacks a needed operation, STOP and tell the user — do not work around it with file tools.

**Run `c3x check` frequently** — after any mutation (`add`, `write`, `set`, `wire`, `delete`). It catches missing required sections, bad entity references, and codemap issues. Treat errors as blockers.

**HARD RULE — ADR is the unit of change:**
Every **change** operation MUST start with `c3x add adr <slug>` as its FIRST action.
No code reads, no file edits, no exploration before the ADR exists.
(Exception: **ref-add** creates its adoption ADR at completion — see `references/ref.md`.)
The ADR is an ephemeral work order — it drives what to update, then gets hidden.
`c3x list` and `c3x check` exclude ADRs by default; use `--include-adr` to see them.

**Stop immediately if:**
- No ADR exists for current change → `c3x add adr <slug>` NOW
- Guessing intent → `AskUserQuestion` (skip if ASSUMPTION_MODE)
- Jumping to component → start Context down
- Updating docs without code check

**File Context — MANDATORY before reading or altering any file:**

**Step 1 — Lookup:**
```bash
bash <skill-dir>/bin/c3x.sh lookup <file-path>
bash <skill-dir>/bin/c3x.sh lookup 'src/auth/**'   # glob for directory-level context
```
Run the moment any file path surfaces. Use glob when working across a directory.
No match = uncharted, proceed with caution.

**Step 2 — Load rules:** For every `rule-*` returned by lookup:
```bash
bash <skill-dir>/bin/c3x.sh read <rule-id>
```
Extract `## Rule`, `## Golden Example`, and `## Not This`. These are hard constraints — code MUST match the golden pattern. Deviations require an Override section in the rule or a new ADR.

**Step 3 — Graph context:** For the first component returned (or each, if few):
```bash
bash <skill-dir>/bin/c3x.sh graph <component-id> --depth 1
```
Shows providers (what this component depends on) and consumers (what depends on it). If your change affects the component's interface, check consumers before proceeding.

**Result:** Returned refs + loaded rule content + graph = the full constraint set. All MUST be honored.

**Layer Navigation:** Context → Container → Component

**Graph Output — Include mermaid graph in responses when relationships matter:**
```bash
bash <skill-dir>/bin/c3x.sh graph <entity-id> --format mermaid
```
Include the output as a mermaid code block. Root selection matters more than depth:
- Graph from a **container** to show its components + cited refs (query, audit, onboard)
- Graph from a **component** to show its constraints and siblings (query, change, sweep)
- Graph from a **ref/rule** to show citation graph (ref/rule Usage, sweep)
- `--depth 1` is default. Use `--depth 2` only for cross-container tracing.
- `--direction forward` for impact. `--direction reverse` for "what depends on this".
- Never graph from `c3-0` — it's always exactly one node, adds no signal.

**File Structure:**
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
No `.c3/` or re-onboard. `c3x init` → discovery → inject CLAUDE.md → show capabilities.
Details: `references/onboard.md`

### query
`c3x query "<terms>"` for search, `c3x list` for topology, `c3x impact <id>` for dependencies.
Details: `references/query.md`

### audit
`c3x check` → `c3x list` → semantic phases. Output: PASS/WARN/FAIL table.
Details: `references/audit.md`

### change
ADR first (`c3x add adr`) → `c3x list` → identify affected entities (refs, affects in frontmatter) → `c3x lookup` each file → fill ADR → approve → execute → `c3x check`.
Provision gate: implement now or `status: provisioned`.
Details: `references/change.md`

### ref
Modes: Add / Update / List / Usage.
Details: `references/ref.md`

### rule
Modes: Add / Update / List / Usage.
Details: `references/rule.md`

### sweep
`c3x impact <id>` → transitive dependency analysis → parallel assessment → synthesize. Advisory only.
Details: `references/sweep.md`

### migrate
Two paths: **v6→v7** (file→DB via `c3x migrate-legacy`) and **v7→v8** (body→nodes via `c3x migrate`). Both are LLM-assisted with evidence gates: dry-run → repair → validate zero issues → migrate → verify content fidelity. Warnings are errors — every warning means silent data loss.
Details: `references/migrate.md`

