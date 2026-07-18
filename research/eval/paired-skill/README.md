# Private paired-eval inputs

This directory contains public templates only. Copy them to a private path,
fill them there, and pass those paths to `scripts/paired_skill_eval.py`.

- `cases.example.jsonl`: synthetic case shape. Replace prompts with approved
  initiatives or historical pre-ADR questions.
- `pricing.example.json`: pricing shape. Replace the model id and prices with a
  current dated snapshot, then set `live_allowed` to `true`.

Do not add filled project cases, gold answers, scoring notes, or raw run output
here. Durable run results must pass the generic allowlist in the harness.

The private Sol-high scorer is responsible for blind, sufficient, correct evaluation.
It receives `{answer}`, `{cases}`, and `{case_id}`, but never the treatment
condition. It must return:

```json
{
  "quality_score": 4.2,
  "correctness_score": 4,
  "trace_completeness_score": 4,
  "reasoning_depth_score": 4,
  "grounding_score": 4,
  "no_hallucination_score": 5,
  "change_usefulness_score": 4,
  "passed": true,
  "independent_review_count": 1,
  "deterministic_evidence_count": 1,
  "scoring_cost_usd": 0.04
}
```

`scoring_cost_usd` is included in both the per-run and total USD walls. Agent
cost remains in `cost_usd`; the retained all-in value is `total_cost_usd`.

Treatment runs are scored only after the usage proxy records successful C3
route, graph, and evidence categories. Control runs must record zero C3 use.
Only category totals, category success counts, and the treatment instruction
hash/layer cross the public retention boundary; commands, arguments, paths,
ids, and outputs stay private.

When collection is deferred, keep the answer manifest and raw answers in a
private directory until scoring is complete, then run the generic post-scorer:

```bash
python3 scripts/paired_skill_postscore.py \
  --manifest /private/answers.jsonl \
  --cases /private/cases.jsonl \
  --score-command '/private/score.py {answer} {cases} {case_id}' \
  --output /private/scores.jsonl \
  --study-id STUDY-001
```

The post-scorer verifies every answer hash before invoking the private scorer,
requires one Sol-high review receipt and one deterministic evidence receipt,
and emits only generic score rows. Delete the private prompts, answers, reviewer
packets, and transcripts after the score ledger is sealed.

To make the paired estimate, pass the sealed generic ledger to the analyzer:

```bash
python3 scripts/paired_skill_analyze.py \
  --input /private/scores.jsonl \
  --study-id STUDY-001 \
  --seed 1 \
  --bootstrap 10000 > /private/analysis.json
```

It fails closed unless every arm is scored, reviewed, evidenced, and paired with
the same run settings. The report contains only arm counts, aggregate quality
deltas, family-level aggregates, efficiency means, and a deterministic bootstrap
interval. It does not emit case ids, prompts, answers, or transcripts.
The report marks studies below 20 unique held-out cases as below the
confirmatory minimum, even when repeated trials create more pairs.
It also marks the frame incomplete until the bootstrap quality-effect
half-width is at most `0.25`.
`protocol-v5-freeze.json` is the generic readiness receipt for the portable edge-source-truth
freeze. It intentionally excludes target names, paths, prompts, answers, entities, and raw model
output. Its route-hit count is a preflight metric, not evidence that the treatment beats baseline.

`protocol-v6-sol-low-br-101.json` is the first valid incremental paired-run receipt. It retains
only generic metrics. One observation can support the blast-radius hypothesis, but cannot establish
a general causal advantage or replace provider-confirmed independent review.

`protocol-v7-three-case-loop.json` closes the first three-case DKR loop. It records the mixed
direction, independent-review disagreement with deterministic scoring, efficiency reads, walls, and
limitations. It is preliminary evidence, not a paper-grade causal claim.

The completed nine-pair loop is recorded in
[`protocol-v8-paper-loop.json`](protocol-v8-paper-loop.json). It retained 18 valid
arms across three case families; its planning reserve remained below the study
ceiling, while actual OpenAI dollar billing was not reported. The skill
condition used fewer tokens, but the two blind judges agreed on only 4 of 9 pair
preferences. The predeclared 80% reliability gate therefore failed. Treat the
effect estimates as DKR evidence for rubric calibration, not as a paper-grade
causal quality claim.

Judge calibration is preserved as a sequential record:

- `protocol-v9-judge-calibration-preregistration.json` and its amendment locally
  record the explicit pick rule before successful calls. The chronology receipt
  is filesystem evidence, not external immutable preregistration.
- `protocol-v9-judge-calibration-result.json` records 5/9 agreement, one v9-rule
  adjudication, and three legacy secondary outcomes. The reliability and
  procedural gates failed.
- `protocol-v10-confirmation-preregistration.json` defines the final high-quality
  confirmation attempt. Its result records a call-cap overrun and an unplanned
  no-output Opus failure; Luna was not run.
- `protocol-v10-study-ledger.json` accounts for $8.7374 of the $10 ceiling while
  separating provider-reported amounts from Codex token-price estimates. Actual
  Codex billing remains unknown.

`protocol-v11-completion-audit.json` checks every original objective requirement.
It records why the current study cannot be called paper-grade and why another
judge pair over the already-observed answers would not repair the confirmatory
design.

Protocols v12 through v17 record the held-out successor attempt. V17 produced
18 valid Sol-low arms across three new case families with zero arm deviations.
Its first primary judge exhausted the call budget without output, so the
confirmatory study stopped. `protocol-v17-result.json` and
`protocol-v17-successor-study-ledger.json` retain that failure without model
substitution or post-hoc reruns.

`protocol-v18-c3-development-freeze.json` converts only V17's aggregate learning
into a generic C3 improvement candidate. It fixes reverse-child blast-radius
traversal and makes sweep closure, evidence, isolation, and unknowns explicit.
This is a frozen development artifact, not evidence of treatment lift; the next
confirmatory study requires fresh held-out cases.

`protocol-v6-frame-terminal.json` closes the first incremental frame. It records accepted and
invalid generic observations, the spend wall, limitations, and the decision not to admit expensive
models. It contains no private case content or raw output.
