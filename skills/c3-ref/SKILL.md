---
name: c3-ref
description: |
  Manages cross-cutting patterns and conventions as first-class architecture artifacts.

  This skill should be used when the user asks to:
  - "add a pattern", "document this convention", "create a ref", "update ref-X"
  - "evolve this pattern", "what patterns exist", "which components use ref-X"
  - "list refs", "show refs", "list patterns"
  - "standardize this approach", "make this a convention", "enforce this across the codebase"

  <example>
  Context: Project with .c3/ directory
  user: "list all C3 refs and show which components cite each"
  assistant: "Using c3-ref to list patterns and their citings."
  </example>

  DO NOT use when "pattern" or "ref" is merely descriptive (e.g., "explain the auth flow pattern" → c3-query).
  DO NOT use for removing/deprecating refs (route to c3-change with ADR).
  Requires .c3/ to exist. Refs are authoritative constraints - violations require explicit override.
---

# C3 Ref - Pattern Management

Cross-cutting patterns (refs) are **authoritative constraints**. They define how things should be done system-wide. This skill makes refs first-class citizens with proper workflows.

## Precondition: C3 Adopted

**STOP if `.c3/README.md` does not exist.**

If missing:
> This project doesn't have C3 docs yet. Use the c3-onboard skill to create documentation first.

Do NOT proceed until `.c3/README.md` is confirmed.

## CRITICAL: Component Categorization

Load `references/component-categories.md` for the full Foundation vs Feature vs Ref rules.

**Key rule for refs:** Refs have NO `## Code References` section. If it needs one, it's a component.

## REQUIRED: Load References

Before proceeding, Read these files (relative to this skill's directory):
1. `references/skill-harness.md` - Red flags and complexity rules
2. `references/component-categories.md` - Foundation vs Feature vs Ref rules
3. `templates/ref.md` - Ref template structure

## Mode Selection

| User Intent | Mode |
|-------------|------|
| "add/create/document a pattern" | **Add** |
| "update/modify/evolve ref-X" | **Update** |
| "list patterns", "what refs exist" | **List** |
| "who uses ref-X", "where is ref-X cited" | **Usage** |
| "remove ref-X", "deprecate this pattern" | Route to **c3-change** (requires ADR) |

---

## Mode: Add

Create a new ref from discovered or proposed pattern.

### Flow

```
Write Ref FIRST → Discover Usage → Update Citings → Create ADR
```

**HARD RULE: Your FIRST Write call must be the ref file.** Do not Read any codebase files, do not Grep, do not look at existing refs before writing. Extract the pattern name, goal, and rules directly from the user's prompt. You can read the `templates/ref.md` template (from the references step above), then immediately write the ref file. Refine it AFTER it exists.

### Steps

**Step 1: Write Ref File IMMEDIATELY**

Extract from the user's prompt:
- Pattern name → slug: `ref-{slug}`
- Pattern goal → what it standardizes
- Key rules → what must be followed

Create `.c3/refs/ref-{slug}.md` using `templates/ref.md` template. Fill Choice and Why sections from what the user told you. Do NOT search the codebase first — the user's description is sufficient for the initial draft.

**Step 2: Discover Usage (BRIEF, max 2 Grep calls)**

Quick codebase scan to find components using this pattern. Do NOT exhaustively explore — just identify the main users for the citing step.

**Step 3: Refine Ref (if needed)**

If discovery reveals additional details (variations, anti-patterns), update the ref file with an Edit call.

**Step 4: Update Citing Components**

For each component that uses this pattern:

1. Read component doc
2. Add ref to `## Related Refs` table (create the table if it doesn't exist)
3. If component doc contains inline pattern content that duplicates the ref, note it for removal

Example `## Related Refs` table entry:

```markdown
## Related Refs

| Ref | Relationship |
|-----|-------------|
| ref-error-handling | Uses error response format |
| ref-retry-pattern | Implements retry with backoff |
```

**Scope:** Only modify `## Related Refs` table. If other content needs changing (e.g., removing duplicated pattern text), route to c3-change.

**Step 5: Create Adoption ADR**

Create mini-ADR at `.c3/adr/adr-YYYYMMDD-ref-{slug}-adoption.md`.

Note: Ref adoption ADRs use `status: implemented` directly because the ref doc IS the deliverable (no code changes to gate).

```markdown
---
id: adr-YYYYMMDD-ref-{slug}-adoption
title: Adopt {Pattern Title} as standard
status: implemented
---

# Adopt {Pattern Title} as Standard

## Problem

{Pattern was implemented inconsistently across N components}

## Decision

Document pattern as ref-{slug}. All existing usages now cite this ref.

## Affected Layers

| Layer | Change |
|-------|--------|
| refs | Added ref-{slug} |
| components | {list of updated components} |
```

---

## Mode: Update

Modify an existing ref with impact analysis.

### Flow

```
Identify Change → Find Citings → Check Compliance → Surface Violations → Execute
```

### Steps

**Step 1: Clarify Change**

Use `AskUserQuestion` to confirm the change type: "What change do you want to make to ref-{slug}?" with options like "Add a new rule", "Modify an existing rule", "Remove a rule", "Clarify/improve documentation".

**Step 2: Find All Citings**

Search all `.c3/` docs for the ref ID using the Grep tool:
- Pattern: `ref-{slug}` in path `.c3/` (recursive)

List all citing components.

**Step 3: Check Compliance**

For each citing component:
- Read code references from component doc
- Check if code still complies with proposed change
- Categorize: compliant / needs-update / breaking

**Step 4: Surface Impact**

Use `AskUserQuestion` to present the impact: "This change affects N components. M are already compliant, K need updates." with options like "Proceed - update ref and K components", "Narrow the change - only affect compliant ones", "Cancel - too much impact".

**Step 5: Execute**

If proceeding:

1. Update ref document (documentation only)
2. Create ADR for ref change
3. For non-compliant components: note as TODO in ADR (do NOT modify code)

```markdown
## Affected Components

| Component | Status | Action |
|-----------|--------|--------|
| c3-101 | compliant | None |
| c3-103 | needs-update | TODO: route to c3-change |
| c3-205 | breaking | TODO: route to c3-change |
```

**Step 6: Route to c3-change for Code Changes**

c3-ref updates ref documentation only. Any code changes in components MUST go through c3-change:

> "Pattern update requires code changes in {N} components. Route to c3-change skill to create an ADR for implementation."

Do not edit component code or component doc content directly from c3-ref (only `## Related Refs` tables may be updated during Add mode).

---

## Mode: List

Show all refs in the system.

### Flow

Use Glob to find all ref files: `.c3/refs/ref-*.md`

For each, read and extract:
- `id`
- `title`
- `goal`
- Count of citings

### Response Format

```
**C3 Patterns (Refs)**

| Ref | Title | Goal | Cited By |
|-----|-------|------|----------|
| ref-error-handling | Error Handling | Consistent error responses | 5 components |
| ref-auth | Authentication | Token-based auth | 3 components |
```

---

## Mode: Usage

Show where a specific ref is used.

### Flow

Search for citings using the Grep tool:
- Pattern: `ref-{slug}` in path `.c3/` with glob `c3-*/c3-*.md`

Read each citing component doc.

### Response Format

```
**ref-{slug} Usage**

**Cited by:**
- c3-101 (Auth Middleware) - JWT validation
- c3-103 (User Service) - Login flow
- c3-205 (API Gateway) - Request auth

**Pattern Summary:**
{Key rules from the ref}
```

---

## Anti-Patterns

| Anti-Pattern | Why It Fails | Correct Approach |
|--------------|--------------|------------------|
| Create ref without user input | Vague, unhelpful pattern doc | Extract specifics from user prompt before writing |
| Update ref without impact check | Break existing code silently | Always check citings |
| Duplicate ref content in components | Maintenance nightmare | Cite, don't duplicate |
| Create ref for one-off pattern | Unnecessary overhead | Refs are for repeated patterns |

