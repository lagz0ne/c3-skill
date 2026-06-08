---
id: c3-109
c3-seal: ba1c5f14632a8dc0a6ac78d46e02c271c69048d3ba78ed1dbfd92b688fd33b2c
title: npm-cli-wrapper
type: component
category: foundation
parent: c3-1
goal: Provide the npm `@c3x/cli` manager that downloads, verifies, caches, and execs the pinned thin C3 release binary without changing caller output mode.
uses:
    - ref-cross-compiled-binary
---

## Goal

Provide the npm `@c3x/cli` manager that downloads, verifies, caches, and execs the pinned thin C3 release binary without changing caller output mode.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Extends the Go CLI distribution surface with a Node-based wrapper while leaving .c3/ read/write behavior inside the selected c3x binary. |
| Boundary | Owns binary discovery, npm package metadata, and delegation environment for packages/cli; does not parse or mutate C3 architecture docs. |
| Collaboration | Coordinates with runtime output policy in c3-108 and release/version policy in c3-1 when wrapper behavior or publication changes. |

## Purpose

Own the thin npm manager used by humans and scripts that want `npx @c3x/cli` or a global npm command. The component resolves the supported OS/arch, pinned C3 version, GitHub Release asset names, versioned user cache, SHA256 checksums, semantic model/vocab prefetch, old-version garbage collection, and final exec into the cached Go binary. It does not parse or mutate C3 architecture docs; all document behavior remains inside the selected Go binary.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | The npm package carries only JavaScript manager code and a pinned C3 binary version; release assets must exist for the requested platform/version. | ref-cross-compiled-binary |
| Inputs | Accepts npm CLI argv, process platform/arch, current working directory, HOME, XDG_CACHE_HOME, C3X_VERSION, C3X_RELEASE_BASE_URL, and C3X_SKIP_MODEL_DOWNLOAD. | c3-1 |
| State / data | Reads and writes only the versioned cache under $XDG_CACHE_HOME/c3x/<version> or ~/.cache/c3x/<version>, including binary, semantic model, vocab, and old-version pruning. | ref-cross-compiled-binary |
| Shared dependencies | Uses Node core modules for download, hashing, filesystem writes, and process exec; delegates C3 document behavior to the cached Go binary. | c3-108 |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Human or script invokes c3x through npm, npx, or a globally installed @c3x/cli package. | c3-1 |
| Primary path | Resolve platform and pinned version, ensure the thin binary and semantic assets exist in the versioned cache with valid SHA256 checksums, prune older versions, then exec the cached binary with original argv and cwd. | ref-cross-compiled-binary |
| Failure path | Unsupported platforms, failed downloads, missing checksums, integrity mismatches, and exec failures exit non-zero with a clear hint to retry, repair the cache, or prefill the npm cache from a connected machine. | rule-dispatcher-error-hint |
| Output mode | The manager does not force C3X_MODE or serialize C3 command output; human, JSON, and agent output remain owned by the selected Go binary and caller environment. | rule-output-via-helpers |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-cross-compiled-binary | ref | Thin npm manager asset naming, cache layout, checksum verification, semantic asset downloads, and old-version pruning. | Distribution ref beats local package prose. | npm stays thin-only; fat skill zips are outside this manager. |
| c3-1 | policy | Packaging and release placement inside the CLI container. | Parent container scope beats local package convenience. | README, package metadata, tests, and release workflow must match wrapper behavior before publish. |
| c3-108 | policy | Runtime output ownership and human/agent presentation boundaries. | Go binary output policy beats wrapper convenience. | The npm manager execs the Go binary and does not serialize C3 command results itself. |
| adr-20260608-fat-default-thin-npm | adr | Decision to make fat skill zips default and npm the thin-only opt-in manager. | Current branch work order beats older fat/thin split ADR wording. | Use this ADR for final verification evidence. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| npm argv | IN | Forward all CLI arguments unchanged to the cached thin Go binary. | c3-1 boundary | cd packages/cli && npm test; cd packages/cli && npm run build. |
| release assets | IN | Resolve only thin asset names: c3x-<version>-<os>-<arch>, semantic model, semantic vocab, and matching .sha256 files. | ref-cross-compiled-binary boundary | packages/cli/src/manager.ts assetNames; npm tests with stub downloader. |
| versioned cache | OUT | Store assets under $XDG_CACHE_HOME/c3x/<version>/ or ~/.cache/c3x/<version>/, chmod the binary after checksum verification, and prune older version directories. | ref-cross-compiled-binary boundary | packages/cli/test/manager.test.mjs prepareRuntime and gcOldVersions tests. |
| child environment | OUT | Set C3X_VERSION and C3_SEMANTIC_CACHE_DIR for the child process while leaving C3 output mode to the caller and Go binary. | c3-108 boundary | packages/cli/src/manager.ts runCli; npm build/test. |
| npm publication | OUT | Package ships manager JavaScript only; no Go binaries, ONNX model, vocab, or fat variant files are included in npm. | c3-1 boundary | packages/cli/package.json files and README cache section. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Wrong thin asset name | Platform mapping or version formatting changes select a non-release asset. | Stub downloader test fails on missing asset name. | cd packages/cli && npm test covers prepareRuntime asset names. |
| Checksum bypass | Cached or downloaded asset is used without matching the release SHA256. | ensureCachedAsset mismatch test rejects corrupted data. | cd packages/cli && npm test covers checksum mismatch rejection. |
| Cache pollution | Preparing a new pinned version leaves old binary/model trees in place. | gcOldVersions test checks old directory removal. | cd packages/cli && npm test covers old-version pruning. |
| Fat or variant drift returns to npm | New env flags or fat-suffixed asset names appear under packages/cli. | Text search finds forbidden fat or variant handling in npm package. | rg -n "fat" packages/cli; rg -n "variant" packages/cli; both return no matches. |
| Package build drift | TypeScript manager exports no longer match tests or bin entry. | tsdown or Node test suite fails. | cd packages/cli && npm run build; cd packages/cli && npm test. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| packages/cli/src/manager.ts | Purpose, Foundational Flow, Business Flow primary path, and Contract release/cache rows. | Helper names may vary; behavior must stay thin-only with SHA256-verified release asset cache and exec. | cd packages/cli && npm test; cd packages/cli && npm run build. |
| packages/cli/test/manager.test.mjs | Change Safety wrong asset, checksum, and cache-pruning risks. | Test data may use stub assets; no network is required. | cd packages/cli && npm test. |
| packages/cli/README.md | Business Flow and Contract npm publication rows. | Copy can stay concise; must not promise bundled binaries or model files. | rg -n "c3x-<version>-<os>-<arch>" packages/cli/README.md. |
| packages/cli/package.json and package-lock.json | Contract npm publication row. | Version may advance; files must keep npm scoped to dist/. | cd packages/cli && npm run build. |
