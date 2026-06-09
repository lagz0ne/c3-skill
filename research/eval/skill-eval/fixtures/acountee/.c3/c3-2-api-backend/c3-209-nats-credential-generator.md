---
id: c3-209
c3-version: 3
c3-seal: f535067b0e782dd66331e86c2323054394079304f70567b642309242429855b1
title: NATS Credential Generator
type: component
category: foundation
parent: c3-2
goal: Generate per-session JWT + nkey for NATS WebSocket authentication
uses:
    - ref-nats-jwt-auth
    - ref-pumped-fn
    - ref-scope-controlled-config
    - ref-structured-logging
---

# NATS Credential Generator

## Goal

Generate per-session JWT + nkey for NATS WebSocket authentication

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | Own NATS Credential Generator behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep NATS Credential Generator decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide durable agent-ready documentation for NATS Credential Generator so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before NATS Credential Generator behavior is changed. | ref-nats-jwt-auth |
| Inputs | Accept only the files, commands, data, or calls that belong to NATS Credential Generator ownership. | ref-nats-jwt-auth |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | ref-nats-jwt-auth |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | ref-nats-jwt-auth |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks NATS Credential Generator to deliver its documented responsibility. | ref-nats-jwt-auth |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | ref-nats-jwt-auth |
| Alternate paths | When a request falls outside NATS Credential Generator ownership, hand it to the parent or sibling component. | ref-nats-jwt-auth |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | ref-nats-jwt-auth |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-nats-jwt-auth | ref | Governs NATS Credential Generator behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| NATS Credential Generator input | IN | Callers must provide context that matches the component goal and parent fit. | c3-2 boundary | c3x lookup plus targeted tests or review. |
| NATS Credential Generator output | OUT | Derived code, docs, and tests must preserve the documented behavior and governance. | c3-2 boundary | c3x check and project test suite. |

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

- `@nats-io/nkeys` — `createUser()`, `fromSeed()`
- `@nats-io/jwt` — `encodeUser()`, `Algorithms`
- Tags: `natsConfig.wsUrl`, `natsConfig.accountSeed`, `natsConfig.subjectPrefix`

## How It Works

The `natsCredentialGenerator` atom parses the account seed on init, then exposes a `generate(email, ttl)` method:

1. Creates an ephemeral user keypair (`createUser()`)
2. Builds user claims with scoped permissions (subscribe-only, WebSocket-only)
3. Signs a JWT with the account key
4. Returns `{ jwt, seed }` to the caller

The loader passes credentials to the client via `loaderData`. The client uses `credsAuthenticator(jwt, seed)` to connect.

## Permission Model

| Permission | Value | Reason |
| --- | --- | --- |
| Subscribe | {prefix}.broadcast, {prefix}.user.{escaped_email} | Receive broadcast + user-specific sync updates |
| Publish | empty allow (no permissions) | Clients are read-only |
| Connection | WEBSOCKET only | Browser access |

Email is escaped for NATS subjects: `@` and `.` replaced with `_`.

## Configuration

| Tag | Example |
| --- | --- |
| natsConfig.wsUrl | wss://nats:8080 |
| natsConfig.accountSeed | SABC... (secret) |
| natsConfig.subjectPrefix | sync |

## Security

- Account seed is server-side only — never sent to client
- User seeds are ephemeral — generated per session, not stored
- JWT expires after TTL (default 1 hour) — client must reconnect
- No publish rights — clients cannot write to NATS
