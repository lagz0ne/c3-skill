---
id: ref-audit-trail
c3-seal: 87f6700bbb524b1dd640fc3a0cef866bfb7e199af8a5964cbc773069c79734c5
title: Audit Trail Pattern
type: ref
goal: Every mutation to persistent records is captured with before/after snapshots, enabling full change reconstruction for compliance and debugging.
---

# Audit Trail Pattern

## Goal

Every mutation to persistent records is captured with before/after snapshots, enabling full change reconstruction for compliance and debugging.

## Choice

Use a hybrid audit strategy: explicit audit writes in flow code for semantic domain actions, and database triggers for low-level table mutation capture.

## Why

This balances clarity and completeness: flows provide business intent, while triggers guarantee no silent data change escapes capture.

Two audit mechanisms exist:

- **Explicit calls** — admin flows call `auditQueries.createAuditEntry` with semantic actions (`create`, `update`, `delete`, `assign_role`, `revoke_role`)
- **DB triggers** — `log_change()` trigger fires on data tables (`invoices`, `pr`, `invoice_services`) with raw SQL operation names (`INSERT`, `UPDATE`, `DELETE`)

Both write to the same `audit` table.

## Data Model

```typescript
// audit table schema
{
  id:            serial,
  action:        text,        // 'create' | 'update' | 'delete' | 'assign_role' | 'revoke_role' | 'INSERT' | 'UPDATE' | 'DELETE'
  table_name:    text,        // 'teams' | 'roles' | 'user_roles' | 'approval_flows' | 'invoices' | 'pr' | ...
  record_id:     integer,     // nullable — FK to the mutated record
  record_before: jsonb,       // full snapshot before mutation (null on create/INSERT)
  record_after:  jsonb,       // full snapshot after mutation (null on delete/DELETE)
  triggered_at:  timestamp,   // default CURRENT_TIMESTAMP
  triggered_by:  text,        // user email or 'system'
  user_agent:    text,        // nullable
  ip_address:    inet,        // nullable
  checksum:      text,        // md5 integrity hash (triggers set this automatically)
  metadata:      jsonb,       // default {} — extensible context
}

// Indexes: action, (table_name, record_id), triggered_at DESC, triggered_by
```

## When to Audit

| Mechanism | Tables | Trigger |
| --- | --- | --- |
| Explicit call | teams, roles, user_roles, approval_flows | Flow code calls createAuditEntry after mutation |
| DB trigger | invoices, pr, invoice_services | log_change() fires on INSERT/UPDATE/DELETE |

Rule: if a table has a `log_change()` trigger, do NOT also call `createAuditEntry` — it would duplicate entries.

## Wiring

### Explicit audit in flows

Flows declare `auditQueries` as a dependency. The query service picks up the active transaction via `seekTag(transactionTag)`, so audit writes are atomic with the mutation.

```typescript
// Golden example: creating a team with audit
export const createTeamFlow = flow({
  deps: { teamQueries, auditQueries, logger },
  parse: CreateTeam.schema.parse,
  factory: async (ctx, { teamQueries, auditQueries, logger }) => {
    const currentUser = ctx.data.seekTag(currentUserTag);

    // 1. Mutate
    const team = await ctx.exec({
      fn: teamQueries.createTeam,
      params: [{ name: ctx.input.name, created_by: currentUser.email }],
    });

    // 2. Audit — same transaction, atomic
    await ctx.exec({
      fn: auditQueries.createAuditEntry,
      params: [{
        action: "create",
        table_name: "teams",
        record_id: team.id,
        record_after: team,
        triggered_by: currentUser.email,
      }],
    });

    return { success: true, team };
  },
});
```

For updates, capture before state first:

```typescript
const existingRole = await ctx.exec({ fn: rbacQueries.getRoleById, params: [{ id: ctx.input.id }] });
// ... perform update ...
await ctx.exec({
  fn: auditQueries.createAuditEntry,
  params: [{
    action: "update",
    table_name: "roles",
    record_id: role.id,
    record_before: existingRole,
    record_after: role,
    triggered_by: currentUser.email,
  }],
});
```

### DB trigger audit

The `log_change()` function is a PostgreSQL trigger that automatically captures mutations. It reads the actor from a session variable:

```sql
-- Transaction sets actor context before any writes
SELECT set_config('app.current_user', <base64-encoded-email>, true);

-- Trigger reads it
current_user_email := current_setting('app.current_user', true);
```

This wiring happens in `executeInDrizzleTransaction` — it base64-encodes the user email and sets `app.current_user` before the callback runs. All writes within that transaction are attributed to the correct user.

### Query service wiring

`auditQueries` is a `service()` that depends on `drizzleDb`. Each method reads the transaction from context:

```typescript
const tx = execCtx.data.seekTag(transactionTag);
const executor = tx ?? db; // falls back to direct DB if no transaction
```

### Related table queries

Audit history supports fetching related records (e.g., invoice + its services):

```typescript
const RELATED_TABLES: Record<string, { table: string; foreignKey: string }[]> = {
  invoices: [{ table: 'invoice_services', foreignKey: 'invoice_id' }],
};
```

This joins audit entries where `record_after->>foreignKey` matches the parent record ID.

## Anti-Patterns

- **Auditing trigger-covered tables explicitly** — invoices/PR already have `log_change()`. Adding `createAuditEntry` creates duplicate entries.
- **Auditing outside a transaction** — the audit write must be atomic with the mutation. If the mutation rolls back, the audit entry should too.
- **Missing before snapshot on update/delete** — fetch the record before mutating. Without `record_before`, change reconstruction is impossible.
- **Using raw SQL operation names in explicit calls** — explicit calls use semantic names (`create`, `update`, `delete`), not SQL names (`INSERT`, `UPDATE`, `DELETE`). SQL names come from triggers only.

## Cited By

- `c3-208` (Audit Flows)
- `c3-210` (Admin Flows)
