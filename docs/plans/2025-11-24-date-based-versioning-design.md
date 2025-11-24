# Date-Based ADR and Version Naming Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace sequential numbering with date-based naming for ADRs and skill versions to eliminate collision in concurrent system changes.

**Architecture:** Migrate from integer versions (3) to date-slug format (YYYYMMDD-slug). ADRs change from adr-nnn to adr-YYYYMMDD-slug. Migration files move from single MIGRATIONS.md to individual files in migrations/ directory.

**Tech Stack:** Markdown, bash, git

---

## Summary

Replace sequential numbering with date-based naming for ADRs and skill versions to eliminate collision in concurrent system changes.

## Changes

### 1. ADR Naming

| Aspect | Old | New |
|--------|-----|-----|
| Pattern | `adr-{nnn}-{slug}.md` | `adr-{YYYYMMDD}-{slug}.md` |
| Example | `adr-001-database.md` | `adr-20251124-database.md` |
| ID | `adr-001` | `adr-20251124-database` |
| Anchor | `{#adr-001-*}` | `{#adr-20251124-*}` |

**Benefits:**
- Eliminates collision when multiple developers create ADRs concurrently
- Natural chronological sorting
- Date provides context without opening file

### 2. Skill Version Naming

| Aspect | Old | New |
|--------|-----|-----|
| Pattern | Integer (`3`) | `YYYYMMDD-slug` |
| Example | `3` | `20251124-adr-date-naming` |
| Location | `VERSION` file | `VERSION` file |
| Frontmatter | `c3-version: 3` | `c3-version: 20251124-adr-date-naming` |

**Benefits:**
- Multiple migrations per day possible (unique slugs)
- Self-documenting version names
- No coordination needed for version numbers

### 3. Migration File Structure

| Aspect | Old | New |
|--------|-----|-----|
| Location | Single `MIGRATIONS.md` | `migrations/` directory |
| File naming | Sections in one file | `YYYYMMDD-slug.md` per version |
| Lookup | Parse single file | List directory, sort, compare |

**Directory structure:**
```
c3-design/
└── migrations/
    ├── 00000000-baseline.md           # Combined v1+v2+v3 history
    └── 20251124-adr-date-naming.md    # New migrations
```

**Baseline:** `00000000-baseline.md` consolidates legacy v1, v2, v3 migrations and sorts before any YYYYMMDD version.

## Search Patterns

### ADR Discovery

| Method | Pattern |
|--------|---------|
| Bash glob | `ls .c3/adr/adr-[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]-*.md` |
| Regex | `adr-[0-9]{8}-[a-z]` |
| By year | `ls .c3/adr/adr-2025*.md` |
| By month | `ls .c3/adr/adr-202511*.md` |

### Version Comparison

Versions sort lexicographically:
- Legacy numeric: `1` < `2` < `3`
- Legacy < date-based: `3` < `20251124-adr-date-naming`
- Date-based: `20251124-foo` < `20251125-bar`

## Migration Strategy

### For Existing ADRs

1. Get original creation date: `git log --follow --format=%aI --diff-filter=A -- <file> | tail -1`
2. Extract YYYYMMDD from ISO date
3. Rename: `adr-{nnn}-{slug}.md` → `adr-{YYYYMMDD}-{slug}.md`
4. Update frontmatter ID: `id: adr-{nnn}` → `id: adr-{YYYYMMDD}-{slug}`
5. Update all internal links in `.c3/**/*.md`

### Collision Handling

If two ADRs have same creation date AND same slug, append `-2` to slug during migration.

### Verification

- No files matching `adr-[0-9]{3}-*.md` pattern
- All ADRs match `adr-[0-9]{8}-*.md` pattern
- No broken links (grep for old `adr-nnn` patterns returns empty)

## Files to Update

| File | Action |
|------|--------|
| `VERSION` | `3` → `20251124-adr-date-naming` |
| `MIGRATIONS.md` | Delete |
| `migrations/00000000-baseline.md` | Create (combined v1+v2+v3) |
| `migrations/20251124-adr-date-naming.md` | Create |
| `skills/c3-naming/SKILL.md` | Update ADR patterns |
| `skills/c3-migrate/SKILL.md` | Update to use migrations/ directory |
| `references/v3-structure.md` | Update ADR patterns |
| `CLAUDE.md` | Add version slug uniqueness rule |

## Uniqueness Rule

Version slugs must be unique across all migrations. Reviewers reject duplicate slugs.

**Check:** `ls migrations/ | grep -c "YYYYMMDD-your-slug"` should return 0 before creating.

## Decisions

- **Date format:** YYYYMMDD (compact, no separators)
- **Legacy migration:** Use git file creation date for existing ADRs
- **Legacy versions:** Consolidated into `00000000-baseline.md`
- **Migration order:** Newest at top when listing (lexicographic sort)
- **Lookup location:** Skill plugin directory `migrations/`

---

## Implementation Tasks

### Task 1: Create migrations directory and baseline file

**Files:**
- Create: `migrations/00000000-baseline.md`
- Delete: `MIGRATIONS.md`

**Step 1: Create migrations directory**

```bash
mkdir -p migrations
```

**Step 2: Create baseline file with combined v1+v2+v3 history**

Create `migrations/00000000-baseline.md` with content:

```markdown
# Baseline Migration (v1 + v2 + v3)

> This file consolidates the historical v1, v2, and v3 migrations into a single baseline.
> Users at any legacy version (1, 2, 3) should migrate to this baseline first.

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
```

**Step 3: Delete old MIGRATIONS.md**

```bash
rm MIGRATIONS.md
```

**Step 4: Commit**

```bash
git add migrations/00000000-baseline.md
git rm MIGRATIONS.md
git commit -m "refactor: move migrations to individual files with baseline"
```

---

### Task 2: Create new migration file for date-based versioning

**Files:**
- Create: `migrations/20251124-adr-date-naming.md`

**Step 1: Create migration file**

Create `migrations/20251124-adr-date-naming.md` with content:

```markdown
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
4. Rename: `adr-{nnn}-{slug}.md` → `adr-{YYYYMMDD}-{slug}.md`

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
```

**Step 2: Commit**

```bash
git add migrations/20251124-adr-date-naming.md
git commit -m "feat: add migration for date-based ADR and version naming"
```

---

### Task 3: Update VERSION file

**Files:**
- Modify: `VERSION`

**Step 1: Update VERSION content**

Change content from:
```
3
```

To:
```
20251124-adr-date-naming
```

**Step 2: Commit**

```bash
git add VERSION
git commit -m "chore: update VERSION to 20251124-adr-date-naming"
```

---

### Task 4: Update c3-naming skill

**Files:**
- Modify: `skills/c3-naming/SKILL.md`

**Step 1: Update ADR pattern in Quick Reference table**

Change:
```markdown
| ADR | `adr-{nnn}` | `.c3/adr/adr-{nnn}-{slug}.md` | `.c3/adr/adr-002-postgresql.md` |
```

To:
```markdown
| ADR | `adr-{YYYYMMDD}-{slug}` | `.c3/adr/adr-{YYYYMMDD}-{slug}.md` | `.c3/adr/adr-20251124-postgresql.md` |
```

**Step 2: Update Search Patterns table**

Change:
```markdown
| ADR | `adr-` | `adr-[0-9]{3}-` | Always 3 digits, lowercase |
```

To:
```markdown
| ADR | `adr-2` | `adr-[0-9]{8}-` | 8 digits (YYYYMMDD), lowercase |
```

**Step 3: Update bash examples**

Change:
```bash
# Find all ADRs
ls .c3/adr/adr-*.md
```

To:
```bash
# Find all ADRs
ls .c3/adr/adr-[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]-*.md

# Find ADRs by year
ls .c3/adr/adr-2025*.md

# Find ADRs by month
ls .c3/adr/adr-202511*.md
```

**Step 4: Update Flow diagram**

Change:
```markdown
    Type -->|ADR| Adr[adr-nnn-slug<br/>lowercase, 3 digits]
```

To:
```markdown
    Type -->|ADR| Adr[adr-YYYYMMDD-slug<br/>date-based, unique slug]
```

**Step 5: Update How to Name section item 4**

Change:
```markdown
4) File paths mirror hierarchy:
   - Context: `.c3/c3-0-system/README.md`
   - Container: `.c3/c3-2-frontend/README.md`
   - Component: `.c3/c3-2-frontend/c3-201-api-client.md`
   - ADR: `.c3/adr/adr-003-caching.md`
```

To:
```markdown
4) File paths mirror hierarchy:
   - Context: `.c3/c3-0-system/README.md`
   - Container: `.c3/c3-2-frontend/README.md`
   - Component: `.c3/c3-2-frontend/c3-201-api-client.md`
   - ADR: `.c3/adr/adr-20251124-caching.md`
```

**Step 6: Update Excellent Example**

Change:
```markdown
ADRs:
- .c3/adr/adr-001-database-choice.md
- .c3/adr/adr-002-auth-strategy.md
```

To:
```markdown
ADRs:
- .c3/adr/adr-20250115-database-choice.md
- .c3/adr/adr-20250220-auth-strategy.md
```

**Step 7: Update Directory Structure Example**

Change:
```markdown
  adr/
    adr-001-database.md
    adr-002-auth.md
```

To:
```markdown
  adr/
    adr-20250115-database.md
    adr-20250220-auth.md
```

**Step 8: Update Checklist**

Change:
```markdown
- IDs follow patterns above (c3-0/c3-N/c3-NNN/adr-nnn).
```

To:
```markdown
- IDs follow patterns above (c3-0/c3-N/c3-NNN/adr-YYYYMMDD-slug).
```

**Step 9: Commit**

```bash
git add skills/c3-naming/SKILL.md
git commit -m "docs: update c3-naming skill for date-based ADR naming"
```

---

### Task 5: Update c3-migrate skill

**Files:**
- Modify: `skills/c3-migrate/SKILL.md`

**Step 1: Update description in frontmatter**

Change:
```yaml
description: Migrate .c3/ documentation to current skill version - reads VERSION, compares against MIGRATIONS.md, executes transforms in batches
```

To:
```yaml
description: Migrate .c3/ documentation to current skill version - reads VERSION, compares against migrations/ directory, executes transforms in batches
```

**Step 2: Update Phase 2 description**

Change:
```markdown
1. Parse `MIGRATIONS.md` from plugin directory
2. For each version from `PROJECT + 1` to `SKILL`:
```

To:
```markdown
1. List files in `migrations/` directory from plugin directory
2. Sort lexicographically (numeric versions first, then YYYYMMDD-slug)
3. For each migration newer than `PROJECT`:
```

**Step 3: Update Version storage table**

Change:
```markdown
**Version storage:**
| Format | Location |
|--------|----------|
| v1/v2 | `.c3/VERSION` file |
| v3+ | `c3-version:` in `.c3/README.md` frontmatter |
```

To:
```markdown
**Version storage:**
| Format | Location |
|--------|----------|
| v1/v2 | `.c3/VERSION` file |
| v3 | `c3-version: 3` in `.c3/README.md` frontmatter |
| v4+ (date-based) | `c3-version: YYYYMMDD-slug` in `.c3/README.md` frontmatter |
```

**Step 4: Update Version comparison logic**

Add after Compare Versions table:
```markdown
**Version ordering:**
- Numeric versions (1, 2, 3) sort before date-based versions
- Date-based versions sort lexicographically: `20251124-foo` < `20251125-bar`
- Example: `1` < `2` < `3` < `20251124-adr-date-naming` < `20251125-next-change`
```

**Step 5: Update Phase 5 Finalize section**

Change:
```markdown
### Update Version

```bash
if [ "$TARGET_VERSION" -ge 3 ]; then
    # v3+: Update frontmatter
    sed -i "s/^c3-version: .*/c3-version: $TARGET_VERSION/" .c3/README.md
    rm -f .c3/VERSION
else
    # v1/v2: Update VERSION file
    echo "$TARGET_VERSION" > .c3/VERSION
fi
```
```

To:
```markdown
### Update Version

```bash
# For numeric versions (1, 2, 3)
if [[ "$TARGET_VERSION" =~ ^[0-9]+$ ]]; then
    if [ "$TARGET_VERSION" -ge 3 ]; then
        sed -i "s/^c3-version: .*/c3-version: $TARGET_VERSION/" .c3/README.md
        rm -f .c3/VERSION
    else
        echo "$TARGET_VERSION" > .c3/VERSION
    fi
else
    # For date-based versions (YYYYMMDD-slug)
    sed -i "s/^c3-version: .*/c3-version: $TARGET_VERSION/" .c3/README.md
    rm -f .c3/VERSION
fi
```
```

**Step 6: Update Related section**

Change:
```markdown
## Related

- [v3-structure.md](../../references/v3-structure.md)
- [MIGRATIONS.md](../../MIGRATIONS.md)
```

To:
```markdown
## Related

- [v3-structure.md](../../references/v3-structure.md)
- [migrations/](../../migrations/) - Individual migration files
```

**Step 7: Commit**

```bash
git add skills/c3-migrate/SKILL.md
git commit -m "feat: update c3-migrate to use migrations/ directory"
```

---

### Task 6: Update references/v3-structure.md

**Files:**
- Modify: `references/v3-structure.md`

**Step 1: Update ID Patterns table**

Change:
```markdown
| ADR | `adr-{nnn}` (nnn=001-999) | `adr-001`, `adr-042` |
```

To:
```markdown
| ADR | `adr-{YYYYMMDD}-{slug}` | `adr-20251124-database`, `adr-20250115-auth` |
```

**Step 2: Update File Paths table**

Change:
```markdown
| ADR | `.c3/adr/adr-{nnn}-{slug}.md` | `.c3/adr/adr-001-database.md` |
```

To:
```markdown
| ADR | `.c3/adr/adr-{YYYYMMDD}-{slug}.md` | `.c3/adr/adr-20251124-database.md` |
```

**Step 3: Update Directory Layout**

Change:
```markdown
├── adr/                   # Architecture Decision Records
│   └── adr-001-*.md
```

To:
```markdown
├── adr/                   # Architecture Decision Records
│   └── adr-YYYYMMDD-*.md  # e.g., adr-20251124-database.md
```

**Step 4: Update ADR frontmatter example**

Change:
```yaml
### ADR
```yaml
---
id: adr-001
title: Use PostgreSQL for persistence
status: accepted
date: 2025-01-15
---
```
```

To:
```yaml
### ADR
```yaml
---
id: adr-20250115-postgresql
title: Use PostgreSQL for persistence
status: accepted
date: 2025-01-15
---
```
```

**Step 5: Update Anchor Format table**

Change:
```markdown
| ADR | `{#adr-nnn-*}` | `{#adr-001-decision}` |
```

To:
```markdown
| ADR | `{#adr-YYYYMMDD-*}` | `{#adr-20251124-decision}` |
```

**Step 6: Update Search Patterns bash section**

Change:
```bash
# Find all ADRs
ls .c3/adr/adr-*.md
```

To:
```bash
# Find all ADRs
ls .c3/adr/adr-[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]-*.md

# Find ADRs by year
ls .c3/adr/adr-2025*.md
```

**Step 7: Commit**

```bash
git add references/v3-structure.md
git commit -m "docs: update v3-structure reference for date-based ADRs"
```

---

### Task 7: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Update Current Version section**

Change:
```markdown
### Current Version

Check `VERSION` file for current version (currently: 3).
Check `MIGRATIONS.md` for migration history and format.
```

To:
```markdown
### Current Version

Check `VERSION` file for current version (currently: `20251124-adr-date-naming`).
Check `migrations/` directory for migration history (files sorted lexicographically).
```

**Step 2: Update Migration Checklist - change MIGRATIONS.md references**

Change:
```markdown
- [ ] MIGRATIONS.md has new version section with:
  - [ ] Changes (human-readable)
  - [ ] Transforms (patterns and replacements)
  - [ ] Verification (how to confirm success)
```

To:
```markdown
- [ ] New migration file in `migrations/YYYYMMDD-slug.md` with:
  - [ ] Changes (human-readable)
  - [ ] Transforms (patterns and replacements)
  - [ ] Verification (how to confirm success)
- [ ] Migration slug is unique (no duplicates in `migrations/`)
```

**Step 3: Update Plugin Structure**

Change:
```markdown
├── VERSION              # Current version number
├── MIGRATIONS.md        # Migration specifications
```

To:
```markdown
├── VERSION              # Current version (YYYYMMDD-slug format)
├── migrations/          # Individual migration files
```

**Step 4: Update Development Workflow**

Change:
```markdown
4. If migration needed: update VERSION, MIGRATIONS.md
```

To:
```markdown
4. If migration needed: update VERSION, create `migrations/YYYYMMDD-slug.md`
```

**Step 5: Add Version Slug Uniqueness section after Migration Checklist**

Add:
```markdown
### Version Slug Uniqueness Rule

Migration slugs must be unique across all migrations. Reviewers must reject PRs with duplicate version slugs.

**Before creating a migration, verify uniqueness:**
```bash
ls migrations/ | grep -c "YYYYMMDD-your-slug"  # Should return 0
```

**Naming format:** `YYYYMMDD-descriptive-slug`
- Date: The date the migration is created
- Slug: Short, descriptive, lowercase, hyphenated

**Examples:**
- `20251124-adr-date-naming` - ADR naming convention change
- `20251125-component-validation` - Component validation rules
```

**Step 6: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md for date-based versioning"
```

---

### Task 8: Final verification and cleanup

**Step 1: Verify all changes**

```bash
# Check VERSION content
cat VERSION
# Expected: 20251124-adr-date-naming

# Check migrations directory
ls migrations/
# Expected: 00000000-baseline.md  20251124-adr-date-naming.md

# Verify MIGRATIONS.md is deleted
[ ! -f MIGRATIONS.md ] && echo "MIGRATIONS.md deleted"

# Grep for any remaining old ADR patterns in skills
grep -r 'adr-\[0-9\]{3}' skills/ references/ || echo "No old patterns found"
grep -r 'adr-001\|adr-002\|adr-003' skills/ references/ || echo "No old examples found"
```

**Step 2: Run git status to verify all changes**

```bash
git status
```

**Step 3: Create final summary commit if needed**

```bash
git log --oneline -10
```
