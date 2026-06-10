# Depth round 1 — Answer Depth Contract result

Skill change under test: commit 3c61be0 (Answer Depth Contract in query.md, failure-boundary
item 6 in Step 0a++, direct/transitive + Verification table in sweep.md, ADR status labels).

| Arm | Generator | Skill | Judge | Pass |
|-----|-----------|-------|-------|------|
| Cold baseline (re-judged control) | codex | pre-contract | current K=3 calibrated | 2/13 (mean 3.75) |
| Depth round 1 | claude/fable | with contract | current K=3 calibrated | 10/13 (mean 4.20) |

Same judge on both arms (control re-run 2026-06-10): recalibration did NOT inflate the old
answers — the same 2 cases pass either way. The delta comes from the answers.

Dimension shift matches the contract's targets: trace_completeness and reasoning_depth moved
from typical 3 to 4–5, change_usefulness mostly 5; correctness/no-hallucination held at 4.

Remaining fails: SLACK-APPROVAL (genuine wrong claim: service-level entry does NOT emit sync
deltas), FILE-IDEMPOTENCY (over-guaranteed audit claims, missed c3-204 owner), STEP-ADVANCE
(gate-failed on abbreviating an ADR id — partially an eval artifact; candidate eval tweak:
resolve unique-prefix ids against the inventory instead of counting them invented).

Caveat: generator changed between arms (codex -> claude/fable); attribution to the skill rests
on the dimension pattern, not a same-generator A/B. Judge remains same-vendor (codex reviewers).
