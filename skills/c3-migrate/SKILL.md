---
name: c3-migrate
description: Migrate .c3/ documentation to current skill version - reads VERSION, compares against MIGRATIONS.md, executes transforms in batches
---

# C3 Migration Skill

## Overview

Migrate project `.c3/` documentation from older versions to current skill version.

**Core principle:** Explicit migration triggered by user. Show plan, get confirmation, execute in batches.

**Announce at start:** "I'm using the c3-migrate skill to upgrade your C3 documentation."

## Quick Reference

| Phase | Key Activities | Output |
|-------|---------------|--------|
| **1. Detect** | Read version, compare to current | Version gap identified |
| **2. Plan** | Parse MIGRATIONS.md, scan files | Migration plan |
| **3. Confirm** | Present changes to user | User approval |
| **4. Execute** | Apply transforms in batches | Updated files |
| **5. Finalize** | Update version, suggest TOC rebuild | Migration complete |

---

## Phase 1: Detect Version

### Read Project Version

```bash
# Check frontmatter first (v3), then VERSION file (v1/v2)
if grep -q '^c3-version:' .c3/README.md 2>/dev/null; then
    PROJECT_VERSION=$(grep '^c3-version:' .c3/README.md | sed 's/c3-version: *//')
elif [ -f ".c3/VERSION" ]; then
    PROJECT_VERSION=$(cat .c3/VERSION)
else
    PROJECT_VERSION=0
fi
```

**Version storage:**
| Format | Location |
|--------|----------|
| v1/v2 | `.c3/VERSION` file |
| v3+ | `c3-version:` in `.c3/README.md` frontmatter |

### Compare Versions

| Condition | Action |
|-----------|--------|
| PROJECT == SKILL | "Already current, no migration needed." Stop. |
| PROJECT > SKILL | "Project newer than skill." Stop. |
| PROJECT < SKILL | Continue to Phase 2 |

---

## Phase 2: Build Migration Plan

1. Parse `MIGRATIONS.md` from plugin directory
2. For each version from `PROJECT + 1` to `SKILL`:
   - Extract transforms section
   - Parse patterns and file globs
3. Scan `.c3/` for affected files
4. Build plan summary

### Plan Format

```markdown
## Migration Plan: v{FROM} → v{TO}

### Version {N} transforms:
- {PATTERN_DESC}: {FILE_COUNT} files

### Batches:
- Batch 1: file1.md, file2.md
- Batch 2: file3.md, file4.md
```

---

## Phase 3: Confirm with User

> "I'll migrate your `.c3/` documentation from v{FROM} to v{TO}.
>
> Changes:
> - {CHANGE_1}
> - {CHANGE_2}
>
> Files affected: {N}
>
> Proceed? [y/n]"

If declined, stop.

---

## Phase 4: Execute Migration

### Batch Processing

Process 3-5 files per batch for trackability.

For each batch:
1. Apply transforms to files
2. Report progress: `Batch 1/3 complete: 3 files updated`

### Error Handling

| Error | Action |
|-------|--------|
| Pattern doesn't match | Log warning, continue |
| File read error | Stop batch, report |

---

## Phase 5: Finalize

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

### Suggest TOC Rebuild

> "Migration complete: v{FROM} → v{TO}
>
> Run `.c3/scripts/build-toc.sh` to refresh TOC."

---

## V1 → V2 Migration

### Transforms

1. **Flatten components:** Move from `components/{container}/` to `components/`
2. **Rename context:** `CTX-system-overview.md` → `README.md`
3. **Update links:** Remove container subfolder from paths

### Verification

```bash
# No nested component directories
[ $(find .c3/components -mindepth 1 -type d | wc -l) -eq 0 ]

# README.md has correct id
grep -q '^id: context$' .c3/README.md
```

---

## V2 → V3 Migration

### Transforms

1. **Convert containers to folders:** `containers/C3-1-*.md` → `c3-1-*/README.md`
2. **Move components into containers:** `components/C3-101-*.md` → `c3-1-*/c3-101-*.md`
3. **Update context:** `id: context` → `id: c3-0`, add `c3-version: 3`
4. **Lowercase ADRs:** `ADR-001-*.md` → `adr-001-*.md`

### Verification

```bash
# No containers/ or components/ directories
[ ! -d ".c3/containers" ] && [ ! -d ".c3/components" ]

# Context has c3-0 and c3-version
grep -q '^id: c3-0$' .c3/README.md
grep -q '^c3-version: 3$' .c3/README.md

# All lowercase
! find .c3 -name "C3-*" -o -name "ADR-*" | grep -q .
```

---

## Red Flags

| Rationalization | Counter |
|-----------------|---------|
| "I'll migrate without asking" | Always confirm with user first |
| "I'll do all files at once" | Batch for trackability |
| "Pattern didn't match, skip silently" | Log warnings |
| "Version update not critical" | Always update on success |

## Related

- [v3-structure.md](../../references/v3-structure.md)
- [MIGRATIONS.md](../../MIGRATIONS.md)
