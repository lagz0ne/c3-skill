# C3 Skill Development Guidelines

## Migration Awareness

**CRITICAL:** When making changes to this plugin, assess whether they require user migration.

### Before Merging Any Change

Ask yourself:
1. Does this change affect the structure of user `.c3/` directories?
2. Does this change the ID patterns (c3-0, c3-N, c3-NNN, adr-nnn)?
3. Does this change frontmatter requirements?
4. Does this change file naming conventions?

**If YES to any:**
- Update `VERSION` file (increment version number)
- Add migration section to `MIGRATIONS.md`
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

Check `VERSION` file for current version (currently: 3).
Check `MIGRATIONS.md` for migration history and format.

### Migration Checklist for Reviewers

When reviewing PRs that claim to need migration:

- [ ] VERSION file incremented
- [ ] MIGRATIONS.md has new version section with:
  - [ ] Changes (human-readable)
  - [ ] Transforms (patterns and replacements)
  - [ ] Verification (how to confirm success)
- [ ] c3-migrate skill updated if transforms are complex
- [ ] README updated if user-facing behavior changes

When reviewing PRs that claim NO migration needed:

- [ ] Confirm changes are purely skill-internal
- [ ] No file structure changes
- [ ] No ID/naming pattern changes
- [ ] No frontmatter changes

## Plugin Structure

```
c3-design/
├── .claude-plugin/      # Plugin manifests only
├── skills/              # Skill definitions (SKILL.md files)
├── commands/            # Slash commands
├── references/          # Shared reference docs for skills
├── docs/plans/          # Design documents
├── .c3/                 # Plugin's own C3 documentation
├── VERSION              # Current version number
├── MIGRATIONS.md        # Migration specifications
└── CLAUDE.md            # This file
```

## Development Workflow

1. Create design doc in `docs/plans/` for significant changes
2. Assess migration impact (see above)
3. Make changes
4. If migration needed: update VERSION, MIGRATIONS.md
5. Commit with descriptive message
6. Request review with migration assessment noted
