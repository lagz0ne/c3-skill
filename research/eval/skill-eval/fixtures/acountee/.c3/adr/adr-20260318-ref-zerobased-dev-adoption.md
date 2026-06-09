---
id: adr-20260318-ref-zerobased-dev-adoption
c3-seal: 1c71b4622090669e98f4149f25581c294c3f7876c9a6239acba3423e09919046
title: Adopt zerobased as standard local dev router
type: adr
goal: Replace portless setup scripts with zerobased for zero-config Docker service routing. Eliminates port conflicts across worktrees and removes manual alias maintenance.
status: implemented
date: "2026-03-18"
affects:
    - ref-zerobased-dev
---

# Adopt zerobased as standard local dev router

## Goal

Replace portless setup scripts with zerobased for zero-config Docker service routing. Eliminates port conflicts across worktrees and removes manual alias maintenance.

## Work Breakdown

- [x] Create `ref-zerobased-dev` documenting the pattern
- [x] Update `CLAUDE.md` dev environment section
- [x] Remove `scripts/portless-setup.sh`
- [x] Update `scripts/reset-and-dev.sh` — replace portless invocation with zerobased
- [x] Update `scripts/reset-and-start.sh` — replace portless invocation with zerobased
- [x] Fix `db.ts` — socket URI parsing for `postgres` lib
- [x] Update `package.json` dev script — portless → zerobased
- [x] Update `.env` and `.env.development` — zerobased connection strings

## Risks

- Reset scripts still reference portless — must be updated before deleting the script
- Existing `.env` files still use `localhost:35432` — developers need to update per ref

## Context

N.A - historical ADR; original context is captured in the git log around the ADR date and in the current code that implements the decision.

## Decision

N.A - historical ADR; the decision matches the Goal above and has already shipped. Current .c3 topology reflects the implemented outcome.

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

## Verification

| Check | Result |
| --- | --- |
| Merged and running in production | PASS - see git log for the merge commit |
