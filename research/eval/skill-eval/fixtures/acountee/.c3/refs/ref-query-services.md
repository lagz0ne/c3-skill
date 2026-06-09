---
id: ref-query-services
c3-seal: e1cd41701253915e2228d2578175db8b3cba8f444e536caaefb07d6a08475976
title: Query Services Pattern
type: ref
goal: Encapsulate database access in service objects that respect transaction boundaries and enable tracing.
---

# Query Services Pattern

## Goal

Encapsulate database access in service objects that respect transaction boundaries and enable tracing.

## Choice

- Use `service()` from @pumped-fn/lite to define query objects with DI lifecycle
- Always check `transactionTag` before executing queries to respect open transactions
- Wrap every database call in `execCtx.exec()` for OpenTelemetry tracing

## Why

- DI-managed services ensure consistent instantiation and testability
- Transaction tag lookup allows queries to participate in cross-flow transactions without explicit passing
- Tracing wrapping gives per-query observability with zero caller burden

## Convention

| Rule | Why |
| --- | --- |
| Use service() from @pumped-fn/lite | DI + lifecycle |
| Check transactionTag first | Respect transaction boundaries |
| Wrap in execCtx.exec() | Enable OTel tracing |

## Pattern

```typescript
export const prQueries = service({
  deps: { drizzleDb },
  factory: (ctx, { drizzleDb: db }) => ({
    getPr: async (execCtx, args: IdArg) => {
      const tx = execCtx.data.seekTag(transactionTag);
      const executor = tx ?? db;

      return execCtx.exec({
        name: 'db:select:pr',
        fn: async () => executor.select().from(pr).where(eq(pr.id, args.id)),
        params: []
      });
    },
  })
});
```

## Services

| Service | Purpose |
| --- | --- |
| prQueries | Payment request CRUD |
| invoiceQueries | Invoice CRUD |
| paymentQueries | Payment CRUD |
| approvalQueries | Approval chain |
| userQueries | User lookup |
| rbacQueries | RBAC |

## Cited By

- c3-2-api (Query Layer)
