---
id: adr-20260126-scope-patterns-adoption
c3-seal: 8d7d876540a022ab53bd60306fa22602d7ac3336a94b6ce5ded10b60ec36013e
title: Adopt Scope-Based Patterns as Standards
type: adr
goal: 'Document and implement the architectural decision: Adopt Scope-Based Patterns as Standards.'
status: implemented
---

# Adopt Scope-Based Patterns as Standards

## Goal

Document and implement the architectural decision: Adopt Scope-Based Patterns as Standards.

## Problem

During notification system implementation, several patterns emerged that improve code consistency, testability, and maintainability:

1. **Logging** - Mixed usage of console.log and rootLogger
2. **Config access** - Closures capturing `process.env` directly
3. **Test scopes** - Custom scope factories duplicating logic
4. **Pub/sub** - Inconsistent dispatcher patterns

## Decision

Document these patterns as refs to ensure consistent application across the codebase.

## Refs Created

| Ref | Goal |
| --- | --- |
| ref-structured-logging | Use rootLogger from scope, not console.log |
| ref-scope-controlled-config | Closures must not capture process.env |
| ref-test-scope-composition | Composable test scopes with defaults |
| ref-pull-dispatcher | Channels subscribe to dispatcher (pull) |

## Affected Layers

| Layer | Change |
| --- | --- |
| refs | Added 4 new refs |
| c3-2-api | server.tsx, migrate.ts follow these patterns |
| c3-211 | Notification system follows pull-dispatcher |

## Evidence

Patterns applied in:

- `src/server.tsx` - structured logging, scope-controlled config
- `src/server/dbs/migrate.ts` - structured logging
- `src/server/resources/serverConfig.ts` - config atom pattern
- `src/server/resources/notificationDispatcher.ts` - pull dispatcher
- `test/notifications/setup.ts` - test scope composition

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
