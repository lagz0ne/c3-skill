---
name: c3
description: |
  Use when working with C3 architecture docs - navigating, understanding, designing, or auditing.
  Triggers: "C3", "architecture", "where is X documented", "impact of changing X".
---

# C3 Architecture Assistant

## ADR Lifecycle

```
proposed → accepted → implemented
```

Create as `proposed` → human accepts → update layer docs → implement → audit passes → mark `implemented`.

## Mode Selection

| Intent | Mode | Reference |
|--------|------|-----------|
| "Where is X?" | Navigate | `references/lookup-patterns.md` |
| "Add/change X" | Design | `references/adr-template.md` |
| "Accept/implement ADR" | Lifecycle | - |
| "Audit C3" | Audit | `references/audit-checks.md` |
| No `.c3/` exists | Adopt | `references/adopt-workflow.md` |

## Layer Reference

| Layer | Reference |
|-------|-----------|
| Context/Container | `references/structure-guide.md` |
| Component | `references/implementation-guide.md` |

## Mode: Navigate

ID-based lookup. Parse ID (`c3-0`, `c3-N`, `c3-NNN`, `adr-YYYYMMDD-slug`) and read the corresponding file.

See `references/lookup-patterns.md` for ID resolution.

## Mode: Adopt

> See `references/adopt-workflow.md`

Discovery subagents scan codebase → confirm containers → confirm components → create `.c3/` with inventories.

**CRITICAL:** Inventory tables only, NOT component docs.

## Mode: Design

Discover (what/why/where) → Assess (read docs, identify layers) → ADR Decision → Create (`references/adr-template.md`) → Handoff

## Mode: Lifecycle

**proposed → accepted:** Update status, parse "Changes Across Layers", update layer docs, record in Audit Record.

**accepted → implemented:** Run audit first. PASS: update status. FAIL: report issues.

## Mode: Audit

> See `references/audit-checks.md`

Scopes: `audit C3` | `audit container c3-1` | `audit adr adr-YYYYMMDD-slug`

## Guidelines

- Never skip ADR lifecycle states
- Update layer docs after accepting ADR
- Never mark implemented without audit passing
