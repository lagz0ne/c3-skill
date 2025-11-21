---
layout: home
title: C3 Architecture Documentation Example
---

# C3 Architecture Documentation Example

Example C3 structure for TaskFlow with downward-only derivation and linked examples.

## Reading Order
**Always top-down: Context → Container → Component**

## Navigation

### Context
- [CTX-001: System Overview](./CTX-001-system-overview.md)

### Containers
| Container | Type | Description |
|-----------|------|-------------|
| [CON-001: Backend](./containers/CON-001-backend.md) | Code | REST API handling business logic |
| [CON-002: Frontend](./containers/CON-002-frontend.md) | Code | Web user interface |
| [CON-003: Postgres](./containers/CON-003-postgres.md) | Infrastructure | Data persistence (leaf node) |

### Components
- Backend:
  - [COM-001: REST Routes](./components/backend/COM-001-rest-routes.md)
  - [COM-002: DB Pool](./components/backend/COM-002-db-pool.md)
  - [COM-003: Logger](./components/backend/COM-003-logger.md)
- Frontend:
  - [COM-004: API Client](./components/frontend/COM-004-api-client.md)

### ADRs
- [ADR-001: REST API Choice](./adr/ADR-001-rest-api.md)

## Derivation Model
```
Context (WHAT exists, HOW they relate)
│
├── Protocols → CON-X#section, CON-Y#section
├── Cross-cutting → CON-X#section
│
↓
Container (WHAT it does, WITH WHAT)
│
├── Code Container → components, relationships (flowchart), data flow (sequence)
├── Cross-cutting → COM links
└── Infra Container (LEAF) → features consumed by Code
│
↓
Component (HOW it works)
   ├─ Nature-driven focus
   ├─ Stack, config, interfaces, behavior, errors, usage
   └─ Terminal leaf
```
