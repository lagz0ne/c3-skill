---
id: adr-20260610-search-hyphen-cli-v10
c3-seal: 07a4936012835e780b8edbb043b13c2ce81abcd0b8e2056c527e0bc79337927a
title: search-hyphen-cli-v10
type: adr
goal: Fix the v10 `c3x search "real-time sync"` hyphen query failure so conceptual searches treat hyphenated natural-language terms as data, not SQLite FTS syntax, and move the npm `@c3x/cli` manager package version to the v10 line.
status: implemented
date: "2026-06-10"
---

## Goal

Fix the v10 `c3x search "real-time sync"` hyphen query failure so conceptual searches treat hyphenated natural-language terms as data, not SQLite FTS syntax, and move the npm `@c3x/cli` manager package version to the v10 line.

## Context

The v10 release introduced `c3x search` as the first conceptual-discovery command. A user-reported query, `c3x search "real-time sync"`, fails with `SQL logic error: no such column: time`, which indicates the query text is reaching SQLite FTS parsing in a shape where `real-time` is parsed as an expression instead of a literal term. The affected code is in the Go CLI analysis/search path and semantic store query path. The same branch also has the skill and bundled c3x release at v10, while `packages/cli` still needs to move toward the 10 line for the npm manager package.

## Decision

Normalize the SQLite FTS query construction used by `c3x search` so user input tokens, including hyphenated terms, are escaped or quoted as literals before reaching `MATCH`. Add regression coverage for `real-time sync` and nearby punctuation cases. Keep command output routed through existing helpers and wrap store/command errors with context. Bump `packages/cli` package metadata to `10.0.0` if the package is still below v10, without changing the npm manager's thin-only behavior.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-1 | container | The Go CLI owns c3x search, release packaging, tests, and npm manager packaging. | c3-1#n1699@v1:sha256:b295c9d36ef741a0e4efaea40550b4d7444f4b3bcae9ad114caa21b3d4cfc19b "Match source files to architecture components via codemap patterns" | Confirm parent responsibilities do not need a broader contract change beyond search correctness and package version metadata. |
| c3-107 | component | Semantic store code owns persisted semantic/keyword search behavior and the SQLite query path that can reject hyphenated input. | c3-107#n1998@v1:sha256:b4745317e94e8953c629e7c21068db0444d6b010b45ee305639f581f6047c9ef "Provide persistent entity, relationship, changelog, codemap, hash, node, and version storage operations for the CLI." | Verify store-layer query escaping with targeted unit tests. |
| c3-118 | component | Analysis commands own search command behavior and user-facing conceptual discovery. cli/cmd/search.go is currently uncharted and should be mapped here. | c3-118#n2531@v1:sha256:bcb7f4508335a6b27c0f688e21aa7450580a18e9271304fbdc49ab162fe57937 "Query, graph, diff, and impact commands for understanding architecture state and change consequences." | Add/verify codemap coverage for search command files and run command-level tests. |
| c3-109 | component | npm @c3x/cli manager owns package metadata and publication versioning. | c3-109#n2109@v1:sha256:8ce64dfe23c3415aa4fc79349bfbaa9294d928eeadd1584584d6421e0c73ae1c "downloads, verifies, caches, and execs the pinned thin C3 release binary" | Bump package metadata only; do not change manager behavior unless tests prove it is needed. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-cross-compiled-binary | Governs npm manager version/cached release binary behavior and thin-only npm distribution. | ref-cross-compiled-binary#n3157@v1:sha256:d4e8aff3ebc7fabf348e6da10383569e74e9cafae95aba2c02a73f285f7e30be "Publish two install paths with one default." | comply |
| ref-embedded-templates | Inherited Go CLI governance because the affected container includes template/scaffold components; this change must not alter template embedding behavior. | ref-embedded-templates#n3167@v1:sha256:db227a0598059041ce49f2746f25fc8501e6ae86b9c9f688f593279ecde35ff8 "Doc templates are bundled in the CLI binary so scaffolding works without external files at any install path." | review |
| ref-frontmatter-docs | Inherited Go CLI governance because the affected container includes frontmatter/content components; this change must preserve metadata/body handling. | ref-frontmatter-docs#n3182@v1:sha256:2113120972193c5dccc7b63aec1ce6de17e31f2aa403b0262fa3fee03501307f "doc uses YAML frontmatter for machine-readable metadata" | review |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| rule-output-via-helpers | c3x search output must continue to use shared output helpers and keep agent-mode TOON behavior. | rule-output-via-helpers#n3218@v1:sha256:ca4a0c295ace95c5e32a1c7ca5b92a583f8637573bdec86d8513c1c6d7d486be "Commands must emit results via" | comply |
| rule-wrap-error-cause | Store/command error changes must preserve causes with %w. | rule-wrap-error-cause#n3238@v1:sha256:4de41fc55bd76795f15eb969b75c91efe9caa49ab89d27986626463fd6e25af7 "All returned errors must wrap the cause" | comply |
| rule-dispatcher-error-hint | If the bug still reaches a user-facing dispatcher error, it must stay actionable. | rule-dispatcher-error-hint#n3198@v1:sha256:182738dc1f72a50edf736dd26e32f990ba604b9e04a107dcb772922f769575e1 "User-facing dispatcher errors must carry an" | review |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Reproduce | Run C3X_MODE=agent bash skills/c3/bin/c3x.sh search "real-time sync" and/or a focused Go test to capture the SQL error. | Failing command/test output before the fix. |
| Search query fix | Update the store/search FTS query builder so hyphenated user tokens are literal query terms. | Focused go test on search/semantic tests. |
| Command coverage | Add or update cli/cmd/search_test.go coverage for the command path if the bug is command-owned. | go test ./cmd -run Search. |
| Codemap coverage | Map cli/cmd/search.go and related search command tests to c3-118 if lookup remains uncharted. | c3x lookup cli/cmd/search.go returns c3-118. |
| npm version | Move packages/cli/package.json and packages/cli/package-lock.json to 10.0.0 if still below 10. | package diff, npm test, and npm run build from packages/cli. |
| Parent delta | Parent Delta: none - c3-1 responsibilities already cover codemap, validation, cross-compiled release binaries, and npm wrapper delegation; c3-0 topology is unchanged. | c3x read c3-1, c3x lookup cli/cmd/search.go, c3x lookup packages/cli/src/version.ts, and c3x check. |
| Verification evidence | Update this ADR verification table before implementation status. | c3x read adr-... --full. |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| c3x search command | Escape or quote FTS query terms so hyphenated input such as real-time does not become a column/expression lookup. | c3x search "real-time sync" exits successfully. |
| Store semantic search tests | Add regression coverage for hyphenated natural-language query terms. | go test ./internal/store -run Search or equivalent exact test. |
| Analysis command tests | Add command-level coverage if search command parsing/output participates in the failure. | go test ./cmd -run Search. |
| Codemap | Add missing search command file mapping to c3-118 if required. | c3x lookup cli/cmd/search.go. |
| npm package metadata | Move @c3x/cli package metadata to v10 while preserving thin-only manager behavior. | cd packages/cli && npm test && npm run build. |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| cli/internal/store/search tests | Hyphenated FTS terms must not produce SQLite column/operator errors. | Targeted Go test. |
| cli/cmd/search_test.go | c3x search "real-time sync" command path remains successful and output-helper based. | Targeted Go test. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh search "real-time sync" | Live local smoke test against this repo's C3 docs. | Command exits 0. |
| packages/cli tests/build | npm package version move does not break thin manager package. | npm test and npm run build. |
| c3x check | C3 docs and codemap remain consistent after ADR/codemap changes. | C3X_MODE=agent bash skills/c3/bin/c3x.sh check. |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Strip hyphens from all search input. | It would make real-time searchable but lose user intent and harm exact keyword matching for hyphenated architecture terms. |
| Replace FTS search with only LIKE queries. | It would avoid FTS syntax errors but throw away the v10 hybrid semantic/keyword search design and likely degrade ranking. |
| Catch the SQLite error and retry with a simplified query only. | It leaves the primary query builder unsafe and makes correctness depend on an error path. |
| Leave npm package below 10. | The user explicitly asked to move CLI toward 10, and package metadata can align with the v10 release without changing manager behavior. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Escaping FTS input reduces search recall. | Preserve tokenized terms and phrase/literal semantics rather than dropping punctuation wholesale. | Existing search tests plus hyphen regression test. |
| Fix only covers real-time but misses other punctuation. | Add tests for at least hyphen and quoted multi-word user input. | Targeted store tests. |
| npm version bump accidentally changes manager behavior. | Limit npm edits to package metadata unless tests require otherwise. | git diff -- packages/cli review plus npm tests/build. |
| Codemap mutation causes C3 drift. | Use c3x set c3-118 codemap ... --append and run c3x check. | C3 check and lookup evidence. |

## Verification

| Check | Result |
| --- | --- |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh search "real-time sync" before fix | FAIL reproduced - search content: SQL logic error: no such column: time (1). |
| Targeted Go test for hyphenated search query | PASS - new store/cmd regression tests failed before fix and pass after splitting hyphenated terms into safe FTS tokens. |
| go test ./cmd -run Search from cli/ | PASS. |
| go test ./internal/store -run Search from cli/ | PASS. |
| go test ./... from cli/ | PASS. |
| cd packages/cli && npm test | PASS - 5 node tests passed after local npm install. |
| cd packages/cli && npm run build | PASS - tsdown built cli, manager, and version entries. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh lookup cli/cmd/search.go | PASS - maps to c3-118 after codemap update. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh search "real-time sync" after fix | PASS - returns results instead of SQL error using rebuilt local linux/amd64 wrapper binary. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check | PASS - total: 81, issues: empty. |
