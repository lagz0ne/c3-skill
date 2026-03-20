# Rule Reference

Manage coding standards as first-class architecture artifacts.

Hard rule: if it's enforceable at code level and has a golden example → create rule, not ref.

## Mode Selection

| Intent | Mode |
|--------|------|
| "add/create/document a coding rule" | **Add** |
| "update/modify rule-X" | **Update** |
| "list rules", "what rules exist" | **List** |
| "who uses rule-X" | **Usage** |
| "remove/deprecate rule-X" | **change** (needs ADR) |

---

## Add

Flow: `Scaffold → Discover → Extract Golden Pattern → Confirm → Fill Content → Discover Usage → Update Citings → ADR`

**HARD RULE: First Bash call must be scaffold.**

### Step 1: Scaffold

```bash
bash <skill-dir>/bin/c3x.sh add rule <slug>
```

### Step 2: Discover (2-5 Grep calls)

Search for existing implementations of the pattern in the codebase.

| Findings | Mode | Action |
|----------|------|--------|
| 0 files | **Describe** | User describes the standard |
| 1 file | **Extract** (low confidence) | Extract, flag to user for confirmation |
| 2+ files | **Extract** (compare top 3) | Structural intersection = golden pattern |
| User provides | **Accept** | Use user's description directly |

### Step 3: Extract Golden Pattern

From discovered code:
- **Shared structure** → `## Golden Example`
- **Varies by context** → multiple code blocks with context labels
- **Clearly wrong** → `## Not This`

### Step 4: Confirm

`AskUserQuestion` — present extracted pattern for approval (ASSUMPTION_MODE: skip).

### Step 5: Quality Gate

Write 1-3 YES/NO compliance questions derivable from `## Rule` + `## Golden Example`. If you can't write them, the standard is too vague — rework before proceeding.

### Step 6: Fill Content

From discovery + user input:
- `## Rule` — one-line statement of what must be true
- `## Golden Example` — canonical code (format-flexible: code blocks, do/don't pairs)
- `## Not This` — anti-patterns with why they're wrong here
- `## Scope`, `## Override` — as needed

### Step 7: Discover Usage (2-3 Grep calls)

Find components using this pattern.

### Step 8: Update Citing Components

For each component using pattern:
1. Run `c3x lookup <file>` per code-map entry
2. Read component doc
3. Add to `## Related Rules`:

```markdown
## Related Rules

| Rule | Role |
|------|------|
| rule-structured-logging | Logging format enforcement |
```

Only modify `## Related Rules`. Other changes → route to change.

### Step 9: Adoption ADR

```yaml
---
id: adr-YYYYMMDD-rule-{slug}-adoption
title: Adopt {Rule Title} as standard
status: implemented
---
```

Rule adoption ADRs use `status: implemented` directly — rule doc IS the deliverable.

---

## Update

Flow: `Clarify → Find Citings → Check Compliance → Surface Impact → Execute`

1. **Clarify:** `AskUserQuestion` — add rule / modify rule / remove rule / clarify docs (ASSUMPTION_MODE: skip)
2. **Find citings:** `c3x list --json` → rule entity → `relationships`. Grep `rule-{slug}` in `.c3/` for depth.
3. **Check compliance:** `c3x lookup <file>` per code-map entry. Compare against `## Golden Example` and `## Not This` for strict compliance. Categorize: compliant / needs-update / breaking.
4. **Surface impact:** `AskUserQuestion` — proceed / narrow / cancel (ASSUMPTION_MODE: skip)
5. **Execute:** Update rule doc + create ADR. Non-compliant → note as TODO in ADR (don't touch code).
6. Code changes → route to change.

---

## List

```bash
bash <skill-dir>/bin/c3x.sh list --json
```

Filter `type: "rule"`. Show: id, title, goal, citing components.

---

## Usage

```bash
bash <skill-dir>/bin/c3x.sh list --json
```

Find `id: "rule-{slug}"`, read `relationships`. Read each citing doc.

---

## Anti-Patterns

| Anti-Pattern | Correct |
|--------------|---------|
| Create rule without golden example | Extract or describe concrete pattern |
| Update rule without compliance check | Always check citings against golden example |
| Duplicate rule content in components | Cite, don't duplicate |
| Create rule for one-off pattern | Rules for repeated standards only |
| Confuse rule with ref | Rule = enforcement, Ref = rationale (use Separation Test) |
