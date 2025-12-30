# Adopt Workflow

**Trigger:** No `.c3/` directory exists, or user asks to "set up C3", "create architecture docs"

**Purpose:** Discover and document existing architecture using inventory-first approach.

---

## Core Principle: Comprehension Before Progression

```
PHASE 1: Context (Firm the Container Inventory)
├── Discover all containers
├── Present overlay: declared vs discovered
├── User confirms container inventory
└── Document container connections
         ↓
    CONTAINER INVENTORY LOCKED
         ↓
PHASE 2: Containers (Build Component Inventories)
├── For each confirmed container:
│   ├── Discover all components
│   ├── Present overlay: declared vs discovered
│   └── User confirms component inventory
└── Create .c3/ structure
```

**Rule:** Never descend to Phase 2 until Phase 1 is FIRM.

---

## Phase 1: Context (Container Inventory)

### Step 1.1: Read Existing Inventory

If `.c3/README.md` exists, parse container inventory table:

```yaml
declared_containers:
  - id: c3-1
    name: API Backend
    responsibility: Request handling
  - id: c3-2
    name: Frontend
    responsibility: User interface
```

If no `.c3/` exists: `declared_containers: []`

### Step 1.2: Discover Containers

**Dispatch Context Discovery subagent** (see `discovery-engine.md`):
- Scans entire codebase for container boundaries
- Returns discovered containers with paths and confidence

```yaml
discovered_containers:
  - path: services/api
    name_hint: API Service
    confidence: high
  - path: frontend
    name_hint: Web Frontend
    confidence: high
  - path: services/worker
    name_hint: Background Worker
    confidence: medium
```

### Step 1.3: Present Overlay

**Show the complete picture before any decisions:**

```
CONTAINER INVENTORY OVERLAY
═══════════════════════════════════════════════════════════════

Declared (in .c3/)          Discovered (in code)        Status
───────────────────────────────────────────────────────────────
c3-1 API Backend            services/api                ✓ Match
c3-2 Frontend               frontend                    ✓ Match
-                           services/worker             ✚ NEW
c3-3 Legacy Adapter         -                           ⚠ MISSING

External Systems Detected:
- PostgreSQL (database) - evidence: docker-compose.yml, services/api/config/db.ts
- Redis (cache) - evidence: services/api/src/cache.ts
- Stripe API (external) - evidence: services/api/src/payments/stripe.ts

═══════════════════════════════════════════════════════════════
```

### Step 1.4: User Confirms Container Inventory

**Ask user to confirm:**
- Accept matches as-is?
- Add NEW discoveries to inventory?
- Handle MISSING (remove from inventory or mark for investigation)?
- Rename/merge/split any containers?
- Validate external systems?

**Output:** Confirmed container list with final names and IDs

### Step 1.5: Document Container Connections

**After containers are firm, capture connections:**
- How do containers communicate? (HTTP, events, shared DB)
- What protocols? (REST, GraphQL, gRPC)
- What are the dependencies? (A calls B, C publishes to D)

This becomes the Context-level interaction diagram.

### Phase 1 Gate

```
┌─────────────────────────────────────────────────────────┐
│ CHECKPOINT: Is Container Inventory FIRM?                │
│                                                         │
│ [ ] All containers identified and named                 │
│ [ ] All external systems documented                     │
│ [ ] Container connections understood                    │
│ [ ] User has explicitly confirmed                       │
│                                                         │
│ If NO to any: Do NOT proceed to Phase 2                 │
└─────────────────────────────────────────────────────────┘
```

---

## Phase 2: Containers (Component Inventories)

**Only enter Phase 2 after Phase 1 gate passes.**

### Step 2.1: For Each Confirmed Container

Process containers sequentially (or in parallel if independent):

#### 2.1.1: Read Existing Component Inventory

If container doc exists (`.c3/c3-{N}-{slug}/README.md`), parse component inventory:

```yaml
declared_components:
  - id: c3-101
    name: Request Handler
    type: Foundation
    status: Documented
  - id: c3-102
    name: Auth Service
    type: Business
    status: ""
```

If container is new: `declared_components: []`

#### 2.1.2: Discover Components

**Dispatch Container Discovery subagent** with container path:
- Scans container scope for component boundaries
- Returns discovered components with types and confidence

```yaml
discovered_components:
  - path: src/handlers
    name_hint: HTTP Handlers
    type: foundation
    confidence: high
  - path: src/services/auth
    name_hint: Auth Service
    type: business
    confidence: high
  - path: src/services/orders
    name_hint: Order Service
    type: business
    confidence: high
```

#### 2.1.3: Present Component Overlay

**Show overlay for this container:**

```
COMPONENT INVENTORY: c3-1 API Backend
═══════════════════════════════════════════════════════════════

Declared (in .c3/)          Discovered (in code)        Status
───────────────────────────────────────────────────────────────
c3-101 Request Handler      src/handlers                ✓ Match
c3-102 Auth Service         src/services/auth           ✓ Match
-                           src/services/orders         ✚ NEW
c3-103 Legacy Auth          -                           ⚠ MISSING

═══════════════════════════════════════════════════════════════
```

#### 2.1.4: User Confirms Component Inventory

**Ask user to confirm for this container:**
- Accept matches?
- Add NEW discoveries?
- Handle MISSING?
- Assign types (Foundation/Business)?

**Output:** Confirmed component list with final names, IDs, types

### Step 2.2: Repeat for All Containers

Continue until all containers have confirmed component inventories.

---

## Phase 3: Create .c3/ Structure

**Only after both phases complete, scaffold documentation:**

### ID Assignment

- Context: `c3-0`
- Containers: `c3-1`, `c3-2`, etc. (sequential)
- Components: `c3-101`, `c3-102` for c3-1; `c3-201`, `c3-202` for c3-2

### Created Structure

```
.c3/
├── README.md (c3-0)
│   └── Contains: Overview, Container Inventory, Interactions Diagram, Externals
│
├── c3-1-{slug}/
│   └── README.md
│       └── Contains: Tech Stack, Component Inventory, Internal Structure Diagram
│
├── c3-2-{slug}/
│   └── README.md
│       └── Contains: Tech Stack, Component Inventory, Internal Structure Diagram
│
├── adr/             ← Empty, for future ADRs
└── settings.yaml    ← Optional custom settings
```

### NOT Created

- **Component docs** - Inventory-first model means docs appear when conventions mature
- **Stub files** - Either a full doc or nothing

---

## Overlay Status Semantics

| Status | Meaning | Action |
|--------|---------|--------|
| `✓ Match` | Declared item found in code | Keep, optionally update path |
| `✚ NEW` | Found in code, not declared | Add to inventory or ignore |
| `⚠ MISSING` | Declared but not found in code | Remove, rename, or investigate drift |

---

## Fallback Behavior

**If discovery subagent fails:**

1. Retry once with same parameters
2. If second failure: Fall back to manual entry
3. Log which steps used manual fallback

**Partial success is acceptable** - proceed with gathered data.

---

## Final Report

```
C3 STRUCTURE CREATED
════════════════════

Context (c3-0):
  Containers: 3
  External Systems: 3

Container Inventories:
  c3-1-api-backend: 5 components
  c3-2-frontend: 4 components
  c3-3-worker: 2 components

Next Steps:
  - Create ADRs for architectural decisions
  - Document components when conventions emerge
  - Run audit periodically to check for drift
```

---

## Anti-Patterns

| Anti-Pattern | Why It's Wrong | Correct Approach |
|--------------|----------------|------------------|
| Jumping to components before confirming containers | Builds on unstable foundation | Complete Phase 1 gate first |
| Skipping overlay presentation | User can't see gaps/drift | Always show declared vs discovered |
| Creating component docs during adopt | Premature - conventions not yet known | Inventory only, docs come later |
| Ignoring MISSING items | Hides drift | Force decision: remove or investigate |
