---
id: ADR-002-postgresql
title: PostgreSQL for Data Persistence
status: accepted
date: 2024-01-15
---

# ADR-002: PostgreSQL for Data Persistence

## Status

Accepted

## Context

We need a database for storing user accounts and tasks with ACID guarantees.

## Decision

Use PostgreSQL 15 as the primary data store.

## Rationale

- **ACID compliance**: Strong consistency for financial-grade data integrity
- **JSON support**: JSONB columns for flexible task metadata
- **Ecosystem**: Excellent tooling, ORMs (Prisma), and managed hosting options
- **Scalability**: Read replicas, partitioning for future growth
- **Full-text search**: Built-in for task search functionality

## Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| MySQL | Familiar, fast reads | Weaker JSON support |
| MongoDB | Flexible schema | No ACID transactions |
| SQLite | Simple, embedded | No concurrency |

## Consequences

- Need connection pooling for efficient resource usage
- Must plan for schema migrations
- Consider read replicas for scaling reads
- Backup strategy required (WAL archiving)
