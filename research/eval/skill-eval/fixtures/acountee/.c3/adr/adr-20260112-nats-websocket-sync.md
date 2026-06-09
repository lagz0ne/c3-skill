---
id: adr-20260112-nats-websocket-sync
c3-seal: f42d567448d35655338dfbca60360bd0361b04d18a686c370f0d9cf3fdbb84c6
title: Replace WebSocket Workaround with NATS
type: adr
goal: 'Document and implement the architectural decision: Replace WebSocket Workaround with NATS.'
status: implemented
date: "2026-01-12"
affects:
    - c3-0
    - c3-1
    - c3-2
    - ref-sync
---

# Replace WebSocket Workaround with NATS

## Goal

Document and implement the architectural decision: Replace WebSocket Workaround with NATS.

## Status

**Implemented** - 2026-01-12

## Problem

TanStack Start (Nitro) doesn't support WebSocket natively. The current workaround runs a separate Bun HTTP+WS server and uses an HTTP bridge (`/_internal/ws-push`) to communicate between Nitro and the WebSocket server. This architecture is clunky: two processes, extra HTTP hop per message, and custom connection management code.

## Decision

Replace the entire WebSocket workaround with NATS:

1. **External NATS server** with WebSocket gateway enabled
2. **Browser connects directly** to NATS via `nats.ws` library
3. **Server publishes** to NATS subjects instead of managing WebSocket connections
4. **NATS auth callout** delegates authentication to our auth service

Topic structure:

- `sync.user.{email}` - User-specific sync messages (for PR filtering)
- `sync.broadcast` - Messages to all users

## Rationale

| Considered | Rejected Because |
| --- | --- |
| Keep current workaround | Complex, two processes, HTTP bridge latency |
| Upgrade to native WS framework | Would require leaving TanStack Start |
| Server-side WS proxy to NATS | Still need WS code in server, defeats simplification |

NATS provides:

- Built-in WebSocket gateway (no server-side WS code needed)
- Pub/sub semantics match our delta broadcast pattern
- Auth callout for delegated authentication
- Horizontal scaling out of the box

## Affected Layers

| Layer | Document | Change |
| --- | --- | --- |
| Context | c3-0 | Add NATS to External Systems, update Linkages |
| Container | c3-1 | Replace WebSocket client with nats.ws, update Tech Stack |
| Container | c3-2 | Remove wsServer, rewrite sync to NATS publish, update Overview |
| Reference | ref-sync | Rewrite for NATS pub/sub model |

## Code Changes

### Delete

- `apps/start/src/server/resources/wsServer.ts` - No longer needed

### Modify (Server)

- `apps/start/src/server/resources/sync.ts`:
Remove HTTP bridge

Add NATS connection

Publish to `sync.user.{email}` for user-specific, `sync.broadcast` for all

- Remove `/_internal/ws-push` and `/_internal/ws-users` endpoints from entry.ts

### Modify (Client)

- `apps/start/src/lib/pumped/atoms/sync.ts`:
Replace `api.connectWebSocket()` with nats.ws subscription

Subscribe to `sync.user.{email}` and `sync.broadcast`

- `apps/start/src/lib/pumped/atoms/api.ts`:
Remove `connectWebSocket()` method

### Add

- NATS auth callout handler (new server function or endpoint)
- NATS connection configuration (env vars for URL, credentials)

## Dependencies

- Add `nats.ws` to frontend dependencies
- Add `nats` to server dependencies

## Infrastructure

- Deploy NATS server with WebSocket enabled
- Configure NATS auth callout to validate JWT tokens
- Expose NATS WebSocket port (typically 8080)

## Verification

- [ ] Browser can connect to NATS via WebSocket
- [ ] Auth callout validates user tokens correctly
- [ ] Delta messages delivered to correct users
- [ ] PR filtering still works (users only see authorized PRs)
- [ ] No `wsServer.ts` or HTTP bridge code remains
- [ ] E2E tests pass
- [ ] C3 audit passes

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
