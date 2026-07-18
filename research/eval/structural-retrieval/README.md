# Structural retrieval microbenchmark

This benchmark asks one narrow question: can a retrieval candidate surface the
schema, query, and event owners needed for a change-impact answer without losing
known route evidence? It is a cheap candidate filter. It does not prove agent
answer quality or the paired-study objective.

## Frozen inputs

The public fixture has 12 invented cases:

- 8 wrong-layer cases put high-overlap prose ahead of anchored schema, query,
  and event facts under the lexical baseline.
- 4 route cases protect strong command-to-handler paths from regression.

Every structural claim is a positive fact with an exact synthetic anchor and a
fact id in that case's complete allowlist. The scorer never infers a missing
edge, owner, or behavior from absence. Repeated fact ids and document ids are
invalid, and every declared UTF-8 byte count is recomputed.

The baseline manifest freezes the fixture SHA-256, scorer SHA-256, invariant
ranking configuration, and baseline metric values. Any candidate session must
keep those bytes unchanged. A benchmark revision starts a new version and a new
baseline; it is not a candidate burst.

`implementation_commit` is forty zeroes while these benchmark bytes are only in
the worktree. It does not claim that the current Git `HEAD` contains them. The
fixture, scorer, manifest, and result SHA-256 values bind the bytes used by a
run; a later portable freeze must replace the sentinel with the commit that
actually contains those same bytes.

## Metrics and keep gate

The deterministic scorer reports owner recall@1/3/5, MRR, wrong-layer MRR,
route recall@5, route MRR, structural-owner precision, false structural claims,
total context bytes, and top-5 context bytes for every case.

A candidate is kept only when all clauses hold:

- `owner_recall_at_5(candidate) - owner_recall_at_5(baseline) >= 0.20` on the
  eight wrong-layer cases;
- structural-owner precision is at least `0.80`;
- false structural claims remain `0`;
- route recall@5, route MRR, and wrong-layer MRR do not regress;
- each case uses no more than `1.05 * baseline_case_bytes` of top-5 context;
- the candidate changes exactly `ranking.mode`, preserves every invariant, and
  cites the frozen baseline manifest hash.

One failed clause means discard. Metrics are not averaged across a failed wall.
Three consecutive discards in one approach family stop candidate execution and
hand control back to discovery.

`NC-STRUCTURAL-ENTITY-DUMP` is the negative control. It places every document
with a structural fact before prose. Its recall looks useful, but its precision
is below `0.80`, so the gate must reject it.

## Replay

From `cli/`:

```bash
go test ./tools/structural-search-eval

go run ./tools/structural-search-eval run \
  --fixtures ../research/eval/structural-retrieval/fixtures.v1.jsonl \
  --candidate ../research/eval/structural-retrieval/candidate-baseline.v1.json \
  --out /tmp/structural-baseline.json

go run ./tools/structural-search-eval gate \
  --baseline /tmp/structural-baseline.json \
  --candidate /tmp/structural-candidate.json \
  --freeze ../research/eval/structural-retrieval/candidate-baseline.v1.json
```

The `run` command fails closed on fixture, scorer, invariant, schema, anchor, or
byte-count drift. The `gate` command writes a machine-readable verdict and exits
nonzero for a discard.
