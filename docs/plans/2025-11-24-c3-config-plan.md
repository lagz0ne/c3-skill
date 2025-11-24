# C3 Config Implementation Plan

Based on: `2025-11-24-c3-config-design.md`

## Tasks

### Task 1: Create c3-config skill

**File:** `skills/c3-config/SKILL.md`

**Content outline:**
```markdown
---
name: c3-config
description: Configure C3 project settings - creates settings.yaml with sensible defaults, refines via Socratic questioning
---

# C3 Config

## Overview
Create and refine `.c3/settings.yaml` with project preferences.

## Flow

1. Check if `.c3/settings.yaml` exists
2. If missing → create with defaults (see template below)
3. Show current settings to user
4. Socratic questions to refine sections
5. Write updates

## Default Settings Template

```yaml
diagrams: |
  mermaid
  sequence: API interactions, request flows
  flowchart: decision logic, error handling

context: |
  system boundaries, actors, external integrations
  avoid implementation details

container: |
  service responsibilities, API contracts
  backend: endpoints, data flows
  frontend: component hierarchy, state

component: |
  technical specifics, configs, algorithms

guard: |
  discovered incrementally via c3-config

handoff: |
  after ADR accepted:
  1. create implementation tasks
  2. notify team
```

## Socratic Questions

### Diagrams
- "What diagram tool does your team use? (mermaid/plantuml/ascii)"
- "Any specific diagram conventions?"

### Guardrails
- "What patterns should be avoided in your architecture?"
- "Any documentation anti-patterns to watch for?"

### Handoff
- "What happens after an ADR is accepted?"
- "What tools/channels should be notified?"
```

**Verification:** File exists, skill is loadable

---

### Task 2: Update c3-adopt to call c3-config

**File:** `skills/c3-adopt/SKILL.md`

**Change:** Add settings check after Phase 1 scaffolding

**Location:** After Step 1.4 (Scaffold .c3/ Directory), add:

```markdown
### Step 1.5: Initialize Settings

Check for settings file:
```bash
ls .c3/settings.yaml 2>/dev/null || echo "Settings not found"
```

If missing:
> "I'll use the c3-config skill to set up project preferences."

Invoke c3-config skill.
```

**Also add to Verification Checklist:**
```markdown
- [ ] `.c3/settings.yaml` - Project settings
```

**Verification:** c3-adopt includes settings step

---

### Task 3: Update c3-init to mention settings

**File:** `commands/c3-init.md`

**Change:** Add settings to the list

```markdown
Use the `c3-adopt` skill for fresh initialization, which will:
1. Create `.c3/` directory structure (v3: containers as lowercase folders)
2. Initialize project settings (`.c3/settings.yaml`)
3. Track version via `c3-version` frontmatter in README.md
4. Guide through Context, Container, and Component discovery
```

**Verification:** c3-init mentions settings

---

### Task 4: Update c3-design to read settings

**File:** `skills/c3-design/SKILL.md`

**Change:** Add settings reading to Prerequisites section

**Location:** After "Prerequisites" heading, add:

```markdown
**Settings:** If `.c3/settings.yaml` exists, read it at session start.
Apply guidance from:
- `diagrams:` when creating visuals
- `context:/container:/component:` when writing docs
- `guard:` as constraints
- `handoff:` after ADR completion
```

**Verification:** c3-design references settings

---

### Task 5: Update v3-structure.md reference

**File:** `references/v3-structure.md`

**Change:** Add settings.yaml to Directory Layout and document format

**Location:** In "Directory Layout" section, add:

```
.c3/
├── settings.yaml          # Project settings
├── README.md              # Context (id: c3-0, c3-version: 3)
...
```

**Add new section:**

```markdown
## Settings File

`.c3/settings.yaml` stores project preferences:

```yaml
diagrams: |
  tool and usage patterns

context: |
  context-layer guidance

container: |
  container-layer guidance

component: |
  component-layer guidance

guard: |
  team guardrails

handoff: |
  post-ADR steps
```

Created via `c3-config` skill, refined incrementally.
```

**Verification:** v3-structure includes settings

---

## Execution Order

1. Task 1 (c3-config skill) - no dependencies
2. Task 5 (v3-structure) - no dependencies
3. Task 2 (c3-adopt) - depends on Task 1
4. Task 3 (c3-init) - depends on Task 2
5. Task 4 (c3-design) - depends on Task 1

**Parallel:** Tasks 1, 5 can run in parallel
**Sequential:** Tasks 2, 3, 4 after Task 1

## Verification

After all tasks:
1. Run c3-config standalone → creates settings.yaml
2. Run c3-adopt on fresh project → includes settings step
3. Run c3-design → reads settings at start
