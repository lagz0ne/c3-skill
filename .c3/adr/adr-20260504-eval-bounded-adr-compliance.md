---
id: adr-20260504-eval-bounded-adr-compliance
c3-seal: ae7b3841ed61aa3346c505a0d2dde8c1cfa4bd64b32f8b07c534fc4eb406d53b
title: eval-bounded-adr-compliance
type: adr
goal: Eliminate the ADR authoring flood where content validation demands every inferred compliance ref or rule before an agent can finish a correct ADR draft.
status: implemented
date: "2026-05-04"
---

## Goal

Eliminate the ADR authoring flood where content validation demands every inferred compliance ref or rule before an agent can finish a correct ADR draft.

## Context

Current ADR add and write validation derives inherited compliance refs and rules from Affected Topology and reports each missing item as an error. The current eval pressure shows this turns one ADR authoring step into many repair turns, especially when broad topology pulls in unrelated refs and rules. Authored compliance rows still need concrete Why required text, but inferred coverage gaps should be advisory during authoring.

## Decision

Keep strict validation for rows the author explicitly writes, but stop treating inferred missing compliance coverage as add or write content errors. Leave inferred coverage as check-time guidance so agents can start from a bounded ADR and improve review coverage without being blocked by a large generated list.

## Affected Topology

| Entity | Type | Why affected | Governance review |
| --- | --- | --- | --- |
| c3-111 | component | ADR add validates creation bodies before insert and currently blocks on inferred compliance coverage. | Review add-cmd validation contract. |
| c3-113 | component | check-cmd owns ADR consistency diagnostics and should keep advisory coverage signals. | Review check-cmd warning behavior. |
| c3-117 | component | docs-state write surfaces the content validation failure shown in the user report. | Review write path keeps authored-row validation. |

## Compliance Refs

| Ref | Why required | Action |
| --- | --- | --- |
| N.A - no governing ref | N.A - current local C3 lookup surfaced component ownership but no ref governance for these files. | N.A - no ref compliance row required. |

## Compliance Rules

| Rule | Why required | Action |
| --- | --- | --- |
| N.A - no governing rule | N.A - current local C3 lookup surfaced no rule governance for these files. | N.A - no rule compliance row required. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| cli/cmd/adr_linkage.go | Split authored row validation from inferred missing coverage so add and write can call strict row checks without coverage flood. | Focused ADR Go tests. |
| cli/cmd/add.go | Use authoring-safe ADR validation during add and dry-run. | Focused ADR add tests. |
| cli/cmd/add_test.go | Replace old implicit-ref blocking expectation with eval-backed bounded authoring expectation. | Red then green focused test. |
| cli/cmd/check_enhanced_test.go | Preserve check-time advisory missing coverage behavior. | Existing check test remains green. |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| ADR validator | Add an authoring path that validates explicit Compliance Refs and Compliance Rules row quality but skips inferred missing rows. | Focused Go tests for bounded add, blank why rejection, and check warnings. |
| Add command | Wire add and dry-run to the authoring path. | Focused Go tests for ADR add. |
| Check command | Keep include-ADR check using full coverage warnings. | Existing check warning test stays green. |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| c3x add adr | Accepts bounded ADR bodies that omit inferred inherited compliance coverage. | Focused Go test. |
| c3x add adr | Still rejects explicit Compliance Refs or Compliance Rules rows with blank Why required. | Existing focused Go test. |
| c3x check include ADR | Still warns about missing inferred compliance coverage for review. | Existing focused Go test. |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Keep inferred coverage as authoring errors | Current eval and user trace show it creates a high-turn repair loop before ADR writing can converge. |
| Remove compliance coverage checks entirely | Check-time warnings are still useful review guidance after the ADR exists. |
| Auto-fill every missing row as N.A | That would increase boilerplate and fight the eval quality guard against boilerplate N.A rows. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| ADRs omit relevant refs forever | Keep check-time include-ADR warnings for missing inferred coverage. | Check warning test stays green. |
| Blank authored compliance rationale slips through | Keep strict Why required validation for rows authors write. | Existing Why column test stays green. |
| Add and check diverge silently | Focused tests cover both authoring and check behavior. | Focused Go tests plus all CLI tests. |

## Verification

| Check | Result |
| --- | --- |
| Focused ADR add and check Go tests | Passed: go test ./cmd focused bounded add, blank why rejection, and check warning tests. |
| go test ./... from cli | Passed under binbag lease. |
| c3local check include ADR only this ADR | Passed before implementation status update. |
