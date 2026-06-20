# Design — the eval-spec on a fact

> Open #1 from [draft-match-op-concept.md](draft-match-op-concept.md). The per-fact thing that
> replaces `codemap:`. It is the **mutable lens** on a **frozen claim** — the binding that lets
> two lanes move on their own clocks. Grounded in today's `.c3/code-map.yaml` (sidecar, fact-id →
> globs, never frozen — the one mutable part).

## What it is
A fact's eval-spec answers, for that one frozen claim: *which uncontrolled external does this
govern, and how do we decide it still holds?* It is the per-fact rubric — and like a harness
topic, it is **hybrid**: a machine half and a judge half.

```
eval-spec  =  op (external-kind)
           +  locator (declared surface: include / exclude)
           +  mechanical line   ← c3x computes this        (the runbook)
           +  judgement rubric  ← a judge reads this, c3x only surfaces it   (rubric-notes)
```

The mechanical line alone, with no rubric, **is the codemap** — a fact that only needs "the
code it names still resolves" carries an op:code / mode:mechanical spec and nothing else. Add a
rubric and the same fact also asks a behaviour question. Generality is free: change `op`.

## Frozen vs mutable — the load-bearing split
- **Frozen:** the claim (the fact body, sealed by `c3-seal`). The thing being measured cannot
  move without a change-unit. This is the fixed point on the *claim* side.
- **Mutable:** the eval-spec. Re-point the lens (code moved dirs), refine inclusion, add a
  falsifier — none of that is a change-unit, exactly as `code-map.yaml` is edited freely today.
- **Usually free:** the external (living code). The eval-spec does **not** freeze it.

This is "make divergence impossible, not fixable" applied correctly: you can always re-aim the
lens, but you can never move the measured claim while pretending the external still matches it.
The verdict stamp (below) records *which* lens was used, so a re-aim is visible in history.

**But a verdict is a one-off, not a guarantee.** Freezing the claim makes the *claim* side
stable; it does **not** make the eval durable, because the external is still free. An eval is
true/false *at the instant it ran*, solid only as a statement about the exact `(claim-merkle,
external-state)` pair it stamped — see the one-off / solidity principle in
[draft-match-op-concept.md](draft-match-op-concept.md). A code-op eval is therefore inherently a
snapshot you re-run on CI; only ops whose external is itself pinned (released artifact, frozen
parallel doc) yield a durable verdict.

## Shape — the code op covers the codemap exactly
Today's `c3-107` codemap entry:
```yaml
# .c3/code-map.yaml
c3-107:
  - cli/internal/store/**
  - cli/tools/semantic-assets/**
```
Becomes a mechanical-only eval-spec — same binding, no new burden:
```yaml
# .c3/eval/c3-107.yaml
op: code
mode: mechanical                 # resolves | structural — no rubric ⇒ pure codemap-equivalent
locator:
  include: [ cli/internal/store/**, cli/tools/semantic-assets/** ]
  exclude: [ cli/internal/store/**_test.go ]   # per-fact carve-out (was code_map_excludes)
```
Add a behaviour question → add `mode: judgement` + a rubric file (the harness `rubric-notes`
shape, beside the spec, **not** frozen into the fact):
```yaml
op: code
mode: judgement
locator: { include: [ cli/internal/store/** ] }
rubric: c3-107.rubric.md         # falsifier table a judge reads; c3x surfaces, never scores
```

## Shape — the other ops prove "general from the start"
```yaml
# docs op — fact in this lane must agree with a fact in a parallel C3 doc set
op: docs
mode: judgement
locator:
  include: [ ../service-b/.c3/** ]
  match:   c3-204                 # which fact(s) in the other lane to equate against
rubric: ...rubric.md              # "do both lanes still describe the same contract?"
```
```yaml
# artifact op — fact governs a build output / generated spec
op: artifact
mode: mechanical                  # the artifact's structure/hash conforms to the claim
locator: { include: [ openapi.json ] }
# (or mode: judgement + rubric: "does the generated spec still match the prose contract?")
```
`op` is the single extensibility point. The topic/rubric/verdict machine is op-agnostic; a new
op = a new `op:` value + a resolver for that external-kind. Nothing else moves.

## How c3x reads it — helper, never gate
- **mechanical line:** c3x fully computes a verdict (`holds` / `drift`) — glob resolves, file
  exists, structural signal matches. This is the deleted `codemapGlobResolves`, reborn as a
  rubric line instead of a `check --strict-codemap` gate.
- **judgement rubric:** c3x **surfaces** it — prints the falsifier table + the evidence it
  gathered (the runbook) — and emits `verdict: needs-judgement`. A human / agent / CI fills the
  verdict. There is no `c3x` exit code that gates `change apply` on conformance.

## The verdict — what "CI-ready" emits
One record per fact, per evaluation — reproducible and lens-aware:
```json
{
  "fact": "c3-107",
  "fact_root_merkle": "db80187…",   // the frozen claim that was measured
  "eval_spec_hash": "9c1f…",        // which lens (so a re-aim is visible)
  "external_state": "git:9f3a1c2",  // which external state it was measured against
  "op": "code", "mode": "mechanical",
  "verdict": "holds | drift | hole | needs-judgement",
  "evidence": [ "cli/internal/store/store.go ✓", "cli/internal/store/import.go ✓" ]
}
```
CI consumes the array; a human reads the `needs-judgement` rows. The fact stays pure — the
stamp lives only in the verdict, never mutates the frozen body.

## Coverage — computed across all specs of an op
- **Inclusion** = each spec's `locator.include` (the declared surface, not the whole repo).
- **Exclusion** = each spec's `locator.exclude` (per-fact).
- **DRIFT** (per fact) = a declared external is gone/changed → mechanical fails, or judge says no.
- **HOLE** (per op) = an external in the op's universe matched by *no* fact's include. Computed
  by union-ing every spec's include for that op and diffing against the op's surface. Reported,
  never gated — the harness's "orphan left standing", generalised.

## Physical layout (proposal)
```
.c3/eval/<fact-id>.yaml          # the mechanical binding (op, locator, mode)  — replaces code-map.yaml
.c3/eval/<fact-id>.rubric.md     # optional judge prose — only when mode: judgement
```
A per-fact directory entry *is* a mini harness-topic: binding (machine) + rubric (judge). The
single `code-map.yaml` migrates to `.c3/eval/*.yaml` with `op: code` / `mode: mechanical`.

## Open after this
- The **resolver interface** per op (code = glob; docs = cross-lane fact match; artifact = path
  + structural probe). One small interface, three impls, to be specced with "first ops".
- Whether `mode` is `mechanical | judgement | both` (a fact can want both halves).
- Naming: `.c3/eval/` vs `.c3/conform/`; `op` vs `against`.
