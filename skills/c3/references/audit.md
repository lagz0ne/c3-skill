# Audit Reference

Validate C3 documentation for consistency, drift, and completeness.

## Audit Phases

Three-tier approach: **structural** (CLI-automated) -> **inventory** (CLI-assisted) -> **semantic** (manual reasoning).

## Progress Checklist

```
Audit Progress:
- [ ] Phase 0: Structural Validation (c3x check)
- [ ] Phase 1: Gather Inventory (c3x list --json)
- [ ] Phase 2: Inventory vs Code
- [ ] Phase 3: Component Categorization
- [ ] Phase 4: Code Reference Validation
- [ ] Phase 5: Diagram Accuracy
- [ ] Phase 6: ADR Lifecycle
- [ ] Phase 7: Ref File Validation
- [ ] Phase 8: Abstraction Boundaries
- [ ] Phase 9: Content Separation
- [ ] Phase 10: Context Files (CLAUDE.md)
```

---

## Phase 0: Structural Validation (CLI)

```bash
bash <skill-dir>/bin/c3x.sh check
bash <skill-dir>/bin/c3x.sh check --json   # machine-readable
```

Detects: broken links, orphan entities, duplicate IDs, missing parents.

Issues found here overlap with Phases 2, 4, 7 — skip re-checking those manually.

## Phase 1: Gather Inventory (CLI)

```bash
bash <skill-dir>/bin/c3x.sh list --json
bash <skill-dir>/bin/c3x.sh list            # quick text topology
```

Returns all entities with id, type, title, path, relationships, frontmatter.

Use this output as source of truth for all subsequent phases. No manual Glob+Read of `.c3/`.

## Phase 2: Inventory vs Code

```
For Context:
  - Compare Containers table <-> actual directories
  - Flag drift in either direction

For each Container:
  - Compare Components inventory <-> actual code modules
  - Flag: major module not in inventory -> FAIL
```

## Phase 3: Component Categorization

```
For each Container:
  - Verify components in Foundation/Feature sections
  - Foundation: "Would changing this break many others?"
  - Feature: "Is this specific to what this product DOES?"
  - Wrong category -> WARN
```

| Category | Description | Examples |
|----------|-------------|---------|
| **Foundation** (01-09) | Primitives, high impact | Layout, Button, Router |
| **Feature** (10+) | Domain-specific | ProductCard, CheckoutScreen |

## Phase 4: Code Reference Validation

```
For each Component:
  - Read ## Code References
  - For each reference:
    - Symbol: grep for definition, flag if not found
    - Pattern: glob, flag if zero matches
    - Path: check exists, flag if missing
  - Report: valid, stale (moved/renamed), broken (missing)

Coverage:
  - Identify major code areas (top-level modules)
  - Major area with zero component references -> WARN
```

## Phase 5: Diagram Accuracy

```
For each diagram:
  - Verify all IDs exist in inventory
  - Stale reference -> FAIL
```

## Phase 6: ADR Lifecycle

```
For each ADR with status=accepted:
  - If >30 days without implemented -> WARN
```

## Phase 7: Ref File Validation

```
For each ref-* file:
  - Verify required sections: Choice and Why (MUST exist)
  - Verify ref cited by at least one component
  - Flag orphan refs (not cited by any component)

For each component citing a ref:
  - Verify cited ref file exists in .c3/refs/
```

## Phase 8: Abstraction Boundaries

**Goal:** Detect when layers take on responsibilities of other layers.

| Signal | Check Method | Violation Type | Severity |
|--------|--------------|----------------|----------|
| Cross-container imports | Grep imports from other c3-* paths | Container bleeding | WARN |
| Global config definition | Grep exported constants used by 3+ files | Context bleeding | WARN |
| Multi-component orchestration | Check if orchestrating vs handing off | Container job | FAIL |
| Pattern redefinition | Compare to cited refs | Ref bypass | FAIL |

## Phase 9: Content Separation

**Goal:** Proper separation between Foundation (code), Feature (composition), Ref (guidance).

**The Code References Test:**
- Component WITH `## Code References` -> implemented (Foundation or Feature)
- Component WITHOUT `## Code References` -> either provisioned or misclassified (should be Ref)
- Ref WITH `## Code References` -> VIOLATION (should be Component)
- Ref with code examples in body -> VALID (golden references)

**Check for missing refs:**
1. Scan dependency manifests for technologies used in 3+ components
2. For each: does a ref exist explaining "how we use it HERE"?

**Check component content:**

| Signal | Indicates | Action |
|--------|-----------|--------|
| "We use X for..." | Technology usage pattern | Extract to ref |
| "Our convention is..." | Cross-cutting pattern | Extract to ref |
| Same pattern in 2+ components | Duplicated knowledge | Create ref |

## Phase 10: Context Files (CLAUDE.md)

**Goal:** Verify CLAUDE.md files propagated to code directories.

1. **Extract expected:** For each component with `## Code References`, parse directory paths
2. **Check presence:** CLAUDE.md exists in each expected directory?
3. **Check c3-generated block:** `<!-- c3-generated: c3-NNN -->` marker matches expected component
4. **Check orphans:** c3-generated blocks referencing deleted components

Expected block format:
```markdown
<!-- c3-generated: c3-201 -->
# c3-201: Component Title

Before modifying this code, read:
- Component: `.c3/c3-2-api/c3-201-component.md`
- Patterns: `ref-error-handling`, `ref-logging`

Full refs: `.c3/refs/ref-{name}.md`
<!-- end-c3-generated -->
```

---

## Output Format

```
**C3 Audit Results**

| Phase | Status | Issues |
|-------|--------|--------|
| Structural (CLI) | PASS/WARN/FAIL | [details] |
| Inventory vs Code | PASS/WARN/FAIL | [details] |
| ... | ... | ... |

**Summary:** N passes, M warnings, K failures
**Action Items:** [list of fixes needed]
```

---

## Drift Resolution

| Situation | Cause | Action |
|-----------|-------|--------|
| Code changed, docs outdated | Intentional change not documented | Create ADR, update docs |
| Docs describe removed code | Forgot to update | Direct fix: remove stale sections |
| New module not in inventory | Recent addition | Direct fix: add to inventory |
| Orphan ADR (accepted, never implemented) | Abandoned change | Close ADR with reason |

Rule of thumb:
- Drift from intentional architectural change -> Create/update ADR
- Drift from doc rot -> Direct fix

---

## Discovery-Based Audit (Alternative)

For deep health checks or after major codebase changes. Compares docs against code reality.

**When to use:** Codebase changed significantly, suspect major drift, after refactoring.

**Phases:**
1. **Read Expectation:** Parse `.c3/` docs via `c3x list --json` + selective Read
2. **Discover Reality:** Scan project structure, tech stack, modules, interactions
3. **Compute Drift:** Compare expectations vs reality -> missing_in_docs, missing_in_code, mismatches
4. **Report:** Severity assessment + actionable recommendations

---

## Audit Scope Options

| Scope | Focus | Checks |
|-------|-------|--------|
| Full system | All layers | All phases |
| Single container | Container + components | Phases 2-9 scoped |
| ADR-specific | ADR + affected layers | Phase 6 + affected entities |
