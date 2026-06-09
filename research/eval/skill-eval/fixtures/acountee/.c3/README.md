---
id: c3-0
c3-version: 3
c3-seal: 956bd9e29c26f1fa8f9b36c53c6ea3934051ccac6ec821f03927182e6d5d9459
title: act
goal: Enable Finance, Admin, and BOD teams to collaborate on approval workflows with real-time sync and audit trails.
uses:
    - ref-jtbd
    - ref-zerobased-dev
---

# act

## Goal

Enable Finance, Admin, and BOD teams to collaborate on approval workflows with real-time sync and audit trails.

## System Context

```mermaid
C4Context
    title act - System Context

    Person(finance, "Finance Team", "Creates and processes payment requests")
    Person(admin, "Admin Team", "Manages invoices and approvals")
    Person(bod, "BOD", "Final approval authority")

    System_Boundary(act, "act System") {
        Container(web, "Web Frontend", "TanStack Start/React", "User interface for approval workflows")
        Container(api, "API Backend", "TanStack Start/Bun", "Business logic, flows, and data access")
    }

    System_Ext(postgres, "PostgreSQL", "Primary data store")
    System_Ext(google, "Google OAuth", "Authentication provider")
    System_Ext(otel, "OpenTelemetry Collector", "Distributed tracing")
    System_Ext(nats, "NATS", "Message broker for real-time sync")

    Rel(finance, web, "Uses")
    Rel(admin, web, "Uses")
    Rel(bod, web, "Uses")
    Rel(web, api, "HTTP")
    Rel(web, nats, "WebSocket")
    Rel(api, nats, "TCP pub/sub")
    Rel(api, postgres, "SQL")
    Rel(api, google, "OAuth 2.0")
    Rel(api, otel, "OTLP traces")
```

## Containers

| ID | Name | Responsibility |
| --- | --- | --- |
| c3-1 | Web Frontend | React UI for approval workflows, real-time sync via NATS |
| c3-2 | API Backend | Business logic via flow(), Drizzle ORM, auth |
| c3-4 | NATS Server | Real-time messaging with JWT auth |

## Abstract Constraints

| Constraint | Rationale | Affected Containers |
| --- | --- | --- |
| All business mutations must be auditable. | Compliance and debugging require reconstructable change history. | c3-1, c3-2 |
| UI state must converge via server-published real-time events. | Prevent stale client state after concurrent mutations. | c3-1, c3-2, c3-4 |
| Access must be identity- and capability-scoped. | Protect admin and operational actions by role/team policies. | c3-1, c3-2 |

## Wiring

| From | To | Protocol | What |
| --- | --- | --- | --- |
| Teams | c3-1 | HTTPS | Web access |
| c3-1 | c3-2 | HTTP | API calls |
| c3-1 | c3-4 | WSS | Real-time sync |
| c3-2 | c3-4 | TCP | Publish events |
| c3-2 | PostgreSQL | TCP/SQL | Data persistence |
| c3-2 | Google OAuth | HTTPS | Authentication |
| c3-2 | OpenTelemetry | OTLP/HTTP | Tracing |

## Deployment

Single unit (`apps/start`) with logical separation:

- `/src/routes/` → Frontend routes
- `/src/server/` → Flows & API routes
- `/src/components/`, `/src/screens/` → React UI
