---
name: c3-component-design
description: Explore Component level impact during scoping - implementation details, configuration, dependencies, and technical specifics
---

# C3 Component Level Exploration

## Overview

Explore Component-level impact during the scoping phase of c3-design. Component is the implementation layer: detailed specifications, configuration, and technical behavior.

**Abstraction Level:** Implementation details. Code examples, configuration snippets, and library usage are appropriate here.

**Announce at start:** "I'm using the c3-component-design skill to explore Component-level impact."

**Reading order:** Navigate Context → Container → Component.

**Reference direction:** Components are terminal - no further derivation. They receive links FROM Container docs but do not link upward or further down.

## When Invoked

Called during EXPLORE phase of c3-design when:
- Hypothesis suggests Component-level impact
- Need to understand implementation implications
- Exploring downstream from Container
- Change affects specific technical behavior

Also called by c3-adopt to CREATE initial Component documentation.

---

## What Belongs at Component Level

### Inclusion Criteria

**INCLUDE at Component level:**

| Element | Why Component | Example |
|---------|--------------|---------|
| Implementation details | HOW it works | Connection pooling algorithm |
| Libraries & versions | Specific dependencies | `pg: 8.11.x` |
| Configuration values | Actual env vars | `DB_POOL_MAX=50` |
| Code examples | Usage patterns | TypeScript snippets |
| Error handling | Specific strategies | Retry with backoff |
| Interfaces/Types | Data structures | `interface Task { ... }` |
| Algorithms | Logic specifics | Token validation steps |
| Performance tuning | Specific optimizations | Pool sizing formula |
| Health checks | Implementation | `SELECT 1` ping |
| Testing approach | How to test | Mock strategies |

**EXCLUDE from Component (push to Container or Context):**

| Element | Why Not Component | Where It Belongs |
|---------|------------------|------------------|
| Container purpose | Too high level | Container |
| API endpoint list | Container scope | Container |
| Middleware order | Container pipeline | Container |
| Technology choice rationale | Container decision | Container |
| System protocols | System-wide | Context |
| Cross-cutting concerns | Span containers | Context |
| Deployment topology | System-wide | Context |

### Litmus Test

Ask: "Could a developer implement this from the documentation?"
- **Yes** → Correct level (Component)
- **No, needs more detail** → Add implementation specifics
- **No, it's about structure** → Push up to Container

---

## Component Nature Types (Open-Ended)

Nature type determines documentation focus. **This is NOT a fixed taxonomy** - use whatever nature type helps describe and organize the component's documentation.

| Nature Type | Documentation Focus |
|-------------|---------------------|
| **Resource/Integration** | Configuration, env differences, how/why config loaded |
| **Business Logic** | Domain flows, rules, edge cases, the "messy" heart |
| **Framework/Entrypoint** | Mixed concerns - auth, errors, signals, protocol handoff, lifecycle |
| **Cross-cutting** | Integration patterns, how it's used everywhere, conventions |
| **Build/Deployment** | Build pipeline, deploy config, CI/CD specifics |
| **Testing** | Test strategies, fixtures, mocking approaches |
| **Contextual** | Situation-specific behavior (caching, websocket, etc.) |
| ... | Whatever the component needs |

**Key principle:** The nature type guides what sections to emphasize, not what the component is allowed to be.

---

## Component Contains What Container Doesn't Enforce

| Element | Description |
|---------|-------------|
| Stack details | Which library, why chosen, exact configuration |
| Environment config | Env vars, defaults, dev vs prod differences |
| Implementation patterns | Conventions, algorithms, code patterns |
| Interfaces/Types | Method signatures, data structures, DTOs |
| Error handling | Specific error codes, retry strategies |
| Usage examples | Code snippets showing how to use it |

**Stack and config live here:** Library selections, versions, and exact configuration belong to the Component. Containers only name the technology stack; Components document the concrete choices and environment differences.

---

## Expressing Relationships at Component Level

### Relationship Types

| Relationship | Expression | Example |
|--------------|------------|---------|
| Dependency injection | Constructor param | `constructor(private db: Pool)` |
| Method call | Interface method | `taskService.create(data)` |
| Event emission | Event name + payload | `emit('task.created', { task })` |
| Data flow | Input → Transform → Output | `request → validate → persist` |
| Error propagation | Throws/catches | `throws DatabaseError` |

### Interface Documentation

```markdown
## Service Interface

```typescript
interface TaskService {
  create(userId: string, data: CreateInput): Promise<Task>;
  update(taskId: string, data: UpdateInput): Promise<Task>;
  delete(taskId: string): Promise<void>;
}
```
```

### Dependency Graph

```markdown
## Dependencies

| This Component | Uses | For |
|----------------|------|-----|
| TaskService | DBPool | Database queries |
| TaskService | Validator | Input validation |
| TaskService | EventEmitter | Domain events |
```

### Event Contracts

```markdown
## Events Emitted

| Event | Payload | When |
|-------|---------|------|
| `task.created` | `{ task, userId }` | After successful create |
| `task.deleted` | `{ taskId, userId }` | After successful delete |
```

### DO NOT Express at Component

- Container-to-container communication (Container level)
- Actor interactions (Context level)
- System-wide patterns (Context level)

---

## Diagrams for Component Level

Use the diagram that best explains how the component works:
- **Flowchart** - decision logic or processing steps
- **Sequence diagram** - calls between functions/components during a request
- **ERD (detailed)** - data relationships owned by this component
- **State chart** - lifecycle and state transitions

Avoid repeating Container-level diagrams; focus on implementation specifics.

---

## Component Level Defines

| Concern | Examples |
|---------|----------|
| **Component identity** | Name, purpose within container |
| **Technical implementation** | Libraries, patterns, algorithms |
| **Configuration** | Environment variables, config files |
| **Dependencies** | External libraries, other components |
| **Interfaces** | Methods, data structures, APIs |
| **Error handling** | Failures, retry logic, fallbacks |
| **Performance** | Caching, optimization, resources |
| **Health checks** | Monitoring, observability |

## Exploration Questions

When exploring Component level, investigate:

### Isolated (at Component)
- What implementation details change?
- What configuration affected?
- What error handling needs updating?

### Upstream (to Container)
- Does this change container responsibilities?
- Does middleware need modification?
- Are container APIs affected?

### Adjacent (same level)
- What sibling components related?
- What shared utilities affected?
- What component interactions change?

### Downstream (consumers)
- What code uses this component?
- What tests need updating?
- What documentation affected?

## Socratic Questions for Component Discovery

When creating or validating Component documentation, ask:

### Purpose
1. "What specific problem does this component solve?"
2. "What would you have to write inline if this didn't exist?"
3. "What's the single most important thing this does?"

### Implementation
4. "What library/framework does this use?"
5. "Why this library over alternatives?"
6. "What's the core algorithm or pattern?"

### Interface
7. "What are the public methods/functions?"
8. "What are the input types and output types?"
9. "What events does this emit or listen to?"

### Configuration
10. "What environment variables does this need?"
11. "What are sensible defaults?"
12. "What must change between dev and prod?"

### Error Handling
13. "What can go wrong?"
14. "What errors are retriable vs fatal?"
15. "How are errors communicated to callers?"

### Testing
16. "How would you test this in isolation?"
17. "What would you mock?"
18. "What edge cases matter?"

## Reading Component Documents

Use c3-locate to retrieve:

```
c3-locate COM-001                    # Overview
c3-locate #com-001-stack             # Library and version
c3-locate #com-001-interfaces        # Methods and types
c3-locate #com-001-config            # Env vars and defaults
c3-locate #com-001-behavior          # Diagrams/behavior
c3-locate #com-001-errors            # Error strategies
c3-locate #com-001-usage             # Usage examples
```

## Impact Signals

| Signal | Meaning |
|--------|---------|
| Interface change | Consumers need updating |
| Configuration change | Deployment affected |
| Dependency change | Security/compatibility review |
| Error handling change | Monitoring may need updates |
| Component should be higher level | Revisit hypothesis at Container |

## Output for c3-design

After exploring Component level, report:
- What Component-level elements are affected
- Impact on sibling components
- Whether Container level needs changes
- Whether this is truly Component-level or should be higher
- Whether hypothesis needs revision

## Abstraction Check

**Critical question:** Does this change belong at Component level?

Signs it should be **higher** (Container/Context):
- Affects multiple components similarly
- Changes middleware behavior
- Alters container responsibilities
- Impacts system protocols

Signs it's correctly at **Component**:
- Isolated to single component
- Implementation detail only
- No upstream contract changes
- Configuration/behavior tweak

If change belongs higher, report this to c3-design for hypothesis revision.

## Document Template Reference

Component documents follow this structure. **Adapt sections based on Nature Type.**

```markdown
---
id: COM-NNN-slug
title: [Component Name] ([Nature Type])
summary: >
  [Why read this document - what it covers]
---

# [COM-NNN-slug] [Component Name] ([Nature Type])

::: info Container
Belongs to [CON-XXX: Container Name](../../containers/CON-XXX-slug.md)
:::

## Overview {#com-nnn-overview}
[Brief description of what this component does]

## Stack {#com-nnn-stack}
- Library: `[e.g., pg 8.11.x]`
- Why: [Reason for choice - stability, features, etc.]

## Interfaces & Types {#com-nnn-interfaces}
```typescript
interface TaskService {
  create(userId: string, data: CreateInput): Promise<Task>;
  update(taskId: string, data: UpdateInput): Promise<Task>;
  delete(taskId: string): Promise<void>;
}
```

## Configuration {#com-nnn-config}
| Env Var | Dev | Prod | Why |
|---------|-----|------|-----|
| DB_POOL_MIN | 2 | 10 | Baseline connections |
| DB_POOL_MAX | 10 | 50 | Scale with load |
| DB_IDLE_TIMEOUT | 30s | 10s | Release faster in prod |

## Behavior {#com-nnn-behavior}
<!--
State diagram, flowchart, or prose explaining key behavior.
Choose diagram type based on what needs explaining (flowchart, sequence,
ERD, state chart).
-->

## Error Handling {#com-nnn-errors}
| Error | Retriable | Action |
|-------|-----------|--------|
| Connection refused | Yes | Retry with backoff |
| Pool exhausted | Yes | Wait up to 5s |
| Query timeout | No | Propagate to caller |

## Usage {#com-nnn-usage}
```typescript
const pool = createPool(config);
const result = await pool.query('SELECT * FROM users WHERE id = $1', [userId]);
```

## Related {#com-nnn-related}
```

**Diagram choice is contextual:** Use flowchart for decision logic, sequence for interactions, state chart for lifecycle, ERD for data relationships.

Use these heading IDs for precise exploration.
