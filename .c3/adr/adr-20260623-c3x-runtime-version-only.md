---
id: adr-20260623-c3x-runtime-version-only
c3-seal: 6b6e4ed57ecb8c66404fe38390abde38dc39e739b182435bcb80f410b3a02499
title: c3x-runtime-version-only
type: adr
goal: Tighten the npm runtime manager contract so project metadata stores only the selected runtime version, and align the parent package membership row with the runtime-manager component goal.
status: done
date: "2026-06-23"
supersedes:
    - adr-20260623-c3x-runtime-manager
---

# C3x Runtime Version Only Metadata

## Goal

Tighten the npm runtime manager contract so project metadata stores only the selected runtime version, and align the parent package membership row with the runtime-manager component goal.

## Context

Independent review found that `.c3/runtime.json` allowed an `updatedAt` field even though the ratified trust boundary is version-only operational data, and the `c3-3` Components table still carried the old pinned-downloader goal contribution. The code, docs, and frozen package facts need to agree on the stricter contract.

## Decision

Project runtime metadata may contain only `version`. The npm manager rejects every other field, including timestamps. The `c3-3` Components row is updated as a parent-delta fact patch to reflect the child component's current runtime-manager contribution.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-0 | system | This narrows one npm distribution surface's trust boundary within the current Go CLI, agent skill, and npm runtime-manager system shape. | c3-0#n3@v1:sha256:4295e84171aebab432093423315f2571c121a774d6338bf7330dd42644c6dfc2 | Top-down context reviewed; no system body patch needed for the runtime metadata narrowing. |
| c3-3 | container | The Components table now names the runtime-manager contribution for c3-301. | c3-3#n493@v1:sha256:7fe11549f06a65e151fbd2f5b82c6f76e0f65f3c67252e74565d1a36b185883b | Parent Delta: updated the member goal contribution row. |
| c3-301 | component | The project-metadata contract row now says version-only and rejects timestamps. | c3-301#n660@v1:sha256:4ec6e5316af5f059d7b0e1f3cf2ae23e9f3ee981157300e4ad88919a7bd737b4 | Updated the Contract row and matching package tests/docs. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-cross-compiled-binary | The narrowed metadata contract still selects release-built platform assets. | ref-cross-compiled-binary#n730@v1:sha256:cea27fa9abdd975d6298f23e899ceb48f2945fde1e933915fdfada258a190136 "Ship c3x as named per-platform release binaries, with standard runtime-manager binaries," | Comply: no asset naming change. |
| ref-fat-thin-distribution | The npm package remains the thin runtime manager and does not bundle binaries. | ref-fat-thin-distribution#n748@v1:sha256:cf8e08bcc48d161ef4120914a8505499068ac512709adeae244258fd1618b031 "Distribute C3 as full-fat skill ZIPs (semantic binaries bundled), Linux portable fat skill" | Comply: no distribution split change for npm. |
| ref-eval-determinism | Inherited from the top-level system; this work does not change eval verdict computation or eval fixtures. | ref-eval-determinism#n739@v1:sha256:d914f393b17de0202b7ae4cdde4df7d173c51fd820b2695487a29efb06f514d7 | N.A - runtime metadata narrowing does not change eval computation. |
| ref-frontmatter-docs | This ADR updates frozen C3 package facts and must preserve frontmatter plus canvas-shaped markdown. | ref-frontmatter-docs#n757@v1:sha256:d4f7719668519e2f2a93de15969bc53c8f0105e7e073231a2f36d7c2626cb361 | Comply: facts move through a change-unit and finish with c3x check. |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| rule-dispatcher-error-hint | New metadata rejection is user-facing and should explain the safe shape. | rule-dispatcher-error-hint#n625@v1:sha256:bd662000c1bc5b93d0b1cc4cf532cc1dc6e4766e5bda6b544f8aab14d21f7dc4 | Comply: unknown fields produce an error: plus version-only hint. |
| rule-output-via-helpers | Inherited from Go CLI components; npm runtime commands are human package-management output. | rule-output-via-helpers#n638@v1:sha256:b5ac8121ffc54be6c8f87ec133e69658fea023e7e73da3859fb85a33869afa29 | N.A - normal Go commands still own structured output. |
| rule-wrap-error-cause | Inherited from Go CLI components; this metadata narrowing is direct validation, not wrapped Go boundary errors. | rule-wrap-error-cause#n650@v1:sha256:b9e4edb84b11060973de3fe6e5c0ab7b5605aa690e00e886335b054bdaab710f | N.A - no Go error-boundary change. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Metadata parser | Reject any .c3/runtime.json field other than version. | packages/cli/src/manager.ts |
| Tests | Assert updatedAt is rejected and direct runtime commands are covered. | packages/cli/test/manager.test.mjs |
| C3 facts | Patch c3-3 Components row and c3-301 Contract row. | .c3/changes/adr-20260623-c3x-runtime-version-only/ |

## Verification

| Check | Result |
| --- | --- |
| cd packages/cli && npm test | Required before done. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh change apply adr-20260623-c3x-runtime-version-only | Required to update frozen facts. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr | Required before done. |
