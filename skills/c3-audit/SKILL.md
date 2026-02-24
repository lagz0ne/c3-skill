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

The audit uses a three-tier approach: **structural** (CLI-automated) → **inventory** (CLI-assisted) → **semantic** (manual reasoning).

Follow the audit procedure from `references/audit-checks.md`. Track progress as you work:

```
Audit Progress:
- [ ] Phase 0: Structural Validation - Run `npx -y c3-kit check` for broken links, orphans, duplicates, missing parents
- [ ] Phase 1: Gather Inventory - Run `npx -y c3-kit list --json` for full entity inventory
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

### Phase 0: Structural Validation (CLI)

Run `npx -y c3-kit check` via Bash to detect structural issues automatically:

```bash
npx -y c3-kit check
```

This catches broken links, orphan entities, duplicate IDs, and missing parent references — issues that Phases 2-7 previously checked manually. For machine-readable output, use `npx -y c3-kit check --json`.

If `npx -y c3-kit check` reports failures, record them immediately. Many will overlap with later phases — skip re-checking those manually.

### Phase 1: Gather Inventory (CLI)

Run `npx -y c3-kit list --json` via Bash to get the full entity inventory:

```bash
npx -y c3-kit list --json
```

This returns all entities with id, type, title, path, relationships, and frontmatter. Use this output as the source of truth for subsequent phases instead of manually running Glob+Read across `.c3/` directories.

For a quick topology overview, use `npx -y c3-kit list` (text format with goals).

### Phases 2-10: Semantic Validation (Manual)

Continue with `references/audit-checks.md` Phases 2-10 using Read+Grep+reasoning. Use the inventory from Phase 1 to drive these checks — no need to re-gather entities.

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
- Impact assessment of proposed change -> Route to c3-sweep skill
- Pattern issues -> Route to c3-ref skill
- Architecture questions -> Route to c3-query skill

**Agent Teams:** This skill can also serve as the Phase 4 auditor role in c3-change Agent Teams flow.

---

## Example

```
User: "audit C3 docs"

Phase 0: `npx -y c3-kit check` → 1 broken link (c3-205 → deleted file), 1 orphan ref → FAIL
Phase 1: `npx -y c3-kit list --json` → 3 containers, 12 components, 4 refs
Phase 2: Inventory vs Code → (broken link already caught in Phase 0, skip) → PASS
Phase 3: Categories → c3-103 has no Code References (should be ref?) → FAIL
Phase 4: Code References → (stale paths already caught in Phase 0, skip) → PASS
Phase 5-10: PASS

Summary: 9 passes, 2 failures (Phase 0 structural, Phase 3 semantic)
Action Items:
  1. Fix broken link in c3-205 (detected by npx -y c3-kit check)
  2. Reclassify c3-103 as ref or add Code References
  3. Resolve orphan ref (detected by npx -y c3-kit check)
```
