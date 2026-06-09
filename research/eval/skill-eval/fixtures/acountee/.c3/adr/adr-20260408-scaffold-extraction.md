---
id: adr-20260408-scaffold-extraction
c3-seal: da786af2d645f9b5e11a38e0628396265b80106b6bd0ee2a6d64a011e98006e4
title: scaffold-extraction
type: adr
goal: Extract export template from apps/start into a shared @acountee/shared package so bank-transfer export logic and SFT test infrastructure can be reused across apps without duplication.
status: implemented
date: "2026-04-08"
---

# scaffold-extraction

## Goal

Extract export template from apps/start into a shared @acountee/shared package so bank-transfer export logic and SFT test infrastructure can be reused across apps without duplication.

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
