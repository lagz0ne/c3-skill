# C3 Skill Development Guidelines

## Migration Awareness

**CRITICAL:** When making changes to this plugin, assess whether they require user migration.

### Before Merging Any Change

Ask yourself:
1. Does this change affect the structure of user `.c3/` directories?
2. Does this change the ID patterns (c3-0, c3-N, c3-NNN, adr-YYYYMMDD-slug)?
3. Does this change frontmatter requirements?
4. Does this change file naming conventions?

**If YES to any:**
- Update `VERSION` file (use `YYYYMMDD-slug` format)
- Create migration file in `migrations/YYYYMMDD-slug.md`
- Update `c3-migrate` skill if transforms are needed
- **Require reviewer confirmation before merge**

**If NO to all:**
- No migration needed (skill-internal changes only)

### Types of Changes

| Change Type | Migration Needed? | Example |
|-------------|-------------------|---------|
| Skill file size/organization | No | Slimming skills, extracting to references |
| New reference files in plugin | No | Adding `references/*.md` |
| Documentation updates | No | README, design docs |
| New optional skill | No | Adding a new skill |
| **File structure changes** | **YES** | Changing where containers/components live |
| **ID pattern changes** | **YES** | Changing from C3-1 to c3-1 |
| **Frontmatter changes** | **YES** | Adding required fields |
| **Naming convention changes** | **YES** | Uppercase to lowercase |

### Current Version

Check `VERSION` file for current version (currently: `20251124-adr-date-naming`).
Check `migrations/` directory for migration history (files sorted lexicographically).

### Migration Checklist for Reviewers

When reviewing PRs that claim to need migration:

- [ ] VERSION file incremented
- [ ] New migration file in `migrations/YYYYMMDD-slug.md` with:
  - [ ] Changes (human-readable)
  - [ ] Transforms (patterns and replacements)
  - [ ] Verification (how to confirm success)
- [ ] Migration slug is unique (no duplicates in `migrations/`)
- [ ] c3-migrate skill updated if transforms are complex
- [ ] README.md updated if user-facing behavior changes (ID patterns, naming conventions, examples)
- [ ] references/v3-structure.md updated if structure patterns change

When reviewing PRs that claim NO migration needed:

- [ ] Confirm changes are purely skill-internal
- [ ] No file structure changes
- [ ] No ID/naming pattern changes
- [ ] No frontmatter changes

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

### Files to Update When Patterns Change

When changing ID patterns, naming conventions, or structure:

| File | What to Update |
|------|----------------|
| `README.md` | Examples, documentation structure, ID format descriptions |
| `references/v3-structure.md` | ID patterns, file paths, frontmatter examples |
| `skills/c3-naming/SKILL.md` | Naming patterns, search patterns, examples |
| `skills/c3-migrate/SKILL.md` | Version comparison, transforms, verification |
| `skills/c3-locate/SKILL.md` | ID resolution patterns, lookup examples |

**Tip:** Search for old patterns across all files before marking migration complete:
```bash
grep -r 'old-pattern' skills/ references/ README.md
```

## Plugin Structure

```
c3-design/
├── .claude-plugin/      # Plugin manifests only
├── skills/              # Skill definitions (SKILL.md files)
├── commands/            # Slash commands
├── references/          # Shared reference docs for skills
├── docs/plans/          # Design documents
├── .c3/                 # Plugin's own C3 documentation
├── VERSION              # Current version (YYYYMMDD-slug format)
├── migrations/          # Individual migration files
└── CLAUDE.md            # This file
```

## Development Workflow

1. Create design doc in `docs/plans/` for significant changes
2. Assess migration impact (see above)
3. Make changes
4. If migration needed: update VERSION, create `migrations/YYYYMMDD-slug.md`
5. Commit with descriptive message
6. Request review with migration assessment noted

## Feature Evaluation

**IMPORTANT:** Continuously evaluate features as you work on this plugin.

### When Adding Features

Ask yourself:
1. Does this feature overlap with an existing skill? → Consider merging
2. Is this feature used by multiple skills? → Extract to `references/`
3. Does this feature need a command entry point? → Add to `commands/`

### When Working on Existing Features

Ask yourself:
1. Is this skill still useful? → If not, consider deprecating
2. Is this skill too complex? → Consider splitting
3. Are there patterns repeated across skills? → Extract to shared reference
4. Is the skill description accurate? → Update if behavior changed

### Feature Inventory

Periodically review:
- `skills/` - Are all skills still relevant and well-scoped?
- `commands/` - Do commands map clearly to user intents?
- `references/` - Is shared content factored out appropriately?
- `README.md` - Does it reflect current capabilities?

### Red Flags

- Skill that duplicates another skill's purpose
- Command without a corresponding skill
- Reference file not used by any skill
- Feature documented in README but not implemented
- Implemented feature not documented in README
