---
name: c3
description: |
  C3 architecture toolkit. Manages architecture docs in .c3/ directory.
  Operations: onboard, query, audit, change, ref, sweep.
  Triggers: /c3, architecture questions when .c3/ exists, C3 docs management.
  Classifies intent from request, loads operation-specific reference, executes.

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
---

# C3

Single skill for all C3 architecture operations.

## CLI

Binary: `bin/c3x.sh` (relative to this skill's directory)

All CLI calls:
```bash
bash <skill-dir>/bin/c3x.sh <command> [args]
```

| Command | Purpose |
|---------|---------|
| `init` | Scaffold `.c3/` with templates |
| `list` | Topology: entities, relationships, frontmatter (`--json`, `--flat`) |
| `check` | Structural validation: broken links, orphans, duplicates (`--json`) |
| `add <type> <slug>` | Create entity with auto-numbering + wiring (`--container`, `--feature`) |

Types for `add`: `container`, `component`, `ref`, `adr`

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

---

## Dispatch

1. Analyze request -> classify op from table above
2. Ambiguous -> `AskUserQuestion` with 6 ops as options
3. Load `references/<op>.md` (Read file relative to this skill's dir)
4. Execute per reference instructions
5. Use Task tool for parallelism when reasoning suggests it helps (anonymous subagents)

---

## Precondition

Run before every operation (except onboard):

```bash
bash <skill-dir>/bin/c3x.sh list --json
```

- Fails or empty -> route to **onboard**
- Exception: onboard does NOT need existing `.c3/`

---

## ASSUMPTION_MODE

If `AskUserQuestion` is denied even once -> set `ASSUMPTION_MODE = true` for the rest of this session.

When `ASSUMPTION_MODE` is true:
- NEVER call `AskUserQuestion` again
- High-impact decisions: state assumption clearly, mark `[ASSUMED]`
- Low-impact decisions: auto-proceed silently
- Every instruction that says "Use AskUserQuestion" becomes "make your best assumption and mark [ASSUMED]"

---

## Shared Rules

**Complexity-First Documentation:**

| Level | Signals | Doc Depth |
|-------|---------|-----------|
| trivial | Single purpose, stateless | Purpose + deps only |
| simple | Few concerns, basic state | + key components |
| moderate | Multiple concerns, caching | + discovered aspects |
| complex | Orchestration, security-critical | Full aspect discovery |
| critical | Distributed txns, compliance | Comprehensive + rationale |

**Red Flags — STOP Immediately:**
- Editing code without ADR -> create ADR first
- Guessing user intent -> `AskUserQuestion` (skip if ASSUMPTION_MODE)
- Jumping to component -> start from Context down
- Updating docs without code check -> verify code first

**Component Categories:**
- Foundation (01-09): infrastructure, `## Code References` required
- Feature (10+): business logic, `## Code References` required
- Ref: conventions only, NO `## Code References`

**Layer Navigation:** Always top-down: Context -> Container -> Component

**File Structure:**
```
.c3/
├── README.md                    # Context (c3-0)
├── adr/                         # Architecture Decision Records
│   └── adr-YYYYMMDD-slug.md
├── refs/                        # Cross-cutting patterns
│   └── ref-slug.md
└── c3-N-name/                   # Container
    ├── README.md                # Container overview
    └── c3-NNN-component.md      # Component doc
```

---

## Operations

### onboard
- **Pre:** No `.c3/` or user wants re-onboard
- **Flow:** `c3x init` -> Socratic discovery (3 stages: Inventory -> Details -> Finalize) -> gate each stage
- **Post:** Inject CLAUDE.md routing, show capabilities overview
- **Details:** Load `references/onboard.md`

### query
- **Pre:** `.c3/` exists (precondition check)
- **Flow:** `c3x list --json` -> match entity -> Read doc -> explore code
- **Subagents:** parallel container exploration when multi-scope
- **Details:** Load `references/query.md`

### audit
- **Pre:** `.c3/` exists (precondition check)
- **Flow:** `c3x check` (structural) -> `c3x list --json` (inventory) -> Phases 2-10 (semantic via Read+Grep+reasoning)
- **Output:** Table of phases with PASS/WARN/FAIL + action items
- **Details:** Load `references/audit.md`

### change
- **Pre:** `.c3/` exists (precondition check)
- **Flow:** `c3x list --json` -> impact analysis -> ADR (`c3x add adr <slug>`) -> user approves -> execute (`c3x add component/ref` for scaffolding) -> audit (`c3x check` + verify)
- **Provision gate:** after ADR approval, ask implement now or design only
- **Details:** Load `references/change.md`

### ref
- **Pre:** `.c3/` exists (precondition check)
- **Modes:** Add (`c3x add ref <slug>` -> fill -> discover usage -> update citings), Update (find citings -> check compliance -> surface impact), List (`c3x list --json` -> filter type=ref), Usage (find citings from JSON)
- **Details:** Load `references/ref.md`

### sweep
- **Pre:** `.c3/` exists (precondition check)
- **Flow:** `c3x list --json` -> identify affected entities -> parallel per-entity assessment via subagents -> synthesize constraints, risks, recommendations
- **Advisory only** — route to change for implementation
- **Details:** Load `references/sweep.md`

---

## CLAUDE.md Injection (onboard)

After onboard completes, inject into project CLAUDE.md:

```markdown
# Architecture
This project uses C3 docs in `.c3/`.
For architecture questions, changes, audits -> `/c3`.
Operations: query, audit, change, ref, sweep.
```

## Capabilities Reveal (onboard)

Show after onboard:

```
## Your C3 toolkit is ready

| Command | What it does |
|---------|-------------|
| `/c3` query | Ask about architecture |
| `/c3` audit | Validate docs |
| `/c3` change | Modify architecture |
| `/c3` ref | Manage patterns |
| `/c3` sweep | Impact assessment |

Just say `/c3` + what you want.
```
