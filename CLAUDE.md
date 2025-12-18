# C3 Skill Development Guidelines

> **STOP** - Before ANY action in this repo, you MUST run:
> ```
> Skill tool → superpowers-developing-for-claude-code:developing-claude-code-plugins
> Skill tool → superpowers-developing-for-claude-code:working-with-claude-code
> ```
> This is a Claude Code plugin. These skills are NON-NEGOTIABLE.

---

## MANDATORY: Claude Code Plugin Development Context

**THIS IS A CLAUDE CODE PLUGIN.** You MUST load plugin development context before ANY work in this repository.

### Required Skills - Load IMMEDIATELY

Before doing ANYTHING in this codebase, you MUST invoke these skills:

1. **`superpowers-developing-for-claude-code:developing-claude-code-plugins`**
   - Use for: Creating, modifying, testing, releasing, or maintaining plugins
   - Contains: Plugin lifecycle patterns, manifest structure, skill authoring

2. **`superpowers-developing-for-claude-code:working-with-claude-code`**
   - Use for: Understanding Claude Code CLI, hooks, MCP, configuration
   - Contains: Official documentation for all Claude Code features

### Activation Harness

**STOP. Before you do anything else, ask yourself:**

```
[ ] Have I loaded developing-claude-code-plugins skill?
[ ] Have I loaded working-with-claude-code skill?
```

**If NO to either:** Use the Skill tool NOW. Do not proceed without this context.

**Common rationalizations that mean you're about to fail:**
- "I already know how plugins work" → WRONG. Load the skill for current patterns.
- "This is just a small change" → WRONG. Small changes can break plugin loading.
- "I'll check if I need it" → WRONG. You need it. This is a plugin repo.
- "The CLAUDE.md has enough context" → WRONG. Skills have implementation details.
- "I remember from last time" → WRONG. Skills get updated. Load current version.

**Why this matters:**
- Plugin manifests have specific structure requirements
- Skills have naming conventions and file organization rules
- Commands must be registered correctly
- Breaking any of these = plugin fails to load for users

### When to Re-load Skills

Re-invoke these skills when:
- Starting a new session (context may have compacted)
- Switching between different types of work
- Encountering unexpected plugin behavior
- Before modifying `.claude-plugin/` manifests
- Before creating new skills or commands

---

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

Check `VERSION` file for current version (currently: `20251216-enforcement-harnesses`).
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
| `references/naming-conventions.md` | Naming patterns, search patterns, examples |
| `skills/c3-migrate/SKILL.md` | Version comparison, transforms, verification |
| `references/lookup-patterns.md` | ID resolution patterns, lookup examples |

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
2. Assess migration impact (see Migration Decision below)
3. Make changes
4. If migration needed: update VERSION, create `migrations/YYYYMMDD-slug.md`
5. Commit using conventional commits (see below)
6. Request review with migration assessment noted

### Conventional Commits

**Always use conventional commit format:**

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
| Type | When to Use |
|------|-------------|
| `feat` | New feature or capability |
| `fix` | Bug fix |
| `docs` | Documentation only changes |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `chore` | Maintenance tasks (version bumps, dependency updates) |
| `test` | Adding or correcting tests |

**Scopes:** `skills`, `references`, `commands`, `migrations`, or specific skill name

**Examples:**
```
feat(skills): add NO CODE enforcement harness to c3-component-design
fix(c3-audit): correct layer rule validation logic
docs(references): update v3-structure examples
chore: bump version to 20251216-enforcement-harnesses
refactor(c3-adopt): extract scaffolding to separate section
```

### Version Changes

**Three types of versions to track:**

| Version Type | Location | When to Update |
|--------------|----------|----------------|
| **Plugin version** | `.claude-plugin/plugin.json` + `marketplace.json` | Any user-visible change (features, fixes, harnesses) |
| **Skill version** | `VERSION` file | Any significant skill change |
| **C3 version** | User's `.c3/README.md` frontmatter | When user doc structure changes |

**Plugin version:** Semver in `.claude-plugin/*.json`. Bump for any user-visible change:
- **Major** (X.0.0): Breaking changes to skill behavior
- **Minor** (x.Y.0): New features, new harnesses, new skills
- **Patch** (x.y.Z): Bug fixes, wording improvements

**Skill version:** Update for any notable change (features, fixes, harness additions). Use `YYYYMMDD-slug` format.

**C3 version:** Only changes when user `.c3/` structure requirements change (rare).

### Migration Decision

**Ask these questions for EVERY change:**

```
1. Does this change how user .c3/ files are structured?     → Migration needed
2. Does this change ID patterns users must follow?          → Migration needed
3. Does this change required frontmatter fields?            → Migration needed
4. Does this change file naming conventions?                → Migration needed
5. Is this purely skill-internal (enforcement, wording)?    → No migration, but bump VERSION
```

**Decision matrix:**

| Change | Bump Plugin? | Bump VERSION? | Create Migration? | Update c3-migrate? |
|--------|--------------|---------------|-------------------|-------------------|
| Skill wording/enforcement | Minor | Yes | No | No |
| New reference file | Minor | Yes | No | No |
| New optional skill | Minor | Yes | No | No |
| Bug fix | Patch | No | No | No |
| File structure change | Minor | Yes | **Yes** | Yes |
| ID pattern change | Minor | Yes | **Yes** | Yes |
| Frontmatter change | Minor | Yes | **Yes** | Yes |
| Breaking behavior change | **Major** | Yes | **Yes** | Yes |

**When creating a migration file, include:**
1. **Changes** - Human-readable description
2. **Transforms** - Pattern replacements (if automatic migration possible)
3. **Verification** - Commands to confirm migration success

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

---

## C3 Architecture Overview

### The C3 Hierarchy

C3 = **Context → Container → Component**

```
Context (c3-0)           ← Eagle-eye introduction, defines WHAT exists
    │
    ├── Container (c3-1) ← Architectural command center, defines HOW things organize
    │       ├── Component (c3-101)  ← Implementation details
    │       └── Component (c3-102)
    │
    └── Container (c3-2)
            └── Component (c3-201)
```

**Key Principle:** Each layer inherits from above and defines contracts for below.

### Layer Responsibilities

| Layer | Abstraction | Key Questions | Changes Are... |
|-------|-------------|---------------|----------------|
| **Context** | Eagle-eye | What containers exist? How do they talk? | Rare, system-wide |
| **Container** | Command center | What components? What patterns? What tech? | Moderate, container-wide |
| **Component** | Implementation | How does this work? What config? | Frequent, isolated |

---

## Skills Organization

### Core Design Skills (The Layer Skills)

These three skills handle the actual architectural documentation:

| Skill | Purpose | When Used |
|-------|---------|-----------|
| `c3-context-design` | Document Context level | Adding/changing containers, protocols |
| `c3-container-design` | Document Container level | Tech stack, components, patterns, diagrams |
| `c3-component-design` | Document Component level | Implementation details, config, code patterns |

**Design Principle:** Container is the "sweet spot" - it has the most architectural impact and is where strategic diagram decisions happen.

### Orchestration Skills

| Skill | Purpose | When Used |
|-------|---------|-----------|
| `c3-design` | Main entry point, orchestrates layer skills | Any architectural change |
| `c3-adopt` | Initialize C3 for any project (existing or fresh) | New project adoption |

### Utility Skills

| Skill | Purpose | When Used |
|-------|---------|-----------|
| `c3-config` | Manage settings.yaml | Configure project preferences |
| `c3-migrate` | Handle version migrations | After plugin updates |
| `c3-audit` | Verify docs match code reality | Post-implementation verification |

**Note:** Document lookup and naming conventions are implemented as reference patterns, not skills:
- `references/lookup-patterns.md` - ID-based document retrieval patterns
- `references/naming-conventions.md` - ID and naming convention patterns

---

## Design Principles

### 1. Settings & Defaults System

Each layer skill has:
- `defaults.md` - Built-in defaults (include/exclude, litmus test, diagrams)
- Loads `.c3/settings.yaml` at start
- Merges settings with defaults based on `useDefaults` flag

**Merge Logic:**
```
defaults.md → loaded first
settings.yaml → merged/overrides based on useDefaults
                ↓
         Active Configuration
```

**When modifying layer skills:** Ensure the "Load Settings & Defaults" section exists and is applied throughout reasoning.

### 2. Litmus Tests for Layer Placement

Each layer has a litmus test to determine if content belongs there:

| Layer | Litmus Test |
|-------|-------------|
| **Context** | "Would changing this require coordinating multiple containers or external parties?" |
| **Container** | "Is this about WHAT this container does and WITH WHAT, not HOW internally?" |
| **Component** | "Could a developer implement this from the documentation?" |

### 3. Diagram Philosophy

**Key Insight:** There's no one-size-fits-all diagram. Each container needs different diagrams based on complexity.

**Container level has highest diagram ROI** because:
- Full context from above
- Controls component organization below
- Where readers understand the architecture

**Diagram Decision Framework (in c3-container-design):**
1. Characterize container (simple/moderate/complex)
2. Evaluate each diagram type (clarity + value + cost)
3. Check combinations for redundancy
4. Document decisions with justifications

### 4. Inter-Container Interactions

**Key Insight:** Inter-container communication is NOT Container-to-Container. It's Component-to-Component, mediated by Container.

```
Container A "talks to" Container B means:
- Component in A (e.g., "B Client") initiates
- Component in B (e.g., "Request Handler") receives
- The protocol comes from Context
- Container A documents its client component
- Container B documents its handler component
```

### 5. ADR Verification

ADRs have:
- `Changes Across Layers` - What should change in docs
- `Verification Checklist` - What to check in code

`c3-audit` verifies:
1. Are ADR doc changes actually made?
2. Does code match what docs describe?
3. Are verification items satisfied?

---

## Key Files to Know

### Layer Skills (modify for architectural behavior)

| File | Purpose |
|------|---------|
| `skills/c3-context-design/SKILL.md` | Context level exploration |
| `skills/c3-context-design/defaults.md` | Context defaults (include/exclude/litmus/diagrams) |
| `skills/c3-container-design/SKILL.md` | Container level exploration (heaviest skill) |
| `skills/c3-container-design/defaults.md` | Container defaults |
| `skills/c3-component-design/SKILL.md` | Component level exploration |
| `skills/c3-component-design/defaults.md` | Component defaults |

### Reference Files (shared knowledge)

| File | Purpose |
|------|---------|
| `references/hierarchy-model.md` | C3 layer inheritance diagram |
| `references/v3-structure.md` | File paths, ID patterns, frontmatter |
| `references/diagram-patterns.md` | Diagram syntax and examples |
| `references/adr-template.md` | ADR structure |
| `references/archetype-hints.md` | Container type patterns |

### Settings & Config

| File | Purpose |
|------|---------|
| `skills/c3-config/SKILL.md` | Settings management skill |
| `.c3/settings.yaml` (user's project) | Project-specific settings |

---

## Common Modification Scenarios

### "I want to change what goes in Context vs Container"

1. Edit `skills/c3-context-design/defaults.md` (include/exclude lists)
2. Edit `skills/c3-container-design/defaults.md` (include/exclude lists)
3. Update litmus tests if needed
4. No migration needed (skill-internal change)

### "I want to change diagram guidance"

1. For defaults: Edit layer skill's `defaults.md`
2. For decision framework: Edit `skills/c3-container-design/SKILL.md` Phase 4
3. For syntax examples: Edit `references/diagram-patterns.md`

### "I want to add a new setting"

1. Add to `skills/c3-config/SKILL.md` settings structure
2. Update layer skills to read and apply the new setting
3. Update `references/v3-structure.md` if it affects file structure

### "I want to change how ADR verification works"

1. Edit `skills/c3-audit/SKILL.md` (Mode 2: ADR Audit)
2. If changing ADR structure: also edit `references/adr-template.md`

### "I want to change the extended thinking in a layer skill"

1. Edit the relevant `<extended_thinking>` block in the SKILL.md
2. Ensure the thinking aligns with settings/defaults loading
3. No migration needed

---

## Session History: 2025-11-27 Changes

### c3-container-design (Major Enhancement)
- Positioned as "architectural command center"
- Added extensive diagram decision framework with extended thinking
- Clarified inter-container interactions are component-mediated
- Added diagram type analysis (capabilities/limitations of each type)
- Added diagram combination patterns and anti-patterns
- Added settings loading section

### c3-context-design (Simplified)
- Reframed as "eagle-eye introduction"
- Focused on two core jobs: container inventory + container interactions
- Removed heavy extended thinking (lighter touch for this layer)
- Added settings loading section

### c3-audit (Enhanced)
- Added audit modes: Full, ADR, Container, Quick
- Added ADR verification workflow (verify doc changes + code conformance)
- Added verification checklist results tracking

### All Layer Skills
- Added "Load Settings & Defaults" section
- Settings are loaded at start and applied throughout reasoning
- Merge logic: defaults.md + settings.yaml based on useDefaults flag
