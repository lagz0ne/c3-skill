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
| **1. Discovery & Scaffolding** | Task Explore codebase, create scaffolding | `.c3/` directory, container map |
| **2. Context Discovery** | Socratic questions about system | Understanding for c3-0 |
| **3. Container Discovery** | Archetype-guided exploration, Socratic questions | Understanding for c3-{N} |
| **4. Component Identification** | Identify key components | c3-{N}{NN} stubs |
| **5. Configure Settings** | Call c3-config if settings.yaml missing | `.c3/settings.yaml` |
| **6. Generate & Verify** | Delegate to sub-skills, build TOC | Complete documentation |
| **7. Platform (Recommended)** | Deployment, networking, secrets, CI/CD | `.c3/platform/` docs |

## Guardrails

**Required reading before starting:**
- [derivation-guardrails.md](../../references/derivation-guardrails.md) - Hierarchy and abstraction rules
- [v3-structure.md](../../references/v3-structure.md) - File paths and ID patterns
- [role-taxonomy.md](../../references/role-taxonomy.md) - Component role vocabulary
- [archetype-hints.md](../../references/archetype-hints.md) - Container archetype patterns
- [discovery-questions.md](../../references/discovery-questions.md) - Socratic questioning guide

**Key rules:**
- **Reading order:** Context → Container → Component
- **Downward-only links:** Parent links to children, never reverse
- **Infra containers are leaf nodes:** No components beneath them

---

## Phase 1: Discovery & Scaffolding

### Step 1.1: Full Codebase Exploration

**Use Task Explore (very thorough) to build container map:**

Prompt for Task Explore:
```
Explore this codebase thoroughly to identify all deployable containers.

Look for:
1. Package managers: package.json, go.mod, requirements.txt, Cargo.toml, pom.xml
2. Entry points: main.*, index.*, app.*, server.*
3. Docker/container files: Dockerfile, docker-compose.yml
4. Framework indicators: next.config.*, nuxt.config.*, angular.json
5. Infrastructure: terraform/, helm/, k8s/

For each potential container found:
- Directory path
- Likely archetype (backend, frontend, worker, infrastructure)
- Key technologies detected
- Entry points found

Return a structured container map.
```

### Step 1.2: Present Container Map

<thinking>
Evaluate discovered containers:
- Did Task Explore find clear container boundaries?
- Are archetypes correctly identified (backend/frontend/worker/infra)?
- Any ambiguous cases (monorepo? shared code?)?
- What's missing that the user might expect?

Confidence assessment:
- High confidence: [containers with clear markers]
- Need confirmation: [ambiguous containers]
- Possibly missing: [gaps in typical architecture]
</thinking>

Show discovered containers to user:

```markdown
## Discovered Containers

| # | Path | Archetype | Technologies |
|---|------|-----------|--------------|
| 1 | /api | Backend | Node.js, Express, PostgreSQL |
| 2 | /web | Frontend | Next.js, React |
| 3 | /worker | Worker | Node.js, BullMQ |
| 4 | - | Infrastructure | PostgreSQL, Redis |

Does this match your understanding? Any containers missing or incorrectly identified?
```

Use AskUserQuestion if choices needed.

### Step 1.3: Check Prerequisites

```bash
ls -la .c3 2>/dev/null && echo "WARNING: .c3 already exists"
```

If `.c3/` exists, ask:
> "I found existing `.c3/` documentation. Would you like me to:
> 1. Review and update existing documentation
> 2. Back up and create fresh documentation
> 3. Abort and preserve what's there"

### Step 1.4: Scaffold .c3/ Directory

Create structure based on confirmed containers:

```bash
mkdir -p .c3/adr
```

The TOC can be rebuilt using the c3-toc skill.

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

Create container folders based on discovered containers:
```
.c3/
├── README.md           # Context stub
├── c3-1-{slug}/        # First container
│   └── README.md
├── c3-2-{slug}/        # Second container
│   └── README.md
├── adr/                # ADR directory
└── TOC.md              # Auto-generated TOC
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

<thinking>
Synthesize user answers into context model:
- System identity: Clear name and purpose?
- Boundary: What's inside vs external is defined?
- Containers: All identified and typed?
- Protocols: How containers communicate?
- Cross-cutting: Auth, logging, monitoring covered?

Gaps remaining:
- [any unanswered questions]
- [areas needing follow-up]

Ready to delegate to c3-context-design? [yes/no]
</thinking>

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

**Goal:** For each container identified in Phase 1, build understanding through questions.

### Step 3.1: Archetype-Guided Exploration

Reference [archetype-hints.md](../../references/archetype-hints.md) for the container's archetype.

Use Task Explore to discover:
- Component structure within the container
- Dependencies (upstream, downstream)
- Technology specifics

### Step 3.2: Socratic Refinement

Using [discovery-questions.md](../../references/discovery-questions.md), ask container-level questions:
- Identity: "What is this container's primary responsibility?"
- Technology: Use AskUserQuestion with discovered options
- Dependencies: "What does this call? What calls this?"
- Testing: "How is this tested?"

### Container Questions (per container)

1. **Identity:** "What is [Container]'s single main responsibility?"
2. **Technology:** "What language and framework does it use?"
3. **Structure:** "How is code organized inside?"
4. **APIs:** "What endpoints does it expose? What does it consume?"
5. **Data:** "What data does it own vs read from others?"
6. **Key Components:** "What are the 3-5 most important components?"

### Step 3.3: Delegate to c3-container-design

> "I understand [Container Name]. I'll use the c3-container-design skill to create the documentation."

**V3 Container Creation:**
1. Create folder: `mkdir -p .c3/c3-{N}-{slug}/`
2. Create doc as `README.md` inside with `id: c3-{N}`

---

## Phase 4: Component Identification

**Goal:** Identify which components need documentation now vs later.

<thinking>
Prioritize components for documentation:
- Which are core to business logic? → High priority
- Which have complex configuration? → High priority
- Which integrate with external services? → High priority
- Which are simple utilities/wrappers? → Low priority

For each container, list components by priority:
- Container 1: [high], [medium], [low]
- Container 2: [high], [medium], [low]
...
</thinking>

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

## Phase 5: Configure Settings

### Check for settings.yaml

After scaffolding, check if `.c3/settings.yaml` exists:

```bash
ls .c3/settings.yaml 2>/dev/null && echo "EXISTS" || echo "MISSING"
```

**If MISSING:**
- Announce: "I'll use the c3-config skill to set up project settings."
- Use the `c3-config` skill
- When called from c3-adopt, c3-config creates defaults and offers customization

**If EXISTS:**
- Skip, settings already configured

---

## Phase 6: Generate & Verify

### Generate TOC

Ask Claude to rebuild the TOC (uses plugin's build-toc.sh script).

### Verification Checklist

- [ ] `.c3/README.md` - Context (id: c3-0, c3-version: 3)
- [ ] `.c3/c3-{N}-{slug}/README.md` - Each container
- [ ] `.c3/c3-{N}-{slug}/c3-{N}{NN}-*.md` - Priority components
- [ ] `.c3/TOC.md` - Table of contents

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
3. **Consider Phase 7:** Document platform concerns (deployment, networking, secrets, CI/CD)
   - Even partial platform docs reduce onboarding friction
   - Use `.c3/platform/` for deployment-related documentation
```

---

## Phase 7: Platform Documentation (Recommended)

> **Why recommended?** Platform concerns (deployment, networking, secrets, CI/CD) are often the #1 source of onboarding confusion. Even partial documentation helps.

Ask user:
```
Platform documentation is recommended. Which areas should we cover?

[ ] Deployment strategy - How and where the system runs
[ ] Networking topology - How containers communicate
[ ] Secrets management - How credentials are handled
[ ] CI/CD pipeline - How changes flow to production

Select all that apply, or skip if truly not applicable.
```

If any selected:
1. Create `.c3/platform/` directory
2. Reference [platform-patterns.md](../../references/platform-patterns.md)
3. Use discovery questions for each area
4. Accept TBD for unknowns

### Platform Structure

```
.c3/platform/
├── deployment.md      # c3-0-deployment
├── networking.md      # c3-0-networking
├── secrets.md         # c3-0-secrets
└── ci-cd.md           # c3-0-cicd
```

Platform docs use `c3-0-*` IDs to indicate they're Context-level.

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
