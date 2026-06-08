---
id: ref-cross-compiled-binary
c3-version: 4
c3-seal: 869995bfe49f6096aced6a10581bf15c81741389ae89151e19f8565288aef704
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

REQUIRED pattern from `scripts/build.sh`: the default build variant is fat. Thin builds use default tags and never stage semantic model files in `cli/internal/store/semantic_model/`. Fat builds first run the repo helper `go run ./tools/semantic-assets --embed-dir cli/internal/store/semantic_model --os <target-os> --arch <target-arch>` to reuse a verified local cache or download the canonical pinned all-MiniLM-L6-v2 model/vocab plus target ONNX Runtime from `cli/internal/store/semantic_assets.go`, verify model/vocab SHA256, stage the real files for `go:embed`, pass `-tags embedmodel`, write `-fat` suffixed artifacts plus `.sha256` files under `dist/c3x/fat/`, then restore the tracked stub files with git before exiting.

REQUIRED release packaging pattern from `.github/workflows/distribute.yml`: tag releases upload thin binaries, model/vocab assets, their checksums, and `SHA256SUMS` for the npm manager. The model asset job uses the repo semantic asset helper instead of duplicating model URLs. The build job does not pre-copy model files into `cli/internal/store/semantic_model/`; `scripts/build.sh` owns fat staging and cleanup. Per-platform default skill zips are assembled from fat build outputs, rename the embedded-model binary to `skills/c3/bin/c3x-<version>-<os>-<arch>` inside the zip, and do not publish a thin skill zip.
