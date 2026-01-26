# c3-skill

C3 (Context-Container-Component) architecture design methodology for Claude Code. Structured top-down documentation with mermaid diagrams, ADR-first changes, and TDD implementation.

![Agent Ecosystem](https://diashort.apps.quickable.co/e/ee8922f6)

## Installation

```bash
claude plugin install c3-skill
```

## Quick Start

```bash
# Initialize C3 docs for a new project
/onboard

# Ask architecture questions
/query "where is authentication handled?"

# Make architectural changes
/alter "add rate limiting to API"

# General C3 operations
/c3 audit
```

## Commands

| Command | Description |
|---------|-------------|
| `/onboard` | Initialize C3 architecture documentation for a project |
| `/query <question>` | Navigate C3 docs - "where is X", "how does X work", "explain X" |
| `/alter <change>` | Make changes via ADR - "add component", "refactor X", "fix bug" |
| `/c3 [action]` | General C3 operations - audit, validate, route to appropriate workflow |

## Simple Example

### 1. Initialize C3 docs for your project

```bash
/onboard
```

Claude will analyze your codebase and ask questions:

```
? What is this system called?
> "TaskManager API"

? What are the main containers?
> "API server, PostgreSQL database, Redis cache"

? What does the API server handle?
> "User auth, task CRUD, notifications"
```

Result: `.c3/` directory with architecture documentation.

### 2. Ask architecture questions

```bash
/query "where is authentication handled?"
```

Response:
```
Authentication is handled in c3-1-api container:

- c3-101-auth-middleware: JWT validation, token refresh
- c3-102-user-service: Login, registration, password reset

See: .c3/c3-1-api/c3-101-auth-middleware.md
```

### 3. Make architectural changes

```bash
/alter "add rate limiting to the API"
```

Claude will:
1. Ask clarifying questions (which endpoints? what limits?)
2. Analyze impact (affected components, patterns)
3. Create ADR at `.c3/adr/adr-20260126-rate-limiting.md`
4. Ask: "Execute now or create plan only?"
5. If execute: implement with TDD, create tasks, verify

### 4. Check the result

```bash
/query "show me the rate limiting ADR"
```

```
ADR: adr-20260126-rate-limiting.md
Status: implemented

Decision: Add Redis-based rate limiting to public API endpoints
- 100 requests/minute for authenticated users
- 20 requests/minute for anonymous

Affected: c3-101-auth-middleware, c3-103-rate-limiter (new)
```

---

## Architecture

### Agent Ecosystem

![Agent Ecosystem](https://diashort.apps.quickable.co/e/ee8922f6)

### Agents

| Agent | Role | Model |
|-------|------|-------|
| **c3-navigator** | Answer architecture questions | inherit |
| **c3-orchestrator** | Orchestrate architectural changes via ADR | opus |
| **c3-dev** | Execute ADR with TDD workflow | opus |
| **c3-adr-transition** | Transition ADR from accepted to implemented | haiku |
| **c3-analyzer** | Analyze affected components (internal) | sonnet |
| **c3-impact** | Trace dependencies and risks (internal) | sonnet |
| **c3-patterns** | Check pattern compliance (internal) | sonnet |
| **c3-synthesizer** | Combine analysis into understanding (internal) | opus |
| **c3-summarizer** | Extract facts from docs (internal) | haiku |
| **c3-content-classifier** | Classify content for audit (internal) | haiku |

### Workflow: Making Changes

1. **Intent Clarification** - Socratic dialogue to understand what and why
2. **Parallel Analysis** - analyzer, impact, patterns run concurrently
3. **Synthesis** - Combine findings, validate all checks pass
4. **ADR Generation** - Create decision record with approved files
5. **ADR Acceptance** - User reviews and accepts, captures base commit
6. **Delegation** - User chooses: plan only, execute now, or manual
7. **Implementation** - c3-dev implements with TDD (if execute now)
8. **Transition** - ADR marked implemented after verification

### Workflow: c3-dev TDD

![TDD Workflow](https://diashort.apps.quickable.co/e/e43110a1)

**Task States:** `pending → in_progress → blocked → testing → implementing → completed`

## C3 Documentation Structure

![C3 Structure](https://diashort.apps.quickable.co/e/c08ec10d)

```
.c3/
├── README.md              # System context (c3-0)
├── TOC.md                 # Auto-generated table of contents
├── c3-1-{container}/      # Container documentation
│   ├── README.md          # Container overview
│   └── c3-1XX-{component}.md  # Component docs
├── refs/                  # Cross-cutting patterns
│   └── ref-{name}.md      # Pattern documentation
└── adr/                   # Architecture Decision Records
    └── adr-YYYYMMDD-{slug}.md
```

## Key Concepts

### ADR (Architecture Decision Record)

Every change goes through an ADR:
- **proposed** - Created, under review
- **accepted** - Approved, ready for implementation
- **implemented** - Code complete, verified

### Refs (Reference Patterns)

Cross-cutting patterns that components must follow:
- `ref-error-handling.md` - Error conventions
- `ref-auth.md` - Authentication patterns
- etc.

### Task-Based Implementation

c3-dev creates tasks for each work item:
```yaml
metadata:
  adr: adr-20260126-auth-refactor  # Links to ADR
  status: testing                   # Granular state
```

Summary task required for ADR transition:
```yaml
metadata:
  adr: adr-20260126-auth-refactor
  type: summary
  status: completed
```

## References

| File | Purpose |
|------|---------|
| `references/skill-harness.md` | Behavioral constraints, complexity rules |
| `references/layer-navigation.md` | How to traverse C3 docs |
| `references/adr-template.md` | ADR structure and lifecycle |
| `references/implementation-guide.md` | Component documentation patterns |
| `references/audit-checks.md` | Full audit procedure |

## License

MIT
