# Architecture Route Enrichment

Status: implementation direction
Date: 2026-07-09

## Goal

Enrich C3's existing query, impact-analysis, frozen-fact, drift-prevention, and graph surfaces with
route clues between architecture facts and code detail.

The route answers two questions:

> During brainstorming or impact analysis, what architecture path explains the change?

> During code work or issue fixing, what stable implementation shape should I inspect first?

It is not a new primitive. It is not a correction layer. It is not a proof layer. It is a signal
layer inside current C3 commands.

```text
current C3 surfaces
  search / query
  graph
  lookup
  eval code bindings
  check drift warnings
        |
        v
route enrichment
  facts + graph neighbors + eval anchors + lanes + drift labels + route hash
```

## Name

Product name: **Route Enrichment**

Pilot/eval name: **Trace Graph**

Why this name:

| Candidate | Decision | Reason |
| --- | --- | --- |
| Code Spine | reject for now | Sounds too canonical and structural. |
| Trace Graph | keep as pilot/eval term | Useful for the replay fixture, but too easy to mistake for a new graph primitive. |
| Code Signal Graph | keep as fallback | Accurate, but clunky. |
| Implementation Trace | keep as fallback | Clear, but less graph-shaped. |
| Route Enrichment | use | Says it enriches existing query/graph/check routes instead of adding a new surface. |

## Boundary

Route enrichment sits inside existing C3 surfaces between frozen facts and mutable source files.

| Layer | Owns | Route-enrichment relation |
| --- | --- | --- |
| C3 fact | Stable architecture claim, ownership, refs, rules | Source of meaning and ownership. |
| Search/query | Concept to candidate facts/docs | Carries facts, graph neighbors, lanes, anchors, and drift labels in the existing result. |
| Graph | Current relationship graph | Carries route facets on current graph nodes, not a second graph. |
| Eval | One-off conformance verdict against external state | Supplies the fact-to-code anchor bindings and verifies selected anchors. |
| Check | Structural/canonical drift prevention | Warns when eval code anchors stop matching files. |
| Code map / code binding | Fact to file/glob binding | Supplies lower-level anchor resolution. |
| Code | Actual implementation | Opened after the route says where to inspect. |

## Current Surface Integration

| Surface | Enrichment | Anti-goal |
| --- | --- | --- |
| `c3 search` | Add route facts, graph neighbors, eval anchors, lanes, drift labels, and a stable route hash to existing result rows. | Do not claim search ranking improved. |
| `c3 graph` | Add the same route facets to current graph nodes. | Do not create a second graph command or model. |
| `c3 lookup` | Continue to resolve files through eval code bindings. | Do not replace lookup with a new code map. |
| `c3 eval` | Continue as conformance verdict. | Do not treat route enrichment as proof. |
| `c3 check` | Warn when an eval `code:` anchor matches no files. | Do not mutate or repair code/facts from the warning. |

## Two-Phase Contract

C3 has two work phases. Route enrichment must serve both.

| Phase | C3 job | Route-enrichment job | Done signal |
| --- | --- | --- | --- |
| Brainstorming / impact analysis | Explain ownership, refs, blast radius, and likely affected paths. | Show the clue graph from business/ref/component meaning to neighboring responsibilities. | A human or agent can name the right affected facts before opening code. |
| Code / fix execution | Guide concrete inspection and repair work. | Show the first stable implementation anchors, then hand off to code, tests, and eval for proof. | A human or agent can name the first files/symbols/tests to inspect without broad grep-first wandering. |

Route enrichment is allowed to guide a code fix. It is not allowed to certify that the fix is correct.
The proof still comes from tests, eval, review, or direct runtime evidence.

## Query Reasoning / RAG Contract

Route enrichment also helps C3 answer questions. It turns retrieval into a smaller reasoning path:

```text
user question
  -> search candidates
  -> current graph and eval anchors around the best candidates
  -> route-enriched context
  -> answer or first code/fix route
```

Without this layer, RAG returns chunks. The agent must infer the architecture path from snippets.
With this layer, RAG returns candidates plus a route: owner, governing refs/rules, neighboring
responsibilities, stable anchors, and stale/missing-anchor signals.

| Query need | Route-enrichment help | Drift benefit |
| --- | --- | --- |
| Retrieve the right fact | Adds graph/role context around search hits. | A renamed file does not erase the architecture clue. |
| Reason across facts | Gives explicit edges for ownership, policy, state, output, and runtime. | The agent follows declared hops instead of inventing hidden links from snippets. |
| Build a context pack | Selects facts, refs/rules, and anchors as separate evidence units. | Stale anchor/hash signals can down-rank or flag a chunk before it enters the answer. |
| Route to code | Names first files/symbols/tests after the architecture path is known. | The answer does not jump from a semantic hit directly to broad grep. |

RAG must treat route enrichment as retrieval guidance, not answer proof. The answer still cites the C3
facts it read and the code/tests/eval it used.

## External Target Evaluation

Route enrichment must survive repos that are less clean than c3-design.

The external-target rule:

```text
target repo has C3 drift
  -> surface drift as a route warning
  -> do not repair or mutate the target
  -> still route query -> C3 facts -> graph neighbors -> lookup anchors -> source anchors
```

This matters because real use will often begin while the target C3 state is imperfect. A red
`c3 check` should not erase useful query reasoning. It should become a drift signal attached to the
context pack.

Temporary eval target: `/home/lagz0ne/dev/acountee`

Replay command:

```bash
python3 scripts/check_trace_graph_acountee_eval.py
```

The checker builds a disposable fixture under `/tmp`, copies only the target `.c3`, symlinks source
directories, and uses the local c3-design wrapper for all C3 commands. It fails if the real acountee
checkout changes, including ignored-file status.

| External signal | Required behavior |
| --- | --- |
| `c3 check` is red | Preserve and report the drift. Do not run repair. |
| C3 `search` works | Use hits as candidates, not proof. |
| C3 `graph` works | Add neighboring facts/refs/custom docs around the candidate. |
| C3 `lookup` works | Attach first file/fact anchors for code inspection. |
| Source anchors resolve | Name stable symbols/tests/files to open first. |
| Source anchors miss | Mark `stale/missing-anchor`; do not hide it. |
| Route context builds | Emit facts, graph neighbors, anchors, lanes, drift labels, and a stable route hash. |

## RAG Route Quality Evaluation

Route enrichment should improve the route after search, not claim the search ranker improved.

```text
baseline:
  c3 search only

candidate:
  c3 search
    -> route-enriched context
    -> facts/custom docs
    -> graph neighbors
    -> files
    -> symbols/tests
    -> drift signal when present
```

Replay command:

```bash
python3 scripts/check_trace_graph_rag_route_quality.py
```

The route-quality checker uses c3-design plus acountee. The acountee cases are deliberately
cross-cutting: frontend/backend approval, UI behavior collection, auth lifecycle, invoice lifecycle,
theming/component papercuts, and NATS sync cycle.

This is a source-anchored replay fixture. It uses expected file and symbol anchors as gold labels to
test whether route-enriched context gives better RAG routing clues than raw search alone. It does not
prove automatic source-anchor discovery.

| Metric | Meaning |
| --- | --- |
| `baseline_route_quality` | What raw `c3 search` can establish by itself. |
| `trace_route_quality` | What the route-enriched context pack establishes after search. |
| `rag_route_quality_delta` | Route lift from enrichment after search. |
| `search_ranking_delta` | Must stay `0.0` unless Hit@k/MRR ranking eval proves otherwise. |
| `auto_discovery_claim_count` | Must stay `0` until a product route discovers anchors without gold labels. |

## Cross-Cutting Lanes

Route enrichment can carry more than one lane through the same codebase. A lane is a query-shaped spine
through time, ownership, and lifecycle state.

| Lane | Route starts with | Must cross |
| --- | --- | --- |
| Frontend/backend ownership | UI screen or server flow | Shared schema, DB query, owning C3 facts |
| Auth lifecycle | Login/guard route | Server auth resource, e2e auth stub, state machine |
| Invoice lifecycle | Invoice UI action | Backend flow/service, import/obsolete/restore states, e2e |
| E2E validation | Test script | Source flow, fixture/state setup, observed UI or DB outcome |
| Components/theming papercuts | UI primitive or variant | Theme bootstrap, screen usage, repeated component pressure |
| Real-time cycle | Frontend subscription or atom | Backend publisher, broker component, notification lifecycle |

## Non-Goals

| Non-goal | Rule |
| --- | --- |
| Correcting code | Route enrichment never rewrites or judges code correctness. |
| Acting as an apply gate | A changed route hash may warn, but must not block `change apply`. |
| Replacing eval | Eval remains the conformance verdict surface. |
| Replacing code bindings | `.c3/eval/<fact>.yaml` remains the file/glob binding surface. |
| Modeling imports | Do not build a language import graph here. |
| Hashing files | Do not hash full file content or line numbers. |
| Claiming completeness | A trace is a useful path, not an exhaustive architecture model. |
| Claiming automatic discovery | The route-quality replay is gold-labeled until a product route proves anchor discovery. |
| Adding primitives | Do not add `trace`, `route`, `spine`, or `impact` commands for this capability. |

## Pilot Graph Model

```yaml
id: trace-lookup-ownership
owner: c3-110
purpose: "Find the read path that maps a file or glob to owning C3 facts."
status: pilot
source_facts:
  - c3-110
  - c3-106
  - c3-102
  - c3-109
nodes:
  - id: lookup-command
    kind: entrypoint
    clue: "User-supplied path or glob enters the read command surface."
    owners: [c3-110]
    anchors:
      paths: [cli/cmd/lookup.go]
      symbols: [RunLookup]
    hash_basis: [kind, clue, owners, anchors.symbols]
edges:
  - from: lookup-command
    to: binding-resolution
    kind: routes-to
```

## Node Fields

| Field | Required | Meaning |
| --- | --- | --- |
| `id` | yes | Stable trace-local id. |
| `kind` | yes | Role category, not code type. |
| `clue` | yes | Human-readable signal. One or two sentences. |
| `owners` | yes | C3 fact/ref/rule ids that supply authority for the clue. |
| `anchors.paths` | optional | File or glob hints. Evidence, not identity. |
| `anchors.symbols` | optional | Stable names to inspect first. |
| `hash_basis` | yes | Explicit list of fields hashed for change signal. |
| `status` | optional | `pilot`, `accepted`, `stale`, or `retired`. |

## Node Kinds

| Kind | Use when |
| --- | --- |
| `entrypoint` | User input, command invocation, event, or public call enters a flow. |
| `binding` | A fact/code, path/entity, or selector relationship is resolved. |
| `state` | Persistent architecture state is read or written. |
| `policy` | A ref or rule governs behavior. |
| `output` | A result leaves the system through a user/agent surface. |
| `runtime` | Runtime selection, install, cache, or exec is the concern. |
| `validation` | A shape or invariant is checked. |

## Edge Kinds

| Kind | Meaning |
| --- | --- |
| `routes-to` | Control or request moves to the next role. |
| `resolves` | A selector is resolved to facts, files, or nodes. |
| `reads` | A node reads state from another owner. |
| `writes` | A node writes state through another owner. |
| `governed-by` | A ref or rule constrains the node. |
| `serializes-as` | A node emits through an output contract. |
| `verifies-with` | A node can be checked by an eval/test command. |

## Route Hash Signal

Route hashes are change signals, not correctness verdicts.

Hash these:

| Input | Why |
| --- | --- |
| entity ids | Captures the owning C3 facts/refs/rules. |
| neighboring fact ids | Captures current graph route shape. |
| eval code globs | Captures first-inspection source anchors. |
| inferred lanes | Captures lifecycle/ownership/cycle signal. |
| drift labels | Captures missing/stale anchor signal. |

Do not hash these:

| Input | Why not |
| --- | --- |
| full file content | Too noisy. Makes the graph file-drift sensitive. |
| line numbers | Too fragile. |
| formatting | No meaning. |
| whole function bodies | Pulls the layer back to code-level correction. |
| ranked search output | Non-deterministic enough to create false movement. |

Canonical hash shape:

```json
{
  "hash_version": 1,
  "graph_id": "trace-lookup-ownership",
  "nodes": [
    {
      "id": "lookup-command",
      "kind": "entrypoint",
      "clue": "User-supplied path or glob enters the read command surface.",
      "owners": ["c3-110"],
      "anchor_symbols": ["RunLookup"],
      "out_edges": ["lookup-command->binding-resolution:routes-to"]
    }
  ]
}
```

Then compute `sha256(canonical_json)`.

## Status Signal

| Signal | Meaning | Next action |
| --- | --- | --- |
| `held` | Hash and anchors match the last accepted trace. | Use trace as first-inspection path. |
| `node_changed` | Node clue, owner, kind, or symbol anchor changed. | Inspect the changed node before opening broad code. |
| `edge_changed` | Flow shape changed. | Inspect producer/consumer boundary. |
| `anchor_missing` | Path or symbol hint no longer resolves. | Re-aim the anchor or retire the node. |
| `unchecked` | No deterministic check was run. | Treat as clue only. |

## Pilot Scorecard

A pilot trace is useful when it passes all rows:

| Check | Target |
| --- | --- |
| Names the user question it helps answer | yes |
| Names C3 source facts/refs/rules | at least 2 |
| Names a first inspection path | yes |
| Names stable anchors above line-level detail | at least 2 |
| Declares hash basis | yes |
| Avoids proof/correction wording | yes |
| Avoids full-file and line-number hash inputs | yes |

The first objective uses six representative questions. A pilot gets credit only when it gives the
right first inspection path before opening code.
