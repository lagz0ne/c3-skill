---
id: c3-213
c3-version: 3
c3-seal: 54220613751107c8c0e8f234033d3266caf40d1440da596c580b94972e6cb6f2
title: Authentication Flows
type: component
category: foundation
parent: c3-2
goal: Google OAuth and test token authentication flows
uses:
    - ref-authentication
    - ref-pumped-fn
    - ref-rbac
    - ref-structured-logging
---

# Authentication Flows

## Goal

Google OAuth and test token authentication flows

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own Authentication Flows behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep Authentication Flows decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for Authentication Flows so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before Authentication Flows behavior is changed. | ref-authentication |
| Inputs | Accept only the files, commands, data, or calls that belong to Authentication Flows ownership. | ref-authentication |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-authentication |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-authentication |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks Authentication Flows to deliver its documented responsibility. | ref-authentication |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-authentication |
| Alternate paths | When a request falls outside Authentication Flows ownership, hand it to the parent or sibling component. | ref-authentication |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-authentication |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-authentication | ref | Governs Authentication Flows behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| Authentication Flows input | IN | Callers must provide context that matches the component goal and parent fit. | c3-2 boundary | c3x lookup plus targeted tests or review. |
| Authentication Flows output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-2 boundary | c3x check and project test suite. |

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

- `@pumped-fn/lite` — `flow()`
- `googleapis` — Google OAuth2 client
- `initUserActor` atom — looks up user in DB, builds `UserActor` with permissions
- `appConfig` — `enableTestToken`, `testToken` settings

## Flows

### authenticateWithGoogle

Input: `{ credentials: Credentials }` (Google OAuth credentials)

1. Calls `gauthSvc.getProfile(credentials)` to exchange credentials for Google profile
2. Extracts and normalizes email (lowercase)
3. Calls `initUserActor(ctx, email, { avatar, displayName, name })` to look up user in DB
4. Returns `UserActor` on success

### authenticateWithTestToken

Input: `{ token: string, email: string }`

1. Checks `appConfig.enableTestToken` is true and token matches `appConfig.testToken`
2. Normalizes email (lowercase)
3. Calls `initUserActor(ctx, email, { avatar, displayName, name })` with generated avatar
4. Returns `UserActor` on success

## Results

| Result | When |
| --- | --- |
| { success: true, user } | Authentication succeeded |
| { success: false, reason: 'NO_CREDENTIALS' } | No email in Google profile |
| { success: false, reason: 'INVALID_TOKEN' } | Test token mismatch or disabled |
| { success: false, reason: 'USER_NOT_FOUND', email } | Email not in users table |

## Google OAuth Service

`gauthSvc` atom wraps `google.auth.OAuth2` with:

- `generateOfflineUrl()` — builds consent URL with email + profile scopes
- `getToken(code)` — exchanges authorization code for tokens
- `getProfile(credentials)` — fetches user info (email, name, picture)

Configuration via `googleOAuthConfig` atom: `clientId`, `clientSecret`, `redirectUri`.

## Cookie Management

Authentication flows return the user; the route handler sets a `user` cookie with the email. Subsequent requests use this cookie via `getCurrentUserMiddleware` to restore the session.
