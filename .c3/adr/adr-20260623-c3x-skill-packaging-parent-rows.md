---
id: adr-20260623-c3x-skill-packaging-parent-rows
c3-seal: 6393a9af59767bcaabdf74e22c2f840e2b3ee1aa4b239d5f37d5837d44cdae3d
title: c3x-skill-packaging-parent-rows
type: adr
goal: Align the parent membership rows for the C3 skill container and cli-wrapper component with the just-landed skill packaging split.
status: done
date: "2026-06-23"
---

# Skill Packaging Parent Rows

## Goal

Align the parent membership rows for the C3 skill container and cli-wrapper component with the just-landed skill packaging split.

## Context

The skill packaging split updated c3-2 and c3-203 goals and responsibilities, but the parent membership rows in c3-0 and c3-2 still show the old local-binary-only summaries. That makes the frozen topology read as if the no-binary npm-manager fallback does not exist.

## Decision

Patch only the two stale parent rows: c3-0's c3-2 row and c3-2's c3-203 row. No implementation behavior changes are part of this follow-up; it is a parent-delta documentation correction for the applied packaging decision.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-0 | system | Its c3-2 container row must summarize the updated Claude skill packaging and runtime-wrapper behavior. | c3-0#n8@v1:sha256:cf977c208b843b43d3bf9e9dc3e264bbfce54da033d44082b71b2f9e919080e0 "Teach an agent to operate C3 through shared skill instructions" | Parent Delta: update one membership row. |
| c3-2 | container | Its c3-203 component row must summarize the wrapper's bundled/source/npm runtime selection behavior. | c3-2#n546@v1:sha256:37b1322deb3f7fc242a5ab98c0b254bbca9ccd1943aebb3ec23fcaced64906e2 "Detect the host platform, select a version-pinned full" | Parent Delta: update one membership row. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-frontmatter-docs | The row patches still mutate frozen facts and must preserve canvas-shaped markdown. | ref-frontmatter-docs#n757@v1:sha256:d4f7719668519e2f2a93de15969bc53c8f0105e7e073231a2f36d7c2626cb361 | Comply: land through a change-unit and check. |
| ref-fat-thin-distribution | The rows describe the artifact split governed by this ref. | ref-fat-thin-distribution#n748@v1:sha256:cf8e08bcc48d161ef4120914a8505499068ac512709adeae244258fd1618b031 "Distribute C3 as full-fat skill ZIPs (semantic binaries bundled), Linux portable fat skill" | Comply: row wording matches fat skill, no-binary skill/plugin, thin npm runtime manager, and later portable fat naming. |
| ref-cross-compiled-binary | Inherited through c3-203, but this follow-up changes only parent row text and not binary naming or platform gates. | ref-cross-compiled-binary#n730@v1:sha256:cea27fa9abdd975d6298f23e899ceb48f2945fde1e933915fdfada258a190136 "Ship c3x as named per-platform release binaries, with standard runtime-manager binaries," | N.A - no implementation or asset-matrix change. |
| ref-eval-determinism | Inherited through the broader system, but this follow-up changes no eval computation or selectors. | ref-eval-determinism#n739@v1:sha256:d914f393b17de0202b7ae4cdde4df7d173c51fd820b2695487a29efb06f514d7 | N.A - parent row text only. |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| rule-dispatcher-error-hint | Inherited through command components, but this follow-up changes no user-facing command failure path. | rule-dispatcher-error-hint#n766@v1:sha256:bd662000c1bc5b93d0b1cc4cf532cc1dc6e4766e5bda6b544f8aab14d21f7dc4 | N.A - parent row text only. |
| rule-output-via-helpers | Inherited through Go CLI output components, but this follow-up changes no command serialization. | rule-output-via-helpers#n779@v1:sha256:b5ac8121ffc54be6c8f87ec133e69658fea023e7e73da3859fb85a33869afa29 | N.A - parent row text only. |
| rule-wrap-error-cause | Inherited through Go command boundaries, but this follow-up changes no error wrapping. | rule-wrap-error-cause#n791@v1:sha256:b9e4edb84b11060973de3fe6e5c0ab7b5605aa690e00e886335b054bdaab710f | N.A - parent row text only. |

## Verification

| Check | Result |
| --- | --- |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr --only adr-20260623-c3x-skill-packaging-parent-rows | Required before done. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr | Required before done. |
