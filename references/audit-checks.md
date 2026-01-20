# Audit Checks Reference

Detailed validation rules for Mode: Audit in the c3 agent.

## Contents

- [Checks Summary](#checks-summary)
- [Audit Procedure](#audit-procedure)
  - Phase 1: Gather
  - Phase 2: Inventory vs Code
  - Phase 3: Component Categorization
  - Phase 4: References Validation
  - Phase 5: Diagram Accuracy
  - Phase 6: ADR Lifecycle
  - Phase 7: Reference Validation
- [Audit Output Template](#audit-output-template)
- [Drift Resolution Guidance](#drift-resolution-guidance)
- [Audit Scope Options](#audit-scope-options)
- [Discovery-Based Audit](#discovery-based-audit-alternative-approach)

---

## Checks Summary

| Check | What It Validates | Pass/Fail Criteria |
|-------|-------------------|-------------------|
| **Inventory vs Code** | Docs match reality | Module missing from inventory = FAIL |
| **Component Categorization** | Foundation/Feature | Wrong category = WARN |
| **Reference Validity** | References resolve to code | Missing symbol/path/glob = FAIL |
| **Reference Coverage** | Major code areas referenced | Major unreferenced area = WARN |
| **Diagram Accuracy** | Diagrams match inventory | Stale reference = FAIL |
| **ADR Lifecycle** | No stale ADRs | Accepted >30 days without implemented = WARN |
| **Ref Files** | ref-* files valid and linked | Missing Goal or orphan ref = WARN |

---

## Audit Procedure

### Phase 1: Gather

```
1. Read .c3/README.md (Context)
2. List .c3/c3-*/ directories (Containers)
3. For each Container: read README.md, list component docs
4. List ADRs (.c3/adr/adr-*.md)
```

### Phase 2: Inventory vs Code

```
For Context:
  - Compare Containers table ↔ actual directories
  - Flag drift in either direction

For each Container:
  - Compare Components inventory ↔ actual code modules
  - Flag: major module not in inventory → FAIL
```

### Phase 3: Component Categorization

```
For each Container:
  - Verify components are in Foundation/Feature sections
  - Apply categorization test:

    Foundation: "Would changing this break many others?"
    Feature: "Is this specific to what this product DOES?"

  - Flag: wrong category → WARN
```

| Category | Description | Examples |
|----------|-------------|----------|
| **Foundation** | Primitives, high impact | Layout, Button, Router |
| **Feature** | Domain-specific | ProductCard, CheckoutScreen |

### Phase 4: References Validation

```
For each Component:
  - Read `## References` section
  - For each reference:
    - Symbol: rg for definition, flag if not found
    - Pattern: glob, flag if zero matches
    - Path: check exists, flag if missing
  - Report: valid, stale (moved/renamed), or broken (missing)

Coverage:
  - Identify major code areas (top-level modules/packages)
  - Flag: major area with zero component references → WARN
```

### Phase 5: Diagram Accuracy

```
For each diagram:
  - Verify all IDs exist in inventory
  - Flag: stale reference → FAIL
```

### Phase 6: ADR Lifecycle

```
For each ADR with status=accepted:
  - If >30 days without implemented → WARN
```

### Phase 7: Reference Validation

```
For each ref-* file in .c3/refs/:
  - Verify Goal section exists
  - Verify ref is cited by at least one component doc (search for ref-* in c3-*/c3-*.md)

For each component doc citing a ref:
  - Verify cited ref-* file exists in .c3/refs/

Check for:
  - [ ] All ref-* files have Goal section
  - [ ] All cited refs (ref-* in component docs) exist in .c3/refs/
  - [ ] No orphan refs (refs not cited by any component)
  - [ ] Refs don't duplicate component content
```

---

## Audit Output Template

```
**C3 Audit Report**

**Scope:** [full / container:c3-N]
**Date:** YYYY-MM-DD

## Summary
| Check | Status |
|-------|--------|
| Inventory vs Code | ✓ PASS / ✗ FAIL |
| Component Categorization | ✓ PASS / ⚠ WARN |
| Reference Validity | ✓ PASS / ✗ FAIL |
| Reference Coverage | ✓ PASS / ⚠ WARN |
| Diagram Accuracy | ✓ PASS / ✗ FAIL |
| ADR Lifecycle | ✓ PASS / ⚠ WARN |
| Ref Files | ✓ PASS / ⚠ WARN |

## Issues
- [issue]: [details]

## Recommendations
- [actionable fix]
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
