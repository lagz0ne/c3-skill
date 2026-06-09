# Skill Eval Loop Summary

## Score Trajectory

| Round | Total | Failed criteria | Stop status |
| --- | ---: | --- | --- |
| 1 | 82 / 87 | U2 search-first failed on all 5 cases | Continued |
| 2 | 87 / 87 | None | Stopped: all cases cleared the round-1 pass bar |

## Round Changes

| Round | Skill/rubric change | Why |
| --- | --- | --- |
| 1 | Clarified `skills/c3/SKILL.md` read-only precondition so conceptual queries must run `c3 search` before `c3 list` or `c3 check`. | Round 1 answers reflected the prior ambiguity: the skill said search-first, but also had a blanket `list/check` precondition. |
| 2 | Updated `skills/c3/references/sweep.md` so impact questions also start with `c3 search`, then expand to graph/list after candidates are known. | NATS impact questions can route to sweep guidance; sweep still had topology-first wording after round 1. |

## What Improved

- Search-first behavior improved from failing every case in round 1 to passing every case in round 2.
- Answers consistently cited exact fixture ids and paired governing refs with why they apply.
- The skill now has aligned query and sweep guidance for conceptual discovery.

## Still Weak

- The scorer is deterministic text matching, so it proves rubric coverage but not deeper semantic quality.
- Fixture `c3 check` rebuilds cache but reports known `c3-0` canonical drift; answers must caveat that rather than treating check as clean.
- The eval covers query/impact behavior only. It does not prove change, audit, canvas, ref, or rule workflows.
