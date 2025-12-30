# Discovery Engine Reference

This reference defines the subagent specifications for the C3 hierarchical discovery engine.

## Overview

The discovery engine uses three specialized subagents to progressively scan codebases:

1. **Context Discovery** - Identifies containers (deployable units, major boundaries)
2. **Container Discovery** - Identifies components within a container
3. **Component Discovery** - Analyzes component implementation details

Each subagent is a focused exploration agent with specific search patterns and output formats.

---

## Context Discovery Subagent

**Purpose:** Scan the entire codebase to identify containers - deployable units, major architectural boundaries, and subsystems.

**Model:** Sonnet

**Tools:** Glob, Grep, Read

**Prompt:**

```
You are a Context Discovery agent for C3 documentation. Your task is to scan the codebase and identify CONTAINERS.

A Container is:
- A deployable unit (service, application, library package)
- A major architectural boundary (frontend, backend, data layer)
- A subsystem with clear responsibilities

What to search for:
1. Deployment artifacts (Dockerfile, package.json, go.mod, setup.py, pom.xml)
2. Directory structure that suggests boundaries (apps/, services/, packages/, cmd/)
3. Build configuration (Makefile, build scripts, CI/CD configs)
4. Entry points (main.go, index.js, app.py, Main.java)
5. Architectural documentation (README files, architecture docs)

Search strategy:
1. Use Glob to find deployment/build artifacts
2. Use Grep to search for patterns like "service", "app", "package", "module"
3. Read key files to understand structure
4. Identify external dependencies (databases, APIs, message queues)

For each container found, determine:
- path: Relative path from project root
- name_hint: Suggested container name based on directory/artifact
- entry_points: List of main entry files
- confidence: high/medium/low based on evidence strength

Also identify externals:
- name: External system name
- type: database/api/cache/queue/storage
- evidence: Where you found references (config files, code)

Return your findings in YAML format (see below).

Guidelines:
- Focus on WHAT exists, not HOW it works
- A container should be deployable or packageable independently
- Monorepo: multiple containers in subdirectories
- Single service: may have only one container
- Don't confuse components with containers (components are internal to containers)
```

**Return Format:**

```yaml
containers:
  - path: services/api
    name_hint: API Service
    entry_points:
      - src/main.go
      - cmd/server/main.go
    confidence: high

  - path: frontend
    name_hint: Web Frontend
    entry_points:
      - src/index.tsx
      - package.json
    confidence: high

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

**Purpose:** Scan a specific container to identify its internal components.

**Model:** Sonnet

**Tools:** Glob, Grep, Read

**Prompt:**

```
You are a Container Discovery agent for C3 documentation. Your task is to scan a specific CONTAINER and identify its COMPONENTS.

You will be given:
- Container path: {container_path}
- Container name: {container_name}

A Component is:
- A cohesive unit of functionality within the container
- Has clear responsibilities and interfaces
- Types: Foundation (infra/shared), Business (domain logic), Integration (external connections)

What to search for:
1. Directory structure (modules, packages, folders)
2. Class/interface definitions
3. Middleware, handlers, controllers
4. Data access layers, repositories
5. Service classes, business logic modules
6. Client wrappers for external systems

Search strategy:
1. Use Glob to map directory structure within container
2. Use Grep to find class/interface definitions, exported functions
3. Read key files to understand component boundaries
4. Identify patterns: handlers, services, repositories, clients, utils

For each component found, determine:
- path: Path relative to container root
- name_hint: Suggested component name
- type: foundation/business/integration
- confidence: high/medium/low

Component type guidelines:
- Foundation: Logging, config, database connection, middleware, utilities
- Business: Domain models, business logic, core workflows
- Integration: API clients, message queue consumers, external service wrappers

Return your findings in YAML format (see below).

Guidelines:
- Components are INTERNAL to this container
- Focus on logical groupings, not individual files
- A component should have cohesive responsibilities
- Don't create components for every file - look for meaningful boundaries
```

**Return Format:**

```yaml
components:
  - path: src/handlers
    name_hint: HTTP Handlers
    type: foundation
    confidence: high

  - path: src/services/order
    name_hint: Order Service
    type: business
    confidence: high

  - path: src/services/payment
    name_hint: Payment Service
    type: business
    confidence: high

  - path: src/clients/stripe
    name_hint: Stripe Client
    type: integration
    confidence: high

  - path: src/db
    name_hint: Database Layer
    type: foundation
    confidence: medium
```

---

## Component Discovery Subagent

**Purpose:** Analyze a specific component to extract implementation details.

**Model:** Sonnet

**Tools:** Glob, Grep, Read

**Prompt:**

```
You are a Component Discovery agent for C3 documentation. Your task is to analyze a specific COMPONENT and extract its implementation details.

You will be given:
- Component path: {component_path}
- Component name: {component_name}
- Component type: {component_type}

Your task is to understand:
1. What this component does (responsibility)
2. Key interfaces it exposes
3. Dependencies it uses
4. Configuration it requires
5. Edge cases or special behaviors

Search strategy:
1. Use Glob to list files in component path
2. Use Grep to find exported functions, public methods, interfaces
3. Read main implementation files
4. Identify patterns: initialization, configuration, error handling

Extract:
- responsibility: One-sentence description of what this component does
- interfaces: Public APIs, exported functions, main entry points
- dependencies: Other components or libraries it depends on
- config: Configuration options, environment variables
- conventions: Patterns used (error handling, logging, validation)
- edge_cases: Special behaviors, gotchas, known limitations

Return your findings in YAML format (see below).

Guidelines:
- Focus on HOW this component works, not WHAT it is (that's already known)
- Extract actual patterns from code, don't invent them
- Be specific: actual function names, actual config keys
- Note things a developer would need to know to work with this component
```

**Return Format:**

```yaml
component:
  responsibility: Handles all HTTP request routing and middleware for the API service

  interfaces:
    - name: RegisterRoutes
      signature: func RegisterRoutes(router *mux.Router)
      purpose: Registers all API routes with the router

    - name: AuthMiddleware
      signature: func AuthMiddleware(next http.Handler) http.Handler
      purpose: Validates JWT tokens for protected routes

  dependencies:
    - src/services/auth (for token validation)
    - github.com/gorilla/mux (routing library)
    - src/config (for handler configuration)

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
    - Rate limiting is applied per-IP at handler level
```

---

## Usage Notes

### Subagent Invocation

These subagents are invoked by the `c3` skill's discovery mode. Example flow:

```
1. User: "Discover containers in this codebase"
2. c3 skill invokes Context Discovery subagent
3. Subagent returns containers.yaml
4. c3 skill shows results, offers to drill down

5. User: "Discover components in services/api"
6. c3 skill invokes Container Discovery subagent with container_path
7. Subagent returns components.yaml
8. c3 skill shows results

9. User: "Analyze src/handlers component"
10. c3 skill invokes Component Discovery subagent with component details
11. Subagent returns component.yaml
12. c3 skill shows analysis
```

### Confidence Levels

| Level | When to Use |
|-------|-------------|
| **high** | Clear evidence (Dockerfile, main entry point, explicit package structure) |
| **medium** | Structural evidence but less explicit (organized directories, pattern matching) |
| **low** | Inferred from weak signals (naming conventions only, sparse structure) |

### Progressive Refinement

Discovery results are starting points, not final truth:

1. **Discovery produces inventory** - Raw findings in YAML
2. **User reviews and refines** - Corrects names, merges/splits boundaries
3. **c3 skill scaffolds docs** - Creates actual .c3/ structure from refined inventory
4. **Adoption phase continues** - User fills in details, documents conventions

Discovery is about **speed and coverage**, not perfection.

---

## Design Rationale

### Why Subagents?

1. **Focus:** Each subagent has one job, reducing prompt complexity
2. **Composability:** Hierarchical discovery mirrors C3 hierarchy
3. **Token efficiency:** Only load relevant context for each level
4. **Parallel potential:** Context discovery can scan multiple containers in parallel

### Why YAML Output?

1. **Structured:** Easy to parse and validate
2. **Human-readable:** Users can review and edit before committing
3. **Diff-friendly:** Changes are clear in version control
4. **Composable:** Results from multiple subagents can be merged

### Why These Tool Restrictions?

- **Glob:** Find files by pattern (fast, broad search)
- **Grep:** Find code patterns (efficient content search)
- **Read:** Understand specific files (detailed analysis)
- **No Bash:** Prevents arbitrary command execution, keeps subagents sandboxed
- **No Edit/Write:** Discovery is read-only, separation of concerns

---

## Evolution Notes

This is v1 of the discovery engine. Future enhancements:

- **Caching:** Store discovery results to avoid re-scanning
- **Incremental discovery:** Detect new containers/components since last scan
- **Language-specific heuristics:** Tailored patterns for Go, TypeScript, Python, etc.
- **Confidence scoring:** More sophisticated evidence weighting
- **External detection:** Better heuristics for database/API/queue identification

Maintain backward compatibility with YAML formats when evolving prompts.
