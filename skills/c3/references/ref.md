# Ref Reference

Manage scoped patterns and conventions as first-class architecture artifacts.

## Component Categories Reminder

- **Foundation/Feature** (components): Has entry in `.c3/code-map.yaml` — code counterpart exists
- **Ref** (pattern): NO code-map entry — no code counterpart. May include golden code examples.

Hard rule: If you cannot name a concrete file, create a ref, not a component.

## Mode Selection

| User Intent | Mode |
|-------------|------|
| "add/create/document a pattern" | **Add** |
| "update/modify/evolve ref-X" | **Update** |
| "list patterns", "what refs exist" | **List** |
| "who uses ref-X", "where is ref-X cited" | **Usage** |
| "remove/deprecate ref-X" | Route to **change** (requires ADR) |

---

## Mode: Add

Create a new ref from discovered or proposed pattern.

### Flow

```
Scaffold via CLI -> Fill Content -> Discover Usage -> Update Citings -> Create ADR
```

**HARD RULE: Your FIRST Bash call must be the scaffold.** Do not Read codebase, Grep, or look at existing refs before writing. Extract pattern name and slug from user's prompt.

### Steps

**Step 1: Scaffold**
```bash
bash <skill-dir>/bin/c3x.sh add ref <slug>
```
Creates `.c3/refs/ref-{slug}.md` with correct frontmatter and structure.

**Step 2: Fill Content via Edit**

From user's prompt:
- `## Goal` — what it standardizes
- `## Choice` — what option was chosen (REQUIRED)
- `## Why` — rationale over alternatives (REQUIRED)
- Other sections as relevant (How, Scope, Not This, Override)

Do NOT search codebase first — user's description is sufficient for initial draft.

**Step 3: Discover Usage (brief, 2-3 Grep calls)**

Quick codebase scan to find components using this pattern. Not exhaustive — just identify main users.

**Step 4: Refine Ref (if needed)**

If discovery reveals additional details (variations, anti-patterns), update ref file.

**Step 5: Update Citing Components**

For each component using this pattern:
1. Read component doc
2. Add ref to `## Related Refs` table (create table if missing)

Example entry:
```markdown
## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-error-handling | Uses error response format |
```

**Scope:** Only modify `## Related Refs` table. Other content changes -> route to change.

**Step 6: Create Adoption ADR**

`.c3/adr/adr-YYYYMMDD-ref-{slug}-adoption.md`

```yaml
---
id: adr-YYYYMMDD-ref-{slug}-adoption
title: Adopt {Pattern Title} as standard
status: implemented
---
```

Ref adoption ADRs use `status: implemented` directly — the ref doc IS the deliverable.

---

## Mode: Update

Modify existing ref with impact analysis.

### Flow

```
Clarify Change -> Find Citings -> Check Compliance -> Surface Impact -> Execute
```

**Step 1: Clarify Change**
`AskUserQuestion`: "What change to ref-{slug}?" (skip if ASSUMPTION_MODE)
- Add a new rule
- Modify an existing rule
- Remove a rule
- Clarify/improve documentation

**Step 2: Find All Citings**
```bash
bash <skill-dir>/bin/c3x.sh list --json
```
Find ref entity by id. `relationships` field lists citing components. For deeper search: Grep `ref-{slug}` in `.c3/`.

**Step 3: Check Compliance**
For each citing component:
- Read code-map entries
- Check if code complies with proposed change
- Categorize: compliant / needs-update / breaking

**Step 4: Surface Impact**
`AskUserQuestion`: "This change affects N components. M compliant, K need updates." (skip if ASSUMPTION_MODE)
- Proceed — update ref and K components
- Narrow — only affect compliant ones
- Cancel — too much impact

**Step 5: Execute**
1. Update ref document (documentation only)
2. Create ADR for ref change
3. Non-compliant components: note as TODO in ADR (do NOT modify code)

**Step 6: Route to change for code changes**
Ref updates documentation only. Code changes in components MUST go through change operation.

---

## Mode: List

```bash
bash <skill-dir>/bin/c3x.sh list --json
```

Filter by `type: "ref"`. Extract: id, title, frontmatter.goal, relationships (citing components).

**Response:**
```
**C3 Patterns (Refs)**

| Ref | Title | Goal |
|-----|-------|------|
| ref-error-handling | Error Handling | Consistent error responses |
```

---

## Mode: Usage

```bash
bash <skill-dir>/bin/c3x.sh list --json
```

Find entry with `id: "ref-{slug}"`, read `relationships` for citing components. Read each citing doc for details.

**Response:**
```
**ref-{slug} Usage**

**Cited by:**
- c3-101 (Auth Middleware) - JWT validation
- c3-103 (User Service) - Login flow

**Pattern Summary:**
{Key rules from the ref}
```

---

## Anti-Patterns

| Anti-Pattern | Why It Fails | Correct |
|--------------|-------------|---------|
| Create ref without user input | Vague, unhelpful | Extract specifics from prompt |
| Update ref without impact check | Silent breakage | Always check citings |
| Duplicate ref content in components | Maintenance nightmare | Cite, don't duplicate |
| Create ref for one-off pattern | Unnecessary overhead | Refs for repeated patterns |
