---
id: c3-4
c3-version: 3
title: PostgreSQL
type: container
parent: c3-0
summary: Primary relational database for all business data
---

# PostgreSQL

PostgreSQL serves as the primary data store for all Acountee business entities including invoices, payment requests, approvals, users, and audit logs. Connected via Drizzle ORM from the API Backend.

## Overview

```mermaid
graph LR
    c3-2[API Backend] -->|Drizzle ORM| c3-4[(PostgreSQL)]

    subgraph c3-4 [PostgreSQL]
        invoices[(invoices)]
        pr[(pr)]
        approval[(approval)]
        users[(users)]
        audit[(audit)]
        files[(files)]
    end
```

## Configuration

| Variable | Purpose | Required |
|----------|---------|----------|
| PGURI | Connection string (postgres://user:pass@host:port/db) | Yes |

## Components

### Tables
> Database tables organized by domain.

| ID | Name | Status | Responsibility |
|----|------|--------|----------------|
| c3-401 | Invoice Tables | active | invoices, invoice_services - imported invoice data |
| c3-402 | PR Tables | active | pr, pr_invoice - payment request records |
| c3-403 | Approval Tables | active | approval, approval_steps, approval_step_users, approval_records |
| c3-404 | User Tables | active | users, roles, user_roles - identity and permissions |
| c3-405 | Audit Tables | active | audit, security_events - logging and compliance |
| c3-406 | File Storage | active | files - binary file storage (bytea) |

## Schema Relationships

| Parent | Child | Relationship |
|--------|-------|--------------|
| invoices | invoice_services | 1:N (line items) |
| pr | approval | 1:1 (approval workflow) |
| approval | approval_steps | 1:N (workflow steps) |
| approval_steps | approval_step_users | 1:N (approvers per step) |
| approval | approval_records | 1:N (who approved) |
| pr | pr_invoice | N:M (linked invoices) |
| users | user_roles | 1:N (role assignments) |

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Connection failure | Application fails to start |
| Pool exhausted | Queries wait, eventually timeout |
| Transaction deadlock | PostgreSQL detects and aborts one |
| Disk full | Write operations fail |

## Testing

| Scenario | Verifies |
|----------|----------|
| Connection | Can connect with PGURI |
| Schema integrity | Foreign keys enforced |
| Transaction rollback | Changes reverted on error |
| Concurrent access | No deadlocks in normal operations |
