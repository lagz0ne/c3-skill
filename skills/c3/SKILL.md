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

| Command | Purpose |
|---------|---------|
| `init` | Scaffold `.c3/` |
| `list` | Topology with files (`--json`, `--flat`, `--compact`) |
| `check` | Structural validation (`--json`, `--fix`) |
| `add <type> <slug>` | Create entity (`--container`, `--feature`, `--json`) |
| `set <id> <field> <val>` | Update frontmatter field |
| `set <id> --section <name>` | Update section content (text or JSON table) |
| `wire <src> <tgt>` | Link component to ref (`--remove` to unlink) |
| `schema <type>` | Section definitions for entity type (`--json`) |
| `codemap` | Scaffold code-map entries for all components, refs + rules |
| `lookup <file-or-glob>` | File or glob → component + refs (`--json`) |
| `coverage` | Code-map coverage stats (JSON default) |
| `delete <id>` | Remove entity + clean all references (`--dry-run`) |
| `query <terms>` | Full-text search across all entities (`--type`, `--limit`, `--json`) |
| `impact <id>` | Transitive impact analysis — who depends on this? (`--depth`, `--json`) |
| `diff` | Show uncommitted entity changes (`--mark <hash>`, `--json`) |
| `export [dir]` | Dump DB to markdown files (escape hatch) |
| `migrate` | Populate content node tree (v7→v8, requires DB) |
| `migrate-legacy` | Import .c3/ markdown files into database (v6→v7, no DB needed) |
| `marketplace add <url>` | Register marketplace rule source (shallow clone) |
| `marketplace list` | Browse available rules (`--source`, `--tag`, `--json`) |
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
bash <skill-dir>/bin/c3x.sh list --json
```
- If output contains "contains markdown files but no database" → route to **migrate**
- Fails/empty → route to **onboard**

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
```bash
bash <skill-dir>/bin/c3x.sh lookup <file-path>
bash <skill-dir>/bin/c3x.sh lookup 'src/auth/**'   # glob for directory-level context
```
Returned refs = hard constraints, every one MUST be honored.
Run the moment any file path surfaces. Use glob when working across a directory.
No match = uncharted, proceed with caution.

**Layer Navigation:** Context → Container → Component

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
`c3x query "<terms>"` for search, `c3x list --json` for topology, `c3x impact <id>` for dependencies.
Details: `references/query.md`

### audit
`c3x check` → `c3x list --json` → semantic phases. Output: PASS/WARN/FAIL table.
Details: `references/audit.md`

### change
ADR first (`c3x add adr --json`) → `c3x list --json` → identify affected entities (refs, affects in frontmatter) → `c3x lookup` each file → fill ADR → approve → execute → `c3x check`.
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

