---
id: c3-201
c3-version: 4
title: skill-router
type: component
category: foundation
parent: c3-2
goal: Classify user intent into one of seven operations (onboard/query/audit/change/ref/rule/sweep) and dispatch to the correct reference doc
summary: SKILL.md entry point — the only file Claude Code loads; must fit triggering constraints (≤1024 chars description)
uses: []
---

# skill-router

## Goal

Classify user intent into one of seven operations and dispatch to the correct operation reference.

## Container Connection

Without this, no operation runs. All natural language c3 requests go through SKILL.md first.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| OUT (provides) | Classified intent + dispatched operation |  |

## Code References

| File | Purpose |
|------|---------|
| `skills/c3/SKILL.md` | Intent classification table + dispatch rules |
