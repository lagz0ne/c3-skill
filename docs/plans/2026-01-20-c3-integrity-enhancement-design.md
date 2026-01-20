# C3 Integrity Enhancement Design

**Date:** 2026-01-20
**Status:** Proposed

## Problem Statement

The C3 skills express the Context → Container → Component hierarchy but lack:
1. **Ref authority** - Refs are descriptive, not normative; violations don't block
2. **Constraint visibility** - Users can't see the full constraint chain for a layer
3. **Ref workflows** - No dedicated journey for creating/modifying refs
4. **Abstraction validation** - Nothing detects when components take on container responsibilities

## Design Goals

- Refs become **blocking constraints** - violations require explicit acknowledgment
- Users can **visualize constraint chains** for any layer
- Refs are **first-class citizens** with their own workflows
- Audit **detects abstraction violations**

---

## Change 1: Make Refs Blocking

### Rationale
Top-down integrity requires that pattern violations are not silently bypassed. Changes that break established patterns must be explicit in the ADR.

### Files to Modify

**`agents/c3-orchestrator.md`**

In Phase 4 (Socratic Refinement), add blocking logic:

```markdown
## Phase 4b: Pattern Violation Gate

If `c3-patterns` returned `breaks`:

1. **Surface violation clearly:**
   ```
   AskUserQuestion:
     question: "This change breaks established pattern ref-{name}. How do you want to proceed?"
     options:
       - "Update the pattern (broader scope)"
       - "Override pattern (requires justification in ADR)"
       - "Rethink the approach"
   ```

2. **If "Override pattern":**
   - ADR MUST include "Pattern Override" section
   - User provides explicit justification
   - ADR records: "Overrides ref-{name} because: {reason}"

3. **If "Update the pattern":**
   - Expand scope to include ref modification
   - Re-run analysis with ref changes
```

**`references/adr-template.md`**

Add optional section:

```markdown
## Pattern Overrides (if applicable)

| Ref | Override Reason | Impact |
|-----|-----------------|--------|
| ref-{name} | {explicit justification} | {what usages are affected} |
```

---

## Change 2: Add `/c3-ref` Skill

### Rationale
Refs deserve their own workflows - creating patterns and evolving them shouldn't be buried in component workflows.

### Files to Create

**`skills/c3-ref/SKILL.md`**

```yaml
---
name: c3-ref
description: |
  Manage cross-cutting patterns and conventions.
  Use when: "add a pattern", "document this convention", "update ref-X", "evolve this pattern"
  Routes: c3-ref add → create ref, c3-ref update → modify ref with validation
---

# C3 Ref - Pattern Management

## Mode: Add

Create a new ref from discovered pattern.

### Flow
1. User describes pattern
2. Search codebase for usages (Grep for pattern indicators)
3. Generate ref from `templates/ref.md`
4. Update citing components' "Related Refs" sections
5. Create mini-ADR documenting pattern adoption

### Example
```
User: "We use retry with exponential backoff everywhere"

1. Search: `rg "retry|backoff|exponential"`
2. Found in: c3-101, c3-103, c3-205
3. Create: .c3/refs/ref-retry-pattern.md
4. Update 3 component docs with Related Refs
5. Create: adr-YYYYMMDD-retry-pattern-adoption.md
```

## Mode: Update

Modify existing ref with impact analysis.

### Flow
1. User describes change to ref
2. Find all citing components
3. For each: check if still compliant
4. Surface non-compliant components
5. User decides: update components or narrow ref change
6. Execute updates

### Example
```
User: "Change retry pattern to use jitter"

1. Read ref-retry-pattern.md
2. Find citings: c3-101, c3-103, c3-205
3. Check each for jitter usage
4. c3-205 doesn't use jitter → surface
5. User: "Update c3-205 too"
6. Create ADR, update ref, update c3-205
```
```

---

## Change 3: Constraint Chain Query

### Rationale
Users need to see the full constraint picture before modifying a layer.

### Files to Modify

**`skills/c3-query/SKILL.md`**

Add new query type:

```markdown
## Query Types

| Type | User Says | Response |
|------|-----------|----------|
| Constraints | "what rules apply to X", "constraints for X" | Full constraint chain |
...

## Step: Constraint Chain (for constraint queries)

When user asks about rules/constraints for a component:

1. **Read component doc** → Extract `Related Refs` section
2. **Read container doc** → Extract any "Components MUST" rules
3. **Read context doc** → Extract any system-wide constraints
4. **Synthesize chain:**

Response format:
```
**Constraint Chain for c3-NNN**

**Context Constraints:**
- [rules from c3-0 that apply]

**Container Constraints:**
- [rules from c3-N that apply]

**Pattern Constraints (from cited refs):**
- ref-X: [key rules]
- ref-Y: [key rules]

**Visualization:**
[mermaid diagram showing inheritance]
```
```

---

## Change 4: Abstraction Validation in Audit

### Rationale
Integrity erodes when components accumulate responsibilities that belong to higher layers.

### Files to Modify

**`references/audit-checks.md`**

Add Phase 8:

```markdown
### Phase 8: Abstraction Boundaries

For each Component:
  - Read component doc
  - Check for abstraction violations:

| Signal | Violation Type | Severity |
|--------|---------------|----------|
| Imports from other containers | Container bleeding | WARN |
| Defines global config/constants used elsewhere | Context bleeding | WARN |
| Orchestrates multiple other components | Container job | WARN |
| Duplicates ref content instead of citing | Ref bypass | WARN |

For each Container:
  - Check for context bleeding:

| Signal | Violation Type | Severity |
|--------|---------------|----------|
| Defines system-wide policies | Context job | WARN |
| Contains components that should be separate container | Boundary violation | WARN |

**Output:**
```
## Abstraction Boundary Check

| Layer | Violation | Recommendation |
|-------|-----------|----------------|
| c3-105 | Imports c3-2-* (other container) | Move to container linkage or shared ref |
| c3-103 | Orchestrates c3-104, c3-105 | Consider elevating to container coordination |
```
```

---

## Change 5: Template Constraints

### Rationale
Templates should explicitly state layer boundaries to guide documentation.

### Files to Modify

**`templates/component.md`**

Add section:

```markdown
## Layer Constraints

This component operates within these boundaries:

**MUST:**
- Focus on single responsibility within its domain
- Cite refs for patterns instead of re-implementing
- Hand off cross-component concerns to container

**MUST NOT:**
- Import directly from other containers (use container linkages)
- Define system-wide configuration (context responsibility)
- Orchestrate multiple peer components (container responsibility)
- Redefine patterns that exist in refs
```

**`templates/container.md`**

Add section:

```markdown
## Layer Constraints

This container operates within these boundaries:

**MUST:**
- Coordinate components within its boundary
- Define how context linkages are fulfilled internally
- Own its technology stack decisions

**MUST NOT:**
- Define system-wide policies (context responsibility)
- Implement business logic directly (component responsibility)
- Bypass refs for cross-cutting concerns
```

---

## Implementation Order

| Priority | Change | Effort | Impact |
|----------|--------|--------|--------|
| 1 | Make Refs Blocking | Medium | High - enforces integrity |
| 2 | Constraint Chain Query | Low | Medium - visibility |
| 3 | Abstraction Validation | Medium | Medium - catches drift |
| 4 | Add `/c3-ref` Skill | High | High - first-class refs |
| 5 | Template Constraints | Low | Low - documentation |

**Recommended approach:** Start with #1 and #2 (high impact, lower effort), then #3 and #5, finally #4 (most effort but builds on others).

---

## Verification

After implementation:

1. **Refs blocking:** Attempt a change that violates a ref → must require explicit override
2. **Constraint query:** Ask "what constraints apply to c3-105" → must show full chain
3. **Abstraction audit:** Run audit on project with intentional violation → must flag
4. **Ref workflow:** Use `/c3-ref add` to document a pattern → must create ref and update citings
