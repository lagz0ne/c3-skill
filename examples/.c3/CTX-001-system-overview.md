# CTX-001 TaskFlow System Overview

## System Boundary {#ctx-001-boundary}
- Inside: Backend API service, managed Postgres database
- Outside: End users, observability pipeline

## Actors {#ctx-001-actors}
| Actor | Role |
|-------|------|
| End User | Submits tasks via HTTP API |

## Containers {#ctx-001-containers}
| Container | Type (Code/Infra) | Description |
|-----------|-------------------|-------------|
| [CON-001-backend](./containers/CON-001-backend.md) | Code | REST API for task submission and retrieval |
| [CON-002-postgres](./containers/CON-002-postgres.md) | Infrastructure | Durable task storage |

## Protocols {#ctx-001-protocols}
| From | To | Protocol | Implementations |
|------|----|----------|-----------------|
| Backend API | Postgres | SQL | [CON-001-backend#con-001-protocols](./containers/CON-001-backend.md#con-001-protocols), [CON-002-postgres#con-002-features](./containers/CON-002-postgres.md#con-002-features) |

## Cross-Cutting {#ctx-001-cross-cutting}
- Authentication: implemented in [CON-001-backend#con-001-cross-cutting](./containers/CON-001-backend.md#con-001-cross-cutting)
- Logging/observability: implemented in [CON-001-backend#con-001-cross-cutting](./containers/CON-001-backend.md#con-001-cross-cutting)
- Error strategy: implemented in [CON-001-backend#con-001-cross-cutting](./containers/CON-001-backend.md#con-001-cross-cutting)

## Deployment Topology {#ctx-001-deployment}
- Single backend container behind HTTPS terminating load balancer
- Managed Postgres instance in the same region
