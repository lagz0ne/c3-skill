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

## Crosscut Phase

Added four indirect cases in `cases/acountee-crosscut.md` and extended
`rubric.jsonl` / `score.py` with trace coverage, explicit sync mechanism,
explicit notification mechanism, and emergent-property checks.

| Round | Total | Main failures | Stop status |
| --- | ---: | --- | --- |
| 1 | 195 / 229 | Missing cross-cut notification ADRs/properties; named fake negative `rule-*` ids; incomplete trace segments. | Continued |
| 2 | 225 / 229 | Trace mostly fixed; exact mechanism names still weak (`NATS JetStream`, `sync.user`, `slackChannel`). | Continued |
| 3 | 229 / 229 | None. | Stopped: all expanded cases passed. |

Round commits:

- `f04b69c` - `crosscut round 1: add trace expansion guidance (total 0->195)`
- `13f7077` - `crosscut round 2: name concrete notification mechanisms (total 195->225)`
- `0cc874a` - `crosscut round 3: record all-pass mechanism trace (total 225->229)`

Skill changes that helped cross-cut tracing:

- Added query guidance to expand from the first owner into action/command owner,
  state mutation owner, sync mechanism, notification mechanism, and emergent
  property.
- Added explicit instruction to avoid invented negative `rule-*` ids when no
  rules exist.
- Added exact mechanism naming guidance so answers copy queue/subject/channel
  names from C3 reads instead of saying only "notification system".

Scorer change:

- Fixed U5 governance-with-why matching to check all answer occurrences of a
  ref id. The first occurrence is often in Evidence commands, which should not
  shadow a later Grounding explanation.

Still weak after crosscut phase:

- Text checks can be satisfied by exact terms, so the eval measures trace shape
  better than deep causal reasoning.
- New crosscut cases are notification-heavy; future cases should add different
  emergent properties such as audit atomicity, config blast radius, and
  transport auth/sync coupling.
- Fixture check still has known acountee drift; the eval uses that as fixture
  context, not a clean-docs proof.
