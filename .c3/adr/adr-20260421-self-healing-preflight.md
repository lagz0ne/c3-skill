---
id: adr-20260421-self-healing-preflight
c3-seal: ac2fd83f4dcaf319ed964fe60a9dedeae5905745a1e1a67079d57c49c75eeee5
title: self-healing-preflight
type: adr
goal: Make ordinary `c3x` operations repair rebuildable canonical/cache drift automatically before dispatch, instead of stopping with a repair instruction when the CLI can fix the state itself.
status: implemented
date: "2026-04-21"
---

## Goal

Make ordinary `c3x` operations repair rebuildable canonical/cache drift automatically before dispatch, instead of stopping with a repair instruction when the CLI can fix the state itself.

## Context

Current preflight runs verification before most commands. When cache or canonical sync is rebuildable, the command can stop early and complain that `.c3/` is not ready even though `repair` or import-style rebuild can recover it. User intent: `repair` may remain explicit, but normal operations should include self-healing and only fail when repair cannot safely prove canonical truth. Affected topology: `c3-108` owns runtime preflight in `cli/main.go`; `c3-119` owns repair/import lifecycle helpers; parent `c3-1` already owns `.c3/` file-system read/write and validation.

## Decision

Move shared read-operation preflight from fail-fast verification to repair-first verification: when canonical docs exist and command is not already a repair/migration/import-style operation, run the existing repair path silently before opening the store. If repair fails, return a grouped error that says auto-repair was attempted and includes the underlying reason. Preserve explicit `repair` as an intentional command, but do not require it for recoverable cache or seal drift.

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Runtime preflight | Update cli/main.go shared canonical readiness branch to invoke repair/import verification before command dispatch. | cli/main.go diff and targeted tests |
| Lifecycle helper reuse | Reuse existing cmd.RunRepair/verification helpers rather than duplicate repair logic in runtime. | cli/cmd/repair.go call path |
| Regression tests | Add failing tests that a normal command recovers from missing/stale cache and does not emit repair-only failure. | cli/main_test.go targeted red/green |
| Parent Delta | No parent table change expected: c3-1 already owns .c3/ reads/writes and validation; behavior remains within c3-108 and c3-119. | c3x read c3-1; c3x graph c3-1 --depth 1 |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| Runtime support | Normal command preflight self-heals rebuildable canonical/cache drift before dispatch. | go test ./... |
| Sync lifecycle | Existing repair path becomes shared automatic recovery, not just explicit command behavior. | temp smoke with broken cache/canonical state |
| User workflow | repair remains available, but recoverable cases no longer require user handoff. | live c3x list/query smoke after drift |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| Unit tests | Normal command succeeds after recoverable drift. | targeted go test |
| Full Go suite | Existing explicit repair, verify, import, migrate flows stay green. | go test ./... |
| C3 validation | Canonical tree remains in sync after implementation. | c3x check --include-adr; c3x verify |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Keep explicit repair-only workflow | It creates unnecessary user handoff when CLI already knows the safe recovery path. |
| Auto-repair only missing cache | Too narrow; stale/broken rebuildable cache and seal drift are same user-facing class. |
| Hide all repair failures | Bad: unrecoverable canonical conflicts must still stop with exact cause. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Auto-repair masks real doc corruption | Only proceed if existing repair/verify succeeds; otherwise surface underlying error. | broken canonical negative test |
| Mutating command partial state | Keep existing mutation snapshot and skip preflight repair for command families that already own repair/migration. | full Go suite |
| Extra work on every command | Repair path can be optimized by existing verification/import checks; correctness wins first. | smoke timing acceptable |

## Verification

| Check | Result |
| --- | --- |
| Red regression | pending |
| Targeted green | pending |
| Full Go suite | pending |
| Live temp smoke | pending |
| C3 check/verify | pending |
