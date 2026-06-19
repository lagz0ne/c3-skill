---
id: adr-20260619-design-narrative
c3-seal: 5ce1e08ef119eff2cb4c7e173b2cba2b9385aaf364d45444bbe816fb22b8c565
title: design-narrative
type: adr
goal: Encode C3's own three-act design story — "shape and freeze", "change-units drive the work", "grow the canvas" — as three frozen facts in the `design-act` canvas, authored through one change-unit, so the narrative that justifies C3's mechanics is itself a governed C3 fact rather than prose scattered across SKILL.md and the references.
status: accepted
date: "2026-06-19"
---

## Goal

Encode C3's own three-act design story — "shape and freeze", "change-units drive the work", "grow the canvas" — as three frozen facts in the `design-act` canvas, authored through one change-unit, so the narrative that justifies C3's mechanics is itself a governed C3 fact rather than prose scattered across SKILL.md and the references.

## Context

C3's design rationale currently lives only as explanatory prose inside `skills/c3/SKILL.md` and the operation references (`onboard.md`, `canvas.md`, `change.md`). That prose is the spine that every mechanic hangs off — facts freeze on a gate flip, every frozen-fact edit rides an atomic change-unit, a canvas is a rung you climb — but nothing freezes the spine itself, so it drifts independently of the mechanics it describes and no `c3 check` ever catches the drift. The `design-act` canvas already exists (verified via `c3 schema design-act`: required Thesis, Move, Tool Guarantee table, Why This Shape, Surfaces table) precisely to hold these acts as first-class facts. The affected topology is the `Claude Skill` container (c3-2) and its doc-surface components, plus the C3 underlay (the `design-act` canvas definition and the change-apply CLI) that will store and gate these new facts. No code behavior changes; this is a documentation-as-governed-fact decision.

## Decision

Author the three acts as three standalone `design-act` facts (`act-1-shape-and-freeze`, `act-2-change-units`, `act-3-grow-the-canvas`) through a single change-unit: this ADR plus three no-base `whole` create-patches, landed all-or-nothing by `c3 change apply`. The acts are deliberately born sealed through the change-unit (not `c3 add`) so the act of recording the narrative is itself the saga the narrative describes — the story is created by the mechanism it documents. Each fact's Tool Guarantee table cites the real CLI source that makes its act true (membership healing, the destruction/inspection/canvas gates, the freeze rule), so the narrative stays anchored to executable code and `c3 check` keeps the canvas honest. This wins over leaving the rationale as free prose (the rejected status quo) because only a frozen fact is drift-checked, and it wins over a single combined fact because the three acts have distinct theses, moves, and source anchors that the canvas wants to keep separable.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-2 | container | Hosts the doc surfaces (SKILL.md, onboard.md, canvas.md, change.md) that today carry the design narrative as prose; the new facts become the governed home of that narrative the skill teaches | c3-2#n2927@v1:sha256:3d2a7925787262d374adf87d566af2d6fc20dea692997a30b28eede29c73e74f "Expose c3 architecture workflows through natural language by routing user intent to the right operation and executing it via c3x." | Confirm the three new facts and the skill's narrative prose tell one consistent story |
| c3-117 | component | Owns c3 add/schema/canvas and the change-unit apply path that creates and seals these facts; the create-patches flow through it | c3-117#n2735@v1:sha256:3a377de9129c41b25fa893dc4431afb0aad2814a699e530c70d0de2567204ff6 "Read, write, set, validate schema, and report status for canonical C3 documents." | Verify the no-base whole create-patch path lands custom-type facts cleanly |
| design-act canvas | N.A - canvas definition, not a topology node | The canvas is the contract the three facts must satisfy; it is user-owned markdown, not a system/container/component | N.A - canvas definition, not a topology node | Confirm each fact satisfies c3 schema design-act (Thesis one sentence, every Tool Guarantee row has a Source) |

## Verification

| Check | Result |
| --- | --- |
| c3 change apply adr-20260619-design-narrative | Three facts materialize sealed, all-or-nothing, no canvas rejection |
| c3 read act-1-shape-and-freeze (and act-2, act-3) | Each renders all five required design-act sections |
| c3 check | issues[0] — canonical markdown in sync, canvas satisfied |
