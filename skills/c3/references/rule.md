# Rule Reference

Manage coding standards as arch artifacts.

Hard rule: enforceable at code level + has golden example → create rule, not ref.

## Mode Selection

| Intent | Mode |
|--------|------|
| "add/create/document a coding rule" | **Add** |
| "update/modify rule-X" | **Update** |
| "list rules", "what rules exist" | **List** |
| "who uses rule-X" | **Usage** |
| "migrate refs to rules", "split ref into rule" | **Migrate** |
| "remove/deprecate rule-X" | **change** (needs ADR) |
| "adopt rule-X", "marketplace adopt" | **Adopt** |

---

## Add

Flow: `Scaffold → Discover → Extract Golden Pattern → Confirm → Fill Content → Discover Usage → Update Citings → ADR`

**HARD RULE: First Bash call must be scaffold.**

### Step 1: Scaffold

```bash
bash <skill-dir>/bin/c3x.sh add rule <slug>
```

### Step 2: Discover (2-5 Grep calls)

Search codebase for existing implementations.

| Findings | Mode | Action |
|----------|------|--------|
| 0 files | **Describe** | User describes standard |
| 1 file | **Extract** (low confidence) | Extract, flag for confirmation |
| 2+ files | **Extract** (compare top 3) | Structural intersection = golden pattern |
| User provides | **Accept** | Use directly |

### Step 3: Extract Golden Pattern

From discovered code:
- **Shared structure** → `## Golden Example`
- **Varies by context** → multiple code blocks with context labels
- **Clearly wrong** → `## Not This`

### Step 4: Confirm

`AskUserQuestion` — approve pattern (ASSUMPTION_MODE: skip).

### Step 5: Quality Gate

Write 1-3 YES/NO compliance questions from `## Rule` + `## Golden Example`. Can't write them → standard too vague, rework.

### Step 6: Fill Content

- `## Rule` — one-line statement of what must be true
- `## Golden Example` — canonical code (code blocks, do/don't pairs)
- `## Not This` — anti-patterns with why wrong
- `## Scope`, `## Override` — as needed

### Step 7: Discover Usage (2-3 Grep calls)

Find components using pattern.

### Step 8: Update Citing Components

Per component using pattern:
1. `c3x lookup <file>` per code-map entry
2. `c3x read <component-id>`
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

Rule adoption ADRs use `status: implemented` — rule doc IS deliverable.

---

## Update

Flow: `Clarify → Find Citings → Check Compliance → Surface Impact → Execute`

1. **Clarify:** `AskUserQuestion` — add/modify/remove rule or clarify docs (ASSUMPTION_MODE: skip)
2. **Find citings:** `c3x list` → rule entity → `relationships`. Depth: `c3x query rule-{slug}`.
3. **Check compliance:** `c3x lookup <file>` per code-map entry. Compare against `## Golden Example` + `## Not This`. Categorize: compliant / needs-update / breaking.
4. **Surface impact:** `AskUserQuestion` — proceed/narrow/cancel (ASSUMPTION_MODE: skip)
5. **Execute:** Update rule doc + create ADR. Non-compliant → TODO in ADR (no code changes).
6. Code changes → route to change.

---

## List

```bash
bash <skill-dir>/bin/c3x.sh list
```

Filter `type: "rule"`. Show: id, title, goal, citing components.

---

## Usage

```bash
bash <skill-dir>/bin/c3x.sh list
```

Find `id: "rule-{slug}"`, read `relationships`. `c3x read <id>` each citing doc.

**Citation Graph:** `c3x graph rule-<slug> --format mermaid` → include as mermaid block.

---

## Migrate

Flow: `Scan → Classify → Split/Convert → Rewire → ADR`

Use when adopting rules in project with existing refs, or auditing refs for rule candidates.

### Step 1: Scan existing refs

```bash
bash <skill-dir>/bin/c3x.sh list
```

Filter `type: "ref"`. `c3x read <ref-id>` each.

### Step 2: Apply Separation Test

Per ref: **"Remove Why section. Doc becomes useless?"**

| Answer | Classification | Action |
|--------|---------------|--------|
| Yes — useless without Why | **Pure ref** | Keep as-is |
| No — still tells what to do | **Pure rule** | Convert |
| Partially — rationale AND enforceable | **Dual-nature** | Split |

Present classification to user (ASSUMPTION_MODE: mark `[ASSUMED]`).

### Step 3a: Convert (pure rule)

Ref entirely about enforcement, not rationale.

1. `c3x add rule <slug>` — scaffold
2. Copy content, adapting sections:
   - `## Goal` → keep
   - `## How` → `## Golden Example`
   - `## Choice` → `## Rule` (one-line)
   - `## Not This` → keep
   - `## Why` → `origin:` if parent ref exists, else drop
3. Set `origin: [ref-{old-slug}]` if old ref kept for rationale
4. `c3x wire <component> rule-{slug}` per citer
5. If deleting ref: `c3x wire --remove <component> ref-{slug}` per citer, then delete
6. Move ref's file patterns to rule in code-map

### Step 3b: Split (dual-nature)

Ref has both rationale AND enforcement.

1. **Narrow ref** — keep only rationale:
   - `## Choice` — decision
   - `## Why` — reasoning
   - Remove/thin `## How` to high-level only
   - Keep `## Not This` if about rejected alternatives (not code anti-patterns)

2. **Create rule** — extract enforcement:
   - `c3x add rule <slug>`
   - `## Rule` — one-line from `## How`
   - `## Golden Example` — code patterns from `## How`
   - `## Not This` — code anti-patterns (not rejected alternatives)
   - Set `origin: [ref-{original-slug}]`

3. **Rewire citations:**
   - Need rationale → keep `ref-{slug}` in `uses:`
   - Need enforcement → add `rule-{slug}` via `c3x wire`
   - Most need both → keep ref, add rule

4. **Update code-map:**
   - Ref keeps/narrows patterns
   - Rule gets enforcement file patterns

### Step 4: Adoption ADR

One ADR per migration batch:

```yaml
---
id: adr-YYYYMMDD-migrate-refs-to-rules
title: Migrate enforcement refs to coding rules
status: implemented
affects: [ref-X, ref-Y, rule-A, rule-B]
---
```

Body: list what was converted, split, or kept.

### Step 5: Verify

```bash
bash <skill-dir>/bin/c3x.sh check
bash <skill-dir>/bin/c3x.sh list
```

Confirm: no orphan refs, all rules have golden examples, all citations intact.

---

## Adopt

Flow: `Preview → Discover Overlap → Guided Merge → Write → Wire → ADR`

Adopt marketplace rule into project `.c3/rules/`.

### Step 1: Preview

```bash
bash <skill-dir>/bin/c3x.sh marketplace show <rule-id>
```

Display full rule content. If `--source` needed, `AskUserQuestion` (ASSUMPTION_MODE: pick first match).

### Step 2: Discover Overlap (2-5 Grep calls)

Search project for patterns overlapping marketplace rule:
- Existing rules/refs covering similar ground (`c3x query`)
- Code matching `## Golden Example`
- Anti-patterns matching `## Not This`

Significant overlap found → present to user before merge.

### Step 3: Section-by-Section Guided Merge

Per section (Goal, Rule, Golden Example, Not This, Scope):

`AskUserQuestion` with options (ASSUMPTION_MODE: adopt as-is):
- **Adopt as-is** — marketplace version verbatim
- **Adapt** — rewrite for project conventions, tech stack, naming
- **Skip** — omit (only optional sections: Scope, Override)

Required sections (Rule, Golden Example) cannot be skipped.

### Step 4: Write

```bash
bash <skill-dir>/bin/c3x.sh add rule <slug>
bash <skill-dir>/bin/c3x.sh set rule-<slug> goal "<adapted goal>"
bash <skill-dir>/bin/c3x.sh set rule-<slug> --section "Rule" "<adapted rule statement>"
bash <skill-dir>/bin/c3x.sh set rule-<slug> --section "Golden Example" "<adapted example>"
bash <skill-dir>/bin/c3x.sh set rule-<slug> --section "Not This" "<adapted anti-patterns>"
```

### Step 5: Wire

Per component from overlap search:
```bash
bash <skill-dir>/bin/c3x.sh wire <component-id> rule-<slug>
```

### Step 6: Adoption ADR

```bash
bash <skill-dir>/bin/c3x.sh add adr adopt-rule-<slug>
bash <skill-dir>/bin/c3x.sh set adr-YYYYMMDD-adopt-rule-<slug> status implemented
```

Body: note source marketplace and adaptations.

---

## Anti-Patterns

| Anti-Pattern | Correct |
|--------------|---------|
| Rule without golden example | Extract or describe concrete pattern |
| Update rule without compliance check | Always check citings against golden example |
| Duplicate rule content in components | Cite, don't duplicate |
| Rule for one-off pattern | Rules for repeated standards only |
| Confuse rule with ref | Rule = enforcement, Ref = rationale (Separation Test) |
| Adopt without checking overlap | Always discover existing patterns first |
| Adopt and keep marketplace default verbatim | Adapt to project conventions |
