---
id: ref-server-functions
c3-seal: 301925682edd1dcd695811c847c7c01f0b21384d85720fbad36789de741188be
title: TanStack Start Server Functions
type: ref
goal: Consistent wiring between TanStack Start `createServerFn` and pumped-fn flows.
---

# TanStack Start Server Functions

## Goal

Consistent wiring between TanStack Start `createServerFn` and pumped-fn flows.

## Choice

- `createServerFn` with `inputValidator` + handler calling `execContext.exec()`
- Two valid patterns depending on where validation happens: `input` (pre-validated) vs `rawInput` (flow validates)

## Why

This keeps transport and domain concerns explicit: server functions remain thin wiring layers while flows stay the source of orchestration and business validation.

## input vs rawInput

`execContext.exec()` accepts **either** `input` or `rawInput` (mutually exclusive, enforced by types):

- **`input: data`** — Typed, pre-validated data. The flow's `parse` is skipped. Use when the server function's `inputValidator` already validates via Zod schema.
- **`rawInput: data`** — Untyped (`unknown`). The flow's `parse` function runs to validate/transform. Use when `inputValidator` is a lightweight shape function and the flow owns validation.

## Pattern 1: Zod in inputValidator + input

Use when `inputValidator` already Zod-parses the data.

```typescript
export const adminCreateTeam = createServerFn({ method: 'POST' })
  .middleware([transactionMiddleware])
  .inputValidator((input: teamFlows.CreateTeam.Input) => teamFlows.CreateTeam.schema.parse(input))
  .handler(async ({ data, context: { execContext } }) => {
    const result = await execContext.exec({ flow: teamFlows.createTeamFlow, input: data })
    return { ...result, executionId }
  })
```

Validation happens at the transport boundary. `data` is already the correct type.

## Pattern 2: Shape function + rawInput

Use when the flow owns validation.

```typescript
const prIdInput = (data: { pr_id: number }) => data

export const approvePr = createServerFn({ method: 'POST' })
  .middleware([transactionMiddleware])
  .inputValidator(prIdInput)
  .handler(async ({ data, context: { execContext } }) => {
    const result = await execContext.exec({ flow: prFlows.approvePr, rawInput: data })
    return { ...result, executionId }
  })
```

`inputValidator` provides TypeScript shape only. The flow's `parse` handles runtime validation.

## Pattern 2b: FormData + rawInput

For file uploads via `formDataMiddleware`, there is no `inputValidator` — the raw form data goes directly to the flow:

```typescript
export const createPr = createServerFn({ method: 'POST' })
  .middleware([formDataMiddleware])
  .handler(async ({ context: { execContext, rawForm } }) => {
    const result = await execContext.exec({ flow: prFlows.createPr, rawInput: rawForm })
    return { ...result, executionId }
  })
```

## Not This

```typescript
// WRONG: mixing input with a flow that expects rawInput (double validation or type mismatch)
execContext.exec({ flow: prFlows.approvePr, input: data })

// WRONG: using rawInput when inputValidator already Zod-parsed (skips nothing, but flow.parse may not exist)
execContext.exec({ flow: teamFlows.createTeamFlow, rawInput: data })
```

## Cited By

- c3-201 (Flow Pattern)
- c3-203 (Middleware Stack)
