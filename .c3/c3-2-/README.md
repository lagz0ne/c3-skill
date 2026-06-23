---
id: c3-2
c3-seal: dbddc2ca6b61b7d7ad206d6591071ba0973b460c230eec11a3a800c99b801eea
title: Claude Skill
type: container
parent: c3-0
goal: Teach an agent to operate C3 through shared skill instructions, Claude and Codex plugin packaging, and a wrapper that runs the selected C3 runtime.
---

# Claude Skill

## Goal

Teach an agent to operate C3 through shared skill instructions, Claude and Codex plugin packaging, and a wrapper that runs the selected C3 runtime.

## Components

| ID | Name | Category | Status | Goal Contribution |
| --- | --- | --- | --- | --- |
| c3-201 | skill-definition |  | active | Be the skill's entry document — classify the agent's intent, teach the one-model / three-act story, and state the contract every operation reference cites. |
| c3-202 | operation-references |  | active | Provide the per-operation guides SKILL.md routes to — one reference per op that teaches how to run that operation against the frozen facts. |
| c3-203 | cli-wrapper |  | active | Detect the host platform, select a version-pinned full or Linux portable packaged binary, build it from source when available, or delegate no-binary installs to the pinned npm runtime manager, then exec it with the agent's arguments. |

## Responsibilities

Carry the skill definition (SKILL.md: the intent router and the three-act model), the per-operation reference guides, the Claude and Codex plugin manifests, and the platform-detecting wrapper that invokes C3 directly from a full-fat bundled/source-built binary, a Linux portable bundled binary, or indirectly through the pinned npm runtime manager for no-binary installs. Owns no architecture logic — it is the agent-facing teaching and packaging layer over c3-1.

## Complexity Assessment

Medium-low: prose plus plugin metadata and a thin shell wrapper. The intelligence lives in the binary; the skill's job is correct routing, faithful invocation, and packaging that works for full-fat sandbox installs, Linux portable fat installs, and no-binary plugin installs.
