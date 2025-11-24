# C3 Audit Design

## Summary

Add `/c3-audit` command to verify `.c3/` documentation against reality.

Opposite of `c3-adopt`: adopt creates docs from code, audit verifies docs match code.

## Problem

Documentation drifts from implementation:
- Containers added/removed without updating docs
- ADRs started but never completed
- Settings defined but not followed
- Components documented but code changed

No way to detect these gaps systematically.

## Solution

### Command: `/c3-audit`

Slash command that performs full reconciliation between `.c3/` and codebase.

### Audit Categories

| Category | What's Checked |
|----------|----------------|
| **Config compliance** | `.c3/` follows `settings.yaml` preferences |
| **Vision vs fact** | Documented containers/components exist in code, undocumented code detected |
| **ADR status** | Proposed ADRs stale? Accepted ADRs implemented? Abandoned? |
| **Skill compliance** | Docs follow skill templates (frontmatter, sections, anchors) |

### Audit Flow

```
/c3-audit
  1. Read .c3/ structure (TOC, all docs)
  2. Read settings.yaml
  3. Task Explore codebase (very thorough)
  4. Compare:
     - Documented containers vs actual directories
     - Documented components vs actual code
     - settings.yaml preferences vs doc content
     - ADR statuses vs implementation state
     - Doc structure vs skill templates
  5. Generate findings report
  6. Ask user: how to handle findings?
  7. Execute chosen handoff
```

### Findings Report Format

```markdown
## Audit Findings

### Config Compliance
- ⚠️ Diagrams use PlantUML but settings.yaml specifies mermaid
- ✅ Tone matches settings

### Vision vs Fact
- ❌ c3-3-worker documented but /worker directory not found
- ⚠️ /services/notification exists but not documented
- ✅ c3-1-backend matches /api

### ADR Status
- ⚠️ ADR-001 proposed 30+ days ago, still not accepted
- ❌ ADR-002 accepted but verification items incomplete
- ✅ ADR-003 accepted and implemented

### Skill Compliance
- ❌ c3-2-frontend missing frontmatter id
- ⚠️ c3-101-db-pool missing {#c3-101-*} anchors
- ✅ Context doc follows template

### Summary
- 2 critical (❌)
- 4 warnings (⚠️)
- 4 passing (✅)
```

### Handoff Options

Check `settings.yaml` for `audit:` preference first. If not set, ask:

```
How would you like to handle these findings?

1. Manual - I'll review and fix myself
2. Tasks - Create vibe-kanban tasks for each finding
3. Agents - Dispatch subagents to fix in parallel

(Save preference to settings.yaml for next time?)
```

**Manual:** Just show report, done.

**Tasks:** Create tasks via vibe-kanban MCP:
- One task per finding category or per finding (user choice)
- Task description includes finding details and suggested fix

**Agents:** Use Task tool with subagent_type:
- Group independent findings
- Dispatch parallel agents to fix
- Report results back

**Settings storage:**
```yaml
audit: |
  handoff: tasks
  # or: manual, agents
```

### Deep Reconciliation (Vision vs Fact)

Use Task Explore to:
1. Discover all deployable containers in codebase
2. Compare against `.c3/c3-{N}-*/` folders
3. For each container, discover components
4. Compare against documented `c3-{N}{NN}-*.md` files
5. Flag: documented but missing, exists but undocumented

### Config Compliance Checks

| Setting | Check |
|---------|-------|
| `diagrams:` | Scan docs for diagram blocks, verify tool matches |
| `context:` | Context doc follows guidance |
| `container:` | Container docs follow guidance |
| `component:` | Component docs follow guidance |
| `guard:` | Docs don't violate stated guardrails |
| `handoff:` | ADRs include handoff steps |

### Skill Compliance Checks

| Check | How |
|-------|-----|
| Frontmatter | `id`, `c3-version`, `title` present |
| Anchors | Headings use `{#id-*}` format |
| Sections | Required sections per layer skill |
| Links | Downward only (no upward links) |

## Files

| File | Type |
|------|------|
| `commands/c3-audit.md` | New command |

## Integration

- Uses `settings.yaml` (from c3-config)
- Uses Task Explore for codebase discovery
- Uses vibe-kanban MCP for task creation
- Uses Task tool for agent dispatch
