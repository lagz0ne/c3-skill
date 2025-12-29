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
feat(skills): add inventory-first model to c3-structure
fix(c3-implementation): correct component template sections
docs(references): update v3-structure examples
chore: bump version to 20251229-skill-consolidation
refactor(agents): extract audit details to reference file
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

## C3 Philosophy (For Maintainers)

**Why does this section exist?** Skills embody the philosophy. When modifying skills, use these principles to evaluate whether changes maintain coherence.

### Why Three Layers?

Three layers exist because decisions require **three types of situational awareness**:

| Layer | Awareness Type | Question It Answers |
|-------|----------------|---------------------|
| **Context** | System awareness | "WHERE am I? What boundaries exist?" |
| **Container** | Relationship awareness | "WHAT is connected? What depends on what?" |
| **Component** | Implementation awareness | "HOW does this work?" |

**Two layers fail:** No place to understand relationships without implementation details.
**Four+ layers fail:** Navigation overhead exceeds value.

**Skill implication:** Each layer skill should help users gain ONE type of awareness, then navigate to next layer if needed.

### Why Unidirectional Flow?

**Reading direction matters:** Context → Container → Component

```
WRONG: Jump to Component first
       → Miss boundaries (Context)
       → Miss relationships (Container)
       → Make changes that break things you didn't know existed

RIGHT: Start at Context, descend as needed
       → Calibrate scope
       → Understand connections
       → Act with full awareness
```

**Skill implication:** Layer skills enforce "load parent first" gates. This is not bureaucracy - it prevents costly mistakes.

### Why "Just Enough"?

C3 docs exist to enable **two types of decisions**:

| Decision | What User Is Doing |
|----------|-------------------|
| **ADAPT** | Add to / evolve the system |
| **ALTER** | Fix / modify current behavior |

**The test:** A layer is documented correctly when it enables ADAPT/ALTER decisions at that level - no more, no less.

| State | Symptom | Fix |
|-------|---------|-----|
| **Over-documented** | Info that doesn't help decisions at THIS level | Move to lower layer or delete |
| **Under-documented** | Missing info that blocks decisions | Add what's missing, nothing more |

**Skill implication:** When evaluating skill output, ask "Does this enable the user to decide, or does it just add words?"

### The Complete Cycle

```
C3 Docs (current state)
    ↓
Agent navigates (gather awareness)
    ↓
ADR created (decision recorded)
    ↓
Implementation plan (tactical)
    ↓
Code changes
    ↓
C3 Docs updated (new state) ←── cycle completes
```

**Drift:** When code diverges from docs, audit brings them back in sync.

**Skill implication:** ADRs bridge understanding to action. Layer skills update docs after ADR acceptance. Audit verifies alignment.

### Coherence Check for Skill Changes

Before modifying any skill, ask:

1. **Awareness type:** Does this skill help users gain the RIGHT type of awareness for its layer?
2. **Direction:** Does this skill enforce reading parent layers first?
3. **Just enough:** Does this skill produce output that enables decisions without over-documenting?
4. **Cycle position:** Does this skill know its role in the ADR → docs → audit cycle?

If any answer is "no" or "unclear," the change may break coherence.

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

### Core Design Skills (2 Skills)

| Skill | Purpose | When Used |
|-------|---------|-----------|
| `c3-structure` | System structure and relationships | Context + Container: inventory, interactions, tech stack |
| `c3-implementation` | Component conventions | How components work, hand-offs, edge cases |

**Design Principle:** Structure skill handles WHAT exists. Implementation skill handles HOW it works.

### Agent

The **c3 agent** (`agents/c3.md`) is the single entry point for all C3 work:
- Discovery and adoption
- Design and documentation
- Configuration
- Audit and verification

### Layer Mapping

| Layer | Skill |
|-------|-------|
| Context (c3-0) | `c3-structure` |
| Container (c3-N) | `c3-structure` |
| Component (c3-NNN) | `c3-implementation` |

**Note:** Reference patterns:
- `references/lookup-patterns.md` - ID-based document retrieval
- `references/naming-conventions.md` - ID and naming conventions
- `references/container-patterns.md` - Component inventory by relationship type

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

**Diagram Decision Framework (in c3-structure):**
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

The c3 skill's audit mode verifies:
1. Are ADR doc changes actually made?
2. Does code match what docs describe?
3. Are verification items satisfied?

---

## Key Files to Know

### Skills

| File | Purpose |
|------|---------|
| `skills/c3-structure/SKILL.md` | Context + Container structure |
| `skills/c3-structure/structure-template.md` | Templates for Context and Container |
| `skills/c3-implementation/SKILL.md` | Component implementation |
| `skills/c3-implementation/implementation-template.md` | Template for Component |

### Reference Files

| File | Purpose |
|------|---------|
| `references/core-principle.md` | The C3 principle (upper defines WHAT, lower implements HOW) |
| `references/v3-structure.md` | File paths, ID patterns, frontmatter |
| `references/diagram-patterns.md` | Diagram syntax and guidance |
| `references/container-patterns.md` | Component inventory by container type |
| `references/adr-template.md` | ADR structure |

### Settings & Config

| File | Purpose |
|------|---------|
| `skills/c3-config/SKILL.md` | Settings management skill |
| `.c3/settings.yaml` (user's project) | Project-specific settings |

---

## Common Modification Scenarios

### "I want to change what goes in Context vs Container"

1. Edit `skills/c3-structure/SKILL.md` (Context Level and Container Level sections)
2. Update litmus tests if needed
3. No migration needed (skill-internal change)

### "I want to change diagram guidance"

1. Edit `references/diagram-patterns.md`
2. No migration needed

### "I want to add a new setting"

1. Add to `skills/c3-config/SKILL.md` settings structure
2. Update skills to read and apply the new setting
3. Update `references/v3-structure.md` if it affects file structure

### "I want to change how ADR verification works"

1. Edit `agents/c3.md` (audit mode section)
2. If changing ADR structure: also edit `references/adr-template.md`

---

## Session History

### 2025-12-29: Major Restructuring

**Skills consolidated from 3 to 2:**
- `c3-context-design` + `c3-container-design` → `c3-structure`
- `c3-component-design` → `c3-implementation`

**Key changes:**
- Inventory-first model: Container inventory is source of truth; component docs appear when conventions mature
- Removed tech-specific guidance (trust model knowledge)
- ALTER/ADAPT focus: documents exist for future changes
- ~54% reduction in total content

**References consolidated:**
- Merged `diagram-decision-framework.md` into `diagram-patterns.md`
- Created `container-patterns.md` (slimmed from `container-archetypes.md`)
- Removed: `archetype-hints.md`, `socratic-method.md`, `hierarchy-model.md`
