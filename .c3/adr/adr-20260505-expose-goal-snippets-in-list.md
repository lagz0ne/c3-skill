---
id: adr-20260505-expose-goal-snippets-in-list
c3-seal: 59aa8e65717533fecfc82d2ab3c3e7e9c684bd4a04431934ca11439d865ad226
title: expose-goal-snippets-in-list
type: adr
goal: Expose bounded goal snippets in list output so agents can navigate C3 topology without extra reads.
status: implemented
date: "2026-05-05"
---

## Goal

Expose bounded goal snippets in list output so agents can navigate C3 topology without extra reads.

## Context

Current agent-mode `c3x list` output shows only `id`, `type`, `title`, `parent`, and `status`, so agents can identify nodes but cannot see the durable reason each node exists. Users must run extra reads to distinguish similarly named components or refs. The affected topology is the Go CLI list command and its agent-facing TOON output contract.

## Decision

Add a short `goal` field to structured list rows, with truncation applied before output so navigation gains purpose context while TOON output stays bounded. Keep the change inside list-cmd presentation logic and reuse existing entity goal data instead of reading document bodies.

## Affected Topology

| Entity | Type | Why affected | Governance review |
| --- | --- | --- | --- |
| c3-1 | container | Parent owns all c3x commands and CLI presentation contracts. | Review parent responsibilities and record parent delta evidence. |
| c3-112 | component | list-cmd owns topology output and is the direct implementation target. | Verify list-cmd tests and real c3x list output. |
| c3-108 | component | runtime-support owns shared agent TOON output helpers consumed by list-cmd. | Review only if shared output helper behavior changes. |

## Compliance Refs

| Ref | Why required | Action |
| --- | --- | --- |
| N.A - no cited refs govern c3-112 beyond its existing c3-101/c3-102 dependencies. | N.A - this change only reshapes list output from existing entity metadata. | N.A - no new ref required. |

## Compliance Rules

| Rule | Why required | Action |
| --- | --- | --- |
| N.A - lookup found no rule entities attached to c3-112. | N.A - no rule-specific golden pattern applies. | N.A - no new rule required. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Tests | Add focused list-cmd expectation that agent TOON output includes a bounded goal column. | Failing focused test before implementation, then passing after implementation. |
| CLI output | Update list structured output row shape to include goal snippets. | c3x list shows goal in agent mode. |
| Docs decision | Record parent delta decision after implementation. | ADR note or final status includes no-delta/update evidence. |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| cli/cmd/list.go | Include bounded entity goal text in structured list output rows. | Focused list test and real local c3x list. |
| cli/cmd/list_test.go | Cover agent TOON list goal field behavior. | go test ./cmd -run TestRunList_TOONOutput. |
| Shared output helpers | N.A - keep TOON helper contract unchanged unless tests reveal a need. | go test ./... proves no shared regression. |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| list-cmd unit test | Fails if agent TOON list omits goal field. | Focused Go test. |
| full CLI test suite | Catches presentation regressions across command parsing and output helpers. | go test ./... from cli/. |
| real c3local list smoke | Confirms built local wrapper emits goal in this repository. | C3X_MODE=agent bash skills/c3/bin/c3x.sh list. |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Keep list minimal and require read for purpose. | This preserves token minimalism but hurts navigation because agents cannot triage topology candidates from list alone. |
| Add full goal text without truncation. | Some goals are long and would bloat the topology view, weakening list as a quick navigation primitive. |
| Add goal only to human output. | The user pain is agent navigation, so agent TOON output needs the bounded content. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Agent output becomes too verbose. | Truncate goal snippets before serialization. | Unit test checks long goal is shortened. |
| Existing tests expect the old TOON header. | Update only intentional list tests, leave shared output contract unchanged. | go test ./.... |
| Parent docs drift because list-cmd behavior changed. | Review parent responsibilities and record no-delta or update decision. | c3x check --include-adr plus parent delta note. |

## Verification

| Check | Result |
| --- | --- |
| cd cli && go test ./cmd -run TestRunList_TOONOutput | Passed; focused TOON list test confirms goal column and truncation. |
| cd cli && go test ./... | Passed; full CLI Go test suite has no regression. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh list | Passed; local wrapper output shows entities[36]{id,type,title,goal,parent,status} with clipped goal snippets. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check | Passed; structural C3 docs check reports no issues. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr --only adr-20260505-expose-goal-snippets-in-list | Passed; focused ADR contract check reports no issues. |
