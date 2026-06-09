---
id: adr-20260407-refresh-c3-codemap-and-enforcement
c3-seal: 5c879a77397de68801f9e1fe129875553146d162f04cb963f0ec884d1d14bf58
title: Refresh c3 codemap and enforcement
type: adr
goal: Update the project C3 codemap to reflect the current live test support files, verify structural integrity after the refresh, and surface any remaining coverage drift that is caused by tracked-but-deleted paths outside the current c3x enforcement surface.
status: implemented
date: "2026-04-07"
affects:
    - c3-204
    - c3-211
---

# Refresh c3 codemap and enforcement

## Goal

Update the project C3 codemap to reflect the current live test support files, verify structural integrity after the refresh, and surface any remaining coverage drift that is caused by tracked-but-deleted paths outside the current c3x enforcement surface.

## Context

N.A - historical ADR; original context is captured in the git log around the ADR date and in the current code that implements the decision.

## Decision

N.A - historical ADR; the decision matches the Goal above and has already shipped. Current .c3 topology reflects the implemented outcome.

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
