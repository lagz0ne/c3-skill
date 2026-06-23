---
id: adr-20260623-c3x-fat-thin-runtime-resolution-wording
c3-seal: 88338696d9b1947548d2c41d40d661e3ca5eb9aa14d80e2ec03a2a491763c53c
title: c3x-fat-thin-runtime-resolution-wording
type: adr
goal: Clarify the fat/thin distribution reference so the npm manager is described as honoring externally supplied C3X_VERSION only as an override, otherwise resolving project or latest runtime versions.
status: proposed
date: "2026-06-23"
---

## Goal

Clarify the fat/thin distribution reference so the npm manager is described as honoring externally supplied C3X_VERSION only as an override, otherwise resolving project or latest runtime versions.

## Context

After the no-binary wrapper correction, the wrapper no longer exports C3X_VERSION before npm delegation. The governing fat/thin reference still says the npm manager "pins or resolves C3X_VERSION", which blurs package pinning, environment override, and normal project/latest runtime resolution.

## Decision

Patch one sentence in `ref-fat-thin-distribution` so it says the npm manager honors C3X_VERSION when externally set for development/tests, otherwise resolves project or latest runtime versions and downloads verified release assets.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-0 | N.A - system unchanged | The top-level distribution surfaces do not change. | c3-0#n3@v1:sha256:4295e84171aebab432093423315f2571c121a774d6338bf7330dd42644c6dfc2 "Build and distribute C3" | No system patch. |
| c3-2 | N.A - container unchanged | The skill container wording already covers full/portable/no-binary installs. | c3-2#n540@v1:sha256:f0177f46a4bad8f33630a5c2228d6ca7e14117c9787ab7a9b59d45846ffb5866 "Teach an agent to operate C3" | No container patch. |
| c3-203 | component | The wrapper behavior is the governed consumer of the fat/thin distribution wording, but its component contract was already corrected separately. | c3-203#n612@v1:sha256:e8eac431ae98762afad308e9cfdf1ce065aa471537113656131449a7e2e16f4e "Carry bin/c3x.sh and bin/VERSION" | No component patch in this unit. |
| ref-fat-thin-distribution | N.A - governing ref update | The How section contained the stale C3X_VERSION wording being corrected. | ref-fat-thin-distribution#n754@v1:sha256:26dbf94b44e558bae89871d1cf5f4408db14f239e09a5fdc821449ac41e25892 "CI assembles skill archives" | Patch one sentence. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-cross-compiled-binary | The wrapper still selects only binary names from the release matrix before no-binary fallback. | ref-cross-compiled-binary#n730@v1:sha256:cea27fa9abdd975d6298f23e899ceb48f2945fde1e933915fdfada258a190136 "Ship c3x as named per-platform release binaries" | Comply: no binary-matrix change. |
| ref-fat-thin-distribution | This is the governing reference being updated. | ref-fat-thin-distribution#n754@v1:sha256:26dbf94b44e558bae89871d1cf5f4408db14f239e09a5fdc821449ac41e25892 "CI assembles skill archives" | Update-ref: clarify C3X_VERSION override versus project/latest runtime resolution. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Frozen ref | Replace the stale sentence in the How section. | .c3/changes/adr-20260623-c3x-fat-thin-runtime-resolution-wording/01-ref-fat-thin-how.patch.md |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| c3 check | Confirms the ref update remains canvas-valid and cited evidence is fresh. | C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr |

## Verification

| Check | Result |
| --- | --- |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh change view adr-20260623-c3x-fat-thin-runtime-resolution-wording | Required before done. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr | Required before done. |
