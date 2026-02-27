---
id: ref-cross-compiled-binary
c3-version: 4
title: Cross-Compiled Binary Distribution
goal: CLI is distributed as pre-built binaries for 4 targets so users need no Go toolchain
via: []
---

# Cross-Compiled Binary Distribution

## Goal

CLI is distributed as pre-built binaries for 4 targets so users need no Go toolchain to use c3x.

## Choice

Cross-compile for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`. Bundle in the release zip. Use `c3x.sh` wrapper to select the correct binary at runtime.

## Why

- **Zero runtime deps**: Users need only bash + the plugin zip — no Go, no npm, no Python
- **Fast startup**: Native binary, no interpreter overhead
- **Plugin distribution**: Claude Code plugins are zip files; bundled binaries make installs self-contained

## How

```bash
bash scripts/build.sh   # cross-compiles all 4 targets → skills/c3/bin/c3x-{ver}-{os}-{arch}
```

CI (`distribute.yml`) builds on push to `dev`, force-commits binaries to `main`, then packages the release zip from `main`.

## Stale Binary Cleanup

`c3x.sh` removes old versioned binaries after finding the current one — prevents accumulation of old binaries in the installed cache.

## Not This

Do not auto-download binaries at runtime — plugin installs must be self-contained. Do not commit binaries to `dev` (gitignored; only `main` tracks them via force-add).
