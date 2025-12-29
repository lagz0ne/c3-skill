# Adopt Workflow

**Trigger:** No `.c3/` directory exists, or user asks to "set up C3", "create architecture docs"

## Greenfield vs Brownfield

| Scenario | Discovery | Approach |
|----------|-----------|----------|
| **Greenfield** | User describes what they'll build | Create from description |
| **Brownfield** | Existing codebase | Explore code, document current state |

## Step 1: Discovery

### Greenfield
From user description, identify:
- System purpose
- Main containers (typically 1-3)
- Technology choices
- External actors

### Brownfield
Explore codebase to identify:
- Entry points (what runs?)
- Boundaries (what's separate?)
- Technologies used
- External integrations

## Step 2: Create Structure

```bash
mkdir -p .c3/adr
```

## Step 3: Create Context (c3-0)

Write `.c3/README.md`:

```yaml
---
id: c3-0
c3-version: 3
title: [System Name]
summary: [One-line description]
---
```

**Required sections:**
- Overview
- Containers table (ID, Name, Responsibility)
- Interactions (Mermaid diagram)
- External Actors

## Step 4: Create Containers

For EACH container:

1. Create folder: `.c3/c3-{N}-{slug}/`
2. Write `README.md`:

```yaml
---
id: c3-{N}
c3-version: 3
title: [Container Name]
type: container
parent: c3-0
summary: [One-line description]
---
```

**Required sections:**
- Inherited From Context
- Technology Stack table
- Components table (inventory)
- Internal Structure (Mermaid diagram)

## Step 5: Component Inventory (NOT Docs)

**CRITICAL: Inventory-first model**

List ALL components in Container's components table:

```markdown
## Components

| ID | Name | Type | Responsibility |
|----|------|------|----------------|
| c3-101 | Request Handler | Foundation | HTTP routing |
| c3-102 | Auth Service | Business | Authentication |
| c3-103 | User Service | Business | User management |
```

**Do NOT create component docs at adopt time.**

Component docs appear later when:
- Conventions emerge that consumers must follow
- Hand-off patterns become non-obvious
- Edge cases need documentation

## Step 6: Create TOC

Write `.c3/TOC.md` listing all created docs.

## Step 7: Confirm Completion

```
C3 documentation created:
├─ .c3/README.md (Context)
├─ .c3/c3-1-{slug}/README.md (Container)
├─ .c3/c3-2-{slug}/README.md (Container)
└─ .c3/TOC.md

Components inventoried but not documented yet.
Create component docs when conventions mature.

Next:
- Use `c3-structure` skill for structural changes
- Use `c3-implementation` skill when documenting components
- Create ADRs for architectural decisions
```

## Atomic Completion

**NEVER end before completing all containers.**

Adopt creates the full structure in one pass. Individual component docs come later through normal development flow.
