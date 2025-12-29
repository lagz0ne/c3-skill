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
