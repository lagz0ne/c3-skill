---
id: adr-20260422-short-lived-socket-write-coordinator
c3-seal: 3feb2d964ae5799db43b41917b18d39b61de26d6471d041355a1eec4d21e758a
title: short-lived-socket-write-coordinator
type: adr
goal: Reduce failed concurrent C3 writes by routing bursty mutating CLI invocations through a short-lived, per-directory Unix socket coordinator that serializes canonical `.c3/` mutations without requiring a long-running daemon.
status: implemented
date: "2026-04-22"
---

## Goal

Reduce failed concurrent C3 writes by routing bursty mutating CLI invocations through a short-lived, per-directory Unix socket coordinator that serializes canonical `.c3/` mutations without requiring a long-running daemon.

## Context

C3 already opens SQLite with WAL and busy timeout, which supports concurrent readers and one writer, but C3 mutating commands also modify canonical markdown files and may run repair/export rollback flows. Parallel agents can invoke mutating commands at the same time, so SQLite WAL alone cannot protect the combined DB plus canonical-file workflow. The target is not true parallel SQLite writes; it is safer parallel command submission with deterministic single-writer execution.

Affected topology is Go CLI runtime support, sync lifecycle commands, docs-state commands, add command, wiring, and store usage. Existing canonical `.c3/` drift may limit final doc verification until the tree is reconciled.

## Decision

Implement a minimal lock-backed short-lived socket write coordinator. A mutating command first tries to forward its request to an existing per-`.c3` Unix socket. If no leader is alive, the command becomes leader by acquiring an atomic per-directory lock, binds the short socket path, runs its own request, accepts queued peer requests for a short idle window, serializes all requests through the existing internal command path, then exits. The direct mutation path remains as fallback and for unsupported platforms.

The coordinator is intentionally not a durable daemon. It exists only during a burst of writes and is guarded by a lock so socket races degrade to direct serialized execution rather than corrupting canonical state.

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Runtime dispatch | Add coordinator hook in cli/main.go before direct mutating execution, with bypass for leader-internal command execution | run() and commandMutatesCanonical tests |
| Coordinator package | Add short socket path hashing, request/response encoding, stale-socket probe, leader lock, idle shutdown, deadlines | New cli/internal/coord tests |
| Mutation safety | Serialize mutating command execution including stdin, cwd, C3X mode, stdout/stderr/exit propagation | Parallel integration test |
| Platform fallback | Keep non-Unix or disabled coordinator behavior on direct path with lock safety | Build/test on current platform; code guards |
| Documentation state | Update C3 docs only if canonical drift permits; otherwise record blocker in task handoff | c3x verify --only or explicit blocker |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| CLI runtime | New coordinator dispatch path around mutating commands | Targeted runtime tests and go test ./... |
| CLI synchronization | Serialized write requests cover export/repair/import paths that touch canonical files | Parallel mutation smoke and sync/verify check |
| Help/hints | No user-facing help change in first slice unless a flag is added | Help tests only if flags/help change |
| C3 docs | Runtime/support and sync-lifecycle docs may need update after implementation | c3x check --include-adr and c3x verify if drift is reconciled |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| Coordinator tests | Prove socket path length, stale socket handling, leader race, timeout behavior | go test ./internal/coord |
| Main integration test | Prove burst mutating commands serialize and preserve canonical sync | Targeted CLI test |
| Existing command tests | Ensure output formats, rollback, and mutation semantics still pass | go test ./... in cli/ |
| Local build | Ensure wrapper uses this checkout binary | bash scripts/build.sh |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Rely only on SQLite WAL | WAL does not protect canonical markdown export/repair races and still permits only one SQLite writer |
| Long-running daemon per directory | More lifecycle surface than needed for current goal; user asked for short-time socket-based coordination |
| File lock only | Correct but does not let burst writers hand requests to one owner; still frequent timeout/failure under agent bursts |
| Shell out from coordinator to c3x | Risks recursion through the coordinator and makes stdin/output/error handling messier than internal dispatch |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Stale socket after crash | Probe before unlink, use short hash path under private runtime dir | Unit test with stale socket file |
| Two leaders race | Atomic lock before bind; loser retries forwarding or direct locked path | Leader race test |
| Recursive forwarding | Explicit internal-execution flag disables coordinator while leader runs queued commands | Runtime test or code assertion |
| Hung leader | Dial/read/write deadlines and short idle timeout | Timeout unit test |
| Output mode drift | Forward argv, stdin, cwd, and relevant environment mode; return stdout/stderr/exit | JSON/TOON/human targeted tests |
| Canonical drift already present | Avoid unrelated doc repair; report blocked verification clearly | Final verification section |

## Verification

| Check | Result |\n| --- | --- |\n| Red parallel mutation test before implementation | Failed before implementation because `cli/internal/coord` did not exist. |\n| `go test ./internal/coord -count=1` | Passed. |\n| Targeted CLI coordinator integration test | Passed: `go test . -run 'TestRun_(ConcurrentMutationsUseShortLivedCoordinator|CoordinatorForwardsPipedInput)' -count=1`. |\n| `go test ./...` from `cli/` | Passed under binbag lease `39e9a43c`. |\n| `bash scripts/build.sh` | Passed under binbag lease `7d502be1`; rebuilt local v9.5.0 wrapper binaries. |\n| C3 focused verification | Pending final `c3x check --include-adr` and focused verify after C3 doc updates. |
