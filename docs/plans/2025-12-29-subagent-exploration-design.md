# C3 Discovery Engine

**Status:** Design Complete
**Date:** 2025-12-30
**Covers:** Adopt + Audit (unified engine)

---

## Problem

The c3 agent's adopt and audit modes do heavy sequential exploration that:
1. Exhausts context (reading many files)
2. Takes too long (no parallelization)
3. Doesn't leverage C3's hierarchical model

## Solution

A **hierarchical discovery engine** that:
- Follows C3's top-down structure (Context → Container → Component)
- Uses subagents (Sonnet) to offload exploration
- Presents findings via AskUserQuestion for efficient user confirmation
- Serves both Adopt and Audit modes

---

## Core Model: Expectation vs Reality

```
┌─────────────────────────────────────────────────────────────┐
│                    DISCOVERY ENGINE                          │
│  (Unified for Adopt and Audit)                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  REALITY (from code)        EXPECTATION (from .c3/ docs)    │
│        ↓                            ↓                        │
│  Discovery subagents          Read existing docs            │
│        ↓                            ↓                        │
│        └────────── COMPARE ─────────┘                       │
│                       ↓                                      │
│         ┌─────────────┴─────────────┐                       │
│         ↓                           ↓                        │
│     ADOPT                        AUDIT                       │
│  (no docs exist)             (docs exist)                   │
│  Reality + User → Create     Reality vs Expectation → Drift │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Hierarchical Discovery (All Layers Required)

Discovery follows C3's top-down hierarchy. Each layer's inventory informs what to search for in the next layer.

```
┌─────────────────────────────────────────────────────────────┐
│  LAYER 1: CONTEXT DISCOVERY                                 │
├─────────────────────────────────────────────────────────────┤
│  Subagent searches for:                                     │
│  - Apps / deployable units                                  │
│  - Connections between units                                │
│  - External actors (DBs, APIs, services)                    │
│                                                              │
│  Returns: Container inventory candidates                     │
│                                                              │
│  ASK USER: Confirm containers + responsibilities            │
└─────────────────────────────────────────────────────────────┘
                              ↓
                    (uses confirmed inventory)
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  LAYER 2: CONTAINER DISCOVERY (per confirmed container)     │
├─────────────────────────────────────────────────────────────┤
│  Subagent searches for:                                     │
│  - Internal organization / separation of concerns           │
│  - Significant modules within the container                 │
│  - Tech stack per container                                 │
│                                                              │
│  Returns: Component inventory candidates                     │
│                                                              │
│  ASK USER: Confirm components per container                 │
└─────────────────────────────────────────────────────────────┘
                              ↓
                    (uses confirmed inventory)
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  LAYER 3: COMPONENT DISCOVERY (per confirmed component)     │
├─────────────────────────────────────────────────────────────┤
│  Subagent searches for:                                     │
│  - Implementation patterns                                  │
│  - Connections to other components                          │
│  - Type classification (foundation/business/integration)    │
│                                                              │
│  Returns: Component details (type, responsibility, status)  │
│                                                              │
│  ASK USER: Confirm component roles + relationships          │
│                                                              │
│  ⚠️  CRITICAL: This is where code meets strategy            │
│  - Adopt: Not finished until component inventory complete   │
│  - Audit: Biggest drifts are at component level             │
└─────────────────────────────────────────────────────────────┘
```

---

## Subagent Specifications

All subagents use **Sonnet** model and are primed with C3 inventory context.

### Context Discovery Subagent

**Prompt context:**
```
You are scanning a codebase to identify containers for C3 documentation.

A "container" in C3 is a deployable unit or major boundary:
- Separate apps (frontend, backend, mobile)
- Services in a microservices architecture
- Major packages in a monorepo

Find:
1. Deployable units / entry points
2. How they connect to each other
3. External systems they depend on

Return findings aligned to this inventory table:
| ID | Container | Responsibility |
```

**Searches:**
- Directory tree (depth 2-3)
- Entry points (main files, Dockerfiles, package.json scripts)
- Package manager signals (workspaces, multi-package)
- Infra signals (docker-compose, k8s configs)

**Returns:**
```yaml
containers:
  - path: apps/backend
    name_hint: Backend API
    entry_points: [src/main.ts, Dockerfile]
    confidence: high

  - path: apps/frontend
    name_hint: Frontend App
    entry_points: [src/index.tsx]
    confidence: high

externals:
  - name: PostgreSQL
    type: database
    evidence: [pg dependency, DATABASE_URL]

connections:
  - from: frontend
    to: backend
    type: http
    evidence: [API_URL in frontend config]
```

### Container Discovery Subagent

**Input:** Confirmed container from Layer 1 + its path

**Prompt context:**
```
You are scanning a container to identify components for C3 documentation.

A "component" in C3 is a significant internal unit:
- Foundation: shared infrastructure (DB, HTTP, auth middleware)
- Business: domain logic (user service, order processing)
- Integration: external adapters (API clients, message handlers)

Scan the internal structure of: {container_path}

Return findings aligned to this inventory table:
| ID | Component | Type | Responsibility |
```

**Searches:**
- Internal directory structure
- Module organization patterns
- Dependency imports between modules

**Returns:**
```yaml
container: apps/backend
tech_stack:
  - layer: Runtime
    tech: Node.js
  - layer: Framework
    tech: Express

components:
  - path: src/auth
    name_hint: Auth Service
    type: business
    confidence: high
    imports_from: [src/db]

  - path: src/api
    name_hint: API Routes
    type: foundation
    confidence: high
```

### Component Discovery Subagent

**Input:** Confirmed component from Layer 2 + its path

**Prompt context:**
```
You are analyzing a component to understand its implementation for C3.

Determine:
1. Type: foundation | business | integration
2. Responsibility: what it does (one line)
3. Connections: what it talks to (other components, externals)
4. Patterns: notable implementation patterns

This informs the component's entry in the container inventory.
```

**Returns:**
```yaml
component: src/auth
type: business
responsibility: Handles user authentication and session management
connections:
  - target: src/db
    type: internal
    purpose: User credential storage
  - target: Redis
    type: external
    purpose: Session cache
patterns:
  - JWT tokens
  - Refresh token rotation
```

---

## User Interaction (AskUserQuestion)

Batch aggressively with multi-select. Minimize round-trips.

### Layer 1: Context Confirmation

```
Call 1:
  Q1: "What is this system?" (single-select or Other)
      - E-commerce platform
      - SaaS application
      - Internal tool

  Q2: "Which are separate containers?" (multi-select)
      - [x] apps/backend → Backend API
      - [x] apps/frontend → Frontend App
      - [ ] scripts/ → (likely not a container)
```

### Layer 2: Container Components (per container)

```
Call 2 (for Backend API):
  Q1: "Container purpose?" (prefilled, confirm or edit)
      - "Core API serving frontend and mobile clients"

  Q2: "Which are significant components?" (multi-select)
      - [x] src/auth → Auth Service
      - [x] src/api → API Routes
      - [x] src/db → Database Layer
      - [ ] src/utils → (utility, exclude)
```

### Layer 3: Component Roles

```
Call 3 (batched for confirmed components):
  Q1: "Component types correct?" (multi-select corrections)
      - [x] Auth Service: business ✓
      - [ ] API Routes: change to foundation

  Q2: "Anything missing from inventory?" (Other for additions)
```

---

## Mode-Specific Behavior

### Adopt Mode

1. Run all three discovery layers
2. ASK at each layer for confirmation
3. **Complete when:** All component inventories confirmed
4. **Output:** Create .c3/ docs from confirmed inventories

```
Reality (discovered) + User Confirmation → Create .c3/ docs
```

### Audit Mode

1. Read existing .c3/ docs (Expectation)
2. Run all three discovery layers (Reality)
3. Compare at each layer:
   - Context: containers in docs vs containers in code
   - Container: components in inventory vs components in code
   - Component: documented details vs actual implementation

```yaml
drift_report:
  context:
    missing_in_docs: []
    missing_in_code: []

  containers:
    c3-1-backend:
      missing_in_inventory:
        - src/cache  # new module, not documented
      missing_in_code:
        - c3-105-legacy  # documented but deleted

  components:
    c3-101-auth:
      type_mismatch: null
      responsibility_drift: "Now also handles OAuth, not just JWT"
```

---

## Decisions Summary

| Question | Decision |
|----------|----------|
| Subagent model | Sonnet |
| Discovery approach | Hierarchical (Context → Container → Component) |
| Each layer | Mandatory (component level is critical) |
| User input | AskUserQuestion, batched with multi-select |
| Fallback (nothing found) | Fall back to manual questions |
| Unified engine | Yes - same discovery for Adopt and Audit |

---

## Implementation Notes

### Parallelization Opportunities

- **Layer 1:** Single subagent (context discovery)
- **Layer 2:** N parallel subagents (one per confirmed container)
- **Layer 3:** M parallel subagents (one per confirmed component, batched)

### Subagent Prompt Template

Each subagent prompt includes:
1. C3 inventory model explanation
2. Layer-specific search targets
3. Structured output format (YAML)
4. Confidence indicators

### Error Handling

- **Confused signals:** ASK user for guidance
- **Nothing found:** Fall back to manual questions
- **Low confidence:** Include in options but mark as "likely exclude"

---

## Next Steps

1. Implement context discovery subagent
2. Implement container discovery subagent
3. Implement component discovery subagent
4. Wire up AskUserQuestion flow
5. Add audit mode diff engine
6. Test on real codebases
