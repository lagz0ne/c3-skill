# Table of Contents

## Context
- [CTX-001-system-overview](./CTX-001-system-overview.md) — System landscape and protocols

## Containers
- [C3-1-backend](./containers/C3-1-backend.md) — REST API service, business logic, persistence
- [C3-2-frontend](./containers/C3-2-frontend.md) — React SPA consuming the backend
- [C3-3-postgres](./containers/C3-3-postgres.md) — Managed PostgreSQL (leaf infrastructure)

## Components
**Backend**
- [C3-106-rest-routes](./components/backend/C3-106-rest-routes.md) — HTTP entrypoint
- [C3-102-auth-middleware](./components/backend/C3-102-auth-middleware.md) — Token validation
- [C3-103-task-service](./components/backend/C3-103-task-service.md) — Business logic
- [C3-101-db-pool](./components/backend/C3-101-db-pool.md) — Connection pooling
- [C3-104-logger](./components/backend/C3-104-logger.md) — Structured logging
- [C3-105-error-handler](./components/backend/C3-105-error-handler.md) — Error formatting

**Frontend**
- [C3-201-api-client](./components/frontend/C3-201-api-client.md) — Backend communication wrapper

## ADRs
- [ADR-001-rest-api](./adr/ADR-001-rest-api.md) — Protocol choice for frontend-backend communication
- [ADR-002-postgresql](./adr/ADR-002-postgresql.md) — Database engine selection
