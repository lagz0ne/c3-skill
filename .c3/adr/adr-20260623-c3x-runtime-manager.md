---
id: adr-20260623-c3x-runtime-manager
c3-seal: ccac6f68dbc44d07bb5dcd5ec6766e1de2c6e60c4770decf99f6a76d5278c393
title: c3x-runtime-manager
type: adr
goal: Make the npm `@c3x/cli` entrypoint a runtime manager for GitHub Release assets while preserving normal C3 command passthrough. The package must list available runtimes, install selected runtimes with visible progress, store only safe per-project runtime metadata, select the runtime on command execution, and remove cache entries only through explicit uninstall or prune commands.
status: superseded
date: "2026-06-23"
---

# C3x Runtime Manager

## Goal

Make the npm `@c3x/cli` entrypoint a runtime manager for GitHub Release assets while preserving normal C3 command passthrough. The package must list available runtimes, install selected runtimes with visible progress, store only safe per-project runtime metadata, select the runtime on command execution, and remove cache entries only through explicit uninstall or prune commands.

## Context

The previous npm package fact described a pinned thin downloader that garbage-collected other versions while preparing one runtime. That no longer matches the distribution model: the binary is no longer controlled by a bundled skill artifact, and the npm manager must own release lookup, cache lifecycle, project runtime selection, and explicit cache cleanup. The affected topology is the C3 product distribution system, the `c3-3` npm client container, and its `c3-301` binary-downloader component.

## Decision

Use a namespaced manager surface, `c3x runtime ...`, so runtime management cannot collide with existing Go CLI commands such as `c3x list`. Normal commands continue to resolve and exec the Go binary. Runtime resolution uses `C3X_VERSION` for development and tests, then `.c3/runtime.json` operational project metadata, then the latest non-draft non-prerelease GitHub Release, with the newest installed runtime as an offline fallback. Project metadata may contain only a version and update timestamp; it may not contain a URL, path, command, checksum, or release source. Asset downloads still verify SHA256 before atomic activation, and cache cleanup becomes explicit through `runtime uninstall` and `runtime prune`.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-0 | system | The npm runtime manager changes one product distribution surface while remaining inside the current Go CLI, agent skill, and npm runtime-manager system shape. | c3-0#n3@v1:sha256:4295e84171aebab432093423315f2571c121a774d6338bf7330dd42644c6dfc2 | Top-down context reviewed; no system body patch needed for the npm runtime-manager change. |
| c3-3 | container | The npm package responsibility changes from pinned runtime forwarding to runtime-manager ownership plus passthrough. | c3-3#n635@v1:sha256:2c16db898e7fd9ef2949b68f05efcbbee8252d7b8557d2a0d9f17b8b012784ce | Parent Delta: updated responsibilities and complexity to include runtime management. |
| c3-301 | component | The component owns packages/cli/src, where release lookup, install, project metadata, run selection, uninstall, prune, and progress now live. | c3-301#n500@v1:sha256:70569c86bdff790c595cf02aa8a10440bc66d7c9f6db5e0869f260baa121e555 | Updated component goal, purpose, contracts, governance notes, and derived-material expectations. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-cross-compiled-binary | Runtime installation still consumes the supported release asset names and platform matrix. | ref-cross-compiled-binary#n730@v1:sha256:cea27fa9abdd975d6298f23e899ceb48f2945fde1e933915fdfada258a190136 "Ship c3x as named per-platform release binaries, with standard runtime-manager binaries," | Comply: preserve platform gating and release asset naming. |
| ref-fat-thin-distribution | The npm package remains the thin runtime-manager surface within the broader fat-skill, no-binary-skill, and thin-npm distribution split. | ref-fat-thin-distribution#n748@v1:sha256:cf8e08bcc48d161ef4120914a8505499068ac512709adeae244258fd1618b031 "Distribute C3 as full-fat skill ZIPs (semantic binaries bundled), Linux portable fat skill" | Review and comply: update notes from pinned download to runtime manager download. |
| ref-eval-determinism | Inherited from the top-level system; this work does not change eval verdict computation or eval fixtures. | ref-eval-determinism#n739@v1:sha256:d914f393b17de0202b7ae4cdde4df7d173c51fd820b2695487a29efb06f514d7 | N.A - runtime manager does not change eval computation. |
| ref-frontmatter-docs | This ADR updates frozen C3 package facts and must preserve frontmatter plus canvas-shaped markdown. | ref-frontmatter-docs#n757@v1:sha256:d4f7719668519e2f2a93de15969bc53c8f0105e7e073231a2f36d7c2626cb361 | Comply: facts move through a change-unit and finish with c3x check. |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| rule-dispatcher-error-hint | New npm manager errors are user-facing and should point to the next recovery move when actionable. | rule-dispatcher-error-hint#n625@v1:sha256:bd662000c1bc5b93d0b1cc4cf532cc1dc6e4766e5bda6b544f8aab14d21f7dc4 | Review and comply: runtime command errors use error: and targeted hint: strings where recovery is clear. |
| rule-output-via-helpers | Inherited from Go CLI components; this npm manager does not emit agent-mode TOON or JSON command results. | rule-output-via-helpers#n638@v1:sha256:b5ac8121ffc54be6c8f87ec133e69658fea023e7e73da3859fb85a33869afa29 | N.A - normal Go commands still own structured output; npm runtime commands are human package-management output. |
| rule-wrap-error-cause | Inherited from Go CLI components; the npm manager still needs contextual download and metadata failures. | rule-wrap-error-cause#n650@v1:sha256:b9e4edb84b11060973de3fe6e5c0ab7b5605aa690e00e886335b054bdaab710f | Review and comply: download errors preserve URL/status context and unsafe metadata errors name the rejected field. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Runtime commands | Add runtime versions, runtime installed, runtime install, runtime use, runtime uninstall, and runtime prune. | packages/cli/src/cli.ts, packages/cli/src/manager.ts |
| Runtime metadata | Read and write .c3/runtime.json as operational version metadata only. | readProjectRuntimeConfig, writeProjectRuntimeConfig tests |
| Cache lifecycle | Stop automatic old-version deletion during normal preparation; keep explicit uninstall and prune. | package tests for prepare, uninstall, and prune |
| Documentation | Update package README and npm CLI spec/plan to match current runtime-manager behavior. | packages/cli/README.md, docs/superpowers/specs/2026-03-17-c3x-npm-cli-design.md |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| npm manager tests | Add runtime acceptance tests for release listing, latest default, project selection, metadata rejection, install progress, uninstall protection, and prune. | cd packages/cli && npm test |
| frozen package facts | Patch c3-3 and c3-301 through this ADR so the C3 model matches the runtime-manager package. | c3x change apply adr-20260623-c3x-runtime-manager, then c3x check --include-adr |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| packages/cli/test/manager.test.mjs | Catches stale pinned-only behavior, unsafe metadata, and eager cache pruning. | Node test suite |
| runtime metadata validator | Rejects URLs, paths, malformed versions, and unknown fields before execution. | readProjectRuntimeConfig |
| checksum path | Downloads .sha256, verifies bytes, and only then renames into cache. | ensureCachedAsset |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Reuse top-level c3x list for runtime listing | It would collide with the Go CLI topology command and break established command meaning. |
| Keep package-pinned version as the default | It would keep the npm package as a stale facade and fail the latest-release objective. |
| Let project metadata set a release URL | It would create a repo-controlled path to unmanaged binary execution. |
| Continue automatic old-version garbage collection | It would make explicit multi-version install and project selection unreliable. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Project metadata becomes an execution authority | Store only a semver or latest; reject URLs, paths, and unknown fields. | package metadata rejection tests |
| Partial or tampered runtime becomes executable | Verify SHA256 and atomic rename after a temporary write. | checksum mismatch test |
| Runtime commands collide with Go commands | Keep the npm-only surface under runtime. | CLI smoke and package tests |
| Frozen C3 docs drift from the package | Land c3-3 and c3-301 fact patches in this ADR. | c3x check --include-adr |

## Verification

| Check | Result |
| --- | --- |
| cd packages/cli && npm test | Required before done. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh change view adr-20260623-c3x-runtime-manager | Required before applying C3 patches. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh change apply adr-20260623-c3x-runtime-manager | Required to update frozen facts. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr | Required before done. |
| independent reviewer pass on runtime security and C3 fact sync | Required by the no-single-LLM anti-goal. |
