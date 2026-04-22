---
id: adr-20260416-uplevel-strict-adrs
c3-seal: e629eeffc55ca995371334da9e109870253fc6c7b20c04b29c0989a86ec8229c
title: uplevel-strict-adrs
type: adr
goal: Up-level strict documentation ADRs so they preserve design intent, enforcement details, migration behavior, risks, and verification evidence instead of acting as thin change markers.
status: implemented
date: "2026-04-16"
affects:
    - adr-20260415-strict-component-docs-semantic-tightening
    - adr-20260415-strict-enriched-component-docs
    - c3-103
    - c3-108
    - c3-113
    - c3-117
    - c3-2
    - c3-201
    - c3-214
---

# uplevel-strict-adrs

## Goal

Up-level strict documentation ADRs so they preserve design intent, enforcement details, migration behavior, risks, and verification evidence instead of acting as thin change markers.

## Context

Strict component documentation now drives future code, derived docs, migration repair, and reviewer expectations. The two strict ADRs existed, but each was mostly a goal statement. That left future agents without the decision matrix, rejected alternatives, enforcement surfaces, migration consequences, proof commands, and exact underlay C3 files behind the strict documentation rules.

The follow-up constraint is stricter: `c3x` must be the single source of C3 enforcement. Skill prose may classify intent and reference workflows, but enforceable steps, ADR shape, hints, help, repair loops, and failure guidance must be discoverable through CLI commands.

## Decision

Use ADRs as durable design ledgers for strict documentation changes, but make the CLI own the ADR shape and workflow prompts. `c3x schema adr`, `c3x add adr`, command `help[]`, `c3x read <adr> --full`, `c3x write`, `c3x check --include-adr`, and `c3x verify` are the authoritative flow. The skill points to those commands and does not duplicate the ADR checklist as a second policy source.

Strict-related ADRs should record the decision matrix, enforcement surfaces, underlay C3 changes, migration impact, rejected softer alternatives, risks, and verification evidence. The ADR should not duplicate every component contract; it should preserve why the contract exists and how the tool enforces it.

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Existing strict ADRs | Expanded the strict enriched docs ADR and semantic tightening ADR with context, decision, underlay C3 changes, enforcement, alternatives, migration impact, risks, and verification sections. | adr-20260415-strict-enriched-component-docs; adr-20260415-strict-component-docs-semantic-tightening |
| ADR schema | Added CLI-owned ADR sections for Context, Decision, Work Breakdown, Underlay C3 Changes, Enforcement Surfaces, Alternatives Considered, Risks, and Verification. | cli/internal/schema/schema.go; c3x schema adr |
| ADR template | Added matching embedded ADR headings and table headers so newly created ADR files start with the durable ledger shape. | cli/internal/templates/adr.md; cli/internal/templates/templates_test.go |
| CLI help and hints | Added add-help ADR workflow and ADR cascade hints that point to c3x schema/read/write/check/verify. | cli/cmd/help.go; cli/cmd/cascade_hints.go |
| Schema presentation | Human schema output now prints section purpose text so c3x gives content guidance, not just headings. | cli/cmd/schema.go; cli/cmd/schema_test.go |
| Skill boundary | Marked the skill as reference routing and moved enforcement authority to c3x output. | skills/c3/SKILL.md; skills/c3/references/change.md |
| C3 component contracts | Updated CLI and skill components touched by the underlay change. | c3-103; c3-108; c3-113; c3-117; c3-201; c3-214; c3-2 |
| Build reproducibility | Made scripts/build.sh force CGO_ENABLED=0 for all targets after local rebuild exposed a cgo -m64 failure. | scripts/build.sh; ref-cross-compiled-binary |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| Schema registry | adr schema gained Context, Decision, Work Breakdown, Underlay C3 Changes, Enforcement Surfaces, Alternatives Considered, Risks, and Verification sections with purposes and table columns. | cli/internal/schema/schema.go; TestRunSchema_ADRIncludesDecisionLedger; TestRunSchema_JSON_ADRUnderlayColumns |
| Schema command | RunSchema text output now prints section purpose lines so the CLI explains what to put in each section. | cli/cmd/schema.go; go test ./cmd -run TestRunSchema_ADRIncludesDecisionLedger |
| ADR template | Embedded adr.md now includes durable ledger headings and table headers. | cli/internal/templates/adr.md; TestRead_ADRTemplateIncludesDecisionLedger |
| Add help | c3x add --help now includes the CLI-owned ADR workflow and removes unsupported ADR --goal example. | cli/cmd/help.go; TestShowHelp_AddADRWorkflowPointsAtSchema |
| ADR hints | ADR add/read/write result hints route to c3x schema adr, c3x read, c3x write, and c3x check --include-adr && c3x verify. | cli/cmd/cascade_hints.go; TestRunAdd_AdrAgentHintsUseCLISchema |
| Skill reference | Skill top-level and change reference now state that c3x owns enforcement, schemas, help, hints, repair steps, and verification. | skills/c3/SKILL.md; skills/c3/references/change.md |
| C3 docs | Updated component contracts and derived materials for schema, template, runtime help/hints, docs-state, skill-router, change-operation, and the Claude Skill container. | c3x read c3-103/c3-108/c3-113/c3-117/c3-201/c3-214/c3-2 --full |
| Build wrapper | scripts/build.sh now sets CGO_ENABLED=0 for each cross-compile target so local rebuilds do not depend on host cgo compiler flags. | bash scripts/build.sh passed after the script change; ref-cross-compiled-binary How section updated. |
| Mutation rollback | Dispatcher now snapshots .c3 before mutating commands and restores it when command handling, database close, or canonical export fails, making add adr all-or-nothing across cache and canonical markdown. | cli/main.go mutationSnapshot; cli/main_test.go TestRun_AddADRRollsBackWhenCanonicalExportFails. |
| ADR creation completeness | c3x add adr now rejects thin ADR bodies and requires all ADR schema sections, table rows, and table columns before inserting the ADR entity. | cli/cmd/add.go validateADRCreationBody; cli/cmd/add_test.go TestRunAdd_AdrRequiresCompleteBody. |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| c3x schema adr | Shows ADR sections, purpose hints, required status, and table columns from CLI schema. | cli/cmd/schema.go; cli/internal/schema/schema.go |
| c3x add adr <slug> | Creates ADR work order from stdin and emits agent hints to inspect, fill, and verify the ADR through c3x. | cli/cmd/add.go; cli/cmd/cascade_hints.go |
| c3x add --help | Teaches ADR creation through CLI schema plus stdin body, not skill-local placeholders or unsupported flags. | cli/cmd/help.go |
| c3x write <adr> / c3x write --section | Keeps ADR repair and fill operations inside the CLI mutation path. | cli/cmd/write.go; cli/cmd/help.go |
| c3x check --include-adr and c3x verify | Validates ADR-inclusive docs and canonical sync before completion. | c3-214 Contract Verification close row |
| Skill references | Route agents to c3x commands and CLI output instead of duplicating checklists. | skills/c3/SKILL.md; skills/c3/references/change.md |
| mutating dispatcher rollback | Failed mutating commands restore the pre-command .c3 tree, including c3.db and canonical markdown, before returning the error. | cli/main.go; TestRun_AddADRRollsBackWhenCanonicalExportFails. |
| c3x add adr completeness gate | Creation fails before insert when any ADR ledger section, required table row, or required table column is missing. | TestRunAdd_AdrRequiresCompleteBody. |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Keep ADR checklist only in skill reference | Creates two sources of truth and lets CLI-created ADRs remain thin when the skill is not loaded or ignored. |
| Make ADR sections mandatory for every ADR | Too blunt for tiny decisions; strictness should come from workflow hints, review, and --include-adr validation until concrete semantic ADR validation is introduced. |
| Only improve ADR template | Better starting file but weak repair loop; agents need schema/help/hints/failure guidance from the CLI after creation too. |
| Put detailed prose in command help only | Help is discoverable, but schema output is the durable content contract used by add/set/write/check flows. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| ADRs become verbose boilerplate | CLI schema asks for concrete underlay changes, enforcement surfaces, alternatives, risks, and verification tied to changed behavior. | c3x schema adr; c3x check --include-adr |
| Agents duplicate component docs inside ADRs | ADR records why and enforcement; component docs keep ownership and contract. | c3x read c3-214 --full; c3x read affected ADRs --full |
| Skill prose becomes a competing policy engine | Skill states c3x is enforcement source and references CLI schema/help/hints. | rg "Enforcement source" skills/c3/SKILL.md; rg "The CLI is the source of truth" skills/c3/references/change.md |
| Underlay details become stale | ADRs and component docs name exact files and tests so future schema/help/hint edits have obvious doc targets. | c3x lookup cli/cmd/schema.go; c3x lookup cli/cmd/cascade_hints.go; go test ./... |
| Agent mode leaks JSON again | New tests assert TOON-style ADR hints use help[n]; shared output tests still cover agent mode serialization. | go test ./cmd; rg json.NewEncoder cli --glob '!**/*_test.go' |
| add adr partially creates work order | Mutating dispatcher rollback restores cache and canonical files when export fails after add succeeds internally. | go test -count=1 . -run TestRun_AddADRRollsBackWhenCanonicalExportFails. |
| Thin ADR exists before decision detail | c3x add adr validates complete decision-ledger sections before insert instead of relying on later incremental fill. | go test ./cmd -run TestRunAdd_AdrRequiresCompleteBody. |

## Verification

| Check | Result |
| --- | --- |
| Red targeted tests | New ADR schema/hint/help/template tests failed before schema purpose output and test expectation fixes. |
| Green targeted tests | go test ./cmd ./internal/templates -run TestRunSchema_ADR |
| ADR complete creation gate | go test ./cmd -run TestRunAdd_AdrRequiresCompleteBody passed; temp-project thin ADR smoke failed before creation with all-or-nothing hints. |
| ADR creation rollback | go test -count=1 . -run TestRun_AddADRRollsBackWhenCanonicalExportFails passed, proving failed add adr does not leave an ADR file or cache entity. |
| Full Go tests | go test ./... in cli passed. |
| Build script | bash scripts/build.sh passed with CGO_ENABLED=0 enforced by the script. |
| C3 validation | C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr and c3x verify passed after complete-creation doc updates. |
