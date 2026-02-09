---
name: living-entity
description: |
  Read-only impact assessment for proposed changes using C3 docs.
  Activates a team lead who coordinates specialized workers for constraint-checked analysis.
  Advisory only — does NOT make changes (route to c3-change for implementation).

  This skill should be used when the user asks to:
  - "assess C3 impact of X", "what would break in C3 if..."
  - "check C3 constraints for X", "is this change safe per C3 docs"
  - "affected C3 components", "which C3 containers are impacted by X"
  - "living entity analysis", "impact assessment against C3 architecture"

  <example>
  user: "assess C3 impact if I replace the auth system"
  assistant: "Using living-entity to assess architectural impact."
  </example>

  DO NOT use for: "what is X" / "explain X" (c3-query), implementation (c3-change), patterns (c3-ref).
  Requires .c3/ to exist.
---

# Living Entity: Impact Assessment Team

Assess the architectural impact of proposed changes through a team of agents that read C3 docs dynamically.

## Precondition: C3 Adopted

**STOP if `.c3/README.md` does not exist.**

If missing:
> This project doesn't have C3 architecture docs yet. Use the c3-onboard skill to create documentation first.

## How It Works

This skill creates an Agent Team with a lead and specialized workers:

```
You (user)
  ↕ conversation
Team Lead (living-entity-lead agent, delegate mode)
  ↕ coordinates
Workers:
  - Container workers (read container docs, delegate to component workers)
  - Component workers (inspect code against docs + refs)
  - Ref workers (check convention compliance)
```

## Setup

Tell the lead about the change you're considering. The lead will:

1. **Read topology** — Parse `.c3/TOC.md` and `README.md` for system structure
2. **Identify entities** — Match the change to affected containers, refs, and ADRs
3. **Delegate** — Spawn container + ref workers in parallel for deep inspection
4. **Synthesize** — Collect advisories into a unified impact assessment

## Team Configuration

The lead operates in delegate mode (coordination only, never modifies code or docs). The lead tries `TeamCreate` first for full Agent Teams coordination; if unavailable, falls back to `Task` with `subagent_type`. Either way, workers read C3 docs directly.

### Worker Tiers

| Worker | Role |
|--------|------|
| Container | Reads container README, identifies affected components, spawns component workers |
| Component | Inspects actual code against component doc + applicable refs |
| Ref | Checks whether the change complies with or violates conventions |

## Assessment Output

The lead synthesizes all worker advisories into:

1. **Affected Entities** — which containers and components, with reasons
2. **Constraint Chain** — all conventions, refs, and ADRs that apply
3. **File Changes** — specific files that would need modification
4. **Risks** — edge cases, relationship impacts, ADR conflicts
5. **Recommended Approach** — step-by-step plan respecting all constraints

## Routing

- Want to implement after assessment? → c3-change skill
- Architecture questions without change context → c3-query skill
- Pattern management → c3-ref skill
- Standalone audit → c3-audit skill
