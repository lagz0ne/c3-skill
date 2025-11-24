# Migration: 20251124-adr-date-naming

> From: `3` (or `00000000-baseline`)
> To: `20251124-adr-date-naming`

## Changes

- ADR naming changes from sequential `adr-{nnn}-{slug}` to date-based `adr-{YYYYMMDD}-{slug}`
- ADR IDs now include the full slug: `adr-20251124-database` instead of `adr-001`
- Skill versioning changes from integer to `YYYYMMDD-slug` format
- Migration files now in `migrations/` directory instead of single `MIGRATIONS.md`
- `c3-version` frontmatter value changes from integer to date-slug format

## Transforms

### Rename ADR files

**For each file matching:** `.c3/adr/adr-[0-9][0-9][0-9]-*.md`

1. Get creation date:
   ```bash
   git log --follow --format=%aI --diff-filter=A -- "$FILE" | tail -1
   ```
2. Extract YYYYMMDD from ISO date
3. Extract slug from filename (portion after `adr-nnn-`)
4. Rename: `adr-{nnn}-{slug}.md` -> `adr-{YYYYMMDD}-{slug}.md`

### Update ADR frontmatter

**Pattern:** `^id: adr-[0-9]{3}$`
**Replace:** `id: adr-{YYYYMMDD}-{slug}` (using same date and slug as filename)

### Update c3-version in context

**File:** `.c3/README.md` (or `.c3/c3-0-*/README.md`)
**Pattern:** `^c3-version: 3$`
**Replace:** `c3-version: 20251124-adr-date-naming`

### Update internal links

**Files:** `.c3/**/*.md`
**Pattern:** `adr-[0-9]{3}` (in links and references)
**Replace:** Corresponding `adr-{YYYYMMDD}-{slug}`

## Verification

```bash
# No legacy ADR filenames
! ls .c3/adr/adr-[0-9][0-9][0-9]-*.md 2>/dev/null

# All ADRs use date format
ls .c3/adr/adr-[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]-*.md

# c3-version updated
grep -q '^c3-version: 20251124-adr-date-naming$' .c3/README.md

# No broken links to old ADR format
! grep -r 'adr-[0-9]\{3\}[^0-9]' .c3/ 2>/dev/null
```
