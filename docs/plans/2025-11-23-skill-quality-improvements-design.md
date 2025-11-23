# Skill Quality Improvements Design

**Date:** 2025-11-23
**Status:** Approved
**Goal:** Reduce skill file sizes to <300 lines, improve cross-skill consistency

## Problem

Current skill files are too large and contain duplicated content:
- c3-adopt: 525 lines
- c3-component-design: 464 lines
- c3-container-design: 458 lines
- c3-migrate: 406 lines
- c3-context-design: 373 lines

This creates:
- High cognitive load for Claude processing skills
- Maintenance burden when patterns change
- Inconsistent terminology across skills

## Solution

### 1. Central References Directory

Create `references/` at plugin root with shared content:

```
references/
├── v3-structure.md          # V3 file paths, ID patterns, hierarchical layout
├── derivation-guardrails.md # Reading order, downward links, naming rules
├── skill-protocol.md        # Announcement format, quick reference table format
└── socratic-method.md       # Question-answer patterns, discovery flow
```

### 2. Skill Slimming Strategy

| Skill | Current | Reduction Strategy | Target |
|-------|---------|-------------------|--------|
| c3-adopt | 525 | Remove embedded build-toc.sh script, extract derivation guardrails, link to socratic-method.md | ~280 |
| c3-component-design | 464 | Remove verbose template examples, extract to appendix, link to derivation-guardrails | ~280 |
| c3-container-design | 458 | Same as component - condense templates, extract shared patterns | ~280 |
| c3-migrate | 406 | Keep migration steps inline, move verbose verification scripts to MIGRATIONS.md | ~290 |
| c3-context-design | 373 | Condense diagram examples, link to derivation-guardrails | ~250 |

### 3. Cross-Skill Consistency

Standardize:
- Section ordering: Overview → When to Use → Quick Reference → The Process → [Skill-specific] → Related Skills
- Terminology: Context/Container/Component (not CTX/CON/COM in prose)
- Link format for references and related skills

## Reference Content Breakdown

### v3-structure.md
Content extracted from:
- c3-locate: V3 ID patterns, file path patterns
- c3-naming: Quick reference table, directory structure example
- c3-adopt: V3 scaffolding, file paths

### derivation-guardrails.md
Content extracted from:
- c3-adopt: Derivation Guardrails section
- c3-context-design: Reference Direction Principle
- c3-container-design: Reference Direction Principle
- c3-component-design: Abstraction Check

### skill-protocol.md
Content standardized from:
- All skills: "Announce at start" pattern
- Quick Reference table format
- When to Use criteria

### socratic-method.md
Content extracted from:
- c3-adopt: The Socratic Approach, Question Flow Pattern
- c3-context-design: Socratic Questions section
- c3-container-design: Socratic Questions section
- c3-component-design: Socratic Questions section

## Implementation Order

1. Create reference files (in parallel)
2. Slim each skill (in sequence, largest first)
3. Verify all skills <300 lines
4. Test skill invocations

## Success Criteria

- [ ] All 9 skills under 300 lines
- [ ] 4 reference files created
- [ ] No broken links between skills and references
- [ ] Consistent section ordering across skills
- [ ] Skills still function correctly when invoked
