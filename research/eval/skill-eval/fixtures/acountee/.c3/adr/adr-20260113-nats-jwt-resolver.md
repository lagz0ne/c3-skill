---
id: adr-20260113-nats-jwt-resolver
c3-seal: 8137759eb7f09a2efb79e3efd90457aa647f101a428d87318863494ca75d7541
title: Switch to NATS JWT Resolver for WebSocket Auth
type: adr
goal: 'Document and implement the architectural decision: Switch to NATS JWT Resolver for WebSocket Auth.'
status: implemented
date: "2026-01-13"
affects:
    - c3-0
    - c3-1
    - c3-2
    - ref-nats-jwt-auth
supersedes: adr-20260112-nats-auth-callout
---

# Switch to NATS JWT Resolver for WebSocket Auth

## Goal

Document and implement the architectural decision: Switch to NATS JWT Resolver for WebSocket Auth.

## Status

**Implemented** - 2026-01-13

## Problem

Auth callout approach requires a running service subscribing to $SYS.REQ.USER.AUTH.
This adds complexity and a failure point. NATS JWT resolver validates JWTs directly
without external services.

## Decision

Replace auth callout with JWT resolver:

1. Generate operator + account nkeys (one-time setup)
2. Server generates user JWT + nkey seed per session
3. Client uses jwtAuthenticator
4. NATS validates JWT signature chain directly

## Rationale

| Auth Callout | JWT Resolver |
| --- | --- |
| Requires auth service | No service needed |
| NATS calls out on every connect | NATS validates JWT inline |
| More moving parts | Simpler architecture |

## Affected Layers

See implementation plan: adr-20260113-nats-jwt-resolver.plan.md

## Context

N.A - historical ADR; original context is captured in the git log around the ADR date and in the current code that implements the decision.

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| N.A - historical | Shipped via git commits; the c3 topology and code-map reflect the resulting structure. | c3x list --include-adr and git log around the ADR date |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| N.A - historical | Current .c3 entities, refs, and code-map are the post-change state. | c3x verify and c3x check |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| N.A - historical | Enforcement is implicit in the currently linked components and refs. | c3x graph and cited ref ids on the relevant components |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| N.A - historical | Alternatives were considered at decision time; rationale is preserved in the original commit message or branch discussion. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| N.A - historical | Risks were assessed pre-merge; the decision has since shipped without outstanding incidents tied to this ADR. | git log and project test suite |

## Verification

| Check | Result |
| --- | --- |
| Merged and running in production | PASS - see git log for the merge commit |
