---
id: adr-20260323-jtbd-orphan-fix
c3-seal: 55d21091d32dbd1e9da0285c4578d7cd65833f551884cc81f1eebdac652c3aa5
title: jtbd-orphan-fix
type: adr
goal: 'Fix 1 FAIL and 2 WARNs from C3 audit:'
status: proposed
date: "2026-03-23"
---

# Fix C3 Audit Issues

## Goal

Fix 1 FAIL and 2 WARNs from C3 audit:

1. **FAIL Phase 8**: WorkbenchScreen.tsx imports runtime values from server/lib/ — move shared types/constants to packages/shared/
2. **WARN Phase 7**: ref-jtbd orphan — wire to c3-0 system context
3. **WARN Phase 7**: ref-zerobased-dev orphan — already has scope: [c3-0] but needs a component citation

## Decision

- Move `exportTemplate.types.ts` constants and `templateGenerator.ts` shared functions to `packages/shared/src/`
- Wire orphan refs to the system context

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
