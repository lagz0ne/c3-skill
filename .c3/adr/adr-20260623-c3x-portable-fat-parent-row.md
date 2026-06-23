---
id: adr-20260623-c3x-portable-fat-parent-row
c3-seal: 8c8f739dc9ac33dcb775dfa834a72f31f4bae6087dc4eca6bddf0ae8e93bf333
title: c3x-portable-fat-parent-row
type: adr
goal: Bring the `c3-2` Components membership row for `c3-203` back into alignment with the wrapper goal introduced by the portable fat skill distribution decision.
status: proposed
date: "2026-06-23"
---

## Goal

Bring the `c3-2` Components membership row for `c3-203` back into alignment with the wrapper goal introduced by the portable fat skill distribution decision.

## Context

The portable fat skill change updated `c3-203` so the wrapper now selects a full bundled binary, then a Linux portable bundled binary, before source and npm fallbacks. The `c3-2` Components table still carries the prior generated summary for `c3-203`, so the parent row understates the wrapper's distribution behavior even though the child fact is already correct.

## Decision

Patch only the `c3-2` table row for `c3-203` to mirror the new wrapper goal. Do not change the system goal, the skill container goal, or the wrapper fact again.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-0 | N.A - parent unchanged | The system-level distribution surfaces do not change; this only repairs a child summary row under the skill container. | c3-0#n8@v1:sha256:cf977c208b843b43d3bf9e9dc3e264bbfce54da033d44082b71b2f9e919080e0 "Teach an agent to operate C3 through shared skill instructions" | No system patch. |
| c3-2 | container | The membership row for c3-203 must reflect the child wrapper's current goal. | c3-2#n546@v1:sha256:37b1322deb3f7fc242a5ab98c0b254bbca9ccd1943aebb3ec23fcaced64906e2 "Detect the host platform, select a version-pinned full" | Patch one table row. |
| c3-203 | N.A - child already updated | The child fact already carries the portable/full binary wording this row summarizes. | c3-203#n603@v1:sha256:54edea1eb796d0b265904816dad70fc6cf7ee1e15430a72c1d27662a9ad038ce "Detect the host platform, select a version-pinned full or Linux portable packaged binary," | No child patch. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-frontmatter-docs | Frozen fact updates must stay canvas-shaped and mutate through a change-unit. | ref-frontmatter-docs#n757@v1:sha256:d4f7719668519e2f2a93de15969bc53c8f0105e7e073231a2f36d7c2626cb361 | Comply: stage one table-row patch and verify with local c3x checks. |
| ref-cross-compiled-binary | The child wrapper is governed by the binary-matrix ref, but this ADR only repairs parent wording after that ref was updated. | ref-cross-compiled-binary#n730@v1:sha256:cea27fa9abdd975d6298f23e899ceb48f2945fde1e933915fdfada258a190136 "Ship c3x as named per-platform release binaries, with standard runtime-manager binaries," | Comply: no ref patch in this parent-row unit. |
| ref-fat-thin-distribution | The child wrapper is governed by the fat/thin distribution ref, but this ADR only repairs parent wording after that ref was updated. | ref-fat-thin-distribution#n748@v1:sha256:cf8e08bcc48d161ef4120914a8505499068ac512709adeae244258fd1618b031 "Distribute C3 as full-fat skill ZIPs (semantic binaries bundled), Linux portable fat skill" | Comply: no ref patch in this parent-row unit. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Parent row | Update the c3-203 row in .c3/c3-2-/README.md to mention full and Linux portable packaged binaries. | .c3/changes/adr-20260623-c3x-portable-fat-parent-row/01-c3-2-component-row.patch.md |

## Verification

| Check | Result |
| --- | --- |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh change view adr-20260623-c3x-portable-fat-parent-row | Required before done. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr | Required before done. |
