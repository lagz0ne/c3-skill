---
id: adr-20260415-force-agent-toon-output
c3-seal: 04311ca7ba2d3853efe9be7522c8fe8ca8ec01216d8e7b8a66b8a069a060040a
title: force-agent-toon-output
type: adr
goal: Ensure every c3x command that emits structured output honors agent mode by returning TOON or plain actionable text, not raw JSON, with systematic tests for commands such as add and other direct encoder paths.
status: implemented
date: "2026-04-15"
---

# force-agent-toon-output

## Goal

Ensure every c3x command that emits structured output honors agent mode by returning TOON or plain actionable text, not raw JSON, with systematic tests for commands such as add and other direct encoder paths.
