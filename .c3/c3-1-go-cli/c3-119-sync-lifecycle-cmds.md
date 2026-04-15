---
id: c3-119
c3-seal: 11ee05ffdfcdeb2f36aec84e920b89dbf2b60696642fcb1d301a2c4e4f1f352a
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
| Actor / caller | Human or LLM agent runs import, migrate, repair, export, delete, sync, or git guardrail commands while maintaining canonical .c3/ docs. | c3-1 |
| Primary path | Command inspects current architecture state, performs the smallest deterministic lifecycle operation it can prove, and returns result hints that name follow-up commands only when they are needed. | c3-1 |
| Alternate paths | When strict component migration would fail, preflight every component, group all blockers, make no migration writes, and return a repair plan with inspect/fix/verify steps. | adr-20260415-migrate-batch-repair-plan |
| Failure behavior | Failures must include the failed entity, reason, and next command loop so users and LLM agents do not guess whether to import, migrate, check, repair, or verify next. | adr-20260415-migrate-batch-repair-plan |
## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-1 | policy | Governs sync-lifecycle-cmds behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |
## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| migration input | IN | Reads canonical .c3/ files and disposable cache state; component strictness is checked before migration writes begin. | c3-1 boundary | c3x lookup cli/cmd/migrate_v2.go; go test -count=1 ./cmd -run TestRunMigrateV2_AggregatesStrictComponentBlockers. |
| migration blocker plan | OUT | Reports every strict component blocker discovered in preflight, explains all-or-nothing behavior, and gives inspect/import/migrate/check/verify commands. | c3-119 boundary | cli/cmd/migrate_v2_test.go; TestRunMigrateV2_AggregatesStrictComponentBlockers; copied Kanna .c3 smoke. |
| lifecycle output | OUT | Derived docs and cache changes must preserve canonical sync and never require users to infer the next lifecycle command after a failure. | c3-1 boundary | c3x check --include-adr; c3x verify; go test -count=1 ./.... |
## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Partial repair state | migrate writes some docs before discovering strict blockers. | Compare canonical content hash before/after a blocked migrate smoke. | Run copied .c3 blocked migrate smoke and confirm unchanged SHA256 content hash. |
| Weak failure guidance | A lifecycle command exits with only an error string or a one-entity blocker. | Inspect command output for grouped entity IDs, reason, fix loop, and verification command. | go test -count=1 ./cmd -run TestRunMigrateV2_AggregatesStrictComponentBlockers. |
| Contract drift | Import, repair, or migrate changes command order or cache behavior without updating component docs. | Compare changed files against c3-119 Contract and Business Flow. | Run go test -count=1 ./..., c3x check --include-adr, and c3x verify. |
## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/cmd/migrate_v2.go | Goal, Business Flow, Contract, and Change Safety sections. | Internal helper names may vary; behavior must keep preflight all-or-nothing blockers and actionable failure hints. | Targeted migrate blocker test and copied legacy .c3 smoke. |
| cli/cmd/migrate_v2_test.go | Contract rows for migration input and blocker plan. | Test fixtures may change; assertions must keep grouped blockers, no-write behavior, and next-command hints covered. | go test -count=1 ./cmd -run TestRunMigrateV2_AggregatesStrictComponentBlockers. |
| Skill docs and prompts | Failure guidance and lifecycle command order from Business Flow. | Copy may be shorter, but must not add extra just-in-case command chains. | c3x check --include-adr and c3-119 review. |
