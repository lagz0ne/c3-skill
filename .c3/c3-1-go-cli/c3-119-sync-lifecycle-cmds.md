---
id: c3-119
c3-seal: 7bf2efe76faa3b5a329322fd71dcd9eda39f13bf00cc50a0eb89467b0d210581
title: sync-lifecycle-cmds
type: component
category: feature
parent: c3-1
goal: Handle import, export, sync, repair, migrate, delete, and git guardrail flows around canonical C3 state.
---

# sync-lifecycle-cmds
## Goal

Handle import, export, sync, repair, migrate, delete, and git guardrail flows around canonical C3 state.

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN | Persistence APIs | c3-107 |
| IN | Structured content pipeline | c3-106 |
| IN | Runtime wiring | c3-108 |
| OUT | Canonical sync and recovery flows |  |
## Container Connection

Owns branch-safe movement between canonical files and local cache.
