---
id: adr-20260123-admin-features-docs
c3-seal: b2c4563db6ad0240301435b9221e87fe49f1258c95e36a1b1155a0e0bb8469c6
title: Document Admin Feature Components
type: adr
goal: 'Document and implement the architectural decision: Document Admin Feature Components.'
status: implemented
date: "2026-01-23"
affects:
    - c3-1
    - c3-2
approved-files:
    - .c3/c3-1-web-frontend/README.md
    - .c3/c3-1-web-frontend/c3-107-admin-screens.md
    - .c3/c3-2-api-backend/README.md
    - .c3/c3-2-api-backend/c3-210-admin-flows.md
    - .c3/TOC.md
base-commit: a3483d229fa5d3541b0f9e852e4bd94213ecd761
---

# Document Admin Feature Components

## Goal

Document and implement the architectural decision: Document Admin Feature Components.

## Status

**Implemented** - 2026-01-23

## Problem

ADR-20260121-admin-management-features was implemented but the C3 component documentation was never created. The code exists and works, but there's no architectural documentation for:

- 5 admin screens (UserManagement, TeamManagement, AuditLog, ApprovalConfig, Organization)
- Admin flows (teamFlows, roleFlows, userFlows, approvalConfigFlows)

This creates doc-code drift where the audit shows undocumented components.

## Decision

Create C3 component documentation for the implemented admin features:

### c3-1 Web Frontend

- **c3-107-admin-screens.md**: Document all 5 admin screens as a single component (they share AdminSidebar, follow consistent patterns)

### c3-2 API Backend

- **c3-210-admin-flows.md**: Document team, role, user, and approvalConfig flows as a single component (they form a cohesive admin feature set)

### Container README Updates

- Add c3-107 to c3-1 components table
- Add c3-210 to c3-2 components table

## Rationale

| Considered | Rejected Because |
| --- | --- |
| One doc per screen (5 docs) | Excessive granularity; screens share patterns and are cohesive feature |
| Skip documentation | Creates ongoing doc-code drift; makes onboarding harder |

Grouping related screens/flows into single docs matches the pattern used for c3-205 (PR Flows) which documents multiple related flows together.

## Affected Layers

| Layer | Document | Change |
| --- | --- | --- |
| c3-1 | README.md | Add c3-107 to components table |
| c3-1 | c3-107-admin-screens.md | New component doc |
| c3-2 | README.md | Add c3-210 to components table |
| c3-2 | c3-210-admin-flows.md | New component doc |

## Verification

- [x] c3-107-admin-screens.md created with all 5 screens documented
- [x] c3-210-admin-flows.md created with team/role/user/approvalConfig flows
- [x] c3-1/README.md lists c3-107 in components table
- [x] c3-2/README.md lists c3-210 in components table
- [x] TOC.md updated to include new docs
- [x] Code references in docs point to actual file locations

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
