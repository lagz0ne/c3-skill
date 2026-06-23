# @c3x/cli — npm runtime manager for c3x

## Problem

The npm `c3x` entrypoint used to be a minimal facade around binaries found inside installed skills.
That made the npm package depend on a separate skill install. The distribution split now has four
artifact shapes: per-platform full-fat skill ZIPs bundle the binary plus semantic ONNX runtime for
sandboxed or offline use; Linux portable fat skill ZIPs bundle a pure-Go core binary for broader
distro/sandbox compatibility; a platform-neutral skill/plugin ZIP carries no binary and delegates
through the pinned npm runtime manager; and the npm package stays thin and manages standalone GitHub
Release runtime assets directly. It must list available versions, install and cache them, select the
runtime for a project, and remove cache entries explicitly.

## Design

### Package

- **Name**: `@c3x/cli`
- **Location**: `packages/cli/`
- **Source**: `src/cli.ts`, `src/manager.ts`, `src/version.ts`
- **Built with**: tsdown -> `dist/*.mjs`
- **Bin**: `{ "c3x": "dist/cli.mjs" }`
- **Runtime dependencies**: Node.js standard library only

### Command Surface

Normal C3 commands pass through unchanged:

```bash
c3x
c3x --help
c3x --version
c3x list
c3x check
c3x search "auth"
```

Root discovery commands (`c3x`, `c3x --help`, `c3x --version`) are served by the npm manager and must not resolve releases or download runtime assets.

Skill wrapper discovery commands use the first local bundled binary available. A full-fat skill ZIP
provides `c3x-<version>-<os>-<arch>`; a Linux portable fat ZIP provides
`c3x-<version>-linux-<arch>-portable`. Only the no-binary skill/plugin artifact reaches the npm
manager fallback.

Runtime-manager commands are namespaced so they do not collide with Go CLI commands:

```bash
c3x runtime versions
c3x runtime installed
c3x runtime install latest
c3x runtime install 11.3.0
c3x runtime use 11.3.0
c3x runtime uninstall 11.3.0
c3x runtime prune
```

### Runtime Resolution

On normal command execution, the npm manager resolves a runtime version, prepares it in cache, then
execs the cached Go binary with the original argv.

Resolution order:

1. `C3X_VERSION`, for development and tests.
2. Project runtime metadata from `.c3/runtime.json`, when present.
3. Latest non-draft, non-prerelease GitHub Release.
4. Newest installed cache entry as an offline fallback when the release list cannot be read.

Project metadata is operational data, not a frozen C3 fact. It may store only:

```json
{
  "version": "11.3.0"
}
```

It must never store a binary URL, release base URL, executable path, shell command, timestamp,
checksum, or arbitrary source. Malformed versions, paths, URLs, and unknown metadata fields are
rejected before execution.

### Asset Fetch And Cache

Assets are cached under:

```text
$XDG_CACHE_HOME/c3x/<version>/
~/.cache/c3x/<version>/
```

The manager downloads and verifies:

- `c3x-<version>-<os>-<arch>`
- `c3x-semantic-model-all-MiniLM-L6-v2-<revision>.onnx`
- `c3x-semantic-vocab-all-MiniLM-L6-v2-<revision>.txt`
- matching `.sha256` files

Each asset is written to a temporary path, checksum-verified, and atomically renamed into place. A
checksum mismatch or failed download never activates a partial binary.

Installed versions are not garbage-collected during normal execution. Cache cleanup is explicit:
`runtime uninstall` removes a named version, and `runtime prune` keeps the project-selected runtime
or the newest installed runtime.

### Progress

`c3x runtime install` reports progress on stderr as stable, human-readable steps:

- asset preparation
- download byte progress with a text progress bar when the server provides length
- checksum verification or cache hit
- final ready path

Normal passthrough commands stay quiet unless the manager must fetch missing runtime assets before execution; in that case they print throttled stderr progress. Root help/version paths stay entirely local and never trigger runtime resolution.

Portable fat skill ZIPs are outside the npm cache path: they are selected by the shell wrapper before
the npm manager is invoked. They must not trigger npm package installation or runtime downloads while
their bundled binary is present.

## Security Rules

- Project metadata may select only a version or `latest`.
- Project metadata cannot configure release hosts, URLs, paths, commands, or checksums.
- Runtime assets must match published `.sha256` digests before activation.
- Hidden release-source overrides are for tests/development only and remain undocumented.
- The package does not implement any C3 architecture behavior; it only resolves, verifies, caches,
  and execs the Go binary.

## Verification

Required before release:

```bash
cd packages/cli && npm test
C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr
```

When Go CLI behavior changes too:

```bash
cd cli && go test ./...
```
