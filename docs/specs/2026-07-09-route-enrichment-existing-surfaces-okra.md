# Route Enrichment Existing Surfaces OKRA

Status: verified implementation
Started: 2026-07-09

## Objective

Enrich existing C3 surfaces with route hints without adding a new user-facing primitive.

Metric: `existing_surface_enrichment_count`

Target:

```text
existing_surface_enrichment_count >= 3
```

Current implementation surfaces:

| Surface | Enrichment | Evidence |
| --- | --- | --- |
| `c3 search` | Existing result rows carry route facts, graph neighbors, eval anchors, lanes, drift labels, and route hash. | `go test ./...`; `c3x search ...` sample |
| `c3 graph` | Existing graph JSON nodes carry route facets. Human text gets a compact route line when useful. | `go test ./...`; `c3x graph ... --json` sample |
| `c3 check` | Existing check warns when eval `code:` anchors match no files. | `go test ./...`; `c3x check` |

## Anti-Goals

| Anti-goal | Metric | Type | Read method |
| --- | ---: | --- | --- |
| Do not add a primitive | `new_user_facing_primitive_count == 0` | tripwire | `TestRouteEnrichmentDoesNotAddPrimitiveCommands` |
| Do not claim search ranking improved | `search_ranking_claim_count == 0` | tripwire | docs/code review; no ranking metric changed |
| Do not replace eval/lookup | `eval_lookup_replacement_count == 0` | tripwire | route helper reads eval bindings; lookup/eval commands remain |
| Do not hash noisy code detail | `noisy_route_hash_basis_count == 0` | tripwire | route hash basis excludes full file content and line numbers |
| Do not make stale anchors silent | `stale_anchor_warning_count >= 1` in fixture | tripwire | check test with missing eval code anchor |

## DKR Checkpoints

| DKR | Decision target | Checkpoint | Confidence |
| --- | --- | --- | ---: |
| DKR-1 Primitive boundary | Decide whether to add `trace`/`route` command | Do not add commands. Enrich current outputs. | 0.95 |
| DKR-2 Insertion point | Decide where route data belongs | Search already has `Context`; graph already has node shape; eval bindings already own code anchors. | 0.9 |
| DKR-3 Drift prevention | Decide how frozen facts learn about stale code anchors | `c3 check` warning is the right first step; eval remains the verdict surface. | 0.85 |

Candidate CKRs and candidate PKRs are not promoted until the orchestrator accepts the supporting
DKR learning checkpoint.

## CKRs

| CKR | Metric | Target | Current |
| --- | --- | ---: | ---: |
| CKR-1 Existing-surface coverage | `existing_surface_enrichment_count` | `>= 3` | `3` |
| CKR-2 Primitive boundary | `new_user_facing_primitive_count` | `0` | `0` |
| CKR-3 Drift warning | `stale_anchor_warning_count` | `>= 1` | `1` |

## PKRs

| PKR | Linked CKR | Output | Done read |
| --- | --- | --- | --- |
| PKR-1 Add internal route helper | CKR-1 | `cli/cmd/route_enrichment.go` | Done: `go test ./...` |
| PKR-2 Enrich search rows | CKR-1 | `SearchResultRow.Route` and compact `route` column | Done: search sample emits `route` column |
| PKR-3 Enrich graph nodes | CKR-1 | `graphNode.Route` | Done: graph sample emits node `route` facets |
| PKR-4 Warn on stale eval anchors | CKR-3 | `checkEvalCodeAnchors` | Done: stale-anchor test and acountee replay surface warnings |
| PKR-5 Guard no new primitive | CKR-2 | command-registry test | Done: command-registry test and OKRA checker report `new_user_facing_primitive_count: 0` |

## Verification Reads

| Read | Result |
| --- | --- |
| `go test ./...` from `cli/` | pass |
| `python3 scripts/check_trace_graph_okra.py` | pass; `new_user_facing_primitive_count: 0` |
| `python3 scripts/check_trace_graph_acountee_eval.py` | pass; external target mutated `0` files |
| `python3 scripts/check_trace_graph_rag_route_quality.py` | pass; route quality delta `0.7863`; search ranking delta `0.0` |
| `C3X_MODE=agent bash skills/c3/bin/c3x.sh check` | pass; `total: 41`, `ok: true` |

## Three Anti-Goal Eval Points

| Point | Trace |
| --- | --- |
| Admissibility before acting | Change is scoped to existing command outputs and check warnings. |
| Direct read after acting | Focused Go tests exercise `search`, `graph`, `check`, and command registry. |
| Paired with objective | Surface count reaches target while new primitive count stays `0`. |

## Human-Only Frame

The human owns these calls:

- Decide whether route facets should be shown in more existing surfaces.
- Decide whether stale eval anchors should become errors later.
- Decide whether product wording should retire the Trace Graph pilot name entirely.
