---
id: c3-201
c3-version: 3
c3-seal: efd36720107cd7124a9e1eed481abf6143e11a9ac7a8a17cc291d029ab2585a8
title: Flow Pattern
type: component
category: foundation
parent: c3-2
goal: Typed business operations with DI, Zod parsing, and OTel tracing via @pumped-fn/lite
uses:
    - ref-pumped-fn
    - ref-structured-logging
---

# Flow Pattern

## Goal

Typed business operations with DI, Zod parsing, and OTel tracing via @pumped-fn/lite

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own Flow Pattern behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep Flow Pattern decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for Flow Pattern so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before Flow Pattern behavior is changed. | ref-pumped-fn |
| Inputs | Accept only the files, commands, data, or calls that belong to Flow Pattern ownership. | ref-pumped-fn |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-pumped-fn |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-pumped-fn |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks Flow Pattern to deliver its documented responsibility. | ref-pumped-fn |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-pumped-fn |
| Alternate paths | When a request falls outside Flow Pattern ownership, hand it to the parent or sibling component. | ref-pumped-fn |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-pumped-fn |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-pumped-fn | ref | Governs Flow Pattern behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| Flow Pattern input | IN | Callers must provide context that matches the component goal and parent fit. | c3-2 boundary | c3x lookup plus targeted tests or review. |
| Flow Pattern output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-2 boundary | c3x check and project test suite. |

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

- `@pumped-fn/lite` — `flow()`, `atom()`, `tags`, `service()`
- `zod` — input validation and coercion
- Execution context — tags for `currentUser`, `transaction`, `executionId`

## Anatomy

Every flow follows the same structure: namespace with Zod schema + types, then the `flow()` call.

```typescript
export namespace ApprovePr {
  export const schema = z.object({
    pr_id: z.coerce.number()
  }).transform(({ pr_id }) => ({ prId: pr_id }));

  export type Input = z.output<typeof schema>;
  export type Result = { success: true } | { success: false; reason: string };
}

export const approvePr = flow({
  deps: { prService, notificationService, logger, sync, executionId: tags.optional(executionIdTag) },
  parse: ApprovePr.schema.parse,
  factory: async (ctx, { prService, notificationService, logger, sync, executionId }): Promise<ApprovePr.Result> => {
    const currentUser = ctx.data.seekTag(currentUserTag);
    if (!currentUser) {
      return { success: false, reason: 'USER_NOT_FOUND' };
    }

    const result = await ctx.exec({ fn: prService.approve, params: [ctx.input.prId, currentUser.email] });

    if (executionId) {
      await sync.ack(executionId);
    }

    return { success: true };
  }
});
```

## Key Concepts

| Concept | How |
| --- | --- |
| Dependencies | deps object — atoms and services resolved from scope |
| Input validation | parse runs Zod schema; throws on failure (400 response) |
| Request-scoped state | ctx.data.seekTag() reads tags (currentUser, transaction) |
| Sub-operations | ctx.exec({ fn, params }) calls service methods within the same context |
| Sync notification | sync.ack(executionId) triggers real-time UI update |

## Conventions

| Rule | Why |
| --- | --- |
| Always define namespace with schema + types | Self-documenting API |
| Use z.coerce for form data | FormData values are always strings |
| Return { success, reason } | Consistent error handling across all flows |
| Access user via seekTag(currentUserTag) | User is request-scoped, not a dependency |
| Group flows into a const prFlows = { ... } as const map | Enables type-safe dispatch by name |

## Execution

Flows are invoked through `execContext.exec()` from server functions:

```typescript
// Pre-validated input (Zod in inputValidator)
await execContext.exec({ flow: createTeamFlow, input: data })

// Flow owns validation (rawInput)
await execContext.exec({ flow: approvePr, rawInput: data })
```

## Query Flows

Read-only flows skip user checks and return data directly:

```typescript
export const listPrs = flow({
  deps: { prQueries },
  factory: async (ctx, { prQueries }) => {
    return await ctx.exec({ fn: prQueries.listPr, params: [] });
  },
});
```
