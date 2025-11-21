# C3 Derivation Model Design

> Design for the hierarchical derivation relationship between Context, Container, and Component layers.

## Problem

Current C3 skills describe each layer independently but lack clarity on:
- How higher layers constrain lower layers
- How to navigate between layers
- What each layer should contain vs delegate down
- How relationships should be documented and traced

## Core Insight

**Higher layers derive lower layers.** When you define something at Context level (like a protocol between Frontend and Backend), that becomes a constraint that MUST be addressed when describing the lower levels.

## Design

### Reading Order

Always top-down: **Context → Container → Component**

### Reference Direction

References only flow DOWN - higher layer links to lower layer implementations.

- Context defines protocols → links to Container#sections implementing them
- Container defines components → links to Component docs
- **No upward links needed** - reader already came from above

This creates:
- Single source of truth (no duplicate relationship definitions)
- Natural reading flow follows derivation
- Maintenance only at one location

---

## Context Level

**Abstraction:** Bird's eye view - WHAT exists and HOW they relate

### Contains

| Element | Description | Links Down To |
|---------|-------------|---------------|
| System boundary | What's inside vs outside | - |
| Actors | Users, external systems, third parties | - |
| Container inventory | WHAT containers exist (not how) | Container docs |
| Protocols | How containers communicate | Container#sections implementing each side |
| System cross-cutting | Auth strategy, logging approach, error handling patterns | Container#sections implementing them |
| Deployment topology | High-level infrastructure | - |

### Diagrams

- System context diagram (actors, boundaries)
- Container overview diagram (what exists, protocols between)
- Deployment topology (high-level)

### Example Structure

```markdown
# CTX-001 System Overview

## Containers {#ctx-001-containers}
| Container | Type | Description |
|-----------|------|-------------|
| [CON-001-backend](./containers/CON-001-backend.md) | Code | REST API |
| [CON-002-frontend](./containers/CON-002-frontend.md) | Code | Web UI |
| [CON-003-postgres](./containers/CON-003-postgres.md) | Infrastructure | Data store |

## Protocols {#ctx-001-protocols}
| From | To | Protocol | Implementations |
|------|-----|----------|-----------------|
| Frontend | Backend | REST/HTTPS | [CON-002#api-calls], [CON-001#rest-endpoints] |
| Backend | Postgres | SQL | [CON-001#db-access], [CON-003#config] |

## Cross-Cutting {#ctx-001-cross-cutting}
### Authentication
JWT-based, implemented in: [CON-001#auth-middleware], [CON-002#auth-handling]

### Logging
Structured JSON with correlation IDs, implemented in: [CON-001#logging], [CON-002#logging]
```

---

## Container Level

**Abstraction:** WHAT it does and WITH WHAT, not implementation details

### Two Container Types

| Type | Has Components? | Documentation Focus |
|------|-----------------|---------------------|
| **Code Container** | Yes | Tech stack, component inventory, how protocols are implemented |
| **Infrastructure Container** | No (leaf node) | Engine, config, features provided to code containers |

### Code Container Contains

| Element | Description | Links Down To |
|---------|-------------|---------------|
| Technology stack | Language, framework, runtime | - |
| Component inventory | WHAT components exist | Component docs |
| Component relationships | How components connect | Flowchart diagram |
| Data flow | How data moves through container | Sequence diagram |
| Container cross-cutting | Logging, error handling, validation within this container | Component docs |
| Protocol implementations | How this container implements Context protocols | Component#sections |

### Infrastructure Container Contains

| Element | Description | Consumed By |
|---------|-------------|-------------|
| Engine/technology | PostgreSQL 15, NATS, Redis | - |
| Configuration | Settings, tuning | Code container components |
| Features provided | WAL, pub/sub, streams | Code container components reference these |

**Infrastructure containers are LEAF NODES** - no component level beneath them. Their features become inputs to code container components.

### Diagrams

- **Flowchart** - component relationships/connections
- **Sequence diagram** - data flow through container

### Example Structure (Code Container)

```markdown
# CON-001 Backend Container (Code)

## Technology Stack {#con-001-stack}
- Runtime: Node.js 20
- Framework: Express 4.18
- Language: TypeScript 5.x

## Component Relationships {#con-001-relationships}

​```mermaid
flowchart LR
    Entry[REST Routes] --> Auth[Auth Middleware]
    Auth --> Business[Order Flow]
    Business --> DB[DB Pool]
    DB --> External[Postgres]

    Auth -.-> Log[Logger]
    Business -.-> Log
    DB -.-> Log
​```

## Data Flow {#con-001-data-flow}

​```mermaid
sequenceDiagram
    participant Client
    participant Routes
    participant Auth
    participant OrderFlow
    participant DBPool

    Client->>Routes: POST /orders
    Routes->>Auth: validate token
    Auth-->>Routes: user context
    Routes->>OrderFlow: createOrder(user, data)
    OrderFlow->>DBPool: insert
    DBPool-->>OrderFlow: order
    OrderFlow-->>Routes: result
    Routes-->>Client: 201 Created
​```

## Container Cross-Cutting {#con-001-cross-cutting}

### Logging {#con-001-logging}
- Structured JSON, correlation IDs passed through
- Implemented by: [COM-006-logger]

### Error Handling {#con-001-error-handling}
- Unified error format, error codes catalog
- Implemented by: [COM-007-error-handler]

## Components {#con-001-components}
| Component | Nature | Responsibility |
|-----------|--------|----------------|
| [COM-001-rest-routes] | Entrypoint | HTTP handling |
| [COM-002-auth-middleware] | Cross-cutting | Token validation |
| [COM-003-db-pool] | Resource | Connection management |
| [COM-004-order-flow] | Business | Order processing |
```

### Example Structure (Infrastructure Container)

```markdown
# CON-003 Postgres Container (Infrastructure)

## Engine {#con-003-engine}
PostgreSQL 15

## Configuration {#con-003-config}
| Setting | Value | Why |
|---------|-------|-----|
| max_connections | 100 | Support pooling from backend |
| wal_level | logical | Enable event streaming |

## Features Provided {#con-003-features}
| Feature | Used By |
|---------|---------|
| WAL logical replication | [CON-001-backend] → [COM-005-event-streaming] |
| LISTEN/NOTIFY | [CON-001-backend] → [COM-003-db-pool] |
```

---

## Component Level

**Abstraction:** Implementation details - HOW it works

### What Component Contains

Component contains what Container doesn't enforce:

| Element | Description |
|---------|-------------|
| Stack details | Which library, why chosen, exact configuration |
| Environment config | Env vars, defaults, dev vs prod differences |
| Implementation patterns | Conventions, algorithms, code patterns |
| Interfaces/Types | Method signatures, data structures, DTOs |
| Error handling | Specific error codes, retry strategies |
| Usage examples | Code snippets showing how to use it |

### Component Nature Types (Open-Ended)

Nature type determines documentation focus. Not a fixed taxonomy - use whatever helps code quality.

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

### Diagrams

Use common UML techniques where they help explain:
- **Flowchart** - decision logic, processing steps
- **Sequence diagram** - component interactions, request flow
- **ERD** - data relationships
- **State chart** - lifecycle, state transitions

Choice is contextual based on what needs explaining.

### Example Structure

```markdown
# COM-003 DB Pool (Resource Nature)

## Overview
Connection pooling for PostgreSQL

## Stack {#com-003-stack}
- Library: `pg` 8.11.x
- Why: Native driver, proven stability, supports LISTEN/NOTIFY

## Configuration {#com-003-config}
| Env Var | Dev | Prod | Why |
|---------|-----|------|-----|
| DB_POOL_MIN | 2 | 10 | Baseline connections |
| DB_POOL_MAX | 10 | 50 | Scale with load |
| DB_IDLE_TIMEOUT | 30s | 10s | Release faster in prod |

## Behavior {#com-003-behavior}

​```mermaid
stateDiagram-v2
    [*] --> Idle
    Idle --> Acquiring: getConnection()
    Acquiring --> Active: success
    Acquiring --> Waiting: pool exhausted
    Waiting --> Acquiring: connection released
    Waiting --> Error: timeout
    Active --> Idle: release()
    Error --> [*]
​```

## Error Handling {#com-003-errors}
| Error | Retriable | Action |
|-------|-----------|--------|
| Connection refused | Yes | Retry with backoff |
| Pool exhausted | Yes | Wait up to 5s |
| Query timeout | No | Propagate to caller |

## Usage {#com-003-usage}
​```typescript
const pool = createPool(config);
const result = await pool.query('SELECT * FROM users WHERE id = $1', [userId]);
​```
```

---

## Derivation Chain Summary

```
Context (WHAT exists, HOW they relate)
│
├── Protocols → CON-X#section, CON-Y#section
├── Cross-cutting → CON-X#section
│
↓
Container (WHAT it does, WITH WHAT)
│
├── Code Container
│   ├── Components → COM-X, COM-Y
│   ├── Relationships → Flowchart
│   ├── Data flow → Sequence diagram
│   └── Container cross-cutting → COM-Z
│
├── Infrastructure Container (LEAF)
│   └── Features → consumed by Code Container components
│
↓
Component (HOW it works)
│
├── Nature determines focus
├── Stack details, config, implementation
└── Terminal - no further derivation
```

## Implementation Tasks

1. Update `skills/c3-context-design/SKILL.md` with:
   - Downward linking to Container#sections
   - Protocol and cross-cutting tables with implementation links

2. Update `skills/c3-container-design/SKILL.md` with:
   - Two container types (Code vs Infrastructure)
   - Flowchart for component relationships
   - Sequence diagram for data flow
   - Container cross-cutting section
   - Downward linking to Component docs

3. Update `skills/c3-component-design/SKILL.md` with:
   - Open-ended nature types
   - Stack details ownership
   - Environment configuration focus
   - Appropriate diagram guidance

4. Update example documents to follow new structure

5. Update `skills/c3-adopt/SKILL.md` to create documents following this model
