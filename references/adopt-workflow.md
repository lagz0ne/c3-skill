# Adopt Workflow

**Trigger:** No `.c3/` directory exists, or user asks to "set up C3", "create architecture docs"

**Purpose:** Discover and document existing architecture using automated codebase scanning.

## Discovery-Based Approach

The adopt workflow uses a **discovery engine** with specialized subagents to scan the codebase and propose documentation structure. User confirms at each layer before proceeding.

**Reference:** `references/discovery-engine.md` for detailed subagent specifications.

## 7-Step Discovery Flow

### 1. Context Discovery (Find Containers)

**Agent dispatches Context Discovery subagent:**
- Scans codebase for container boundaries
- Identifies entry points, technologies, external systems
- Returns container candidates with confidence scores

**Output:** YAML with containers and externals

### 2. Confirm Containers with User

**Agent presents findings:**
- List of discovered containers (name, path, entry points, confidence)
- Detected external systems (name, type, evidence)

**User confirms:**
- Which containers to document
- Renames/merges/splits as needed
- Validates external systems

### 3. Container Discovery (Find Components per Container)

**For each confirmed container, agent dispatches Container Discovery subagent:**
- Scans container scope for component boundaries
- Identifies component types (foundation/business/integration)
- Returns component candidates with confidence scores

**Output:** YAML with components per container

### 4. Confirm Components per Container

**Agent presents findings per container:**
- List of discovered components (name, path, type, confidence)

**User confirms:**
- Which components to include in inventory
- Renames/merges/splits as needed
- Validates component types

### 5. Component Discovery (Optional Detail)

**ONLY if user explicitly requests detailed analysis.**

**Agent dispatches Component Discovery subagent for select components:**
- Analyzes responsibility, interfaces, dependencies, config
- Returns detailed component characteristics

**Default:** Skip this step. Component inventory is sufficient at adopt time.

### 6. Confirm Component Details (If Gathered)

**Agent presents analysis for reviewed components:**
- Responsibility, interfaces, dependencies, config, conventions

**User confirms accuracy.**

**Note:** Details are for understanding only - no component docs are created (inventory-first model).

### 7. Create .c3/ Structure

**Agent scaffolds documentation using confirmed inventories:**

**ID Assignment:**
- Context: `c3-0`
- Containers: `c3-1`, `c3-2`, `c3-3`, etc. (sequential)
- Components: `c3-101`, `c3-102` for c3-1; `c3-201`, `c3-202` for c3-2, etc.

**Created:**
```
.c3/
├── README.md (c3-0) ← Context with container inventory
├── c3-1-{slug}/
│   └── README.md    ← Container with component inventory
├── c3-2-{slug}/
│   └── README.md    ← Container with component inventory
├── adr/             ← Empty, for future ADRs
└── settings.yaml    ← Optional custom settings
```

**NOT Created:**
- Component docs (inventory-first model)

**Component docs appear later when:**
- Conventions emerge that consumers must follow
- Hand-off patterns become non-obvious
- Edge cases need documentation

## Inventory-First Model

**CRITICAL:** Adopt creates **inventory tables**, NOT component docs.

```markdown
## Components (in Container README)

| ID | Name | Type | Responsibility |
|----|------|------|----------------|
| c3-101 | Request Handler | Foundation | HTTP routing |
| c3-102 | Auth Service | Business | Authentication |
| c3-103 | User Service | Business | User management |
```

**Rationale:** Most components don't need docs at adoption time. Inventory enables navigation and impact analysis. Docs come later when implementation details matter.

## Fallback Behavior

**If discovery subagent fails or times out:**

1. **Retry once** with same parameters
2. **If second failure:**
   - Fall back to manual entry mode
   - Ask user to provide information directly
3. **Log failure** in adoption report (note which steps used manual fallback)

**Partial success is acceptable** - proceed with whatever data was successfully gathered.

## Atomic Completion

**NEVER end before completing all confirmed containers.**

Adopt creates the full structure in one pass. Individual component docs come later through normal development flow (via `c3-implementation` skill).

## Final Report

```
**C3 structure created**

Containers documented:
- c3-1-{name} ({N} components in inventory)
- c3-2-{name} ({N} components in inventory)

Externals documented:
- {external_name} ({type})

Next steps:
- Use `c3-structure` skill for structural changes (containers, context)
- Use `c3-implementation` skill when documenting components
- Create ADRs for architectural decisions
```
