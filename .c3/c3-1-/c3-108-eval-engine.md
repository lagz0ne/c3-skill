---
id: c3-108
c3-seal: 7b7535547848c2179e4abf967e5a5a362ba88567f800ce34fe92daa995ff87ec
title: eval-engine
type: component
category: foundation
parent: c3-1
goal: Run a fact's conformance pipeline — check a frozen claim against the uncontrolled external it governs — and produce a one-off, stamped verdict (holds / drift / needs-judgement) that is never an apply gate.
uses:
    - ref-eval-determinism
    - rule-wrap-error-cause
---

# eval-engine

## Goal

Run a fact's conformance pipeline — check a frozen claim against the uncontrolled external it governs — and produce a one-off, stamped verdict (holds / drift / needs-judgement) that is never an apply gate.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | The conformance engine: it interprets each fact's eval pipeline, composes the five ops, and stamps the verdict the eval command and CI read. |
| Boundary | Owns the pipeline interpreter, the predicates, and the verdict stamp; it resolves globs through codemap-lib and reads specs from disk, but decides nothing about doc shape (schema) and persists nothing (store). |
| Collaboration | The eval command loads the specs and passes the engine the project dir, the code bindings, and the fact ids; the engine runs each spec to a verdict and hands the array back to serialize. |

## Purpose

Interpret an eval spec (.c3/eval/<fact>.yaml) as a composition of five ops — gather, filter, transform, eval, loop — over a frame of gathered material, and reduce it to a stamped verdict. Gather is the only boundary-crossing op; eval is the only terminal that asserts; loop fans a sub-pipeline over a collection and rolls the per-item verdicts into one. Non-goals: glob resolution (codemap-lib), spec discovery and authoring, doc-shape validation (schema), and persistence (store).

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-eval-determinism | ref | Which gather sources the engine may admit, and the meaning of the stamped verdict | Invariant binds the engine over any convenience | A verdict is solid only for the (claim, external-state) pair it stamped; only deterministic gather sources are admissible. |
| rule-wrap-error-cause | rule | Errors the engine returns when a gather fails wrap the underlying cause | Standard applies to all engine error paths | A gather that cannot start is a real error; a non-zero exit (no match) is a legitimately empty frame, not a failure. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| gather | IN | Acquire material into the frame from exactly one source — file, command, files (glob), facts (id-glob), code (a fact's declared globs), literal, or each — the only op that crosses to the external or fact bank | Acquires only; asserts nothing; an empty or non-zero-exit acquisition is an empty frame, not an error | cli/internal/eval/eval.go |
| filter / transform | IN/OUT | Reshape the frame already in hand — keep values matching a predicate, or trim / first / lines — never acquire new material | Operate only on the gathered frame | cli/internal/eval/eval.go |
| eval | OUT | Assert on the frame to produce the verdict — exists, equals, all_equal, contains_all, contains, count, or judgement, yielding holds / drift / needs-judgement | The only terminal op; the only op that yields a verdict | cli/internal/eval/eval.go |
| loop | IN/OUT | Fan a sub-pipeline over a gathered collection, binding $item, and roll the per-item verdicts into one — holds iff every item holds, and the evidence names which item drifted | Membership comes from a deterministic selector, never ranked retrieval | cli/internal/eval/eval.go |
| verdict | OUT | Stamp ExternalState as the hash of the gathered frame; the verdict is solid only for that (claim, external-state) pair, and the run always exits success | A one-off signal, never an apply gate | cli/internal/eval/eval.go |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/internal/eval/**.go | Contract | The interpreter internals — op dispatch, predicate implementations, the stamp function — may vary as long as the five-op contract and the stamped verdict hold | go test ./internal/eval/... |
