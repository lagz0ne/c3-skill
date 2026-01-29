---
name: c3
description: |
  PRIMARY ROUTER for C3 architecture tasks and audit tool for documentation consistency.

  ROUTING RULES (check in order):
  1. **PROVISION-RELATED** (design without implementing):
     → route to c3-provision skill (e.g., "provision X", "design X architecture", "plan X", "envision X", "architect X")
  2. **REF-RELATED** (contains "ref", "refs", "pattern", "patterns", "convention", "standard"):
     → MUST route to c3-ref skill (e.g., "what ref", "show refs", "what patterns")
  3. **CHANGES** (add/modify/remove/fix/refactor/implement):
     → route to c3-alter skill
  4. **QUESTIONS** (where/what/how/explain about components):
     → route to c3-query skill
  5. **NO .c3/** directory:
     → route to onboard skill

  AUDIT MODE: Use when user asks to "audit architecture", "audit C3", "run C3 audit", "run audit",
  "validate C3", "validate architecture", "validate docs", "check C3 docs", "verify docs", "verify C3",
  "verify documentation", "verify architecture", "check documentation", "are docs up to date",
  "is documentation current", "check for stale docs", "docs out of sync", "sync docs", "refresh docs",
  "update docs to match code", "documentation drift", "check architecture", "is C3 current",
  "I have a C3 project", "help me with C3", "show me my architecture", "C3 help".
---

# C3 Architecture Assistant

## REQUIRED: Load References

Before proceeding, use Glob to find and Read these files:
1. `**/references/skill-harness.md` - Red flags and complexity rules
2. `**/references/layer-navigation.md` - How to traverse C3 docs

## Intent Recognition & Routing

This skill is the **primary entry point** for C3 tasks. Route based on intent:

| User Says | Intent | Route To | Agent Chain |
|-----------|--------|----------|-------------|
| "provision/design/plan/envision/architect X" | Design-only | `/c3-provision` | (inline, stops at ADR) |
| "where/what/how/explain/show me" | Question | `/c3-query` | c3-navigator → c3-summarizer |
| "add/modify/remove/fix/refactor/implement" | Change | `/c3-alter` | c3-orchestrator → c3-analysis → c3-synthesizer → c3-dev |
| "pattern/convention/standard/ref/how should we" | Pattern | `/c3-ref` | (inline) |
| "audit/validate/check/verify/sync" | Audit | this skill | → c3-content-classifier |
| (no .c3/ directory) | Initialize | `/onboard` | (inline) |

**When unclear:** Ask "Do you want to explore (query), design only (provision), change and implement (alter), manage patterns (ref), or audit?"

---

## Mode: Adopt

Route to `/onboard` skill for the full staged learning loop.

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
3. Run Phase 2-9: Validate each check
4. Output audit report with PASS/FAIL/WARN per check
5. List actionable fixes for any failures
```

---

## Proactive Pattern Awareness

If hooks are configured, C3 provides **ambient pattern awareness** without explicit invocation:

| Hook | Script | Purpose |
|------|--------|---------|
| SessionStart | `c3-context-loader` | Loads system goal, all refs, code→component mapping |
| PreToolUse (Edit/Write) | `c3-edit-context` | Surfaces relevant patterns when editing documented files |
| PreToolUse (Edit/Write) | `c3-gate` | Gates edits to ADR-approved files only |

This enables **drift prevention** - changes naturally align with established patterns.
