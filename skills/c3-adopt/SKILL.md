---
name: c3-adopt
description: Initialize C3 architecture documentation for any project (existing codebase or fresh start) - uses Socratic questioning to build understanding, then delegates to layer skills for document creation
---

# C3 Adopt - Initialize Architecture Documentation

## Overview

Bootstrap C3 (Context-Container-Component) architecture documentation through Socratic questioning and delegation. Works for both existing codebases and fresh projects.

**Announce at start:** "I'm using the c3-adopt skill to initialize architecture documentation for this project."

## When to Use

| Scenario | Path |
|----------|------|
| Existing codebase needs C3 docs | Full workflow (explore → question → document) |
| Brand new project, no code yet | Fresh start (question → scaffold → document) |
| `.c3/` exists but needs rebuild | Ask user: update, backup+recreate, or abort |

## Quick Reference

| Phase | Key Activities | Output |
|-------|---------------|--------|
| **0. Project Detection** | Check for code, check for existing .c3/ | Determine path |
| **1. Discovery & Scaffolding** | Explore codebase OR Socratic discovery | `.c3/` directory, container map |
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

## Phase 0: Project Detection

### Step 0.1: Check for Existing .c3/

```bash
ls -la .c3 2>/dev/null && echo "EXISTS" || echo "MISSING"
```

**If EXISTS:** Ask user:
> "I found existing `.c3/` documentation. Would you like me to:
> 1. Review and update existing documentation
> 2. Back up and create fresh documentation
> 3. Abort and preserve what's there"

### Step 0.2: Detect Project Type

```bash
# Check if this is a fresh project (no meaningful code)
CODE_FILES=$(find . -maxdepth 3 -type f \( -name "*.js" -o -name "*.ts" -o -name "*.py" -o -name "*.go" -o -name "*.java" -o -name "*.rs" \) 2>/dev/null | head -5 | wc -l)
PACKAGE_FILES=$(find . -maxdepth 2 -type f \( -name "package.json" -o -name "go.mod" -o -name "requirements.txt" -o -name "Cargo.toml" -o -name "pom.xml" \) 2>/dev/null | wc -l)

if [ "$CODE_FILES" -eq 0 ] && [ "$PACKAGE_FILES" -eq 0 ]; then
    echo "FRESH_PROJECT"
else
    echo "EXISTING_CODEBASE"
fi
```

| Detection | Path |
|-----------|------|
| `FRESH_PROJECT` | Skip codebase exploration, use Socratic-only discovery |
| `EXISTING_CODEBASE` | Full exploration + Socratic refinement |

---

## Phase 1: Discovery & Scaffolding

### Path A: Existing Codebase (Full Exploration)

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

### Path B: Fresh Project (Socratic-Only Discovery)

When no code exists, build the container map through questions:

1. **System Identity**
   - "What is the name of this system/project?"
   - "In one sentence, what will it do for users?"

2. **Planned Architecture**
   - "What containers/services do you plan to build?"
   - "Will this be a monolith, microservices, or serverless?"

3. **Technology Choices**
   - "What languages/frameworks are you considering?"
   - "What databases or external services will you use?"

Build container map from answers:

```markdown
## Planned Containers

| # | Name | Archetype | Planned Technologies |
|---|------|-----------|---------------------|
| 1 | api | Backend | [TBD or specified] |
| 2 | web | Frontend | [TBD or specified] |

Note: Architecture based on planned design, not code discovery.
```

### Step 1.3: Scaffold .c3/ Directory

Create structure based on confirmed containers:

```bash
mkdir -p .c3/adr
```

The TOC can be rebuilt using the plugin's `build-toc.sh` script (ask Claude to "rebuild the TOC").

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

Create `README.md` (context) with v3 frontmatter and template:
```markdown
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
| Project settings | `c3-config` |

**References (not skills):**
- [lookup-patterns.md](../../references/lookup-patterns.md) - ID-based document retrieval
- [naming-conventions.md](../../references/naming-conventions.md) - Naming patterns

## Related Skills

- [c3-context-design](../c3-context-design/SKILL.md)
- [c3-container-design](../c3-container-design/SKILL.md)
- [c3-component-design](../c3-component-design/SKILL.md)
- [c3-config](../c3-config/SKILL.md)
