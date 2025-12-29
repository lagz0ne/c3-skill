# Migration: 20251229-skill-consolidation

## Summary

Consolidated 3 layer skills into 2 skills. No user `.c3/` changes required.

## Changes

### Skills Renamed

| Old Skill | New Skill |
|-----------|-----------|
| `c3-context-design` | `c3-structure` |
| `c3-container-design` | `c3-structure` |
| `c3-component-design` | `c3-implementation` |

### References Consolidated

| Removed | Merged Into |
|---------|-------------|
| `archetype-hints.md` | Removed (trust model knowledge) |
| `socratic-method.md` | Removed (trust model knowledge) |
| `hierarchy-model.md` | Concepts inline in skills |
| `diagram-decision-framework.md` | `diagram-patterns.md` |
| `container-archetypes.md` | `container-patterns.md` (slimmed) |

### New Model: Inventory-First

Container component inventory is now the source of truth:
- List ALL components in Container inventory
- Component docs appear when conventions mature
- No stubs - either full doc exists or just inventory entry

## User Impact

**No migration required.** This change is skill-internal:
- User `.c3/` directory structure unchanged
- ID patterns unchanged
- Frontmatter unchanged
- File naming unchanged

Users invoking old skill names will need to use new names:
- `c3-context-design` → `c3-structure`
- `c3-container-design` → `c3-structure`
- `c3-component-design` → `c3-implementation`

## Verification

No verification needed - skill-internal change only.
