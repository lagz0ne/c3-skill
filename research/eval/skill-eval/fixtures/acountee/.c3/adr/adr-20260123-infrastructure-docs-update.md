---
id: adr-20260123-infrastructure-docs-update
c3-seal: e1662f53112cb5d779fdcc9f241862d8a58aa843c7d6b61be146fd2ce122b2ea
title: Update Infrastructure Documentation (NATS, Middleware, Migrations)
type: adr
goal: 'Document and implement the architectural decision: Update Infrastructure Documentation (NATS, Middleware, Migrations).'
status: implemented
date: "2026-01-23"
affects:
    - c3-2
    - c3-203
    - c3-204
    - c3-4
approved-files:
    - .c3/c3-4-nats-server/README.md
    - .c3/c3-2-api-backend/c3-203-middleware-stack.md
    - .c3/c3-2-api-backend/c3-204-drizzle-orm.md
    - .c3/TOC.md
base-commit: a3483d229fa5d3541b0f9e852e4bd94213ecd761
---

# Update Infrastructure Documentation

## Goal

Document and implement the architectural decision: Update Infrastructure Documentation (NATS, Middleware, Migrations).

## Status

**Implemented** - 2026-01-23

## Problem

Recent changes introduced infrastructure updates that aren't reflected in C3 documentation:

1. **c3-4 NATS Server**: JetStream now enabled for notification system, env-based JWT injection, updated permissions (`$JS.API.>`, `_INBOX.>`)
**c3-4 NATS Server**: JetStream now enabled for notification system, env-based JWT injection, updated permissions (`$JS.API.>`, `_INBOX.>`)
2. **c3-203 Middleware Stack**: `teamCapabilities` added to UserActor (commit d0ad6bc)
**c3-203 Middleware Stack**: `teamCapabilities` added to UserActor (commit d0ad6bc)
3. **c3-204 Drizzle ORM**: Migration file location and auto-migration system not documented
**c3-204 Drizzle ORM**: Migration file location and auto-migration system not documented

These updates create doc-code drift in existing component documentation.

## Decision

Update existing C3 component docs to reflect current state:

### c3-4 NATS Server README

- Add JetStream configuration section
- Document `NOTIFICATIONS` stream for notification system
- Update permission model to include `$JS.API.>`, `_INBOX.>`
- Note environment variable-based JWT injection

### c3-203 Middleware Stack

- Add `teamCapabilities` to UserActor type documentation
- Update the middleware output description

### c3-204 Drizzle ORM

- Document migration file location: `apps/start/src/server/dbs/migrations/*.sql`
- Document auto-migration system in `server.tsx`
- Note timestamp prefix convention (`0001_description.sql`)

## Rationale

| Considered | Rejected Because |
| --- | --- |
| Create new component docs | Updates are to existing components, not new ones |
| Defer documentation | Increases doc-code drift; makes troubleshooting harder |

## Affected Layers

| Layer | Document | Change |
| --- | --- | --- |
| c3-4 | README.md | Add JetStream, update permissions |
| c3-2 | c3-203-middleware-stack.md | Add teamCapabilities |
| c3-2 | c3-204-drizzle-orm.md | Add migration docs |

## Verification

- [x] c3-4/README.md documents JetStream configuration
- [x] c3-4/README.md lists `$JS.API.>` and `_INBOX.>` permissions
- [x] c3-203 documents `teamCapabilities` property
- [x] c3-204 documents migration file location
- [x] c3-204 documents auto-migration via `runMigrations`

## Implementation Notes

Completed 2026-01-23. All documentation updated to reflect current infrastructure state:

1. **c3-4 NATS Server**: Added JetStream Configuration section with storage limits, NOTIFICATIONS stream documentation, updated Permission Model table with `$JS.API.>` and `_INBOX.>`, added JWT Injection at Runtime section explaining env-based approach
**c3-4 NATS Server**: Added JetStream Configuration section with storage limits, NOTIFICATIONS stream documentation, updated Permission Model table with `$JS.API.>` and `_INBOX.>`, added JWT Injection at Runtime section explaining env-based approach
2. **c3-203 Middleware Stack**: Added complete UserActor Type section with all properties including `teamCapabilities`, added property table with descriptions, added Permission vs Capability clarification, updated edge cases for team-related scenarios
**c3-203 Middleware Stack**: Added complete UserActor Type section with all properties including `teamCapabilities`, added property table with descriptions, added Permission vs Capability clarification, updated edge cases for team-related scenarios
3. **c3-204 Drizzle ORM**: Added Migration File Location section with full path, listed current migration files, documented naming convention (NNNN_description.sql), added Migration Tracking section with pgmigrations schema, documented Migration Flow steps, added Idempotent Patterns guidance
**c3-204 Drizzle ORM**: Added Migration File Location section with full path, listed current migration files, documented naming convention (NNNN_description.sql), added Migration Tracking section with pgmigrations schema, documented Migration Flow steps, added Idempotent Patterns guidance

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
