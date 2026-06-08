---
id: ref-cross-compiled-binary
c3-version: 4
c3-seal: 21fe68957486f62b79096a4cf537bade92ffa81b29f5bf286dcc8fc29e1c85b0
title: Cross-Compiled Binary Distribution
type: ref
goal: Standardize how C3 distributes platform-specific Go executables and semantic model assets without forcing every install channel to carry every large binary blob.
via: []
---

## Goal

Standardize how C3 distributes platform-specific Go executables and semantic model assets without forcing every install channel to carry every large binary blob.

## Choice

Publish two distribution variants. Thin is the default: release assets provide `c3x-<version>-<os>-<arch>` binaries plus semantic model/vocab assets, and launchers cache them under the versioned C3X cache. Fat is explicit: `go build -tags embedmodel` produces `c3x-<version>-<os>-<arch>-fat` with the ONNX model embedded for offline plugin artifacts.

## Why

Phase C semantic search introduced an all-MiniLM-L6-v2 ONNX model that is too large to duplicate into npm and every skill install. GitHub Releases are already the binary distribution boundary, while npm should remain a small manager package and the source skill should stop committing all platform binaries. The split preserves air-gapped installs through fat artifacts while keeping normal installs small and version-pinned.

## How

REQUIRED pattern from `skills/c3/bin/c3x.sh`: the launcher reads `skills/c3/bin/VERSION`, maps `uname` to `linux|darwin` and `amd64|arm64`, then selects by env. `C3X_VARIANT=fat` or `C3X_USE_FAT=1` requires a local `c3x-<version>-<os>-<arch>-fat`; thin mode resolves `$C3X_CACHE_DIR`, `$XDG_CACHE_HOME/c3x`, or `$HOME/.cache/c3x`, downloads `c3x-<version>-<os>-<arch>` from the GitHub Release when absent, verifies `<asset>.sha256`, exports `C3X_VERSION` and `C3_SEMANTIC_CACHE_DIR`, and execs the cached binary.

REQUIRED pattern from `packages/cli/src/manager.ts`: npm resolves Node `process.platform`/`process.arch` to the same asset names, uses `~/.cache/c3x/<version>/`, verifies SHA256 before chmod/exec, downloads the ONNX model and vocab as release assets rather than npm files, and removes old version directories after preparing the pinned version.

REQUIRED pattern from `scripts/build.sh`: thin builds use default tags; fat builds pass `-tags embedmodel` and write `-fat` suffixed artifacts plus `.sha256` files under `dist/c3x/<variant>/`.

OPTIONAL release packaging pattern from `.github/workflows/distribute.yml`: tag releases upload thin binaries, model/vocab assets, `SHA256SUMS`, a thin skill zip without committed Go binaries, and per-platform fat skill zips containing only the matching embedded-model binary.
