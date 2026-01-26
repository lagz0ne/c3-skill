# c3-skill

C3 (Context-Container-Component) architecture methodology for Claude Code. Top-down documentation with ADR-first changes and TDD implementation.

![Agent Ecosystem](https://diashort.apps.quickable.co/e/ee8922f6)

## Installation

```bash
claude plugin install c3-skill
```

## Commands

| Command | Triggers | Purpose |
|---------|----------|---------|
| `/onboard` | "adopt C3", "init C3", "scaffold docs" | Initialize C3 documentation for a project |
| `/query` | "where is X", "how does X work", "explain X" | Navigate architecture docs |
| `/alter` | "add component", "refactor X", "fix bug" | Make changes via ADR workflow |
| `/c3` | "audit", "validate", general C3 ops | Route to appropriate workflow |

## Skills

| Skill | Triggers | Purpose |
|-------|----------|---------|
| `c3-query` | "where/what is X", "what does X do", "explain/describe X", "diagram of X", "dependencies of X", "who/what calls X", "trace X", "flow of X" | Deep architecture navigation |
| `c3-alter` | "add/change/edit/update/remove X", "implement/extend/improve X", "make X do Y", "refactor X", "fix bug", "rename/move/split/merge/delete component" | Full ADR change workflow |
| `c3-ref` | "add/define pattern", "best practice for X", "coding standard", "guideline for X", "how should we X", "what's the convention", "list refs" | Manage cross-cutting patterns |
| `onboard` | "adopt/init/initialize/start/bootstrap C3", "create C3 docs", "start documenting", "document this project/codebase" | Staged Socratic discovery |
| `c3` | "audit/validate C3", "check/verify docs", "sync/refresh docs", "is documentation current", "docs out of sync" | Routing + audit |

## Quick Example

```bash
# 1. Initialize (asks questions about your system)
/onboard

# 2. Query (finds and explains architecture)
/query "where is authentication handled?"
# → c3-1-api: c3-101-auth-middleware, c3-102-user-service

# 3. Change (ADR workflow with impact analysis)
/alter "add rate limiting"
# → Clarify → Analyze → ADR → Accept → Execute (optional)
```

## Architecture

### Agents

**User-facing:**

| Agent | Role | Model |
|-------|------|-------|
| `c3-navigator` | Answer architecture questions | inherit |
| `c3-orchestrator` | Orchestrate changes via ADR | opus |
| `c3-dev` | Execute ADR with TDD | opus |
| `c3-adr-transition` | Transition ADR to implemented | haiku |

**Internal (dispatched by orchestrator):**

| Agent | Role | Model |
|-------|------|-------|
| `c3-analysis` | Comprehensive analysis (state, impact, patterns) | sonnet |
| `c3-synthesizer` | Combine analysis, validate readiness | opus |
| `c3-summarizer` | Extract facts from docs | haiku |
| `c3-content-classifier` | Classify content for audit | haiku |

### Change Workflow

```
Intent → Analysis → Synthesis → ADR → Accept → Execute → Transition
             │                          │
      c3-analysis       ← validation gate ──┘
      (state+impact+patterns)
```

1. **Clarify** - Socratic dialogue (what? why? scope?)
2. **Analyze** - c3-analysis performs state, impact, patterns
3. **Synthesize** - Combine findings, validate checks
4. **ADR** - Create decision record with approved files
5. **Accept** - User reviews, captures base commit
6. **Execute** - c3-dev implements with TDD (optional)
7. **Transition** - Mark implemented after verification

### TDD Workflow (c3-dev)

![TDD Workflow](https://diashort.apps.quickable.co/e/e43110a1)

Task states: `pending → in_progress → blocked → testing → implementing → completed`

## Documentation Structure

```
.c3/
├── README.md                    # System context (c3-0)
├── TOC.md                       # Auto-generated TOC
├── c3-N-{container}/            # Container docs
│   ├── README.md                # Container overview
│   └── c3-NXX-{component}.md    # Component docs
├── refs/                        # Cross-cutting patterns
│   └── ref-{name}.md            # Pattern constraints
└── adr/                         # Decision records
    └── adr-YYYYMMDD-{slug}.md   # ADR lifecycle
```

![C3 Structure](https://diashort.apps.quickable.co/e/c08ec10d)

## Proactive Pattern Awareness

C3 provides **ambient context** without explicit invocation:

### SessionStart
When entering a C3-documented project, Claude automatically learns:
- System goal and key decisions
- All refs (patterns) with their purposes
- File → component mapping

### Edit/Write Context
When editing files in documented areas, Claude receives:
- Which component owns the file
- Applicable patterns (refs)
- Guidance to maintain consistency

This enables **drift prevention** - changes naturally align with established patterns.

## Key Concepts

| Concept | Description |
|---------|-------------|
| **ADR** | Architecture Decision Record. States: `proposed → accepted → implemented` |
| **Ref** | Cross-cutting pattern. Components must follow; violations require override justification |
| **Task** | c3-dev work item linked to ADR via metadata |

## Hooks

| Hook | Script | Purpose |
|------|--------|---------|
| SessionStart | `c3-context-loader` | Load patterns + file mapping into context |
| PreToolUse (Edit/Write) | `c3-edit-context` | Surface applicable refs when editing |
| PreToolUse (Edit/Write) | `c3-gate` | Gate edits to ADR-approved files |
| PreToolUse (Bash) | `pre-commit-toc` | Rebuild TOC on git commit |

## References

| File | Purpose |
|------|---------|
| `references/skill-harness.md` | Behavioral constraints |
| `references/layer-navigation.md` | C3 doc traversal |
| `references/adr-template.md` | ADR structure |
| `references/audit-checks.md` | Audit procedure |

## License

MIT
