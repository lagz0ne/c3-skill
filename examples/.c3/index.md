---
layout: home
title: C3 Architecture Documentation Example
---

# C3 Architecture Documentation Example

Example C3 structure for TaskFlow with downward-only derivation and linked examples. Start at Context, then drill into Containers and Components.

## Reading Order

**Always top-down: Context → Container → Component**

References only flow DOWN - higher layers link to lower layer implementations.

## Navigation

### Context (Bird's Eye View)
- [CTX-001: System Overview](./CTX-001-system-overview.md) - Start here

### Containers (What & With What)

| Container | Type | Description |
|-----------|------|-------------|
| [C3-1: Backend](./containers/C3-1-backend.md) | Code | REST API handling business logic |
| [C3-2: Frontend](./containers/C3-2-frontend.md) | Code | Web user interface |
| [C3-3: Postgres](./containers/C3-3-postgres.md) | Infrastructure | Data persistence (leaf node) |

### Components (How It Works)
- Backend Components:
  - [C3-101: DB Pool](./components/backend/C3-101-db-pool.md) - Resource nature
  - [C3-102: Auth Middleware](./components/backend/C3-102-auth-middleware.md) - Cross-cutting nature
  - [C3-103: Task Service](./components/backend/C3-103-task-service.md) - Business logic nature
  - [C3-104: Logger](./components/backend/C3-104-logger.md) - Cross-cutting nature
  - [C3-105: Error Handler](./components/backend/C3-105-error-handler.md) - Cross-cutting nature
  - [C3-106: REST Routes](./components/backend/C3-106-rest-routes.md) - Entrypoint example
- Frontend Components:
  - [C3-201: API Client](./components/frontend/C3-201-api-client.md)

### Architecture Decision Records
- [ADR-001: REST API Choice](./adr/ADR-001-rest-api.md)
- [ADR-002: PostgreSQL Choice](./adr/ADR-002-postgresql.md)

## Derivation Model

This example demonstrates the **hierarchical derivation relationship**:

```
Context (WHAT exists, HOW they relate)
│
├── Protocols → C3-<digit>#section
├── Cross-cutting → C3-<digit>#section
│
↓
Container (WHAT it does, WITH WHAT)
│
├── Code Container
│   ├── Components → C3-<digit><NN>
│   ├── Relationships → Flowchart
│   ├── Data flow → Sequence diagram
│   └── Container cross-cutting → C3-<digit><NN>
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

## Conventions Used

| Level | ID Pattern | Location | Links Down To |
|-------|-----------|----------|---------------|
| Context | `CTX-###-slug` | `.c3/` | Container docs, Container#sections |
| Container (Code/Infra) | `C3-<C>-slug` (`C` = container digit) | `.c3/containers/` | Component docs (if code) |
| Component | `C3-<C><NN>-slug` (inherits container digit) | `.c3/components/{container}/` | - (terminal) |
| ADR | `ADR-###-slug` | `.c3/adr/` | - |
