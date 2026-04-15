---
id: adr-20260415-force-agent-toon-output
c3-seal: 10c53cffab82dd6ade217e098555c127da16563b1c83e26a45e454af3bf57ffd
title: force-agent-toon-output
type: adr
goal: Ensure every c3x command that emits structured output honors agent mode by returning TOON or plain actionable text, not raw JSON, with systematic tests for commands such as add and other direct encoder paths.
status: implemented
date: "2026-04-15"
---

# force-agent-toon-output
## Goal

Ensure every c3x command that emits structured output honors agent mode by returning TOON or plain actionable text, not raw JSON, with systematic tests for commands such as add and other direct encoder paths.
