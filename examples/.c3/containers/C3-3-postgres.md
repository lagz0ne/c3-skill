id: C3-3-postgres
title: PostgreSQL Container (Infrastructure)
summary: >
  Managed PostgreSQL for TaskFlow. Leaf infrastructure container—features are consumed by code containers; no component level beneath.
---

# [C3-3-postgres] PostgreSQL Container (Infrastructure)

## Engine {#c3-3-engine}

PostgreSQL 15 (managed service)

**Why PostgreSQL:**
- Strong ACID compliance for data integrity
- Rich query capabilities (JSON, full-text search)
- Mature ecosystem and tooling
- Excellent performance for relational workloads

See [ADR-002: PostgreSQL Choice](../adr/ADR-002-postgresql.md) for decision rationale.

## Configuration {#c3-3-config}

| Setting | Value | Why |
|---------|-------|-----|
| `max_connections` | 100 | Support pooling from backend (50) + headroom |
| `shared_buffers` | 256MB | 25% of container memory |
| `work_mem` | 4MB | Per-operation memory for sorting |
| `maintenance_work_mem` | 64MB | For VACUUM, CREATE INDEX |
| `wal_level` | logical | Enable event streaming (future) |
| `log_statement` | ddl | Log schema changes |
| `log_min_duration_statement` | 1000 | Log slow queries (>1s) |

### Development vs Production {#c3-3-dev-vs-prod}

| Setting | Development | Production |
|---------|-------------|------------|
| `max_connections` | 20 | 100 |
| `shared_buffers` | 128MB | 256MB |
| `log_statement` | all | ddl |
| `ssl` | off | on |

## Features Provided {#c3-3-features}

| Feature | Used By | Purpose |
|---------|---------|---------|
| Connection pooling support | [C3-1-backend](./C3-1-backend.md#c3-1-db-access) → [C3-101-db-pool](../components/backend/C3-101-db-pool.md) | Efficient connection reuse |
| LISTEN/NOTIFY | [C3-1-backend](./C3-1-backend.md#c3-1-db-access) → [C3-101-db-pool](../components/backend/C3-101-db-pool.md) | Real-time event notifications |
| JSON/JSONB columns | [C3-1-backend](./C3-1-backend.md#c3-1-components) → [C3-103-task-service](../components/backend/C3-103-task-service.md) | Flexible task metadata |
| Full-text search | [C3-1-backend](./C3-1-backend.md#c3-1-components) → [C3-103-task-service](../components/backend/C3-103-task-service.md) | Task search functionality |
| WAL logical replication | Future: Event streaming | CDC for external systems |

## Schema Overview {#c3-3-schema}

```mermaid
erDiagram
    users {
        uuid id PK
        string email UK
        string password_hash
        timestamp created_at
    }
    tasks {
        uuid id PK
        uuid user_id FK
        string title
        text description
        string status
        timestamp due_date
        jsonb metadata
        timestamp created_at
    }
    users ||--o{ tasks : owns
```

## Backup & Recovery {#c3-3-backup}

| Aspect | Strategy |
|--------|----------|
| Full backup | Daily pg_dump at 02:00 UTC |
| Point-in-time | WAL archiving to S3 |
| Retention | 7 days full, 30 days WAL |
| RTO | < 1 hour |
| RPO | < 5 minutes |

## Health Checks {#c3-3-health}

```sql
-- Liveness check
SELECT 1;

-- Readiness check (connections available)
SELECT count(*) < 90 AS ready
FROM pg_stat_activity
WHERE state != 'idle';
```

## Deployment {#c3-3-deployment}

**Docker image:** `postgres:15-alpine`

**Resource limits:**
- Memory: 1GB
- CPU: 1 core
- Storage: 10GB (SSD)

**Persistence:**
- Data volume: `/var/lib/postgresql/data`
- Named volume in docker-compose
- EBS/Persistent disk in cloud

## Related {#c3-3-related}

- [CTX-001: System Overview](../CTX-001-system-overview.md#ctx-001-database-protocol)
- [C3-1: Backend](./C3-1-backend.md#c3-1-db-access) - Primary consumer
- [ADR-002: PostgreSQL Choice](../adr/ADR-002-postgresql.md)
