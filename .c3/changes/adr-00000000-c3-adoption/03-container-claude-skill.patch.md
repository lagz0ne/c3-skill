---
target: c3-2
scope: whole
type: container
parent: c3-0
title: Claude Skill
---
# Claude Skill

## Goal

Teach an agent to operate C3 — route intent to the right operation and run each one through the local CLI binary.

## Components

| ID | Name | Category | Status | Goal Contribution |
| --- | --- | --- | --- | --- |

## Responsibilities

Carry the skill definition (SKILL.md: the intent router and the three-act model), the per-operation reference guides, and the platform-detecting wrapper that invokes the Go binary. Owns no architecture logic — it is the agent-facing teaching layer over c3-1.

## Complexity Assessment

Low: prose plus a thin shell wrapper. The intelligence lives in the binary; the skill's job is correct routing and faithful invocation.
