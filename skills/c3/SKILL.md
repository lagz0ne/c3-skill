---
name: c3
description: |
  Routes C3 architecture requests and audits existing C3 documentation for consistency.
  Use when the user asks to "audit architecture", "validate C3", "check C3 docs", "verify docs",
  "check documentation", "are docs up to date", or when no .c3/ directory exists (routes to onboarding).
  Routes navigation to c3-navigator agent, changes to c3-orchestrator agent.
---

# C3 Architecture Assistant

## REQUIRED: Load References

Before proceeding, use Glob to find and Read these files:
1. `**/references/skill-harness.md` - Red flags and complexity rules
2. `**/references/layer-navigation.md` - How to traverse C3 docs

## Mode Selection

| Condition | Mode |
|-----------|------|
| No `.c3/README.md` | **Adopt** - Route to `/onboard` skill |
| Has `.c3/` + "audit" intent | **Audit** - Validate docs |
| Has `.c3/` + question/navigation | Use `c3-skill:c3-navigator` agent |
| Has `.c3/` + change request | Use `c3-skill:c3-orchestrator` agent |

---

## Mode: Adopt

Route to `/onboard` skill for the full staged learning loop.

See `skills/onboard/SKILL.md` for the complete workflow.

---

## Mode: Audit

**REQUIRED:** Load `**/references/audit-checks.md` for full procedure.

| Scope | Command |
|-------|---------|
| Full system | `audit C3` |
| Container | `audit container c3-1` |
| ADR | `audit adr adr-YYYYMMDD-slug` |

**Checks:** Inventory vs code, categorization, reference validity, diagrams, ADR lifecycle, abstraction boundaries, content separation

**Content Separation Check (Phase 9):**
- Verifies components contain domain logic, refs contain usage patterns
- Uses `c3-skill:c3-content-classifier` agent for LLM-based analysis
- Detects: missing refs for technologies, integration patterns in components, duplicated patterns

**Example:**
```
User: "Check if C3 docs are up to date"

1. Load audit-checks.md
2. Run Phase 1: Gather (list containers, components, ADRs)
3. Run Phase 2-9: Validate each check (including content separation)
4. Output audit report with PASS/FAIL/WARN per check
5. List actionable fixes for any failures
```
