---
id: adr-20260421-partial-verify-skip-adr
c3-seal: 8f202bc5a3fa7ff8b48a4d28c2d157345a646eedeb7474333653903629297010
title: partial-verify-skip-adr
type: adr
goal: Allow normal c3x verify to validate active C3 source truth while ignoring in-progress ADR canonical drift unless the caller explicitly asks for ADR validation.
status: implemented
date: "2026-04-21"
---

## Goal

Allow normal c3x verify to validate active C3 source truth while ignoring in-progress ADR canonical drift unless the caller explicitly asks for ADR validation.

## Context

ADRs are work-order and historical records hidden from list, query, and check by default. Teams working on the same branch can have an ADR under active edit while still needing c3x verify to prove non-ADR C3 docs, cache rebuild, and canonical sync. Current verify already excludes ADR schema validation through RunCheckV2, but seal and sync checks still scan adr/*.md, so an in-progress ADR can block unrelated branch work.

## Decision

Make c3x verify partial by default for ADRs and selectable for any canonical doc set. Default verify skips ADR files for seal and sync drift, while c3x verify --include-adr preserves full ADR seal, sync, and schema validation. Add repeatable c3x verify --only <id-or-path-or-glob> so agents can verify a focused set of docs by entity ID, canonical path, file name, or path glob. Thread --include-adr and --only through VerifyOptions, RepairOptions, lifecycle sync checks, missing-cache import rebuilds, and check validation instead of inventing a second command.

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Runtime option path | Pass parsed IncludeADR and repeatable Only selectors into verify dispatch and preflight verify calls. | cli/main.go; cli/cmd/options.go; cli/cmd/options_test.go. |
| Lifecycle verification | Filter adr/*.md from default seal/sync snapshots unless IncludeADR is true, and filter snapshots to --only selectors when supplied. | cli/cmd/repair.go; cli/cmd/sync.go. |
| Missing-cache rebuild | Let default/scoped verify rebuild cache while tolerating drift outside the selected verification set. Direct import remains strict. | cli/cmd/import.go; TestRun_VerifySkipsADRDriftByDefault. |
| Scoped check | Scope schema, relationship, code-map, and layer warnings to selected entities during verify --only so unrelated docs do not block focused verification. | cli/cmd/check_enhanced.go; TestRun_VerifyOnlySkipsUnselectedComponentDrift. |
| Help and tests | Document default partial verify, --include-adr, and repeatable --only; add red/green tests for ADR drift and non-ADR selected/unselected drift. | cli/cmd/help.go; cli/main_test.go; cli/cmd/help_test.go. |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| c3-119 sync-lifecycle-cmds | Verify and sync-check ADR/default filtering plus repeatable --only path/entity filtering become explicit lifecycle behavior; parent c3-1 unchanged because lifecycle command ownership already includes verify/repair/sync. | c3x lookup cli/cmd/repair.go; c3x read c3-119; go test targeted verify tests. |
| c3-108 runtime-support | Existing IncludeADR parser flag is reused and new Only selectors are parsed/passed for verify dispatch, with no new command. | c3x lookup cli/main.go; c3x read c3-108; go test ./cmd -run TestParseArgs. |
| c3-113 check-cmd | Check validation accepts selected entities during verify --only and filters global issue sources to selected docs. | c3x lookup cli/cmd/check_enhanced.go; c3x read c3-113; go test targeted verify-only tests. |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| c3x verify | Skips ADR seal and sync drift by default while still checking non-ADR canonical docs. | TestRun_VerifySkipsADRDriftByDefault. |
| c3x verify --include-adr | Fails on broken ADR seal or ADR sync drift and validates ADR schema. | TestRun_VerifyIncludeADRReportsADRDrift. |
| c3x verify --only <selector> | Limits seal, sync, missing-cache import tolerance, and schema checks to selected entity IDs or canonical paths. | TestRun_VerifyOnlySkipsUnselectedComponentDrift; TestRun_VerifyOnlyReportsSelectedComponentDrift. |
| c3x check --include-adr | Remains the explicit ADR detail validation path for full-tree ADR work. | Existing RunCheckV2 IncludeADR behavior. |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Add a new verify-partial command | Duplicates verify semantics and makes agent guidance harder. |
| Ignore ADRs in all verify modes | Removes the current full validation path needed before finishing ADR work. |
| Keep current full verify only | Blocks same-branch collaboration whenever ADR files are mid-edit. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| ADR problems get missed before final handoff | Keep --include-adr and help text for ADR work completion. | c3x verify --include-adr targeted smoke. |
| Non-selected doc drift hides intended work | Make --only exact by entity ID, canonical path, file name, or glob; selected-drift tests prove the positive and negative cases. | TestRun_VerifyOnlySkipsUnselectedComponentDrift; TestRun_VerifyOnlyReportsSelectedComponentDrift. |
| Non-ADR drift accidentally filtered in default mode | Default filtering only skips canonical paths under adr/ with .md suffix unless --only is explicitly supplied. | Targeted tests mutate ADR and component docs separately. |
| Preflight auto-repair hides selected edits | Pass IncludeADR and Only through normal preflight so selected drift still blocks, while unrelated drift can be ignored. | go test targeted preflight/verify tests. |

## Verification

| Check | Result |
| --- | --- |
| Red scoped verify-only tests | Failed before implementation because --only c3-1 still reported BROKEN_SEAL for c3-101-auth.md; passed after scoped path/entity filtering. |
| Targeted scoped and ADR verify tests | Passed; proves selected non-ADR drift fails, unselected non-ADR drift passes, default ADR drift passes, and --include-adr ADR drift fails. |
| Parser and verify help tests | Passed; proves repeatable --only parsing and verify help surface. |
| Check, parser, and help hint tests | Passed after scoped check matcher simplification and scoped verify hint updates. |
| Full Go suite | Passed under binbag lease bf25950a after final matcher simplification. |
| Local wrapper build | Passed under binbag lease 556456da; rebuilt local c3x-9.4.4 binaries. |
| Default local c3 verify | Exited 0 with zero issues and canonical sync OK. |
| Focused local c3 verify | c3x verify --only c3-119 exited 0 with zero issues and canonical sync OK. |
| Full ADR local c3 verify | Exited 0; reported existing historical ADR section warnings unrelated to this change; canonical sync OK. |
| Full ADR local c3 check | Exited 0; reported existing historical ADR section warnings unrelated to this change. |
