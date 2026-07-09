# Architecture Trace Graph OKRA

Status: pilot objective achieved after independent challenge remediation; external acountee gate added
Owner: human owns the frame; agents may draft pilots and checks inside the frame
Started: 2026-07-09

## Objective

Prove whether **Architecture Trace Graph** is a useful C4-ish clue layer between C3 facts and code
detail across both C3 work phases:

1. brainstorming / impact analysis
2. code inspection / issue fixing
3. query reasoning / RAG

Metric: `trace_graph_decision_readiness`

Target: `trace_graph_decision_readiness >= 0.8`

Decision readiness is read from six representative C3 questions. A question scores 1 when the
pilot trace gives both:

- the right impact-analysis path through C3 facts/refs/rules
- the right first code/fix inspection path through stable anchors
- for query/RAG questions, the right retrieval reasoning path before context is packed into an answer

Formula:

```text
trace_graph_decision_readiness = useful_first_signal_count / representative_question_count
```

Current read:

| Metric | Target | Observed | Status |
| --- | ---: | ---: | --- |
| `representative_question_count` | `6` | `6` | met |
| `pilot_trace_count` | `>= 3` | `6` | met |
| `useful_first_signal_rate` | `>= 5/6` | `6/6` | met |
| `impact_analysis_signal_rate` | `>= 5/6` | `6/6` | met |
| `code_fix_signal_rate` | `>= 5/6` | `6/6` | met |
| `query_reasoning_signal_rate` | `>= 1/1` | `1/1` | met |
| `trace_graph_decision_readiness` | `>= 0.8` | `1.0` | met |
| `false_authority_claim_count` | `0` | `0` | met |
| `node_hash_noise_failures` | `0` | `0` | met |

Evidence:

- Spec: `docs/specs/2026-07-09-architecture-trace-graph.md`
- Replay command: `python3 scripts/check_trace_graph_okra.py`
- External target replay: `python3 scripts/check_trace_graph_acountee_eval.py`
- C3 check: `C3X_MODE=agent bash skills/c3/bin/c3x.sh check` returned `ok: true`
- C3 facts read for pilots: `c3-101`, `c3-102`, `c3-103`, `c3-106`, `c3-108`, `c3-109`,
  `c3-110`, `c3-203`, `c3-301`, `c3-401`, `c3-402`, `ref-eval-determinism`,
  `rule-output-via-helpers`
- Final challenger verdict: `intact` for the remediated path-anchor, hyphenated hash-noise, and
  apply-gate false-pass gaps.

## Anti-Goals

| Anti-goal | Metric | Type | Current read |
| --- | ---: | --- | --- |
| Do not become a correction/proof gate | `conformance_claim_count == 0` and `apply_gate_added_count == 0` | tripwire | held |
| Do not become analysis-only | `code_fix_route_missing_count == 0` | tripwire | held |
| Do not become chunk-only RAG | `query_reasoning_route_missing_count == 0` | tripwire | held |
| Do not require repairing external targets | `target_repair_or_mutation_count == 0` | tripwire | held |
| Do not become file-drift noise | `full_file_hash_basis_count == 0` and `line_number_hash_basis_count == 0` | tripwire | held |
| Do not add permanent C3 surface too early | `permanent_cli_or_canvas_change_before_pilot_count == 0` | tripwire | held |
| Keep OKRA integrity | `anti_goal_bypass_or_dishonesty_count == 0` | tripwire | held |

Anti-goal evidence:

- No CLI files changed.
- No `.c3/canvases/*` files changed.
- No frozen C3 fact files edited.
- Trace hash basis in the spec excludes full file content and line numbers.
- The spec says Trace Graph is a signal layer, not a proof, correction, eval, or apply-gate layer.
- Each pilot below names both an impact-analysis path and a code/fix inspection path.
- The query/RAG pilot names retrieval, trace expansion, context packing, and ranking-measurement
  anchors instead of relying on free-floating chunks.
- The acountee external replay surfaces target canonical drift, does not run repair, and verifies
  the real checkout git status is unchanged.

## DKR Checkpoints

| DKR | Decision target | Checkpoint | Confidence |
| --- | --- | --- | ---: |
| DKR-1 Naming and boundary | Pick the concept name and non-goals | Use **Architecture Trace Graph**, short **Trace Graph**. Reject **Code Spine** for now because it sounds canonical. | 0.85 |
| DKR-2 Node model | Decide the smallest useful node shape | Use graph id, owner, purpose, source facts, nodes, edges, anchors, and explicit hash basis. | 0.8 |
| DKR-3 Hash signal | Decide what the node hash should include | Hash role/meaning/owner/edge/symbol signals only. Exclude full files, line numbers, formatting, and whole bodies. | 0.8 |
| DKR-4 Pilot traces | Test representative questions | Six pilot traces below all provide a first inspection path. | 0.8 |
| DKR-5 Query/RAG route | Decide whether Trace Graph helps retrieval reasoning | Add one RAG-specific pilot that routes from search candidates through trace expansion into cited context and code/fix anchors. | 0.8 |
| DKR-6 External drift target | Decide whether Trace Graph still helps when target `c3 check` is red | Use acountee as a temporary `/tmp` eval target; treat red check as drift context, not a blocker. | 0.8 |

Candidate CKRs and PKRs are not promoted until the supporting DKR checkpoint is accepted by the
orchestrator.

## CKRs

| CKR | Metric | Target | Current |
| --- | --- | ---: | ---: |
| CKR-1 Signal utility | `useful_first_signal_rate` | `>= 5/6` | `6/6` |
| CKR-1b Two-phase coverage | `impact_analysis_signal_rate` and `code_fix_signal_rate` | `>= 5/6` each | `6/6` each |
| CKR-1c Query/RAG coverage | `query_reasoning_signal_rate` | `>= 1/1` | `1/1` |
| CKR-2 Low drift | `node_hash_noise_failures` | `0` | `0` |
| CKR-3 Honest semantics | `false_authority_claim_count` | `0` | `0` |
| CKR-4 External target resilience | `trace_graph_acountee_decision_readiness` | `>= 0.8` | `1.0` |

## PKRs

| PKR | Linked CKR | Output | Done read |
| --- | --- | --- | --- |
| PKR-1 Draft one-page spec | CKR-2, CKR-3 | `docs/specs/2026-07-09-architecture-trace-graph.md` | done |
| PKR-2 Write pilot traces | CKR-1 | Six pilots in this file | done |
| PKR-2b Add query/RAG pilot | CKR-1c | Query Reasoning / RAG pilot in this file | done |
| PKR-3 Run scorecard | CKR-1, CKR-2, CKR-3 | Current metric read table plus replay checker | done |
| PKR-4 Run acountee external gate | CKR-4 | `scripts/check_trace_graph_acountee_eval.py` | done |

## External Target Eval: Acountee

Purpose:

```text
Prove Trace Graph can build a useful first context pack on a target repo that currently has C3
canonical drift, without repairing or mutating that target.
```

Action envelope:

```text
allowed: copy target .c3 into /tmp, symlink source paths into /tmp, run local c3-design wrapper,
read/search/graph/lookup, inspect source anchors
forbidden: run c3 repair on acountee, edit /home/lagz0ne/dev/acountee, hide target drift
human boundary: only the human decides whether acountee should be repaired later
```

Pilot question:

> How does PR approval/approval-chain work, and where would a fix start?

Observed replay:

| Metric | Target | Observed | Status |
| --- | ---: | ---: | --- |
| `external_target_count` | `1` | `1` | met |
| `target_c3_check_red_visible_count` | `1` | `1` | met |
| `target_repair_or_mutation_count` | `0` | `0` | met |
| `target_visible_status_unchanged_count` | `1` | `1` | met |
| `target_ignored_status_unchanged_count` | `1` | `1` | met |
| `query_reasoning_context_pack_count` | `1` | `1` | met |
| `search_candidate_signal_count` | `>= 5` | `5` | met |
| `c3_fact_signal_count` | `>= 5` | `5` | met |
| `graph_neighbor_signal_count` | `>= 4` | `4` | met |
| `lookup_anchor_signal_count` | `>= 8` | `8` | met |
| `source_anchor_signal_count` | `>= 12` | `15` | met |
| `trace_context_pack_node_count` | `>= 6` | `6` | met |
| `trace_context_pack_edge_count` | `>= 5` | `5` | met |
| `trace_graph_acountee_decision_readiness` | `>= 0.8` | `1.0` | met |

Target drift surfaced:

```text
ok: false
only_in_tree: c3-4-nats-server-external/component-nats-core-broker-.md
missing_from_tree: c3-4-nats-server-external/component-nats-core-broker-component-nats-core-broker-.md
error: sync check failed: canonical markdown drift detected
hint: run c3x repair, then rerun c3x check
```

Trace context pack:

```text
search candidates:
  c3-105 PaymentRequestsScreen
  c3-205 PR Flows
  ref-approval-chain Approval Chain Pattern
  ui-flow-pr-approval-flow
  ui-state-machine-payment-request-lifecycle

graph neighbors:
  ref-approval-chain cited by c3-105, c3-107, c3-205, c3-210

lookup anchors:
  apps/start/src/screens/PaymentRequestsScreen.tsx -> c3-105, ref-approval-chain
  apps/start/src/server/flows/pr.ts -> c3-205, ref-approval-chain, ref-pumped-fn
  apps/start/src/server/dbs/queries/approval.ts -> ref-approval-chain, ref-query-services
  packages/shared/src/approval.ts -> ref-approval-chain

source anchors:
  approvalFlowSchema, ApprovalFlow, createPr, requestApprovals, approvePr, approveAll,
  approvalQueries, updateApprovalCurrentStep, requestForApprovals, notifyNextApprovers,
  canApprove, selectedApprovalFlow, stepsToApprovalFlow

trace pack:
  nodes: 6
  edges: 5
  hash_basis: graph id/version, node id/kind/clue/owners, C3 ids, paths, symbols, edges
```

Decision:

```text
The Trace Graph idea survives the harder acountee target if the product contract explicitly treats
red target checks as drift context instead of a prerequisite for query reasoning.
```

## Pilot Trace 1: Lookup Ownership

Question:

> Which path handles mapping a file or glob to the owning C3 facts and governing refs/rules?

First inspection path:

```text
c3-110 read-cmds
  -> c3-106 codemap-lib
  -> c3-102 store
  -> c3-109 cmd-support / rule-output-via-helpers
```

Impact-analysis signal:

```text
If a lookup answer is wrong or incomplete, inspect ownership boundaries first:
c3-110 command responsibility, c3-106 binding semantics, c3-102 stored relationship data.
```

Code/fix signal:

```text
Start at RunLookup, then GlobFiles / IsGlobPattern, then store lookup APIs, then shared output
helpers if the bug is serialization-shaped.
```

C3 source facts:

| Fact | Signal used |
| --- | --- |
| `c3-110` | Owns read-only `read`, `lookup`, `graph`, `search`, `check` command surface. |
| `c3-106` | Owns glob matching and file-to-fact binding resolution. |
| `c3-102` | Owns typed persistence for entities, relationships, nodes, code-map, and semantic vectors. |
| `c3-109` | Owns shared command registry, option parsing, and output helper. |
| `rule-output-via-helpers` | Governs structured command output through shared helpers. |

Trace nodes:

| Node | Kind | Clue | Owners | Stable anchors | Hash basis |
| --- | --- | --- | --- | --- | --- |
| `lookup-command` | entrypoint | User-supplied path or glob enters the read command surface. | `c3-110` | `RunLookup`, `cli/cmd/lookup.go` | kind, clue, owners, symbols, edges |
| `binding-resolution` | binding | The path/glob resolves against fact code bindings. | `c3-106` | `GlobFiles`, `IsGlobPattern` | kind, clue, owners, symbols, edges |
| `fact-read` | state | Matched owners and relationships are read as typed store results. | `c3-102` | `cli/internal/store/**` | kind, clue, owners, symbols, edges |
| `agent-output` | output | Result leaves through shared TOON/JSON helpers. | `c3-109`, `rule-output-via-helpers` | `WriteTableOutput`, `WriteObjectOutput`, `writeJSON` | kind, clue, owners, symbols, edges |

Edges:

```text
lookup-command -> binding-resolution -> fact-read -> agent-output
agent-output -> rule-output-via-helpers:governed-by
```

Scorecard: useful first signal. It points to the read command, binding resolver, store read, and
output contract before opening implementation files.

## Pilot Trace 2: Eval Conformance

Question:

> Before changing eval gather semantics, where should I inspect first?

First inspection path:

```text
c3-108 eval-engine
  -> ref-eval-determinism
  -> c3-106 codemap-lib
  -> cli/internal/eval/**
```

Impact-analysis signal:

```text
If an eval result moves unexpectedly, inspect deterministic gather policy before changing code.
The blast radius is c3-108 plus ref-eval-determinism, then code binding through c3-106.
```

Code/fix signal:

```text
Start at eval spec loading and gather/assert paths in cli/internal/eval, then binding resolution in
codemap-lib if the issue is unresolved code selectors.
```

C3 source facts:

| Fact | Signal used |
| --- | --- |
| `c3-108` | Owns gather, filter, transform, eval, loop, and verdict stamping. |
| `ref-eval-determinism` | Governs deterministic gathered frames and ExternalState hashing. |
| `c3-106` | Resolves code bindings and globs used by eval. |

Trace nodes:

| Node | Kind | Clue | Owners | Stable anchors | Hash basis |
| --- | --- | --- | --- | --- | --- |
| `eval-spec` | entrypoint | A fact's eval spec selects external state to measure. | `c3-108` | `.c3/eval/<fact>.yaml`, `EvalBindings` | kind, clue, owners, symbols, edges |
| `deterministic-gather` | policy | Gather must produce byte-stable frames for unchanged subjects. | `ref-eval-determinism`, `c3-108` | `gather`, `ExternalState` | kind, clue, owners, symbols, edges |
| `code-binding` | binding | `code:` selectors resolve through the code binding resolver. | `c3-106`, `c3-108` | `GlobFiles`, `gather code` | kind, clue, owners, symbols, edges |
| `verdict` | output | Holds/drift/needs-judgement is a one-off signal, never an apply gate. | `c3-108` | `assert`, `Verdict` | kind, clue, owners, symbols, edges |

Edges:

```text
eval-spec -> deterministic-gather -> code-binding -> verdict
deterministic-gather -> ref-eval-determinism:governed-by
```

Scorecard: useful first signal. It keeps the first inspection at eval semantics and determinism
instead of starting with random call sites.

## Pilot Trace 3: Agent Output Format

Question:

> What should I inspect before changing structured agent output?

First inspection path:

```text
rule-output-via-helpers
  -> c3-109 cmd-support
  -> c3-110 read-cmds
  -> concrete command tests
```

Impact-analysis signal:

```text
If agent output drifts, inspect rule-output-via-helpers first, then c3-109 shared command-output
ownership, then c3-110 commands that consume it.
```

Code/fix signal:

```text
Start at WriteTableOutput / WriteObjectOutput / writeJSON / writeHints. Only inspect individual
commands after the shared helper path is understood.
```

C3 source facts:

| Fact | Signal used |
| --- | --- |
| `rule-output-via-helpers` | Commands serialize structured results through shared helpers. |
| `c3-109` | Owns the shared output helper and option parsing. |
| `c3-110` | Read commands are high-volume consumers of structured output. |

Trace nodes:

| Node | Kind | Clue | Owners | Stable anchors | Hash basis |
| --- | --- | --- | --- | --- | --- |
| `output-rule` | policy | Structured result format is centralized. | `rule-output-via-helpers` | `WriteTableOutput`, `WriteObjectOutput` | kind, clue, owners, symbols, edges |
| `format-switch` | output | Agent mode chooses compact TOON unless explicit JSON path applies. | `c3-109` | `writeJSON`, `writeHints` | kind, clue, owners, symbols, edges |
| `read-command-consumers` | output | Read commands must reuse helpers instead of hand-rolling output. | `c3-110` | `RunLookup`, `RunSearch`, `RunCheckV2` | kind, clue, owners, symbols, edges |

Edges:

```text
output-rule -> format-switch:governed-by
format-switch -> read-command-consumers:serializes-as
```

Scorecard: useful first signal. It tells a maintainer to inspect the rule and helper path before
editing individual command formatting.

## Pilot Trace 4: Runtime Resolution

Question:

> Where does version/platform runtime selection live for skill and npm entrypoints?

First inspection path:

```text
c3-203 cli-wrapper
  -> c3-301 binary-downloader
  -> ref-cross-compiled-binary
  -> ref-fat-thin-distribution
```

Impact-analysis signal:

```text
If runtime startup or version selection breaks, separate c3-203 skill-wrapper behavior from c3-301
npm thin-client behavior before reading Go command code.
```

Code/fix signal:

```text
Start at skills/c3/bin/c3x.sh for skill invocation bugs. Start at packages/cli/src/manager.ts for
npm install/cache/checksum/runtime-selection bugs.
```

C3 source facts:

| Fact | Signal used |
| --- | --- |
| `c3-203` | Owns skill wrapper platform detection, bundled binary selection, source build fallback, and npm delegation. |
| `c3-301` | Owns npm thin-client runtime selection, cache, checksum verification, and exec. |
| `ref-cross-compiled-binary` | Governs supported OS/arch and release asset names. |
| `ref-fat-thin-distribution` | Governs full-fat skill, portable skill, and thin npm split. |

Trace nodes:

| Node | Kind | Clue | Owners | Stable anchors | Hash basis |
| --- | --- | --- | --- | --- | --- |
| `skill-wrapper-entry` | runtime | Skill commands enter through `skills/c3/bin/c3x.sh`. | `c3-203` | `skills/c3/bin/c3x.sh`, `skills/c3/bin/VERSION` | kind, clue, owners, symbols, edges |
| `asset-selection` | runtime | Wrapper and npm client select release assets from platform/version. | `c3-203`, `c3-301`, `ref-cross-compiled-binary` | `asset_name`, `assetNames`, `resolvePlatform` | kind, clue, owners, symbols, edges |
| `cache-and-verify` | runtime | npm path downloads, verifies sha256, and atomically activates the runtime. | `c3-301`, `ref-fat-thin-distribution` | `ensureCachedAsset`, `runCli` | kind, clue, owners, symbols, edges |

Edges:

```text
skill-wrapper-entry -> asset-selection -> cache-and-verify
asset-selection -> ref-cross-compiled-binary:governed-by
cache-and-verify -> ref-fat-thin-distribution:governed-by
```

Scorecard: useful first signal. It prevents runtime questions from starting in the Go command layer,
which is explicitly a non-goal of the wrapper/downloader components.

## Pilot Trace 5: Canvas Validation

Question:

> Before changing fact/canvas validation, what shape path should I inspect?

First inspection path:

```text
c3-103 schema
  -> c3-101 doc-model
  -> c3-102 store
  -> c3-110 check/read commands
```

Impact-analysis signal:

```text
If a validation issue appears, decide first whether it is canvas shape, markdown parsing, stored
node state, or command reporting.
```

Code/fix signal:

```text
Start at schema resolution for shape bugs, doc-model parsing for markdown/table bugs, store
round-trip for seal/node bugs, and RunCheckV2 for reporting bugs.
```

C3 source facts:

| Fact | Signal used |
| --- | --- |
| `c3-103` | Owns canvas definitions, section/column shape, edge-writer columns, and validation. |
| `c3-101` | Parses frontmatter, sections, tables, node trees, and document rendering. |
| `c3-102` | Persists typed entities, relationships, nodes, versions, and seals. |
| `c3-110` | Exposes check/read command surfaces that report validation issues. |

Trace nodes:

| Node | Kind | Clue | Owners | Stable anchors | Hash basis |
| --- | --- | --- | --- | --- | --- |
| `canvas-shape` | validation | Entity type shape resolves from built-in and project canvases. | `c3-103` | `ParseCanvasDocument`, `ResolveCanvas`, `DefinitionForDir` | kind, clue, owners, symbols, edges |
| `document-parse` | binding | Raw markdown becomes frontmatter, sections, tables, and node tree. | `c3-101` | `ParseFrontmatter`, `ParseSections`, `ParseTable` | kind, clue, owners, symbols, edges |
| `store-roundtrip` | state | Parsed nodes and versions are written/read through typed store APIs. | `c3-102` | `WriteEntity`, `ReadEntity`, `RenderMarkdown` | kind, clue, owners, symbols, edges |
| `validation-output` | output | Check reports shape/seal/layer/citation issues through read-command output. | `c3-110` | `RunCheckV2`, `RunRead` | kind, clue, owners, symbols, edges |

Edges:

```text
canvas-shape -> document-parse -> store-roundtrip -> validation-output
validation-output -> rule-output-via-helpers:governed-by
```

Scorecard: useful first signal. It points to schema before markdown parsing and storage, which keeps
shape changes separate from document and persistence internals.

## Pilot Trace 6: Query Reasoning / RAG

Question:

> How should C3 retrieval guide an answer before the agent opens code or builds a context pack?

First inspection path:

```text
c3-110 read-cmds search
  -> c3-102 store search / semantic vectors
  -> c3-401 search-eval ranking metrics
  -> c3-402 semantic-assets distribution path
```

Impact-analysis signal:

```text
If a query answer is weak, inspect c3-110 search responsibility, c3-102 retrieval/storage signals,
c3-401 ranking measurement, and c3-402 semantic asset readiness before changing prompts or code.
```

Code/fix signal:

```text
Start at RunSearch, collectSearchRows, expandHybridRows, fuseSemanticRows, enrichSearchRow,
SearchContent, SearchWithLimit, and SearchSemanticWithOptions before changing broader RAG behavior.
```

Query/RAG signal:

```text
Use Trace Graph as the bridge from ranked search hits to a cited context pack: candidate ids,
match_sources, graph context, owners, refs/rules, stable anchors, and stale/missing-anchor signals.
```

C3 source facts:

| Fact | Signal used |
| --- | --- |
| `c3-110` | Owns `search` as the read command that fuses semantic, keyword, and graph context. |
| `c3-102` | Owns the store surfaces behind FTS, semantic vectors, entities, relationships, and nodes. |
| `c3-401` | Owns the labelled query corpus and Hit@1/3/5/MRR search-quality report. |
| `c3-402` | Owns semantic model/runtime asset preparation for embedded and release modes. |

Trace nodes:

| Node | Kind | Clue | Owners | Stable anchors | Hash basis |
| --- | --- | --- | --- | --- | --- |
| `query-entry` | entrypoint | Natural-language question enters the read command surface. | `c3-110` | `RunSearch`, `cli/cmd/search.go` | kind, clue, owners, symbols, edges |
| `candidate-retrieval` | binding | Keyword/content/entity hits and graph expansion produce candidate C3 ids. | `c3-110`, `c3-102` | `collectSearchRows`, `expandHybridRows`, `SearchContent`, `SearchWithLimit` | kind, clue, owners, symbols, edges |
| `semantic-fusion` | binding | Optional semantic hits are fused with keyword/graph candidates. | `c3-110`, `c3-102`, `c3-402` | `EnsureSemanticIndexWithOptions`, `SearchSemanticWithOptions`, `fuseSemanticRows`, `cli/internal/store/semantic.go` | kind, clue, owners, symbols, edges |
| `context-pack` | output | Search results carry match sources and graph/code context for the agent to read next. | `c3-110` | `SearchResultRow`, `SearchContext`, `enrichSearchRow` | kind, clue, owners, symbols, edges |
| `ranking-measure` | validation | Search-quality changes are measured against labelled cases. | `c3-401` | `cli/tools/search-eval/**`, `MRR` | kind, clue, owners, symbols, edges |
| `semantic-asset-readiness` | runtime | Semantic retrieval depends on pinned model/runtime assets being prepared or downloadable. | `c3-402` | `cli/tools/semantic-assets/**`, `docs/specs/2026-06-08-local-onnx-semantic-search.md` | kind, clue, owners, symbols, edges |

Edges:

```text
query-entry -> candidate-retrieval -> semantic-fusion -> context-pack
context-pack -> ranking-measure:verifies-with
semantic-fusion -> c3-402:governed-by
```

Scorecard: useful query/RAG signal. It keeps retrieval grounded in C3 ids, match provenance,
graph context, and ranking metrics instead of treating embedding snippets as the answer.

## Six-Question Scorecard

| Question | Trace | Impact signal | Code/fix signal | Stable anchors | Anti-goals held | Score |
| --- | --- | --- | --- | --- | --- | ---: |
| Which path handles mapping a file/glob to owning facts? | Lookup Ownership | yes | yes | 7 | yes | 1 |
| Before changing eval gather semantics, where should I inspect? | Eval Conformance | yes | yes | 6 | yes | 1 |
| What should I inspect before changing structured agent output? | Agent Output Format | yes | yes | 7 | yes | 1 |
| Where does version/platform runtime selection live? | Runtime Resolution | yes | yes | 6 | yes | 1 |
| Before changing fact/canvas validation, what shape path should I inspect? | Canvas Validation | yes | yes | 9 | yes | 1 |
| How should C3 retrieval guide an answer before code/context packing? | Query Reasoning / RAG | yes | yes | 12 | yes | 1 |

Read:

```text
useful_first_signal_count = 6
impact_analysis_signal_count = 6
code_fix_signal_count = 6
query_reasoning_signal_count = 1
representative_question_count = 6
trace_graph_decision_readiness = 1.0
```

## Three Anti-Goal Eval Points

| Point | Trace |
| --- | --- |
| Admissibility before acting | Initial pilot work only added docs. The follow-on route-enrichment OKRA permits CLI output enrichment while still forbidding new primitives, `.c3/canvases` changes, and frozen-fact changes. |
| Direct read after acting | This pilot produced docs/task artifacts; implementation evidence now lives in `2026-07-09-route-enrichment-existing-surfaces-okra.md`. |
| Paired with objective | Objective metrics read green while anti-goals still read held. |

## Flags

| Flag | Current state | Evidence |
| --- | --- | --- |
| Cannot | closed | Six pilot traces found stable clues. |
| Breaking | closed | No proof/correction/apply-gate wording accepted; no new primitive/canvas/frozen-fact changes. Existing CLI surfaces may carry route enrichment. |
| Pointless | closed | Each representative question scored a useful first inspection path. |
| Stalled | closed | DKR outputs became a spec plus pilots in this round. |
| Authority drift | closed | No frame, threshold, or action-envelope change made without human direction. |

## Independent Challenge Trail

| Pass | Verdict | Finding | Resolution |
| --- | --- | --- | --- |
| Challenger 1 | drifted | Checker counted labels and table rows more than substance; proof/apply-gate and hash-noise false passes existed. | Replaced checker with expected pilots, C3 fact reads, documented/source anchor checks, authority scans, and negative controls. |
| Challenger 2 | drifted | Phase blocks could be useless long text; pilot hash-basis rows and apply-gate variants could still false-pass. | Added phase-specific required terms, OKRA table hash-noise scans, broader apply-gate patterns, and targeted negative controls. |
| Challenger 3 | drifted | Path-only anchors were not checked; hyphenated hash-noise and another apply-gate wording could still pass. | Added `path_anchors`, documented/source path resolution, hyphen normalization, broader apply-gate wording, and targeted negative controls. |
| Challenger 4 | intact | The three challenged gaps were closed for the pilot stage. | No further remediation required for this OKRA objective. |
| Query/RAG extension | checker-held | User noticed query reasoning/RAG was missing from the dogfood content. | Added a sixth pilot and replay checks for retrieval reasoning, context-pack routing, search anchors, semantic assets, and ranking measurement. |

Current negative controls:

```text
apply-gate-positive
change-apply-required
change-apply-completed-before
certifies-fix-positive
analysis-only
useless-impact-signal
useless-code-fix-signal
useless-query-rag-signal
noisy-hash-basis
pilot-row-noisy-hash-basis
pilot-row-hyphenated-noisy-hash-basis
missing-path-anchor
missing-anchor
```

## Human-Only Frame

The human owns these calls:

- Accept or rename **Architecture Trace Graph**.
- Decide whether this pilot can become a C3 canvas, eval convention, or CLI surface.
- Change the thresholds above.
- Retire the anti-goals.

Recommended next action:

Run an independent challenge on the five pilots. If the challenge holds, promote the concept to a
design proposal. If it finds dead or noisy hypotheses, retire those pilot nodes instead of carrying
them as architecture noise.
