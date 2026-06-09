---
id: c3-203
c3-version: 3
c3-seal: 35774f7d76531700a94b5dbfd167daef918bc7c8e1dece1b6db6ac25f9330097
title: Middleware Stack
type: component
category: foundation
parent: c3-2
goal: executionContextMiddleware + getCurrentUserMiddleware for request handling
uses:
    - ref-pumped-fn
    - ref-scope-controlled-config
    - ref-structured-logging
---

# Middleware Stack

## Goal

executionContextMiddleware + getCurrentUserMiddleware for request handling

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own Middleware Stack behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep Middleware Stack decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for Middleware Stack so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before Middleware Stack behavior is changed. | ref-pumped-fn |
| Inputs | Accept only the files, commands, data, or calls that belong to Middleware Stack ownership. | ref-pumped-fn |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-pumped-fn |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-pumped-fn |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks Middleware Stack to deliver its documented responsibility. | ref-pumped-fn |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-pumped-fn |
| Alternate paths | When a request falls outside Middleware Stack ownership, hand it to the parent or sibling component. | ref-pumped-fn |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-pumped-fn |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-pumped-fn | ref | Governs Middleware Stack behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| Middleware Stack input | IN | Callers must provide context that matches the component goal and parent fit. | c3-2 boundary | c3x lookup plus targeted tests or review. |
| Middleware Stack output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-2 boundary | c3x check and project test suite. |

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

- `@tanstack/react-start` — `createMiddleware()`
- `@pumped-fn/lite` — scope, context, tags
- `cookie` — `parseCookie()` for reading auth cookie

## Bootstrap (server.tsx)

Before the middleware stack runs, `server.tsx` resolves zerobased-injected env vars:

- `ZB_NATS_80` (HTTP URL) → `NATS_WS_URL` (WebSocket URL) via `http` → `ws` replacement
- `ZB_POSTGRES_5432` → used by `pgConfig` in c3-204

These are set before Zod env validation so downstream code sees standard env var names.

## Middleware Chain

```
Request → executionContextMiddleware → getCurrentUserMiddleware → Handler → Response
```

### executionContextMiddleware

Creates a `@pumped-fn/lite` execution context from the scope and sets a unique execution ID. Closes the context in `finally` regardless of outcome.

```typescript
export const executionContextMiddleware = createMiddleware().server(
  async ({ next, context: { scope } }) => {
    const execContext = scope.createContext({});
    execContext.data.setTag(executionIdTag, createExecutionId());
    try {
      return await next({ context: { execContext } });
    } finally {
      await execContext.close();
    }
  }
);
```

### getCurrentUserMiddleware

Chains after `executionContextMiddleware`. Reads the `user` cookie, looks up the user in the database, fetches team capabilities, and sets `currentUserTag` on the execution context.

```typescript
export const getCurrentUserMiddleware = createMiddleware()
  .middleware([executionContextMiddleware])
  .server(async ({ next, context: { scope, execContext }, request }) => {
    const cookies = parseCookie(request.headers.get("cookie") ?? "");
    const userEmail = cookies["user"]?.toLowerCase();

    if (userEmail) {
      const dbUser = await userQueriesService.getUser(execContext, { email: userEmail });
      if (dbUser) {
        const teamCapabilities = dbUser.team
          ? (await teamQueriesService.getTeamByName(execContext, { name: dbUser.team }))?.capabilities ?? []
          : [];

        execContext.data.setTag(currentUserTag, {
          email: userEmail,
          permissions,
          team: dbUser.team || "default",
          teamCapabilities,
          can: (p) => permissions.includes(p),
          asserts: (p) => { if (!permissions.includes(p)) throw new Error(`Permission denied: ${p}`) },
          setPermissions: () => {},
        });
      }
    }
    return await next({});
  });
```

## Auth Logic

| Cookie State | Result |
| --- | --- |
| No cookie | currentUserTag not set — handler must check |
| Cookie present, user not in DB | currentUserTag not set |
| Cookie present, valid user | currentUserTag set with permissions + team capabilities |
| User has no team | team defaults to 'default', teamCapabilities is [] |

## Permission vs Capability

- **Permissions** (`string[]`): Direct user permissions (e.g., `approvals`, `payments`). Use `can()` / `asserts()` for authorization.
- **Team Capabilities** (`string[]`): Team-level feature flags (e.g., `notifications`, `workbench`). Use for feature gating.
