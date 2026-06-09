---
id: ref-scope-controlled-config
c3-seal: bbbb2d5cced6d61b2eb8f13f3788d3f6ea57293a65383b9fc176cf2e7a3d4efd
title: Scope-Controlled Configuration
type: ref
goal: Closures must not capture process.env; all config access through scope tags.
---

# Scope-Controlled Configuration

## Goal

Closures must not capture process.env; all config access through scope tags.

## Choice

- All configuration accessed via scope tags resolved at runtime, never via closure-captured `process.env`
- A config atom aggregates related tags into a single resolvable unit
- Env vars are only read once at scope creation time and injected as tags

## When

- Any code that runs after scope creation
- Especially async functions like `initialize()` that might look like they need env access

## Why

- Testable: tests provide their own tags
- No hidden dependencies in closures
- Side effects controlled through scope
- Clear dependency graph

## Conventions

| Rule | Example |
| --- | --- |
| Define tags for config | natsConfig.subjectPrefix |
| Create config atom | serverConfig collects all tags |
| Resolve config from scope | await scope.resolve(serverConfig) |
| Never capture env in closure | No env.NATS_* inside async functions |

## Pattern

```typescript
// BAD - captures env in closure
async function initialize() {
  log.info({ prefix: env.NATS_SUBJECT_PREFIX })  // closure capture
}

// GOOD - resolve config from scope
const serverConfig = atom({
  deps: {
    natsSubjectPrefix: tags.required(natsConfig.subjectPrefix),
    smtp: tags.required(emailConfig.smtp),
  },
  factory: (_ctx, deps) => deps,
})

async function initialize() {
  const config = await scope.resolve(serverConfig)
  log.info({ prefix: config.natsSubjectPrefix })  // from scope
}
```

## Scope Creation

Config is injected via tags at scope creation time only:

```typescript
const scope = createScope({
  tags: [
    natsConfig.subjectPrefix(env.NATS_SUBJECT_PREFIX),
    emailConfig.smtp({ host: env.SMTP_HOST, ... }),
  ]
})
```

## Cited By

- c3-2-api (server.tsx, serverConfig atom)
