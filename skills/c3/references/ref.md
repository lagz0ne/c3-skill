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
c3 add ref <slug>
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

**First:** `c3 schema ref` — the output leads with `REJECT IF:` bullets that ARE the rejection contract. Per-section `fill:` and `rejected when:` lines apply the same gate at section level. Draft to the contract, do not freehand.

- `## Goal` — what it standardizes
- `## Choice` — option chosen (REQUIRED)
- `## Why` — rationale (REQUIRED)
- `## How` — golden pattern (code blocks, do/don't pairs, checklists)
- `## Not This` — rejected alternatives + anti-examples
- `## Scope`, `## Override` — as needed

### Step 4: Discover Usage (2-3 Grep calls)

Find components using pattern.

### Step 5: Update Citing Components

A component is a frozen fact and its `uses`/cite edges are derived from its own body (`## Related Refs` / frontmatter `refs:`) at import — never created by `c3 wire`. So how you add a citation depends on whether the component already exists:

- **Brand-new citer (being created now):** author `## Related Refs` directly in its body file, then `c3 add component <slug> --file body.md`. The edge appears at import — no `c3 wire`.
- **Existing citer (frozen):** adding a citation is an edit to its frozen body, so it MUST ride as a change-unit patch on the `## Related Refs` block. Per component using pattern:
  1. `c3 lookup <file>` per code-map entry — loads constraint chain
  2. `c3 read <component-id> --section "Related Refs" --cite` — cite the block
  3. Author `.c3/changes/<adr-id>/<seq>-<slug>.patch.md` adding the row, then `c3 change apply <adr-id>`

The added row looks like:

```markdown
## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-error-handling | Uses error response format |
```

Only the `## Related Refs` block changes. Other changes → route to change as their own patch.

### Step 6: Adoption ADR

Canonical flow — never type or `set` a terminal status; the latch does it:

```bash
c3 add adr ref-{slug}-adoption < adr-body.md
# wire the ref / land the deliverable so the ADR's per-row After cites resolve fresh
c3 change accept adr-YYYYMMDD-ref-{slug}-adoption
c3 check --fix   # auto-latches accepted → done once every After cite resolves
```

Final state:
```yaml
---
id: adr-YYYYMMDD-ref-{slug}-adoption
title: Adopt {Pattern Title} as standard
status: done
---
```

Ref adoption ADRs end in `status: done` — ref doc IS deliverable. `accepted → done` is a one-way auto-done latch: you never type or `set` `done`; `c3 check --fix` actualizes it when the per-row *After* cites all resolve fresh. Once `done`, the ADR is terminal/historical and exempt from `c3 check` validation.

---

## Update

Flow: `Clarify → Find Citings → Check Compliance → Surface Impact → Execute`

1. **Clarify:** `AskUserQuestion` — add/modify/remove rule or clarify docs (ASSUMPTION_MODE: skip)
2. **Find citings:** `c3 list` → ref entity → `relationships`. Depth: `c3 graph ref-{slug} --direction reverse`.
3. **Check compliance:** `c3 lookup <file>` per code-map entry. Categorize: compliant / needs-update / breaking.
4. **Surface impact:** `AskUserQuestion` — proceed/narrow/cancel (ASSUMPTION_MODE: skip)
5. **Execute:** A ref is a frozen fact — `c3 write`/`c3 set`/`c3 wire` on an existing ref is refused ("…is a fact — facts are frozen and change only through a change-unit"). Create the ADR as the change-unit, then route the ref edit through it: `c3 read ref-{slug} --section <name> --cite` → author `.c3/changes/<adr-id>/<seq>-<slug>.patch.md` → `c3 change apply <adr-id>`. Non-compliant → TODO in ADR (no code changes).
6. Code changes → route to change.

---

## List

```bash
c3 list
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
c3 list
```

Find `id: "ref-{slug}"`, read `relationships`. `c3 read <id>` each citing doc.

**Citation Graph:** `c3 graph ref-<slug> --format mermaid` → include as mermaid block.

```
**ref-{slug} Usage**

**Cited by:**
- c3-101 (Auth Middleware) - JWT validation

**Citation Graph:**
(mermaid block from c3 graph)

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
