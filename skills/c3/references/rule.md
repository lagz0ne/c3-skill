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
| "migrate refs to rules", "split ref into rule" | **Migrate** |
| "remove/deprecate rule-X" | **change** (needs ADR) |
| "adopt rule-X", "install from marketplace", "marketplace adopt" | **Adopt** |

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

Rule adoption ADRs use `status: implemented` directly — rule doc IS the deliverable.

---

## Update

Flow: `Clarify → Find Citings → Check Compliance → Surface Impact → Execute`

1. **Clarify:** `AskUserQuestion` — add rule / modify rule / remove rule / clarify docs (ASSUMPTION_MODE: skip)
2. **Find citings:** `c3x list --json` → rule entity → `relationships`. Search via `c3x query rule-{slug}` for depth.
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

Find `id: "rule-{slug}"`, read `relationships`. `c3x read <id>` each citing doc.

---

## Migrate

Flow: `Scan → Classify → Split/Convert → Rewire → ADR`

Use when adopting rules in a project that already has refs, or when auditing existing refs for rule candidates.

### Step 1: Scan existing refs

```bash
bash <skill-dir>/bin/c3x.sh list --json
```

Filter `type: "ref"`. For each ref, `c3x read <ref-id>`.

### Step 2: Apply Separation Test

For each ref, ask: **"Remove the Why section. Does the doc become useless?"**

| Answer | Classification | Action |
|--------|---------------|--------|
| Yes — doc is useless without Why | **Pure ref** | Keep as-is |
| No — doc still tells you what to do | **Pure rule** | Convert |
| Partially — has both rationale AND enforceable standard | **Dual-nature** | Split |

Present classification table to user for approval (ASSUMPTION_MODE: mark `[ASSUMED]`).

### Step 3a: Convert (pure rule)

The ref is entirely about enforcement, not rationale.

1. `c3x add rule <slug>` — scaffold the rule doc
2. Copy content from ref, adapting sections:
   - `## Goal` → keep
   - `## How` → becomes `## Golden Example`
   - `## Choice` → becomes `## Rule` (one-line statement)
   - `## Not This` → keep
   - `## Why` → move to `origin:` if there's a parent ref, otherwise drop
3. Set `origin: [ref-{old-slug}]` if the old ref is being kept for rationale
4. Update all citing components: `c3x wire <component> rule-{slug}` for each
5. If ref is being deleted: `c3x wire --remove <component> ref-{slug}` for each citer, then delete ref
6. Update code-map: move ref's file patterns to rule

### Step 3b: Split (dual-nature)

The ref has both rationale (why we chose this) AND enforcement (what code must look like).

1. **Narrow the ref** — keep only rationale sections:
   - `## Choice` — the decision
   - `## Why` — the reasoning
   - Remove or thin `## How` to high-level guidance only
   - Keep `## Not This` if it's about rejected alternatives (not code anti-patterns)

2. **Create the rule** — extract enforcement content:
   - `c3x add rule <slug>`
   - `## Rule` — one-line standard extracted from `## How`
   - `## Golden Example` — code patterns from ref's `## How`
   - `## Not This` — code anti-patterns (not rejected alternatives)
   - Set `origin: [ref-{original-slug}]`

3. **Rewire citations:**
   - Components that need the rationale → keep `ref-{slug}` in `uses:`
   - Components that need enforcement → add `rule-{slug}` via `c3x wire`
   - Most components need both → keep ref, add rule

4. **Update code-map:**
   - Ref keeps its patterns (or narrows them)
   - Rule gets the file patterns where enforcement applies

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

Body should list what was converted, split, or kept.

### Step 5: Verify

```bash
bash <skill-dir>/bin/c3x.sh check
bash <skill-dir>/bin/c3x.sh list
```

Confirm: no orphan refs, all rules have golden examples, all citations intact.

---

---

## Adopt

Flow: `Preview → Discover Overlap → Guided Merge → Write → Wire → ADR`

Adopt a rule from a registered marketplace source into the project's `.c3/rules/`.

### Step 1: Preview

```bash
bash <skill-dir>/bin/c3x.sh marketplace show <rule-id>
```

Display full rule content. If `--source` needed to disambiguate, prompt with `AskUserQuestion` (ASSUMPTION_MODE: pick first match).

### Step 2: Discover Overlap (2-5 Grep calls)

Search the project codebase for existing patterns that overlap with the marketplace rule:
- Existing rules/refs covering similar ground (check via `c3x query`)
- Code matching the rule's `## Golden Example`
- Anti-patterns matching `## Not This`

If significant overlap found, present to user before merge.

### Step 3: Section-by-Section Guided Merge

For each rule section (Goal, Rule, Golden Example, Not This, Scope):

`AskUserQuestion` with options (ASSUMPTION_MODE: adopt as-is):
- **Adopt as-is** — take marketplace version verbatim
- **Adapt** — LLM rewrites section for project conventions, tech stack, naming
- **Skip** — omit section (only optional sections: Scope, Override)

Required sections (Rule, Golden Example) cannot be skipped.

### Step 4: Write

```bash
bash <skill-dir>/bin/c3x.sh add rule <slug>
```

Then fill content:
```bash
bash <skill-dir>/bin/c3x.sh set rule-<slug> goal "<adapted goal>"
bash <skill-dir>/bin/c3x.sh set rule-<slug> --section "Rule" "<adapted rule statement>"
bash <skill-dir>/bin/c3x.sh set rule-<slug> --section "Golden Example" "<adapted example>"
bash <skill-dir>/bin/c3x.sh set rule-<slug> --section "Not This" "<adapted anti-patterns>"
```

### Step 5: Wire

For each component the overlap search identified:
```bash
bash <skill-dir>/bin/c3x.sh wire <component-id> rule-<slug>
```

### Step 6: Adoption ADR

```bash
bash <skill-dir>/bin/c3x.sh add adr adopt-rule-<slug>
bash <skill-dir>/bin/c3x.sh set adr-YYYYMMDD-adopt-rule-<slug> status implemented
```

Body: note the source marketplace and any adaptations made.

---

## Anti-Patterns

| Anti-Pattern | Correct |
|--------------|---------|
| Create rule without golden example | Extract or describe concrete pattern |
| Update rule without compliance check | Always check citings against golden example |
| Duplicate rule content in components | Cite, don't duplicate |
| Create rule for one-off pattern | Rules for repeated standards only |
| Confuse rule with ref | Rule = enforcement, Ref = rationale (use Separation Test) |
| Adopt rule without checking overlap | Always discover existing patterns first |
| Adopt rule and keep marketplace default verbatim | Adapt to project conventions |
