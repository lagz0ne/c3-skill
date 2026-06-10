# C3 Skill Eval Judge Summary

## Scope

Cold baseline answers live in `research/eval/skill-eval/runs/judge-cold-baseline/`.

The deterministic text scorer gives this run `374/374`. Every case includes the expected IDs, key terms, local C3 command shape, and no hallucinated fixture IDs.

The semantic judge was then run on all 13 answers with:

```bash
python3 research/eval/skill-eval/judge/judge.py <CASE_ID> research/eval/skill-eval/runs/judge-cold-baseline/<CASE_ID>.md --output research/eval/skill-eval/runs/judge-cold-baseline/judge/<CASE_ID>.json
```

## Judge Rubric

The judge scores six dimensions from 1 to 5:

| Dimension | Weight | Bar |
| --- | ---: | --- |
| Correctness | 25% | Claims match the fixture and case ground truth. |
| Trace completeness | 20% | The answer traces action -> state change -> mechanism -> dependent/observer -> emergent property. |
| Reasoning depth | 20% | The answer explains why/how the behavior emerges, including boundary and failure mode. |
| Grounding | 15% | Important claims are tied to cited C3 reads, graphs, searches, or case ground truth. |
| No hallucination | 10% | No invented IDs, guarantees, rules, or unsupported overclaims. |
| Change usefulness | 10% | The answer gives enough owners, risks, and checks to safely assess or make a change. |

Pass requires overall >= 4.0, correctness >= 4, no-hallucination >= 4, and no dimension below 3.

## Quality Baseline

| Case | Text score | Judge | Verdict | C | T | R | G | H | U |
| --- | ---: | ---: | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| ADMIN-1 | 20/20 | 3.45 | fail | 4 | 3 | 3 | 3 | 5 | 3 |
| APPROVAL-1 | 23/23 | 3.60 | fail | 4 | 3 | 3 | 4 | 5 | 3 |
| AUTH-1 | 24/24 | 3.35 | fail | 4 | 3 | 3 | 3 | 4 | 3 |
| CROSSCUT-MASS-APPROVAL-1 | 30/30 | 3.85 | fail | 4 | 4 | 4 | 3 | 4 | 4 |
| CROSSCUT-NOTIFICATION-BELL-1 | 28/28 | 4.10 | pass | 4 | 4 | 4 | 4 | 5 | 4 |
| CROSSCUT-SLACK-APPROVAL-1 | 30/30 | 3.85 | fail | 4 | 4 | 4 | 3 | 4 | 4 |
| CROSSCUT-STEP-ADVANCE-1 | 29/29 | 3.95 | fail | 4 | 4 | 4 | 3 | 5 | 4 |
| NATS-1 | 22/22 | 3.35 | fail | 4 | 3 | 3 | 3 | 4 | 3 |
| PROPERTY-AUDIT-ATOMICITY-1 | 32/32 | 4.00 | pass | 4 | 4 | 4 | 4 | 4 | 4 |
| PROPERTY-CONFIG-BLAST-RADIUS-1 | 44/44 | 3.85 | fail | 4 | 4 | 4 | 3 | 4 | 4 |
| PROPERTY-FILE-IDEMPOTENCY-1 | 33/33 | 3.75 | fail | 4 | 4 | 4 | 3 | 3 | 4 |
| PROPERTY-TRANSPORT-SYNC-COUPLING-1 | 36/36 | 3.85 | fail | 4 | 4 | 4 | 3 | 4 | 4 |
| UI-1 | 23/23 | 3.60 | fail | 4 | 3 | 3 | 4 | 5 | 3 |

Legend: C = correctness, T = trace completeness, R = reasoning depth, G = grounding, H = no hallucination, U = change usefulness.

Summary: 13/13 pass deterministic text matching; 2/13 pass semantic judging.

## Key Gap

The text scorer says the run is perfect because the answers contain the right IDs, commands, terms, and negative rule facts. The judge shows that this is not enough. Eleven cases pass text matching but fail semantic quality:

| Case | Main semantic failure |
| --- | --- |
| ADMIN-1 | Correct owners, but missing supporting recipes, most governing refs, `ref-sync`, ADR affected topology, and change-safety boundaries. |
| APPROVAL-1 | Correct primary owners, but omits `c3-211`, `c3-212` cleanup detail, backend flow refs, audit double-write caveat, and full action/state/observer trace. |
| AUTH-1 | Correct auth layers, but misses `ref-nats-jwt-auth` graph, under-grounds `c3-4`, and compresses login -> RBAC -> execution context -> NATS validation. |
| CROSSCUT-MASS-APPROVAL-1 | Good shape, but does not tie async error suppression to `c3-205`, misses per-PR conditional notification and user-subject delivery. |
| CROSSCUT-SLACK-APPROVAL-1 | Correct flow-entry idea, but grounding is thin, `c3-205` side effects are not explicit, and one Slack config caveat is unsupported. |
| CROSSCUT-STEP-ADVANCE-1 | Near pass, but grounding is too assertive, first-step vs later-step notification behavior is incomplete, and dispatch/channel facts are missing. |
| NATS-1 | Correct JWT-resolver distinction, but misses affected topology, websocket-sync ADR, permission/resolver failure chain, and verification path. |
| PROPERTY-CONFIG-BLAST-RADIUS-1 | Names blast radius, but overstates dependent flow behavior and does not separate direct prefix consumers from reverse-graph dependents. |
| PROPERTY-FILE-IDEMPOTENCY-1 | Captures idempotency, but adds unsupported caveats and does not fully state ZIP partial-state boundary or prove no-rule fact from evidence. |
| PROPERTY-TRANSPORT-SYNC-COUPLING-1 | Names coupling, but under-grounds `c3-209`/`c3-4`/`c3-101`, blurs historical ADR status, and compresses `executionId` tracing. |
| UI-1 | Correct refs and owners, but misses concrete layout semantics: sticky headers, facet/BIG grids, empty states, tablet behavior, and pattern rationale. |

The important failure mode: term-complete answers can still be shallow. They cite IDs and say the right property name, but do not prove the causal chain or tie each claim to the exact fixture evidence. The judge catches asserted-not-grounded claims, missing dependent/observer sides, unsupported caveats, and subtly wrong overclaims that text matching cannot see.

## Depth Improvements Needed

To score well semantically, the C3 skill needs stronger depth requirements, not more keyword requirements:

1. Cross-cut and property answers must include an explicit chain: action owner -> mutation/state owner -> mechanism/ref -> dependent/observer -> emergent property -> failure boundary.
2. Grounding must bind each material claim to a cited read/graph/search result. Evidence commands alone should not count.
3. Reverse graphs must separate direct consumers from indirect dependents, and answers must not assign behavior to every graph neighbor.
4. Historical ADRs must be labeled as current, superseded, or historical context, so old decisions are not implied to be active.
5. Caveats must be evidence-backed. Generic fixture drift or "no rules" claims need supporting `list`, `search`, or case-grounded evidence.
6. Change-usefulness needs concrete checks: owner files/entities, config/permission/runtime checks, sync/notification assertions, and failure-mode verification.
7. UI/layout answers need pattern semantics, not just ref lists: which ref owns which behavior, what breaks if the pattern changes, and how to verify consistency.

## Self-Judge Caveat

The judge was run through the local `judge.py` harness, which invokes a separate `codex exec --ephemeral --sandbox read-only` reviewer with the rubric and schema. That gives separation from the answer-generation pass, but it is still same-vendor/self-judge biased and not a truly independent evaluator.

Recommendation: keep these JSON verdicts as a useful strict-reviewer baseline, but rerun the same prompts through an independent model or human reviewer before using the numeric scores as final eval truth.

## Evidence

- Deterministic score: `374/374`.
- Judge result files: `research/eval/skill-eval/runs/judge-cold-baseline/judge/*.json`.
- Harness: `research/eval/skill-eval/judge/judge.py`, `judge-rubric.md`, `judge-schema.json`.
