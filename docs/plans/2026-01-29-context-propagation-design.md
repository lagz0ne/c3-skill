# C3 Context Propagation Design

## Problem

Claude (and subagents) rush to read code instead of checking C3 architecture docs and refs first. Current hooks fire too late:

- **SessionStart hook** - Context gets lost over time, especially in subagents
- **PreToolUse on Edit/Write** - By the time this fires, Claude has already made decisions

The component-to-refs relationship is real and important. Changing code in a specific area (e.g., express integration) needs relevant context (foundational patterns, refs for error handling, logging practices) loaded to enforce consistency.

## Solution

Use Claude's native CLAUDE.md loading instead of runtime hooks. Generate CLAUDE.md files in strategic positions based on C3 component Code References.

### How It Works

```
C3 Component (c3-201-auth.md)
  ├─ Code References: src/routes/auth/, src/middleware/auth.ts
  ├─ Related Refs: ref-error-handling, ref-logging
  │
  ▼ /c3 apply generates
  │
src/routes/auth/CLAUDE.md
  ├─ "This directory is part of c3-201 (Authentication)"
  ├─ "Follow these patterns: ref-error-handling, ref-logging"
  └─ "Read .c3/refs/ref-*.md for details"
```

### Benefits Over Hooks

| Hook approach | CLAUDE.md approach |
|---------------|-------------------|
| Runtime injection | Pre-built, always there |
| Complex caching logic | No caching needed |
| Subagent issues | Works natively for all agents |
| Hook timing issues | Context loaded at session start |
| Requires hook infrastructure | Just files in the repo |

## Design

### CLAUDE.md Template (Minimal Pointer)

```markdown
<!-- c3-generated: c3-201 -->
# c3-201: Authentication Service

Before modifying this code, read:
- Component: `.c3/c3-2-api/c3-201-auth.md`
- Patterns: `ref-error-handling`, `ref-logging`

Full refs: `.c3/refs/ref-{name}.md`
<!-- end-c3-generated -->
```

The `<!-- c3-generated: ID -->` marker enables:
- Safe regeneration (script knows which blocks to replace)
- Component ownership tracking
- Stale/orphan detection

### Conflict Handling

When CLAUDE.md already exists:

```
If CLAUDE.md exists:
  ├─ Has <!-- c3-generated --> block? → Replace block only
  └─ No block? → Append block at end
Else:
  └─ Create new file with block
```

User content outside the c3-generated block is preserved.

### Files Are Committed

Generated CLAUDE.md files are committed to the repo:
- Always present, even for new contributors
- Version controlled, visible in PRs
- No build step required for context to work

## Commands

### `/c3 audit` (Enhanced)

Add **Phase 10: Context Files** to existing audit:

```
Phase 10: Context Files

For each Component with Code References:
  1. Extract directory paths from Code References section
  2. For each directory:
     - Check if CLAUDE.md exists
     - Check if <!-- c3-generated: {component-id} --> block exists
     - Check if block content matches current component/refs

Output:
| Status | Meaning |
|--------|---------|
| Missing | CLAUDE.md doesn't exist in documented path |
| Stale | c3-generated block doesn't match current refs |
| Orphaned | c3-generated block references non-existent component |
```

Audit is read-only, reports findings for user decision.

### `/c3 apply`

Generates/updates CLAUDE.md files:

```
For each component's Code References:
  → Create CLAUDE.md if missing
  → Replace c3-generated block if stale
  → Remove orphaned blocks (optional, with flag?)

Writes files to repo.
```

### Workflow

```
┌─────────────────────────────────────────────────────────┐
│  /c3 audit                                               │
│  ─────────────────────────────────────────────────────   │
│  Phases 1-9: (existing checks)                           │
│  Phase 10: Context Files                                 │
│    → Reports missing CLAUDE.md                           │
│    → Reports stale c3-generated blocks                   │
│    → Reports orphaned blocks                             │
│  Read-only, reports findings                             │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼ User reviews findings
┌─────────────────────────────────────────────────────────┐
│  /c3 apply                                               │
│  ─────────────────────────────────────────────────────   │
│  For each component's Code References:                   │
│    → Create CLAUDE.md if missing                         │
│    → Replace c3-generated block if stale                 │
│    → Remove orphaned blocks                              │
│  Writes files, user commits to repo                      │
└─────────────────────────────────────────────────────────┘
```

## Enforcement Model

C3 is the source of truth. The audit mechanism works regardless of who/what wrote the code:

1. **Audit runs anytime** - On demand, ADR transition, CI, etc.
2. **Reports alignment issues** - Code vs C3 docs, CLAUDE.md freshness
3. **User decides action** - Fix code OR update C3 docs
4. **No blocking** - Audit is informational, user retains control

ADR `provisioned → implemented` transition is just one occasion where this check runs.

## Implementation Tasks

1. **Update `references/audit-checks.md`** - Add Phase 10 specification
2. **Update `skills/c3/SKILL.md`** - Document `/c3 apply` command
3. **Create apply logic** - Script or agent to generate CLAUDE.md files
4. **Update audit agent** - Implement Phase 10 checks
5. **Test with real C3 project** - Verify CLAUDE.md generation and audit detection

## Open Questions

- Should `/c3 apply` auto-commit or leave files staged for user to commit?
- Should orphaned blocks be auto-removed or just reported?
- How to handle components with many Code References (large CLAUDE.md)?
