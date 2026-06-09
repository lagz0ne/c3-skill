---
id: ref-structured-logging
c3-seal: af9aba1be913b7c3c16a4f43a6ca18ff4830af224e3396596ba9c55a87c77cab
title: Structured Logging
type: ref
goal: Consistent logging through scope-resolved rootLogger, never console.log in server code.
---

# Structured Logging

## Goal

Consistent logging through scope-resolved rootLogger, never console.log in server code.

## Choice

- Resolve `rootLogger` from DI scope; create child loggers with component name
- Structured data-first format: `log.info({ key: value }, 'message')`
- Pre-scope code uses `process.stderr.write()`; client code uses `console.error()`

## When

- All server-side code that needs to log
- Pre-scope code (env validation) uses `process.stderr.write()`

## Why

- Structured logs enable filtering, aggregation, tracing
- Scope resolution ensures consistent config (logtail, dev mode)
- Child loggers provide automatic context (component name)

## Conventions

| Rule | Example |
| --- | --- |
| Resolve logger from scope | const logger = await scope.resolve(rootLogger) |
| Create child with name | const log = logger.child({ name: 'server' }) |
| Structured data first, message last | log.info({ key: value }, 'message') |
| Pre-scope uses stderr | process.stderr.write('msg\n') |

## Pattern

```typescript
async function initialize() {
  const logger = await scope.resolve(rootLogger)
  const log = logger.child({ name: 'server' })

  log.info({ version, built }, 'starting server')
  log.error({ error: err }, 'operation failed')
}
```

## Not This

```typescript
// WRONG: pino.transport() uses thread-stream which doesn't bundle with Vite/Nitro
const transportStream = transport({ targets: [...] })

// CORRECT: sync destination works everywhere
pino.destination({ fd: 1, sync: true })
```

Production logs JSON to stdout — container runtime collects logs.

## Exceptions

| Context | Method | Reason |
| --- | --- | --- |
| Pre-scope (env validation) | process.stderr.write() | Scope not available |
| Client-side (screens) | console.error() | Browser debugging |
| Test files | console.log() | Test output visibility |

## Cited By

- c3-2-api (server initialization, migrations)
