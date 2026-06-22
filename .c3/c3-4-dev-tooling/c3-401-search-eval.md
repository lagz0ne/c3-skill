---
id: c3-401
c3-seal: 3b9cbeac151faa9289ca2511ac6a1d7e6c0f17faae09c6d7da42339d23fd035b
title: search-eval
type: component
category: foundation
parent: c3-4
goal: Measure whether a change to c3x search ranking is a real, numeric win — run a fixed set of labelled queries through the live search path against the local store and report Hit@1/3/5 and MRR so retrieval changes are kept only on a measured improvement.
uses:
    - rule-wrap-error-cause
---

# search-eval

## Goal

Measure whether a change to c3x search ranking is a real, numeric win — run a fixed set of labelled queries through the live search path against the local store and report Hit@1/3/5 and MRR so retrieval changes are kept only on a measured improvement.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-4 |
| Role | The search-ranking quality harness: a standalone main that drives the CLI's search code over a curated case set and emits a metrics report, never compiled into the shipped binary. |
| Boundary | Owns the eval case corpus, the ranking metrics, and the report shape; it does not implement search itself and changes no ranking behavior. |
| Collaboration | It opens the project's .c3/c3.db through the store, invokes cmd.RunSearch per case, ranks the returned ids against each case's expected ids, and prints a JSON report a developer compares across runs. |

## Purpose

Owns the labelled query corpus (paraphrase and keyword cases mapping a natural-language query to the entity ids that should rank for it), the ranking-metric computation (Hit@1, Hit@3, Hit@5, MRR, overall and by case kind), and the JSON report shape, plus a `--db`/`-k`/`--semantic`/`--no-semantic` command surface and version-env seeding for semantic runs. Non-goals: implementing search or ranking (that is the CLI search command and the store), maintaining the semantic index, and gating anything — it is a measurement tool a human reads, never an apply check.

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| rule-wrap-error-cause | rule | Errors the harness surfaces from the store and search path are wrapped with %w (e.g. "%s: %w", "search: %w") so the failing case and cause stay reachable. | Standard applies to every error path in the harness | Bare leaf fail exits for flag/db errors are the constructed-leaf exception the rule allows. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| eval cases | IN | A fixed corpus of {id, kind, query, expected} cases the harness ranks search output against. | The labels are the fixture; the harness does not learn or mutate them. | cli/tools/search-eval/main.go |
| metrics report | OUT | A JSON report of Hit@1/3/5 and MRR, overall and by kind, plus per-case rank and hits, for cross-run comparison. | Reports numbers only; it draws no keep/revert conclusion itself. | cli/tools/search-eval/main.go |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/tools/search-eval/**.go | Contract | The case set and metric internals may evolve as long as the case-then-metrics-report contract holds | go test ./tools/... |
