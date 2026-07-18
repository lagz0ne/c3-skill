# Structural retrieval v3

This directory is a generic seven-case contract for measuring pre-change
impact retrieval. It contains no product repository names, paths, prompts, or
answers. The four `v2-*` cases keep the v2 wrong-layer and behavioral
route shapes. Their source-v2 hashes and ordinal mapping are retained in
`benchmark.v3.json` under `translation_metadata`.

The three held-out countercases are deliberately different:

| case | corpus shape | required negative rule |
| --- | --- | --- |
| `counter-multi-peer-owner` | one owner, two parentless neutral peers, one forbidden child decoy | owner and both neutral facts are in scope; forbidden child is not |
| `counter-no-target` | two neutral facts and one forbidden fact, no owner | owner metrics are N/A; preserve neutral context; v3 may `omit` or `flagged` with a generic reason and no forbidden row |
| `counter-bound-route` | one bound owner→anchor route plus an unbound owner/anchor pair | exact route witness earns credit; unsupported unbound rows earn none |

Every entity and fact has one disjoint role binding. The corpus uses all five
roles (`owner`, `neutral`, `forbidden`, `unsupported`, and `unknown`). Neutral
context is bound through `neutral_fact_ids`; unknown and unsupported records
are explicit oracle fields. Duplicate entity IDs in `rows[0:5]` reject before
scoring. Precision and negative counts use unique non-neutral top-five rows.

Route credit requires the exact bound entity, content ID, graph source and
edge, direct entity/content FTS miss IDs, and exact values for `facts`,
`graph`, `lanes`, and `hash`. A non-empty but arbitrary route field does not
earn credit. Direct-FTS probe arrays are part of each route oracle; a boolean
alone is not a miss proof.

## Thresholds and byte wall

The legacy v2 walls remain: owner-recall-at-5 delta ≥ 0.20, structural-owner
precision ≥ 0.80, forbidden count = 0, route recall/MRR non-regression, and
per-case canonical row bytes ≤ 1.05× baseline. The v3 walls are:

| slice | required result |
| --- | --- |
| multi-peer owner | owner recall, neutral context recall, and owner precision = 1.0; unsupported count = 0 |
| no-target | neutral context recall = 1.0; forbidden exposure = 0; policy = 1.0 |
| bound route | route recall, MRR, and field equality = 1.0; unbound exposure = 0 |
| every case | candidate compact-row bytes / same-case fresh B-v3 bytes ≤ 1.05 |

Canonical row bytes cover every returned row, while retrieval, duplicate,
precision, forbidden, unsupported, route, and policy metrics inspect only the
top five. The B-v3 byte baseline is per case, finite, non-zero, and captured
from an unchanged C3 run with the same seven fixtures.

## Baseline gate

`B-v3-baseline.json` is a fresh unchanged-controller capture. The generic
capture script translates each v3 fixture into a synthetic corpus, invokes the
frozen v2 arm once per case with semantic search disabled, and retains only
same-case canonical row-byte counts and hashes. Temporary stores and raw rows
are discarded. `B-v3-baseline-held.json` records the earlier held prerequisite;
it is historical evidence, not the active baseline.

The baseline capture has no candidate source, capability, activation,
authorization, execution, or effect authority. An independent validator must
replay the capture script and accept the seven finite, non-zero byte baselines
before candidate work can start.

The standalone scorer is `cli/tools/structural-search-eval-v3`. This benchmark
does not authorize candidate execution or claim C3/product/paper value.
