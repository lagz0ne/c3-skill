# C3 Config Design

## Summary

Add `.c3/settings.yaml` for project-specific preferences, configured via `c3-config` skill.

## Problem

Teams have different preferences for:
- Diagram tools and styles
- Documentation tone per layer
- Team-specific guardrails
- Post-ADR handoff steps

Currently no way to capture these; skills use hardcoded defaults.

## Solution

### File: `.c3/settings.yaml`

```yaml
diagrams: |
  mermaid
  sequence: API interactions, request flows
  flowchart: decision logic, error handling

context: |
  system boundaries, actors, external integrations
  avoid implementation details

container: |
  service responsibilities, API contracts
  backend: endpoints, data flows
  frontend: component hierarchy, state

component: |
  technical specifics, configs, algorithms

guard: |
  discovered incrementally via c3-config

handoff: |
  after ADR accepted:
  1. create implementation tasks
  2. notify team

audit: |
  handoff: tasks
  # options: manual, tasks, agents
```

**Format:** YAML with text values (flexible prose, not rigid schema)

**Sections:**
- `diagrams` - tool and usage patterns
- `context` - context-layer guidance
- `container` - container-layer guidance
- `component` - component-layer guidance
- `guard` - team guardrails (discovered via Socratic questioning)
- `handoff` - post-ADR steps
- `audit` - audit findings handoff preference (manual/tasks/agents)

### Skill: `c3-config`

**Purpose:** Create and refine settings.yaml

**Flow:**
1. Check if `.c3/settings.yaml` exists
2. If missing → create with sensible defaults
3. Socratic questions to refine sections
4. Write updates incrementally

**Invocation:**
- Standalone: user runs `c3-config` directly
- Via c3-adopt: called if settings.yaml missing
- Via c3-init: called as part of initialization

**Onboarding approach:**
- Provide sensible defaults with narrative examples
- Socratic questions focus on empty/minimal sections
- User can edit file directly anytime

### Integration

**c3-adopt changes:**
- After creating `.c3/` structure, check for settings.yaml
- If missing → call c3-config

**c3-init changes:**
- Include c3-config in initialization flow

**c3-design changes:**
- Read `.c3/settings.yaml` at skill start (if exists)
- Apply guidance throughout session:
  - Diagram preferences when creating visuals
  - Layer guidance when writing docs
  - Guardrails as constraints
  - Handoff steps after ADR completion

## Files Changed

| File | Change |
|------|--------|
| `skills/c3-config/SKILL.md` | New skill |
| `skills/c3-adopt/SKILL.md` | Call c3-config if settings missing |
| `skills/c3-design/SKILL.md` | Read settings at start |
| `commands/c3-init.md` | Include c3-config in flow |
| `references/v3-structure.md` | Document settings.yaml |

## Migration

Existing projects without settings.yaml:
- Run `c3-config` directly, or
- Run `c3-adopt` which triggers c3-config

No version field needed - c3-version lives in README.md.
