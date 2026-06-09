---
id: adr-20260123-devtools-docs
c3-seal: ecb07c1d5656c9883817fb98b0ae7fa5db066314d2ee8ac160bb7f8669b791df
title: Document DevTools Component
type: adr
goal: 'Document and implement the architectural decision: Document DevTools Component.'
status: implemented
date: "2026-01-23"
affects:
    - c3-1
approved-files:
    - .c3/c3-1-web-frontend/README.md
    - .c3/c3-1-web-frontend/c3-108-devtools.md
    - .c3/TOC.md
base-commit: a3483d229fa5d3541b0f9e852e4bd94213ecd761
---

# Document DevTools Component

## Goal

Document and implement the architectural decision: Document DevTools Component.

## Status

**Implemented** - 2026-01-23

## Problem

The DevShell component was recently refactored (commit 926761f) from a sidebar to a bottom toolbar, resulting in 689+ lines of code. This developer experience component provides:

- Viewport controls
- Theme toggle
- Fullscreen mode
- Log viewer panel
- Keyboard shortcuts

Despite being substantial infrastructure, it has no C3 documentation.

## Decision

Create C3 component documentation for DevTools:

### c3-1 Web Frontend

- **c3-108-devtools.md**: Document the DevShell component, its features, and integration points

The doc will cover:

- Purpose (development tooling, not production)
- Feature breakdown (viewport, theme, logs, shortcuts)
- Keyboard shortcuts (F for fullscreen, ESC to exit)
- Integration with root layout

## Rationale

| Considered | Rejected Because |
| --- | --- |
| Skip documentation (dev-only) | Component is substantial, affects root layout, future devs need orientation |
| Merge into c3-101 State Management | DevTools is UI component, not state management |

## Affected Layers

| Layer | Document | Change |
| --- | --- | --- |
| c3-1 | README.md | Add c3-108 to components table |
| c3-1 | c3-108-devtools.md | New component doc |

## Verification

- [x] c3-108-devtools.md created
- [x] Documents DevShell features
- [x] Documents keyboard shortcuts
- [x] c3-1/README.md lists c3-108 in components table
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
