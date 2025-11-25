# Layer Settings Externalization Design

**Date:** 2025-01-25
**Status:** Approved for implementation

## Problem

Currently, layer instructions (what to include/exclude, litmus tests, diagram preferences) are hardcoded in each layer design skill (`c3-context-design`, `c3-container-design`, `c3-component-design`). This means:

1. Users cannot customize what belongs at each layer for their team's needs
2. All projects use the same rigid rules regardless of domain
3. To change layer behavior, users would need to fork the plugin

## Solution

Externalize layer instructions to:
1. **Per-skill defaults files** (`defaults.md`) - canonical defaults owned by each skill
2. **settings.yaml** - user customizations that layer on top of (or replace) defaults

## Design

### File Structure

```
skills/
├── c3-context-design/
│   ├── SKILL.md           # Skill logic (slimmed down)
│   └── defaults.md        # Context layer defaults
├── c3-container-design/
│   ├── SKILL.md
│   └── defaults.md        # Container layer defaults
├── c3-component-design/
│   ├── SKILL.md
│   └── defaults.md        # Component layer defaults
```

### defaults.md Structure

Each `defaults.md` contains the canonical defaults for that layer:

```markdown
# [Layer] Layer Defaults

## Include
- item 1
- item 2
...

## Exclude
- item 1 → [where it should go]
- item 2 → [where it should go]
...

## Litmus Test
"[Question to ask when deciding if content belongs at this layer]"
- Yes → [action]
- No → [action]

## Diagrams
- Primary: [type]
- Secondary: [type]
- Avoid: [types]
```

### settings.yaml Structure

Enhanced layer sections in `.c3/settings.yaml`:

```yaml
# Existing general settings (unchanged)
diagrams: |
  mermaid
  sequence: API interactions, request flows
  flowchart: decision logic, error handling

# Enhanced layer sections
context:
  useDefaults: true   # Load from skills/c3-context-design/defaults.md
  guidance: |
    system boundaries, actors, external integrations
    avoid implementation details
  include: |          # Optional: add to defaults
    compliance requirements
    audit trails
  exclude: |          # Optional: add to defaults
    internal service details
  litmus: |           # Optional: override default litmus
    "Does this affect external stakeholders or system contracts?"

container:
  useDefaults: true
  guidance: |
    service responsibilities, API contracts
    backend: endpoints, data flows
    frontend: component hierarchy, state
  # No include/exclude/litmus = just use defaults + guidance

component:
  useDefaults: true
  guidance: |
    technical specifics, configs, algorithms
  diagrams: |         # Optional: layer-specific diagram prefs
    sequence: method calls between components
    flowchart: decision logic

# Rest unchanged
guard: |
  discovered incrementally via c3-config

handoff: |
  after ADR accepted:
  1. create implementation tasks
  2. notify team
```

### Configuration Keys

| Key | Type | Description |
|-----|------|-------------|
| `useDefaults` | boolean | `true` (default): load skill's defaults.md then merge. `false`: only use settings.yaml |
| `guidance` | prose | General prose guidance for the layer (existing behavior) |
| `include` | list | Items that belong at this layer |
| `exclude` | list | Items to push to other layers |
| `litmus` | prose | Decision question for placement |
| `diagrams` | list | Diagram types appropriate for this layer |

### Merge Logic

When `useDefaults: true` (default):
- `include`: defaults + user additions (union)
- `exclude`: defaults + user additions (union)
- `litmus`: user overrides default (replacement), or default if not specified
- `diagrams`: user overrides default (replacement), or default if not specified
- `guidance`: user prose is additional context

When `useDefaults: false`:
- Defaults are NOT loaded
- Only user-provided config is used
- Missing keys result in empty/no guidance for that aspect

### Skill Behavior

Each layer skill follows this pattern:

```
1. Read own defaults.md (for reference)
2. Read .c3/settings.yaml
3. Check layer config:
   - If useDefaults: true (or missing) → merge defaults + user customizations
   - If useDefaults: false → use only user-provided config
4. Display merged configuration at start of exploration
5. Apply include/exclude/litmus during documentation decisions
```

**Example output at skill start:**

```
I'm using c3-context-design skill to explore Context-level impact.

Layer configuration:
- Include: system boundary, actors, protocols, compliance requirements (custom)
- Exclude: technology choices, configuration values, internal service details (custom)
- Litmus: "Does this affect external stakeholders or system contracts?"
- Diagrams: System Context, Container Overview

Proceeding with exploration...
```

## Files to Change

### Create

1. `skills/c3-context-design/defaults.md` - Extract from current SKILL.md
2. `skills/c3-container-design/defaults.md` - Extract from current SKILL.md
3. `skills/c3-component-design/defaults.md` - Extract from current SKILL.md

### Modify

1. `skills/c3-context-design/SKILL.md` - Slim down, add config loading logic
2. `skills/c3-container-design/SKILL.md` - Slim down, add config loading logic
3. `skills/c3-component-design/SKILL.md` - Slim down, add config loading logic
4. `skills/c3-config/SKILL.md` - Handle enhanced layer sections, show useDefaults option

## Migration Impact

**Migration: `20251125-layer-settings`**

This change is **backward compatible** but requires version bump:

**What changes:**
- New defaults.md files are plugin files
- settings.yaml schema enhanced (new keys per layer)
- `useDefaults` defaults to `true`, so existing behavior preserved

**Backward compatibility:**
- Existing prose-only format (`context: |`) continues to work
- Users can opt-in to structured format for customization
- No automatic transforms needed

**Migration files:**
- `VERSION` → `20251125-layer-settings`
- `migrations/20251125-layer-settings.md` - documents the change
- `skills/c3-migrate/SKILL.md` - updated with migration guide

## Benefits

1. **Customizability** - Teams can adapt layer rules to their domain
2. **Sensible defaults** - Works out-of-box, customization optional
3. **Integrity** - Defaults live next to skills that use them (no drift)
4. **Transparency** - Config displayed at skill start
5. **Backward compatible** - Existing settings.yaml files continue working
