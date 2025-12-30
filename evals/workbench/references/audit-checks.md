# Audit Checks Reference

Detailed validation rules for Mode: Audit in the c3 agent.

---

## Checks Summary

| Check | What It Validates | Pass/Fail Criteria |
|-------|-------------------|-------------------|
| **Frontmatter Validity** | Required fields per `v3-structure.md` | Missing id/title/type/parent = FAIL |
| **ID Pattern Compliance** | IDs follow `c3-{N}`, `c3-{N}{NN}` patterns | Wrong format = FAIL |
| **Inventory vs Code** | Significant code modules listed in inventory | Major module missing = FAIL |
| **Inventory-First Compliance** | No orphan component docs | Doc without inventory entry = FAIL |
| **Component Doc Completeness** | Existing docs have required sections | Missing required section = FAIL |
| **No Code Blocks** | Component docs use tables, not code | JSON/YAML/code snippets = FAIL |
| **Structure Integrity** | Parent exists before child | Container without Context = FAIL |
| **Diagram Accuracy** | Diagrams reference existing items | Diagram shows deleted container = FAIL |
| **ADR Lifecycle Integrity** | No orphan accepted ADRs | Accepted >30 days without implemented = WARN |

---

## Audit Procedure

### Phase 1: Gather

```
1. Read .c3/README.md (Context) - check frontmatter
2. List all .c3/c3-*/ directories (Containers)
3. For each Container:
   - Read README.md - check frontmatter, inventory table
   - List component docs (.c3/c3-*/c3-*.md)
4. List all ADRs (.c3/adr/adr-*.md) - check lifecycle status
```

### Phase 2: Validate Structure (per v3-structure.md)

```
For each doc:
  - Parse frontmatter: id, c3-version, title, type, parent, summary
  - Validate ID pattern: lowercase, correct format
  - Validate parent exists
  - Validate folder-ID match (c3-1 in c3-1-*/)
  - Validate slug format (lowercase-hyphenated)
```

**Required frontmatter by layer:**

| Layer | Required Fields |
|-------|-----------------|
| Context | id, c3-version, title, summary |
| Container | id, c3-version, title, type, parent, summary |
| Component | id, c3-version, title, type, parent, summary |
| ADR | id, type, status, title, affects |

### Phase 3: Cross-Reference Inventory

```
For each Container:
  - Parse Components table (inventory)
  - List component doc files that exist
  - Flag: doc exists but NOT in inventory → FAIL (inventory-first violation)
  - Note: inventory entry without doc → OK (inventory-first model)

For Context:
  - Parse Containers table
  - List container directories that exist
  - Flag: directory exists but NOT in table → FAIL (drift)
  - Flag: table entry but no directory → FAIL (drift)
```

### Phase 4: Validate Required Sections

**Context (c3-0) required sections:**

| Section | Purpose |
|---------|---------|
| Overview | What the system does |
| Containers | Table: ID, Name, Purpose |
| Container Interactions | Mermaid diagram |
| External Actors | Who/what interacts with system |

**Container (c3-N) required sections:**

| Section | Purpose |
|---------|---------|
| Technology Stack | Table: Layer, Tech, Purpose |
| Components | Table: ID, Name, Responsibility |
| Internal Structure | Mermaid diagram (optional but recommended) |

**Component (c3-NNN) required sections:**

| Section | Purpose |
|---------|---------|
| Contract | What this component provides |
| Interface | IN/OUT boundary diagram (Mermaid) |
| Hand-offs | Table: exchanges with other components |
| Conventions | Rules for consumers |
| Edge Cases | Error handling, failures |

**Component prohibited content:**
- No code blocks (except Mermaid)
- No JSON/YAML examples
- No interface definitions

### Phase 5: ADR Lifecycle Check

```
For each ADR:
  - Parse status from frontmatter
  - Parse Audit Record dates
  - If status=accepted AND no implemented date:
    - Calculate days since accepted
    - If >30 days → WARN (orphan ADR)
```

### Phase 6: Code Sampling (if significant codebase exists)

```
- Sample major directories (src/, lib/, packages/)
- Identify obvious modules not in any inventory
- Note: Sanity check, not exhaustive
```

---

## Audit Output Template

```
**C3 Audit Report**

**Scope:** [full / container:c3-1 / adr:adr-YYYYMMDD-slug]
**Date:** YYYY-MM-DD

## Summary
| Check | Status |
|-------|--------|
| Frontmatter Validity | ✓ PASS / ✗ FAIL |
| ID Pattern Compliance | ✓ PASS / ✗ FAIL |
| Inventory vs Code | ✓ PASS / ✗ FAIL |
| Inventory-First Compliance | ✓ PASS / ✗ FAIL |
| Component Doc Completeness | ✓ PASS / ✗ FAIL / ⚠ N/A (no docs) |
| No Code Blocks | ✓ PASS / ✗ FAIL / ⚠ N/A |
| Structure Integrity | ✓ PASS / ✗ FAIL |
| Diagram Accuracy | ✓ PASS / ✗ FAIL |
| ADR Lifecycle Integrity | ✓ PASS / ⚠ WARN |

## Issues Found

### Critical (Must Fix)
- [issue]: [details]

### Warnings (Should Fix)
- [issue]: [details]

### Info
- [observation]: [details]

## Recommendations
- [actionable fix]

## Next Steps
[Based on findings - see Drift Resolution below]
```

---

## Drift Resolution Guidance

When drift is detected, determine the cause:

| Situation | Cause | Action |
|-----------|-------|--------|
| Code changed, docs outdated | Intentional change not documented | Create ADR to formalize, then update docs |
| Docs describe removed code | Code deleted, forgot to update docs | Direct fix: remove stale doc sections |
| New module not in inventory | Recent addition | Direct fix: add to inventory |
| Component doc without inventory entry | Created doc before inventory | Direct fix: add to Container inventory first |
| Orphan ADR (accepted, never implemented) | Abandoned change | Close ADR with reason, or implement |

**Rule of thumb:**
- Drift from **intentional architectural change** → Create/update ADR
- Drift from **doc rot** (forgot to update) → Direct fix

---

## Audit Scope Options

| Scope | Command | Checks |
|-------|---------|--------|
| Full system | `audit C3` | All checks, all layers |
| Single container | `audit container c3-1` | Container + its components |
| ADR-specific | `audit adr adr-YYYYMMDD-slug` | ADR lifecycle + affected layers |

---

## Discovery-Based Audit (Alternative Approach)

The validation-based audit above checks **docs against rules**. Discovery-based audit checks **docs against reality** by running discovery first, then comparing.

### When to Use Discovery-Based Audit

| Situation | Use Discovery-Based |
|-----------|-------------------|
| Codebase changed significantly since docs written | Yes |
| Suspect major drift but don't know specifics | Yes |
| New team member needs to understand what's really there | Yes |
| Validating docs are accurate after refactoring | Yes |
| Regular health check for established systems | Yes |
| Just checking formatting and structure | No - use validation-based |

### Discovery-Based Audit Phases

#### Phase 0: Read Expectation

**Goal:** Parse .c3/ documentation to understand what SHOULD exist

```
1. Read .c3/README.md (Context)
   - Parse Containers table → expected containers
   - Parse External Actors → expected external systems
   - Parse Container Interactions diagram → expected relationships

2. For each Container (.c3/c3-*/README.md):
   - Parse Technology Stack table → expected tech layers
   - Parse Components table (inventory) → expected components
   - Parse Internal Structure diagram → expected component relationships

3. For each Component (.c3/c3-*/c3-*.md):
   - Parse Contract → expected responsibilities
   - Parse Hand-offs → expected component interactions
   - Parse Conventions → expected patterns
```

**Output:** Structured expectations data (containers, components, tech, relationships)

#### Phase 1: Discover Reality

**Goal:** Run discovery subagents to find what ACTUALLY exists in the codebase

**Discovery Strategy:**

```
1. Project Structure Discovery
   - Find all directories (categorize by depth, naming patterns)
   - Identify major modules (src/, lib/, packages/, apps/)
   - Detect build/config artifacts

2. Technology Stack Discovery
   - Parse package.json / pyproject.toml / Cargo.toml / go.mod
   - Identify frameworks (search for imports, decorators)
   - Find build tools (webpack, vite, etc.)

3. Component/Module Discovery
   - Search for major code patterns (classes, functions, modules)
   - Group by directory/namespace
   - Identify entry points (main.py, index.ts, etc.)

4. Interaction Discovery
   - Find HTTP clients (fetch, axios, requests)
   - Find HTTP servers (express, fastapi, gin)
   - Find message queue usage (kafka, rabbitmq)
   - Find database connections (sql, nosql)
   - Parse imports/dependencies between modules
```

**Output:** Structured reality data (actual modules, actual tech, actual connections)

#### Phase 2: Compute Drift

**Goal:** Compare expectations (Phase 0) vs reality (Phase 1) to find drift

**Drift Categories:**

| Drift Type | Meaning | Example |
|------------|---------|---------|
| **missing_in_docs** | Exists in code, not documented | Major module found but not in any inventory |
| **missing_in_code** | Documented but doesn't exist | Component doc describes deleted module |
| **mismatches** | Both exist but details differ | Docs say Express, code uses Fastify |

**Drift Computation:**

```
For Containers:
  expected_containers = parse Containers table from c3-0
  actual_directories = discover major directories
  → missing_in_docs = actual_directories - expected_containers
  → missing_in_code = expected_containers - actual_directories

For Components:
  expected_components = parse Components table from c3-N
  actual_modules = discover code modules in container path
  → missing_in_docs = actual_modules - expected_components
  → missing_in_code = expected_components - actual_modules

For Technology Stack:
  expected_tech = parse Technology Stack table from c3-N
  actual_tech = discover from package files + imports
  → mismatches = where expected != actual

For Interactions:
  expected_interactions = parse diagrams + hand-offs
  actual_interactions = discover client/server/db code
  → missing_in_docs = actual - expected
  → missing_in_code = expected - actual
```

**Output:** Drift report (categorized by type and severity)

#### Phase 3: Report Findings

**Goal:** Present drift findings with actionable recommendations

**Report Structure:**

```markdown
**Discovery-Based Audit Report**

**Scope:** [full / container:c3-N]
**Date:** YYYY-MM-DD

## Discovery Summary

**Expected (from .c3/ docs):**
- X containers
- Y components
- Z technology layers

**Discovered (from codebase):**
- X actual directories
- Y actual modules
- Z actual technologies

## Drift Analysis

### Missing in Docs (Exists in code, not documented)

**Containers:**
- `src/analytics/` - appears to be analytics service, not in Context

**Components:**
- `src/api/auth/session-manager.ts` - session handling module, not in c3-1 inventory

**Technology:**
- Redis (found in docker-compose.yml, not in tech stack)

### Missing in Code (Documented but doesn't exist)

**Components:**
- `c3-102-cache-layer.md` documents caching but no cache code found

**Technology:**
- Docs say PostgreSQL but no pg imports found

### Mismatches (Both exist but details differ)

**Technology:**
- Docs: "Express.js for HTTP server"
- Reality: Fastify (found in package.json + app.ts imports)

**Interactions:**
- Docs: "API calls Auth service via REST"
- Reality: No HTTP client code found, appears to use direct function calls

## Severity Assessment

| Drift Item | Severity | Why |
|------------|----------|-----|
| `src/analytics/` missing from docs | High | Major module completely undocumented |
| Redis not in tech stack | Medium | Infrastructure component missing |
| Express vs Fastify mismatch | Medium | Core tech difference |
| Stale cache component doc | Low | Doc for deleted feature |

## Recommendations

### Immediate Actions (High Severity)

1. **Document analytics container**
   - Create c3-3-analytics/ directory
   - Add to Context Containers table
   - Run container design for proper inventory

2. **Update tech stack**
   - Add Redis to c3-1 tech stack
   - Correct Express → Fastify in c3-1 tech stack

### Follow-Up Actions (Medium/Low Severity)

3. Remove c3-102-cache-layer.md (stale doc)
4. Verify PostgreSQL - may be in different service not yet documented
5. Update interaction docs to reflect direct function calls vs REST

## Next Steps

1. Create ADR for analytics container (architectural addition)
2. Direct fix: tech stack corrections in c3-1
3. Direct fix: remove stale component doc
```

### Discovery-Based vs Validation-Based

| Aspect | Validation-Based | Discovery-Based |
|--------|------------------|-----------------|
| **Starting point** | .c3/ documentation | Codebase reality |
| **Primary check** | Docs follow structure rules | Docs match code |
| **Best for** | Format compliance, structure integrity | Accuracy, completeness |
| **Finds** | Malformed docs, missing sections | Missing docs, outdated info |
| **When docs missing** | Fails validation | Shows what should be documented |
| **Speed** | Fast (just parse docs) | Slower (must discover code) |
| **Accuracy** | 100% for rules, doesn't check code | Depends on discovery quality |

**Recommendation:** Run both periodically. Validation-based for quick checks. Discovery-based for deep health checks.
