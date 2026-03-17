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

### Step 4: Validate plugin.json (CRITICAL)

**Claude Code uses auto-discovery for plugin components. Explicit path declarations break plugin loading.**

Read `.claude-plugin/plugin.json` and **REMOVE** these fields if present:
- `"commands": "..."`
- `"skills": "..."`
- `"agents": "..."`
- `"hooks": "..."`

These fields cause the plugin to not load properly. Auto-discovery finds components automatically from standard directories (`commands/`, `skills/`, `agents/`, `hooks/`).

### Step 5: Update Version Files

Read, then update **all three** files with the new version:

1. `skills/c3/bin/VERSION` - Replace content with just `A.B.C` (no quotes, no newline prefix) — **CI and c3x.sh read this to detect releases and select binaries**
2. `.claude-plugin/plugin.json` - Update `"version": "X.Y.Z"` (and ensure no explicit paths per Step 4)
3. `.claude-plugin/marketplace.json` - Update `"version": "X.Y.Z"` in the plugins array

### Step 6: Update README.md

Read `README.md` and update it to reflect any new features, changed commands, or modified `.c3/` structure introduced in this release. Focus on:

- **What You Get** table: any new `/c3` operations or changed behavior
- **`c3x` CLI** command list: new commands, removed commands, changed flags
- **The `.c3/` Directory** tree: new files or changed structure
- **Layer validation table**: if `check` behavior changed

Keep it concise — README is user-facing marketing, not a changelog. Only update what's visibly different to a new user.

### Step 7: Update CHANGELOG.md

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

### Step 7: Summary

After updates, show:
1. Version change: `X.Y.Z` → `A.B.C`
2. Files modified
3. Changelog entry preview
4. Remind user to commit and push to **dev** (CI handles the rest):
   ```bash
   git add skills/c3/bin/VERSION .claude-plugin/plugin.json .claude-plugin/marketplace.json CHANGELOG.md README.md
   git commit -m "chore: release vA.B.C"
   git push origin dev
   ```

**What CI does automatically on push to dev:**
- Runs `go test ./...`
- Cross-compiles Go CLI for 4 targets (linux/darwin × amd64/arm64)
- Merges dev → main (with binaries force-added — they're gitignored on dev)
- Detects new VERSION → creates git tag `vA.B.C` + GitHub Release with plugin zip

**Do NOT manually create git tags** — CI owns that.
