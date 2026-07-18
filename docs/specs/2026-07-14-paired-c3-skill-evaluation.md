# Paired C3 Skill Evaluation

## Decision

Measure the C3 skill as a treatment against the same agent without C3. Keep
answer quality and execution cost separate. A cheaper wrong answer is not a win.

The study estimates the paired treatment effect on the existing 1-5 semantic
quality score. The study stops when the 95% paired confidence-interval
half-width is at most `0.25`, after at least 20 and no more than 60 held-out
cases.

## Trial shape

```text
private case + frozen Git HEAD
                |
         randomized pair
         /             \
  without C3          with C3
         \             /
      blind scoring before labels
                |
    generic numeric result rows
       raw workspace deleted
```

Both arms use the same model, task, Git HEAD, timeout, token ceiling, runtime
supervisor, and tool environment. The treatment arm keeps `.c3`, mounts this
checkout's local C3 skill, and receives the C3 treatment instruction. Codex
receives it as `developer_instructions`; Claude receives it through
`--append-system-prompt`. The control arm removes `.c3`, does not mount the
skill, and receives neither provider-level treatment instruction.

Both arms first remove ambient `AGENTS.md`, `CLAUDE.md`, `.agents`, `.claude`,
and `.codex` material from the complete disposable tree. The runner then copies
one concise, C3-neutral repository quickstart to root `AGENTS.md` and
`CLAUDE.md`. The files must be byte-identical, no nested instruction may remain,
and both arms record the same baseline hash. This compares C3 with a normally
prepared agent rather than with an agent deprived of repository guidance.

For the treatment only, the runner appends the same bounded C3 impact block to
both root instruction files after hashing the neutral baseline. It records a
separate treatment hash and provider instruction layer. The block requires one
successful route, graph, and evidence command before narrow source checks.
Missing a successful route, graph, or evidence command, control-arm C3 use, or
an instruction hash or layer mismatch invalidates the run before scoring.
No-result exploratory C3 calls remain counted but do not invalidate a run after
all three required categories have at least one success.

The source repository is never mounted. The runner exports its committed HEAD
into a disposable writable workspace. Dirty and untracked source files are not
part of the trial.

## Quality contract

Score every answer from 1 to 5 on:

| Dimension | Weight |
|---|---:|
| Correctness | 25% |
| Trace completeness | 20% |
| Reasoning depth | 20% |
| Grounding | 15% |
| No hallucination | 10% |
| Change usefulness | 10% |

Pass requires overall `>= 4.0`, correctness `>= 4`, no hallucination `>= 4`,
and no dimension below 3. A score record must report at least two independent
reviews and at least one deterministic evidence check. The scorer never
receives the arm label. This is label blinding, not guaranteed treatment
blinding: an answer may reveal that it used C3 through its evidence or wording.
The paper must report that limit.

## Budget contract

The planner selects the cheapest `eligible: true` and `live_allowed: true`
model in a local frozen pricing file. The same selected model is used for both
arms.

Defaults:

| Wall | Threshold |
|---|---:|
| Pilot runs | `<= 24` |
| Total pilot cost | `<= $5.00` |
| Estimated or actual cost per run | `<= $0.50` |
| Tokens per run | `<= 250000` |
| Time per run | `<= 900 seconds` |
| Tool calls started | `<= 6`; terminate when a seventh starts |
| Combined stdout and stderr | `< 524288 bytes` before process termination |

`plan` never calls an agent. `run` still refuses unless `--run` is present.
Live execution also refuses missing pricing, pricing older than 30 days, an
example-only model, or a projected budget breach. An actual breach stops later
runs. A provider-neutral process supervisor terminates either CLI at the same
tool-event, output-byte, or wall-time wall and emits a generic guard receipt.
Post-run token and cost reads remain fail-closed because neither provider can
guarantee an exact token stop before the current model turn completes.

## Retention boundary

Raw prompts, answers, transcripts, workspaces, entity ids, paths, diffs,
repository names, and reviewer prose live only in a temporary private
directory. Scoring happens before that directory is deleted.

The durable JSONL schema permits only opaque study/case ids, task family, arm,
agent/model, numeric quality dimensions, pass/fail, tokens, turns, tool calls,
elapsed time, cost, price version, protocol-deviation enums, runner version,
and the treatment skill hash. Unknown fields fail closed.
Generic C3 command-category counts and instruction hashes/layers are also
retained; command arguments, paths, entity ids, and output are not.

## Private inputs

Copy the templates from `research/eval/paired-skill/` outside this repository.
Do not commit the filled files.

Each case is one JSONL object:

```json
{"case_id":"BR-001","family":"blast_radius","prompt":"What is affected if the configured subject namespace changes?"}
```

Use `BR-NNN` for blast-radius cases and `PI-NNN` for pre-initiative
change-unit cases. The prompt may contain project details because the case file
is private; its opaque id is the only case data retained.

The scoring command receives three placeholders and must print one JSON object:

```text
/private/score.py {answer} {cases} {case_id}
```

Required score-only fields are the seven quality fields, `passed`,
`independent_review_count >= 2`, and `deterministic_evidence_count >= 1`.

## Commands

No-spend plan:

```bash
python3 scripts/paired_skill_eval.py plan \
  --repo ~/dev/acountee \
  --cases /private/c3-eval/cases.jsonl \
  --pricing /private/c3-eval/pricing.json \
  --repeat 2
```

Live pilot, only after cases and the scorer are frozen:

```bash
python3 scripts/paired_skill_eval.py run --run \
  --repo ~/dev/acountee \
  --cases /private/c3-eval/cases.jsonl \
  --pricing /private/c3-eval/pricing.json \
  --score-command '/private/c3-eval/score.py {answer} {cases} {case_id}' \
  --results research/eval/runs/STUDY-001.jsonl \
  --study-id STUDY-001 \
  --repeat 2 \
  --budget-usd 5.00 \
  --max-cost-per-run-usd 0.50 \
  --auth session
```

The retained `results` path must contain only generic rows. It remains invalid
input for a paper until the paired statistics and confidence-interval reporter
are completed after the first pilot estimates variance.

## Standing walls

Generic-only retention, no treatment leakage, arm parity, source immutability,
semantic review, no single-LLM truth, frozen cases/rubric, budget ceilings, and
OKRA integrity remain standing after each run. Only a human can retire them at
merge, handoff, or decommission.
