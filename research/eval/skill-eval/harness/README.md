# C3 blindbox growth harness

This harness runs real Codex, Claude, and Kilo against isolated C3 growth
topics. It is separate from the acountee answer benchmark so the existing
architecture-answer scores stay comparable.

The first topic, `grow-warehouse-system`, checks whether an agent can use the
local C3 skill to grow a system through multiple containers, components,
concepts, feature additions, migrations, and document gardening.

## Generate

Dry-run first:

```bash
research/eval/skill-eval/harness/bin/run-blindbox.sh \
  --agent all \
  --topic grow-warehouse-system \
  --auth session \
  --run-timeout 900 \
  --dry-run
```

Run selected agents:

```bash
research/eval/skill-eval/harness/bin/run-blindbox.sh \
  --agent codex \
  --topic grow-warehouse-system \
  --auth session \
  --run-timeout 900

research/eval/skill-eval/harness/bin/run-blindbox.sh \
  --agent claude \
  --topic grow-warehouse-system \
  --auth session \
  --run-timeout 900

research/eval/skill-eval/harness/bin/run-blindbox.sh \
  --agent kilo \
  --topic grow-warehouse-system \
  --auth session \
  --model kilo/kilo-auto/free \
  --kilo-agent code \
  --run-timeout 900
```

Run every Kilo model currently marked free:

```bash
research/eval/skill-eval/harness/bin/run-blindbox.sh \
  --agent kilo-free \
  --topic grow-warehouse-system \
  --auth session \
  --refresh-models \
  --run-timeout 900
```

## Isolation

The runner uses `bwrap`, clears the environment, uses an empty home, and does
not mount the repository root. It mounts:

- the prompt packet as `/prompt.md`;
- the local C3 skill directory as `/opt/c3/skills/c3`;
- one writable project workspace under `harness/runs/<label>.workspace`;
- the run directory as `/runs` for final answers and sidecars.

Session auth mode copies only the selected CLI's known session credential file
into a temporary auth directory, mounts that copy, and deletes it after the run.

## Review

Review with:

- candidate output from `harness/runs/<label>.md`;
- workspace artifacts from `harness/runs/<label>.workspace`;
- `harness/rubric.md`;
- `harness/review-lenses.md`;
- `harness/topics/grow-warehouse-system/rubric-notes.md`.

Compare agents on:

- score and pass/fail dimensions;
- whether the agent used the local C3 wrapper;
- whether growth happened through rung migration rather than partial facts;
- whether frontend/backend/integration/database boundaries stayed clear;
- whether migrations and document gardening were concrete;
- Kilo free-model spread and repeated failure patterns.
