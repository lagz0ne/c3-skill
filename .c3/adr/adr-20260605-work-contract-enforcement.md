---
id: adr-20260605-work-contract-enforcement
c3-seal: 66baf5da72035dd63196dea87c48b47fa0478ff1d710c29b899a798a53fb3503
title: work-contract-enforcement
type: adr
goal: 'Implement the c3x work contract as executable behavior: canvas-defined content types must author and validate through the same definition, lookup and graph context must feed a hybrid search command, and read-only agent work must avoid unnecessary canonical preverify while still exposing verifiable command surfaces.'
status: implemented
date: "2026-06-05"
---

## Goal

Implement the c3x work contract as executable behavior: canvas-defined content types must author and validate through the same definition, lookup and graph context must feed a hybrid search command, and read-only agent work must avoid unnecessary canonical preverify while still exposing verifiable command surfaces.

## Context

Current review proved that canvas listing/schema/check works for built-in definitions, lookup and graph work for code-map and relationship traversal, and store-level FTS exists. It also proved gaps: project-defined document canvases cannot be added as first-class entity content, table/column validation is not shared uniformly across add/write/check, no public hybrid search command exists, and read-only commands run the expensive canonical verification path before doing cache-backed context work. The affected topology is the Go CLI under c3-1, especially add/write/check, store/search, lookup/graph, runtime dispatch, and help/capabilities.

## Decision

Add BDD-style red tests first with concrete research-note, lookup, hybrid search, and read-only drift fixtures. Then implement minimal green paths: generic canvas-defined document entities, shared body validation for required sections/tables/typed columns, a search command that merges metadata FTS, content FTS, and graph context, and a read-only fast path that trusts the local cache unless the user explicitly runs check/repair or the cache is missing. Command outputs added for agents must use the shared output helpers and actionable errors.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-1 | container | Owns CLI file-system operations, numbering, codemap matching, validation, and binary command behavior that this work changes. | c3-1#n1370@v1:sha256:f042753b02d7565bc9f051f933dd9b4da6c4fc04079eba6a5620cab6a75bec83 "Own all file-system read/write for .c3/ architecture docs" | Review parent responsibilities after command and validation changes; record Parent Delta. |
| c3-111 | component | Add must create canvas-defined document entities, not only fixed built-in entity types. | c3-111#n1879@v1:sha256:e6fad9e0289bcae50e2398556208897e696140a220eb4c0d3bb5d74085bf8d00 "Create new containers, components, refs, rules, or ADRs with correct numbering and wired into the parent doc." | Ensure numbering, IDs, body validation, and canonical export stay deterministic. |
| c3-117 | component | Write/schema/canvas/read behavior must preserve canvas definition as the content contract. | c3-117#n2154@v1:sha256:3a377de9129c41b25fa893dc4431afb0aad2814a699e530c70d0de2567204ff6 "Read, write, set, validate schema, and report status for canonical C3 documents." | Verify write uses shared definition validation and docs-state command output remains agent-safe. |
| c3-113 | component | Check must enforce the same required sections, tables, columns, and primitive semantics as add/write. | c3-113#n1974@v1:sha256:cf24ddb428f8aecd7dc88a46115e6dfaf1a4a29525c628bdaed6f02ce94b4435 "Validate structural integrity of .c3/ docs, ref and rule compliance — required fields, numbering, wiring, scope cross-checks, origin validation." | Keep focused validator failures traceable to one owning layer. |
| c3-118 | component | Analysis gains the public hybrid search command alongside graph behavior. | c3-118#n2201@v1:sha256:bcb7f4508335a6b27c0f688e21aa7450580a18e9271304fbdc49ab162fe57937 "Query, graph, diff, and impact commands for understanding architecture state and change consequences." | Confirm search output includes FTS hits, graph context, snippets, and source labels. |
| c3-107 | component | Store adds or exposes combined search over entity metadata, node content, relationships, and codemap context. | c3-107#n1671@v1:sha256:b4745317e94e8953c629e7c21068db0444d6b010b45ee305639f581f6047c9ef "Provide persistent entity, relationship, changelog, codemap, hash, node, and version storage operations for the CLI." | Verify store search remains deterministic and does not hide graph relationship errors. |
| c3-108 | component | Runtime dispatch, preverify policy, capabilities, and agent output formatting change. | c3-108#n1715@v1:sha256:ae80704ae7172ccccc82f6ba7b67f4fe434e41a8e3571e164a7b2165e4e4f06b "Provide CLI bootstrap, option parsing, output shaping, config resolution, and human/agent presentation helpers." | Ensure read-only fast path is explicit, errors keep hints, and structured outputs stay TOON in agent mode. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-frontmatter-docs | Generic canvas-defined entities still use C3 markdown frontmatter plus body as the persisted user-owned content format. | ref-frontmatter-docs#n2859@v1:sha256:2113120972193c5dccc7b63aec1ce6de17e31f2aa403b0262fa3fee03501307f "Every .c3/ doc uses YAML frontmatter for machine-readable metadata and a Markdown body for human-readable content." | Comply; generic entities must preserve frontmatter/body parsing and canonical export. |
| ref-cross-compiled-binary | The affected CLI container cites distribution behavior, but this implementation does not change build targets or binary packaging. | ref-cross-compiled-binary#n2826@v1:sha256:aef925fd21543311cf7dac4a5afcfc8663f3e4e0e1a592ec0c772f4edea7b46f "CLI is distributed as pre-built binaries for 4 targets so users need no Go toolchain to use c3x." | Review only; no distribution code change expected. |
| ref-embedded-templates | The affected CLI container cites embedded template behavior, but this implementation works through canvas definitions and does not alter template embedding. | ref-embedded-templates#n2844@v1:sha256:db227a0598059041ce49f2746f25fc8501e6ae86b9c9f688f593279ecde35ff8 "Doc templates are bundled in the CLI binary so scaffolding works without external files at any install path." | Review only; no embedded template code change expected. |
| N.A - no additional ref | No other existing ref directly governs hybrid search ranking or read-only preverify policy in this repository. | N.A - no citation exists for absent governing ref. | Review after implementation; create a ref only if the search/ranking pattern becomes reusable across commands. |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| rule-output-via-helpers | New search/capabilities structured outputs must remain TOON in agent mode and avoid command-local serialization. | rule-output-via-helpers#n2893@v1:sha256:4c8f10ca43e5ccb0607defedf34c5c672702e512f4fed9259f8a5fa19ef866e3 "All command results serialize through one output layer so agent mode always yields TOON and human/JSON formats stay consistent across commands." | Comply in new command output paths. |
| rule-dispatcher-error-hint | New parser/dispatch failures and unsupported command paths need actionable error: and hint: text. | rule-dispatcher-error-hint#n2873@v1:sha256:c88ba0daf18b558254516480e305ec64cacb996ef674a8540d5b99e103488cb4 "User-facing CLI errors from the top-level dispatcher guide the user to a next step, so a failure is actionable rather than a bare message." | Comply in search/add/write validation errors. |
| rule-wrap-error-cause | New store/search/canvas validation errors must preserve underlying causes at layer boundaries. | rule-wrap-error-cause#n2913@v1:sha256:20a5bd788231e5b7b7403d387c6414f0d5b8b31303d720d83369abbe18c9ab26 "Every returned error in the Go CLI preserves its cause and context so failures stay traceable across the dispatcher, store, and command layers." | Comply where lower-level calls can fail. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| BDD tests | Add concrete tests for research-note canvas add/write/check, lookup/graph fixture, hybrid search result context, and read-only preverify behavior. | cli/cmd/add_test.go; cli/cmd/readwrite_test.go; cli/cmd/check_enhanced_test.go; cli/cmd/search_test.go; cli/main_test.go. |
| Generic canvas entities | Allow project-defined canvas IDs such as research-note to create/read/write/check first-class content while preserving canonical docs and store rows. | c3-111 and c3-117 affected topology rows. |
| Shared validation | Reuse one definition-driven validator for add/write/check required sections, table rows, column presence, and typed primitives. | c3-113 affected topology row. |
| Hybrid search | Add store and command layer that combines entity FTS, content FTS, graph context, codemap paths, snippets, and match source labels. | c3-107 and c3-118 affected topology rows. |
| Runtime efficiency | Change read-only dispatch so lookup/read/graph/search do not run canonical preverify warning path when cache exists; keep check/repair as explicit verification commands. | c3-108 affected topology row. |
| Help/capabilities | Add search to help/capabilities and make add help say types come from canvas list rather than a fixed complete list. | rule-output-via-helpers and rule-dispatcher-error-hint rows. |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| Tests | Add BDD-style test cases using concrete fixtures for canvas content, lookup/graph, hybrid search, and read-only fast path. | go test ./cmd ./internal/store ./internal/schema from cli. |
| Validator | Move shared section/table/column/primitive validation into a reusable path invoked by add, write, and check. | go test ./cmd -run TestRunAdd and go test ./cmd -run TestRunWrite and go test ./cmd -run TestRunCheck. |
| Store/search | Add hybrid search projection with content and graph context. | go test ./internal/store -run TestHybridSearch. |
| CLI command | Add search command parser, dispatcher, help, capabilities, and agent-safe output. | go test . ./cmd -run TestRun_Search and go test ./cmd -run TestShowHelp and go test . -run TestRun_Capabilities. |
| Runtime | Change read-only preverify policy and add test proving lookup ignores stale canonical seal when cache exists. | go test . -run TestRun_ReadOnlyLookupSkipsCanonicalPreverify. |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| c3x add/write/check | Reject missing research-note required sections, missing table columns, invalid citations, invalid entity_id, invalid enum, placeholder check, and placeholder edge values. | cli/cmd/add_test.go; cli/cmd/readwrite_test.go; cli/cmd/check_enhanced_test.go. |
| c3x lookup/graph | Resolve file path to owner, refs, rules, and graph context for authoring. | cli/cmd/lookup_test.go; cli/cmd/graph_test.go. |
| c3x search | Return hybrid results with match sources, snippets, graph context, and codemap path labels. | cli/cmd/search_test.go; cli/internal/store/search_test.go. |
| c3x capabilities/help | Public surface includes search and add help points to canvas list as type source. | cli/cmd/help_test.go; cli/main_test.go. |
| c3x check | Remains explicit full validation and continues to pass after canonical state updates. | C3X_MODE=agent bash skills/c3/bin/c3x.sh check. |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Only add tests for existing graph/lookup and leave hybrid search absent. | The user success story explicitly requires searchable content by hybrid RAG, so this would preserve the current gap instead of enforcing the work. |
| Build generic canvas entities without shared add/write/check validation. | The current failure is contract drift between authoring and validation; separate validators would keep that drift risk alive. |
| Run canonical preverify for every read-only command forever. | It makes agent context gathering inefficient and contradicts the desired efficient work path; explicit check/repair already owns full verification. |
| Add vector embeddings before FTS plus graph hybrid search. | The repo already has SQLite FTS and graph/codemap relationships; a deterministic local hybrid search can satisfy the contract without introducing network/model dependencies. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Generic entity types break SQLite type constraints or canonical export paths. | Add failing tests around research-note creation and update store/export schema deliberately. | go test ./cmd -run TestRunAdd_CanvasDefinedEntity. |
| Shared validation changes component strictness behavior unexpectedly. | Keep component strict rules as extra checks after definition validation and run focused strict component tests. | go test ./cmd -run TestStrictComponentDocs. |
| Hybrid search ranking becomes nondeterministic. | Use deterministic source weights and stable ID sort for tie breaks. | go test ./internal/store -run TestHybridSearch. |
| Read-only fast path hides real canonical drift. | Keep c3x check/repair explicit and make mutating commands refresh cache before mutation as they do today. | go test . -run TestRun_ReadOnlyLookupSkipsCanonicalPreverify plus C3X_MODE=agent bash skills/c3/bin/c3x.sh check. |

## Verification

| Check | Result |
| --- | --- |
| GOCACHE=/tmp/go-cache go test ./cmd -count=1 | PASS; command package validates canvas add/write/check, hybrid search, help, export/import, and command output paths. |
| GOCACHE=/tmp/go-cache go test . -run 'TestBDD_(ReadOnlyLookupSkipsCanonicalPreverifyWhenCacheExists | RunSearchHybridJSONDispatch)' -count=1 |
| GOCACHE=/tmp/go-cache go test ./internal/store -run 'TestOpen_MigratesEntityTypeCheckForCanvasTypes | TestSearchContent' -count=1 |
| GOCACHE=/tmp/go-cache go test ./... -count=1 | PASS; all CLI packages passed. Sandbox prints devbox read-only unlink warning before tests, but exit code is 0. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check | PASS before final ADR status transition; total 77, no issues. |
