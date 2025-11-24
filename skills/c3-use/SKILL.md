---
name: c3-use
description: Entry point for using C3 architecture docs - checks for .c3/ directory, guides reading if exists, offers adoption if not
---

# C3 Use - Getting Started with Architecture Docs

## Overview

Entry point for working with C3 architecture documentation. Checks if `.c3/` exists and guides accordingly.

**Announce at start:** "I'm using the c3-use skill to help you work with architecture documentation."

## Quick Reference

| Scenario | Action |
|----------|--------|
| `.c3/` exists | Guide user through the documentation |
| `.c3/` missing | Offer to adopt C3 for the project |

## The Process

### Step 1: Check for .c3/ Directory

```bash
ls -d .c3 2>/dev/null && echo "EXISTS" || echo "MISSING"
```

**If MISSING:** Go to [No C3 Found](#no-c3-found)

**If EXISTS:** Continue to Step 2

### Step 2: Load and Present Overview

Read the Context document:
```bash
cat .c3/README.md
```

Read the TOC:
```bash
cat .c3/TOC.md
```

Present a brief orientation:

```markdown
## This Project Uses C3 Architecture Documentation

**System:** [Name from README.md]
**Purpose:** [One-line summary]

### Structure
- **Context** (c3-0): System overview, boundaries, actors
- **Containers** (c3-1, c3-2, ...): Deployable units
- **Components** (c3-101, c3-102, ...): Internal parts of containers
- **ADRs**: Architecture Decision Records

### Quick Navigation
| Document | What You'll Learn |
|----------|-------------------|
| `.c3/README.md` | System overview and boundaries |
| `.c3/TOC.md` | Full documentation index |
| `.c3/c3-{N}-*/README.md` | Container details |
| `.c3/adr/` | Decision history |

### How to Read
1. Start with **README.md** (Context) for the big picture
2. Pick a container relevant to your work
3. Drill into components as needed
4. Check ADRs for "why" behind decisions

What would you like to explore?
```

### Step 3: Help Navigate

Based on user's interest, use appropriate skills:

| User Wants | Action |
|------------|--------|
| Understand a specific part | Use `c3-locate` to find and show content |
| Make changes | Use `c3-design` to guide architecture updates |
| Find decision history | Point to `.c3/adr/` or use `c3-locate` |
| Update settings | Use `c3-config` |

---

## No C3 Found

When `.c3/` directory doesn't exist:

```markdown
This project doesn't have C3 architecture documentation yet.

C3 (Context-Container-Component) helps document:
- System boundaries and actors
- Deployable containers and their responsibilities
- Key components and their configurations
- Architecture decisions and their rationale

Would you like to set up C3 documentation for this project?
```

**If user says yes:**
- Announce: "I'll use the c3-adopt skill to initialize architecture documentation."
- Use the `c3-adopt` skill

**If user says no:**
- Acknowledge and offer to help with other tasks

## C3 Structure Explained

For users unfamiliar with C3, explain:

### Layers

| Layer | ID Pattern | Purpose | Example |
|-------|------------|---------|---------|
| **Context** | `c3-0` | System boundary, actors, external integrations | The whole system |
| **Container** | `c3-{N}` | Deployable unit (service, app, database) | Backend API, Frontend |
| **Component** | `c3-{N}{NN}` | Internal part of a container | Auth middleware, DB pool |

### ID System

- `c3-0` - Always the Context (system overview)
- `c3-1`, `c3-2` - Containers (numbered 1-9)
- `c3-101`, `c3-102` - Components in container 1
- `c3-201`, `c3-215` - Components in container 2
- `adr-YYYYMMDD-slug` - Architecture Decision Records

### Reading Order

```
Context (c3-0)
    ↓
Container of interest (c3-1, c3-2, ...)
    ↓
Components as needed (c3-101, c3-102, ...)
    ↓
Related ADRs for decision history
```

## Related Skills

| Skill | When to Use |
|-------|-------------|
| `c3-adopt` | Initialize C3 for a project |
| `c3-design` | Make architecture changes |
| `c3-locate` | Find specific documents by ID |
| `c3-config` | Update project settings |
