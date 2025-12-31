# Template-Based C3 Adoption Design

**Date:** 2025-12-31
**Status:** Proposed

## Summary

Simplify C3 adoption from discovery-heavy workflow to template-first approach. Copy templates, replace tokens, let AI fill inventories. 80% structure from templates, 20% content from AI.

## Problem

Current adoption workflow:
1. Loads abstract references into context
2. Runs discovery loops with overlay presentations
3. Requires multiple confirmation gates
4. Generates docs from scratch

**Issues:**
- Context-heavy (hallucination risk)
- Time-consuming
- Over-engineered for the task

## Solution

**Two-round template-based adoption:**

```
Round 1 (Bash)              Round 2 (AI Subagent)
─────────────────           ─────────────────────
Copy templates        →     Analyze codebase
Replace tokens        →     Fill inventories
Create structure      →     Add diagrams + linkages
                      →     User reviews final result
```

## Philosophy

### Layer Hierarchy

```
CONTEXT (Strategic)
│   WHAT containers exist
│   WHY they connect (inter-container protocols)
│
└──→ CONTAINER (Bridge: Tactical for Context, Strategic for Components)
     │   Receives protocol expectations from Context
     │   WHAT components fulfill those protocols
     │   WHY components connect internally
     │
     └──→ COMPONENT (Tactical)
          HOW the protocol is actually implemented
```

### Container as Bridge

Container fulfills connections defined at Context level:
- Context says: "API ←→ Database (persistence)"
- Container maps: "DB Client component handles persistence"
- Component details: "Uses pg-pool with connection pooling..."

### Consistent Document Structure

Every C3 document follows:

```
┌─────────────────────────────────────┐
│ 1. DIAGRAM (mermaid)                │
│    Visual grab of connectivity      │
│    Uses IDs for quick reference     │
├─────────────────────────────────────┤
│ 2. INVENTORY TABLE                  │
│    ID | Name | Type | Status        │
│    Consistent structure across all  │
├─────────────────────────────────────┤
│ 3. LINKAGES                         │
│    From → To + WHY (reasoning)      │
│    Not just THAT they connect       │
└─────────────────────────────────────┘
```

## Templates

### Template 1: Context (c3-0)

```markdown
---
id: c3-0
c3-version: 3
title: {{PROJECT_NAME}}
summary: {{ONE_LINE_SUMMARY}}
---

# {{PROJECT_NAME}}

<!-- AI: 2-3 sentences on what this system does and why it exists -->

## Overview

​```mermaid
graph LR
    %% Actors
    A1[User]
    A2[Scheduler]

    %% Containers
    c3-1[API Backend]
    c3-2[Frontend]
    c3-3[Worker]

    %% External
    E1[(Database)]
    E2[Stripe API]

    %% Connections
    A1 -->|HTTP| c3-2
    c3-2 -->|REST| c3-1
    c3-1 -->|SQL| E1
    c3-1 -->|events| c3-3
    c3-1 -->|payment| E2
    A2 -->|trigger| c3-3
​```

## Actors

| ID | Actor | Type | Interacts With | Purpose |
|----|-------|------|----------------|---------|
| A1 | User | user | c3-2 | Primary system user |
| A2 | Scheduler | scheduled | c3-3 | Triggers background jobs |

## Containers

| ID | Name | Type | Status | Purpose |
|----|------|------|--------|---------|
| c3-1 | API Backend | service | | Business logic |
| c3-2 | Frontend | app | | User interface |
| c3-3 | Worker | service | | Background processing |

## External Systems

| ID | Name | Type | Purpose |
|----|------|------|---------|
| E1 | PostgreSQL | database | Primary persistence |
| E2 | Stripe | api | Payment processing |

## Linkages

| From | To | Protocol | Reasoning |
|------|-----|----------|-----------|
| A1 | c3-2 | HTTP | Users access via browser |
| c3-2 | c3-1 | REST | Frontend fetches data, API owns logic |
| c3-1 | E1 | SQL | API owns persistence |
| c3-1 | c3-3 | events | Async work decoupled from request |
| c3-1 | E2 | HTTPS | Payment delegation |
| A2 | c3-3 | trigger | Scheduled job initiation |
```

### Template 2: Container (c3-N)

```markdown
---
id: c3-{{N}}
c3-version: 3
title: {{CONTAINER_NAME}}
type: container
parent: c3-0
summary: {{ONE_LINE_SUMMARY}}
---

# {{CONTAINER_NAME}}

<!-- AI: 2-3 sentences on what this container does and its role in the system -->

## Overview

​```mermaid
graph TD
    %% Foundation
    c3-{{N}}01[Request Handler]
    c3-{{N}}02[Auth Middleware]

    %% Auxiliary
    c3-{{N}}03[DB Client]

    %% Business
    c3-{{N}}05[User Service]

    %% Flows
    c3-{{N}}01 -->|validates| c3-{{N}}02
    c3-{{N}}01 -->|delegates| c3-{{N}}05
    c3-{{N}}05 -->|persists| c3-{{N}}03
​```

## Components

### Foundation

| ID | Name | Concern | Status | Responsibility |
|----|------|---------|--------|----------------|
| c3-{{N}}01 | Request Handler | entry | | Receives external calls |
| c3-{{N}}02 | Auth Middleware | identity | | Validates caller identity |

### Auxiliary

| ID | Name | Concern | Status | Responsibility |
|----|------|---------|--------|----------------|
| c3-{{N}}03 | DB Client | library-wrapper | | Persistence abstraction |

### Business

| ID | Name | Concern | Status | Responsibility |
|----|------|---------|--------|----------------|
| c3-{{N}}05 | User Service | domain | | User operations |

### Presentation

N/A

## Fulfillment

<!-- How this container fulfills connections from Context (c3-0) -->

| Link (from c3-0) | Fulfilled By | Constraints |
|------------------|--------------|-------------|
| A1 → c3-{{N}} via HTTP | c3-{{N}}01 | REST only, JSON body, requires CORS |
| c3-{{N}} → E1 via SQL | c3-{{N}}03 | PostgreSQL 14+, connection pooling |

## Linkages

| From | To | Contract | Reasoning |
|------|-----|----------|-----------|
| c3-{{N}}01 | c3-{{N}}02 | validate-request | Every request needs identity before processing |
| c3-{{N}}01 | c3-{{N}}05 | user-ops | Handler delegates, doesn't own business logic |
| c3-{{N}}05 | c3-{{N}}03 | persist | Service owns logic, DB Client owns storage |
```

**Concern categories (all present, ready for growth):**

| Category | Purpose | Examples |
|----------|---------|----------|
| Foundation | How container operates | Entry, identity, integration |
| Auxiliary | Consistency enforcers | Library wrappers, framework usage, utilities |
| Business | Domain flows | Services organized by domain |
| Presentation | UI concerns (frontend) | Styling, composition, state |

### Template 3: Component (c3-NNN)

```markdown
---
id: c3-{{N}}{{NN}}
c3-version: 3
title: {{COMPONENT_NAME}}
type: component
parent: c3-{{N}}
summary: {{ONE_LINE_SUMMARY}}
---

# {{COMPONENT_NAME}}

<!-- AI: 2-3 sentences on what this component does -->

## Overview

​```mermaid
graph LR
    %% Inputs
    IN1([HTTP Request]) --> c3-{{N}}{{NN}}
    IN2([Auth Token]) --> c3-{{N}}{{NN}}

    %% This component
    c3-{{N}}{{NN}}[{{COMPONENT_NAME}}]

    %% Outputs
    c3-{{N}}{{NN}} --> OUT1([Validated Request])
    c3-{{N}}{{NN}} --> OUT2([Error Response])
​```

## Interface

| Direction | What | Format | From/To |
|-----------|------|--------|---------|
| IN | HTTP Request | Express.Request | External |
| IN | Auth Token | JWT string | Header |
| OUT | Validated Request | Request + User | c3-{{N}}05 |
| OUT | Error Response | 401/403 JSON | External |

## Hand-offs

| To | What | Contract | Mechanism |
|----|------|----------|-----------|
| c3-{{N}}05 | Validated request | User attached to context | Next middleware |
| External | Auth failure | Standard error shape | HTTP response |

## Implementation

### Technology

| Aspect | Choice | Rationale |
|--------|--------|-----------|
| Framework | Express.js | Project standard |
| JWT Library | jsonwebtoken | Widely supported |
| Validation | zod | Type-safe schemas |

### Conventions

- All routes under `/api/*` pass through this middleware
- Token must be in `Authorization: Bearer <token>` header
- User object attached to `req.user` after validation

### Edge Cases

| Scenario | Behavior |
|----------|----------|
| Missing token | 401 Unauthorized |
| Expired token | 401 + `token_expired` code |
| Invalid signature | 401 + `invalid_token` code |
| Valid token, user deleted | 403 Forbidden |
```

### Template 4: ADR-000 (Initialization)

```markdown
---
id: adr-00000000-c3-adoption
c3-version: 3
title: C3 Architecture Documentation Adoption
type: adr
status: implemented
date: {{DATE}}
affects: [c3-0]
---

# C3 Architecture Documentation Adoption

## Overview

​```mermaid
graph TD
    ADR[This ADR] -->|establishes| C3[C3 Structure]
    C3 -->|contains| CTX[c3-0 Context]
    CTX -->|contains| CON[Containers]
    CON -->|contains| CMP[Components]
​```

## Status

**Implemented** - {{DATE}}

## Problem

| Situation | Impact |
|-----------|--------|
| No architecture docs | Onboarding takes weeks |
| Knowledge in heads | Bus factor risk |
| Ad-hoc decisions | Inconsistent patterns |

## Decision

Adopt C3 (Context-Container-Component) methodology for architecture documentation.

## Structure Created

| Level | ID | Name | Purpose |
|-------|-----|------|---------|
| Context | c3-0 | {{PROJECT_NAME}} | System overview |
| Container | c3-1 | {{CONTAINER_1}} | {{PURPOSE_1}} |
| Container | c3-2 | {{CONTAINER_2}} | {{PURPOSE_2}} |

## Rationale

| Consideration | C3 Approach |
|---------------|-------------|
| Layered abstraction | Context → Container → Component |
| Change isolation | Strategic vs tactical separation |
| Growth-ready | Structure exists before content |
| Decision tracking | ADRs capture evolution |

## Consequences

### Positive

- Architecture visible and navigable
- Onboarding accelerated
- Decisions documented with reasoning

### Negative

- Maintenance overhead (docs can drift)
- Initial time investment

## Verification

- [ ] `.c3/README.md` exists (c3-0)
- [ ] All containers have `README.md`
- [ ] Diagrams use consistent IDs
- [ ] Linkages have reasoning

## Audit Record

| Phase | Date | Notes |
|-------|------|-------|
| Adopted | {{DATE}} | Initial C3 structure created |
```

## Adoption Flow

### Round 1: Bash Script

```bash
#!/bin/bash
# c3-init.sh

PROJECT="${PROJECT:-MyProject}"
C1="${C1:-backend}"
C2="${C2:-frontend}"
DATE=$(date +%Y-%m-%d)

# Create structure
mkdir -p .c3/c3-1-$C1
mkdir -p .c3/c3-2-$C2
mkdir -p .c3/adr

# Copy and replace templates
envsubst < templates/context.md > .c3/README.md
envsubst < templates/container.md > .c3/c3-1-$C1/README.md
envsubst < templates/container.md > .c3/c3-2-$C2/README.md
envsubst < templates/adr-000.md > .c3/adr/adr-00000000-c3-adoption.md
```

### Round 2: AI Subagent

Subagent receives:
- Template structure already in place
- Tokens already replaced

Subagent does:
1. Analyze codebase (same discovery logic)
2. Fill inventory tables
3. Create mermaid diagrams with IDs
4. Add linkages with reasoning
5. Complete fulfillment sections

User reviews final result (no intermediate confirmations).

## What Changes

### Removed

- `references/adopt-workflow.md` - replaced by templates
- `references/discovery-engine.md` - simplified, subagent prompts in skill
- Overlay presentation logic
- Phase gates and confirmation loops

### Added

- `templates/context.md`
- `templates/container.md`
- `templates/component.md`
- `templates/adr-000.md`
- `scripts/c3-init.sh`

### Modified

- `skills/c3/SKILL.md` - Adopt mode uses templates, simpler flow

## Migration

No user migration needed. This changes how adoption works, not the output structure. Existing `.c3/` directories remain valid.

## Verification

- [ ] Templates produce valid C3 structure
- [ ] Bash script creates correct directories
- [ ] AI subagent fills all sections
- [ ] Output matches v3-structure.md requirements
