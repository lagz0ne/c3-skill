# C3 Skill Versioning and Migration Design

## Summary

Add versioning and migration capabilities to the C3 skill so projects using C3 can be upgraded when the skill evolves.

## Context

- C3 is a Claude Code plugin used across multiple repos
- When the skill evolves (naming conventions, frontmatter, heading IDs), existing projects have outdated formats
- Need a way to detect version gaps and migrate documents to current format

## Decisions

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Scope | Document format only | Focus on naming, frontmatter, heading IDs - not workflow changes |
| Version format | Sequential integers (1, 2, 3...) | Simple - each increment means migration needed |
| Version storage | `.c3/VERSION` dedicated file | Clean separation, easy to read/write |
| Migration definitions | Single `MIGRATIONS.md` | One file to maintain, all transforms in one place |
| Execution | Sequential batching (3-5 files) | Controlled parallelism, easier to track |
| Trigger | Explicit `/c3-migrate` command | User controls when migration happens |

## Design

### Version File Structure

**Location:** `.c3/VERSION` in each project

**Format:** Plain text, single integer
```
3
```

**Behavior:**
- Missing file = version 0 (unversioned/legacy)
- Created by `c3-init` with current skill version
- Updated by `c3-migrate` after successful migration

**In skill repo:** `VERSION` file at plugin root with current version

### MIGRATIONS.md Structure

**Location:** Plugin root (`/MIGRATIONS.md`)

**Format:**
```markdown
# C3 Skill Migrations

Current version: 4

## Version 4 (from 3)
### Changes
- Heading IDs: `{#doc-id-section}` → `{#section}` (simplified)

### Transforms
- **Pattern:** `{#(CTX|C3|ADR)-[^}]+-([^}]+)}`
- **Replace:** `{#$2}`
- **Files:** `*.md` in `.c3/`

## Version 3 (from 2)
### Changes
- Container naming: `CON-NNN-slug` → `C3-N-slug`
- Component naming: `COM-NNN-slug` → `C3-NNN-slug`

### Transforms
- **Pattern:** `CON-(\d+)-`
- **Replace:** `C3-$1-`
- **Files:** `.c3/containers/*.md`, `TOC.md`

## Version 2 (from 1)
### Changes
- Added required `summary` field in frontmatter

### Transforms
- **Check:** frontmatter has `summary:`
- **Action:** If missing, prompt user or add placeholder
- **Files:** All `.c3/*.md`
```

**Key points:**
- Each version section describes what changed and how to transform
- Transforms can be regex patterns, structural checks, or prompts
- Listed newest-first for readability

### c3-migrate Skill Workflow

**Trigger:** User runs `/c3-migrate` or invokes `c3-skill:c3-migrate`

**Flow:**
```
1. READ .c3/VERSION (default 0 if missing)
2. READ MIGRATIONS.md from plugin
3. COMPARE versions
   ├── Same → "Already at version N, no migration needed"
   └── Different → Continue
4. BUILD migration plan
   - Collect all transforms from current+1 to target
   - Scan .c3/ to find affected files
   - Group into batches of 3-5 files
5. PRESENT plan to user
   "Migrating v2 → v4:
    - 12 files need container renaming (CON→C3)
    - 8 files need heading ID updates
    Proceed? [y/n]"
6. EXECUTE batches
   - For each batch: spawn subagent with file list + transforms
   - Subagent applies transforms, reports results
   - Main skill tracks progress, shows completion
7. UPDATE .c3/VERSION to target version
8. SUGGEST: "Run .c3/scripts/build-toc.sh to refresh TOC"
```

**Subagent task example:**
```
"Apply these transforms to these 5 files:
 - Pattern: CON-(\d+)- → C3-$1-
 - Files: C3-1-backend.md, C3-2-frontend.md, ...
 Report: files changed, lines modified, any issues"
```

### Integration with Other Skills

**c3-init:**
- After creating `.c3/` structure, write `VERSION` file with current version

**c3-adopt:**
- When adopting existing project, detect if `.c3/VERSION` exists
- If missing → assume v0, inform user: "Run `/c3-migrate` to update to current format"

**Other skills (c3-design, c3-locate, etc.):**
- No changes required
- User responsible for running migration when needed

## New Artifacts

**In plugin repo:**
```
c3-design/
├── VERSION                    # Current skill version
├── MIGRATIONS.md              # All migration definitions
├── skills/
│   └── c3-migrate/
│       └── SKILL.md           # Migration skill
└── commands/
    └── c3-migrate.md          # Slash command
```

**In each project:**
```
project/.c3/
└── VERSION                    # Project's current version
```
