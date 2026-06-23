---
id: adr-20260623-c3x-wrapper-no-binary-runtime-resolution
c3-seal: 317e09231099411e3f3fd3a9ab86e86778c791e9c50973f342b843bdc1c5828e
title: c3x-wrapper-no-binary-runtime-resolution
type: adr
goal: Correct the no-binary C3 skill wrapper contract so passive root commands stay local and real npm-delegated commands do not force C3X_VERSION over the npm manager's project/latest runtime resolution.
status: proposed
date: "2026-06-23"
---

## Goal

Correct the no-binary C3 skill wrapper contract so passive root commands stay local and real npm-delegated commands do not force C3X_VERSION over the npm manager's project/latest runtime resolution.

## Context

The portable fat skill work made bundled full and Linux portable binaries authoritative when present, but independent review found the no-binary fallback still had two stale behaviors: passive root help/version could reach `npm exec`, and the wrapper exported C3X_VERSION before npm delegation. The first behavior can download an npm package for a passive command; the second makes the platform-neutral no-binary skill override project runtime metadata instead of letting the pinned manager package resolve the runtime.

## Decision

Keep bundled and source-built binaries first so their native help/version output remains authoritative. Only when no bundled/source runtime exists should the shell wrapper answer root help/version locally, and only real commands should delegate through `npm exec --yes --package @c3x/cli@${VERSION} -- c3x` without exporting C3X_VERSION. This preserves package pinning for the manager while keeping runtime version selection inside the manager.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-0 | N.A - system unchanged | The system distribution story stays the same; this change only clarifies no-binary wrapper behavior inside the skill runtime path. | c3-0#n3@v1:sha256:4295e84171aebab432093423315f2571c121a774d6338bf7330dd42644c6dfc2 "Build and distribute C3" | No system patch. |
| c3-2 | N.A - container unchanged | The skill container owns the wrapper surface, but its full/portable/no-binary responsibility wording already covers this correction. | c3-2#n540@v1:sha256:f0177f46a4bad8f33630a5c2228d6ca7e14117c9787ab7a9b59d45846ffb5866 "Teach an agent to operate C3" | No container patch. |
| c3-203 | component | The wrapper contract said npm fallback exports C3X_VERSION and did not name local passive handling for no-binary installs. | c3-203#n612@v1:sha256:e8eac431ae98762afad308e9cfdf1ce065aa471537113656131449a7e2e16f4e "Carry bin/c3x.sh and bin/VERSION" | Patch only the stale purpose, runtime contract, and derived-material rows. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-cross-compiled-binary | The wrapper still selects only release asset names produced by the binary matrix before any no-binary fallback behavior applies. | ref-cross-compiled-binary#n730@v1:sha256:cea27fa9abdd975d6298f23e899ceb48f2945fde1e933915fdfada258a190136 "Ship c3x as named per-platform release binaries" | Comply: do not change the platform matrix or bundled full/portable selection order. |
| ref-fat-thin-distribution | The change affects the no-binary skill/plugin path inside the fat/thin distribution split. | ref-fat-thin-distribution#n748@v1:sha256:cf8e08bcc48d161ef4120914a8505499068ac512709adeae244258fd1618b031 "Distribute C3 as full-fat skill ZIPs" | Comply: keep no-binary installs delegated to the pinned npm manager package without turning the wrapper into runtime authority. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Wrapper | Add local passive handling after bundled/source runtime checks and remove wrapper-created C3X_VERSION from npm fallback. | skills/c3/bin/c3x.sh |
| Tests | Assert no-binary help/version do not invoke npm and npm fallback receives no wrapper-forced C3X_VERSION. | scripts/test_skill_release_packaging.py |
| Frozen fact | Update c3-203 rows that described C3X_VERSION export as universal. | .c3/changes/adr-20260623-c3x-wrapper-no-binary-runtime-resolution/*.patch.md |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| Wrapper tests | Catch passive npm invocation and runtime-version override regressions. | scripts/test_skill_release_packaging.py |
| c3 check | Confirms the wrapper fact remains canvas-valid after the change-unit applies. | C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Fat/source help gets replaced by the wrapper's generic help | Place passive handling after bundled/source runtime checks. | Wrapper portable-before-npm test still proves bundled portable handles --help. |
| No-binary real commands lose package pinning | Keep the npm package spec pinned to @c3x/cli@${VERSION}; only remove wrapper-created C3X_VERSION. | npm fallback argument capture test. |

## Verification

| Check | Result |
| --- | --- |
| python3 scripts/test_skill_release_packaging.py | Required before done. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh change view adr-20260623-c3x-wrapper-no-binary-runtime-resolution | Required before done. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr | Required before done. |
