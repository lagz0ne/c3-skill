---
layout: home
title: C3 Architecture Documentation Example
---

# C3 Architecture Documentation Example

This is an example C3 documentation structure for a simple web application called **TaskFlow** - a task management system with backend, frontend, and database components.

## Reading Order

**Always top-down: Context → Container → Component**

References only flow DOWN - higher layers link to lower layer implementations.

## Navigation

### Context (Bird's Eye View)
- [CTX-001: System Overview](./CTX-001-system-overview.md) - Start here

### Containers (What & With What)

| Container | Type | Description |
|-----------|------|-------------|
| [CON-001: Backend](./containers/CON-001-backend.md) | Code | REST API handling business logic |
| [CON-002: Frontend](./containers/CON-002-frontend.md) | Code | Web user interface |
| [CON-003: Postgres](./containers/CON-003-postgres.md) | Infrastructure | Data persistence (leaf node) |

### Components (How It Works)
- Backend Components:
  - [COM-001: DB Pool](./components/backend/COM-001-db-pool.md) - Resource nature
  - [COM-002: Auth Middleware](./components/backend/COM-002-auth-middleware.md) - Cross-cutting nature
  - [COM-003: Task Service](./components/backend/COM-003-task-service.md) - Business logic nature

### Architecture Decision Records
- [ADR-001: REST API Choice](./adr/ADR-001-rest-api.md)

## Derivation Model

This example demonstrates the **hierarchical derivation relationship**:

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

## Conventions Used

| Level | ID Pattern | Location | Links Down To |
|-------|-----------|----------|---------------|
| Context | `CTX-NNN-slug` | `.c3/` | Container docs, Container#sections |
| Container (Code) | `CON-NNN-slug` | `.c3/containers/` | Component docs |
| Container (Infra) | `CON-NNN-slug` | `.c3/containers/` | - (leaf node) |
| Component | `COM-NNN-slug` | `.c3/components/{container}/` | - (terminal) |
| ADR | `ADR-NNN-slug` | `.c3/adr/` | - |
