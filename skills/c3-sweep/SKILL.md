---
name: c3-sweep
description: |
  Sweeps architecture for impact before a change — spawns agents per C3 entity for constraint-checked analysis.
  Advisory only — does NOT make changes (route to c3-change for implementation).

  This skill should be used when the user asks to:
  - "assess C3 impact of X", "what would break in C3 if..."
  - "check C3 constraints for X", "is this change safe per C3 docs"
  - "affected C3 components", "which C3 containers are impacted by X"
  - "sweep C3 architecture", "impact assessment against C3 architecture"

  <example>
  user: "assess C3 impact if I replace the auth system"
  assistant: "Using c3-sweep to assess architectural impact."
  </example>

  DO NOT use for: "what is X" / "explain X" (c3-query), implementation (c3-change), patterns (c3-ref).
  Requires .c3/ to exist.
---

# C3 Sweep: Impact Assessment Team

Sweep the architecture for impact of proposed changes through a team of agents that read C3 docs dynamically.

## Precondition: C3 Adopted

Run `npx -y c3-kit list --json` via Bash. **STOP if it fails or returns empty.**

If missing:
> This project doesn't have C3 architecture docs yet. Use the c3-onboard skill to create documentation first.

## How It Works

This skill creates an Agent Team with a lead and specialized workers:

```
You (user)
  ↕ conversation
Team Lead (c3-sweep-lead agent, delegate mode)
  ↕ coordinates
Workers:
  - Container workers (read container docs, delegate to component workers)
  - Component workers (inspect code against docs + refs)
  - Ref workers (check convention compliance)
```

## Execution

**HARD RULE: Your FIRST action must be to spawn the c3-sweep-lead agent.** Do not read C3 docs yourself, do not create teams yourself, do not assess impact yourself. The lead handles everything.

Spawn the lead:
```
Task tool:
  subagent_type: "c3-skill:c3-sweep-lead"
  prompt: "<pass the user's full change request / question here>"
  mode: "delegate"
```

The lead will:
1. **Load topology** — Run `npx -y c3-kit list --json` to get full system structure (entities, relationships, frontmatter)
2. **Identify entities** — Match the change to affected containers, refs, and ADRs from the JSON
3. **Delegate** — Spawn container + ref workers in parallel for deep inspection
4. **Synthesize** — Collect advisories into a unified impact assessment

The lead communicates back when it needs clarification. Relay these to the user and pass their responses back.

## Team Configuration

The lead operates in delegate mode (coordination only, never modifies code or docs). The lead creates its own Agent Team via `TeamCreate` and spawns workers into it. Workers read C3 docs directly.

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
