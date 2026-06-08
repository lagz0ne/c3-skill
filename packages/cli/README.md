# @c3x/cli

Thin npm manager for `c3x`, the C3 architecture documentation CLI.

The package does not bundle Go binaries or the ONNX model. On invocation it resolves the pinned C3 version, downloads release assets from GitHub Releases, verifies SHA256 checksums, caches them, and execs the cached Go binary.

## Install

```bash
npm install -g @c3x/cli
c3x list
```

`npx @c3x/cli list` works the same way.

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

Old version directories are removed after the pinned version is prepared.

## Environment

| Variable | Use |
| --- | --- |
| `C3X_VERSION` | Override the pinned Go binary version. |
| `C3X_RELEASE_BASE_URL` | Override the GitHub Release asset base URL. |
| `XDG_CACHE_HOME` | Select the cache root. |
| `C3X_SKIP_MODEL_DOWNLOAD` | Download only the Go binary; the Go CLI can fetch semantic assets later. |

If the first download happens offline, use the fat C3 skill artifact or prefill the cache from a connected machine.

## License

MIT
