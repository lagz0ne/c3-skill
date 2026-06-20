---
id: c3-2
c3-seal: 18f4d279702b8ebefa8ece4877c87be95effdd39b229700fe664595af90e0c20
title: Claude Skill
type: container
parent: c3-0
goal: Teach an agent to operate C3 — route intent to the right operation and run each one through the local CLI binary.
---

# Claude Skill

## Goal

Teach an agent to operate C3 — route intent to the right operation and run each one through the local CLI binary.

## Components

| ID | Name | Category | Status | Goal Contribution |
| --- | --- | --- | --- | --- |
| c3-201 | skill-definition |  | active | Be the skill's entry document — classify the agent's intent, teach the one-model / three-act story, and state the contract every operation reference cites. |
| c3-202 | operation-references |  | active | Provide the per-operation guides SKILL.md routes to — one reference per op that teaches how to run that operation against the frozen facts. |
| c3-203 | cli-wrapper |  | active | Detect the host platform, select the version-pinned packaged binary, build it from source only when absent, and exec it with the agent's arguments. |

## Responsibilities

Carry the skill definition (SKILL.md: the intent router and the three-act model), the per-operation reference guides, and the platform-detecting wrapper that invokes the Go binary. Owns no architecture logic — it is the agent-facing teaching layer over c3-1.

## Complexity Assessment

Low: prose plus a thin shell wrapper. The intelligence lives in the binary; the skill's job is correct routing and faithful invocation.
