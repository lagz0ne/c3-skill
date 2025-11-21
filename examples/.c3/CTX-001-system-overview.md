---
id: CTX-001-system-overview
title: TaskFlow System Architecture Overview
summary: >
  Bird's-eye view of TaskFlow, the actors, containers, protocols, and cross-cutting
  concerns. Start here, then drill into containers and components via downward links.
---

# [CTX-001-system-overview] TaskFlow System Architecture Overview

## System Boundary {#ctx-001-boundary}
- Inside: Backend API service, Frontend SPA, managed Postgres database
- Outside: End users, observability pipeline, email service

## Actors {#ctx-001-actors}
| Actor | Role |
|-------|------|
| End User | Submits tasks via HTTP API/SPA |

## Containers {#ctx-001-containers}
| Container | Type (Code/Infra) | Description |
|-----------|-------------------|-------------|
| [CON-001-backend](./containers/CON-001-backend.md) | Code | REST API for task submission and retrieval |
| [CON-002-frontend](./containers/CON-002-frontend.md) | Code | React SPA for users |
| [CON-003-postgres](./containers/CON-003-postgres.md) | Infrastructure | Durable task storage |

## Protocols {#ctx-001-protocols}
| From | To | Protocol | Implementations |
|------|----|----------|-----------------|
| Frontend | Backend | REST/HTTPS | [CON-002-frontend#con-002-api-calls](./containers/CON-002-frontend.md#con-002-api-calls), [CON-001-backend#con-001-protocols](./containers/CON-001-backend.md#con-001-protocols) |
| Backend | Postgres | SQL/TCP | [CON-001-backend#con-001-protocols](./containers/CON-001-backend.md#con-001-protocols), [CON-003-postgres#con-003-features](./containers/CON-003-postgres.md#con-003-features) |

## Cross-Cutting {#ctx-001-cross-cutting}
- Authentication: implemented in [CON-001-backend#con-001-cross-cutting](./containers/CON-001-backend.md#con-001-cross-cutting) and [CON-002-frontend#con-002-auth-handling](./containers/CON-002-frontend.md#con-002-auth-handling)
- Logging/observability: implemented in [CON-001-backend#con-001-cross-cutting](./containers/CON-001-backend.md#con-001-cross-cutting)
- Error strategy: implemented in [CON-001-backend#con-001-cross-cutting](./containers/CON-001-backend.md#con-001-cross-cutting) and surfaced to users via [CON-002-frontend#con-002-error-handling](./containers/CON-002-frontend.md#con-002-error-handling)

## Deployment Topology {#ctx-001-deployment}
- Backend container behind HTTPS-terminating load balancer
- Managed Postgres instance in the same region
- Frontend served as static assets via CDN
