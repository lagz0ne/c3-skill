---
target: c3-301
scope: block
base: c3-301#n509@v1:sha256:7783ee34eec9aca2e29aacc11d23ac74b6ebb8bf32ac716655785ed4d1cd02a2
---
Carry packages/cli/src: cli.ts routes the npm-only `runtime` namespace to the manager and forwards every other argv to runCli; manager.ts resolves supported platforms, lists non-draft GitHub releases, validates `.c3/runtime.json` as version-only operational metadata, defaults to the latest release with installed-cache fallback, installs release assets with progress, verifies `.sha256` before atomic activation, writes runtime manifests, refuses unsafe project metadata, protects project-selected runtimes from accidental uninstall, prunes cache entries explicitly, and execs the cached binary with C3X_VERSION and the semantic-cache dir exported; version.ts holds the package version pin and semantic model revision for release/test surfaces. Non-goals: implementing any C3 command (the binary), teaching operations (the skill), or building the binary (CI).
