---
id: c3-101
c3-version: 3
c3-seal: c9433ede2d5b96bd90ade758a293c9efb9fb46ab5885c7171e163c3ed763d2e5
title: State Management
type: component
category: foundation
parent: c3-1
goal: Reactive atoms via @pumped-fn/lite for api, user, stores, sync
uses:
    - ref-pumped-fn
    - ref-sync
---

# State Management

## Goal

Reactive atoms via @pumped-fn/lite for api, user, stores, sync

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own State Management behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep State Management decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for State Management so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before State Management behavior is changed. | ref-pumped-fn |
| Inputs | Accept only the files, commands, data, or calls that belong to State Management ownership. | ref-pumped-fn |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-pumped-fn |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-pumped-fn |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks State Management to deliver its documented responsibility. | ref-pumped-fn |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-pumped-fn |
| Alternate paths | When a request falls outside State Management ownership, hand it to the parent or sibling component. | ref-pumped-fn |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-pumped-fn |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-pumped-fn | ref | Governs State Management behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| State Management input | IN | Callers must provide context that matches the component goal and parent fit. | c3-1 boundary | c3x lookup plus targeted tests or review. |
| State Management output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-1 boundary | c3x check and project test suite. |

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

- `@pumped-fn/lite` -- atom, controller, preset, createScope, tag
- `@pumped-fn/react-lite` -- useAtom, ScopeProvider
- `@nats-io/nats-core` -- WebSocket transport for NATS sync atom

## Atoms

| Atom | Deps | Purpose |
| --- | --- | --- |
| executionTracker | -- | Promise-based wait/notify for correlating server actions to NATS acks |
| api | executionTracker | HTTP client: get, post, act (FormData), actJson, actForm |
| actions | api | Shorthand wrappers: act, actf, actd, actj |
| user | api | Current authenticated user with can(permission) helper |
| invoices | -- | Invoice[] store |
| prs | -- | PaymentRequest[] store |
| payments | -- | Payment[] store |
| approvalFlow | -- | Record<string, ApprovalFlow> via tag lookup |
| notifications | -- | NotificationState with add/mark-read/dismiss helpers |
| unreadCount | notifications | Derived: count of unread notifications |
| natsSync | user, prs, invoices, payments, notifications, executionTracker + NATS tags | WebSocket connection to NATS; subscribes to sync.broadcast (delta) and sync.user.<email> (notifications) |

## SSR Hydration

Route loader fetches data server-side, then creates a client scope with presets:

```typescript
const newScope = createScope({
  tags: [
    natsWsUrlTag(loaderData.natsWsUrl),
    natsCredentialsTag(loaderData.natsCredentials),
  ],
  presets: [
    preset(invoices, loaderData.invoices),
    preset(prs, loaderData.prs),
    preset(payments, loaderData.payments),
    preset(approvalFlow, loaderData.approvalFlow),
    preset(user, createUserWithCan(loaderData.user)),
  ]
})

// Wrap app
<ScopeProvider scope={scope}>
  <Outlet />
</ScopeProvider>
```

Scope is disposed on unmount (triggers NATS connection drain via `ctx.cleanup`).

## Usage

```typescript
// Read
const currentUser = useAtom(user)

// Mutate via controller
import { controller } from '@pumped-fn/lite'
const prsCtrl = useAtom(controller(prs))
prsCtrl.update(prev => [...prev, newPr])

// Actions
const { act } = useActions()
const result = await act("createPr", { name, amount })
if (result.executionId) await result.wait() // waits for NATS ack
```

## NATS Sync Wiring

`natsSync` connects via WebSocket with JWT auth, subscribes to two subjects:

- **`sync.broadcast`** -- delta + ack messages update prs/invoices/payments stores and notify `executionTracker`
- **`sync.user.<email>`** -- notification messages added to notifications store

Subject naming uses the default server prefix (`NATS_SUBJECT_PREFIX=sync`). The current frontend atom subscribes to `sync.*` directly; if prefix changes, frontend subscription wiring must be updated to match.

`executionTracker` correlation requires exact `executionId` key equality across HTTP result and NATS messages. Canonical type is string. `result.wait()` resolves when matching sync message arrives, with a 2s timeout fallback.
