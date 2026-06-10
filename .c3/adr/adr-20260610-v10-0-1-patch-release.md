---
id: adr-20260610-v10-0-1-patch-release
c3-seal: 3859b6e6e055736d22c6304e585c7f22d24968b090c7a4910458338d14fc8338
title: v10-0-1-patch-release
type: adr
goal: Release the verified v10 fixes as C3 v10.0.1 so the GitHub Release assets, fat skill metadata, and thin npm wrapper all pin a post-fix binary instead of the already-published v10.0.0 assets.
status: implemented
date: "2026-06-10"
---

## Goal

Release the verified v10 fixes as C3 v10.0.1 so the GitHub Release assets, fat skill metadata, and thin npm wrapper all pin a post-fix binary instead of the already-published v10.0.0 assets.

## Context

The repository already has remote tag `v10.0.0` and a published GitHub Release from commit `7bf755a`, while the current branch contains follow-up fixes for hyphenated search and agent TOON output. npm has not published any 10.x `@c3x/cli` package, but publishing npm `10.0.0` now would make the wrapper download the old `v10.0.0` GitHub asset. Release flow is CI-owned: tag pushes build thin binaries, semantic assets, and fat skill zips through `.github/workflows/distribute.yml`, while pushes to `dev` publish the npm wrapper when `packages/cli/package.json` has a new version.

## Decision

Bump the release line to `10.0.1` and keep every version-bearing release surface in sync: `skills/c3/bin/VERSION`, `.claude-plugin` manifests, `packages/cli` package metadata, and the npm wrapper pin in `packages/cli/src/version.ts`. Do not force-move or overwrite `v10.0.0`; create a normal patch release tag after the fixes are committed. Add codemap coverage for `.claude-plugin` manifests under the distribution ref so plugin release metadata is tracked by C3 going forward. Align launcher platform checks with the v10 release matrix by supporting Linux amd64/arm64 and Darwin arm64, and by rejecting Darwin amd64 before a missing asset is requested.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-109 | component | The npm wrapper pins the GitHub Release binary version, package version, and supported platform normalization. | c3-109#n2337@v1:sha256:0c07b1d57423025038e57010f6e16e5eadfc2854aeed40cd14aca1176c643e96 "Boundary \| Owns binary discovery, npm package metadata, and delegation environment for packages/cli; does not parse or mutate C3 architecture docs." | Keep npm package version, wrapper pin, and supported platform checks aligned with release assets. |
| N.A - ref-governed release metadata | N.A - release metadata | skills/c3/bin/VERSION, skills/c3/bin/c3x.sh, and .claude-plugin manifests are distribution metadata governed by ref-cross-compiled-binary, not a command component. | ref-cross-compiled-binary#n3377@v1:sha256:90bdc1dce31722c7d28535fa7aaf8b4ae1ea7ebb6955aaa1c7a44d15305b9e1c "Standardize how C3 distributes platform-specific Go executables and semantic model assets without forcing every install channel to carry every large binary blob" | Follow the distribution ref and codemap the manifest files to that ref. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-cross-compiled-binary | Governs thin npm asset names, fat skill zips, skills/c3/bin/VERSION, scripts/build.sh, release workflows, and launcher distribution behavior. | ref-cross-compiled-binary#n3377@v1:sha256:90bdc1dce31722c7d28535fa7aaf8b4ae1ea7ebb6955aaa1c7a44d15305b9e1c "Standardize how C3 distributes platform-specific Go executables and semantic model assets without forcing every install channel to carry every large binary blob" | comply |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| N.A - no directly applicable release rule | N.A - this patch ADR governs release metadata, npm platform support, and distribution pinning; output and command-error rules are enforced by the existing behavioral ADRs in this release. | N.A - no release metadata rule applies. | N.A - no rule action. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Skill release metadata | Change skills/c3/bin/VERSION, .claude-plugin/plugin.json, and .claude-plugin/marketplace.json from 10.0.0 to 10.0.1. | GitHub Release v10.0.0 already exists, so patch release metadata must advance. |
| npm wrapper metadata | Change packages/cli/package.json, packages/cli/package-lock.json, and packages/cli/src/version.ts to 10.0.1. | npm view @c3x/cli versions --json has no 10.x; wrapper must point to the new GitHub Release. |
| Platform support alignment | Reject Darwin x64 in packages/cli/src/manager.ts and skills/c3/bin/c3x.sh because .github/workflows/distribute.yml only builds Linux amd64/arm64 and Darwin arm64 release assets. | Pre-release audit verified c3x-10.0.0-darwin-amd64 is absent from GitHub Release assets while the launcher paths accepted it. |
| Changelog | Add a 10.0.1 entry summarizing the hyphen search fix, agent TOON default, marketplace show JSON sunset, release metadata correction, and platform support alignment. | release.yml extracts changelog text for new tags when present. |
| C3 codemap | Map .claude-plugin/plugin.json and .claude-plugin/marketplace.json to ref-cross-compiled-binary. | c3x lookup .claude-plugin/** reported uncharted files. |
| Git release action | Commit changes, push the branch/target branch, then create/push tag v10.0.1 only after verification passes. | .github/workflows/distribute.yml builds assets on v* tag pushes. |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| Codemap coverage | Add .claude-plugin release manifest paths to .c3/code-map.yaml under ref-cross-compiled-binary. | c3x lookup '.claude-plugin/**' should resolve manifests after the change. |
| CLI tests/help/validators | N.A - the behavioral CLI changes were already covered by adr-20260610-search-hyphen-cli-v10 and adr-20260610-force-toon-default. | Run their existing verification commands before release. |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| .github/workflows/distribute.yml | On tag v10.0.1, builds thin binaries, model/vocab assets, fat skill zips, checksums, and uploads GitHub Release assets. | Workflow read showed tag-triggered Build & Distribute. |
| .github/workflows/npm-publish.yml | On dev push touching packages/cli/**, publishes npm only when the package version differs from npm. | Workflow read showed version check against npm and npm publish --access public. |
| scripts/build.sh | Builds local thin/fat binaries with -X main.version=${VERSION} and default version from skills/c3/bin/VERSION. | Script read showed --version handling and VERSION fallback. |
| Package tests | npm test builds and tests the wrapper before CI publish. | packages/cli/package.json defines test as build plus node tests. |
| C3 checks | c3x check validates codemap and architecture docs after release metadata changes. | Project workflow requires local C3 validation before done. |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Force-move or clobber v10.0.0 | The tag and GitHub Release are already public with assets; mutating it would make checksums/downloads ambiguous and break normal release provenance. |
| Publish npm 10.0.0 after the fixes | The wrapper would pin C3X_VERSION = 10.0.0 and download the already-published old binary assets, leaving the fixed code unreleased for npm users. |
| Leave plugin manifests at 10.0.0 and only bump npm | Fat skill zips and marketplace metadata would advertise the wrong version while npm points to a different release line. |
| Create 10.1.0 | The changes are patch fixes and metadata correction after a major release, not a new minor feature line. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Version surfaces drift | Update all 10.0.0 active release metadata in one patch and search for remaining active 10.0.0 references. | rg -n "10\\.0\\.0" over active release surfaces after edits. |
| CI publishes npm before GitHub assets exist | Push/tag ordering should create the GitHub v10.0.1 release before npm users rely on the wrapper pin. | Confirm release workflow/tag push plan before pushing npm-triggering branch changes. |
| Launcher accepts an unpublished platform | Reject Darwin x64 in npm and skill launchers, matching the v10 release matrix. | npm manager test for resolvePlatform('darwin', 'x64') and separate rg checks over active launcher sources. |
| Generated or ignored artifacts enter the commit | Stage only tracked source/docs/metadata and ADR/codemap files. | git status --porcelain and git diff --cached --name-only. |
| Patch release skips behavioral proof | Re-run Go tests, npm tests/build, live wrapper smoke, C3 checks, and release build smoke. | Commands in Verification table must pass before tagging. |

## Verification

| Check | Result |
| --- | --- |
| go build ./... && go build -tags embedmodel ./... && go test ./... from cli/ | PASS - all Go packages built and tested. |
| npm test && npm run build from packages/cli/ | PASS - manager tests passed, including Darwin x64 rejection, and dist rebuilt at 10.0.1. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh search "real-time sync" | PASS - returned results without SQL FTS hyphen error. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh marketplace show rule-output-via-helpers --json | PASS - expected failure with sunset JSON error and preview hint. |
| bash scripts/build.sh --version 10.0.1 --variant thin --os linux --arch amd64 --out-dir dist/release-smoke | PASS - built thin linux amd64 artifact under dist/release-smoke. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh lookup '.claude-plugin/**' | PASS - plugin manifests resolve to ref-cross-compiled-binary. |
| rg -n "darwin/amd64" packages/cli/src/manager.ts skills/c3/bin/c3x.sh | PASS - no active launcher matches. |
| rg -n "supports linux/darwin on x64/arm64" packages/cli/src/manager.ts skills/c3/bin/c3x.sh | PASS - no stale support hint matches. |
| npm pack --dry-run --json from packages/cli/ | PASS - package would publish as @c3x/cli@10.0.1 with README, package.json, and dist files only. |
| node -e "import('./dist/version.mjs').then(m => console.log(m.C3X_VERSION))" from packages/cli/ | PASS - printed 10.0.1. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr --only adr-20260610-v10-0-1-patch-release | PASS - exit 0 with one non-blocking c3-109 citation snippet warning before terminal implementation. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check | PASS - canonical markdown in sync with .c3. |
| git diff --check | PASS - no whitespace errors. |
