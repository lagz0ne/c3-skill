---
id: c3-109
c3-seal: c338a60df165c278942c16ca3dfafa53a9134b15927fe1ee7f9370613a14f3ec
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
| Actor / caller | Human or script invokes c3x through npm, npx, or a globally installed package. | c3-1 |
| Primary path | Resolve platform and pinned version, ensure the thin binary and semantic assets exist in the versioned cache with valid SHA256 checksums, prune older versions, then exec the cached binary with original argv and cwd. | ref-cross-compiled-binary |
| Failure path | Unsupported platforms, failed downloads, missing checksums, integrity mismatches, and exec failures exit non-zero with a clear hint to retry with network/cache repair or use the fat C3 build. | rule-dispatcher-error-hint |
| Output mode | The manager does not force C3X_MODE; human/default output remains the Go binary default and agent output remains controlled by skill launchers or explicit caller environment. | rule-output-via-helpers |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-108 | policy | Output-mode ownership and human/agent presentation boundaries. | Runtime output policy beats wrapper convenience. | The npm shim strips inherited C3X_MODE; explicit c3x flags still pass through. |
| adr-20260415-npm-cli-human-mode | adr | Decision to chart the npm shim and keep npm delegation human/default by default. | Release-specific decision for this wrapper change. | Added with the 9.1.0 release bump. |
| c3-1 | policy | Packaging and release placement inside the CLI container. | Parent container scope beats local package convenience. | README and package metadata must match wrapper behavior before publish. |
| ref-cross-compiled-binary | ref | The npm wrapper depends on the prebuilt-binary distribution: it resolves skills/c3/bin/VERSION and c3x.sh produced by cross-compilation, with no Go toolchain at runtime. | Cited ref contract beats uncited local prose. | Wrapper binary discovery must track the cross-compiled artifact layout per ref-cross-compiled-binary. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| npm argv | IN | Accept wrapper flag --agent and forward all remaining args unchanged to the selected c3x.sh. | c3-1 boundary | npm run build; temp shim smoke via node packages/cli/dist/cli.mjs list. |
| install discovery | OUT | Select the highest semver candidate across project, Claude, Codex, and marketplace paths, with priority as tie-breaker. | c3-1 boundary | packages/cli/src/cli.ts discovery code plus packages/cli/README.md resolution table. |
| child environment | OUT | Remove inherited C3X_MODE before delegation so npm callers receive human/default output unless they pass explicit c3x output flags. | c3-108 boundary | C3X_MODE=agent node packages/cli/dist/cli.mjs list temp shim smoke prints unset. |
| npm publication | OUT | Package version changes when wrapper behavior changes so the Publish @c3x/cli workflow can publish. | c3-1 boundary | packages/cli/package.json and packages/cli/package-lock.json at 0.1.2. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Agent-mode leak | Wrapper inherits C3X_MODE=agent from a parent process and forces agent output for npm users. | Temp fake install prints child environment while parent exports C3X_MODE=agent. | Run npm run build and C3X_MODE=agent node packages/cli/dist/cli.mjs list from a temp project; expect unset. |
| Discovery regression | Changes to candidate paths, semver sorting, or --agent filtering select the wrong install. | Compare packages/cli/src/cli.ts discovery logic to packages/cli/README.md resolution order. | Run npm run build and a temp project-scope shim smoke. |
| Publish skip | Code changes without npm package version bump leave npm workflow no-op. | Compare packages/cli/package.json against .github/workflows/npm-publish.yml version check behavior. | Run jq empty packages/cli/package.json packages/cli/package-lock.json; verify version 0.1.2. |
| C3 ownership drift | Wrapper files change without mapped component ownership. | c3x lookup packages/cli/src/cli.ts should resolve to this component. | Run c3x lookup packages/cli/src/cli.ts, c3x check --include-adr, and c3x verify. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| packages/cli/src/cli.ts | Purpose, Business Flow, Contract child environment row. | Implementation details may vary; inherited C3X_MODE must not reach c3x.sh. | npm run build; C3X_MODE=agent node packages/cli/dist/cli.mjs list temp shim smoke. |
| packages/cli/package.json and package-lock.json | Contract npm publication row and Change Safety publish-skip row. | Patch version may advance; package name and bin contract stay stable. | jq empty packages/cli/package.json packages/cli/package-lock.json. |
| packages/cli/README.md | Business Flow actor row and Governance notes. | Copy can be concise; must not claim npm sets agent mode automatically. | rg C3X_MODE packages/cli/README.md has no stale automatic-agent claim. |
