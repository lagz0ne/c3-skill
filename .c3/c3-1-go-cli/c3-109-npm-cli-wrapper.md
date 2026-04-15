---
id: c3-109
c3-seal: 287a90ed0e71aeac7c020c292173d844ed48153f1fb470848f56779d705c1063
title: npm-cli-wrapper
type: component
category: foundation
parent: c3-1
goal: Provide the npm `@c3x/cli` shim that discovers an installed C3 binary and delegates commands without changing the caller's intended output mode.
---

## Goal

Provide the npm `@c3x/cli` shim that discovers an installed C3 binary and delegates commands without changing the caller's intended output mode.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Extends the Go CLI distribution surface with a Node-based wrapper while leaving .c3/ read/write behavior inside the selected c3x binary. |
| Boundary | Owns binary discovery, npm package metadata, and delegation environment for packages/cli; does not parse or mutate C3 architecture docs. |
| Collaboration | Coordinates with runtime output policy in c3-108 and release/version policy in c3-1 when wrapper behavior or publication changes. |
## Purpose

Own the thin npm wrapper used by humans and scripts that want `npx @c3x/cli` or a global npm command. The component discovers candidate skill/plugin installs, chooses the highest semver with deterministic priority tie-breaks, and invokes that install's `c3x.sh`. It must not silently force agent output; agent mode belongs to the skill wrapper or explicit CLI flags.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | At least one candidate install exposes skills/c3/bin/VERSION and skills/c3/bin/c3x.sh; otherwise the wrapper prints searched locations and exits non-zero. | c3-1 |
| Inputs | Accepts npm CLI argv, wrapper flag --agent claude or --agent codex, current working directory, home directory, and installed skill/plugin version files. | c3-1 |
| State / data | Reads version files only for discovery; package metadata and lockfile define the npm publication contract. | c3-1 |
| Shared dependencies | Uses Node standard library for process execution and filesystem discovery; build output comes from tsdown. | c3-108 |
## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Human shell, npm script, or automation invokes c3x through npx @c3x/cli or a globally installed npm binary. | c3-1 |
| Primary path | Parse wrapper-only flags, discover candidates from project, Claude, Codex, and marketplace paths, select highest semver, then delegate remaining args to c3x.sh. | c3-1 |
| Alternate paths | --agent claude limits non-project candidates to Claude locations; --agent codex limits non-project candidates to Codex locations. | c3-1 |
| Failure behavior | If no candidate exists or the child command fails, print actionable install/search information or propagate the child exit status. | c3-108 |
## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-108 | policy | Output-mode ownership and human/agent presentation boundaries. | Runtime output policy beats wrapper convenience. | The npm shim strips inherited C3X_MODE; explicit c3x flags still pass through. |
| adr-20260415-npm-cli-human-mode | adr | Decision to chart the npm shim and keep npm delegation human/default by default. | Release-specific decision for this wrapper change. | Added with the 9.1.0 release bump. |
| c3-1 | policy | Packaging and release placement inside the CLI container. | Parent container scope beats local package convenience. | README and package metadata must match wrapper behavior before publish. |
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
