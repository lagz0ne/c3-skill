---
id: c3-106
c3-seal: af388b3fd8051177d66d7ad866af15626823ec6a8050cd1a092ecc491fbf12ac
title: content-lib
type: component
category: foundation
parent: c3-1
goal: Parse, render, and bridge structured markdown content to node-tree storage for C3 entities.
---

# content-lib
## Goal

Parse, render, and bridge structured markdown content to node-tree storage for C3 entities.

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN | Entity node persistence | c3-107 |
| OUT | Markdown parse/render pipeline |  |
## Container Connection

Provides document body semantics used by read, write, import, and export flows.
