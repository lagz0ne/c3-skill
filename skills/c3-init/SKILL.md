---
name: c3-init
description: Initialize C3 documentation structure from scratch - creates .c3/ directory and guides initial setup
---

# C3 Init

## Overview

Create `.c3/` directory structure for a new project. For existing projects with code, use `c3-adopt` instead.

**Announce at start:** "I'm using the c3-init skill to initialize C3 documentation structure."

## Quick Reference

| Phase | Key Activities | Output |
|-------|---------------|--------|
| **1. Pre-flight** | Check no existing .c3/, confirm intent | Ready to init |
| **2. Create Structure** | Create directories and base files | .c3/ skeleton |
| **3. Configure** | Set up settings.yaml | Project preferences |
| **4. Handoff** | Guide next steps | User knows how to proceed |

## When to Use

| Scenario | Use This Skill | Use Instead |
|----------|---------------|-------------|
| Brand new project, no code yet | Yes | - |
| Existing codebase needs C3 docs | No | `c3-adopt` |
| .c3/ exists but needs rebuild | No | `c3-migrate` or delete + re-init |

## Phase 1: Pre-flight Check

```bash
# Check if .c3/ already exists
if [ -d ".c3" ]; then
  echo "EXISTS"
else
  echo "CLEAR"
fi
```

**If EXISTS:**
- Stop and inform user
- Suggest: "C3 structure already exists. Use `c3-design` to work with it, or `c3-migrate` if it needs updating."

**If CLEAR:**
- Confirm with user: "I'll create a fresh C3 structure. This project appears to be [new/has code]. Proceed?"

## Phase 2: Create Structure

### Directory Structure (v3)

```bash
mkdir -p .c3/adr
```

### Create TOC.md

```bash
cat > .c3/TOC.md << 'EOF'
# Table of Contents

## Context
- [System Overview](README.md) - c3-0

## Containers
<!-- Add containers as c3-N-slug/README.md -->

## ADRs
<!-- Add ADRs as adr/adr-YYYYMMDD-slug.md -->
EOF
```

### Create Context README

```bash
cat > .c3/README.md << 'EOF'
---
id: c3-0
c3-version: 3
title: System Overview
---

# System Overview

## Overview {#c3-0-overview}
<!--
System purpose in 1-2 sentences.
What problem does it solve? For whom?
-->

## System Boundary {#c3-0-boundary}
<!--
What's INSIDE vs OUTSIDE the system.
-->

### Inside (Our System)
- [List containers]

### Outside (External)
- [List external systems]

## Actors {#c3-0-actors}
<!--
Who/what interacts with the system.
-->

| Actor | Type | Interacts Via | Implemented By |
|-------|------|---------------|----------------|
| | | | |

## Containers {#c3-0-containers}
<!--
Inventory of all containers.
-->

| Container | ID | Type | Responsibility |
|-----------|-----|------|----------------|
| | | | |

## Protocols {#c3-0-protocols}
<!--
How containers communicate.
-->

| From | To | Protocol | Contract |
|------|-----|----------|----------|
| | | | |

## Cross-Cutting Concerns {#c3-0-cross-cutting}
<!--
System-wide patterns.
-->

### Authentication {#c3-0-auth}
<!-- Strategy and implementation -->

### Logging {#c3-0-logging}
<!-- Pattern and implementation -->

### Error Handling {#c3-0-errors}
<!-- Strategy and implementation -->
EOF
```

### TOC Generation Note

**TOC Generation:** The `build-toc.sh` script lives in the c3-skill plugin, not in user projects.

When you need to rebuild the TOC:
- Ask Claude to "rebuild the TOC" (c3-toc skill will handle it)
- Or run the plugin's script directly if needed

**No local script is created** - the plugin manages TOC generation.

## Phase 3: Configure (Optional)

Ask user: "Would you like to set up project preferences now?"

If yes, invoke `c3-config` skill to create settings.yaml.

If no, inform: "You can run `/c3-config` later to set up preferences."

## Phase 4: Handoff

Summarize what was created:

```
C3 structure initialized:

.c3/
├── README.md          # Context document (c3-0)
├── TOC.md             # Table of contents
└── adr/               # Architecture Decision Records

Note: TOC generation is handled by the plugin's c3-toc skill.

Next steps:
1. Edit .c3/README.md to define your system context
2. Use /c3 to design architecture changes
3. Use /c3-config to set up project preferences
```

## Checklist

- [ ] No existing .c3/ (or user confirmed overwrite)
- [ ] Directory structure created (.c3/adr/)
- [ ] TOC.md created
- [ ] Context README created with template
- [ ] User informed that TOC generation is handled by plugin
- [ ] User informed of next steps

## Related

- [c3-adopt](../c3-adopt/SKILL.md) - For existing codebases
- [c3-config](../c3-config/SKILL.md) - Configure preferences
- [c3-design](../c3-design/SKILL.md) - Main design workflow
- [v3-structure.md](../../references/v3-structure.md) - Structure reference
