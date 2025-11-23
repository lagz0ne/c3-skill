# C3 Skill Migrations

> This file defines all migrations between C3 skill versions.
> Each version section describes changes and transforms needed.

Current version: 3

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

## Version 2 (from 1)

### Changes
- Flatten components directory (remove container subfolders)
- Primary context becomes README.md with `id: context`
- Path: `components/{container}/C3-*.md` → `components/C3-*.md`
- Path: `CTX-system-overview.md` → `README.md`

### Transforms

**Move nested components to flat:**
- **Files:** `components/*/*.md`
- **Action:** Move to `components/` (parent directory)
- **Command:** `mv .c3/components/*/*.md .c3/components/ && rmdir .c3/components/*/`

**Rename primary context:**
- **Pattern:** `CTX-system-overview.md`
- **Replace:** `README.md`
- **Frontmatter:** Change `id: CTX-system-overview` → `id: context`

**Update internal links:**
- **Pattern:** `](./components/[^/]+/(C3-[0-9])`
- **Replace:** `](./components/$1`
- **Files:** `.c3/**/*.md`

**Update context links:**
- **Pattern:** `CTX-system-overview.md`
- **Replace:** `README.md`
- **Files:** `.c3/**/*.md`

### Verification
- `.c3/VERSION` contains `2`
- No subdirectories in `.c3/components/`
- `.c3/README.md` exists with `id: context`
- All component files match `.c3/components/C3-[0-9][0-9][0-9]-*.md`
- No broken internal links (grep for old paths returns empty)

---

## Version 3 (from 2)

### Changes
- Containers become folders containing their components
- No more `containers/` or `components/` type folders
- All lowercase naming (c3-1-backend instead of C3-1-backend)
- IDs are numeric only (c3-1, c3-101 instead of C3-1-backend, C3-101-db-pool)
- Context ID changes from `context` to `c3-0`
- VERSION file removed, `c3-version` added to README.md frontmatter
- ADRs lowercase (adr-001 instead of ADR-001)

### Transforms

**Convert containers to folders:**
- **Pattern:** `containers/C3-{N}-{slug}.md`
- **Action:** Create `c3-{n}-{slug}/README.md`, move content, update id to `c3-{n}`

**Move components into container folders:**
- **Pattern:** `components/C3-{N}{NN}-{slug}.md`
- **Action:** Move to `c3-{n}-*/c3-{n}{nn}-{slug}.md`, update id to `c3-{n}{nn}`

**Update context:**
- Change `id: context` to `id: c3-0`
- Add `c3-version: 3` to README.md frontmatter
- Remove VERSION file

**Lowercase ADRs:**
- Rename `ADR-{NNN}-{slug}.md` to `adr-{nnn}-{slug}.md`
- Update id to lowercase

**Update internal links:**
- `](./containers/C3-` → `](./c3-`
- `](./components/C3-` → container-relative paths

### Verification
- `.c3/README.md` contains `c3-version: 3` and `id: c3-0`
- No `containers/` or `components/` directories exist
- Container folders match `c3-[0-9]-*/` pattern
- Each container folder has `README.md`
- All IDs are numeric only

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
