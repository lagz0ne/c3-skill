---
id: c3-119
c3-seal: 957063d128fe20f55ec06c03b65cf282a1e9a0e0e63df2affe41bbfbf39a144f
title: sync-lifecycle-cmds
type: component
category: feature
parent: c3-1
goal: Handle import, export, sync, repair, migrate, delete, and git guardrail flows around canonical C3 state.
---

# sync-lifecycle-cmds

## Goal

Handle import, export, sync, repair, migrate, delete, and git guardrail flows around canonical C3 state.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own sync-lifecycle-cmds behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep sync-lifecycle-cmds decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |

## Purpose

Provide the lifecycle commands that keep canonical `.c3/` state repairable. These commands should lead within CLI capability: detect batch blockers, explain why execution stopped, and give the next smallest useful repair loop without inventing multi-command chains for edge cases.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before sync-lifecycle-cmds behavior is changed. | c3-1 |
| Inputs | Accept only the files, commands, data, or calls that belong to sync-lifecycle-cmds ownership. | c3-1 |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | c3-1 |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | c3-1 |

## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Human or LLM agent runs import, migrate, repair, export, delete, sync, git guardrail commands, or reaches shared repair through normal command preflight while maintaining canonical .c3/ docs. | c3-1 |
| Primary path | Command inspects current architecture state, performs the smallest deterministic lifecycle operation it can prove, and returns result hints that name follow-up commands only when they are needed. | c3-1 |
| Alternate paths | Normal store-backed commands may call repair automatically before dispatch; explicit migrate and sync export keep their command-owned sequencing. Migration blockers still preflight every component, group all blockers, make no migration writes, and return a scoped repair plan. | adr-20260421-self-healing-preflight; adr-20260415-migrate-repair-command-flow |
| Failure behavior | Failures must include the failed entity, reason, matched bad token where available, writesMade proof, auto-repair context where applicable, and next command loop only when the CLI cannot safely repair on its own. | adr-20260421-self-healing-preflight |

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-1 | policy | Governs sync-lifecycle-cmds behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |
| adr-20260415-migrate-repair-command-flow | adr | Tight migration repair commands and failure output. | Specific repair-flow ADR beats older broad batch-repair wording. | Adds machine blocker reports, repair-plan, cache clear, scoped repair, writesMade proof, matched token, and continue. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| migration input | IN | Reads canonical .c3/ files and disposable cache state; component strictness is checked before migration writes begin. | c3-1 boundary | c3x lookup cli/cmd/migrate_v2.go; go test -count=1 ./cmd -run TestRunMigrateV2_AggregatesStrictComponentBlockers. |
| migration blocker report | OUT | c3x migrate --dry-run --json reports grouped blockers with writesMade:false, issue hints, matched bad-token examples, and next safe commands. | c3-119 boundary | cli/cmd/migrate_v2_test.go TestRunMigrateV2_JSONBlockerReport. |
| migration repair plan | OUT | c3x migrate repair-plan emits exact safe steps and current blockers: scoped repair, cache clear, import, continue, then check/verify. | c3-119 boundary | cli/cmd/migrate_v2_test.go TestRunMigrateRepairPlanGivesSafeLoop. |
| scoped migration repair | IN/OUT | c3x migrate repair <id> --section <name> only repairs sections named by current migration blockers and validates the generated strict result before writing. | c3-119 boundary | cli/cmd/migrate_v2_test.go TestRunMigrateRepairSectionOnlyRepairsCurrentBlockerSection. |
| cache clear | OUT | c3x cache clear deletes only disposable .c3/c3.db* and .c3/.c3.import.tmp.db* files so repair loops do not require manual rm commands. | c3-119 boundary | cli/cmd/cache_test.go TestRunCacheClearRemovesOnlyDisposableCacheFiles. |
| lifecycle output | OUT | Derived docs and cache changes must preserve canonical sync and never require users to infer the next lifecycle command after a failure. | c3-1 boundary | c3x check --include-adr; c3x verify; go test -count=1 ./.... |
| shared repair preflight | OUT | RunRepair remains an explicit lifecycle command and is also reusable by runtime preflight to rebuild cache, reseal canonical docs, and verify before normal command dispatch. | c3-119 repair mechanics; c3-108 chooses when to invoke it. | cli/cmd/repair.go RunRepair; cli/main.go self-healing preflight; go test ./.... |
| scoped verify | OUT | c3x verify excludes adr/*.md by default and accepts repeatable --only selectors for entity IDs, canonical paths, or globs so agents can verify a focused doc set while unrelated docs are in progress. --include-adr restores ADR participation for the selected or full set. | c3-119 lifecycle verification boundary; c3-108 parses and passes the selector; c3-113 scopes schema checks. | cli/cmd/repair.go; cli/cmd/sync.go; cli/cmd/import.go; cli/main_test.go TestRun_VerifyOnlySkipsUnselectedComponentDrift and TestRun_VerifyOnlyReportsSelectedComponentDrift. |

## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Partial repair state | migrate writes some docs before discovering strict blockers. | Compare canonical content hash before/after a blocked migrate smoke. | Run copied .c3 blocked migrate smoke and confirm unchanged SHA256 content hash. |
| Weak failure guidance | A lifecycle command exits with only an error string, a one-entity blocker, or a manual rm instruction. | Inspect command output for grouped entity IDs, reason, writesMade proof, matched token, fix loop, and verification command. | go test -count=1 ./cmd -run TestRunMigrateV2_JSONBlockerReport; go test -count=1 ./cmd -run TestRunMigrateRepairPlanGivesSafeLoop. |
| Contract drift | Import, repair, or migrate changes command order or cache behavior without updating component docs. | Compare changed files against c3-119 Contract and Business Flow. | Run go test -count=1 ./..., c3x check --include-adr, and c3x verify. |
| ADR drift masks release readiness | Default verify skips in-progress ADR canonical drift while branch work continues. | Explicit verify --include-adr reports ADR broken seals, sync drift, and schema issues before completion. | go test . -run TestRun_VerifyIncludeADRReportsADRDrift -count=1; c3x verify --include-adr before final handoff. |
| Scoped verify misses intended doc | An --only selector does not match the entity ID or canonical path shape the user expected. | Selected-drift tests mutate c3-101-auth.md and prove --only c3-101 fails while --only c3-1 ignores it. | go test . -run TestRun_VerifyOnly -count=1; use exact entity IDs or canonical relative paths in automation. |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/cmd/migrate_v2.go | Goal, Business Flow, Contract, and Change Safety sections. | Internal helper names may vary; behavior must keep preflight all-or-nothing blockers, scoped repairs, writesMade proof, matched-token hints, and continue flow. | cli/cmd/migrate_v2.go; go test -count=1 ./cmd -run TestRunMigrateV2_JSONBlockerReport. |
| cli/cmd/migrate_v2_test.go | Contract rows for migration blocker report, repair plan, and scoped repair. | Test fixtures may change; assertions must keep grouped blockers, no-write behavior, matched token, and next-command hints covered. | cli/cmd/migrate_v2_test.go; go test -count=1 ./cmd -run TestRunMigrateRepairSectionOnlyRepairsCurrentBlockerSection. |
| cli/cmd/cache.go and cli/cmd/cache_test.go | Contract cache clear row. | Cache filename patterns may grow only if they remain disposable local cache files. | go test -count=1 ./cmd -run TestRunCacheClear. |
| Skill docs and prompts | Failure guidance and lifecycle command order from Business Flow. | Copy may be shorter, but must not add extra just-in-case command chains. | c3x check --include-adr and c3-119 review. |
