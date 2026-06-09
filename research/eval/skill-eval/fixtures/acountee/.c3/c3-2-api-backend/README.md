---
id: c3-2
c3-version: 3
c3-seal: 83a7e426986ee6e06104181dc45551974c1a76fd65508a11d2cab3c352d08928
title: API Backend
type: container
parent: c3-0
goal: Execute business operations (approvals, payments, invoices) through typed flows, persist data via Drizzle, handle auth, broadcast updates to clients.
---

# API Backend

## Goal

Execute business operations (approvals, payments, invoices) through typed flows, persist data via Drizzle, handle auth, broadcast updates to clients.

## Responsibilities

- Own server-side orchestration for domain mutations and queries.
- Enforce authentication, authorization, and request-scoped execution context.
- Persist domain state and publish real-time synchronization events.

## Components

| ID | Name | Category | Status | Goal Contribution |
| --- | --- | --- | --- | --- |
| c3-201 | Flow Pattern | foundation | active | Standardized, typed orchestration model for backend actions |
| c3-202 | Execution Context | foundation | active | Scoped request metadata and transaction coordination |
| c3-203 | Middleware Stack | foundation | active | Transport-level auth, validation, and execution wiring |
| c3-204 | Drizzle ORM | foundation | active | Typed persistence and query composition |
| c3-209 | NATS Credential Generator | foundation | active | Auth material for secure real-time connections |
| c3-211 | Notification System | foundation | active | Durable and targeted notification delivery infrastructure |
| c3-213 | Authentication Flows | foundation | active | User/session authentication workflows |
| c3-205 | PR Flows | feature | active | Payment request business workflows |
| c3-206 | Invoice Flows | feature | active | Invoice import and linking workflows |
| c3-207 | Payment Flows | feature | active | Payment method lifecycle workflows |
| c3-208 | Audit Flows | feature | active | Audit query/export behavior |
| c3-210 | Admin Flows | feature | active | Team, role, user, and approval configuration management |
| c3-212 | Workbench Flows | feature | active | Bulk operational workflows for finance workbench |
| c3-214 | User Notification Flows | feature | active | End-user notification retrieval and acknowledgement flows |
| c3-215 | Slack Bot Integration | feature | active | Slack bot for PR approval notifications, commands, and interactive actions |

## Wiring

| From | To | Protocol | What |
| --- | --- | --- | --- |
| c3-1 (Frontend) | /act | HTTP | Mutations & queries |
| /auth.google.url | Google OAuth | HTTPS | Authentication |
| NATS Publisher | c3-4 (NATS) | TCP | Broadcast sync events |
| Query Services | PostgreSQL | TCP/SQL | Data access with RLS |
| Flows | OpenTelemetry | OTLP/HTTP | Distributed tracing |

## Entry Points

- **`POST /act`** — Single endpoint for mutations (FormData: action + data → executionId)
- **`/auth.google.url`, `/cb`** — OAuth flow
- **NATS Publisher** — Async sync events via JetStream

## Core Pattern: flow()

All business logic uses `flow()` with DI, Zod validation, tracing:

```typescript
export const approvePr = flow({
  deps: { prService, logger, sync },       // DI
  parse: ApprovePr.schema.parse,            // Validation
  factory: async (ctx, deps) => {           // Logic
    const user = ctx.data.seekTag(currentUserTag);
    // ...
    return { success: true };
  }
});
```

## Middleware Stack

- **executionContextMiddleware** — Request-scoped context with tags (currentUser, transaction, executionId)
- **getCurrentUserMiddleware** — Extract user from auth headers

## Data Layer

- **Drizzle ORM** — Type-safe SQL via schema + query services
- **Transaction tag** — All DB access within request scope
- **set_config('request.jwt.claim.user_id', ...)** — PostgreSQL RLS

## Domains

| Domain | Flows |
| --- | --- |
| Payment Requests | approvePr, rejectPr, createPr, unlinkPr, etc. |
| Invoices | linkPaymentRequest, importFiles, etc. |
| Payments | createPayment, updatePayment, deletePayment |
| Admin | Team, role, user, approval config (owner role) |
| Workbench | Invoice cleanup, PR export, paid PR import |
| Audit | getAuditHistory, exportAuditTrail |
| Notifications | User-facing fetch, read, dismiss + JetStream publisher |
| Auth | Google OAuth, test token flow, NATS credential generator |

## Tech Stack

| Layer | Technology |
| --- | --- |
| Runtime | Bun |
| Framework | TanStack Start |
| ORM | Drizzle |
| DI | @pumped-fn/lite |
| Validation | Zod |
| Tracing | OpenTelemetry |
