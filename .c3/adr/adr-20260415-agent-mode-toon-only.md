---
id: adr-20260415-agent-mode-toon-only
c3-seal: 059425397256f9e76b767440ad5a83d4c7925df9c7c796eeae6b306febad441f
title: agent-mode-toon-only
type: adr
goal: Make `C3X_MODE=agent` a TOON-only output contract inside the C3 CLI so agent invocations no longer receive JSON unless they run outside agent mode. Keep explicit `--json` available for human/non-agent callers, but let agent mode override legacy JSON serialization paths.
status: proposed
date: "2026-04-15"
---

## Goal

Make `C3X_MODE=agent` a TOON-only output contract inside the C3 CLI so agent invocations no longer receive JSON unless they run outside agent mode. Keep explicit `--json` available for human/non-agent callers, but let agent mode override legacy JSON serialization paths.
