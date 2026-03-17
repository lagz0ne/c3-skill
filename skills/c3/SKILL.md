---
name: c3
description: |
  This skill should be used when the user invokes /c3 or asks architecture questions
  about a project with a .c3/ directory. Trigger phrases: "adopt C3", "onboard this
  project", "where is X", "audit the architecture", "check docs", "add a component",
  "implement feature", "what breaks if I change X", "add a ref". Handles operations:
  onboard, query, audit, change, ref, sweep. Classifies intent, loads ref, executes.

  <example>
  user: "adopt C3 for this project"
  assistant: "Using c3 to onboard this project."
  </example>

  <example>
  user: "where is auth in the C3 docs?"
  assistant: "Using c3 to query the architecture."
  </example>

  <example>
  user: "add a new API component"
  assistant: "Using c3 to orchestrate the change."
  </example>

  <example>
  user: "what breaks if I change the auth API?"
  assistant: "Using c3 to assess impact."
  </example>

  <example>
  user: "audit C3 docs for drift"
  assistant: "Using c3 to audit."
  </example>

  <example>
  user: "add a ref for error handling"
  assistant: "Using c3 to create a ref."
  </example>
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
| `codemap` | Scaffold `.c3/code-map.yaml` with stubs for all components + refs |
| `lookup <file-or-glob>` | File or glob → component + refs (`--json`) |
| `coverage` | Code-map coverage stats (JSON default) |
| `delete <id>` | Remove entity + clean all references (`--dry-run`) |

Types for `add`: `container`, `component`, `ref`, `adr`, `recipe`

---

## Intent Classification

| Keywords | Op | Reference |
|----------|----|-----------|
| adopt, init, scaffold, bootstrap, onboard, "create .c3", "set up architecture" | **onboard** | `references/onboard.md` |
| where, explain, how, diagram, trace, "show me", "what is", "list components" | **query** | `references/query.md` |
| audit, validate, "check docs", drift, "docs up to date", "verify docs" | **audit** | `references/audit.md` |
| add, change, fix, implement, refactor, remove, migrate, provision, design | **change** | `references/change.md` |
| pattern, convention, "create ref", "update ref", "list refs", standardize | **ref** | `references/ref.md` |
| impact, "what breaks", assess, sweep, "is this safe" | **sweep** | `references/sweep.md` |
| recipe, "trace end-to-end", "cross-cutting flow", "how does X flow" | **query** (read) / **change** (create) | `references/query.md` / `references/change.md` |

---

## Dispatch

1. Classify op (ambiguous → `AskUserQuestion` with 6 options)
2. Load `references/<op>.md`
3. Execute (use Task tool for parallelism)

---

## Precondition

Before every op except onboard:
```bash
bash <skill-dir>/bin/c3x.sh list --json
```
Fails/empty → route to **onboard**

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

**HARD RULE — .c3/ is CLI-only:**
NEVER use Edit, Write, or any file tool on files inside `.c3/`. ALL mutations go through c3x:
- Create: `c3x add`, `c3x init`, `c3x codemap`
- Read:   `c3x list`, `c3x schema`, `c3x lookup`, `c3x coverage`
- Update: `c3x set`, `c3x wire` (`--remove` to unwire)
- Delete: `c3x delete`
- Validate: `c3x check`

Exception: `code-map.yaml` patterns may be edited directly (flat YAML, no integrity constraints).
If c3x lacks a needed mutation, STOP and tell the user — do not work around it.

**Run `c3x check` frequently** — after creating/editing any `.c3/` doc. It catches broken YAML frontmatter, missing required sections, bad entity references, and codemap issues. Treat errors (`✗`) as blockers.

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
`c3x list --json` → match entity (includes refs, affects, files) → Read doc → explore code.
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

### sweep
`c3x list --json` → filter by refs/affects to find affected entities → parallel assessment → synthesize. Advisory only.
Details: `references/sweep.md`

