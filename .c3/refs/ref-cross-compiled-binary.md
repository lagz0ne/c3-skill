---
id: ref-cross-compiled-binary
c3-version: 4
c3-seal: f39d35dba48e2cd134924c88af62da72e79cd0023a481eb469ff2b443e08aa98
title: Cross-Compiled Binary Distribution
type: ref
goal: Standardize how C3 distributes platform-specific Go executables and semantic model assets without forcing every install channel to carry every large binary blob.
via: []
---

## Goal

Standardize how C3 distributes platform-specific Go executables and semantic model assets without forcing every install channel to carry every large binary blob.

## Choice

Publish two install paths with one default. The default C3 skill distribution is a per-platform fat skill zip built with `go build -tags embedmodel`; each zip contains one local `c3x-<version>-<os>-<arch>` binary and runs without downloading the Go binary or semantic model. The thin path is npm-only: `@c3x/cli` downloads the thin `c3x-<version>-<os>-<arch>` binary, semantic model, vocab, and SHA256 files from GitHub Releases into a versioned user cache before exec.

## Why

Phase C semantic search introduced an all-MiniLM-L6-v2 ONNX model that should not be duplicated inside the npm package, but the skill zip still needs to preserve the current self-contained "just works" install behavior. Making fat skill zips primary keeps agent/plugin installs offline-capable and removes launcher ambiguity. Keeping thin downloads inside `@c3x/cli` keeps npm small, version-pinned, checksum-verified, and explicit about its network/cache responsibility.

## How

REQUIRED pattern from `skills/c3/bin/c3x.sh`: the launcher reads `skills/c3/bin/VERSION`, maps `uname` to supported platform names, and execs `skills/c3/bin/c3x-<version>-<os>-<arch>` when that local binary is present. If no packaged binary exists and the checkout has `cli/go.mod` plus a Go toolchain, it builds the same local binary with `-tags embedmodel` and execs it for source-tree development. The launcher does not read `C3X_VARIANT` or `C3X_USE_FAT`, does not resolve a thin cache, and does not download release assets.

REQUIRED pattern from `packages/cli/src/manager.ts`: npm is thin-only. It resolves Node `process.platform` and `process.arch` to the release binary name `c3x-<version>-<os>-<arch>`, uses `~/.cache/c3x/<version>/` or `$XDG_CACHE_HOME/c3x/<version>/`, verifies SHA256 before chmod/exec, downloads the ONNX model and vocab as release assets, and removes old version directories after preparing the pinned version.

REQUIRED pattern from `scripts/build.sh`: the default build variant is fat. Thin builds use default tags for npm release assets; fat builds pass `-tags embedmodel` and write `-fat` suffixed artifacts plus `.sha256` files under `dist/c3x/<variant>/`.

REQUIRED release packaging pattern from `.github/workflows/distribute.yml`: tag releases upload thin binaries, model/vocab assets, their checksums, and `SHA256SUMS` for the npm manager. Per-platform default skill zips are assembled from fat build outputs, rename the embedded-model binary to `skills/c3/bin/c3x-<version>-<os>-<arch>` inside the zip, and do not publish a thin skill zip.
