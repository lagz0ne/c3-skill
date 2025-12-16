# Migration: 20251216-workflow-guidelines

## Changes

Added development workflow guidelines to CLAUDE.md:

- **Conventional commits** - Required format with types (feat, fix, docs, refactor, chore, test) and scopes
- **Version changes** - Clarified skill version vs C3 version distinction
- **Migration decision** - Added decision matrix and questions for determining migration needs

## Transforms

**No automatic transforms required.**

This is a documentation-only change affecting plugin development guidelines.

## Verification

```bash
# Verify CLAUDE.md has new sections
grep -c "Conventional Commits" CLAUDE.md        # Should return 1
grep -c "Migration Decision" CLAUDE.md          # Should return 1
grep -c "Version Changes" CLAUDE.md             # Should return 1
```
