---
name: c3-change
description: |
  Orchestrates architectural changes through Agent Teams with ADR-first workflow.

  This skill should be used when the user asks to:
  - "add a component", "add feature", "add X to Y", "new component", "create new service"
  - "add endpoint", "add API", "add middleware", "add handler", "add route"
  - "change architecture", "modify X", "update component", "refactor X"
  - "fix X", "fix bug", "fix issue", "implement X", "implement feature"
  - "remove component", "delete component", "deprecate X"
  - "provision X", "design the architecture for X", "plan X architecture"
  - "replace X with Y", "migrate X", "upgrade X"

  DO NOT use for: "what is X", "how does X work", "explain X" (route to c3-query).
  DO NOT use for: "add pattern", "create ref", "update ref" (route to c3-ref).
  DO NOT use for: "audit C3", "validate docs" (route to c3-audit).
  Requires .c3/ to exist. All changes flow through ADR process with Agent Teams.
---

# C3 Change - Agent Teams Workflow

Orchestrate architectural changes through a team of Claude Code agents.

## Precondition: C3 Adopted

**STOP if `.c3/README.md` does not exist.**

If missing:
> This project doesn't have C3 docs yet. Use the c3-onboard skill to create documentation first.

## How It Works

This skill creates an Agent Team with a lead and specialized workers:

```
You (user)
  ↕ conversation
Team Lead (c3-lead agent, delegate mode)
  ↕ coordinates
Workers:
  - Analyst (Phase 1: impact investigation)
  - Reviewer (Phase 1: challenge findings)
  - Implementer x N (Phase 3: execute tasks)
  - Auditor (Phase 4: verify docs vs code)
```

## Setup

Tell the lead about your change. The lead will:

1. **Understand** — Spawn analyst + reviewer to investigate impact
2. **ADR** — Write Architecture Decision Record with Work Breakdown
3. **Execute** — Decompose into tasks, spawn implementers
4. **Audit** — Spawn auditor to verify C3 docs match code

## Team Configuration

The lead operates in delegate mode (coordination only, never writes code). Workers are spawned as Agent Teams workers when available, or as Task subagents otherwise. Either way, workers are full Claude Code sessions that read C3 docs directly.

## Phase Details

### Phase 1: Understand
Lead reads C3 docs, clarifies with you, then spawns analyst and reviewer workers who debate the impact. Lead synthesizes and presents findings.

### Phase 2: ADR
Lead writes ADR with Work Breakdown. You review and accept. No separate plan files — the ADR IS the plan.

**ADR format:** Always use YAML frontmatter (not markdown-style headers):
```yaml
---
id: adr-YYYYMMDD-{slug}
title: [Decision Title]
status: proposed | accepted | provisioned | implemented
date: YYYY-MM-DD
base-commit: [git hash]
affects: [c3-N, c3-NNN]
approved-files: []
---
```
This applies to BOTH provision and implementation modes.

### Phase 2b: Provision Gate
After ADR acceptance, the lead asks: **implement now or design only?**

- **Implement now** → Continue to Phase 3
- **Design only (provision)** → Lead creates component docs with `status: provisioned` in the main `.c3/` tree, marks ADR `provisioned`, done. No Code References section (no code exists yet). These docs are visible to c3-query and c3-audit immediately.

To **implement a provisioned design** later, invoke c3-change again. The lead picks up the existing provisioned ADR + docs as starting context and resumes from Phase 3.

### Phase 3: Execute
Lead creates tasks from Work Breakdown. Each task points to C3 component docs and refs (mandatory reading). Implementer workers claim tasks and execute. Lead reviews each completion.

### Phase 4: Audit
Lead spawns auditor to compare C3 docs vs actual changes. Auditor updates CLAUDE.md files. Lead transitions ADR to implemented.

## Regression

Late discoveries during any phase trigger the regression decision tree:
- Changes the problem → back to Phase 1
- Changes the approach → back to Phase 2
- Expands scope → amend ADR
- Implementation detail → adjust tasks

Workers surface discoveries to the lead. Lead decides how far to regress. User confirms anything that affects the ADR.

## Routing

- Architecture questions → c3-query skill
- Pattern management → c3-ref skill
- Standalone audit → c3-audit skill
- New project documentation → c3-onboard skill
