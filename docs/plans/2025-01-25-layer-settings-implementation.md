# Layer Settings Externalization Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Externalize hardcoded layer instructions from skills to configurable defaults.md files + settings.yaml customization.

**Architecture:** Each layer skill (context/container/component) gets a co-located defaults.md containing canonical include/exclude/litmus/diagrams. Skills read defaults.md + settings.yaml, merge based on `useDefaults` flag. The c3-config skill is updated to handle enhanced layer sections.

**Tech Stack:** Markdown skill files, YAML settings

**Design doc:** `docs/plans/2025-01-25-layer-settings-externalization-design.md`

---

## Task 1: Create Context Layer Defaults

**Files:**
- Create: `skills/c3-context-design/defaults.md`

**Step 1: Create the defaults file**

Create `skills/c3-context-design/defaults.md` with content extracted from SKILL.md:

```markdown
# Context Layer Defaults

## Include

| Element | Example |
|---------|---------|
| System boundary | "TaskFlow system includes..." |
| Actors | Users, Admin, External APIs |
| Container inventory | Links to container docs |
| Protocols | REST, gRPC, WebSocket |
| Cross-cutting concerns | Auth strategy, logging approach |
| Deployment topology | Cloud, multi-region |

## Exclude

| Element | Push To |
|---------|---------|
| Technology choices | Container |
| Middleware specifics | Container |
| API endpoints | Container |
| Configuration values | Component |
| Code examples | Component |

## Litmus Test

> "Would changing this require coordinating multiple containers or external parties?"

- **Yes** → Context level
- **No** → Push to Container

## Diagrams

| Type | Use For |
|------|---------|
| **Primary: System Context** | Bird's-eye view of system boundary and actors |
| **Secondary: Container Overview** | High-level container relationships |
| **Avoid** | Sequence diagrams with methods, class diagrams, flowcharts with logic |
```

**Step 2: Verify file exists**

```bash
ls -la skills/c3-context-design/defaults.md
cat skills/c3-context-design/defaults.md | head -20
```

**Step 3: Commit**

```bash
git add skills/c3-context-design/defaults.md
git commit -m "feat(c3-context-design): add defaults.md for layer configuration

Extracts include/exclude/litmus/diagrams from SKILL.md into
dedicated defaults file. Skills will read this + settings.yaml
and merge based on useDefaults flag."
```

---

## Task 2: Create Container Layer Defaults

**Files:**
- Create: `skills/c3-container-design/defaults.md`

**Step 1: Create the defaults file**

Create `skills/c3-container-design/defaults.md`:

```markdown
# Container Layer Defaults

## Include

| Element | Example |
|---------|---------|
| Technology stack | Node.js 20, Express 4.18 |
| Container responsibilities | "Handles API requests" |
| Component relationships | Flowchart of connections |
| Data flow | Sequence diagram |
| Component inventory | Links to component docs |
| API surface | Endpoints exposed |
| Data ownership | "Owns User accounts, Tasks" |
| Inter-container communication | "REST to Backend, SQL to DB" |

## Exclude

| Element | Push To |
|---------|---------|
| System boundary | Context |
| Cross-cutting concerns | Context |
| Implementation code | Component |
| Library specifics | Component |
| Configuration values | Component |

## Litmus Test

> "Is this about WHAT this container does and WITH WHAT, not HOW internally?"

- **Yes** → Container level
- **No (system-wide)** → Push to Context
- **No (implementation)** → Push to Component

## Diagrams

| Type | Use For |
|------|---------|
| **Required: Component Relationships** | Flowchart showing how components interact |
| **Required: Data Flow** | Sequence diagram showing request paths |
| **Avoid** | System context, actor diagrams, detailed class diagrams |
```

**Step 2: Verify file exists**

```bash
ls -la skills/c3-container-design/defaults.md
cat skills/c3-container-design/defaults.md | head -20
```

**Step 3: Commit**

```bash
git add skills/c3-container-design/defaults.md
git commit -m "feat(c3-container-design): add defaults.md for layer configuration"
```

---

## Task 3: Create Component Layer Defaults

**Files:**
- Create: `skills/c3-component-design/defaults.md`

**Step 1: Create the defaults file**

Create `skills/c3-component-design/defaults.md`:

```markdown
# Component Layer Defaults

## Include

| Element | Example |
|---------|---------|
| Stack details | `pg: 8.11.x` - why chosen |
| Environment config | `DB_POOL_MAX=50` (dev vs prod) |
| Implementation patterns | Connection pooling algorithm |
| Interfaces/Types | Method signatures, DTOs |
| Error handling | Retry strategies, error catalog |
| Usage examples | TypeScript snippets |

## Exclude

| Element | Push To |
|---------|---------|
| Container purpose | Container |
| API endpoint list | Container |
| Technology choice rationale | Container |
| System protocols | Context |

## Litmus Test

> "Could a developer implement this from the documentation?"

- **Yes** → Correct level
- **No, needs more detail** → Add specifics
- **No, it's about structure** → Push to Container

## Diagrams

| Type | Use For |
|------|---------|
| Flowchart | Decision logic |
| Sequence | Method calls |
| State chart | Lifecycle/state |
| ERD | Data structures |
| Class diagram | Type relationships |
| **Avoid** | System context, container overview, deployment diagrams |
```

**Step 2: Verify file exists**

```bash
ls -la skills/c3-component-design/defaults.md
cat skills/c3-component-design/defaults.md | head -20
```

**Step 3: Commit**

```bash
git add skills/c3-component-design/defaults.md
git commit -m "feat(c3-component-design): add defaults.md for layer configuration"
```

---

## Task 4: Update c3-context-design SKILL.md

**Files:**
- Modify: `skills/c3-context-design/SKILL.md`

**Step 1: Add configuration loading section after Overview**

After the "## Overview" section (around line 14), add this new section:

```markdown
## Configuration Loading

**At skill start:**

1. Read `defaults.md` from this skill directory
2. Read `.c3/settings.yaml` (if exists)
3. Check `context` section in settings:
   - If `useDefaults: true` (or missing) → merge defaults + user customizations
   - If `useDefaults: false` → use only user-provided config
4. Display merged configuration

**Merge rules:**
- `include`: defaults + user additions (union)
- `exclude`: defaults + user additions (union)
- `litmus`: user overrides default (replacement)
- `diagrams`: user overrides default (replacement)

**Display at start:**
```
Layer configuration (Context):
- Include: [merged list]
- Exclude: [merged list]
- Litmus: [active litmus test]
- Diagrams: [active diagram types]
```
```

**Step 2: Replace the "What Belongs at Context Level" section**

Replace the entire "## What Belongs at Context Level" section (lines ~36-63) with:

```markdown
## What Belongs at Context Level

See `defaults.md` for canonical include/exclude lists.

Check `.c3/settings.yaml` for project-specific overrides under the `context` section.

Apply the active litmus test when deciding content placement.
```

**Step 3: Replace the "Diagrams" section**

Replace the "## Diagrams" section (lines ~65-100) with:

```markdown
## Diagrams

See `defaults.md` for default diagram recommendations.

Check `.c3/settings.yaml` for project-specific diagram preferences under `context.diagrams`.

Use the project's `diagrams` setting (root level) for tool preference (mermaid, PlantUML, etc.).
```

**Step 4: Verify changes**

```bash
grep -n "Configuration Loading" skills/c3-context-design/SKILL.md
grep -n "defaults.md" skills/c3-context-design/SKILL.md
```

**Step 5: Commit**

```bash
git add skills/c3-context-design/SKILL.md
git commit -m "refactor(c3-context-design): reference defaults.md for layer config

- Add Configuration Loading section
- Replace hardcoded include/exclude with defaults.md reference
- Replace hardcoded diagrams with defaults.md reference
- Skills now merge defaults + settings.yaml based on useDefaults flag"
```

---

## Task 5: Update c3-container-design SKILL.md

**Files:**
- Modify: `skills/c3-container-design/SKILL.md`

**Step 1: Add configuration loading section after Overview**

After "## Overview" section (around line 14), add:

```markdown
## Configuration Loading

**At skill start:**

1. Read `defaults.md` from this skill directory
2. Read `.c3/settings.yaml` (if exists)
3. Check `container` section in settings:
   - If `useDefaults: true` (or missing) → merge defaults + user customizations
   - If `useDefaults: false` → use only user-provided config
4. Display merged configuration

**Merge rules:**
- `include`: defaults + user additions (union)
- `exclude`: defaults + user additions (union)
- `litmus`: user overrides default (replacement)
- `diagrams`: user overrides default (replacement)

**Display at start:**
```
Layer configuration (Container):
- Include: [merged list]
- Exclude: [merged list]
- Litmus: [active litmus test]
- Diagrams: [active diagram types]
```
```

**Step 2: Replace the "What Belongs at Container Level" section**

Replace the entire "## What Belongs at Container Level" section (lines ~51-84) with:

```markdown
## What Belongs at Container Level

See `defaults.md` for canonical include/exclude lists.

Check `.c3/settings.yaml` for project-specific overrides under the `container` section.

Apply the active litmus test when deciding content placement.
```

**Step 3: Replace the "Diagrams" section**

Replace the "## Diagrams" section (lines ~86-110) with:

```markdown
## Diagrams

See `defaults.md` for default diagram recommendations.

Check `.c3/settings.yaml` for project-specific diagram preferences under `container.diagrams`.

Use the project's `diagrams` setting (root level) for tool preference (mermaid, PlantUML, etc.).
```

**Step 4: Verify changes**

```bash
grep -n "Configuration Loading" skills/c3-container-design/SKILL.md
grep -n "defaults.md" skills/c3-container-design/SKILL.md
```

**Step 5: Commit**

```bash
git add skills/c3-container-design/SKILL.md
git commit -m "refactor(c3-container-design): reference defaults.md for layer config"
```

---

## Task 6: Update c3-component-design SKILL.md

**Files:**
- Modify: `skills/c3-component-design/SKILL.md`

**Step 1: Add configuration loading section after Overview**

After "## Overview" section (around line 14), add:

```markdown
## Configuration Loading

**At skill start:**

1. Read `defaults.md` from this skill directory
2. Read `.c3/settings.yaml` (if exists)
3. Check `component` section in settings:
   - If `useDefaults: true` (or missing) → merge defaults + user customizations
   - If `useDefaults: false` → use only user-provided config
4. Display merged configuration

**Merge rules:**
- `include`: defaults + user additions (union)
- `exclude`: defaults + user additions (union)
- `litmus`: user overrides default (replacement)
- `diagrams`: user overrides default (replacement)

**Display at start:**
```
Layer configuration (Component):
- Include: [merged list]
- Exclude: [merged list]
- Litmus: [active litmus test]
- Diagrams: [active diagram types]
```
```

**Step 2: Replace the "What Belongs at Component Level" section**

Replace the entire "## What Belongs at Component Level" section (lines ~61-95) with:

```markdown
## What Belongs at Component Level

See `defaults.md` for canonical include/exclude lists.

Check `.c3/settings.yaml` for project-specific overrides under the `component` section.

Apply the active litmus test when deciding content placement.
```

**Step 3: Replace the "Diagrams" section**

Replace the "## Diagrams" section (lines ~114-127) with:

```markdown
## Diagrams

See `defaults.md` for default diagram recommendations.

Check `.c3/settings.yaml` for project-specific diagram preferences under `component.diagrams`.

Use the project's `diagrams` setting (root level) for tool preference (mermaid, PlantUML, etc.).
```

**Step 4: Verify changes**

```bash
grep -n "Configuration Loading" skills/c3-component-design/SKILL.md
grep -n "defaults.md" skills/c3-component-design/SKILL.md
```

**Step 5: Commit**

```bash
git add skills/c3-component-design/SKILL.md
git commit -m "refactor(c3-component-design): reference defaults.md for layer config"
```

---

## Task 7: Update c3-config SKILL.md for Enhanced Layer Sections

**Files:**
- Modify: `skills/c3-config/SKILL.md`

**Step 1: Update the Settings File Structure section**

Replace the settings.yaml example in "## Settings File Structure" (lines ~33-62) with the enhanced version:

```yaml
diagrams: |
  mermaid
  sequence: API interactions, request flows
  flowchart: decision logic, error handling

context:
  useDefaults: true
  guidance: |
    system boundaries, actors, external integrations
    avoid implementation details
  include: |
    # Optional: add items to defaults
  exclude: |
    # Optional: add items to defaults
  litmus: |
    # Optional: override default litmus test
  diagrams: |
    # Optional: override default diagrams

container:
  useDefaults: true
  guidance: |
    service responsibilities, API contracts
    backend: endpoints, data flows
    frontend: component hierarchy, state

component:
  useDefaults: true
  guidance: |
    technical specifics, configs, algorithms

guard: |
  discovered incrementally via c3-config

handoff: |
  after ADR accepted:
  1. create implementation tasks
  2. notify team

audit: |
  handoff: tasks
```

**Step 2: Update the Sections table**

Replace the "**Sections:**" table with:

```markdown
**Sections:**
| Section | Purpose |
|---------|---------|
| `diagrams` | Diagram tool and usage patterns (global) |
| `context` | Context-layer configuration |
| `context.useDefaults` | Load defaults from skill's defaults.md (default: true) |
| `context.guidance` | Prose guidance for context documentation |
| `context.include` | Additional items to include at context level |
| `context.exclude` | Additional items to exclude from context level |
| `context.litmus` | Override default litmus test |
| `context.diagrams` | Override default diagram recommendations |
| `container` | Container-layer configuration (same keys as context) |
| `component` | Component-layer configuration (same keys as context) |
| `guard` | Team guardrails and constraints |
| `handoff` | Post-ADR completion steps |
| `audit` | Audit findings handoff preference |
```

**Step 3: Add new Section-Specific Questions for layer config**

In "#### Section-Specific Questions" section, add after Diagrams questions:

```markdown
**Layer Configuration (Context/Container/Component):**
- "Do you want to use the default include/exclude rules, or customize them?"
- "Are there specific items that should ALWAYS be at this layer for your project?"
- "Are there items that should NEVER be at this layer?"
- "Do you want a custom litmus test for deciding content placement?"
```

**Step 4: Verify changes**

```bash
grep -n "useDefaults" skills/c3-config/SKILL.md
grep -n "Layer Configuration" skills/c3-config/SKILL.md
```

**Step 5: Commit**

```bash
git add skills/c3-config/SKILL.md
git commit -m "feat(c3-config): support enhanced layer configuration

- Add useDefaults flag per layer
- Add include/exclude/litmus/diagrams per layer
- Update settings.yaml example
- Add section-specific questions for layer config"
```

---

## Task 8: Update README.md to Document New Features

**Files:**
- Modify: `README.md`

**Step 1: Add Layer Configuration section**

In the README, find the "## Getting Started" or similar section and add:

```markdown
## Layer Configuration

Each C3 layer (Context, Container, Component) has configurable rules for what content belongs at that level.

### Defaults

Each layer skill includes a `defaults.md` with sensible defaults:
- `skills/c3-context-design/defaults.md`
- `skills/c3-container-design/defaults.md`
- `skills/c3-component-design/defaults.md`

### Customization

Customize layer rules in `.c3/settings.yaml`:

```yaml
context:
  useDefaults: true    # Load defaults, then merge customizations
  guidance: |
    your team's context-level guidance
  include: |
    compliance requirements
    audit trails
  exclude: |
    internal service details
  litmus: |
    "Does this affect external stakeholders?"

container:
  useDefaults: true
  guidance: |
    your team's container-level guidance

component:
  useDefaults: false   # Only use what's specified here
  guidance: |
    your team's component-level guidance
  include: |
    your custom include list
```

### Configuration Keys

| Key | Type | Description |
|-----|------|-------------|
| `useDefaults` | boolean | `true`: merge with defaults. `false`: only use settings.yaml |
| `guidance` | prose | General documentation guidance |
| `include` | list | Items that belong at this layer |
| `exclude` | list | Items to push to other layers |
| `litmus` | prose | Decision question for content placement |
| `diagrams` | list | Diagram types for this layer |
```

**Step 2: Verify changes**

```bash
grep -n "Layer Configuration" README.md
grep -n "useDefaults" README.md
```

**Step 3: Commit**

```bash
git add README.md
git commit -m "docs: add layer configuration documentation to README"
```

---

## Task 9: Update VERSION File

**Files:**
- Modify: `VERSION`

**Step 1: Update VERSION to new slug**

```bash
echo "20251125-layer-settings" > VERSION
```

**Step 2: Verify**

```bash
cat VERSION
```

Expected: `20251125-layer-settings`

**Step 3: Commit**

```bash
git add VERSION
git commit -m "chore: bump VERSION to 20251125-layer-settings"
```

---

## Task 10: Create Migration File

**Files:**
- Create: `migrations/20251125-layer-settings.md`

**Step 1: Create migration file**

Create `migrations/20251125-layer-settings.md`:

```markdown
# Migration: 20251125-layer-settings

> From: `20251124-toc-fix-links`
> To: `20251125-layer-settings`

## Changes

- Layer sections in settings.yaml now support structured configuration
- New keys per layer: `useDefaults`, `include`, `exclude`, `litmus`, `diagrams`
- Existing prose-only layer config (e.g., `context: |`) still works (backward compatible)
- Layer design skills now read defaults from co-located `defaults.md` files

## Transforms

### Upgrade settings.yaml (Optional)

**If exists:** `.c3/settings.yaml`

Users MAY upgrade from prose-only format to structured format:

**Before (still works):**
```yaml
context: |
  system boundaries, actors, external integrations
  avoid implementation details
```

**After (new capabilities):**
```yaml
context:
  useDefaults: true
  guidance: |
    system boundaries, actors, external integrations
    avoid implementation details
  include: |
    # Optional: add items to defaults
  exclude: |
    # Optional: add items to defaults
  litmus: |
    # Optional: override default litmus test
```

**No automatic transforms needed** - the change is backward compatible.

Existing prose-only format continues to work; users can opt-in to structured format when they want customization.

## Verification

```bash
# VERSION updated
cat .c3/README.md | grep -q 'c3-version: 20251125-layer-settings' && echo "VERSION: OK"

# settings.yaml still valid (if exists)
if [ -f ".c3/settings.yaml" ]; then
  grep -q '^context:' .c3/settings.yaml && echo "context section: OK"
  grep -q '^container:' .c3/settings.yaml && echo "container section: OK"
  grep -q '^component:' .c3/settings.yaml && echo "component section: OK"
fi
```

## Backward Compatibility

| Old Format | New Format | Works? |
|------------|------------|--------|
| `context: \|` (prose) | `context:` (structured) | ✅ Yes, prose still valid |
| No settings.yaml | No settings.yaml | ✅ Yes, skills use defaults.md |

Users only need to migrate if they want the new customization features.
```

**Step 2: Verify file exists**

```bash
ls -la migrations/20251125-layer-settings.md
cat migrations/20251125-layer-settings.md | head -20
```

**Step 3: Commit**

```bash
git add migrations/20251125-layer-settings.md
git commit -m "docs: add migration file for 20251125-layer-settings

Backward compatible change - existing settings.yaml continues to work.
Users can opt-in to structured layer config for customization."
```

---

## Task 11: Update c3-migrate SKILL.md

**Files:**
- Modify: `skills/c3-migrate/SKILL.md`

**Step 1: Add migration section after V2 → V3 section**

After the "## V2 → V3 Migration" section (around line 195), add:

```markdown
---

## 20251125-layer-settings Migration

### Changes

- Layer sections support structured configuration with `useDefaults`, `include`, `exclude`, `litmus`, `diagrams`
- Existing prose-only format still works (backward compatible)
- Skills read defaults from `defaults.md` files

### Transforms

**No automatic transforms required.**

This migration is backward compatible:
- Existing `context: |` prose format continues to work
- Skills fall back to defaults.md when settings not customized

### Optional Upgrade

Users who want layer customization can manually convert:

**From:**
```yaml
context: |
  prose guidance here
```

**To:**
```yaml
context:
  useDefaults: true
  guidance: |
    prose guidance here
  include: |
    custom items
```

### Verification

```bash
# VERSION shows 20251125-layer-settings or later
grep 'c3-version:' .c3/README.md

# settings.yaml still valid (basic check)
if [ -f ".c3/settings.yaml" ]; then
  grep -q '^context:' .c3/settings.yaml || grep -q '^context$' .c3/settings.yaml
fi
```
```

**Step 2: Verify changes**

```bash
grep -n "20251125-layer-settings" skills/c3-migrate/SKILL.md
```

**Step 3: Commit**

```bash
git add skills/c3-migrate/SKILL.md
git commit -m "docs(c3-migrate): add 20251125-layer-settings migration guide"
```

---

## Task 12: Final Verification

**Step 1: Verify all defaults.md files exist**

```bash
ls -la skills/c3-context-design/defaults.md
ls -la skills/c3-container-design/defaults.md
ls -la skills/c3-component-design/defaults.md
```

**Step 2: Verify all SKILL.md files reference defaults.md**

```bash
grep -l "defaults.md" skills/*/SKILL.md
```

Expected output:
```
skills/c3-component-design/SKILL.md
skills/c3-config/SKILL.md
skills/c3-container-design/SKILL.md
skills/c3-context-design/SKILL.md
```

**Step 3: Verify configuration loading sections exist**

```bash
grep -l "Configuration Loading" skills/c3-*-design/SKILL.md
```

Expected output:
```
skills/c3-component-design/SKILL.md
skills/c3-container-design/SKILL.md
skills/c3-context-design/SKILL.md
```

**Step 4: Verify VERSION and migration files**

```bash
cat VERSION
ls -la migrations/20251125-layer-settings.md
grep -n "20251125-layer-settings" skills/c3-migrate/SKILL.md
```

Expected:
- VERSION contains `20251125-layer-settings`
- Migration file exists
- c3-migrate skill references the new migration

**Step 5: Git log to verify commits**

```bash
git log --oneline -15
```

**Step 6: Report completion**

All files created and modified:
- 3 defaults.md files (context, container, component)
- 4 SKILL.md files updated (3 layer skills + c3-config)
- README.md updated with documentation
- VERSION bumped to 20251125-layer-settings
- Migration file created
- c3-migrate skill updated

Feature branch ready for review/merge.
