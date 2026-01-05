---
description: Bump version and generate changelog based on changes
argument-hint: [major|minor|patch]
allowed-tools: Bash(git:*), Read, Edit, Write, AskUserQuestion
---

# C3 Plugin Release

## Arguments

$ARGUMENTS

## Current State

- Current version: !`cat "$CLAUDE_PROJECT_DIR/.claude-plugin/plugin.json" | grep '"version"' | head -1 | sed 's/.*"version": "\([^"]*\)".*/\1/'`
- Last git tag: !`cd "$CLAUDE_PROJECT_DIR" && git describe --tags --abbrev=0 2>/dev/null || echo "none"`

## Instructions

You are preparing a release for the C3 plugin. Follow these steps:

### Step 1: Gather Change Information

Run these git commands to understand what changed:

```bash
# Get commits since last tag (or all if no tags)
LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [ -n "$LAST_TAG" ]; then
    git log "$LAST_TAG"..HEAD --oneline --no-merges
else
    git log --oneline --no-merges -20
fi
```

Also check the diff summary:
```bash
if [ -n "$LAST_TAG" ]; then
    git diff "$LAST_TAG" --stat | tail -5
fi
```

### Step 2: Determine Version Bump

Analyze the commits using conventional commit prefixes:

| Prefix | Bump Type | Examples |
|--------|-----------|----------|
| `feat` | **minor** | New feature, new skill, new command |
| `fix` | **patch** | Bug fix, typo correction |
| `docs` | **patch** | Documentation only |
| `refactor` | **patch** | Code restructure, no behavior change |
| `chore` | **patch** | Maintenance, dependency updates |
| `BREAKING CHANGE` or exclamation suffix | **major** | Breaking changes to skill behavior |

**Decision rules:**
1. If user specified `major`, `minor`, or `patch` in $ARGUMENTS, use that
2. Otherwise, analyze commits and recommend based on:
   - Any `feat:` commits → minor
   - Any breaking changes → major
   - Only fixes/docs/chores → patch

### Step 3: Confirm with User

Use AskUserQuestion to confirm:
- The determined version bump (showing current → new version)
- Summary of changes to include in changelog

### Step 4: Update Version Files

Read, then update both files with the new version:

1. `.claude-plugin/plugin.json` - Update `"version": "X.Y.Z"`
2. `.claude-plugin/marketplace.json` - Update `"version": "X.Y.Z"` in the plugins array

### Step 5: Update CHANGELOG.md

Read CHANGELOG.md (create if doesn't exist). Add a new entry at the top following this format:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features (from `feat:` commits)

### Changed
- Changes to existing features (from `refactor:`, updates)

### Fixed
- Bug fixes (from `fix:` commits)

### Documentation
- Doc updates (from `docs:` commits)
```

**Rules for changelog:**
- Only include sections that have entries
- Group related changes
- Write user-facing descriptions (not commit messages verbatim)
- Include the justification for the version bump type

### Step 6: Summary

After updates, show:
1. Version change: `X.Y.Z` → `A.B.C`
2. Files modified
3. Changelog entry preview
4. Remind user to commit these changes:
   ```
   git add .claude-plugin/plugin.json .claude-plugin/marketplace.json CHANGELOG.md
   git commit -m "chore: release vA.B.C"
   git tag vA.B.C
   ```
