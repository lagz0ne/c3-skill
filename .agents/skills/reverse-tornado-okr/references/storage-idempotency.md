# Storage and Idempotency

Use this reference when the goal is being run as an automated loop, has side effects, or must survive interruption and resume.

## Store

Persist these records before running the first state-changing move:

- `frame.json`: write-once objective, target, anti-goal, thresholds, owner, and ratification time.
  Include `frame_version`, `frame_hash`, metric contracts, anti-goal coverage review, action
  envelope, and the human approval record.
- `tree.json`: current DKR/CKR/PKR decomposition and worker scopes.
- `results/<idempotency-key>.json`: write-once committed move result.
- `ledger.jsonl`: append-only direct objective and anti-goal readings with `observed_at`,
  `recorded_at`, source, query/report hash, window, value, unit, and freshness status.
- `flags.jsonl`: append-only `cannot`, `breaking`, `pointless`, and `authority_drift` flags with
  lifecycle status and resolution records.

Do not rewrite frame or ledger records to make a run look safer after the fact. Add a new reading or
a human-ratified frame revision instead. A revision is a new immutable frame record with diff,
reason, approver, timestamp, affected metrics, and whether it relaxes any guardrail.

## Idempotency Keys

Use a stable key for every committing move:

```text
<frame-version>/<frame-hash>/<metric-contract-hash>/<action-approval-id>/<tree-node-id>/<move-kind>/<scope-hash>/<input-hash>/<attempt-policy>
```

Include enough input to distinguish materially different side effects. Do not include timestamps, worker ids, random ids, or retry counts unless they change the intended effect.

The key must change when the frame, metric definition, guardrail threshold, authority envelope, or
human approval changes. Reusing a result across frame revisions is allowed only for explicitly
read-only discovery outputs that the orchestrator re-admits under the current frame.

Dry-run propose-cost moves do not need idempotency keys because they do not commit side effects. If a dry-run writes to an external system, it is not a dry-run and must be keyed.

## Dispatch Sequence

1. Read the current ratified frame and refuse to continue if the frame is still only a candidate.
2. Refuse committing moves when required metric readings are stale, missing, or blocked by an
   unresolved flag unless the human has recorded a waiver for this move.
3. Refuse and flag `authority_drift` if the worker or move tries to alter the frame, expand scope,
   bypass approval, relax a threshold, or leave the action envelope.
4. Build the candidate move and run admissibility against the anti-goal.
5. If cost is unknown, run a propose-cost dry-run and admit or veto from that result.
6. Construct the idempotency key for an admitted committing move.
7. If `results/<key>.json` exists, reuse it and do not dispatch the worker again.
8. Dispatch the worker with the frozen scope and key.
9. Write the result once.
10. Read direct objective and anti-goal metrics from source and append to `ledger.jsonl`.
11. Evaluate `cannot`, `breaking`, `pointless`, and `authority_drift` flags.

## Resume Sequence

On restart, load the current frame revision, tree, results, ledger, and flags. Check unresolved
blocking flags and metric freshness before rebuilding the next move. Never infer success from an
existing subtree; use ledger readings only if they satisfy the current metric contract, otherwise
take a fresh direct read.

If the frame changed while a worker was running, do not commit that worker's result automatically.
Record the worker output as evidence, re-run admissibility under the new frame, and either commit
under a new key or discard/redo the move.
