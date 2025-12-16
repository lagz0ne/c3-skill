# Migration: 20251216-enforcement-harnesses

## Changes

### NO CODE Enforcement Strengthened (c3-component-design)

Added explicit prohibitions and guidance for:
- JSON/YAML schemas → use tables with Field|Type|Purpose
- Example payloads → use tables with Example Value column
- Pseudocode schemas (`{ field: type }`) → prohibited
- Wire format examples → prohibited

New red flags added:
- "Example payload" in code blocks
- `{ field: type }` notation
- "Wire format" justification
- Nested JSON/YAML

Clarifications added:
- Why Mermaid allowed but JSON not (parseable = code)
- references/ is NOT an escape hatch for schemas

### V3 Structure Enforcement Added (c3-adopt + layer skills)

Added mandatory structure validation:
- Context IS `.c3/README.md` (not `context/c3-0.md`)
- Containers are folders: `.c3/c3-{N}-{slug}/README.md`
- Components inside container folders (not `components/` folder)

Prohibited v2 patterns:
- `context/` folder
- `containers/` folder
- `components/` folder
- Any `c3-0.md` file

File location reminders added to all layer skills.

## Transforms

**No automatic transforms required.**

This migration adds skill-internal enforcement. User `.c3/` directories are not affected if already using v3 structure.

## Verification

```bash
# Verify skills have enforcement sections
grep -l "NO CODE ENFORCEMENT" skills/c3-component-design/SKILL.md
grep -l "V3 STRUCTURE ENFORCEMENT" skills/c3-adopt/SKILL.md

# Verify layer skills have file location reminders
grep "File Location" skills/c3-*-design/SKILL.md
```
