---
name: c3-container-design
description: Explore Container level impact during scoping - technology choices, component organization, middleware, and inter-container communication
---

# C3 Container Level Exploration

## Overview

Explore Container-level impact during the scoping phase of c3-design. Container is the middle layer: individual services, their technology, and component organization.

**Abstraction Level:** WHAT and WHY, not HOW. Characteristics and architecture, not implementation code.

**Announce at start:** "I'm using the c3-container-design skill to explore Container-level impact."

## When Invoked

Called during EXPLORE phase of c3-design when:
- Hypothesis suggests Container-level impact
- Need to understand service-level implications
- Exploring downstream from Context
- Exploring upstream from Component
- Change affects technology stack or middleware

Also called by c3-adopt to CREATE initial Container documentation.

---

## Two Container Types

Containers come in two types with different documentation focus:

| Type | Has Components? | Documentation Focus |
|------|-----------------|---------------------|
| **Code Container** | Yes | Tech stack, component inventory, how protocols are implemented |
| **Infrastructure Container** | No (leaf node) | Engine, config, features provided to code containers |

**Infrastructure containers are LEAF NODES** - no component level beneath them. Their features become inputs to code container components.

---

## What Belongs at Container Level

### Inclusion Criteria

**INCLUDE at Container level:**

| Element | Why Container | Example |
|---------|--------------|---------|
| Technology stack | Container-specific choices | Node.js 20, Express 4.18 |
| Container responsibilities | What this container does | "Handles API requests" |
| Component relationships | How components connect (flowchart) | Entry → Auth → Business → DB |
| Data flow | How data moves through (sequence diagram) | Request → Validate → Process → Store |
| Component inventory | WHAT components exist (links to Component docs) | [COM-001], [COM-002] |
| Container cross-cutting | Logging, error handling within container | [COM-006-logger], [COM-007-error-handler] |
| Protocol implementations | How this container implements Context protocols | [COM-001#rest-endpoints] |
| API surface | Endpoints exposed | `POST /api/v1/tasks` |
| Data ownership | What data this owns | "User accounts, Tasks" |
| Inter-container communication | How it talks to siblings | "REST to Backend, SQL to DB" |
| Configuration approach | How config is managed | Environment variables |
| Deployment specifics | Container deployment | Docker image, resources |

**EXCLUDE from Container (push to Context or Component):**

| Element | Why Not Container | Where It Belongs |
|---------|------------------|------------------|
| System boundary | Affects multiple containers | Context |
| Cross-cutting concerns | Span containers | Context |
| Protocol decisions | System-wide | Context |
| Implementation code | Too detailed | Component |
| Library specifics | Implementation | Component |
| Configuration values | Implementation | Component |
| Error handling details | Implementation | Component |
| Algorithm specifics | Implementation | Component |

### Litmus Test

Ask: "Is this about WHAT this container does and WITH WHAT, not HOW it does it internally?"
- **Yes** → Container level
- **No (system-wide)** → Push up to Context
- **No (implementation)** → Push down to Component

---

## Expressing Relationships at Container Level

### Relationship Types

| Relationship | Expression | Example |
|--------------|------------|---------|
| Container → Container | Protocol + purpose | "Calls Auth Service via gRPC for validation" |
| Container → Database | Connection type | "PostgreSQL via connection pool" |
| Container → External | Integration type | "SMTP to SendGrid" |
| Layer → Layer (internal) | Arrow with label | Routes → Services → Repositories |
| Component → Component | Dependency | "TaskService depends on DBPool" |

### Relationship Table Format

```markdown
## Component Dependencies

| Component | Depends On | Relationship |
|-----------|------------|--------------|
| TaskService | DBPool | Uses for queries |
| TaskService | AuthMiddleware | Protected by |
| Routes | TaskService | Delegates to |
```

### Internal Structure Diagram

Show layers and component groups:

```markdown
## Component Organization

| Layer | Components | Responsibility |
|-------|------------|----------------|
| API | Routes, Middleware | HTTP handling |
| Business | Services, Validators | Domain logic |
| Data | Repositories, DBPool | Persistence |
```

### DO NOT Express at Container

- Actor interactions (Context level)
- System-wide protocols (Context level)
- Method signatures (Component level)
- Data structures (Component level)

---

## Diagrams for Container Level

### Primary: Component Relationships Flowchart

**Purpose:** Show how components connect to each other within the container.

```mermaid
flowchart LR
    Entry[REST Routes] --> Auth[Auth Middleware]
    Auth --> Business[Order Flow]
    Business --> DB[DB Pool]
    DB --> External[Postgres]

    Auth -.-> Log[Logger]
    Business -.-> Log
    DB -.-> Log
```

**When to use:** Always include to show component organization and dependencies.

### Secondary: Data Flow Sequence Diagram

**Purpose:** Show how data moves through the container during a request.

```mermaid
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
```

**When to use:** Always include to show data transformation and request lifecycle.

### Tertiary: Middleware Pipeline Diagram

**Purpose:** Show request processing flow through middleware.

```mermaid
graph LR
    A[Request] --> B[CORS]
    B --> C[Body Parser]
    C --> D[Auth]
    D --> E[Rate Limit]
    E --> F[Handler]
    F --> G[Error Handler]
    G --> H[Response]
```

**When to use:** When container has significant middleware/pipeline.

### Avoid at Container Level

| Diagram Type | Why Not | Where It Belongs |
|--------------|---------|------------------|
| System context diagram | Too high level | Context |
| Actor diagrams | System level | Context |
| Class diagrams with methods | Too detailed | Component |
| Detailed sequence with code | Implementation | Component |
| State machines for logic | Implementation | Component |

---

## Container Level Defines

| Concern | Examples |
|---------|----------|
| **Container identity** | Name, purpose, responsibilities |
| **Technology stack** | Language, framework, runtime |
| **Component organization** | Internal structure, layers |
| **Middleware pipeline** | Auth, rate limiting, request flow |
| **APIs** | Endpoints exposed and consumed |
| **Data responsibilities** | What data this container owns |
| **Deployment specifics** | Container-level deployment |

## Exploration Questions

When exploring Container level, investigate:

### Isolated (at Container)
- What container responsibilities change?
- What middleware pipeline affected?
- What APIs need modification?

### Upstream (to Context)
- Does this change system boundaries?
- Do protocols need updating?
- Are cross-cutting concerns affected?

### Adjacent (same level)
- What sibling containers related?
- What inter-container communication affected?
- What shared dependencies exist?

### Downstream (to Components)
- Which components inside this container affected?
- What new components needed?
- How does component organization change?

## Socratic Questions for Container Discovery

When creating or validating Container documentation, ask:

### Identity & Purpose
1. "What is the single responsibility of this container?"
2. "If this container disappeared, what would break?"
3. "What would you name this container in one word?"

### Technology
4. "What language and framework does this use?"
5. "Why was this technology chosen over alternatives?"
6. "What are the key libraries/dependencies?"

### Structure
7. "How is code organized inside? Layers? Modules?"
8. "What are the main entry points?"
9. "How do requests flow through this container?"

### APIs
10. "What endpoints does this container expose?"
11. "What APIs does it consume from other containers?"
12. "What is the API versioning strategy?"

### Data
13. "What data does this container own?"
14. "What data does it read from other sources?"
15. "How is data validated and transformed?"

### Configuration
16. "How is this container configured?"
17. "What differs between dev and production?"
18. "What secrets are required?"

## Reading Container Documents

Use c3-locate to retrieve:

```
c3-locate CON-001                    # Overview
c3-locate #con-001-technology-stack  # Tech choices
c3-locate #con-001-middleware        # Request pipeline
c3-locate #con-001-components        # Internal structure
c3-locate #con-001-api-endpoints     # API surface
c3-locate #con-001-communication     # Inter-container
c3-locate #con-001-data              # Data ownership
c3-locate #con-001-configuration     # Config approach
c3-locate #con-001-deployment        # Deployment details
```

## Impact Signals

| Signal | Meaning |
|--------|---------|
| New middleware layer needed | Cross-component change |
| API contract change | Consumers affected |
| Technology stack change | Major container rewrite |
| Data ownership change | Migration needed |
| New container needed | Context-level impact |

## Output for c3-design

After exploring Container level, report:
- What Container-level elements are affected
- Impact on adjacent containers
- Components that need deeper exploration
- Whether Context level needs revisiting
- Whether hypothesis needs revision

## Document Template Reference

### Code Container Template

Code containers have components and use **downward linking**:

```markdown
---
id: CON-NNN-slug
title: [Container Name] Container (Code)
summary: >
  [Why read this document - what it covers]
---

# [CON-NNN-slug] [Container Name] Container (Code)

## Overview {#con-nnn-overview}
<!--
High-level description of container purpose and responsibilities.
-->

## Technology Stack {#con-nnn-stack}
- Runtime: Node.js 20
- Framework: Express 4.18
- Language: TypeScript 5.x

## Component Relationships {#con-nnn-relationships}
<!--
Flowchart showing how components connect.
-->
```mermaid
flowchart LR
    Entry[REST Routes] --> Auth[Auth Middleware]
    Auth --> Business[Order Flow]
    Business --> DB[DB Pool]
    DB --> External[Postgres]

    Auth -.-> Log[Logger]
    Business -.-> Log
    DB -.-> Log
```

## Data Flow {#con-nnn-data-flow}
<!--
Sequence diagram showing how data moves through container.
-->
```mermaid
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
```

## Container Cross-Cutting {#con-nnn-cross-cutting}

### Logging {#con-nnn-logging}
- Structured JSON, correlation IDs passed through
- Implemented by: [COM-006-logger](./components/CON-NNN/COM-006-logger.md)

### Error Handling {#con-nnn-error-handling}
- Unified error format, error codes catalog
- Implemented by: [COM-007-error-handler](./components/CON-NNN/COM-007-error-handler.md)

## Components {#con-nnn-components}
<!--
Links DOWN to Component docs. Reader follows links to dive deeper.
-->
| Component | Nature | Responsibility |
|-----------|--------|----------------|
| [COM-001-rest-routes](./components/CON-NNN/COM-001-rest-routes.md) | Entrypoint | HTTP handling |
| [COM-002-auth-middleware](./components/CON-NNN/COM-002-auth-middleware.md) | Cross-cutting | Token validation |
| [COM-003-db-pool](./components/CON-NNN/COM-003-db-pool.md) | Resource | Connection management |
| [COM-004-order-flow](./components/CON-NNN/COM-004-order-flow.md) | Business | Order processing |

## Related {#con-nnn-related}
```

### Code Container Checklist (must be true to call CON done)

- [ ] Technology stack recorded
- [ ] Protocol implementations table maps every CTX protocol to specific components/sections
- [ ] Flowchart shows component relationships (must exist)
- [ ] Sequence diagram shows data flow (must exist)
- [ ] Cross-cutting choices mapped to implementing components
- [ ] Component inventory complete with Nature + Responsibility
- [ ] All anchors use `{#con-xxx-*}` format for stable linking

### Infrastructure Container Template

Infrastructure containers are **LEAF NODES** - no components, focus on features provided:

```markdown
---
id: CON-NNN-slug
title: [Infrastructure Name] Container (Infrastructure)
summary: >
  [Why read this document - what it covers]
---

# [CON-NNN-slug] [Infrastructure Name] Container (Infrastructure)

## Engine {#con-nnn-engine}
PostgreSQL 15

## Configuration {#con-nnn-config}
| Setting | Value | Why |
|---------|-------|-----|
| max_connections | 100 | Support pooling from backend |
| wal_level | logical | Enable event streaming |

## Features Provided {#con-nnn-features}
<!--
Features that code containers consume. No downward links - this is a leaf node.
-->
| Feature | Used By |
|---------|---------|
| WAL logical replication | [CON-001-backend] → [COM-005-event-streaming] |
| LISTEN/NOTIFY | [CON-001-backend] → [COM-003-db-pool] |
```

### Infrastructure Container Checklist (must be true to call CON done)

- [ ] Engine/version stated
- [ ] Configuration table with settings and rationale
- [ ] Features table lists capabilities with links to consuming code containers/components
- [ ] No component-level sections (this is a leaf node)
- [ ] All anchors use `{#con-xxx-*}` format for stable linking

### Reference Direction Principle

**References only flow DOWN** from Container to Component:

- Container defines components → links to Component docs
- Container defines cross-cutting → links to Component docs implementing them
- **No upward links needed** - reader already came from Context

Use these heading IDs for precise exploration.
