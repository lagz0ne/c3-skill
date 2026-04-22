---
id: adr-20260415-cascade-help-hints
c3-seal: 5a0a8c450434f86bec5d4b86bb5c94bf9437aa6001a5a34d0326ce10f4e00270
title: cascade-help-hints
type: adr
goal: Inject result-aware cascade steering into c3x agent-mode help hints so component-level changes force parent/container delta decisions before completion.
status: implemented
date: "2026-04-15"
affects:
    - c3-1
    - c3-108
    - c3-111
    - c3-113
    - c3-114
    - c3-117
    - c3-118
    - c3-2
    - c3-210
---

# Cascade Help Hints

## Goal

Inject result-aware cascade steering into c3x agent-mode help hints so component-level changes force parent/container delta decisions before completion.

## Work Breakdown

- Add JSON-safe help hints for lookup/read/check/diff agent outputs.
- Add text/mermaid help[] steering for add/write/set/graph success paths.
- Make component hints parent-first: read/graph container before component detail.
- Make new component hints explicitly top-down: update container Components/Responsibilities before component detail is considered done.
- Tighten skill change flow with a Contract Cascade Gate and mandatory Parent Delta evidence.

## Parent Delta

updated: c3-1 Go CLI responsibility remains valid; command outputs now steer architecture integration. c3-2 Claude Skill operation guidance now requires top-down parent context and Parent Delta evidence. No new component added.

## Verification

- Red tests first: cascade hint tests failed across lookup/read/add/write/set/graph/check/diff.
- Green tests: targeted cascade hint suite passed.
- Full Go suite: go test ./... passed in cli/.
- Smoke: C3X_MODE=agent go run . --c3-dir ../.c3 lookup cli/cmd/lookup.go emits parent-first help for c3-1/c3-114.
- Smoke: C3X_MODE=agent go run . --c3-dir ../.c3 read c3-114 emits parent-first help.
- Smoke: C3X_MODE=agent go run . --c3-dir ../.c3 check --include-adr emits cascade review help and zero issues.
