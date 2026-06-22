---
target: c3-301
scope: whole
type: component
parent: c3-3
title: binary-downloader
---
# binary-downloader

## Goal

Resolve the pinned platform binary, download and checksum-verify it into a per-version cache, then exec it with the user's arguments forwarded.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-3 |
| Role | The whole of the npm thin client: the install-time downloader and runtime launcher behind the `c3x` bin. |
| Boundary | Owns platform resolution, version pinning, asset download, checksum verification, caching, and the exec; ships none of the architecture logic it launches. |
| Collaboration | Fetches the same per-platform binary cli-wrapper selects locally; pins the version the wrapper's VERSION also tracks; downloads from the GitHub Release the build matrix publishes. |

## Purpose

Carry packages/cli/src: cli.ts forwards argv into runCli and surfaces errors; manager.ts resolves the platform (x64 to amd64, gates linux x64/arm64 and darwin arm64), pins the version from C3X_VERSION or version.ts, computes the cache dir under XDG_CACHE_HOME/HOME, garbage-collects other versions, downloads each asset and its .sha256 (following GitHub's redirects), verifies the digest before an atomic rename, and execs the cached binary with C3X_VERSION and the semantic-cache dir exported; version.ts holds the pinned C3X_VERSION and model revision. Non-goals: implementing any C3 command (the binary), teaching operations (the skill), or building the binary (CI).

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-cross-compiled-binary | ref | The per-platform asset name and supported OS/arch matrix the client resolves and downloads | Client downloads only an asset the build matrix produces | manager.ts resolvePlatform gates linux x64/arm64 + darwin arm64 and assetNames builds `c3x-${version}-${os}-${arch}`. |
| ref-fat-thin-distribution | ref | Why the npm package ships only a downloader and pulls the pinned binary on demand instead of bundling it | Thin client fetches; the binary stays the single source | manager.ts downloads from the GitHub Release base URL at the pinned version, the thin-client half of the split. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| c3x CLI entry | IN | Invoked with argv; prepares the runtime and execs the cached binary, propagating its exit status | Forwards argv unchanged; no flags consumed by the wrapper | packages/cli/src/cli.ts; manager.ts runCli / defaultExec |
| asset fetch + cache | OUT | Each asset is downloaded, its sha256 matched against the published checksum, then atomically renamed into the per-version cache; a mismatch aborts | A checksum-failing or partially written asset never becomes the live binary | packages/cli/src/manager.ts ensureCachedAsset; checksum-mismatch throw |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| packages/cli/src/{cli,manager}.ts | Contract | Caching layout, redirect handling, and exec mechanics may vary while pinned-platform fetch + verify hold | npm run test (build + node --test) |
| packages/cli/src/version.ts | Goal | The pinned version and model revision change each release | C3X_VERSION matches skills/c3/bin/VERSION and the release tag |
