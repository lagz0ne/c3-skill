# Component References Design

## Summary

Add explicit code references to component documents to make the link between architecture docs and implementation less brittle.

## Problem

Currently, the connection between component docs and actual code is inferred through naming conventions and discovery patterns during audit. This is brittle - docs and code can drift apart without detection.

## Solution

Add a `## References` section to all component templates containing a simple bullet list of code references.

### Reference Format

```markdown
## References

- ClassName
- functionName()
- src/path/to/relevant/**/*.ts
```

**Reference types (implicit, no markers needed):**
- Symbol names (classes, functions, hooks) - most stable, survives file moves
- Glob patterns - for components spanning multiple files
- File paths - fallback for config files or single-file implementations

**Guidelines:**
- Symbols first, patterns for groups, exact paths sparingly
- Keep it short: 3-7 references typical
- If more needed, component may be too broad

### Section Placement

After main content, before `## Testing`:

```markdown
## Contract / Conventions / Behavior
...

## References

- ...

## Testing
...
```

## Operation Changes

### Onboard

**Doc-first approach:**
1. Create component doc based on architectural understanding (what it does, why)
2. Then add references linking to where that concept is implemented in code

References are a lookup index from architecture to code, not a reflection of code structure.

**Workflow change:**
- Stages 3-5: Create component docs (unchanged)
- New step after each component doc: locate implementing code, populate `## References`

### ADR (Alter)

Reference maintenance folds into existing workflow:

**During ADR creation:**
- If change adds/removes/moves code, note which component references need updating
- ADR scope can flag: "References affected: `c3-101`, `c3-205`"

**During ADR execution:**
- After code changes, update `## References` in affected component docs
- Part of existing "update affected docs" step

### Implementation Plan

**Before execution:**
- Update `## References` in affected components with code locations from the plan
- Front-loads reference maintenance while knowledge is fresh

**Plan template addition:**
```markdown
## Pre-execution checklist
- [ ] Update `## References` in affected components with code locations from this plan
```

### Audit

**New audit checks:**
- Reference validity - do referenced symbols/files exist?
- Reference staleness - did code move/rename without updating references?
- Coverage - are there major code areas not referenced by any component?

**Validation approach:**
```
For each component:
  - Read ## References
  - For each reference:
    - Symbol: grep for definition, flag if not found
    - Pattern: glob, flag if zero matches
    - Path: check exists, flag if missing
  - Report: valid, stale, or broken references
```

**Bidirectional validation:**
- Current: infer code→doc mapping via naming conventions
- New: also check doc→code mapping via explicit references
- Catches drift from both directions

## Files to Modify

| File | Change |
|------|--------|
| `templates/component-foundation.md` | Add `## References` section |
| `templates/component-auxiliary.md` | Add `## References` section |
| `templates/component-feature.md` | Add `## References` section |
| `references/audit-checks.md` | Add reference validation checks |
| `references/plan-template.md` | Add pre-execution checklist item |
| `skills/c3/SKILL.md` | Update onboard workflow with reference step |
| `skills/c3-alter/SKILL.md` | Note reference updates in ADR workflow |

## Design Decisions

1. **Dedicated section over frontmatter** - more visible, allows prose if needed
2. **Single "references" list** - no defines/uses distinction, simpler to maintain
3. **No type markers** - format makes type clear (symbols vs paths vs patterns)
4. **Doc-first** - architecture drives structure, references are just pointers to implementation
