---
id: c3-204
c3-version: 3
c3-seal: 709085e60de6b3e8e1cf45408b0228b1fc8ab9b5fc8bf1ddd228972b20a10e37
title: Drizzle ORM
type: component
category: foundation
parent: c3-2
goal: Type-safe PostgreSQL access via Drizzle ORM with auto-migration system
uses:
    - ref-pumped-fn
    - ref-query-services
    - ref-scope-controlled-config
    - ref-structured-logging
---

# Drizzle ORM

## Goal

Type-safe PostgreSQL access via Drizzle ORM with auto-migration system

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own Drizzle ORM behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep Drizzle ORM decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for Drizzle ORM so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before Drizzle ORM behavior is changed. | ref-pumped-fn |
| Inputs | Accept only the files, commands, data, or calls that belong to Drizzle ORM ownership. | ref-pumped-fn |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-pumped-fn |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-pumped-fn |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks Drizzle ORM to deliver its documented responsibility. | ref-pumped-fn |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-pumped-fn |
| Alternate paths | When a request falls outside Drizzle ORM ownership, hand it to the parent or sibling component. | ref-pumped-fn |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-pumped-fn |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-pumped-fn | ref | Governs Drizzle ORM behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| Drizzle ORM input | IN | Callers must provide context that matches the component goal and parent fit. | c3-2 boundary | c3x lookup plus targeted tests or review. |
| Drizzle ORM output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-2 boundary | c3x check and project test suite. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |
| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| Code, docs, tests, prompts | Goal, Governance, Contract, and Change Safety sections. | Names and framework shape may vary; behavior and boundaries may not. | c3x check, c3x verify, and relevant tests. |

## Architecture Details

## Dependencies

- `drizzle-orm` + `postgres` (postgres-js driver)
- `@pumped-fn/lite` — atoms for connection, transaction, migration
- `PGURI` or `ZB_POSTGRES_5432` env var — PostgreSQL connection string (validated with Zod). Zerobased auto-injects `ZB_POSTGRES_5432` with a Unix socket URI.

## Connection

`pgConfig` atom reads `PGURI` (falls back to `ZB_POSTGRES_5432`) and passes it through `parseConnectionString()`. Socket-style URIs (`postgresql://user@/db?host=/path/to/socket`) can't be parsed by Node's URL constructor, so the function detects the `?host=` suffix and converts to a `postgres()` options object with `host`, `user`, `pass`, `database`.

`drizzleDb` atom creates the Drizzle instance with schema and a custom pino-based query logger that redacts binary data.

```typescript
export const drizzleDb = atom({
  deps: { pgConfig, logger },
  factory: async (ctx, { pgConfig, logger }) => {
    const client = postgres(pgConfig.connectionString);
    return drizzle({ client, schema, logger: queryLogger });
  },
});
```

## Schema

All tables defined in `schema.ts` using Drizzle's `pgTable()`. Key tables: `pr`, `users`, `invoices`, `payments`, `approval`, `approvalSteps`, `approvalRecords`, `audit`, `teams`, `roles`, `notificationLog`, `notificationPreferences`, `files`.

JSONB columns need `$type<>()` for proper TypeScript typing:

```typescript
permissions: jsonb().$type<string[]>().default([]).notNull(),
payload: jsonb().$type<object>().notNull(),
```

## Transactions

Two transaction atoms in `transaction.ts`:

| Atom | Purpose |
| --- | --- |
| executeInDrizzleTransaction | Wraps callback in transaction, sets app.current_user (base64-encoded email) for PostgreSQL audit triggers |
| unsafeExecuteInDrizzleTransaction | Transaction without user context — for system operations |

The transaction is stored on the execution context via `transactionTag`. Query services read it to participate in the same transaction.

## Auto-Migration

`runMigrations` atom runs SQL files from `dbs/migrations/*.sql` on app startup:

1. Creates `pgmigrations` tracking table if not exists
2. Reads already-run migration names from DB
3. Loads `*.sql` files from disk, sorted by filename prefix
4. Applies pending migrations sequentially
5. Records each in `pgmigrations` on success
6. On failure: throws error, app exits (no partial apply)

Migration files use `NNNN_description.sql` naming. Use idempotent patterns (`IF NOT EXISTS`, `ON CONFLICT DO NOTHING`).

```bash
# Generate migration from schema changes
pnpm run db:generate

# Push schema directly (dev only)
pnpm run db:push

# Open Drizzle Studio
pnpm run db:studio
```
