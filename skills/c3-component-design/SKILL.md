---
name: c3-component-design
description: Explore Component level impact during scoping - implementation details, configuration, dependencies, and technical specifics
---

# C3 Component Level Exploration

## Overview

Explore Component-level impact during the scoping phase of c3-design. Component is the implementation layer: detailed specifications, configuration, and technical behavior.

**Abstraction Level:** Implementation details. Code examples, configuration snippets, and library usage are appropriate here.

**Announce at start:** "I'm using the c3-component-design skill to explore Component-level impact."

## Configuration Loading

**At skill start:**

1. Read `defaults.md` from this skill directory
2. Read `.c3/settings.yaml` (if exists)
3. Check `component` section in settings:
   - If `useDefaults: true` (or missing) → merge defaults + user customizations
   - If `useDefaults: false` → use only user-provided config
4. Display merged configuration

**Merge rules:**
- `include`: defaults + user additions (union)
- `exclude`: defaults + user additions (union)
- `litmus`: user overrides default (replacement)
- `diagrams`: user overrides default (replacement)

**Display at start:**
```
Layer configuration (Component):
- Include: [merged list]
- Exclude: [merged list]
- Litmus: [active litmus test]
- Diagrams: [active diagram types]
```

## When Invoked

Called during EXPLORE phase of c3-design when:
- Hypothesis suggests Component-level impact
- Need to understand implementation implications
- Change affects specific technical behavior

Also called by c3-adopt to CREATE initial Component documentation.

## Quick Reference

| Direction | Question | Action |
|-----------|----------|--------|
| **Isolated** | What implementation details change? | Investigate this component |
| **Upstream** | Does this change container responsibilities? | May need c3-container-design |
| **Adjacent** | What sibling components related? | Check dependencies |
| **Downstream** | What code uses this component? | Update consumers |

## Component Nature Types

Nature type determines documentation focus:

| Nature | Focus |
|--------|-------|
| **Resource/Integration** | Configuration, env differences, connection handling |
| **Business Logic** | Domain flows, rules, edge cases |
| **Framework/Entrypoint** | Auth, errors, lifecycle, protocol handoff |
| **Cross-cutting** | Integration patterns, conventions |
| **Build/Deployment** | CI/CD, build config |
| **Testing** | Test strategies, fixtures |

## Role Mapping

Map discovered components to roles from @references/role-taxonomy.md:

| Nature Type | Typical Roles |
|-------------|---------------|
| Resource/Integration | Database Access, Cache Access, External Client |
| Business Logic | Business Logic, Saga Coordinator |
| Framework/Entrypoint | Request Handler, Event Consumer |
| Cross-cutting | Health/Readiness, Observability |
| Testing | (testing infrastructure, not a role) |

Use roles as vocabulary: "This resource component provides Database Access - correct?"

---

## What Belongs at Component Level

See `defaults.md` for canonical include/exclude lists.

Check `.c3/settings.yaml` for project-specific overrides under the `component` section.

Apply the active litmus test when deciding content placement.

---

## Expressing Relationships

| Relationship | Expression |
|--------------|------------|
| Dependency injection | `constructor(private db: Pool)` |
| Method call | `taskService.create(data)` |
| Event emission | `emit('task.created', { task })` |
| Data flow | `request → validate → persist` |

### Dependencies Table

```markdown
| This Component | Uses | For |
|----------------|------|-----|
| TaskService | DBPool | Database queries |
| TaskService | Validator | Input validation |
```

---

## Diagrams

See `defaults.md` for default diagram recommendations.

Check `.c3/settings.yaml` for project-specific diagram preferences under `component.diagrams`.

Use the project's `diagrams` setting (root level) for tool preference (mermaid, PlantUML, etc.).

---

## Discovery Approach

### Step 1: Understand Flow

Look in all directions:
- **Upward:** What calls this component?
- **Downward:** What does this component call?
- **Adjacent:** What runs alongside this?

### Step 2: Task Explore (if needed)

For complex components, use Task Explore:
```
Explore this component at [path]:
- What does it depend on? (imports, injections)
- What depends on it? (who imports this)
- What's the main pattern? (service, repository, handler)
- How are errors handled?
```

### Step 3: Socratic Questions

Use @references/discovery-questions.md for component-level questions.
Use AskUserQuestion when choices are clear.

### Exploration Questions

| Direction | Questions |
|-----------|-----------|
| **Isolated** | What implementation details change? What configuration affected? |
| **Upstream** | Does this change container responsibilities? Are APIs affected? |
| **Adjacent** | What sibling components related? What shared utilities affected? |
| **Downstream** | What code uses this component? What tests need updating? |

---

## Socratic Questions

See [socratic-method.md](../../references/socratic-method.md) for techniques.

**Purpose:**
- "What specific problem does this component solve?"
- "What's the single most important thing this does?"

**Implementation:**
- "What library/framework does this use? Why?"
- "What's the core algorithm or pattern?"

**Configuration:**
- "What environment variables does this need?"
- "What must change between dev and prod?"

**Error Handling:**
- "What can go wrong?"
- "What errors are retriable vs fatal?"

---

## Abstraction Check

**Signs it should be HIGHER (Container/Context):**
- Affects multiple components similarly
- Changes middleware behavior
- Alters container responsibilities

**Signs it's correctly at Component:**
- Isolated to single component
- Implementation detail only
- No upstream contract changes

If change belongs higher, report to c3-design for hypothesis revision.

---

## Testing Discovery

Reference @references/testing-discovery.md for testing documentation.

Ask: "Are there tests for this component?"

If yes, document location and how to run.
If no, document as TBD - don't enforce.

```markdown
## Testing {#c3-nnn-testing}
- Location: `src/__tests__/componentName.test.ts`
- Run: `npm test componentName`
```

---

## Document Template

```markdown
---
id: c3-{N}{NN}
c3-version: 3
title: [Component Name]
nature: [Resource/Business/etc]
---

# [Component Name]

## Overview {#c3-nnn-overview}
[Purpose and responsibility]

## Stack {#c3-nnn-stack}
- Library: [name] [version]
- Why: [rationale]

## Configuration {#c3-nnn-config}
| Env Var | Dev | Prod | Why |
|---------|-----|------|-----|

## Behavior {#c3-nnn-behavior}
[Diagram if non-trivial]

## Error Handling {#c3-nnn-errors}
| Error | Retriable | Action |
|-------|-----------|--------|

## Usage {#c3-nnn-usage}
[Code example]

## Dependencies {#c3-nnn-deps}
- Upstream: [components this depends on]
- Downstream: [components that depend on this]
```

### Checklist

- [ ] **Scoping verified** (up/down/aside per derivation-guardrails.md)
- [ ] Nature chosen and documented
- [ ] Stack and configuration documented
- [ ] Interfaces/types specified
- [ ] Behavior explained (diagram if complex)
- [ ] Error handling table present
- [ ] Usage example included
- [ ] Dependencies listed
- [ ] Anchors use `{#c3-nnn-*}` format

## Output for c3-design

After exploring, report:
- Component-level elements affected
- Impact on sibling components
- Whether Container level needs changes
- Whether hypothesis needs revision

## Related

- [derivation-guardrails.md](../../references/derivation-guardrails.md)
- [v3-structure.md](../../references/v3-structure.md)
