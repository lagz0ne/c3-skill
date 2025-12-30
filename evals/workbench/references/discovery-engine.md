# Discovery Engine Reference

Subagent specifications for C3 inventory-first discovery.

---

## Overview

The discovery engine uses specialized subagents that:
1. **Read existing inventory** (declared items from `.c3/`)
2. **Discover from code** (scan codebase for actual items)
3. **Return overlay data** (declared + discovered for comparison)

```
┌─────────────────────────────────────────────────────────┐
│                    DISCOVERY FLOW                       │
├─────────────────────────────────────────────────────────┤
│                                                         │
│   Read .c3/     →    Scan Code    →    Return Overlay   │
│   (declared)         (discovered)      (both + status)  │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## Context Discovery Subagent

**Purpose:** Identify containers and produce overlay against existing container inventory.

**Model:** Sonnet

**Tools:** Glob, Grep, Read

### Prompt

```
You are a Context Discovery agent. Your task is to:
1. Read existing container inventory (if .c3/ exists)
2. Scan codebase for container boundaries
3. Return overlay comparing declared vs discovered

STEP 1: READ EXISTING INVENTORY
-------------------------------
Check if .c3/README.md exists. If yes, parse the Containers table:

| ID | Name | Responsibility |
|----|------|----------------|

Extract into:
declared_containers:
  - id: c3-1
    name: API Backend
    responsibility: Request handling

If no .c3/ exists: declared_containers: []

STEP 2: DISCOVER CONTAINERS
---------------------------
Scan codebase for containers. A Container is:
- A deployable unit (service, application, library package)
- A major architectural boundary (frontend, backend, data layer)
- A subsystem with clear responsibilities

Search for:
1. Deployment artifacts (Dockerfile, package.json, go.mod, setup.py, pom.xml)
2. Directory structure suggesting boundaries (apps/, services/, packages/, cmd/)
3. Build configuration (Makefile, build scripts, CI/CD configs)
4. Entry points (main.go, index.js, app.py, Main.java)

For each container found:
- path: Relative path from project root
- name_hint: Suggested name based on directory/artifact
- confidence: high/medium/low

Also detect external systems:
- name: External system name
- type: database/api/cache/queue/storage
- evidence: Config files, code references

STEP 3: RETURN OVERLAY
----------------------
Return YAML with both declared and discovered, enabling overlay comparison.

Guidelines:
- Focus on WHAT exists, not HOW it works
- A container should be deployable or packageable independently
- Don't confuse components with containers
- Monorepo: multiple containers in subdirectories
- Single service: may have only one container
```

### Return Format

```yaml
declared_containers:
  - id: c3-1
    name: API Backend
    responsibility: Request handling, business logic
  - id: c3-2
    name: Frontend
    responsibility: User interface

discovered_containers:
  - path: services/api
    name_hint: API Service
    entry_points:
      - src/main.go
      - cmd/server/main.go
    confidence: high
    matches_declared: c3-1  # null if no match

  - path: frontend
    name_hint: Web Frontend
    entry_points:
      - src/index.tsx
    confidence: high
    matches_declared: c3-2

  - path: services/worker
    name_hint: Background Worker
    entry_points:
      - cmd/worker/main.go
    confidence: medium
    matches_declared: null  # NEW - not in declared

externals:
  - name: PostgreSQL
    type: database
    evidence: services/api/config/database.yml, docker-compose.yml

  - name: Redis
    type: cache
    evidence: services/api/src/cache/redis.go

  - name: Stripe API
    type: api
    evidence: services/api/src/payments/stripe.go
```

---

## Container Discovery Subagent

**Purpose:** Identify components within a container and produce overlay against existing component inventory.

**Model:** Sonnet

**Tools:** Glob, Grep, Read

### Input Parameters

```yaml
container_path: services/api
container_name: API Backend
container_id: c3-1
```

### Prompt

```
You are a Container Discovery agent. Your task is to:
1. Read existing component inventory for this container
2. Scan container scope for component boundaries
3. Return overlay comparing declared vs discovered

Container: {container_name}
Path: {container_path}
ID: {container_id}

STEP 1: READ EXISTING INVENTORY
-------------------------------
Check if .c3/{container_id}-*/README.md exists. If yes, parse Components table:

| ID | Name | Type | Responsibility | Status |
|----|------|------|----------------|--------|

Extract into:
declared_components:
  - id: c3-101
    name: Request Handler
    type: Foundation
    status: Documented
  - id: c3-102
    name: Auth Service
    type: Business
    status: ""

If container is new: declared_components: []

STEP 2: DISCOVER COMPONENTS
---------------------------
Scan container scope for components. A Component is:
- A cohesive unit of functionality within the container
- Has clear responsibilities and interfaces

Search for:
1. Directory structure (modules, packages, folders)
2. Class/interface definitions
3. Handlers, controllers, middleware
4. Services, repositories, data access
5. Client wrappers for external systems

Component types:
- Foundation: Cross-cutting (HTTP framework, logger, config, middleware)
- Business: Domain logic (auth service, order processor)

For each component found:
- path: Path relative to container root
- name_hint: Suggested component name
- type: foundation/business
- confidence: high/medium/low

STEP 3: RETURN OVERLAY
----------------------
Return YAML with both declared and discovered, enabling overlay comparison.

Guidelines:
- Components are INTERNAL to this container
- Focus on logical groupings, not individual files
- A component should have cohesive responsibilities
- Don't create components for every file
```

### Return Format

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
  - id: c3-103
    name: Legacy Auth
    type: Business
    status: Skip: deprecated

discovered_components:
  - path: src/handlers
    name_hint: HTTP Handlers
    type: foundation
    confidence: high
    matches_declared: c3-101

  - path: src/services/auth
    name_hint: Auth Service
    type: business
    confidence: high
    matches_declared: c3-102

  - path: src/services/orders
    name_hint: Order Service
    type: business
    confidence: high
    matches_declared: null  # NEW

  # Note: c3-103 Legacy Auth has no discovered match → MISSING
```

---

## Component Discovery Subagent (Optional)

**Purpose:** Deep-dive analysis of a specific component. **Only used when explicitly requested.**

**Model:** Sonnet

**Tools:** Glob, Grep, Read

### Input Parameters

```yaml
component_path: src/handlers
component_name: HTTP Handlers
component_type: foundation
container_path: services/api
```

### Prompt

```
You are a Component Discovery agent. Your task is to analyze a specific component for implementation details.

Component: {component_name}
Path: {container_path}/{component_path}
Type: {component_type}

This is OPTIONAL deep-dive analysis. Most components don't need this at adopt time.

Analyze:
1. Responsibility: One-sentence description
2. Interfaces: Public APIs, exported functions, entry points
3. Dependencies: Other components or libraries used
4. Config: Configuration options, environment variables
5. Conventions: Patterns used (error handling, logging, validation)
6. Edge cases: Special behaviors, gotchas, limitations

Guidelines:
- Focus on HOW this component works
- Extract actual patterns from code
- Note things a developer would need to know
- Be specific: actual function names, actual config keys
```

### Return Format

```yaml
component:
  responsibility: Handles all HTTP request routing and middleware for the API

  interfaces:
    - name: RegisterRoutes
      signature: func RegisterRoutes(router *mux.Router)
      purpose: Registers all API routes

    - name: AuthMiddleware
      signature: func AuthMiddleware(next http.Handler) http.Handler
      purpose: Validates JWT tokens for protected routes

  dependencies:
    - src/services/auth (token validation)
    - github.com/gorilla/mux (routing)
    - src/config (handler configuration)

  config:
    - API_TIMEOUT: Request timeout in seconds
    - MAX_BODY_SIZE: Maximum request body size
    - CORS_ORIGINS: Allowed CORS origins

  conventions:
    - All handlers return JSON responses
    - Errors are logged with request ID
    - Panics are recovered and converted to 500 responses

  edge_cases:
    - WebSocket routes bypass normal middleware
    - File uploads use streaming to avoid memory issues
```

---

## Overlay Computation

The orchestrating agent (not the subagents) computes the overlay:

```python
def compute_overlay(declared, discovered):
    overlay = []

    # Match discovered to declared
    for disc in discovered:
        if disc.matches_declared:
            overlay.append({
                "declared": find_by_id(declared, disc.matches_declared),
                "discovered": disc,
                "status": "MATCH"
            })
        else:
            overlay.append({
                "declared": None,
                "discovered": disc,
                "status": "NEW"
            })

    # Find declared with no discovered match
    matched_ids = [d.matches_declared for d in discovered if d.matches_declared]
    for decl in declared:
        if decl.id not in matched_ids:
            overlay.append({
                "declared": decl,
                "discovered": None,
                "status": "MISSING"
            })

    return overlay
```

---

## Confidence Levels

| Level | Evidence | Example |
|-------|----------|---------|
| **high** | Explicit artifact | Dockerfile, package.json with clear name |
| **medium** | Structural evidence | Organized directory, pattern matching |
| **low** | Weak signals | Naming conventions only, sparse structure |

---

## Design Rationale

### Why Read Before Discover?

1. **Drift detection:** Comparing declared vs discovered reveals drift
2. **User context:** Seeing existing inventory helps user understand proposals
3. **ID preservation:** Matching discovered to declared maintains ID stability

### Why Overlay Format?

1. **Single source:** All data in one response
2. **Comparison-ready:** Both sides present for overlay rendering
3. **Decision-enabling:** User can see exactly what needs action

### Why YAML Output?

1. **Structured:** Easy to parse and validate
2. **Human-readable:** Users can review before committing
3. **Diff-friendly:** Changes are clear in version control

### Tool Restrictions

- **Glob:** Find files by pattern (fast, broad search)
- **Grep:** Find code patterns (efficient content search)
- **Read:** Understand specific files (detailed analysis)
- **No Bash:** Prevents arbitrary command execution
- **No Edit/Write:** Discovery is read-only
