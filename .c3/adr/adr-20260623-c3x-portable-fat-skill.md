---
id: adr-20260623-c3x-portable-fat-skill
c3-seal: c9e126eca7327724c7cad5cbe83eee37c37c8ea593d7492a322ecf3835fafebc
title: c3x-portable-fat-skill
type: adr
goal: Add a portable Linux fat skill distribution path for C3 so sandboxed or distro-diverse Linux environments can run a bundled `c3x` binary without npm, Go, glibc-linked ONNX runtime, or first-use downloads, while preserving the existing full-fat semantic build and the no-binary plugin install path.
status: proposed
date: "2026-06-23"
---

## Goal

Add a portable Linux fat skill distribution path for C3 so sandboxed or distro-diverse Linux environments can run a bundled `c3x` binary without npm, Go, glibc-linked ONNX runtime, or first-use downloads, while preserving the existing full-fat semantic build and the no-binary plugin install path.

## Context

The release already ships full-fat skill ZIPs that bundle the platform `c3x` binary and embedded semantic assets, plus a platform-neutral no-binary skill/plugin ZIP that falls through to the pinned npm runtime manager. The full-fat Linux build is feature-complete but cgo/native-runtime oriented, so it is not the best artifact for Alpine, musl, distroless-like sandboxes, or tightly isolated environments. The user wants a more generic fat form for most Linux distros without weakening the current fat archive or making passive wrapper commands download runtime assets.

## Decision

Add a Linux-only `portable` build variant that compiles `c3x` with `CGO_ENABLED=0` and pure-Go build tags, emits `c3x-{VERSION}-linux-{arch}-portable`, and packages it as `c3-skill-linux-{arch}-portable-v{VERSION}.zip`. The wrapper selection order becomes full bundled binary, Linux portable bundled binary, source build, then pinned npm manager. Full-fat artifacts remain the semantic/native ONNX path; portable fat artifacts are core bundled runtime artifacts where semantic ONNX search is unavailable and search falls back to keyword/graph behavior.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-0 | N.A - parent distribution shape unchanged | The top-level three-surface story remains Go CLI, agent skill, and npm runtime manager; portable fat is a variant inside the skill/binary distribution surface. | c3-0#n8@v1:sha256:cf977c208b843b43d3bf9e9dc3e264bbfce54da033d44082b71b2f9e919080e0 "Teach an agent to operate C3 through shared skill instructions" | No system goal change. |
| c3-2 | container | The skill container now distinguishes full-fat, portable-fat, and no-binary installer paths in its packaging responsibility. | c3-2#n548@v1:sha256:bf815d792d8f9c65f3ecc2298b6e6a4edda15389ec9c7736dfcca49aa979317e "Carry the skill definition (SKILL.md: the intent router and the three-act model), the per-" | Update parent responsibility/complexity wording only. |
| c3-203 | component | The wrapper gains an intermediate portable-binary selection path before source/npm fallback. | c3-203#n603@v1:sha256:54edea1eb796d0b265904816dad70fc6cf7ee1e15430a72c1d27662a9ad038ce "Detect the host platform, select a version-pinned full or Linux portable packaged binary," | Update wrapper goal, purpose, governance, contract, and derived-material evidence. |
| ref-cross-compiled-binary | N.A - governing ref update | The binary matrix now includes full semantic builds and separate pure-Go Linux portable builds. | ref-cross-compiled-binary#n730@v1:sha256:cea27fa9abdd975d6298f23e899ceb48f2945fde1e933915fdfada258a190136 "Ship c3x as named per-platform release binaries, with standard runtime-manager binaries," | Update the binary-distribution reference. |
| ref-fat-thin-distribution | N.A - governing ref update | The distribution split changes from full-fat/no-binary/thin to full-fat/portable-fat/no-binary/thin. | ref-fat-thin-distribution#n748@v1:sha256:cf8e08bcc48d161ef4120914a8505499068ac512709adeae244258fd1618b031 "Distribute C3 as full-fat skill ZIPs (semantic binaries bundled), Linux portable fat skill" | Update the fat/thin distribution reference. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-cross-compiled-binary | Portable fat changes the release binary matrix and artifact naming the wrapper resolves. | ref-cross-compiled-binary#n730@v1:sha256:cea27fa9abdd975d6298f23e899ceb48f2945fde1e933915fdfada258a190136 "Ship c3x as named per-platform release binaries, with standard runtime-manager binaries," | Update-ref: include full semantic binaries plus Linux portable binaries. |
| ref-fat-thin-distribution | Portable fat changes the skill archive shapes and wrapper fallback expectations. | ref-fat-thin-distribution#n748@v1:sha256:cf8e08bcc48d161ef4120914a8505499068ac512709adeae244258fd1618b031 "Distribute C3 as full-fat skill ZIPs (semantic binaries bundled), Linux portable fat skill" | Update-ref: describe full-fat, portable-fat, no-binary, and thin npm surfaces. |
| ref-frontmatter-docs | Frozen facts must stay frontmatter plus canvas-shaped markdown and mutate through a change-unit. | ref-frontmatter-docs#n757@v1:sha256:d4f7719668519e2f2a93de15969bc53c8f0105e7e073231a2f36d7c2626cb361 | Comply: stage block patches and verify with local c3x checks. |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| rule-dispatcher-error-hint | The wrapper error hint changes when the portable artifact becomes a valid recovery path. | rule-dispatcher-error-hint#n766@v1:sha256:bd662000c1bc5b93d0b1cc4cf532cc1dc6e4766e5bda6b544f8aab14d21f7dc4 | Comply: keep the missing-runtime error actionable. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Build script | Add portable and release variants; portable sets CGO_ENABLED=0, uses pure-Go build tags, and is limited to Linux targets. | scripts/build.sh |
| Release assembly | Package portable/ artifacts as c3-skill-linux-{arch}-portable-v{VERSION}.zip without replacing full-fat archives. | scripts/assemble_release_assets.sh |
| Wrapper dispatch | Prefer full bundled binary, then Linux portable bundled binary, then source build, then npm manager. | skills/c3/bin/c3x.sh |
| Tests | Assert archive shape, portable-before-npm wrapper dispatch, static ELF properties, and bwrap network-isolated --help. | scripts/test_skill_release_packaging.py |
| Docs | Explain full-fat vs portable-fat vs no-binary behavior. | README.md; packages/cli/README.md; npm CLI spec/plan |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| Release tooling | scripts/build.sh adds the Linux portable variant and release variant; workflows call the release variant. | python3 scripts/test_skill_release_packaging.py; workflow diff |
| Skill wrapper | skills/c3/bin/c3x.sh adds the Linux portable binary fallback before source/npm. | Wrapper test in scripts/test_skill_release_packaging.py |
| Frozen architecture facts | c3-2, c3-203, ref-cross-compiled-binary, and ref-fat-thin-distribution receive scoped patches. | C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| scripts/test_skill_release_packaging.py | Verifies full-fat, portable-fat, and no-binary skill archives plus wrapper portable selection. | Focused packaging test |
| readelf portable smoke | Fails if the local portable Linux binary has an ELF interpreter. | readelf -l assertion in packaging test |
| bwrap portable smoke | Runs portable c3x --help in a network-isolated bubblewrap sandbox. | bwrap --unshare-net ... /c3x --help in packaging test |
| release workflows | Build the release variant so Linux produces thin, full-fat, and portable artifacts. | .github/workflows/release.yml; .github/workflows/distribute.yml |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Replace full-fat Linux with portable only | That would remove offline semantic ONNX behavior from users who want the full feature set. |
| Build musl/cgo ONNX full-fat artifacts | The upstream ONNX Runtime path is native and heavier; it is not a small packaging change and would widen the security/build surface. |
| Publish one universal Linux binary for amd64 and arm64 | Linux ELF binaries are architecture-specific; the release must keep separate arch artifacts. |
| Make macOS universal in this slice | The Go binary can be lipo-combined, but the ONNX Runtime dylib path is arch-specific; separate mac work is safer. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Portable fat is mistaken for full semantic fat | Name artifacts with -portable and document semantic ONNX unavailability. | README/spec updates and archive tests. |
| Passive commands invoke npm despite a bundled portable binary | Wrapper checks portable before source/npm. | Wrapper portable-before-npm test. |
| Portable build silently regains a dynamic interpreter | Build with CGO_ENABLED=0 and assert no ELF interpreter. | readelf -l assertion. |
| Frozen docs drift or lose quality | Use scoped C3 change-unit patches and rerun local checks. | c3x check --include-adr. |

## Verification

| Check | Result |
| --- | --- |
| python3 scripts/test_skill_release_packaging.py | Required before done. |
| cd packages/cli && npm test | Required before done. |
| cd cli && go test ./... | Required before done. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check | Required before done. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr | Required before done. |
| Independent multi-agent review of portability/security/doc consistency | Required before done. |
