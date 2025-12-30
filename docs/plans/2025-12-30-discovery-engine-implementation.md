# C3 Discovery Engine Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a hierarchical discovery engine using subagents to power both Adopt and Audit modes.

**Architecture:** Three-layer discovery (Context → Container → Component) using Sonnet subagents, with AskUserQuestion for user confirmation at each layer. Unified engine serves both Adopt (create docs) and Audit (detect drift).

**Tech Stack:** Claude Code agents, YAML structured output, AskUserQuestion tool

**Design Doc:** `docs/plans/2025-12-29-subagent-exploration-design.md`

---

## Phase 1: Foundation

### Task 1: Create Discovery Reference File

**Files:**
- Create: `references/discovery-engine.md`

**Step 1: Create the reference file with subagent prompts**

```markdown
# Discovery Engine Reference

Subagent specifications for hierarchical C3 discovery.

---

## Context Discovery Subagent

**Model:** Sonnet
**Tools:** Glob, Grep, Read

**Prompt:**

You are scanning a codebase to identify containers for C3 documentation.

A "container" in C3 is a deployable unit or major boundary:
- Separate apps (frontend, backend, mobile)
- Services in a microservices architecture
- Major packages in a monorepo

**Your task:**
1. Scan directory structure (depth 2-3)
2. Identify entry points (main files, Dockerfiles, package.json scripts)
3. Detect package manager patterns (workspaces, multi-package)
4. Find infrastructure signals (docker-compose, k8s)
5. Identify external dependencies (databases, APIs)

**Return YAML:**

```yaml
containers:
  - path: <relative path>
    name_hint: <suggested name>
    entry_points: [<files>]
    confidence: high | medium | low

externals:
  - name: <service name>
    type: database | api | cache | queue
    evidence: [<how detected>]

connections:
  - from: <container>
    to: <container>
    type: http | grpc | events
    evidence: [<how detected>]
```

---

## Container Discovery Subagent

**Model:** Sonnet
**Tools:** Glob, Grep, Read
**Input:** Container path from Layer 1

**Prompt:**

You are scanning a container to identify components for C3 documentation.

A "component" in C3 is a significant internal unit:
- Foundation: shared infrastructure (DB layer, HTTP server, auth middleware)
- Business: domain logic (user service, order processing)
- Integration: external adapters (API clients, message handlers)

**Scan:** {container_path}

**Your task:**
1. Analyze internal directory structure
2. Identify module boundaries
3. Detect import relationships
4. Classify component types

**Return YAML:**

```yaml
container: <path>
tech_stack:
  - layer: Runtime | Framework | Database | Cache
    tech: <technology>
    purpose: <why used>

components:
  - path: <relative path within container>
    name_hint: <suggested name>
    type: foundation | business | integration
    confidence: high | medium | low
    imports_from: [<other components>]
```

---

## Component Discovery Subagent

**Model:** Sonnet
**Tools:** Glob, Grep, Read
**Input:** Component path from Layer 2

**Prompt:**

You are analyzing a component to understand its implementation for C3.

**Analyze:** {component_path}

**Determine:**
1. Type: foundation | business | integration
2. Responsibility: what it does (one sentence)
3. Connections: what it talks to
4. Key patterns: notable implementation approaches

**Return YAML:**

```yaml
component: <path>
type: foundation | business | integration
responsibility: <one sentence>
connections:
  - target: <component or external>
    type: internal | external
    purpose: <why connected>
patterns: [<notable patterns>]
```
```

**Step 2: Commit**

```bash
git add references/discovery-engine.md
git commit -m "docs(references): add discovery engine subagent specifications"
```

---

### Task 2: Update Agent Tools

**Files:**
- Modify: `agents/c3.md`

**Step 1: Add Task tool to agent**

Current tools line:
```
tools: Glob, Grep, Read, Edit, Write, TodoWrite, Skill
```

Change to:
```
tools: Glob, Grep, Read, Edit, Write, TodoWrite, Skill, Task, AskUserQuestion
```

**Step 2: Commit**

```bash
git add agents/c3.md
git commit -m "feat(agents): add Task and AskUserQuestion tools to c3 agent"
```

---

## Phase 2: Context Discovery (Layer 1)

### Task 3: Add Adopt Mode with Context Discovery

**Files:**
- Modify: `agents/c3.md`

**Step 1: Replace Mode: Adopt section**

Find the section starting with `## Mode: Adopt (New Project)` and replace with:

```markdown
## Mode: Adopt (New Project)

**Trigger:** No `.c3/` directory exists, or user asks to "set up C3", "create architecture docs"

### Step 1: Context Discovery

Dispatch context discovery subagent:

```
Task tool:
  subagent_type: general-purpose
  model: sonnet
  prompt: |
    [Load references/discovery-engine.md - Context Discovery Subagent section]

    Scan this codebase and return container candidates.
    Return YAML only, no explanation.
```

### Step 2: Confirm Containers

Use AskUserQuestion to confirm discovered containers:

```
Q1: "What type of system is this?" (single)
    Options from: [Web application, API service, CLI tool, Library, Monorepo]

Q2: "Which are separate containers?" (multi-select)
    Options from discovery results, include confidence
```

### Step 3: Container Discovery (per confirmed container)

For each confirmed container, dispatch container discovery subagent:

```
Task tool (parallel, one per container):
  subagent_type: general-purpose
  model: sonnet
  prompt: |
    [Load references/discovery-engine.md - Container Discovery Subagent section]

    Container path: {container_path}
    Scan and return component candidates.
```

### Step 4: Confirm Components

Use AskUserQuestion per container:

```
Q1: "Purpose of {container_name}?" (single, prefilled from discovery)

Q2: "Which are significant components?" (multi-select)
    Options from discovery results
```

### Step 5: Component Discovery

For each confirmed component, dispatch component discovery subagent:

```
Task tool (parallel, batched):
  subagent_type: general-purpose
  model: sonnet
  prompt: |
    [Load references/discovery-engine.md - Component Discovery Subagent section]

    Component path: {component_path}
    Analyze and return details.
```

### Step 6: Confirm Component Details

Use AskUserQuestion:

```
Q1: "Component types correct?" (multi-select for corrections)

Q2: "Any components missing?" (Other for additions)
```

### Step 7: Create .c3/ Structure

After all confirmations, create docs using inventory-first:

1. Create `.c3/README.md` (Context) with container inventory
2. Create `.c3/c3-N-{slug}/README.md` (Container) with component inventory
3. Do NOT create component docs yet (inventory-first)

Use `c3-structure` skill patterns for document structure.
```

**Step 2: Commit**

```bash
git add agents/c3.md
git commit -m "feat(agents): add discovery-based adopt mode"
```

---

## Phase 3: Audit Mode

### Task 4: Add Audit Mode with Drift Detection

**Files:**
- Modify: `agents/c3.md`

**Step 1: Update Mode: Audit section**

Find `## Mode: Audit (Standalone Health Check)` and replace with:

```markdown
## Mode: Audit (Standalone Health Check)

**Trigger:** User says "audit C3", "check docs", "verify architecture"

**Purpose:** Compare code reality vs documented expectation.

### Step 1: Read Expectation

Read existing .c3/ docs:
- `.c3/README.md` → extract container inventory
- `.c3/c3-*/README.md` → extract component inventories

### Step 2: Discover Reality

Run same discovery as Adopt mode:
1. Context discovery → container candidates
2. Container discovery → component candidates
3. Component discovery → details

### Step 3: Compare

For each layer, compute diff:

```yaml
drift_report:
  context:
    missing_in_docs: [containers in code but not documented]
    missing_in_code: [containers documented but not found]

  containers:
    {container_id}:
      missing_in_inventory: [components in code but not listed]
      missing_in_code: [components listed but not found]

  components:
    {component_id}:
      type_mismatch: [if type changed]
      responsibility_drift: [if purpose changed]
```

### Step 4: Report

Present drift report using structure from `references/audit-checks.md`.

For each drift:
- Explain what changed
- Recommend fix (update docs or remove stale entries)
```

**Step 2: Commit**

```bash
git add agents/c3.md
git commit -m "feat(agents): add discovery-based audit mode with drift detection"
```

---

## Phase 4: Integration

### Task 5: Update Adopt Workflow Reference

**Files:**
- Modify: `references/adopt-workflow.md`

**Step 1: Update to reference discovery engine**

Replace content to align with new discovery-based flow:

```markdown
# Adopt Workflow

**Trigger:** No `.c3/` directory exists, or user asks to "set up C3"

## Discovery-Based Adoption

Adopt uses the hierarchical discovery engine:

1. **Context Discovery** → Identify containers
2. **User Confirmation** → Confirm containers via AskUserQuestion
3. **Container Discovery** → Identify components per container
4. **User Confirmation** → Confirm components via AskUserQuestion
5. **Component Discovery** → Analyze component details
6. **User Confirmation** → Confirm types and responsibilities
7. **Create Docs** → Write .c3/ structure

See `references/discovery-engine.md` for subagent specifications.

## Inventory-First

At adopt completion:
- Context doc has container inventory table
- Container docs have component inventory tables
- NO component docs created (inventory-first model)

## Fallback

If discovery finds nothing useful:
- Fall back to manual questions
- Ask user to describe system structure
- Build inventory from responses
```

**Step 2: Commit**

```bash
git add references/adopt-workflow.md
git commit -m "docs(references): update adopt workflow for discovery engine"
```

---

### Task 6: Update Audit Checks Reference

**Files:**
- Modify: `references/audit-checks.md`

**Step 1: Add discovery-based audit section**

Add after the existing content:

```markdown
---

## Discovery-Based Audit

The audit now uses the discovery engine to find reality:

### Phase 0: Read Expectation

```
1. Parse .c3/README.md → container inventory
2. For each container:
   - Parse .c3/c3-*/README.md → component inventory
3. Build expectation model
```

### Phase 1: Discover Reality

```
1. Run context discovery → container candidates
2. Run container discovery → component candidates
3. Run component discovery → details
4. Build reality model
```

### Phase 2: Compute Drift

```
For each layer:
  - missing_in_docs = reality - expectation
  - missing_in_code = expectation - reality
  - mismatches = different details for same item
```

### Phase 3: Report

Present findings using audit output template.
```

**Step 2: Commit**

```bash
git add references/audit-checks.md
git commit -m "docs(references): add discovery-based audit procedure"
```

---

## Phase 5: Testing

### Task 7: Create Test Case for Discovery Adopt

**Files:**
- Create: `tests/cases/06-discovery-adopt.md`

**Step 1: Create test case**

```markdown
# Test: Discovery-Based Adopt

## Setup

Fixture: `fixtures/06-discovery-adopt/`

```
apps/
  backend/
    src/
      main.ts
      auth/
        index.ts
      api/
        routes.ts
    package.json
    Dockerfile
  frontend/
    src/
      index.tsx
      components/
    package.json
docker-compose.yml
```

## Query

```
Set up C3 documentation for this project.
```

## Expect

### PASS: Discovery Should Find

| Element | Check |
|---------|-------|
| 2 containers detected | backend, frontend |
| Components in backend | auth, api |
| External actor | detected from docker-compose |

### PASS: User Interaction

| Element | Check |
|---------|-------|
| AskUserQuestion called | For container confirmation |
| AskUserQuestion called | For component confirmation |

### PASS: Output Structure

| Element | Check |
|---------|-------|
| .c3/README.md | Has container inventory |
| .c3/c3-1-backend/README.md | Has component inventory |
| .c3/c3-2-frontend/README.md | Has component inventory |
| No component docs | Inventory-first respected |

### FAIL: Should NOT Include

| Element | Failure Reason |
|---------|----------------|
| Component .md files | Should be inventory-first |
| Code blocks in docs | NO CODE rule |
```

**Step 2: Commit**

```bash
git add tests/cases/06-discovery-adopt.md
git commit -m "test: add discovery-based adopt test case"
```

---

### Task 8: Create Test Case for Discovery Audit

**Files:**
- Create: `tests/cases/07-discovery-audit.md`

**Step 1: Create test case**

```markdown
# Test: Discovery-Based Audit

## Setup

Fixture: `fixtures/07-discovery-audit/`

Pre-existing .c3/ docs with deliberate drift:
- c3-1-backend lists component c3-103-legacy (deleted in code)
- Code has new src/cache module (not in inventory)

## Query

```
Audit the C3 documentation.
```

## Expect

### PASS: Drift Detection

| Element | Check |
|---------|-------|
| missing_in_inventory | src/cache detected |
| missing_in_code | c3-103-legacy detected |

### PASS: Report Format

| Element | Check |
|---------|-------|
| Drift report generated | Following audit template |
| Recommendations included | How to fix each drift |
```

**Step 2: Commit**

```bash
git add tests/cases/07-discovery-audit.md
git commit -m "test: add discovery-based audit test case"
```

---

## Phase 6: Version Bump

### Task 9: Bump Plugin Version

**Files:**
- Modify: `.claude-plugin/plugin.json`
- Modify: `.claude-plugin/marketplace.json`

**Step 1: Update version to 1.13.0 (minor bump for new feature)**

In both files, change:
```json
"version": "1.12.2"
```
to:
```json
"version": "1.13.0"
```

**Step 2: Commit**

```bash
git add .claude-plugin/plugin.json .claude-plugin/marketplace.json
git commit -m "chore: bump version to 1.13.0 for discovery engine"
```

---

## Summary

| Phase | Tasks | What's Built |
|-------|-------|--------------|
| 1. Foundation | 1-2 | Reference file, agent tools |
| 2. Context Discovery | 3 | Adopt mode with discovery |
| 3. Audit Mode | 4 | Audit mode with drift detection |
| 4. Integration | 5-6 | Updated references |
| 5. Testing | 7-8 | Test cases |
| 6. Version | 9 | Version bump |

**Total:** 9 tasks, ~45 minutes estimated
