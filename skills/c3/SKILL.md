---
name: c3
description: |
  Architecture documentation system for codebases with a .c3/ directory. Use this skill
  whenever the user invokes /c3, mentions C3/c3, asks about project architecture, wants
  to onboard/audit/change/query architecture docs, manage patterns (refs), or assess
  impact of changes. Trigger phrases include: "adopt C3", "onboard", "where is X in the
  architecture", "audit docs", "check if docs are up to date", "add a component",
  "implement feature", "what breaks if I change X", "add a ref", "trace the flow".
  Even if the user doesn't say "c3" explicitly, use this skill if the project has a .c3/
  directory and the request relates to architecture documentation.
---

# C3

CLI: `bash <skill-dir>/bin/c3x.sh <command> [args]`

| Command | Purpose |
|---------|---------|
| `init` | Scaffold `.c3/` |
| `list` | Topology (`--json`, `--flat`, `--compact`) |
| `check` | Structural validation (`--json`, `--include-adr`) |
| `add <type> <slug>` | Create entity (`--container`, `--feature`) |
| `codemap` | Scaffold `.c3/code-map.yaml` stubs |
| `lookup <file-or-glob>` | File → component + refs (`--json`) |
| `coverage` | Code-map coverage stats |
| `set <id> <field> <value>` | Update frontmatter field or section |
| `wire <src> cite <tgt>` | Link component → ref citation |
| `unwire <src> cite <tgt>` | Remove component → ref citation |
| `schema <type>` | Show known sections for entity type |

Types for `add`: `container`, `component`, `ref`, `adr`, `recipe`

---

## Intent Classification

| Intent | Op | Reference |
|--------|----|-----------|
| Adopt, onboard, scaffold, bootstrap, "set up architecture" | **onboard** | `references/onboard.md` |
| Where, explain, how, trace, "show me", "what is", diagram | **query** | `references/query.md` |
| Audit, validate, "check docs", drift, "docs up to date" | **audit** | `references/audit.md` |
| Add, change, fix, implement, refactor, remove, migrate, design | **change** | `references/change.md` |
| Pattern, convention, "create ref", "update ref", standardize | **ref** | `references/ref.md` |
| Impact, "what breaks", assess, sweep, "is this safe" | **sweep** | `references/sweep.md` |

Ambiguous intent → `AskUserQuestion` listing the 6 operations.

---

## Dispatch

1. Classify intent from the table above
2. Load `references/<op>.md`
3. Execute (use Task tool for parallelism when appropriate)

**Precondition** — before every op except onboard:
```bash
bash <skill-dir>/bin/c3x.sh list --json
```
Fails or empty → route to **onboard**.

---

## ASSUMPTION_MODE

Interactive questions (via `AskUserQuestion`) help clarify intent, scope, and approval. But some users prefer speed over confirmation.

**How it activates:** If the user declines or dismisses an `AskUserQuestion` prompt (e.g., says "just do it", "skip questions", or selects a dismissive option), set `ASSUMPTION_MODE = true` for the rest of the session.

**Once active:**
- Stop asking `AskUserQuestion` entirely
- For high-impact decisions: state what you're assuming and mark `[ASSUMED]` so the user can correct if needed
- For low-impact decisions: auto-proceed silently

---

## Core Principles

These apply across all operations. Correctness is the priority — always maximize accuracy of documentation over speed.

**1. Validate after mutations** — Run `c3x check` after creating or editing any `.c3/` doc. It catches broken frontmatter, missing fields, bad references. Treat errors (`✗`) as blockers.

**2. ADR-first for changes** — Every change operation starts with `c3x add adr <slug>` before any code exploration. The ADR is an ephemeral work order that captures intent, impact, and work breakdown. It drives what gets updated, then becomes hidden. Exception: ref-add creates its adoption ADR at completion.

**3. File context via lookup** — When modifying code or making decisions about specific files, run `c3x lookup <file-path>` to understand which component owns the file and which refs (patterns/constraints) govern it. This is especially important during change and sweep operations. For read-only queries where you're tracing code, use lookup selectively — it's a tool for understanding constraints, not a gate before every file read.

**4. Navigate top-down** — Context → Container → Component. Don't jump straight to a component without understanding its container and the system context.

**5. LSP first, then search** — When exploring or navigating code, always prefer LSP tools (go-to-definition, find-references, hover, workspace symbols) over Grep/Glob. LSP provides precise, type-aware results. Only fall back to Grep/Glob when LSP is unavailable or returns no results. This applies to all operations that involve reading or understanding source code.

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
No `.c3/` or re-onboard. Scaffold → discover → document → validate.
Details: `references/onboard.md`

### query
Load topology → match entity → read docs + explore code → respond.
Details: `references/query.md`

### audit
CLI scan → goal-oriented exploration → synthesize findings.
Details: `references/audit.md`

### change
ADR first → understand impact → approve → execute → validate.
Provision gate: implement now or `status: provisioned`.
Details: `references/change.md`

### ref
Modes: Add / Update / List / Usage.
Details: `references/ref.md`

### sweep
Topology → affected entities → assess constraints → synthesize. Advisory only.
Details: `references/sweep.md`
