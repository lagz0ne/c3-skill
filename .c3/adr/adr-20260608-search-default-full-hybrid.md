---
id: adr-20260608-search-default-full-hybrid
c3-seal: 9a1afc2e088a08129185d5d27d726a076caccbc3d878c9e3872212020e22f8cf
title: search-default-full-hybrid
type: adr
goal: 'Make `c3x search <query>` default to the highest-value retrieval path: semantic embeddings, keyword/BM25 FTS, and graph/context expansion fused together, without requiring a manual `c3x index` first.'
status: implemented
date: "2026-06-08"
---

## Goal

Make `c3x search <query>` default to the highest-value retrieval path: semantic embeddings, keyword/BM25 FTS, and graph/context expansion fused together, without requiring a manual `c3x index` first.

## Context

Search currently fuses semantic hits only when the caller explicitly asks for semantic indexing or when a usable semantic index already exists. A fresh checkout therefore defaults to keyword and graph results even though the semantic model path exists for fat builds and the thin build can cache or download assets. The affected topology is the Go CLI search command boundary, runtime option/help behavior, and store-owned semantic index lifecycle.

## Decision

Enable semantic search by default unless `--no-semantic` is passed. Before ranking, search asks the store to ensure the semantic index is fresh by comparing current entity semantic text hashes against stored embedding hashes, embedding only missing or changed entities, deleting stale rows, and reusing fresh vectors on repeat searches. If semantic assets or runtime are genuinely unavailable, search falls back to keyword plus graph results without failing the command. Manual `c3x index` remains an explicit full rebuild path.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-1 | container | Search behavior and semantic asset use are inside the Go CLI command and storage boundary. | c3-1#n1624@v1:sha256:1f10b60594a91ba79c9f39e43539257e287197fb4d527506c35b42a09cf7bb95 "Compile to a self-contained binary for all supported platforms" | Review parent responsibilities for no-delta or update after implementation. |
| c3-107 | component | Store-lib owns semantic vectors, entity hashes, and the index freshness/reuse contract. | c3-107#n1921@v1:sha256:b4745317e94e8953c629e7c21068db0444d6b010b45ee305639f581f6047c9ef "Provide persistent entity, relationship, changelog, codemap, hash, node, and version storage operations for the CLI." | Add store tests proving fresh, reused, and stale incremental indexing. |
| c3-108 | component | Runtime-support owns search option parsing/help and agent/human command output boundaries. | c3-108#n1965@v1:sha256:ae80704ae7172ccccc82f6ba7b67f4fe434e41a8e3571e164a7b2165e4e4f06b "Provide CLI bootstrap, option parsing, output shaping, config resolution, and human/agent presentation helpers." | Keep output on shared helpers and preserve actionable error behavior. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-cross-compiled-binary | Fat/thin distribution defines whether a semantic model is embedded, cached, or downloaded for default search. | ref-cross-compiled-binary#n3078@v1:sha256:90bdc1dce31722c7d28535fa7aaf8b4ae1ea7ebb6955aaa1c7a44d15305b9e1c "Standardize how C3 distributes platform-specific Go executables and semantic model assets without forcing every install channel to carry every large binary blob." | review |
| ref-frontmatter-docs | This ADR and any C3 doc changes must remain canonical frontmatter markdown managed by c3x. | ref-frontmatter-docs#n3111@v1:sha256:27959709c9aa210a07dabf189c66e2aae5c1376d69c62343c583a0f8e2ee7d9a "tables and diagrams" | comply |
| ref-embedded-templates | The fat build path uses embedded semantic assets alongside the existing embedded-template convention, so this default-search change must not alter template embedding behavior. | ref-embedded-templates#n3090@v1:sha256:db227a0598059041ce49f2746f25fc8501e6ae86b9c9f688f593279ecde35ff8 "Doc templates are bundled in the CLI binary so scaffolding works without external files at any install path." | review |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| rule-output-via-helpers | Search output must continue through shared output helpers so agent mode stays TOON and JSON stays explicit. | rule-output-via-helpers#n3139@v1:sha256:4c8f10ca43e5ccb0607defedf34c5c672702e512f4fed9259f8a5fa19ef866e3 "All command results serialize through one output layer so agent mode always yields TOON and human/JSON formats stay consistent across commands." | comply |
| rule-wrap-error-cause | Store/search errors added around semantic freshness and model fallback must preserve causes. | rule-wrap-error-cause#n3159@v1:sha256:20a5bd788231e5b7b7403d387c6414f0d5b8b31303d720d83369abbe18c9ab26 "Every returned error in the Go CLI preserves its cause and context so failures stay traceable across the dispatcher, store, and command layers." | comply |
| rule-dispatcher-error-hint | User-facing failures remain actionable; semantic model absence should not become a dispatcher crash. | rule-dispatcher-error-hint#n3119@v1:sha256:c88ba0daf18b558254516480e305ec64cacb996ef674a8540d5b99e103488cb4 "User-facing CLI errors from the top-level dispatcher guide the user to a next step, so a failure is actionable rather than a bare message." | review |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| cli/internal/store | Add semantic index freshness detection that compares current entity text hashes with stored embedding hashes and upserts only changed/missing entities. | Store unit tests with a fake embedder count calls and stale refresh behavior. |
| cli/cmd/search.go | Default search to semantic-on plus graph hybrid, auto-ensure the index, fuse semantic hits, and fall back only when semantic is unavailable. | Search unit tests and CLI smoke show match_sources includes semantic on a fresh index. |
| cli/cmd/options.go and help | Preserve --semantic/--no-semantic; document that --semantic is compatibility and --no-semantic is opt-out. | Options/help tests and help text inspection. |
| cli/tools/search-eval | Make the no-flag eval path use the new default semantic hybrid and add an explicit opt-out baseline. | Eval before/after metrics show default moves from keyword baseline to semantic numbers. |
| C3 docs | Record default behavior change in this ADR. | c3x check --include-adr. |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| Search command runtime | RunSearch treats semantic and graph expansion as default search behavior, while --no-semantic disables only semantic. | go test ./cmd and fresh CLI smoke. |
| Store semantic index | New store ensure path detects absent/stale/current vectors and embeds only missing or changed entities. | go test ./internal/store. |
| Search eval harness | Default eval calls no longer pass NoSemantic; opt-out baseline is explicit. | go run ./tools/search-eval --db ../.c3/c3.db --k 5 and --no-semantic. |
| Help/options surface | Help text describes semantic-on default and fallback behavior. | go test ./cmd -run Test plus help text review. |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| Store tests | Fresh index builds all vectors; repeat ensure embeds zero entities; stale entity embeds one changed entity. | go test ./internal/store. |
| Search tests | Default search degrades when the model is unavailable and opt-out excludes semantic sources. | go test ./cmd. |
| CLI smoke | Fresh repo with no semantic vectors returns fused results with semantic in match_sources before any manual c3x index. | Local c3local search ... --json smoke after deleting embedding rows. |
| Eval harness | Default no-flag eval reports semantic hybrid metrics instead of keyword baseline. | Search eval default and opt-out runs. |
| Build matrix | Thin and fat builds compile cleanly. | go build ./... and go build -tags embedmodel ./.... |
| C3 check | ADR and canonical docs stay valid. | c3local check --include-adr. |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Keep requiring manual c3x index | Fresh default search stays least valuable and violates the requested default contract. |
| Rebuild all embeddings on every search | Repeat search would run real ONNX inference across all entities and make default search slow. |
| Silently disable semantic unless an index already exists | This preserves the current low-value default and hides the model capability. |
| Fail search when semantic assets are missing | Offline thin users would lose keyword plus graph search even though that path is still useful. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Repeat search becomes slow | Compare entity text hashes and only embed changed/missing rows. | Store fake embedder count and CLI repeat timing/update evidence. |
| Offline thin build hangs or crashes | Treat semantic unavailable errors as non-fatal for search and keep keyword plus graph fallback. | Empty-cache offline smoke with C3_SEMANTIC_OFFLINE=1. |
| Stale vectors survive entity edits | Hash current semantic text including metadata, nodes, and code-map paths before each ensure. | Store stale-refresh test mutates an entity and verifies exactly one new embed. |
| Eval default remains keyword-only | Change harness default and prove no-flag metrics match semantic run. | Eval before/after output. |
| Help text drifts from behavior | Update search help to state semantic-on default and opt-out. | Help test/build plus final diff. |

## Verification

| Check | Result |
| --- | --- |
| Fresh c3x search <term> with no semantic rows includes semantic in match_sources without manual c3x index. | PASS: deleted entity_embeddings; C3X_MODE=agent bash skills/c3/bin/c3x.sh search "which architecture unit owns a source path" --limit 5 returned rows with match_sources containing semantic; first run elapsed 2.62s and built 80 embeddings. |
| Repeat c3x search <term> reuses the existing semantic index and does not re-embed all entities. | PASS: repeat search elapsed 0.21s; entity_embeddings stayed at 80 rows and min/max updated_at stayed 2026-06-08 11:31:50 after a delayed repeat. |
| Empty model cache plus offline env degrades to keyword plus graph with no crash or hang. | PASS: copied DB with zero embeddings plus C3_SEMANTIC_CACHE_DIR=../.tmp-empty-semantic-cache C3_SEMANTIC_OFFLINE=1 go run . --c3-dir ../.tmp-offline-c3 search "which architecture unit owns a source path" --limit 5 returned keyword/graph JSON in 0.65s and no semantic sources. |
| go run ./tools/search-eval --db ../.c3/c3.db --k 5 default reports semantic hybrid metrics. | PASS: mode keyword_graph_onnx_semantic, overall Hit@5 0.875, paraphrase Hit@1 0.25. |
| go run ./tools/search-eval --db ../.c3/c3.db --k 5 --no-semantic reports keyword baseline metrics. | PASS: mode keyword_graph, overall Hit@5 0.625. |
| cd cli && go build ./... | PASS. |
| cd cli && go build -tags embedmodel ./... | PASS. |
| cd cli && go test ./... | PASS. |
| git diff --check | PASS. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr | PASS exit 0; full ADR scan still reports legacy ADR warnings outside this change. Focused --only adr-20260608-search-default-full-hybrid is clean. |
