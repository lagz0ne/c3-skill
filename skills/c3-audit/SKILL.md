---
name: c3-audit
description: |
  Audits C3 architecture documentation for consistency, drift, and completeness.

  This skill should be used when the user asks to:
  - "audit C3", "validate docs", "check architecture", "verify C3 docs"
  - "run C3 audit", "check documentation", "are docs up to date", "docs out of sync"
  - "validate architecture", "check for drift", "verify docs match code"

  DO NOT use for: "update docs", "change docs" (route to c3-change for changes).
  Requires .c3/ to exist. Routes navigation to c3-query, changes to c3-change, patterns to c3-ref.
---

# C3 Audit - Architecture Validation

Validate C3 documentation for consistency, drift, and completeness.

## Precondition: C3 Adopted

**STOP if `.c3/README.md` does not exist.**

If missing:
> This project doesn't have C3 docs yet. Use the c3-onboard skill to create documentation first.

Do NOT proceed until `.c3/README.md` is confirmed.

## REQUIRED: Load References

Before proceeding, Read these files (relative to this skill's directory):
1. `references/skill-harness.md` - Red flags and complexity rules
2. `references/audit-checks.md` - Full 10-phase audit procedure

## Audit Phases

Follow the 10-phase audit procedure from `references/audit-checks.md`. Track progress as you work:

```
Audit Progress:
- [ ] Phase 1: Gather - Collect all C3 docs
- [ ] Phase 2: Inventory vs Code - Docs match reality
- [ ] Phase 3: Component Categorization - Foundation/Feature/Ref correct
- [ ] Phase 4: Code Reference Validation - Code References point to real files
- [ ] Phase 5: Diagram Accuracy - Diagrams match current state
- [ ] Phase 6: ADR Lifecycle - ADRs in valid states
- [ ] Phase 7: Ref File Validation - Refs cited correctly
- [ ] Phase 8: Abstraction Boundaries - Layers don't leak
- [ ] Phase 9: Content Separation - Foundation vs Feature vs Ref
- [ ] Phase 10: Context Files - CLAUDE.md presence/freshness
```

## Output Format

```
**C3 Audit Results**

| Phase | Status | Issues |
|-------|--------|--------|
| Inventory vs Code | PASS/WARN/FAIL | [details] |
| ... | ... | ... |

**Summary:** N passes, M warnings, K failures
**Action Items:** [list of fixes needed]
```

## Discovery-Based Audit (Alternative)

For deep health checks or after major codebase changes, use the discovery-based approach documented in `references/audit-checks.md` (section: Discovery-Based Audit). This compares docs against code reality rather than structure rules.

## Routing

If during audit the user wants to fix issues:
- Documentation changes -> Route to c3-change skill
- Pattern issues -> Route to c3-ref skill
- Architecture questions -> Route to c3-query skill

**Agent Teams:** This skill can also serve as the Phase 4 auditor role in c3-change Agent Teams flow.

---

## Example

```
User: "audit C3 docs"

Phase 1: Gathered 3 containers, 12 components, 4 refs
Phase 2: Inventory vs Code → c3-205 references deleted file → FAIL
Phase 3: Categories → c3-103 has no Code References (should be ref?) → FAIL
Phase 4: Code References → 2 stale paths → FAIL
Phase 5-10: PASS

Summary: 7 passes, 3 failures
Action Items:
  1. Update c3-205 Code References (stale path)
  2. Reclassify c3-103 as ref or add Code References
  3. Fix 2 stale paths in Phase 4
```
