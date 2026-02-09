# c3-skill

C3 (Context-Container-Component) architecture methodology for Claude Code. Concept-first documentation with Agent Teams for coordinated change.

## Installation

```bash
claude plugin install c3-skill
```

## Skills

| Skill | Purpose |
|-------|---------|
| `c3-onboard` | Create C3 docs from scratch via staged Socratic discovery |
| `c3-query` | Navigate architecture docs and answer questions |
| `c3-change` | Coordinated change workflow via Agent Teams (ADR-first) |
| `c3-ref` | Manage cross-cutting chosen options (patterns, conventions) |
| `c3-audit` | Audit C3 docs for consistency, drift, and completeness |

## Quick Example

```bash
# 1. Initialize (Socratic discovery of your system)
# Use the c3-onboard skill

# 2. Query (find and explain architecture)
# "where is authentication handled?"
# → c3-1-api: c3-101-auth-middleware (foundation)

# 3. Change (Agent Teams workflow)
# "add rate limiting"
# → Understand → ADR → Execute (teammates) → Audit
```

## Architecture

### Agents

| Agent | Role | Model |
|-------|------|-------|
| `c3-navigator` | Answer architecture questions via doc traversal | inherit |
| `c3-lead` | Team lead for Agent Teams change workflow | opus |

### Change Workflow (Agent Teams)

```
Understand → ADR → Execute → Audit
    │          │        │        │
 analyst    lead    implement  auditor
 reviewer           teammates  teammate
```

**4-Phase Flow:**

1. **Understand** — Analyst + reviewer teammates analyze impact, check refs
2. **ADR** — Lead creates decision record with work breakdown
3. **Execute** — Implementer teammates execute tasks, lead reviews
4. **Audit** — Auditor teammate verifies docs match code reality

**Regression:** Lead can regress to earlier phases when drift is discovered (ADR-anchored decision tree).

## Documentation Structure

```
.c3/
├── README.md                    # System context (abstract constraints)
├── c3-N-{container}/            # Container docs (responsibility allocation)
│   ├── README.md                # Container overview + boundary
│   └── c3-NXX-{component}.md   # Foundation (01-09) or Feature (10+)
├── refs/                        # Cross-cutting chosen options
│   └── ref-{name}.md            # Choice/Why/How/Not This/Scope/Override
└── adr/                         # Decision records
    └── adr-YYYYMMDD-{slug}.md   # proposed → accepted → implemented
```

## Key Concepts

| Concept | Description |
|---------|-------------|
| **Context** | Abstract constraints and system-level goals |
| **Container** | Deployment boundary + responsibility allocator |
| **Foundation** | Platform capabilities others depend on (01-09) |
| **Feature** | Business logic composing foundations (10+) |
| **Ref** | Cross-cutting chosen option — Choice, Why, How, Scope, Override |
| **ADR** | Architecture Decision Record with work breakdown |

## References

| File | Purpose |
|------|---------|
| `references/skill-harness.md` | Behavioral constraints for all skills |
| `references/layer-navigation.md` | C3 doc traversal rules |
| `references/adr-template.md` | ADR structure and lifecycle |
| `references/audit-checks.md` | 10-phase audit procedure |
| `references/component-categories.md` | Foundation vs Feature vs Ref rules |

## License

MIT
