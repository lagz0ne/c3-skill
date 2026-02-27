# Ref Reference

Manage patterns as first-class architecture artifacts.

Hard rule: can't name a concrete file → create ref, not component.

## Mode Selection

| Intent | Mode |
|--------|------|
| "add/create/document a pattern" | **Add** |
| "update/modify ref-X" | **Update** |
| "list patterns", "what refs exist" | **List** |
| "who uses ref-X" | **Usage** |
| "remove/deprecate ref-X" | **change** (needs ADR) |

---

## Add

Flow: `Scaffold → Fill Content → Discover Usage → Update Citings → ADR`

**HARD RULE: First Bash call must be scaffold.**

### Step 1: Scaffold

```bash
bash <skill-dir>/bin/c3x.sh add ref <slug>
```

### Step 2: Fill Content

From user's prompt:
- `## Goal` — what it standardizes
- `## Choice` — option chosen (REQUIRED)
- `## Why` — rationale (REQUIRED)
- Other: How, Scope, Not This, Override

Don't search codebase first — user's description is enough for draft.

### Step 3: Discover Usage (2-3 Grep calls)

Find components using this pattern.

### Step 4: Refine (if needed)

Update ref if discovery reveals variations or anti-patterns.

### Step 5: Update Citing Components

For each component using pattern:
1. Run `c3x lookup <file>` per code-map entry — loads constraint chain
2. Read component doc
3. Add to `## Related Refs`:

```markdown
## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-error-handling | Uses error response format |
```

Only modify `## Related Refs`. Other changes → route to change.

### Step 6: Adoption ADR

```yaml
---
id: adr-YYYYMMDD-ref-{slug}-adoption
title: Adopt {Pattern Title} as standard
status: implemented
---
```

Ref adoption ADRs use `status: implemented` directly — ref doc IS the deliverable.

---

## Update

Flow: `Clarify → Find Citings → Check Compliance → Surface Impact → Execute`

1. **Clarify:** `AskUserQuestion` — add rule / modify rule / remove rule / clarify docs (ASSUMPTION_MODE: skip)
2. **Find citings:** `c3x list --json` → ref entity → `relationships`. Grep `ref-{slug}` in `.c3/` for depth.
3. **Check compliance:** `c3x lookup <file>` per code-map entry. Categorize: compliant / needs-update / breaking.
4. **Surface impact:** `AskUserQuestion` — proceed / narrow / cancel (ASSUMPTION_MODE: skip)
5. **Execute:** Update ref doc + create ADR. Non-compliant → note as TODO in ADR (don't touch code).
6. Code changes → route to change.

---

## List

```bash
bash <skill-dir>/bin/c3x.sh list --json
```

Filter `type: "ref"`. Show: id, title, goal, citing components.

```
**C3 Patterns**

| Ref | Title | Goal |
|-----|-------|------|
| ref-error-handling | Error Handling | Consistent errors |
```

---

## Usage

```bash
bash <skill-dir>/bin/c3x.sh list --json
```

Find `id: "ref-{slug}"`, read `relationships`. Read each citing doc.

```
**ref-{slug} Usage**

**Cited by:**
- c3-101 (Auth Middleware) - JWT validation

**Pattern Summary:** {Key rules}
```

---

## Anti-Patterns

| Anti-Pattern | Correct |
|--------------|---------|
| Create ref without user input | Extract specifics from prompt |
| Update ref without impact check | Always check citings |
| Duplicate ref content in components | Cite, don't duplicate |
| Create ref for one-off pattern | Refs for repeated patterns only |
