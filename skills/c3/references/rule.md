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
c3 add rule <slug>
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

**First:** `c3 schema rule` — the output leads with `REJECT IF:` bullets that ARE the rejection contract. Per-section `fill:` and `rejected when:` lines apply the same gate at section level. Draft to the contract, do not freehand.

- `## Rule` — one-line statement of what must be true
- `## Golden Example` — canonical code (code blocks, do/don't pairs)
- `## Not This` — anti-patterns with why wrong
- `## Scope`, `## Override` — as needed

### Step 7: Discover Usage (2-3 Grep calls)

Find components using pattern.

### Step 8: Cite the rule from each using component

A component's `uses` edges come from the **column its canvas marks `edge: uses`** — authoring that column **is** the citation (display row and graph edge are one). It is **canvas-configurable**, so ask `c3 schema component` where it lives (find the column tagged `→ edge: uses`) rather than memorizing a section — in a freshly-seeded project that's the `Governance` table's `Reference` column, which carries `rule-*` and `ref-*` alike, distinguished by the `Type` column. If no column shows an `→ edge:` tag, the project predates the edge column — cite the legacy way (the `Governance` reference row / frontmatter `uses:`). A citation must resolve. Then:

- **Brand-new citer (created now):** author the reference row into that section in the component's body file, then `c3 add component <slug> --file body.md`. The edge appears at import.
- **Existing citer (frozen):** adding a citation edits the frozen body, so it rides as a change-unit patch on **that** section's block:
  1. `c3 lookup <file>` per code-map entry.
  2. `c3 schema component` → find the reference section; `c3 read <component-id> --section <that-section> --cite` → cite the block.
  3. Author `.c3/changes/<adr-id>/<seq>-<slug>.patch.md` adding the reference row, then `c3 change apply <adr-id>`.

Add the row in the shape `c3 schema component` shows — for today's component canvas, the `Governance` row with `Type: rule`:

```markdown
| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| rule-structured-logging | rule | Logging format enforcement | rule is strict — must match golden example | — |
```

Only that section's block changes. Other changes → route to change as their own patch.

### Step 9: Adoption ADR

ADR status set is `[open, accepted, done, superseded]`. You never `set` a terminal status — `accepted → done` is a one-way auto-done latch. Flow:

```bash
c3 add adr rule-{slug}-adoption < adr-body.md
# land the deliverable: the new rule doc (created via `c3 add rule`) is in place
c3 change accept adr-YYYYMMDD-rule-{slug}-adoption
c3 check --fix   # auto-latches accepted → done once the After cites resolve fresh
```

Final state:
```yaml
---
id: adr-YYYYMMDD-rule-{slug}-adoption
title: Adopt {Rule Title} as standard
status: done
---
```

Rule adoption ADRs end in `status: done` — rule doc IS deliverable. `c3 change accept` flips it to `accepted`; the next `c3 check --fix` latches it to `done` when the per-row *After* cites resolve. You never type or `set` `done`. A terminal (`done`/`superseded`) ADR is historical and exempt from `c3 check` validation.

---

## Update

Flow: `Clarify → Find Citings → Check Compliance → Surface Impact → Execute`

1. **Clarify:** `AskUserQuestion` — add/modify/remove rule or clarify docs (ASSUMPTION_MODE: skip)
2. **Find citings:** `c3 list` → rule entity → `relationships`. Depth: `c3 graph rule-{slug} --direction reverse`.
3. **Check compliance:** `c3 lookup <file>` per code-map entry. Compare against `## Golden Example` + `## Not This`. Categorize: compliant / needs-update / breaking.
4. **Surface impact:** `AskUserQuestion` — proceed/narrow/cancel (ASSUMPTION_MODE: skip)
5. **Execute:** A rule is a frozen fact — `c3 write`/`c3 set` on an existing rule is refused ("… is a fact — facts are frozen and change only through a change-unit"). Route the edit through the change-unit: open the ADR as the change-unit, `c3 read rule-{slug} --section <name> --cite` for each section you touch, author a `<seq>-{slug}.patch.md` under `.c3/changes/<adr-id>/`, then `c3 change apply <adr-id>`. Non-compliant citers → TODO in the ADR (no code changes here).
6. Code changes → route to change.

> `c3 set rule-{slug} codemap '<glob>'` is the one exception — the code binding is not frozen and may be updated directly (or carried as a `.codemap.md` inside the change).

---

## List

```bash
c3 list
```

Filter `type: "rule"`. Show: id, title, goal, citing components.

---

## Usage

```bash
c3 list
```

Find `id: "rule-{slug}"`, read `relationships`. `c3 read <id>` each citing doc.

**Citation Graph:** `c3 graph rule-<slug> --format mermaid` → include as mermaid block.

---

## Migrate

Flow: `Scan → Classify → Split/Convert → Rewire → ADR`

Use when adopting rules in project with existing refs, or auditing refs for rule candidates.

### Step 1: Scan existing refs

```bash
c3 list
```

Filter `type: "ref"`. `c3 read <ref-id>` each.

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

1. `c3 add rule <slug>` — scaffold
2. Copy content, adapting sections:
   - `## Goal` → keep
   - `## How` → `## Golden Example`
   - `## Choice` → `## Rule` (one-line)
   - `## Not This` → keep
   - `## Why` → `origin:` if parent ref exists, else drop
3. Set `origin: [ref-{old-slug}]` if old ref kept for rationale
4. Re-edge each citer to the new rule. Citers are frozen facts whose edges come from their canvas's reference column (`c3 schema component` — today `Governance`/`Reference`), so this is a change-unit patch per citer (one ADR for the batch): `c3 read <component-id> --section Governance --cite` → add the `rule-{slug}` row in a `<seq>-<slug>.patch.md` → `c3 change apply <adr-id>`. A brand-new citer would instead carry the row in its body at `c3 add` time.
5. If deleting ref: drop the `ref-{slug}` row from each citer's reference column the same way (change-unit patch), then delete the ref
6. Move ref's file patterns to rule in code-map

### Step 3b: Split (dual-nature)

Ref has both rationale AND enforcement.

1. **Narrow ref** — keep only rationale:
   - `## Choice` — decision
   - `## Why` — reasoning
   - Remove/thin `## How` to high-level only
   - Keep `## Not This` if about rejected alternatives (not code anti-patterns)

2. **Create rule** — extract enforcement:
   - `c3 add rule <slug>`
   - `## Rule` — one-line from `## How`
   - `## Golden Example` — code patterns from `## How`
   - `## Not This` — code anti-patterns (not rejected alternatives)
   - Set `origin: [ref-{original-slug}]`

3. **Rewire citations** (each citer is a frozen fact — edges come from its canvas's reference column, so re-edging rides as a change-unit patch on that block — `c3 schema component`, today `Governance`; a brand-new citer carries the rows in its body at `c3 add`):
   - Need rationale → keep the `ref-{slug}` row
   - Need enforcement → add a `rule-{slug}` row
   - Most need both → keep ref row, add rule row

4. **Update code-map:**
   - Ref keeps/narrows patterns
   - Rule gets enforcement file patterns

### Step 4: Adoption ADR

One ADR per migration batch. ADR status set is `[open, accepted, done, superseded]`. Born `open` (a fresh ADR may show the unmigrated synonym `proposed`); after the migration lands, `c3 change accept <adr-id>` flips it to `accepted`, and the next `c3 check --fix` auto-latches `accepted → done`. You never `set` a terminal status. Final state:

```yaml
---
id: adr-YYYYMMDD-migrate-refs-to-rules
title: Migrate enforcement refs to coding rules
status: done
affects: [ref-X, ref-Y, rule-A, rule-B]
---
```

Body: list what was converted, split, or kept.

### Step 5: Verify

```bash
c3 check
c3 list
```

Confirm: no orphan refs, all rules have golden examples, all citations intact.

---

## Adopt

Flow: `Preview → Discover Overlap → Guided Merge → Write → Cite from Citers → ADR`

Adopt marketplace rule into project `.c3/rules/` (if using marketplace plugin packs).

### Step 1: Preview

```bash
c3 marketplace show <rule-id>
```

Display full rule content. If `--source` needed, `AskUserQuestion` (ASSUMPTION_MODE: pick first match).

### Step 2: Discover Overlap (2-5 Grep calls)

Search project for patterns overlapping marketplace rule:
- Existing rules/refs covering similar ground (`c3 list`; for body text use targeted `c3 read` output)
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

Adopting creates a brand-new local rule, so author its full body at creation. A rule is a frozen fact — once it exists, `c3 set`/`c3 write` on it is refused, so do NOT scaffold-then-patch. Assemble the adapted sections (Goal, Rule, Golden Example, Not This, Scope — code fences and all) into one body file and create the rule from it:

```bash
# body.md holds the full adapted doc: ## Goal, ## Rule, ## Golden Example, ## Not This, ...
c3 add rule <slug> --file body.md
```

To revise the rule LATER (after it exists), route through a change-unit — see the Update flow.

### Step 5: Cite from each using component

Citers are frozen facts and their edges come from their canvas's reference column (`c3 schema component` — today `Governance`/`Reference`). Per component from overlap search, add a `rule-<slug>` row to that column:

- **Existing citer:** ride it as a change-unit patch — `c3 read <component-id> --section Governance --cite` → author `<seq>-<slug>.patch.md` adding the row → `c3 change apply <adr-id>`.
- **Brand-new citer:** author the reference row in its body at `c3 add component <slug> --file body.md` time; the edge appears at import.

### Step 6: Adoption ADR

ADR status set is `[open, accepted, done, superseded]`; `accepted → done` is a one-way auto-done latch you never `set`. Flow:

```bash
c3 add adr adopt-rule-<slug> < adr-body.md
# rule created and adapted from marketplace; deliverable in place
c3 change accept adr-YYYYMMDD-adopt-rule-<slug>
c3 check --fix   # auto-latches accepted → done once the After cites resolve fresh
```

Body with rationale + adaptation notes → `c3 write adr-YYYYMMDD-adopt-rule-<slug> --file body.md`.

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
