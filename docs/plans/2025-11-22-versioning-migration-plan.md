# C3 Versioning and Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add versioning and migration capabilities to C3 skill so projects can be upgraded when skill evolves.

**Architecture:** Sequential integer versions stored in `.c3/VERSION`. Single `MIGRATIONS.md` defines all transforms. Explicit `/c3-migrate` command triggers migration with batch execution.

**Tech Stack:** Markdown skills, bash scripts, Claude subagents for parallel file transforms.

---

## Task 1: Create Plugin VERSION File

**Files:**
- Create: `VERSION`

**Step 1: Create VERSION file**

Create `/home/lagz0ne/c3-design/VERSION`:
```
1
```

**Step 2: Verify file exists**

Run: `cat VERSION`
Expected: `1`

**Step 3: Commit**

```bash
git add VERSION
git commit -m "chore: add VERSION file for skill versioning"
```

---

## Task 2: Create MIGRATIONS.md

**Files:**
- Create: `MIGRATIONS.md`

**Step 1: Create MIGRATIONS.md**

Create `/home/lagz0ne/c3-design/MIGRATIONS.md`:
```markdown
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

\`\`\`markdown
## Version N (from N-1)

### Changes
- [Human-readable description of what changed]

### Transforms
- **Pattern:** [regex pattern to find]
- **Replace:** [replacement pattern]
- **Files:** [glob pattern for affected files]

### Verification
- [Checks to confirm migration succeeded]
\`\`\`
```

**Step 2: Verify file content**

Run: `head -20 MIGRATIONS.md`
Expected: Header and Version 1 section visible

**Step 3: Commit**

```bash
git add MIGRATIONS.md
git commit -m "docs: add MIGRATIONS.md for version transforms"
```

---

## Task 3: Create c3-migrate Skill

**Files:**
- Create: `skills/c3-migrate/SKILL.md`

**Step 1: Create skills directory**

Run: `mkdir -p skills/c3-migrate`

**Step 2: Create SKILL.md**

Create `/home/lagz0ne/c3-design/skills/c3-migrate/SKILL.md`:
```markdown
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
> Recommended: Run `.c3/scripts/build-toc.sh` to refresh the table of contents."

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
```

**Step 3: Verify skill file**

Run: `head -30 skills/c3-migrate/SKILL.md`
Expected: Frontmatter and Overview section visible

**Step 4: Commit**

```bash
git add skills/c3-migrate/SKILL.md
git commit -m "feat: add c3-migrate skill for version migration"
```

---

## Task 4: Create c3-migrate Command

**Files:**
- Create: `commands/c3-migrate.md`

**Step 1: Create command file**

Create `/home/lagz0ne/c3-design/commands/c3-migrate.md`:
```markdown
---
description: Migrate .c3/ documentation to current skill version
---

Migrate this project's C3 documentation to the current skill version.

Use the `c3-migrate` skill to:
1. Detect current project version from `.c3/VERSION`
2. Compare against current skill version
3. Build and present migration plan
4. Execute transforms with user confirmation
```

**Step 2: Verify command file**

Run: `cat commands/c3-migrate.md`
Expected: Full content visible

**Step 3: Commit**

```bash
git add commands/c3-migrate.md
git commit -m "feat: add /c3-migrate slash command"
```

---

## Task 5: Update c3-adopt to Create VERSION

**Files:**
- Modify: `skills/c3-adopt/SKILL.md`

**Step 1: Read current c3-adopt skill**

Read the file to find the scaffolding section (Phase 1: Establish).

**Step 2: Add VERSION creation to scaffolding**

In Phase 1, after creating directories, add VERSION file creation.

Find this section:
```markdown
### Create Scaffolding

If proceeding with new documentation:

```bash
mkdir -p .c3/{containers,components,adr,scripts}
```
```

Add after the mkdir command:
```markdown
### Create Scaffolding

If proceeding with new documentation:

```bash
mkdir -p .c3/{containers,components,adr,scripts}
```

Create VERSION file with current skill version:
```bash
# Read current version from plugin
SKILL_VERSION=$(cat /path/to/c3-skill/VERSION)
echo "$SKILL_VERSION" > .c3/VERSION
```
```

**Step 3: Verify edit**

Run: `grep -A5 "Create Scaffolding" skills/c3-adopt/SKILL.md`
Expected: VERSION creation visible

**Step 4: Commit**

```bash
git add skills/c3-adopt/SKILL.md
git commit -m "feat: c3-adopt creates VERSION file on init"
```

---

## Task 6: Update c3-adopt to Detect Missing VERSION

**Files:**
- Modify: `skills/c3-adopt/SKILL.md`

**Step 1: Add version detection to prerequisites**

Find the "Check Prerequisites" section and add VERSION detection.

After the existing `.c3/` check, add:
```markdown
### Check Version

If `.c3/` exists but `.c3/VERSION` is missing:

> "I found existing `.c3/` documentation but no VERSION file.
> This may be from an older version of the C3 skill.
>
> After review, consider running `/c3-migrate` to update to current format."
```

**Step 2: Verify edit**

Run: `grep -A8 "Check Version" skills/c3-adopt/SKILL.md`
Expected: Version check section visible

**Step 3: Commit**

```bash
git add skills/c3-adopt/SKILL.md
git commit -m "feat: c3-adopt warns about missing VERSION file"
```

---

## Task 7: Update c3-init Command

**Files:**
- Modify: `commands/c3-init.md`

**Step 1: Read current command**

Read the file to understand current content.

**Step 2: Update to mention VERSION**

Update `/home/lagz0ne/c3-design/commands/c3-init.md`:
```markdown
---
description: Initialize C3 documentation structure from scratch
---

Create `.c3/` directory structure with VERSION file and begin system design.

Use the `c3-adopt` skill for fresh initialization, which will:
1. Create `.c3/` directory structure
2. Set VERSION to current skill version
3. Guide through Context, Container, and Component discovery
```

**Step 3: Verify edit**

Run: `cat commands/c3-init.md`
Expected: Updated content with VERSION mention

**Step 4: Commit**

```bash
git add commands/c3-init.md
git commit -m "docs: update c3-init to mention VERSION file"
```

---

## Task 8: Update README

**Files:**
- Modify: `README.md`

**Step 1: Add versioning section**

Find the "Commands" section and add versioning documentation.

Add new section after "Commands":
```markdown
## Versioning

C3-Skill uses sequential version numbers to track documentation format changes.

### Version File

Each project using C3 has a `.c3/VERSION` file containing the version number:
```
1
```

### Migration

When the skill evolves, run `/c3-migrate` to upgrade your documentation:
```
/c3-migrate
```

This will:
1. Detect your current version
2. Show what changes are needed
3. Apply transforms with your confirmation
```

**Step 2: Update commands list**

Add `/c3-migrate` to the commands list:
```markdown
### Commands

- `/c3` - Design or update architecture (main command)
- `/c3-init` - Initialize `.c3/` structure from scratch
- `/c3-migrate` - Migrate documentation to current skill version
```

**Step 3: Verify edits**

Run: `grep -A10 "Versioning" README.md`
Expected: New versioning section visible

**Step 4: Commit**

```bash
git add README.md
git commit -m "docs: add versioning and migration documentation to README"
```

---

## Task 9: Final Verification

**Step 1: Verify all new files exist**

Run:
```bash
ls -la VERSION MIGRATIONS.md skills/c3-migrate/SKILL.md commands/c3-migrate.md
```
Expected: All 4 files exist

**Step 2: Verify git history**

Run: `git log --oneline -10`
Expected: 8 new commits for versioning/migration

**Step 3: Run TOC build (if applicable)**

Run: `.c3/scripts/build-toc.sh` (if .c3 exists in this repo)

---

## Summary

| Task | File | Action |
|------|------|--------|
| 1 | `VERSION` | Create with `1` |
| 2 | `MIGRATIONS.md` | Create with v1 baseline |
| 3 | `skills/c3-migrate/SKILL.md` | Create migration skill |
| 4 | `commands/c3-migrate.md` | Create slash command |
| 5 | `skills/c3-adopt/SKILL.md` | Add VERSION creation |
| 6 | `skills/c3-adopt/SKILL.md` | Add VERSION detection |
| 7 | `commands/c3-init.md` | Update to mention VERSION |
| 8 | `README.md` | Add versioning docs |
| 9 | - | Final verification |
