# C3 skill-effectiveness eval

Round 1 base eval. This is not a gold benchmark. It is a grounded starting
point for measuring whether an agent using `skills/c3/SKILL.md` and
`skills/c3/references/` can answer architecture questions about a real project.

The scored artifact is the agent answer, including the evidence commands it
claims to have run. Conceptual questions should start with `c3 search`, then use
targeted `read`, `lookup`, `graph`, and `schema` commands to confirm the answer.

## Fixture

Path: `research/eval/skill-eval/fixtures/acountee/.c3`

Source: `~/dev/acountee/.c3`, snapshotted on 2026-06-09.

Included: canonical markdown docs and `code-map.yaml`.

Excluded: `c3.db`, because it is a disposable local cache. Rebuild it locally
with `c3x check` when running the eval, then remove it before committing.

Fixture caveat: this is a real but imperfect base snapshot. Local C3 can rebuild
and query it, but `check` reports the current acountee documentation drift:

```text
Rebuilt local C3 cache from canonical .c3/
total: 91
issues: warning c3-0 missing required column "Boundary" in table: Containers
issues: warning c3-0 missing required column "Status" in table: Containers
issues: warning c3-0 missing required column "Responsibilities" in table: Containers
issues: warning c3-0 missing required column "Goal Contribution" in table: Containers
ONLY_IN_TREE code-map.yaml
sync check failed: canonical markdown drift detected
```

Do not repair the fixture inside this eval. Its imperfection is part of the
round-1 base.

## Local command

Always run C3 through the local wrapper from this repository:

```bash
C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 <cmd>
```

The wrapper needs a rebuilt cache for this fixture:

```bash
C3X_MODE=agent bash skills/c3/bin/c3x.sh --c3-dir research/eval/skill-eval/fixtures/acountee/.c3 check || true
```

## Contents

| File | Role |
| --- | --- |
| `cases/acountee-round1.md` | Questions, expected elements, evidence snippets, one sample skill-guided answer, and inspection. |
| `rubric.md` | Human-readable scoring criteria derived from the sample answer inspection. |
| `rubric.jsonl` | Machine-scoreable seed in the existing `research/eval/rubric.jsonl` style. |
| `fixtures/acountee/.c3/` | Portable C3 snapshot for the real acountee project. |

## Candidate answer contract

Each eval prompt should ask the agent to include:

- `Evidence commands`: exact C3 commands run, in order.
- `Answer`: concise architecture answer with exact entity ids.
- `Grounding`: why cited refs/rules apply.
- `Caveats`: known fixture or evidence limits.

This lets the rubric score both answer quality and skill behavior, especially
whether conceptual discovery started with `c3 search`.

## Round-1 case set

- `AUTH-1`: How is authentication handled and what governs it?
- `NATS-1`: What breaks if I change NATS websocket authentication?
- `ADMIN-1`: What owns administrator features for users, teams, audit, and approval configuration?
- `APPROVAL-1`: Where does approval workflow live and what governs changes to approvals?
- `UI-1`: How should invoice and payment request screens stay consistent across detail and list layouts?

## Iteration loop

1. Add or refresh a fixture snapshot under `fixtures/<project>/.c3`, excluding
   `c3.db`. Note whether the source is clean or imperfect.
2. Pick 3-5 user-shaped architecture questions. Prefer conceptual or paraphrase
   questions where `search` should help more than title matching.
3. For each case, run the skill-guided flow: `search` first, then targeted
   `read`/`graph`/`lookup`/`schema`. Record command evidence and exact ids.
4. Write one sample answer as an agent using the skill would answer. Inspect it
   against ground truth before changing the rubric.
5. Add only rubric criteria that the base result can actually be measured
   against. If a criterion needs transcript data, make the prompt require
   evidence commands.
6. Re-run candidate agents, compare failures to the expected elements, then
   update the C3 skill or references only for observed gaps.

Anti-goal: do not write abstract gold criteria that the base result cannot be
measured against. This eval should validate real skill usefulness, not reward a
rubric invented before seeing a real answer.
