# C3 continuous research + eval loop

This directory is the **memory** of the improvement loop. The eval harness
(`scripts/agent_efficiency_eval.py`) measures C3 skill quality; the gate
(`scripts/eval_gate.py`) decides whether a change moved the needle; this
directory remembers the answer so progress is measurable run-over-run.

## Target metric

**Quality pass-rate across the whole eval matrix.** A record passes when:

| Case type | Pass condition |
|-----------|----------------|
| plain (`task_session`, `debug_session`, `system_design_change`, `skill_task_session`) | `exit_code == 0` and `accuracy_score >= 1.0` |
| ADR (`adr_create`, `skill_content_limit_adr`) | above **and** `adr_quality_score >= 0.8` |
| canvas (`canvas_*`) | above **and** `canvas_quality_passed` |

Token efficiency is a **secondary guardrail**: any record at/over the `no_go`
threshold (250k tokens) forces a discard even if quality held. It never, on its
own, makes a change "better".

## Files

| File | Role | Committed? |
|------|------|-----------|
| `baseline.json` | best-known scorecard; the bar each candidate must clear | yes |
| `history.jsonl` | append-only verdict log (one line per gate run) | yes |
| `runs/*.jsonl` | raw eval records per run | no (gitignored) |
| `../learnings/*.md` | durable narrative of what was tried and why it stuck | yes |

## Run it by hand (one cycle)

```bash
# 1. Make a minimal change to the C3 skill / references / CLI (one thing).
# 2. Rebuild only if the CLI changed:  bash scripts/build.sh
# 3. Eval the affected cases (spends tokens — needs claude/codex CLIs):
python scripts/agent_efficiency_eval.py --run \
  --case adr_create --case canvas_prd \
  --output research/eval/runs/$(git rev-parse --short HEAD).jsonl

# 4. Gate it against the baseline:
python scripts/eval_gate.py \
  --candidate research/eval/runs/<hash>.jsonl \
  --label "tighten adr ref guidance"
#   exit 0 -> keep (re-run with --update-baseline to accept as new best, then commit)
#   exit 1 -> discard (revert the change)
```

## Run it autonomously (the loop)

```
Workflow({ name: "c3-research-eval" })
```

The workflow researches improvement ideas, proposes one minimal change at a
time, evals, gates, and keeps or reverts — recording every verdict here. See
`.claude/workflows/c3-research-eval.js`. It only spends tokens when you launch
it. Stops after 3 consecutive discards (per the autoresearch discipline).

## Discipline (from the `autoresearch` skill)

- **One change at a time** — never tune two things in one eval.
- **Never edit the benchmark mid-session** — changing `agent_efficiency_eval.py`
  cases/floors invalidates baseline comparison; do that as its own deliberate
  baseline reset.
- **Stop after 3 consecutive discards** — the idea well is dry; re-research.
- **Record discards too** — a discarded hypothesis is a learning, not a waste.
