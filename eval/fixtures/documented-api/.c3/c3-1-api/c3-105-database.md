# c3-105: Database Layer

## Purpose

Database abstraction for SQL operations.

## Location

`src/db.ts`

## Responsibilities

- Provide query interface
- Handle connection pooling (in production)
- Abstract database-specific syntax

## Interface

```typescript
db.query(sql: string, params?: unknown[]): Promise<any[]>
```

## Notes

Current implementation is a stub. Production would use:
- MySQL, PostgreSQL, or SQLite
- Connection pooling
- Prepared statements
