---
id: adr-20260123-notification-system-docs
c3-seal: 1ee66f4735f4fa7a7d4c89b5c09371ce677bd9a0de196c1bb49908957e0a4091
title: Document Notification System Components
type: adr
goal: 'Document and implement the architectural decision: Document Notification System Components.'
status: implemented
date: "2026-01-23"
affects:
    - c3-2
approved-files:
    - .c3/c3-2-api-backend/README.md
    - .c3/c3-2-api-backend/c3-211-notification-system.md
    - .c3/TOC.md
base-commit: a3483d229fa5d3541b0f9e852e4bd94213ecd761
---

# Document Notification System Components

## Goal

Document and implement the architectural decision: Document Notification System Components.

## Status

**Implemented** - 2026-01-23

## Problem

ADR-20260121-notification-system was accepted and implementation exists, but the C3 component documentation was never created. The following code exists without C3 docs:

- `notificationPublisher.ts` - Publishes to NATS JetStream
- `notificationDispatcher.ts` - Consumes from NATS, routes to channels
- `notification.ts` (service) - Notification orchestration

This creates doc-code drift where notification infrastructure is undocumented.

## Decision

Create a single C3 component document covering the notification system:

### c3-2 API Backend

- **c3-211-notification-system.md**: Document the publisher, dispatcher, and service as one cohesive notification component

The component doc will cover:

- Publisher → JetStream flow
- Dispatcher → Channel routing
- Integration with PR flows
- NATS subject design (`notifications.>`)

## Rationale

| Considered | Rejected Because |
| --- | --- |
| Separate docs for publisher/dispatcher/service | Over-fragmentation; they form one logical system |
| Document under c3-205 PR Flows | Notification is independent; can extend to other triggers |

## Affected Layers

| Layer | Document | Change |
| --- | --- | --- |
| c3-2 | README.md | Add c3-211 to components table |
| c3-2 | c3-211-notification-system.md | New component doc |

## Verification

- [x] c3-211-notification-system.md created
- [x] Documents publisher, dispatcher, service architecture
- [x] Documents NATS JetStream integration
- [x] c3-2/README.md lists c3-211 in components table
- [x] TOC.md updated
- [x] Code references point to actual files

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
