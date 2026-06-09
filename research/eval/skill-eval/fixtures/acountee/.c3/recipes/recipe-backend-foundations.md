---
id: recipe-backend-foundations
c3-seal: 654a4f83fd266c15c8772a89c3801229949480c28c3576a641e5612af46376cb
title: Backend Foundations
type: recipe
goal: Trace how backend business logic is structured — from DI to flow execution to server function wiring.
sources:
    - c3-2
    - c3-201
    - c3-202
    - c3-203
    - ref-pumped-fn
    - ref-scope-controlled-config
    - ref-server-functions
    - ref-structured-logging
---

# Backend Foundations

## Goal

Trace how backend business logic is structured — from DI to flow execution to server function wiring.

## Narrative

The backend is built on three interlocking foundations: the **flow pattern**
(c3-201), **execution context** (c3-202), and **pumped-fn DI** (ref-pumped-fn).

Every business operation is a `flow()` with: typed deps (atoms resolved
from scope), Zod-parsed input, and OTel tracing. Flows define a namespace
with schema + types, making them self-documenting. Convention: always use
`z.coerce` for form data, return `{ success, reason }`.

Execution context carries request-scoped state via tags: `executionIdTag`
(unique request ID), `currentUserTag` (authenticated user + permissions),
`transactionTag` (active Drizzle transaction), `nameTag` (operation name).
Middleware creates and closes the context; flows read tags via
`ctx.data.seekTag()`.

Server functions (ref-server-functions) wire TanStack Start's
`createServerFn` to flows. Two patterns: `input` (Zod in inputValidator,
pre-validated) vs `rawInput` (flow validates). Never mix — using `input`
with a `rawInput` flow causes double validation.

Configuration is scope-controlled (ref-scope-controlled-config): closures
must not capture `process.env`; all config through scope tags. Logging
goes through scope-resolved `rootLogger` (ref-structured-logging), never
`console.log` in server code.

The DI system (`@pumped-fn/lite`) provides atoms for long-lived resources,
flows for operations, tags for contextual values, controllers for mutable
client stores, and presets for test overrides. Same library on server and
client (<17KB, zero deps).
