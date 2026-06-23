# @c3x/cli

Thin npm manager for `c3x`, the C3 architecture documentation CLI.

The package does not bundle Go binaries or the ONNX model. For real C3 commands it resolves the project-selected C3 runtime, or the latest GitHub Release when the project has no selection, downloads release assets from GitHub Releases with visible progress, verifies SHA256 checksums, caches them, and execs the cached Go binary.

This is the thin npm distribution only. The platform-neutral skill/plugin artifact carries no binary and delegates through this package, while per-platform C3 skill ZIPs remain fat and bundle a binary so they can run in sandboxed or offline environments without depending on the npm cache. Full fat skill ZIPs include the semantic ONNX runtime; Linux portable fat ZIPs use a pure-Go core binary for broader distro compatibility and run without local semantic ONNX.

Root discovery commands stay local. `c3x`, `c3x --help`, and `c3x --version` do not resolve releases or download runtime assets.

## Install

```bash
npm install -g @c3x/cli
c3x list
```

`npx @c3x/cli list` works the same way.

## Runtime Manager

The npm entrypoint owns runtime installation and cache management under the `runtime` namespace:

```bash
c3x runtime versions
c3x runtime installed
c3x runtime install latest
c3x runtime install 11.3.0
c3x runtime use 11.3.0
c3x runtime uninstall 11.3.0
c3x runtime prune
```

`c3x runtime use <version>` writes operational project metadata at `.c3/runtime.json`. That file stores only the selected runtime version; it is not a frozen C3 fact and it never stores a binary URL or executable path.

`c3x runtime install <version|latest>` prints download and verification progress on stderr. Normal C3 commands also print progress only when they must populate missing runtime assets before execution.

## Cache

Assets are cached under:

```text
$XDG_CACHE_HOME/c3x/<version>/
~/.cache/c3x/<version>/
```

The manager downloads:

- `c3x-<version>-<os>-<arch>`
- `c3x-semantic-model-all-MiniLM-L6-v2-<revision>.onnx`
- `c3x-semantic-vocab-all-MiniLM-L6-v2-<revision>.txt`
- matching `.sha256` files

Old version directories are kept until explicitly removed with `c3x runtime uninstall <version>` or `c3x runtime prune`.

## Environment

| Variable | Use |
| --- | --- |
| `C3X_VERSION` | Override the selected Go binary version. Intended for development and tests. |
| `XDG_CACHE_HOME` | Select the cache root. |
| `C3X_SKIP_MODEL_DOWNLOAD` | Download only the Go binary; the Go CLI can fetch semantic assets later. |

If the first download happens offline, prefill the cache from a connected machine.

## License

MIT
