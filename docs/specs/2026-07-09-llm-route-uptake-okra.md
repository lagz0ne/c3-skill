# LLM Route Uptake OKRA

Status: verified v2 full matrix after challenge
Started: 2026-07-09

## Objective

Prove whether LLM agents naturally use C3 route enrichment.

Metric: `llm_route_uptake_score`

Target:

```text
llm_route_uptake_score >= 0.7
paired_uptake_delta > 0.0
```

This is not another route-quality replay. The previous replay proves richer context exists. This
goal measures whether agent answers improve when route context is present and the prompt does not
tell the agent to use route fields.

## Eval Shape

```text
same case
  A: search + graph with route stripped
  B: search + route context pack

same neutral instruction
  -> answer JSON with runner provenance
  -> deterministic scorer
```

Answer artifact:

```json
{
  "meta": {
    "agent_id": "route_a",
    "runner_id": "019f45c2-da81-7293-9c59-4ff97f1de989",
    "source_prompt": "/tmp/llm-route-uptake-real-v2/prompts/<case>/route/prompt.md"
  },
  "answer": {
    "first_files": ["path/to/file.ts"],
    "first_facts": ["c3-or-ref-id"],
    "broad_grep_needed": false,
    "owner_claims": ["c3-or-ref-id"],
    "stale_anchor_noticed": false,
    "fix_start": "one concise sentence naming the first inspection step"
  }
}
```

## Metrics

| Metric | Target | Meaning |
| --- | ---: | --- |
| `first_file_precision` | measured | First files named by the agent match expected source anchors. |
| `first_fact_precision` | measured | First C3 facts named by the agent match expected facts. |
| `broad_grep_avoidance` | measured | Agent does not fall back to broad grep/search-the-repo. |
| `wrong_owner_claim_score` | measured | Agent does not name unrelated owner facts. |
| `stale_anchor_noticing` | measured | Agent notices known acountee drift when route context exposes it. |
| `fix_start_quality` | measured | First fix step connects a source file/symbol with an owning fact. |
| `paired_uptake_delta` | `> 0.0` | Route arm beats no-route on the same case; same-agent when IDs pair, otherwise arm means. |
| `runner_count` | `>= 4` | Answer artifacts include independent worker-run provenance, not only filenames. |

## Anti-Goals

| Anti-goal | Metric | Read method |
| --- | ---: | --- |
| Do not prompt-tune agents to use route fields | `route_instruction_leak_count == 0` | Scan instruction block before context. |
| Do not expose hidden eval answer labels | `answer_key_leak_count == 0` | Scan every prompt for scorer-only labels such as `expected_ids`, `source_anchors`, and `readable_ids`. |
| Do not claim search ranking improved | `search_ranking_claim_count == 0` | Checker reports no ranking metric. |
| Do not accept one LLM self-report as proof | `single_llm_truth_acceptance_count == 0` | Scorer requires at least two distinct `meta.runner_id` values per case/arm. |
| Do not mutate acountee during eval | `target_mutation_count == 0` | Prompt generation records visible and ignored git-status digests; scoring requires the guard. |
| Do not claim product auto-discovery | `auto_discovery_claim_count == 0` | This eval still uses expected source anchors for scoring. |

## DKR Checkpoints

| DKR | Decision target | Checkpoint | Confidence |
| --- | --- | --- | ---: |
| DKR-1 Uptake definition | Decide what counts as natural use | Answer quality changes under neutral prompts, not self-report. | 0.85 |
| DKR-2 Prompt boundary | Decide how to avoid route coaching | Instruction block does not mention route fields; route appears only as context data in arm B. | 0.85 |
| DKR-3 Scoring boundary | Decide what proves useful uptake | Score first files, facts, grep avoidance, owner accuracy, drift noticing, and fix-start quality. | 0.8 |
| DKR-4 External fixture | Decide whether acountee remains target | Keep acountee because it has current drift and cross-cutting lifecycle routes. | 0.85 |

Candidate CKRs and candidate PKRs are not promoted until the orchestrator accepts the supporting
DKR learning checkpoint.

## CKRs

| CKR | Metric | Target | Current |
| --- | --- | ---: | ---: |
| CKR-1 Prompt coverage | `prompt_count` | `20` | `20` |
| CKR-2 Instruction leak guard | `route_instruction_leak_count` | `0` | `0` |
| CKR-3 Live answer coverage | `answer_artifact_count` | `>= 40` | `40` |
| CKR-4 Uptake score | `llm_route_uptake_score` | `>= 0.7` | `0.7844` |
| CKR-5 Paired lift | `paired_uptake_delta` | `> 0.0` | `0.1197` |
| CKR-6 Runner provenance | `runner_count` | `>= 4` | `4` |

## PKRs

| PKR | Linked CKR | Output | Done read |
| --- | --- | --- | --- |
| PKR-1 Add uptake checker | CKR-1..CKR-5 | `scripts/check_llm_route_uptake.py` | self-test passes |
| PKR-2 Generate prompt bundle | CKR-1, CKR-2 | `/tmp/llm-route-uptake-real-v2/prompts` | generated 20 prompts |
| PKR-3 Run smoke agent pair | CKR-3 | one acountee case, two agents per arm | done: `acountee-auth-lifecycle` |
| PKR-4 Run full agent matrix | CKR-3..CKR-5 | 10 cases, two agents per arm | done: 40 answer artifacts |
| PKR-5 Score full matrix | CKR-4..CKR-6 | `python3 scripts/check_llm_route_uptake.py --answers ...` | done: v2 full scorer passed |

## Three Anti-Goal Eval Points

| Point | Trace |
| --- | --- |
| Admissibility before acting | Prompt generator strips route from arm A, masks eval answer labels from arm B, and scans instruction blocks for route-use coaching. |
| Direct read after acting | Scorer reads wrapped answer artifacts and rejects missing runner provenance, single-runner arms, leaked route instructions, or leaked answer-key labels. |
| Paired with objective | Success requires route score target plus positive paired delta while anti-goals remain held. |

## Verification Reads

| Read | Result |
| --- | --- |
| `python3 scripts/check_llm_route_uptake.py --self-test` | pass |
| `python3 scripts/check_llm_route_uptake.py --answers /tmp/llm-route-uptake-real-v2/answers --min-agents-per-arm 2` | pass |
| `answer_artifact_count` | `40` |
| `runner_count` | `4` |
| `baseline_uptake_score` | `0.6648` |
| `llm_route_uptake_score` | `0.7844` |
| `paired_uptake_delta` | `0.1197` |
| `paired_delta_case_count` | `10` |
| `paired_delta_source` | `case_mean: 10`, `same_agent: 0` |
| `route_instruction_leak_count` | `0` |
| `answer_key_leak_count` | `0` |
| `single_llm_truth_acceptance_count` | `0` |
| `target_mutation_count` | `0` |
| `auto_discovery_claim_count` | `0` |

## Smoke Read

Case: `acountee-auth-lifecycle`

| Arm | Agents | Mean score | Read |
| --- | ---: | ---: | --- |
| no-route | 2 | `0.4167` | Named auth facts but had weaker file and drift grounding. |
| route | 2 | `0.7916` | Named concrete auth files/facts and surfaced route-visible drift. |

Smoke metrics:

```text
answer_artifact_count = 4
baseline_uptake_score = 0.4167
llm_route_uptake_score = 0.7916
paired_uptake_delta = 0.3749
route_instruction_leak_count = 0
answer_key_leak_count = 0
single_llm_truth_acceptance_count = 0
target_mutation_count = 0
```

## Full Matrix Read

```text
agent_count = 4
runner_count = 4
answer_artifact_count = 40
case_count = 10
baseline_answer_count = 20
route_answer_count = 20
baseline_uptake_score = 0.6648
llm_route_uptake_score = 0.7844
paired_uptake_delta = 0.1197
answer_key_leak_count = 0
```

Interpretation:

Route context was naturally used by the sampled agents under neutral instructions, after hiding
eval-only answer labels. The lift is smaller than v1 but still clears the target. The strongest gains
are on acountee cross-cutting lifecycle questions where route evidence exposes source anchors, drift,
and ownership lanes. One case, `acountee-ui-behavior-collection`, slightly regressed because the
no-route context was already enough for both no-route workers.

This does not prove search ranking improved, product auto-discovery exists, or every route-enriched
answer is better. Those counters remain `0`; the claim is narrower: route-enriched context improves
query reasoning/RAG answer quality on this sampled C3/acountee matrix.

## Challenged V1 Read

The first full matrix at `/tmp/llm-route-uptake-real` scored higher:

```text
baseline_uptake_score = 0.5298
llm_route_uptake_score = 0.9084
paired_uptake_delta = 0.3785
```

That read is superseded. A challenger found that the route prompt exposed eval-only labels such as
`expected_ids`, `source_anchors`, and `readable_ids`, and that scoring relied on filenames rather
than runner provenance. V2 fixed those proof issues and reran all 40 answers.
