# Learnings: eval-quality-20260610 (autoresearch on the c3 skill-eval itself)

Session improved the EVAL (text scorer + LLM judge), not the skill. Meta-benchmark: tiered probes
(gold / stuffed / stuffed-v2 / wrong / real-shallow) over 4 cases; `eval_quality` = mean of six
discrimination components. Baseline 0.625 → final 1.0 on a benchmark hardened twice mid-session.

## Reusable findings

1. **Text-match scorers are fully gameable by construction.** Bare term dumps maxed the original
   scorer (12/12 probes). Per-line dump filtering (≥70% rubric-term coverage → strip line) stops
   term lists, but prose-padded dilution evades any per-line threshold trivially (stuffed-v2: full
   marks, zero lines stripped).
2. **Fixture-attestation is the strongest deterministic anti-gaming signal found.** U8: ≥75% of an
   answer's content vocabulary (4+ chars, rubric terms excluded) must appear in the fixture corpus.
   Gap is wide and case-consistent (genuine 0.85–0.92 vs padded filler 0.57–0.63). Defeating it
   requires writing fixture-grounded prose about the right terms — which converges on doing the work.
   Depth beyond that is undecidable deterministically: wrong-but-fluent answers still pass text
   (3/4) and only the judge catches them. Division of labor: text scorer = shape + grounding floor;
   judge = truth + depth.
3. **LLM judges punish true-but-beyond-excerpt research as "hallucination" unless told otherwise.**
   Gold answers citing real fixture ADRs the case excerpt didn't mention scored no_hallucination=3
   (gate fail). Fix: redefine hallucination as (a) IDs absent from a fixture ID inventory injected
   into the prompt, (b) contradictions, (c) invented guarantees — and route unverifiable-but-plausible
   detail to the grounding dimension. Gold 2/4 → 4/4 pass with wrong-tier still 4/4 fail.
4. **Single-sample LLM verdicts are not trustworthy at decision boundaries.** Measured flip rate
   0.25 on boundary probes (a gold at overall 4.00 flipped pass→fail between identical runs).
   K=3 reviewers with per-dimension median + majority verdict: observed flips 0/8 × 3 reps, mean
   spread 0.237 → 0.094, at 3× token cost. Residual: a probe whose single-reviewer fail probability
   is ~1/3 still has ~1/4 chance of a bad K=3 majority — margin-thin probes (gold at exactly 4.0)
   should be revised or judged at K=5.
5. **Meta-eval loop mechanics that worked:** content-hash judge cache (env-hash over judge+rubric+cases,
   answer-hash per probe) makes text-only experiments free and judge experiments cheap; probe tiers
   with KNOWN intended outcomes turn "is the eval good?" into a measurable classification task;
   when the metric saturates, extend the benchmark with the previous run's own documented evasion
   (each keep-report listed how to defeat itself — that became the next tier).

## Caveats

- All judge verdicts are codex-judging-codex/claude output (same-vendor bias); numbers are a strict-
  reviewer baseline, not ground truth.
- 4 cases × 5 tiers is a small probe set; thresholds (0.70 dump coverage, 0.75 attestation) have
  wide margins here but should be re-validated on a second fixture before being trusted generally.
- wrong_text_pass_rate=0.75 is accepted by design (mechanism truth is the judge's job), tracked
  but unscored.

## Addendum: skill-depth-20260610 (child session — skill vs the hardened eval)

Loop flipped: eval frozen (after parent runs 7-8: unique-prefix id resolution; verdict from
pass rule on aggregated dims, not reviewer majority — votes [pass,fail,pass] had disagreed
with median overall 3.85), skill as the variable. Baseline 9/13 pass, mean 4.215.

One targeted run reached the stop rule: **13/13 pass, mean 4.438 (+0.223)**. The change:
depth-contract bar 8 — *side effects belong to a layer; read the attachment point (flow vs
service vs storage trigger) before claiming any side effect survives a bypass path* — plus
exact-id citation discipline in bar 1. Honest caveats: (a) compound with generation-prompt
emphasis of existing contract bars, so the +0.223 is skill+emphasis, not the diff alone;
(b) one fixture, one generator (claude/fable), same-vendor judge; (c) 13/13 on the training
fixture says nothing about generalization — second-fixture check is the real test.

Meta-lesson: diagnosed-failure-targeted skill edits (one bar per failure class) moved the
judge mean far beyond its measured noise floor (0.094) in a single iteration; speculative
guidance never had to be written.
