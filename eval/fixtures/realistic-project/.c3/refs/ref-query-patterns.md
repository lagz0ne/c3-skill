---
id: ref-query-patterns
title: Query Patterns
---

# Query Patterns

## Goal

Establish conventions for database queries using Drizzle ORM wrapped in service atoms with typed input/output and transaction support.

## Conventions

| Rule | Why |
|------|-----|
| Wrap queries in `service({deps, factory})` | Enables dependency injection and testing |
| Use Drizzle query builder | Type-safe SQL generation |
| Accept execContext as first parameter | Access transaction context when needed |
| Use generated types from types.ts | Consistent row types across queries |
| Support transactionTag for atomic operations | Multi-table operations in single transaction |
| Return null for not found (not throw) | Caller decides how to handle missing data |
| Use indexed columns in WHERE clauses | Performance for large datasets |

## Testing

| Convention | How to Test |
|------------|-------------|
| Type safety | Use wrong column name, verify compile error |
| Transaction support | Run in tx, rollback, verify no changes |
| Null returns | Query non-existent ID, verify null (not throw) |
| Index usage | EXPLAIN query, verify index scan |
| Generated types | Verify ListPrRow matches actual columns |

## References

- `apps/start/src/server/dbs/queries/` - All query modules
