---
id: adr-20260608-fat-default-thin-npm
c3-seal: 148d6b4da98e8e81544179d76df0f792cabe5afdb6d1db589b93ee4a37336ff1
title: fat-default-thin-npm
type: adr
goal: Flip C3 distribution so packaged skill installs are fat self-contained by default while `@c3x/cli` becomes the only thin opt-in manager path.
status: implemented
date: "2026-06-08"
---

## Goal

Flip C3 distribution so packaged skill installs are fat self-contained by default while `@c3x/cli` becomes the only thin opt-in manager path.

## Context

Current branch made thin the default: `skills/c3/bin/c3x.sh` can download/cache thin release assets and choose fat through variant environment, release packaging treats thin assets as primary, and the npm manager still contains variant/fat logic. User requirement is to preserve self-contained skill zip behavior as the default and keep thin downloads only inside the npm manager. Affected topology is Go CLI distribution, npm wrapper delegation, release workflow, and cross-compiled binary distribution policy.

## Decision

Use fat binaries built with `-tags embedmodel` as the primary per-platform skill zip artifact. Simplify `skills/c3/bin/c3x.sh` to exec a present local platform binary and only fall back to source build for checkout development when no packaged binary exists. Make `@c3x/cli` thin-only: resolve pinned version and platform, download thin binary/model/vocab/checksums from GitHub Releases, verify SHA256, cache under `~/.cache/c3x/<version>/`, and exec. Remove fat/variant switching from npm and skill launcher code so each install path has one responsibility.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-1 | container | Distribution build and npm delegation behavior are CLI container responsibilities. | c3-1#n1557@v1:sha256:1f10b60594a91ba79c9f39e43539257e287197fb4d527506c35b42a09cf7bb95 "Compile to a self-contained binary for all supported platforms" | Review responsibilities and component membership for parent delta. |
| c3-109 | component | The npm manager changes to thin-only release asset caching and exec. | c3-109#n1974@v1:sha256:0a83b9a8acb0b670c3a4deeb6a99f64cedb3ac13ee92ee737c18cc0f4f3fc385 "Own the thin npm manager used by humans and scripts that want npx @c3x/cli or a global npm command. The component resolves the supported OS/arch, pinned C3 ve" | Update component contract if existing text mentions stale variant or discovery behavior. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-cross-compiled-binary | It directly governs skills/c3/bin/c3x.sh, scripts/build.sh, release packaging, and npm release asset layout. | ref-cross-compiled-binary#n3010@v1:sha256:0f8a3f16f77f8de3bea94816048e672c09a4b6700f59f12e3c302a64b3eaf3aa "Publish two distribution variants. Thin is the default: release assets provide c3x-<version>-<os>-<arch> binaries plus semantic model/vocab assets, and launch" | update-ref |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| rule-dispatcher-error-hint | npm manager and launcher failures must remain actionable when platform, checksum, download, or build fallback fails. | rule-dispatcher-error-hint#n3051@v1:sha256:182738dc1f72a50edf736dd26e32f990ba604b9e04a107dcb772922f769575e1 "User-facing dispatcher errors must carry an error: prefix and, when the failure is recoverable, a hint: line naming the next step." | comply |
| rule-output-via-helpers | Go command output stays inside the Go binary; wrapper/launcher only execs and does not add agent-facing JSON paths. | rule-output-via-helpers#n3071@v1:sha256:ca4a0c295ace95c5e32a1c7ca5b92a583f8637573bdec86d8513c1c6d7d486be "Commands must emit results via WriteTableOutput/WriteObjectOutput with a format from ResolveFormat — never call json.Marshal or fmt.Fprintf to seria" | comply |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| skills/c3/bin/c3x.sh | Remove variant and download/cache logic; exec local packaged binary; keep source-build dev fallback. | User requested simplified launcher. |
| .github/workflows/distribute.yml | Build primary per-platform fat skill zips with -tags embedmodel; keep thin binary/model/vocab/checksums for npm. | User requested fat zips as headline/default. |
| packages/cli | Remove fat/variant handling and keep thin-only download, SHA256 verify, cache, exec behavior. | User requested npm manager thin-only. |
| C3 docs | Update distribution ref and component docs if stale after behavior change. | C3 lookup and ref read show thin default text. |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| Distribution ref | Update ref-cross-compiled-binary choice/how text from thin default to fat skill zip default plus thin npm manager. | c3x check --include-adr after update. |
| npm component | Update c3-109 docs only where stale behavior references old discovery or fat fallback. | c3x check --only c3-109. |
| CLI/distribution code | No new C3 validators or command help are required because this change is packaging/runtime wrapper behavior, verified by builds/tests/smoke. | Go builds/tests, npm test/build, script build, launcher smoke. |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| Go compiler | Thin and fat binaries both compile. | cd cli && go build ./...; cd cli && go build -tags embedmodel ./.... |
| Go tests | Existing CLI behavior remains green. | cd cli && go test ./.... |
| npm tests/build | Thin manager tests and package build remain green. | cd packages/cli && npm test; cd packages/cli && npm run build. |
| skill build | Local script builds fat skill artifacts. | bash scripts/build.sh. |
| launcher smoke | Present skill-bin binary execs without variant env or download. | Local dev binary copied into skill bin and bash skills/c3/bin/c3x.sh --version. |
| whitespace guard | No whitespace errors. | git diff --check. |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Keep thin as default and require fat env selection | Violates user requirement that skill zip continues self-contained just-works behavior. |
| Keep variant support in both launcher and npm manager | Keeps duplicated responsibility and makes default behavior ambiguous across install paths. |
| Make npm manager able to download fat binaries too | Inflates npm path and undermines requested thin-only manager boundary. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Skill zip misses embedded model | Build release skill binaries with -tags embedmodel and verify fat build compiles. | cd cli && go build -tags embedmodel ./...; bash scripts/build.sh. |
| npm manager fetches wrong artifact | Remove variant suffix logic and test thin asset naming/checksum paths. | cd packages/cli && npm test. |
| Launcher accidentally downloads in default path | Delete download/cache branches from c3x.sh and smoke with local binary only. | bash skills/c3/bin/c3x.sh --version with present bin, no variant env. |
| Docs drift | Update governing ref/component docs in same change. | c3x check --include-adr. |

## Verification

| Check | Result |
| --- | --- |
| cd cli && go build ./... | PASS. |
| cd cli && go build -tags embedmodel ./... | PASS. |
| cd cli && go test ./... | PASS. |
| cd packages/cli && npm test | PASS after npm ci installed lockfile dev dependencies. |
| cd packages/cli && npm run build | PASS. |
| bash scripts/build.sh | PASS; default built fat artifact under dist/c3x/fat/. |
| bash skills/c3/bin/c3x.sh --version with copied local binary in skill bin and no variant env | PASS; printed 9.9.1 with C3X_RELEASE_BASE_URL pointed at 127.0.0.1:9. |
| git diff --check | PASS. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check | PASS; canonical markdown synced. |
