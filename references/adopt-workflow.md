# Adopt Workflow (New Project Setup)

**Trigger:** No `.c3/` directory exists, or user asks to "set up C3", "create architecture docs"

**CRITICAL:** Complete ALL layers atomically. Do NOT delegate to other skills or end early.

## Step 1: Quick Discovery (non-interactive)

Make reasonable assumptions from user's description:
- System purpose
- Main containers (typically 1-3)
- External actors
- Key technologies mentioned

## Step 2: Create Structure

```bash
mkdir -p .c3/adr
```

## Step 3: Create Context (c3-0) - IMMEDIATELY

Write `.c3/README.md` with:

```yaml
---
id: c3-0
c3-version: 3
title: [System Name]
summary: [One-line description]
---
```

Required sections:
- Overview (what the system does)
- Containers table (ID, Name, Purpose)
- Container Interactions (Mermaid diagram)
- External Actors

## Step 4: Create Containers - IN SAME PASS

For EACH container identified:

1. Create folder: `.c3/c3-{N}-{slug}/`
2. Write `README.md` with:

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

Required sections:
- Technology Stack table
- Components table (ID, Name, Responsibility)
- Internal Structure (Mermaid diagram)

## Step 5: Create TOC

Write `.c3/TOC.md` listing all created docs.

## Step 6: Confirm Completion

```
C3 documentation created:
├─ .c3/README.md (Context)
├─ .c3/c3-1-{slug}/README.md (Container)
├─ .c3/c3-2-{slug}/README.md (Container)
└─ .c3/TOC.md

Next: Add component docs as you build, or create ADRs for architectural decisions.
```

**NEVER end before completing all containers.**
