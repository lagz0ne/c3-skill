---
id: CTX-001-system-overview
title: TaskFlow System Architecture Overview
summary: >
  Bird's-eye view of TaskFlow, the actors, containers, protocols, and cross-cutting
  concerns. Explains how users interact and how containers communicate—start here,
  then drill into containers and components via downward links.
---

# [CTX-001-system-overview] TaskFlow System Architecture Overview

## Overview {#ctx-001-overview}

TaskFlow is a task management application that allows users to create, organize, and track tasks. The system consists of a web frontend, a REST API backend, and a PostgreSQL database.

**Key Features:**
- User authentication and authorization
- Task creation, editing, and deletion
- Task categorization and filtering
- Due date tracking and notifications

## System Boundary {#ctx-001-boundary}
- Inside: Backend API service, Frontend SPA, managed Postgres database
- Outside: End users, observability pipeline, email service

## Actors {#ctx-001-actors}
| Actor | Role |
|-------|------|
| End User | Submits tasks via HTTP API/SPA |
| Email Service | Delivers notifications triggered by backend |

## Architecture {#ctx-001-architecture}

```mermaid
graph TB
    Users[Users] -->|HTTPS| Frontend[Web Frontend]
    Frontend -->|REST API| Backend[Backend API]
    Backend -->|PostgreSQL| DB[(Database)]
    Backend -->|SMTP| Email[Email Service]

    subgraph "TaskFlow System"
        Frontend
        Backend
        DB
    end

    subgraph "External"
        Email
    end
```

## Containers {#ctx-001-containers}

| Container | Type (Code/Infra) | Description |
|-----------|-------------------|-------------|
| [C3-1-backend](./containers/C3-1-backend.md) | Code | REST API handling business logic |
| [C3-2-frontend](./containers/C3-2-frontend.md) | Code | Web user interface |
| [C3-3-postgres](./containers/C3-3-postgres.md) | Infrastructure | Data persistence |

## Protocols {#ctx-001-protocols}

| From | To | Protocol | Implementations |
|------|----|----------|-----------------|
| Frontend | Backend | REST/HTTPS | [C3-2-frontend#c3-2-api-calls](./containers/C3-2-frontend.md#c3-2-api-calls), [C3-1-backend#c3-1-rest-endpoints](./containers/C3-1-backend.md#c3-1-rest-endpoints) |
| Backend | Postgres | SQL/TCP | [C3-1-backend#c3-1-db-access](./containers/C3-1-backend.md#c3-1-db-access), [C3-3-postgres#c3-3-features](./containers/C3-3-postgres.md#c3-3-features) |
| Backend | Email Service | SMTP/TLS | [C3-1-backend#c3-1-email-integration](./containers/C3-1-backend.md#c3-1-email-integration) |

### REST API Protocol {#ctx-001-rest-api}

Primary protocol for frontend-backend communication. Chosen for simplicity and broad tooling support.

```mermaid
sequenceDiagram
    participant U as User
    participant F as Frontend
    participant B as Backend
    participant D as Database

    U->>F: Create Task
    F->>B: POST /api/tasks
    B->>D: INSERT INTO tasks
    D-->>B: task_id
    B-->>F: 201 Created + task
    F-->>U: Show new task
```

**API Characteristics:**
- JSON request/response bodies
- JWT-based authentication
- Versioned via URL prefix (`/api/v1/`)
- Rate limited at 100 requests/minute per user

See [ADR-001: REST API Choice](./adr/ADR-001-rest-api.md) for decision rationale.

### Database Protocol {#ctx-001-database-protocol}

PostgreSQL wire protocol over TCP:
- Connection pooling (10-50 connections)
- SSL/TLS encryption in production
- Read replicas for scaling (future)

## Cross-Cutting Concerns {#ctx-001-cross-cutting}

### Authentication {#ctx-001-authentication}

JWT-based authentication with refresh tokens:
- Access tokens: 15-minute expiry
- Refresh tokens: 7-day expiry, stored in httpOnly cookies
- Token validation at API gateway level

**Implemented in:**
- [C3-1-backend#c3-1-auth-middleware](./containers/C3-1-backend.md#c3-1-auth-middleware) - Token validation
- [C3-2-frontend#c3-2-auth-handling](./containers/C3-2-frontend.md#c3-2-auth-handling) - Token storage and refresh

### Logging {#ctx-001-logging}

Structured JSON logging across all containers:
- Correlation IDs for request tracing
- Log levels: DEBUG, INFO, WARN, ERROR
- Centralized logging via stdout (container-friendly)

**Implemented in:**
- [C3-1-backend#c3-1-logging](./containers/C3-1-backend.md#c3-1-logging) - Backend logging
- [C3-2-frontend#c3-2-logging](./containers/C3-2-frontend.md#c3-2-logging) - Frontend logging

### Error Handling {#ctx-001-error-handling}

Consistent error response format:
```json
{
  "error": {
    "code": "TASK_NOT_FOUND",
    "message": "Task with ID 123 not found",
    "correlationId": "abc-123-def"
  }
}
```

**Implemented in:**
- [C3-1-backend#c3-1-error-handling](./containers/C3-1-backend.md#c3-1-error-handling) - Error formatting
- [C3-2-frontend#c3-2-error-handling](./containers/C3-2-frontend.md#c3-2-error-handling) - Error display

## Deployment {#ctx-001-deployment}

Docker-based deployment suitable for:
- Local development (docker-compose)
- Cloud deployment (Kubernetes, ECS)
- Simple VPS hosting

```mermaid
graph LR
    subgraph "Production Environment"
        LB[Load Balancer]
        LB --> B1[Backend 1]
        LB --> B2[Backend 2]
        B1 --> DB[(PostgreSQL)]
        B2 --> DB
    end
```

**Scaling Strategy:**
- Horizontal scaling for backend (stateless)
- Vertical scaling for database (single primary)
- CDN for static frontend assets

## Related {#ctx-001-related}
- [ADR-001: REST API Choice](./adr/ADR-001-rest-api.md)
- [ADR-002: PostgreSQL Choice](./adr/ADR-002-postgresql.md)
