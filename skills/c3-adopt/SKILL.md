---
name: c3-adopt
description: Initialize C3 architecture documentation for an existing project - uses Socratic questioning to build understanding, then delegates to layer skills for document creation
---

# C3 Adopt - Initialize Architecture Documentation

## Overview

Bootstrap C3 (Context-Container-Component) architecture documentation for an existing codebase through Socratic questioning and delegation.

**Announce at start:** "I'm using the c3-adopt skill to initialize architecture documentation for this project."

## Quick Reference

| Phase | Key Activities | Output |
|-------|---------------|--------|
| **1. Establish** | Check prerequisites, create scaffolding | `.c3/` directory |
| **2. Context Discovery** | Socratic questions about system | Understanding for c3-0 |
| **3. Container Discovery** | Socratic questions per container | Understanding for c3-{N} |
| **4. Component Identification** | Identify key components | c3-{N}{NN} stubs |
| **5. Generate & Verify** | Delegate to sub-skills, build TOC | Complete documentation |

## Guardrails

See [derivation-guardrails.md](../../references/derivation-guardrails.md) for complete rules. Key points:
- **Reading order:** Context → Container → Component
- **Downward-only links:** Parent links to children, never reverse
- **Infra containers are leaf nodes:** No components beneath them
- **Naming:** Use [v3-structure.md](../../references/v3-structure.md) patterns

---

## Phase 1: Establish

### Check Prerequisites

```bash
ls -la .c3 2>/dev/null && echo "WARNING: .c3 already exists"
```

If `.c3/` exists, ask:
> "I found existing `.c3/` documentation. Would you like me to:
> 1. Review and update existing documentation
> 2. Back up and create fresh documentation
> 3. Abort and preserve what's there"

### Create Scaffolding

```bash
mkdir -p .c3/{adr,scripts}
```

Copy `build-toc.sh` from the plugin's `.c3/scripts/` directory.

Create `index.md`:
```markdown
---
layout: home
title: C3 Architecture Documentation
---

# C3 Architecture Documentation

- [Table of Contents](./TOC.md)
- [System Overview](./README.md)
```

Create `README.md` (context) with v3 frontmatter:
```markdown
---
id: c3-0
c3-version: 3
title: System Overview
---

# System Overview

<!-- Context document content goes here -->
```

---

## Phase 2: Context Discovery

**Goal:** Build understanding through questions, NOT code scanning.

See [socratic-method.md](../../references/socratic-method.md) for question techniques.

### Context Questions (ask in order)

1. **System Identity**
   - "What is the name of this system/project?"
   - "In one sentence, what does it do for users?"

2. **System Boundary**
   - "What is INSIDE your system vs EXTERNAL?"
   - "Who or what interacts with your system from outside?"

3. **Containers**
   - "If you deployed this today, what separate processes would run?"
   - "What data stores exist?"

4. **Communication**
   - "How do these pieces talk to each other?"
   - "What external services does your system call?"

5. **Cross-Cutting**
   - "How is authentication handled across the system?"
   - "How do you handle logging and monitoring?"

### Build Context Model

From answers, construct:

```markdown
## Context Understanding

**System:** [Name]
**Purpose:** [One sentence]

**Actors:**
- [Actor 1]

**Containers:**
| Name | Type | Purpose |
|------|------|---------|
| [Name] | [Backend/Frontend/DB] | [What it does] |

**Protocols:**
| From | To | Protocol |
|------|-----|----------|
| [A] | [B] | [REST/gRPC] |
```

### Delegate to c3-context-design

> "I now understand your system context. I'll use the c3-context-design skill to create README.md."

---

## Phase 3: Container Discovery

**Goal:** For each container identified, build understanding through questions.

### Container Questions (per container)

1. **Identity:** "What is [Container]'s single main responsibility?"
2. **Technology:** "What language and framework does it use?"
3. **Structure:** "How is code organized inside?"
4. **APIs:** "What endpoints does it expose? What does it consume?"
5. **Data:** "What data does it own vs read from others?"
6. **Key Components:** "What are the 3-5 most important components?"

### Delegate to c3-container-design

> "I understand [Container Name]. I'll use the c3-container-design skill to create the documentation."

**V3 Container Creation:**
1. Create folder: `mkdir -p .c3/c3-{N}-{slug}/`
2. Create doc as `README.md` inside with `id: c3-{N}`

---

## Phase 4: Component Identification

**Goal:** Identify which components need documentation now vs later.

### Prioritization Questions

- "Which components would cause the most confusion for a new developer?"
- "Which have the most complex configuration?"
- "Which integrate with external services?"

### Priority Matrix

| Priority | Criteria | Action |
|----------|----------|--------|
| **High** | Core business logic, external integrations | Full document |
| **Medium** | Important but straightforward | Stub document |
| **Low** | Utilities, simple wrappers | Note in container |

### Create Component Stubs

Components go in container folders: `.c3/c3-{N}-{slug}/c3-{N}{NN}-{slug}.md`

For high-priority, delegate to c3-component-design.

---

## Phase 5: Generate & Verify

### Generate TOC

```bash
chmod +x .c3/scripts/build-toc.sh
.c3/scripts/build-toc.sh
```

### Verification Checklist

- [ ] `.c3/README.md` - Context (id: c3-0, c3-version: 3)
- [ ] `.c3/c3-{N}-{slug}/README.md` - Each container
- [ ] `.c3/c3-{N}-{slug}/c3-{N}{NN}-*.md` - Priority components
- [ ] `.c3/TOC.md` - Table of contents
- [ ] `.c3/scripts/build-toc.sh` - TOC generator

### Present Summary

```markdown
## C3 Adoption Complete

### Structure Created:
.c3/
├── README.md           # Context (c3-0)
├── TOC.md
├── c3-1-{slug}/        # Container 1
│   ├── README.md
│   └── c3-101-*.md     # Components
├── c3-2-{slug}/        # Container 2
│   └── README.md
├── adr/
└── scripts/

### Gaps Identified:
- [Areas needing more detail]

### Next Steps:
1. Review README.md for accuracy
2. Fill in [specific gap]
```

---

## Delegation Reference

| Understanding Needed | Delegate To |
|---------------------|-------------|
| System boundary, actors, protocols | `c3-context-design` |
| Container structure, tech stack | `c3-container-design` |
| Implementation details, config | `c3-component-design` |
| TOC management | `c3-toc` |
| Document retrieval | `c3-locate` |

## Related Skills

- [c3-context-design](../c3-context-design/SKILL.md)
- [c3-container-design](../c3-container-design/SKILL.md)
- [c3-component-design](../c3-component-design/SKILL.md)
