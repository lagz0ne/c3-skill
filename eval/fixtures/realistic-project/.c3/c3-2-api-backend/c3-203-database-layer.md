---
id: c3-203
c3-version: 3
title: Database Layer
type: component
category: foundation
parent: c3-2
summary: Drizzle ORM setup with PostgreSQL, schema definitions, and transaction support
---

# Database Layer

Provides database connectivity using Drizzle ORM with PostgreSQL, schema definitions in schema.ts, and transaction support via transactionTag for atomic operations.

## Contract

| Provides | Expects |
|----------|---------|
| drizzleDb atom | PGURI environment variable |
| Schema definitions (tables) | Valid PostgreSQL connection |
| db.transaction() | Callback with tx object |
| transactionTag | Set within transaction scope |

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Connection failure | Atom resolution fails |
| Transaction rollback | Changes discarded |
| Connection pool exhausted | Queries wait/timeout |
| Invalid SQL | Drizzle compile error |

## Testing

| Scenario | Verifies |
|----------|----------|
| Connection | scope.resolve(drizzleDb) succeeds |
| Transaction isolation | Changes visible only after commit |
| Rollback | Error in callback rolls back |
| Schema types | Table column types match TypeScript |

## References

- `apps/start/src/server/dbs/` - Database layer (schema, queries, transactions)
