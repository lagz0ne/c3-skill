# Ref Reference

Manage patterns as arch artifacts.

Hard rule: can't name concrete file → create ref, not component.

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

Flow: `Scaffold → Discover → Fill Content → Discover Usage → Update Citings → ADR`

**HARD RULE: First Bash call must be scaffold.**

### Step 1: Scaffold

```bash
bash <skill-dir>/bin/c3x.sh add ref <slug>
```

### Step 2: Discover (2-5 Grep calls)

Search codebase for existing implementations.

| Findings | Mode | Action |
|----------|------|--------|
| 0 files | **Describe** | User describes pattern |
| 1 file | **Extract** (low confidence) | Extract, flag for confirmation |
| 2+ files | **Extract** (compare top 3) | Structural intersection = pattern |
| User provides | **Accept** | Use directly |

### Step 2c: Extract Pattern

From discovered code:
- **Shared structure** → `## How` (golden pattern)
- **Varies by context** → `## Choice` (decision point)
- **Clearly wrong** → `## Not This` (anti-pattern)

Annotate: `// REQUIRED` vs `// OPTIONAL` for structural elements.

### Step 2d: Confirm

`AskUserQuestion` — approve pattern (ASSUMPTION_MODE: skip).

### Step 2e: Quality Gate

Write 1-3 YES/NO compliance questions from `## How`. Can't write them → pattern too vague, rework.

### Step 3: Fill Content

**First:** `c3x schema ref` — the output leads with `REJECT IF:` bullets that ARE the rejection contract. Per-section `fill:` and `rejected when:` lines apply the same gate at section level. Draft to the contract, do not freehand.

- `## Goal` — what it standardizes
- `## Choice` — option chosen (REQUIRED)
- `## Why` — rationale (REQUIRED)
- `## How` — golden pattern (code blocks, do/don't pairs, checklists)
- `## Not This` — rejected alternatives + anti-examples
- `## Scope`, `## Override` — as needed

### Step 4: Discover Usage (2-3 Grep calls)

Find components using pattern.

### Step 5: Update Citing Components

Per component using pattern:
1. `c3x lookup <file>` per code-map entry — loads constraint chain
2. `c3x read <component-id>`
3. Add to `## Related Refs`:

```markdown
## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-error-handling | Uses error response format |
```

Only modify `## Related Refs`. Other changes → route to change.

### Step 6: Adoption ADR

ADRs cannot be created as `implemented` and cannot transition `proposed → implemented` directly. Two-step:

```bash
bash <skill-dir>/bin/c3x.sh add adr ref-{slug}-adoption < adr-body.md
bash <skill-dir>/bin/c3x.sh set adr-YYYYMMDD-ref-{slug}-adoption status accepted
# ref doc is wired and the deliverable is in place
bash <skill-dir>/bin/c3x.sh set adr-YYYYMMDD-ref-{slug}-adoption status implemented
```

Final state:
```yaml
---
id: adr-YYYYMMDD-ref-{slug}-adoption
title: Adopt {Pattern Title} as standard
status: implemented
---
```

Ref adoption ADRs end in `status: implemented` — ref doc IS deliverable. After implemented, the ADR becomes historical and is exempt from `c3x check` validation.

---

## Update

Flow: `Clarify → Find Citings → Check Compliance → Surface Impact → Execute`

1. **Clarify:** `AskUserQuestion` — add/modify/remove rule or clarify docs (ASSUMPTION_MODE: skip)
2. **Find citings:** `c3x list` → ref entity → `relationships`. Depth: `c3x graph ref-{slug} --direction reverse`.
3. **Check compliance:** `c3x lookup <file>` per code-map entry. Categorize: compliant / needs-update / breaking.
4. **Surface impact:** `AskUserQuestion` — proceed/narrow/cancel (ASSUMPTION_MODE: skip)
5. **Execute:** Update ref doc + create ADR. Non-compliant → TODO in ADR (no code changes).
6. Code changes → route to change.

---

## List

```bash
bash <skill-dir>/bin/c3x.sh list
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
bash <skill-dir>/bin/c3x.sh list
```

Find `id: "ref-{slug}"`, read `relationships`. `c3x read <id>` each citing doc.

**Citation Graph:** `c3x graph ref-<slug> --format mermaid` → include as mermaid block.

```
**ref-{slug} Usage**

**Cited by:**
- c3-101 (Auth Middleware) - JWT validation

**Citation Graph:**
(mermaid block from c3x graph)

**Pattern Summary:** {Key rules}
```

---

## Separation Test: Ref vs Rule

Ask: **"Remove Why section. Doc becomes useless?"**

| Answer | Type | Action |
|--------|------|--------|
| Yes — useless without Why | **Ref** | Create ref (this flow) |
| No — still tells what to do | **Rule** | Route to `references/rule.md` Add |
| Both — rationale AND enforcement | **Dual** | Ref for rationale + rule for enforcement (see `references/rule.md` Migrate) |

Primarily about enforcement (golden examples, coding standards) → rule, not ref.

---

## Anti-Patterns

| Anti-Pattern | Correct |
|--------------|---------|
| Create ref without user input | Extract specifics from prompt |
| Update ref without impact check | Always check citings |
| Duplicate ref content in components | Cite, don't duplicate |
| Ref for one-off pattern | Refs for repeated patterns only |
| Ref for enforceable coding standard | Use rule (Separation Test) |
