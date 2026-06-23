---
id: adr-20260623-c3x-runtime-help-progress
c3-seal: 80246d3b33f87a3bff045b90062be887574d22e185d0dcc59162d930a10ee30b
title: c3x-runtime-help-progress
type: adr
goal: Keep wrapper-local discovery commands download-free while making first-use runtime downloads visible and bounded to operations that actually need runtime assets.
status: done
date: "2026-06-23"
---

# C3x Runtime Help And Progress

## Goal

Keep wrapper-local discovery commands download-free while making first-use runtime downloads visible and bounded to operations that actually need runtime assets.

## Context

The npm runtime manager can now resolve and install GitHub Release runtimes, but the root `c3x --help` and empty `c3x` paths should not fetch release metadata or runtime assets just to explain usage. Conversely, a real forwarded C3 command that must populate an empty cache should not sit silently while binary, model, and vocab assets download. This affects only the npm distribution container and its binary-downloader component.

## Decision

Serve root help, empty argv, and root version locally from the npm manager without resolving a runtime. Real forwarded commands still prepare and exec the selected runtime, but automatic cache population reports throttled stderr progress with a text progress bar. Explicit `c3x runtime install` keeps the fuller install progress surface. Runtime-manager read-only commands such as `runtime installed` stay local; `runtime versions` may fetch only the release index.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-0 | system | Its container membership row must describe the npm client as a runtime manager with local discovery commands, not the older pinned downloader. | c3-0#n9@v1:sha256:0887f802f58de6a8d2a359d78a0e3540a4a771c330087776c9b81a901e80a447 | Parent system row updated. |
| c3-3 | container | Its responsibilities now distinguish local wrapper discovery from runtime-preparing commands. | c3-3#n635@v1:sha256:2c16db898e7fd9ef2949b68f05efcbbee8252d7b8557d2a0d9f17b8b012784ce | Parent responsibility updated. |
| c3-301 | component | Its CLI-entry and asset-fetch contract now covers download-free root help/version and visible progress when assets are actually fetched. | c3-301#n658@v1:sha256:0bde9d5d94e0836366f5980b45d577c25b66b801e624a1e528f69fcabb1263fa | Component contract and purpose updated. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-cross-compiled-binary | Runtime downloads still have to target the release matrix assets that CI produces. | ref-cross-compiled-binary#n730@v1:sha256:cea27fa9abdd975d6298f23e899ceb48f2945fde1e933915fdfada258a190136 "Ship c3x as named per-platform release binaries, with standard runtime-manager binaries," | comply |
| ref-fat-thin-distribution | The npm client remains thin and fetches runtime assets on demand; this ADR narrows when that fetch is allowed within the npm surface. | ref-fat-thin-distribution#n748@v1:sha256:cf8e08bcc48d161ef4120914a8505499068ac512709adeae244258fd1618b031 "Distribute C3 as full-fat skill ZIPs (semantic binaries bundled), Linux portable fat skill" | comply |
| ref-eval-determinism | Inherited from the top-level system; this work does not change eval verdict computation or fixtures. | ref-eval-determinism#n739@v1:sha256:d914f393b17de0202b7ae4cdde4df7d173c51fd820b2695487a29efb06f514d7 | N.A - runtime help/progress does not change eval computation. |
| ref-frontmatter-docs | This ADR updates frozen C3 facts and must preserve frontmatter plus canvas-shaped markdown. | ref-frontmatter-docs#n757@v1:sha256:d4f7719668519e2f2a93de15969bc53c8f0105e7e073231a2f36d7c2626cb361 | comply |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| rule-dispatcher-error-hint | The npm manager still emits user-facing errors for unknown runtime commands and metadata/download failures. | rule-dispatcher-error-hint#n766@v1:sha256:bd662000c1bc5b93d0b1cc4cf532cc1dc6e4766e5bda6b544f8aab14d21f7dc4 | comply |
| rule-output-via-helpers | Inherited from Go CLI components; npm manager help/progress is human package-management output, not agent-mode structured command output. | rule-output-via-helpers#n779@v1:sha256:b5ac8121ffc54be6c8f87ec133e69658fea023e7e73da3859fb85a33869afa29 | N.A - normal Go commands still own structured output. |
| rule-wrap-error-cause | Inherited from Go CLI components; npm manager download and metadata failures still need contextual causes. | rule-wrap-error-cause#n791@v1:sha256:b9e4edb84b11060973de3fe6e5c0ab7b5605aa690e00e886335b054bdaab710f | comply |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Runtime manager | Add root help/version short-circuit before runtime resolution and use default progress reporting during automatic runtime preparation. | packages/cli/src/manager.ts |
| Tests | Cover help/no-arg/version paths without downloads, automatic download progress, and progress throttling. | packages/cli/test/manager.test.mjs |
| Docs | Document download-free help/version and real-command progress behavior. | packages/cli/README.md; docs/superpowers/specs/2026-03-17-c3x-npm-cli-design.md |
| C3 facts | Align c3-0, c3-3, and c3-301 with the wrapper-local discovery and progress contract. | c3-0; c3-3; c3-301 |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| runCli tests | Downloader and exec hooks remain untouched for --help, empty argv, and --version. | packages/cli/test/manager.test.mjs |
| progress tests | Real forwarded command emits automatic download progress with a text bar. | packages/cli/test/manager.test.mjs |
| C3 check | Frozen facts and ADR topology remain structurally valid. | c3local check --include-adr |

## Verification

| Check | Result |
| --- | --- |
| cd packages/cli && npm test | Required before done. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr | Required before done. |
