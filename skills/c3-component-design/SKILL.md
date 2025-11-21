---
name: c3-component-design
description: Explore Component level impact during scoping - implementation details, configuration, dependencies, and technical specifics
---

# C3 Component Level Exploration

## Overview

Explore Component-level impact during the scoping phase of c3-design. Component is the implementation layer: detailed specifications, configuration, and technical behavior.

**Abstraction Level:** Implementation details. Code examples, configuration snippets, and library usage are appropriate here.

**Announce at start:** "I'm using the c3-component-design skill to explore Component-level impact."

## When Invoked

Called during EXPLORE phase of c3-design when:
- Hypothesis suggests Component-level impact
- Need to understand implementation implications
- Exploring downstream from Container
- Change affects specific technical behavior

Also called by c3-adopt to CREATE initial Component documentation.

---

## Component Nature Types (Open-Ended)

Nature type determines documentation focus. **Not a fixed taxonomy** - use whatever helps code quality.

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

---

## What Belongs at Component Level

Component contains **what Container doesn't enforce** - the implementation details.

### Inclusion Criteria

**INCLUDE at Component level:**

| Element | Why Component | Example |
|---------|--------------|---------|
| Stack details | Which library, why chosen, exact configuration | `pg: 8.11.x` - why over alternatives |
| Environment config | Env vars, defaults, dev vs prod differences | `DB_POOL_MAX=50` (dev: 10, prod: 50) |
| Implementation patterns | Conventions, algorithms, code patterns | Connection pooling algorithm |
| Interfaces/Types | Method signatures, data structures, DTOs | `interface Task { ... }` |
| Error handling | Specific error codes, retry strategies | Retry with backoff, error catalog |
| Usage examples | Code snippets showing how to use it | TypeScript snippets |
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

Use common UML techniques **where they help explain**. Choice is contextual based on what needs explaining.

### Flowchart - Decision Logic

**When to use:** Algorithm logic, processing steps, decision trees.

```mermaid
flowchart TD
    A[Receive Token] --> B{Token in header?}
    B -->|Yes| C[Extract from header]
    B -->|No| D{Token in cookie?}
    D -->|Yes| E[Extract from cookie]
    D -->|No| F[Return 401]
    C --> G[Validate signature]
    E --> G
    G --> H{Valid?}
    H -->|Yes| I[Inject user context]
    H -->|No| F
    I --> J[Continue]
```

### Sequence Diagram - Component Interactions

**When to use:** Method calls, request flow, component collaboration.

```mermaid
sequenceDiagram
    participant R as Route Handler
    participant S as TaskService
    participant V as Validator
    participant D as Repository
    participant E as EventEmitter

    R->>S: createTask(userId, data)
    S->>V: validate(data)
    V-->>S: validatedData
    S->>D: insert(task)
    D-->>S: taskId
    S->>E: emit('task.created')
    S-->>R: Task
```

### State Chart - Lifecycle/Transitions

**When to use:** Component lifecycle, connection states, state machines.

```mermaid
stateDiagram-v2
    [*] --> Idle
    Idle --> Acquiring: getConnection()
    Acquiring --> Active: success
    Acquiring --> Waiting: pool exhausted
    Waiting --> Acquiring: connection released
    Waiting --> Error: timeout
    Active --> Idle: release()
    Error --> [*]
```

### ERD - Data Relationships

**When to use:** Specific table columns, data structures, relationships.

### Class Diagram - Type Relationships

**When to use:** Complex type hierarchies, interface contracts.

### Diagram Selection Guide

| What to Explain | Diagram Type |
|-----------------|--------------|
| Decision logic, algorithms | Flowchart |
| Method calls, request flow | Sequence diagram |
| Lifecycle, state transitions | State chart |
| Data structures, tables | ERD |
| Type relationships | Class diagram |

### Avoid at Component Level

| Diagram Type | Why Not | Where It Belongs |
|--------------|---------|------------------|
| System context | Too high level | Context |
| Container overview | Too high level | Container |
| Deployment diagrams | Infrastructure | Context/Container |

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
c3-locate #com-001-implementation    # Technical details
c3-locate #com-001-configuration     # Config options
c3-locate #com-001-pool-behavior     # Specific behavior
c3-locate #com-001-error-handling    # Error strategies
c3-locate #com-001-performance       # Performance characteristics
c3-locate #com-001-health-checks     # Health monitoring
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

Component documents focus on **what Container doesn't enforce**. Structure varies by nature type.

### Example: Resource Nature (DB Pool)

```markdown
---
id: COM-NNN-slug
title: [Component Name] Component
summary: >
  [Why read this document - what it covers]
nature: Resource
---

# [COM-NNN-slug] [Component Name] (Resource Nature)

## Overview
Connection pooling for PostgreSQL

## Stack {#com-nnn-stack}
- Library: `pg` 8.11.x
- Why: Native driver, proven stability, supports LISTEN/NOTIFY

## Configuration {#com-nnn-config}
| Env Var | Dev | Prod | Why |
|---------|-----|------|-----|
| DB_POOL_MIN | 2 | 10 | Baseline connections |
| DB_POOL_MAX | 10 | 50 | Scale with load |
| DB_IDLE_TIMEOUT | 30s | 10s | Release faster in prod |

## Behavior {#com-nnn-behavior}

```mermaid
stateDiagram-v2
    [*] --> Idle
    Idle --> Acquiring: getConnection()
    Acquiring --> Active: success
    Acquiring --> Waiting: pool exhausted
    Waiting --> Acquiring: connection released
    Waiting --> Error: timeout
    Active --> Idle: release()
    Error --> [*]
```

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
```

### Example: Business Logic Nature

```markdown
---
id: COM-NNN-slug
title: Order Flow Component
nature: Business Logic
---

# [COM-NNN-slug] Order Flow (Business Logic Nature)

## Overview
Core order processing - the "messy" heart of order handling.

## Domain Rules {#com-nnn-rules}
| Rule | Condition | Action |
|------|-----------|--------|
| Stock check | Before confirm | Validate inventory |
| Price lock | On add to cart | Lock price for 15min |
| Tax calculation | At checkout | Apply regional rules |

## Edge Cases {#com-nnn-edge-cases}
- Partial fulfillment when stock insufficient
- Price change during checkout flow
- Concurrent modifications to same order

## Flow {#com-nnn-flow}
[Flowchart or sequence diagram as appropriate]

## Dependencies {#com-nnn-deps}
- Upstream: [components this depends on]
- Downstream: [components that depend on this]
- Infra features consumed: [CON-xxx#features]
```

### Component Checklist (must be true to call COM done)

- [ ] Nature chosen and documented (resource, business, cross-cutting, entrypoint, etc.)
- [ ] Stack and configuration fully documented (with dev/prod differences)
- [ ] Interfaces/types specified
- [ ] Behavior explained with at least one diagram if non-trivial
- [ ] Error handling table present
- [ ] Usage example shows intended invocation
- [ ] Dependencies section lists upstream/downstream components and consumed infra features
- [ ] All anchors use `{#com-xxx-*}` format for stable linking

### Key Principle

Component documentation is **terminal** - no further derivation. Include everything needed to understand and implement.

Use these heading IDs for precise exploration.
