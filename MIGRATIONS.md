# C3 Skill Migrations

> This file defines all migrations between C3 skill versions.
> Each version section describes changes and transforms needed.

Current version: 1

---

## Version 1 (initial)

### Changes
- Initial versioned release
- Establishes `.c3/VERSION` file convention
- Container naming: `C3-N-slug` (N = single digit container number)
- Component naming: `C3-NNN-slug` (N = container digit + 2-digit component number)
- Context naming: `CTX-slug`
- ADR naming: `ADR-NNN-slug`

### Transforms
- None (baseline version)

### Verification
- `.c3/VERSION` exists and contains `1`
- All container files match `C3-[0-9]-*.md` pattern
- All component files match `C3-[0-9][0-9][0-9]-*.md` pattern

---

## Migration Format Reference

Each future version section should include:

```markdown
## Version N (from N-1)

### Changes
- [Human-readable description of what changed]

### Transforms
- **Pattern:** [regex pattern to find]
- **Replace:** [replacement pattern]
- **Files:** [glob pattern for affected files]

### Verification
- [Checks to confirm migration succeeded]
```
