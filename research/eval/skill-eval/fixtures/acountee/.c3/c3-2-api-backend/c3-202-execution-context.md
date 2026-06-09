---
id: c3-202
c3-version: 3
c3-seal: 2500d661812e6a5aa55c45604ac0a0f4f073d9686336f957d824ebb5422cbf89
title: Execution Context
type: component
category: foundation
parent: c3-2
goal: Request-scoped context with tags for currentUser, transaction, executionId
uses:
    - ref-pumped-fn
    - ref-scope-controlled-config
    - ref-structured-logging
---

# Execution Context

## Goal

Request-scoped context with tags for currentUser, transaction, executionId

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own Execution Context behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep Execution Context decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for Execution Context so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before Execution Context behavior is changed. | ref-pumped-fn |
| Inputs | Accept only the files, commands, data, or calls that belong to Execution Context ownership. | ref-pumped-fn |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-pumped-fn |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-pumped-fn |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks Execution Context to deliver its documented responsibility. | ref-pumped-fn |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-pumped-fn |
| Alternate paths | When a request falls outside Execution Context ownership, hand it to the parent or sibling component. | ref-pumped-fn |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-pumped-fn |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-pumped-fn | ref | Governs Execution Context behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| Execution Context input | IN | Callers must provide context that matches the component goal and parent fit. | c3-2 boundary | c3x lookup plus targeted tests or review. |
| Execution Context output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-2 boundary | c3x check and project test suite. |

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

- `@pumped-fn/lite` — `tag()`, `scope.createContext()`
- Middleware stack — creates context and sets tags

## Tags

Defined in `tags.ts`:

| Tag | Type | Set By | Purpose |
| --- | --- | --- | --- |
| executionIdTag | string | executionContextMiddleware | Unique request ID (crypto.randomUUID()) |
| currentUserTag | UserActor | getCurrentUserMiddleware | Authenticated user with permissions |
| transactionTag | DrizzleTransaction | Transaction middleware | Active DB transaction |
| nameTag | string | execContext.exec() | Operation name for tracing |

## Tag API

```typescript
import { tag } from '@pumped-fn/lite'

export const executionIdTag = tag<string>({ label: 'executionId' })
export const currentUserTag = tag<UserActor>({ label: 'currentUser' })
export const transactionTag = tag<DrizzleTransaction>({ label: 'transaction' })
export const nameTag = tag<string>({ label: 'name' })
```

- `ctx.data.setTag(tag, value)` — write (middleware)
- `ctx.data.seekTag(tag)` — read, returns `undefined` if not set (flows)

## Lifecycle

1. Middleware creates context: `scope.createContext({})`
2. Sets `executionIdTag` and `currentUserTag`
3. Transaction middleware sets `transactionTag` within `db.transaction()`
4. Flow reads tags via `ctx.data.seekTag()`
5. Middleware closes context in `finally` block

## UserActor

```typescript
interface UserActor {
  email: string;
  team: string;
  teamCapabilities: string[];
  name?: string | null;
  displayName?: string | null;
  avatar?: string | null;
  permissions: string[];
  can(permission: string): boolean;
  asserts(permission: string): void;
  setPermissions(permissions: string[]): void;
}
```

## Configuration Tags

Beyond request-scoped tags, `tags.ts` also defines infrastructure configuration:

| Group | Tags | Purpose |
| --- | --- | --- |
| natsConfig | wsUrl, accountSeed, subjectPrefix | NATS connection and JWT signing |
| logConfig | logtail, isDev | Logging backend |
| emailConfig | smtp | SMTP transport settings |
| notificationConfig | requiredChannels | Which notification channels must initialize |
| slackConfig | botToken, signingSecret | Slack bot connection (optional, undefined if not configured) |
| buildConfig | version, time | Build metadata |
