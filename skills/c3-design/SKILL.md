---
name: c3-design
description: Use when designing, updating, or exploring system architecture with C3 methodology - iterative scoping through hypothesis, exploration, and discovery across Context/Container/Component layers
---

# C3 Architecture Design

## Overview

Transform requirements into C3 documentation through iterative scoping. Also supports exploration-only mode.

**Core principle:** Hypothesis → Explore → Discover → Iterate until stable.

**Announce:** "I'm using the c3-design skill to guide architecture design."

**CRITICAL:** Read `references/design-guardrails.md` before starting. These guardrails are non-negotiable.

## Mode Detection

| Intent | Mode |
|--------|------|
| "What's the architecture?" | Exploration |
| "How does X work?" | Exploration |
| "I need to add/change..." | Design |
| "Why did we choose X?" | Exploration |

### Exploration Mode

See `references/design-exploration-mode.md` for full workflow.

Quick: Load TOC → Present overview → Navigate on demand.

### Design Mode

Full workflow with mandatory phases.

## Design Phases

**IMMEDIATELY create TodoWrite items:**
1. Phase 1: Surface Understanding
2. Phase 2: Iterative Scoping
3. Phase 3: ADR Creation (MANDATORY)
4. Phase 4: Handoff

See `references/design-phases.md` for detailed phase requirements.

**Rules:**
- Mark `in_progress` when starting
- Mark `completed` only when gate met
- Phase 3 gate: ADR file MUST exist
- Phase 4 gate: Handoff steps executed

## Quick Reference

| Phase | Output | Gate |
|-------|--------|------|
| 1. Surface | Layer hypothesis | Hypothesis formed |
| 2. Scope | Stable scope | No new discoveries |
| 3. ADR | ADR file | File exists |
| 4. Handoff | Complete | Steps executed |

## Layer Skill Delegation

| Impact | Delegate To |
|--------|-------------|
| Container inventory changes | c3-context-design |
| Component organization | c3-container-design |
| Implementation details | c3-component-design |

## Checklist

- [ ] Mode detected (exploration vs design)
- [ ] If design: TodoWrite phases created
- [ ] Each phase completed with gate
- [ ] ADR created (design mode)
- [ ] Handoff executed

## Related

- `references/design-guardrails.md` - **READ FIRST:** Key principles, common mistakes, red flags
- `references/design-exploration-mode.md`
- `references/design-phases.md`
- `references/adr-template.md`
