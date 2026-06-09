---
id: ref-pumped-fn
c3-seal: a0634b86c7c7f21829b0a15f3bb5aed9591061b4505771f746abc78116948568
title: pumped-fn DI System
type: ref
goal: Consistent use of `@pumped-fn/lite` for dependency injection, lifecycle management, and reactive state.
---

# pumped-fn DI System

## Goal

Consistent use of `@pumped-fn/lite` for dependency injection, lifecycle management, and reactive state.

## Choice

- `@pumped-fn/lite` as the sole DI/effect system for both server and client
- Atoms for scope-level singletons, flows for business operations, tags for contextual values, resources for execution-scoped deps

## Why

- Zero-dependency, <17KB — fits in both server and browser bundles
- Unified model across server resources, business logic, and client state
- Lifecycle management (cleanup, invalidation, GC) prevents resource leaks

## How

Full API reference: `node_modules/@pumped-fn/lite/dist/index.d.mts`

### Primitives

| Primitive | Lifetime | Best for |
| --- | --- | --- |
| atom() | Scope (singleton, cached) | DB connections, NATS, loggers, client stores |
| flow() | Per call | Business operations (approve, create, import) |
| tag() | Set on scope or execution context | Request-scoped values (currentUser, transaction), config injection |
| resource() | Execution chain (fresh per chain, seek-up on nested) | Per-request loggers, transaction-aware executors |
| service() | Scope (singleton, methods receive execCtx) | Query services, domain services |
| controller() | Tied to parent atom | Writable handles for reactive stores; watch: true for reactive derivation |
| preset() | Scope creation | SSR hydration, test doubles (atoms and flows) |

### Golden Example: Server atom with deps and cleanup

```typescript
// resources/slackBot.ts — scope-level singleton
export const slackBot = atom({
  deps: {
    config: tags.required(slackConfigTag),  // tag dependency
    rootLogger,                              // atom dependency
  },
  factory: (ctx, { config, rootLogger }) => {
    if (!config) return null                 // graceful skip when unconfigured

    const logger = rootLogger.child({ name: 'slackBot' })
    const bot = new Chat({ adapters: { slack }, state: createMemoryState() })

    ctx.cleanup(() => bot.shutdown())        // ALWAYS register cleanup
    return bot
  },
})
```

### Golden Example: Flow with Zod parse and sync ack

```typescript
// flows/pr.ts — business operation
export namespace ApprovePr {
  export const schema = z.object({ pr_id: z.coerce.number() })
    .transform(({ pr_id }) => ({ prId: pr_id }))
  export type Input = z.output<typeof schema>
  export type Result = { success: true } | { success: false; reason: string }
}

export const approvePr = flow({
  deps: { prService, sync, executionId: tags.optional(executionIdTag) },
  parse: ApprovePr.schema.parse,
  factory: async (ctx, { prService, sync, executionId }): Promise<ApprovePr.Result> => {
    const currentUser = ctx.data.seekTag(currentUserTag)
    if (!currentUser) return { success: false, reason: 'USER_NOT_FOUND' }

    const result = await ctx.exec({ fn: prService.approve, params: [ctx.input.prId, currentUser.email] })
    if (!result.success) return { success: false, reason: result.reason ?? 'APPROVAL_FAILED' }

    if (executionId) await sync.ack(executionId)
    return { success: true }
  },
})
```

### Golden Example: Query service with transaction tag

```typescript
// dbs/queries/slack.ts — scope singleton, methods receive execCtx
export const slackQueries = service({
  deps: { drizzleDb },
  factory: (_ctx, { drizzleDb: db }) => ({
    getSlackUserByEmail: async (execCtx, args: { userEmail: string }) => {
      const tx = execCtx.data.seekTag(transactionTag)  // respect active transaction
      const executor = tx ?? db                         // fallback to raw db
      return execCtx.exec({ name: 'db:select:slackUserByEmail', fn: async () => {
        const [result] = await executor.select().from(slackUserLinks)
          .where(eq(slackUserLinks.userEmail, args.userEmail))
        return result ?? null
      }, params: [] })
    },
  }),
})
```

### Golden Example: Execution-scoped resource

```typescript
// resources/requestLogger.ts — fresh per execution chain
const requestLogger = resource({
  deps: { logService: logServiceAtom },
  factory: (ctx, { logService }) => {
    const requestId = ctx.data.seekTag(executionIdTag)
    const logger = logService.child({ requestId })
    ctx.onClose(() => logger.flush())
    return logger
  },
})

// Used as a flow dep — resolved per execution, not cached in scope
export const approvePr = flow({
  deps: { requestLogger, prService },
  factory: async (ctx, { requestLogger, prService }) => {
    requestLogger.info('approving PR')
    // ...
  },
})
```

A `resource()` can also encapsulate the transaction-aware executor pattern:

```typescript
const txExecutor = resource({
  deps: { drizzleDb },
  factory: (ctx, { drizzleDb: db }) => {
    const tx = ctx.data.seekTag(transactionTag)
    return tx ?? db
  },
})
```

### Golden Example: Client-side reactive state with controller

```typescript
// lib/pumped/atoms/stores.ts — reactive stores
export const prs = atom({ factory: (): PaymentRequest[] => [] })

// lib/pumped/atoms/natsSync.ts — controller for mutation
export const natsSync = atom({
  deps: {
    prs,
    prsCtrl: controller(prs),              // writable handle for prs atom
    invoicesCtrl: controller(invoices),
    tracker: executionTracker,
    natsWsUrl: tags.required(natsWsUrlTag),
  },
  factory: async (ctx, { prsCtrl, invoicesCtrl, tracker, ... }) => {
    // On delta message from NATS:
    prsCtrl.update(prev => applyDelta(prev, syncMsg.changes.prs!))
    if (syncMsg.executionId) tracker.notify(syncMsg.executionId)
  },
})
```

Watched controllers auto-re-run the parent factory on value change:

```typescript
const derived = atom({
  deps: { src: controller(srcAtom, { resolve: true, watch: true }) },
  factory: (_, { src }) => transform(src.get()),  // re-runs automatically
})

// With custom equality (skip re-run if structurally equal):
const derived = atom({
  deps: { src: controller(srcAtom, { resolve: true, watch: true, eq: (a, b) => a.id === b.id }) },
  factory: (_, { src }) => src.get().name,
})
```

### Golden Example: SSR hydration with preset

```typescript
// routes/_authed.tsx — hydrate client scope from loader data
<ScopeProvider presets={[
  preset(invoices, loaderData.invoices),
  preset(prs, loaderData.prs),
  preset(payments, loaderData.payments),
  preset(user, createUserWithCan({ ... })),
]} tags={[
  natsWsUrlTag(loaderData.natsWsUrl),
  natsCredentialsTag(loaderData.natsCredentials),
]}>
```

Flows can also be preset — useful for test doubles:

```typescript
const scope = createScope({
  presets: [
    preset(approvePr, async (ctx) => ({ success: true })),
    preset(approvePr, mockApprovePr),  // or swap with another flow
  ],
})
```

### Scope options

```typescript
const scope = createScope({
  gc: { enabled: true, graceMs: 3000 },  // automatic garbage collection
  tags: [...],
  presets: [...],
  extensions: [...],
})
```

Atoms can opt out of GC:

```typescript
const criticalConnection = atom({
  keepAlive: true,  // never GC'd, even if no dependents
  factory: () => createPersistentConnection(),
})
```

### Utility types

```typescript
type Config = Lite.Utils.AtomValue<typeof configAtom>
type Result = Lite.Utils.FlowOutput<typeof processFlow>
type Input  = Lite.Utils.FlowInput<typeof processFlow>
type UserId = Lite.Utils.TagValue<typeof userIdTag>
type Deps   = Lite.Utils.DepsOf<typeof myAtom>
```

## Conventions

| Primitive | Server | Client |
| --- | --- | --- |
| atom() | Resources (db, nats, sync) | Reactive stores (prs, invoices, user) |
| flow() | Business operations (approve, create) | — |
| tag() | Request-scoped (currentUser, transaction) | Config injection (natsWsUrl) |
| resource() | Execution-scoped (per-request loggers, tx wrappers) | — |
| service() | Query/domain services (scope singleton, methods take execCtx) | — |
| controller() | — | Mutable store handles; { watch: true } for reactive derivation |
| preset() | Test doubles (atoms + flows) | SSR hydration |

## Anti-patterns

- Capturing `process.env` in atom/flow factories — use tags instead (see ref-scope-controlled-config)
- Creating atoms inside flows — atoms are scope-level; use `resource()` for execution-scoped deps
- Manual invalidation wiring (`ctx.scope.on('resolved', ...) + ctx.invalidate()`) — use `controller({ watch: true })` instead
- Skipping `ctx.cleanup()` on atoms that hold connections or subscriptions

## Cited By

- c3-101 (State Management)
- c3-201 (Flow Pattern)
- c3-202 (Execution Context)
