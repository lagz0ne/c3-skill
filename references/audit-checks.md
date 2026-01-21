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
  - Phase 8: Abstraction Boundaries
  - Phase 9: Content Separation
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
| **Abstraction Boundaries** | Layers stay in their lane | Container job/Ref bypass = FAIL, bleeding = WARN |
| **Content Separation** | Components = domain logic, Refs = usage patterns | Integration content in component = WARN, missing ref for technology = WARN |

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

### Phase 8: Abstraction Boundaries

**Goal:** Detect when layers take on responsibilities that belong to other layers.

```
For each Component:
  1. Read component doc
  2. Check code references for abstraction violations:

  | Signal | Check Method | Violation Type |
  |--------|--------------|----------------|
  | Cross-container imports | Grep for imports from other c3-* paths | Container bleeding |
  | Global config definition | Grep for exported constants used by 3+ files | Context bleeding |
  | Multi-component orchestration | Read hand-offs, check if orchestrating vs handing off | Container job |
  | Pattern redefinition | Compare to cited refs, check for reimplementation | Ref bypass |

For each Container:
  1. Read container doc
  2. Check for context bleeding:

  | Signal | Check Method | Violation Type |
  |--------|--------------|----------------|
  | System-wide policy definition | Check if container defines rules used by other containers | Context job |
  | Cross-container coordination | Check if container orchestrates other containers | Context job |
```

**Severity levels:**

| Violation | Severity | Reason |
|-----------|----------|--------|
| Container bleeding (component imports other containers) | WARN | May be valid cross-container dependency, review needed |
| Context bleeding (component defines global config) | WARN | May indicate missing context documentation |
| Container job (component orchestrates peers) | FAIL | Clear abstraction violation |
| Ref bypass (component redefines pattern) | FAIL | Integrity violation |

**Output format:**

```
## Abstraction Boundary Check

| Layer | Violation | Evidence | Recommendation |
|-------|-----------|----------|----------------|
| c3-105 | Container bleeding | Imports `c3-2-api/auth` | Use container linkage or shared ref |
| c3-103 | Container job | Orchestrates c3-104, c3-105 | Elevate to container coordination |
| c3-201 | Ref bypass | Reimplements error handling | Cite ref-error-handling instead |
```

### Phase 9: Content Separation

**Goal:** Ensure proper separation between domain logic (components) and usage patterns (refs).

**Core Principle:** C3 proactively splits content:
- **Components** document WHAT the system does (business/domain logic)
- **Refs** document HOW we use technologies HERE (usage patterns, conventions)

**The Separation Test:** (see `content-separation.md` for full definition)

> "Would this content change if we swapped the underlying technology?"
> - **Yes** → Integration/usage pattern → belongs in ref
> - **No** → Business/domain logic → belongs in component

#### Step 1: Identify Technologies in Use

```
1. Scan dependency manifests (package.json, go.mod, Cargo.toml, etc.)
2. Scan imports across codebase for frameworks/libraries
3. List technologies used in 3+ components
```

#### Step 2: Check for Missing Refs

For each significant technology/framework:
- Does a ref exist explaining "how we use it HERE"?
- Does the ref capture specific decisions and conventions (not generic docs)?

**What refs should capture (the "use" questions):**

| Question | What It Captures |
|----------|------------------|
| **When** do we use this? | Context, triggers, conditions |
| **Why** this over alternatives? | Decision rationale |
| **Where** is the entry point? | How to invoke, where to start |
| **What** conventions apply? | Constraints, patterns we follow |

#### Step 3: Check Component Content

For each component doc, analyze content for integration patterns that should be refs:

| Signal | Indicates | Action |
|--------|-----------|--------|
| "We use X for..." | Technology usage pattern | Extract to ref |
| "Our convention is..." | Cross-cutting pattern | Extract to ref |
| Setup/config details | Integration knowledge | Extract to ref |
| Same pattern in 2+ components | Duplicated knowledge | Create ref, cite in both |

**NOT ref content (keep in component):**
- Domain-specific rules ("Users are charged when...")
- Business calculations
- Feature-specific behavior
- Entity lifecycle logic

#### Step 4: Check Ref Content

For each ref doc, verify it doesn't contain business logic:

| Signal | Indicates | Action |
|--------|-----------|--------|
| Domain entity behavior | Business logic | Move to component |
| Business rules not tied to technology | Domain logic | Move to component |
| Feature-specific workflows | Domain logic | Move to component |

#### Step 5: Check Ref Scope

Refs should be focused enough to be actionable:
- If a ref covers too much (e.g., "ref-react" for a large React app) → split by concern
- Scoping options: by layer, by feature area, by specific pattern

**Output format:**

```
## Content Separation Check

### Missing Refs

| Technology | Used In | Recommendation |
|------------|---------|----------------|
| [technology] | c3-101, c3-205, c3-301 | Create ref-[topic] |

### Misplaced Content

| Location | Content Type | Should Be | Evidence | Severity |
|----------|--------------|-----------|----------|----------|
| c3-201 Conventions | Integration pattern | ref-error-handling | "We use RFC 7807 format..." | WARN |
| ref-api-patterns | Business rule | c3-pricing | Rate limit tiers by plan | WARN |

### Duplicated Patterns

| Pattern | Found In | Recommendation |
|---------|----------|----------------|
| Retry logic | c3-201, c3-205 | Create ref-retry-pattern |
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
| Abstraction Boundaries | ✓ PASS / ⚠ WARN / ✗ FAIL |
| Content Separation | ✓ PASS / ⚠ WARN |

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
