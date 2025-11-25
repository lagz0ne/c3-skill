---
name: c3-config
description: Create and refine .c3/settings.yaml for project-specific preferences - diagram tools, layer guidance, guardrails, and handoff steps
---

# C3 Config - Project Settings

## Overview

Create and refine `.c3/settings.yaml` for project-specific preferences through Socratic questioning.

**Announce at start:** "I'm using the c3-config skill to configure project settings."

## Quick Reference

| Phase | Key Activities | Output |
|-------|---------------|--------|
| **1. Check Existing** | Look for settings.yaml | Determine create vs update |
| **2. Initialize/Load** | Create defaults or load existing | Working settings |
| **3. Socratic Refinement** | Questions to refine sections | Updated preferences |
| **4. Write Settings** | Save to .c3/settings.yaml | Persisted config |

## Prerequisites

**Required:** `.c3/` directory must exist.

If `.c3/` doesn't exist:
- Stop and suggest: "Use the `c3-adopt` skill to initialize C3 documentation first"

## Settings File Structure

`.c3/settings.yaml` contains project preferences:

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

## The Process

### Phase 1: Check Existing Settings

```bash
ls .c3/settings.yaml 2>/dev/null && echo "EXISTS" || echo "MISSING"
```

- If **EXISTS**: Load and show current settings
- If **MISSING**: Proceed to create with defaults

### Phase 2: Initialize or Load

#### If Missing - Create with Defaults

Create `.c3/settings.yaml` with sensible defaults:

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

audit: |
  handoff: tasks
```

#### If Exists - Load and Display

Show current settings to user:
```
Current settings in .c3/settings.yaml:

diagrams: [current value]
context: [current value]
...

Which section would you like to refine?
```

### Phase 3: Socratic Refinement

**Goal:** Refine sections through targeted questions, not exhaustive interrogation.

#### Refinement Approach

1. **Focus on gaps** - Ask about empty/minimal sections first
2. **One section at a time** - Don't overwhelm with all questions at once
3. **Accept defaults** - User can skip any section

#### Section-Specific Questions

**Diagrams:**
- "What diagramming tool does your team use? (mermaid, PlantUML, draw.io, etc.)"
- "What types of diagrams are most useful for your project?"

**Layer Configuration (Context/Container/Component):**
- "Do you want to use the default include/exclude rules, or customize them?"
- "Are there specific items that should ALWAYS be at this layer for your project?"
- "Are there items that should NEVER be at this layer?"
- "Do you want a custom litmus test for deciding content placement?"

**Context/Container/Component Guidance:**
- "What should documentation at the [layer] level emphasize?"
- "Any specific patterns or anti-patterns to call out?"

**Guardrails:**
- "Are there architectural decisions that should never be revisited?"
- "Any technologies or patterns that are off-limits?"
- "Performance or security constraints to always consider?"

**Handoff:**
- "What happens after an ADR is accepted?"
- "How should implementation tasks be tracked? (GitHub issues, Jira, Linear, etc.)"

**Audit:**
Use AskUserQuestion:
```
Question: "How should audit findings be handled?"
Options:
  - "manual" (Review findings manually, decide case-by-case)
  - "tasks" (Automatically create tasks/tickets for findings)
  - "agents" (Dispatch agents to investigate/fix findings)
```

#### Refinement Flow

```
┌─────────────────────────────────────────────────────┐
│ Show current section value                          │
│        ↓                                            │
│ Ask targeted question                               │
│        ↓                                            │
│ User provides input OR skips                        │
│        ↓                                            │
│ Update section value                                │
│        ↓                                            │
│ Move to next section OR finish                      │
└─────────────────────────────────────────────────────┘
```

### Phase 4: Write Settings

Save updated settings to `.c3/settings.yaml`.

**Verification:**
```bash
# Confirm file exists and has expected sections
ls .c3/settings.yaml
grep -q '^diagrams:' .c3/settings.yaml && echo "diagrams: OK"
grep -q '^context:' .c3/settings.yaml && echo "context: OK"
grep -q '^container:' .c3/settings.yaml && echo "container: OK"
grep -q '^component:' .c3/settings.yaml && echo "component: OK"
grep -q '^guard:' .c3/settings.yaml && echo "guard: OK"
grep -q '^handoff:' .c3/settings.yaml && echo "handoff: OK"
grep -q '^audit:' .c3/settings.yaml && echo "audit: OK"
```

**Summary:**
```
Settings saved to .c3/settings.yaml

Configured:
- diagrams: mermaid with sequence/flowchart patterns
- context: [summary]
- container: [summary]
- component: [summary]
- guard: [summary or "none yet"]
- handoff: [summary]
- audit: [handoff preference]

These settings will be used by c3-design when creating documentation.
You can edit .c3/settings.yaml directly anytime.
```

## Invocation Contexts

| Context | Behavior |
|---------|----------|
| **Standalone** | Full Socratic refinement of all sections |
| **Via c3-adopt** | Create with defaults, minimal questions |
| **Via c3-init** | Create with defaults, offer refinement |

When called from `c3-adopt` or `c3-init`:
- Create defaults immediately
- Ask: "Would you like to customize project settings now, or use defaults?"
- If customize → full refinement
- If defaults → done

## Key Principles

| Principle | Application |
|-----------|-------------|
| **Sensible defaults** | Settings work out-of-box, customization is optional |
| **Incremental discovery** | Guardrails grow over time, not all at once |
| **Flexible format** | YAML with prose values, not rigid schema |
| **User can edit directly** | settings.yaml is human-readable/editable |
| **Non-blocking** | Missing settings doesn't break other skills |

## Related Skills

- [c3-adopt](../c3-adopt/SKILL.md) - Calls c3-config during initialization
- [c3-design](../c3-design/SKILL.md) - Reads settings at start
