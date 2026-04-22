---
id: adr-20260415-strict-component-docs-semantic-tightening
c3-seal: 0cd65ec6c06c64a0eb62d33ae05b5f4291d5dfce6da6372de1aeac8f5626f6a5
title: strict-component-docs-semantic-tightening
type: adr
goal: Tighten strict component documentation so valid-looking tables cannot pass with ungrounded references, weak evidence, overused N.A values, or repeated boilerplate.
status: implemented
date: "2026-04-15"
---

# strict-component-docs-semantic-tightening

## Goal

Tighten strict component documentation so valid-looking tables cannot pass with ungrounded references, weak evidence, overused N.A values, or repeated boilerplate.

## Context

The first strict component-doc pass established the required sections. The next failure mode was subtler: a document could satisfy headings and table shape while still saying little. Examples included ungrounded reference columns, evidence that did not name a command/file/entity, all-`N.A` rows, repeated generated boilerplate, and derivation rows that did not point back to the sections that should govern future work.

## Decision

Add semantic validation on top of section-shape validation:

| Rule | Enforcement intent | Example failure |
| --- | --- | --- |
| Grounded references | Reference columns should cite C3 IDs, refs, rules, ADRs, files, or commands rather than generic prose. | Reference names policy with no entity or file. |
| Evidence must be actionable | Evidence cells should name a command, file path, test, or C3 entity. | Evidence says review after implementation. |
| N.A requires reason | Non-applicable values must use N.A - reason. | N.A alone. |
| No all-N.A rows | A row with every cell non-applicable carries no contract. | Every column says N.A - none. |
| No repeated boilerplate | Duplicate filler across table cells should fail. | Same generic verification copied into unrelated rows. |
| Derived materials must cite governing sections | Future code/docs/tests/prompts must derive from concrete component sections. | Derivation names component docs with no section. |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| Weak token rejection | placeholderPattern rejects TBD, TODO, maybe, optional, later, and if applicable in free text and table cells. | cli/cmd/strict_component_docs.go; cli/cmd/strict_component_docs_test.go |
| N.A contract | validateTextSubstance and validateStrictTableCells require N.A - reason instead of bare N.A. | TestStrictComponentDoc_AllowsNAWithReason; TestStrictComponentDoc_RejectsInvalidNA |
| Enum enforcement | strictColumnEnums restricts Governance Type and Contract Direction to known values or reasoned N.A. | TestStrictComponentDoc_RejectsInvalidDirection |
| Minimum substance | Goal and Purpose have minimum word checks so tiny statements cannot pass. | cli/cmd/strict_component_docs.go; cli/cmd/readwrite_test.go |
| Grounded references | validateStrictTableSemantics requires reference-like cells to cite an entity ID, file path, command, or explicit grounded evidence. | TestStrictComponentDoc_RejectsUngroundedReference |
| Actionable evidence | evidencePattern requires evidence cells to name command/file/entity proof. | cli/cmd/strict_component_docs.go; cli/cmd/check_enhanced_test.go |
| All-N.A row rejection | Table row validation fails when every cell is N.A. | TestStrictComponentDoc_RejectsAllNARow |
| Duplicate boilerplate rejection | Repeated semantic table content is tracked and rejected across cells where it hides missing detail. | cli/cmd/strict_component_docs.go; cli/cmd/strict_component_docs_test.go |
| Derived-material grounding | Derived Materials rows must cite strict sections such as Goal, Governance, Contract, or Change Safety. | TestStrictComponentDoc_RejectsUngroundedDerivation |
| Migration compatibility | strictMigrationComponentBody must emit docs that satisfy the semantic checks or become grouped blockers. | cli/cmd/migrate_v2.go; cli/cmd/migrate_v2_test.go |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| Strict validator | Adds table-cell semantic checks after required section/header checks. | cli/cmd/strict_component_docs.go |
| Check/verify | Reports semantic errors as component validation failures. | cli/cmd/check_enhanced.go |
| Add/write/set | Prevents storing valid-looking but weak component docs. | cli/cmd/add.go; cli/cmd/write.go; cli/cmd/set.go |
| Migration | Converts empty and legacy docs to strict form; blocks unrecoverable weak content before writing. | cli/cmd/migrate_v2.go |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Trust reviewer judgment only | LLM-generated docs would keep accepting shallow tables until humans noticed. |
| Only lint weak tokens | Weak tokens are not the only low-content pattern. |
| Allow plain N.A | Agents cannot tell whether the author considered the field or skipped it. |
| Keep migration permissive | Weak migrated docs would become the new source for generated code and prompts. |

## Migration Impact

Migration must produce semantically valid strict docs for empty or legacy components. If generated content cannot pass these checks, migration should report grouped blockers and avoid partial writes. Existing docs that were shape-valid but semantically weak need targeted edits rather than validator bypasses.

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Validator rejects useful concise docs | Keep checks tied to evidence, references, N.A reasoning, and duplicate boilerplate rather than arbitrary length. | cli/cmd/strict_component_docs_test.go |
| Migration creates repetitive valid-looking rows | Semantic checks run against generated migration output before writes. | cli/cmd/migrate_v2_test.go |
| Future agents forget why strictness exists | This ADR records the quality failure modes and enforcement surfaces directly. | c3x read adr-20260415-strict-component-docs-semantic-tightening --full |
| Underlay predicates drift | Keep this ADR synchronized when strict regexes, enums, row counts, or semantic validators change. | c3x lookup cli/cmd/strict_component_docs.go; go test ./... |

## Verification

| Check | Command or file |
| --- | --- |
| Semantic validator behavior | go test -count=1 ./cmd -run TestStrictComponent |
| Migration recovery and blockers | go test -count=1 ./cmd -run TestRunMigrateV2 |
| C3 canonical health | c3x check --include-adr; c3x verify |
