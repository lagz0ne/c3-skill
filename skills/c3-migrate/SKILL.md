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
| **1. Detect** | Read .c3/VERSION, compare to current | Version gap identified |
| **2. Plan** | Parse MIGRATIONS.md, scan files | Migration plan |
| **3. Confirm** | Present changes to user | User approval |
| **4. Execute** | Apply transforms in batches | Updated files |
| **5. Finalize** | Update VERSION, suggest TOC rebuild | Migration complete |

## Phase 1: Detect Version

### Read Project Version

```bash
# Check if .c3/ exists
if [ ! -d ".c3" ]; then
    echo "No .c3/ directory found. Nothing to migrate."
    exit 0
fi

# Read version (default 0 if missing)
if [ -f ".c3/VERSION" ]; then
    PROJECT_VERSION=$(cat .c3/VERSION)
else
    PROJECT_VERSION=0
fi
```

### Read Current Skill Version

The current version is defined in the plugin's `VERSION` file:
```bash
SKILL_VERSION=$(cat /path/to/c3-skill/VERSION)
```

### Compare Versions

| Condition | Action |
|-----------|--------|
| PROJECT_VERSION == SKILL_VERSION | "Already at version N, no migration needed." Stop. |
| PROJECT_VERSION > SKILL_VERSION | "Project version newer than skill. This shouldn't happen." Stop. |
| PROJECT_VERSION < SKILL_VERSION | Continue to Phase 2 |

## Phase 2: Build Migration Plan

### Parse MIGRATIONS.md

Read `MIGRATIONS.md` from the plugin directory.

For each version from `PROJECT_VERSION + 1` to `SKILL_VERSION`:
1. Extract the `### Transforms` section
2. Parse patterns, replacements, and file globs
3. Collect into migration steps

### Scan Affected Files

For each transform:
1. Glob for matching files in `.c3/`
2. Check which files contain the pattern
3. Count matches per file

### Build Plan Summary

```markdown
## Migration Plan: v{FROM} → v{TO}

### Version {N} transforms:
- {PATTERN_DESC}: {FILE_COUNT} files, {MATCH_COUNT} matches
  - file1.md (3 matches)
  - file2.md (1 match)

### Batches:
- Batch 1: file1.md, file2.md, file3.md
- Batch 2: file4.md, file5.md
```

## Phase 3: Confirm with User

Present the plan:

> "I'll migrate your `.c3/` documentation from version {FROM} to version {TO}.
>
> Changes:
> - {CHANGE_1}
> - {CHANGE_2}
>
> Files affected: {N}
>
> Proceed with migration? [y/n]"

If user declines, stop.

## Phase 4: Execute Migration

### Batch Processing

Process files in batches of 3-5 to balance parallelism and trackability.

For each batch:

1. **Dispatch subagent** with task:
   ```
   Apply these transforms to these files:

   Transforms:
   - Pattern: {REGEX}
     Replace: {REPLACEMENT}

   Files:
   - {FILE_1}
   - {FILE_2}
   - {FILE_3}

   Instructions:
   1. Read each file
   2. Apply pattern replacements
   3. Write updated content
   4. Report: filename, lines changed, any issues
   ```

2. **Wait for completion**

3. **Report progress**
   ```
   Batch 1/3 complete: 3 files updated
   - file1.md: 5 replacements
   - file2.md: 2 replacements
   - file3.md: 1 replacement
   ```

### Error Handling

| Error | Action |
|-------|--------|
| Pattern doesn't match | Log warning, continue |
| File read error | Stop batch, report to user |
| Subagent timeout | Retry once, then stop |

## Phase 5: Finalize

### Update VERSION

```bash
echo "{TARGET_VERSION}" > .c3/VERSION
```

### Suggest TOC Rebuild

> "Migration complete: v{FROM} → v{TO}
>
> {N} files updated across {BATCHES} batches.
>
> Recommended: Run `.c3/scripts/build-toc.sh` to refresh the table of contents.
> (Script missing? Copy from c3-skill plugin's `.c3/scripts/build-toc.sh`)"

### Verification Checklist

Run the `### Verification` checks from MIGRATIONS.md for the target version:
- Report pass/fail for each check
- If any fail, warn user

## Common Patterns

### Regex Transform Examples

| Change | Pattern | Replace |
|--------|---------|---------|
| CON→C3 prefix | `CON-(\d+)-` | `C3-$1-` |
| COM→C3 prefix | `COM-(\d+)-` | `C3-$1-` |
| Heading ID simplify | `\{#(CTX\|C3\|ADR)-[^}]+-([^}]+)\}` | `{#$2}` |
| Add frontmatter field | (structural) | Insert after `---` |

### Structural Transforms

For non-regex transforms (like adding required fields):

1. Parse file frontmatter
2. Check for required field
3. If missing, either:
   - Add with placeholder: `summary: "TODO: Add summary"`
   - Prompt user for value

## V1 → V2 Migration Details

### Detecting V1 Structure

```bash
# V1 has nested component directories
nested_dirs=$(find .c3/components -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)
if [ "$nested_dirs" -gt 0 ]; then
    echo "V1 structure detected: nested component directories"
fi
```

### Transform: Flatten Components

```bash
# Move all component files up one level
for container_dir in .c3/components/*/; do
    [ -d "$container_dir" ] || continue
    mv "$container_dir"*.md .c3/components/ 2>/dev/null || true
    rmdir "$container_dir" 2>/dev/null || true
done
```

### Transform: Rename Context

```bash
# Rename CTX-system-overview.md to README.md
if [ -f ".c3/CTX-system-overview.md" ]; then
    mv ".c3/CTX-system-overview.md" ".c3/README.md"
    # Update frontmatter id
    sed -i 's/^id: CTX-system-overview$/id: context/' ".c3/README.md"
fi
```

### Transform: Update Internal Links

```bash
# Update component links (remove container subfolder from path)
find .c3 -name "*.md" -exec sed -i \
    's|\](./components/[^/]*/\(C3-[0-9]\)|\](./components/\1|g' {} \;

# Update context links
find .c3 -name "*.md" -exec sed -i \
    's|CTX-system-overview\.md|README.md|g' {} \;
```

### V1→V2 Verification

After migration, verify:

```bash
# No nested component directories
[ $(find .c3/components -mindepth 1 -maxdepth 1 -type d | wc -l) -eq 0 ] || echo "FAIL: nested dirs remain"

# README.md exists with correct id
grep -q '^id: context$' .c3/README.md || echo "FAIL: README.md id not updated"

# VERSION updated
[ "$(cat .c3/VERSION)" = "2" ] || echo "FAIL: VERSION not updated"

# No broken component links
! grep -r '](./components/[^/]*/C3-' .c3/*.md .c3/**/*.md || echo "FAIL: old component links remain"
```

## Sub-Skill Usage

| Task | Tool |
|------|------|
| Batch file transforms | Task tool with subagent |
| File reading | Read tool |
| File writing | Edit or Write tool |
| Pattern matching | Grep tool |
| User confirmation | AskUserQuestion tool |

## Red Flags

| Rationalization | Counter |
|-----------------|---------|
| "I'll migrate without asking" | Always confirm with user first |
| "I'll do all files at once" | Batch in groups of 3-5 for trackability |
| "Pattern didn't match, skip silently" | Log warnings for transparency |
| "VERSION file not critical" | Always update VERSION on success |
