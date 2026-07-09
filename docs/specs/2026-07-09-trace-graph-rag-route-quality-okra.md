# Trace Graph RAG Route Quality OKRA

Status: pilot objective achieved after independent challenge
Started: 2026-07-09

## Objective

Prove whether **Trace Graph improves query/RAG route quality** after raw C3 search.

This run does not claim raw search ranking improved.
This run does not claim automatic source-anchor discovery. It is a source-anchored replay fixture:
expected files and symbols are gold labels used to test whether Trace Graph context improves routing
after search.

```text
baseline:
  c3 search only

candidate:
  c3 search
    -> Trace Graph context pack
    -> facts/custom docs
    -> graph neighbors
    -> files
    -> symbols/tests
    -> drift signal when present
```

Metric: `rag_route_quality_delta`

Target:

```text
rag_route_quality_delta >= +0.25
```

Formula:

```text
rag_route_quality_delta = trace_route_quality - baseline_route_quality
```

Current read:

| Metric | Target | Observed | Status |
| --- | ---: | ---: | --- |
| `total_query_count` | `>= 8` | `10` | met |
| `external_query_count` | `>= 4` | `6` | met |
| `cross_cutting_query_count` | `>= 4` | `6` | met |
| `baseline_route_quality` | measured | `0.2138` | read |
| `trace_route_quality` | measured | `1.0` | read |
| `rag_route_quality_delta` | `>= 0.25` | `0.7863` | met |
| `min_case_delta` | `> 0` | `0.75` | met |
| `search_ranking_delta` | `0.0` | `0.0` | held |
| `auto_discovery_claim_count` | `0` | `0` | held |
| `trace_pack_count` | `10` | `10` | met |

Replay command:

```bash
python3 scripts/check_trace_graph_rag_route_quality.py
```

Latest replay artifact:

```text
/tmp/trace-graph-rag-route-acountee.9cul15w2/route-quality/trace-packs.json
```

## Anti-Goals

| Anti-goal | Metric | Type | Current read |
| --- | ---: | --- | --- |
| Do not pretend raw search ranking improved | `search_ranking_claim_without_hitk_mrr_count == 0` and `search_ranking_delta == 0.0` | tripwire | held |
| Do not overfit one query | `total_query_count >= 8` and `external_query_count >= 4` | tripwire | held |
| Do not ignore cross-cutting papercuts | `cross_cutting_query_count >= 4` | tripwire | held |
| Do not hide target drift | `missed_known_drift_count == 0` | tripwire | held |
| Do not mutate external targets | `target_mutation_count == 0` | tripwire | held |
| Do not turn Trace Graph into proof | `proof_or_certification_claim_count == 0` | tripwire | held |
| Do not claim automatic anchor discovery | `auto_discovery_claim_count == 0` | tripwire | held |

Anti-goal evidence:

- The checker reports `search_ranking_delta = 0.0`.
- The checker compares visible and ignored git status for `/home/lagz0ne/dev/acountee`.
- Acountee target drift is expected and must be surfaced for every acountee case.
- The generated context packs use graph-shaped routing evidence, not correctness claims.
- The checker uses expected source files and symbols as gold labels, so automatic discovery remains
  out of scope.
- Hash basis excludes full file content, line numbers, whole function bodies, and formatting.

## DKR Checkpoints

| DKR | Decision target | Checkpoint | Confidence |
| --- | --- | --- | ---: |
| DKR-1 Metric split | Decide whether to measure search or route quality | Keep search ranking separate. This run measures route quality only. | 0.9 |
| DKR-2 Cross-cutting set | Decide whether acountee should remain the external fixture | Use acountee because it has frontend/backend boundaries, auth lifecycle, invoice lifecycle, e2e, NATS sync, theming, and current C3 drift. | 0.85 |
| DKR-3 Baseline shape | Decide the baseline route | Baseline is `c3 search --no-semantic --limit 12` only. It can score expected C3 ids, not files/symbols/drift. | 0.85 |
| DKR-4 Candidate shape | Decide the candidate route | Candidate emits a Trace Graph context pack with facts/custom docs, graph neighbors, lookup anchors, source symbols/tests, and drift. | 0.85 |
| DKR-5 Discovery boundary | Decide whether this proves product discovery | Keep automatic anchor discovery out of scope until a product route discovers files/symbols without gold labels. | 0.9 |

Candidate CKRs and candidate PKRs are not promoted until the orchestrator accepts the supporting
DKR learning checkpoint.

## CKRs

| CKR | Metric | Target | Current |
| --- | --- | ---: | ---: |
| CKR-1 Route quality lift | `rag_route_quality_delta` | `>= 0.25` | `0.7863` |
| CKR-2 Query breadth | `total_query_count` | `>= 8` | `10` |
| CKR-3 External pressure | `external_query_count` | `>= 4` | `6` |
| CKR-4 Cross-cutting coverage | `cross_cutting_query_count` | `>= 4` | `6` |
| CKR-5 Honest search boundary | `search_ranking_delta` | `0.0` | `0.0` |
| CKR-6 Honest discovery boundary | `auto_discovery_claim_count` | `0` | `0` |

## PKRs

| PKR | Linked CKR | Output | Done read |
| --- | --- | --- | --- |
| PKR-1 Add route-quality checker | CKR-1..CKR-5 | `scripts/check_trace_graph_rag_route_quality.py` | done |
| PKR-2 Add c3-design query cases | CKR-2 | Lookup, eval, search/RAG, agent output | done |
| PKR-3 Add acountee cross-cutting cases | CKR-3, CKR-4 | PR approval, UI behavior, auth, invoice, theming, sync | done |
| PKR-4 Generate context packs | CKR-1 | `/tmp/.../route-quality/trace-packs.json` | done |
| PKR-5 Bound discovery claim | CKR-6 | Source-anchored replay caveat in spec, OKRA, checker | done |

## Query Set

| Case | Repo | Theme | Why it matters |
| --- | --- | --- | --- |
| `c3-lookup-ownership` | c3-design | lookup ownership | Maps file/glob questions to facts, refs, and output helpers. |
| `c3-eval-determinism` | c3-design | eval determinism | Separates conformance evidence from source drift. |
| `c3-search-rag` | c3-design | query reasoning and search | Touches retrieval, graph expansion, semantic assets, and ranking eval. |
| `c3-agent-output` | c3-design | agent output contract | Prevents RAG answers from drifting into wrong command-output format. |
| `acountee-pr-approval` | acountee | frontend/backend approval lifecycle | Crosses UI screen, backend flow, DB approval state, shared schema. |
| `acountee-ui-behavior-collection` | acountee | source-backed UI behavior and tests | Crosses route inventory, screen/flow/state-machine facts, generator, test. |
| `acountee-auth-lifecycle` | acountee | auth lifecycle across frontend/backend | Crosses login, auth guard, server auth resource, e2e auth stub. |
| `acountee-invoice-lifecycle` | acountee | invoice lifecycle across UI/backend/e2e | Crosses UI list/detail, backend invoice flows/service, Lightpanda e2e. |
| `acountee-ui-theming-papercuts` | acountee | frontend components theming papercuts | Crosses UI kit, variant system, route theme boot, screen usage. |
| `acountee-realtime-sync-cycle` | acountee | real-time ownership and cycle | Crosses frontend atom, execution wait, NATS publisher, notification path. |

## Result Read

```text
baseline_route_quality = 0.2138
trace_route_quality = 1.0
rag_route_quality_delta = 0.7863
search_ranking_delta = 0.0
```

Interpretation:

```text
Raw c3 search already found many expected C3 ids.
Trace Graph improved the route by adding files, source symbols/tests, graph context, and drift nodes.
This is route-quality evidence, not search-ranking evidence.
This is source-anchored replay evidence, not automatic source-anchor discovery evidence.
```

## Three Anti-Goal Eval Points

| Point | Trace |
| --- | --- |
| Admissibility before acting | The run only added docs/checkers in c3-design and created disposable `/tmp` fixtures. |
| Direct read after acting | `python3 scripts/check_trace_graph_rag_route_quality.py` reported target mutation count `0` and known acountee drift surfaced. |
| Paired with objective | Objective metric improved while `search_ranking_delta` stayed `0.0`, `auto_discovery_claim_count` stayed `0`, and proof/certification claims stayed `0`. |

## Flags

| Flag | Current state | Evidence |
| --- | --- | --- |
| Cannot | closed | Ten query cases produced route packs. |
| Breaking | closed | No external target mutation; no proof/certification language. |
| Pointless | closed | Route-quality delta exceeded target. |
| Stalled | closed | Broad cross-cutting cases replayed in one checker. |
| Authority drift | closed | Human expanded scope; checker incorporated it without changing the objective into search ranking. |

## Independent Challenge

Verdict: intact.

Accepted boundary:

```text
The completion claim is source-anchored route-quality replay.
It is not product RAG quality, answer quality, search-ranking improvement, correctness proof, or
automatic source-anchor discovery.
```

Residual risk:

```text
The proof/ranking/discovery claim counters are hardcoded zeros, not text scanners.
This is acceptable for this pilot because the docs state the boundary and the challenger checked it.
A productized run should replace those counters with claim scanners.
```

## Human-Only Frame

The human owns these calls:

- Decide whether route-quality lift is enough to productize Trace Graph.
- Decide whether the next run should measure answer quality with human or independent judge labels.
- Decide whether to repair acountee's C3 drift separately.
- Change the route-quality target or query set.
