---
name: c3-adopt
description: Initialize C3 architecture documentation for an existing project - uses Socratic questioning to build understanding, then delegates to layer skills for document creation
---

# C3 Adopt - Initialize Architecture Documentation

## Overview

Bootstrap C3 (Context-Container-Component) architecture documentation for an existing codebase through Socratic questioning and delegation.

**Core principle:** Ask questions to build understanding, then delegate document creation to the appropriate layer skill.

**Announce at start:** "I'm using the c3-adopt skill to initialize architecture documentation for this project."

## Quick Reference

| Phase | Key Activities | Output |
|-------|---------------|--------|
| **1. Establish** | Check prerequisites, create scaffolding | `.c3/` directory |
| **2. Context Discovery** | Socratic questions about system | Understanding for CTX |
| **3. Container Discovery** | Socratic questions per container | Understanding for C3-X |
| **4. Component Identification** | Identify key components | C3-XNN stubs |
| **5. Generate & Verify** | Delegate to sub-skills, build TOC | Complete documentation |

## Derivation Guardrails (apply everywhere)

- **Reading order:** Context → Container → Component
- **Downward-only links:** CTX → C3-X sections; C3-X → C3-XNN docs. No upward links.
- **Infra containers are leaf nodes:** No components; their features must be cited by consuming components.
- **Anchors:** Use `{#ctx-*}` for context, `{#c3-xxx-*}` for containers and components.
- **Naming:** Use `c3-naming` (components carry parent container code; IDs match filenames/paths).
- **Templates/checklists:** Use templates from c3-context-design, c3-container-design, c3-component-design and complete all checklist items.

---

## Phase 1: Establish

### Check Prerequisites

```bash
# Verify .c3 doesn't already exist
ls -la .c3 2>/dev/null && echo "WARNING: .c3 already exists"
```

If `.c3/` exists, ask:
> "I found existing `.c3/` documentation. Would you like me to:
> 1. Review and update existing documentation
> 2. Back up and create fresh documentation
> 3. Abort and preserve what's there"

### Create Scaffolding

If proceeding with new documentation:

```bash
mkdir -p .c3/{containers,components,adr,scripts}
```

Create VERSION file with current skill version:
```bash
# Read current version from plugin
SKILL_VERSION=$(cat /path/to/c3-skill/VERSION)
echo "$SKILL_VERSION" > .c3/VERSION
```

Create `build-toc.sh` script (see Appendix A).

Create `index.md`:
```markdown
---
layout: home
title: C3 Architecture Documentation
---

# C3 Architecture Documentation

- [Table of Contents](./TOC.md)
- [System Overview](./CTX-system-overview.md)
```

---

## Phase 2: Context Discovery

**Goal:** Build understanding through questions, NOT code scanning.

### The Socratic Approach

Instead of scanning files, ASK the user. They know their system better than any automated scan.

### Context Questions (ask in order, build on answers)

#### System Identity
1. > "What is the name of this system/project?"
2. > "In one sentence, what does it do for users?"

#### System Boundary
3. > "What is INSIDE your system vs EXTERNAL?"
   - Help them think: "What code do you control vs depend on?"
4. > "Who or what interacts with your system from outside?"
   - Users? Admins? Other systems? Scheduled jobs?

#### Containers (deployable units)
5. > "If you deployed this today, what separate processes would be running?"
   - Hint: "Think Docker containers, serverless functions, databases"
6. > "What data stores exist? (databases, caches, queues)"

#### Communication
7. > "How do these pieces talk to each other?"
   - REST? gRPC? Message queues? Direct function calls?
8. > "What external services does your system call?"
   - Email? Payment? Auth providers?

#### Cross-Cutting
9. > "How is authentication handled across the system?"
10. > "How do you handle logging and monitoring?"

### Build Context Model

From answers, construct:

```markdown
## Context Understanding

**System:** [Name]
**Purpose:** [One sentence]

**Actors:**
- [Actor 1]
- [Actor 2]

**Containers (identified):**
| Name | Type | Purpose |
|------|------|---------|
| [Name] | [Backend/Frontend/DB/etc] | [What it does] |

**Protocols:**
| From | To | Protocol |
|------|-----|----------|
| [A] | [B] | [REST/gRPC/etc] |

**External Dependencies:**
- [Service 1]
- [Service 2]
```

### Delegate to c3-context-design

Once you have understanding:
> "I now understand your system context. I'll use the c3-context-design skill to create CTX-system-overview."

Use `c3-context-design` to create the Context document with:
- System overview
- Architecture diagram (from identified containers)
- Container list (high-level)
- Protocols section
- Cross-cutting concerns
- Deployment section

---

## Phase 3: Container Discovery

**Goal:** For each container identified, build understanding through questions.

### Container Questions (per container)

#### Identity
1. > "For [Container Name]: What is its single main responsibility?"
2. > "What technology does it use? (language, framework)"

#### Structure
3. > "How is code organized inside? Layers? Feature folders?"
4. > "What are the main entry points?"

#### APIs
5. > "What endpoints/interfaces does it expose?"
6. > "What does it consume from other containers?"

#### Data
7. > "What data does this container own?"
8. > "What data does it read but not own?"

#### Key Components
9. > "What are the 3-5 most important components inside?"
   - "The ones a new developer MUST understand"

### Build Container Model

From answers:

```markdown
## Container: [Name]

**Responsibility:** [One sentence]
**Technology:** [Language + Framework]

**Structure:**
| Layer | Purpose |
|-------|---------|
| [Layer] | [What it does] |

**APIs Exposed:**
- [Endpoint 1]
- [Endpoint 2]

**Data Owned:**
- [Data 1]
- [Data 2]

**Key Components:**
| Component | Purpose | Priority |
|-----------|---------|----------|
| [Name] | [What it does] | High/Medium |
```

### Delegate to c3-container-design

> "I understand [Container Name]. I'll use the c3-container-design skill to create C3-X-[slug]."

Repeat for each container.

---

## Phase 4: Component Identification

**Goal:** Identify which components need documentation now vs later.

### Prioritization Questions

For each container's key components:

1. > "Which of these would cause the most confusion for a new developer?"
2. > "Which components have the most complex configuration?"
3. > "Which components integrate with external services?"

### Component Priority Matrix

| Priority | Criteria | Action |
|----------|----------|--------|
| **High** | Core business logic, External integrations, Complex config | Create full C3-XNN document |
| **Medium** | Important but straightforward | Create C3-XNN stub |
| **Low** | Utilities, Simple wrappers | Note in Container doc |

### Create Component Stubs

For High/Medium priority, create stub with:
- Overview (from user description)
- Purpose (responsibility)
- TODO markers for implementation details

### Delegate to c3-component-design

For high-priority components:
> "I'll use the c3-component-design skill to create detailed documentation for [Component Name]."

---

## Phase 5: Generate & Verify

### Generate TOC

```bash
chmod +x .c3/scripts/build-toc.sh
.c3/scripts/build-toc.sh
```

### Verification Checklist

Present to user:

```markdown
## C3 Adoption Complete

### Created:
- [ ] `.c3/CTX-*.md` - System context
- [ ] `.c3/containers/C3-1-*.md` - [Container 1]
- [ ] `.c3/containers/C3-2-*.md` - [Container 2]
- [ ] `.c3/components/*/C3-*.md` - [N] components
- [ ] `.c3/TOC.md` - Table of contents
- [ ] `.c3/scripts/build-toc.sh` - TOC generator

### Gaps Identified:
- [Any areas that need more detail]
- [Components marked for future documentation]

### Recommended Next Steps:
1. Review CTX-system-overview for accuracy
2. Fill in [specific gap]
3. Consider ADR for [identified decision]
```

---

## Socratic Principles

### Why Questions Over Scanning

| Scanning | Questioning |
|----------|-------------|
| Finds files | Finds purpose |
| Shows structure | Shows intent |
| Detects patterns | Discovers context |
| Misses relationships | Builds understanding |
| Can be wrong | User validates |

### Question Flow Pattern

```
Open Question → Listen → Reflect Back → Clarify → Confirm
```

**Example:**
1. "What containers exist?" (open)
2. Listen: "We have a Next.js app and a database"
3. Reflect: "So you have a fullstack Next.js container and PostgreSQL"
4. Clarify: "Is the database managed or self-hosted?"
5. Confirm: "Got it - Next.js frontend+API with managed PostgreSQL"

### When User Doesn't Know

If user is uncertain:
> "That's fine - let's mark this as TBD and move on. We can fill it in later."

Create placeholder in documentation:
```markdown
## [Section] {#id}
<!--
TODO: Needs clarification
-->

[To be documented]
```

---

## Delegation Reference

| Understanding Needed | Delegate To |
|---------------------|-------------|
| System boundary, actors, protocols | `c3-context-design` |
| Container structure, tech stack, APIs | `c3-container-design` |
| Implementation details, config, code | `c3-component-design` |
| TOC management | `c3-toc` |
| Document retrieval | `c3-locate` |

### Delegation Handoff

When delegating, provide:
1. The understanding you've built
2. Which document to create (CTX-slug, C3-2-slug, C3-201-slug, etc.)
3. Key sections to fill

---

## Appendix A: build-toc.sh

```bash
#!/bin/bash
# Build Table of Contents from C3 documentation
set -e

C3_ROOT=".c3"
TOC_FILE="${C3_ROOT}/TOC.md"
TEMP_FILE=$(mktemp)

echo "Building C3 Table of Contents..."

cat > "$TEMP_FILE" << 'EOF'
# C3 Documentation Table of Contents

> **AUTO-GENERATED** - Regenerate with: `.c3/scripts/build-toc.sh`
EOF

echo "" >> "$TEMP_FILE"
echo "> Last generated: $(date '+%Y-%m-%d %H:%M:%S')" >> "$TEMP_FILE"
echo "" >> "$TEMP_FILE"

# Extract frontmatter
extract_field() {
    awk -v field="$2" '/^---$/{f=!f;next} f && $0~"^"field":"{ sub("^"field": *",""); print; exit }' "$1"
}

# Context
for file in $(find "$C3_ROOT" -maxdepth 1 -name "CTX-*.md" 2>/dev/null | sort); do
    [ -z "$(cat "$file" 2>/dev/null)" ] && continue
    id=$(extract_field "$file" "id")
    title=$(extract_field "$file" "title")
    echo "## Context: [$id](./${id}.md)" >> "$TEMP_FILE"
    echo "$title" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
done

# Containers (C3-X-slug.md where X is single digit)
for file in $(find "$C3_ROOT/containers" -name "C3-[0-9]-*.md" 2>/dev/null | sort); do
    id=$(extract_field "$file" "id")
    title=$(extract_field "$file" "title")
    echo "## Container: [$id](./containers/${id}.md)" >> "$TEMP_FILE"
    echo "$title" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
done

# Components (C3-XNN-slug.md where X is container digit, NN is 01-99)
for dir in $(find "$C3_ROOT/components" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | sort); do
    container=$(basename "$dir")
    echo "## Components: ${container}" >> "$TEMP_FILE"
    for file in $(find "$dir" -name "C3-[0-9][0-9][0-9]-*.md" | sort); do
        id=$(extract_field "$file" "id")
        title=$(extract_field "$file" "title")
        echo "- [$id](./components/$container/${id}.md) - $title" >> "$TEMP_FILE"
    done
    echo "" >> "$TEMP_FILE"
done

# ADRs
for file in $(find "$C3_ROOT/adr" -name "ADR-*.md" 2>/dev/null | sort -r); do
    id=$(extract_field "$file" "id")
    title=$(extract_field "$file" "title")
    status=$(extract_field "$file" "status")
    echo "## ADR: [$id](./adr/${id}.md)" >> "$TEMP_FILE"
    echo "$title (${status:-proposed})" >> "$TEMP_FILE"
    echo "" >> "$TEMP_FILE"
done

mv "$TEMP_FILE" "$TOC_FILE"
echo "Done: $TOC_FILE"
```

---

## Example Session

```
User: Initialize C3 docs for my project
