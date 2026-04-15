---
id: adr-20260415-mutation-preverify-repair-bypass
c3-seal: 3ed645e2a28811a43414ac1960f7adaa4a5308f7319752036c57c826e6acdc87
title: mutation-preverify-repair-bypass
type: adr
goal: Allow mutating C3 commands such as add, write, set, wire, delete, codemap, and migrate to reach their own payload validation and canonical export paths even when existing canonical docs fail preverification, while keeping read-only commands gated by verify.
status: implemented
date: "2026-04-15"
---

# mutation-preverify-repair-bypass
## Goal

Allow mutating C3 commands such as add, write, set, wire, delete, codemap, and migrate to reach their own payload validation and canonical export paths even when existing canonical docs fail preverification, while keeping read-only commands gated by verify.
