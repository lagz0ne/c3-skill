# Reference System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace auxiliary component type with a unified reference system (`.c3/refs/`) that enables flexible, goal-driven documentation at system-wide scope.

**Architecture:** Remove auxiliary category, add flat `.c3/refs/` folder with `ref-{slug}.md` files. Update all templates to Goal-first with universal guidance. Update ref hygiene at each C3 level (context/container/component).

**Tech Stack:** Markdown templates, YAML frontmatter, Mermaid diagrams

---

## Overview Diagrams

- Association: https://diashort.apps.quickable.co/d/05ce9541
- Structure + Ref Flow: https://diashort.apps.quickable.co/d/aa0e3ede
- Flows: https://diashort.apps.quickable.co/d/88d7581a

## Target Structure

```
.c3/
├── README.md              (c3-0 Context)
├── refs/                  (flat, system-wide)
│   └── ref-*.md
├── adr/
│   └── adr-*.md
├── c3-1-slug/
│   ├── README.md          (Container)
│   ├── c3-101-slug.md     (Component - foundation)
│   └── c3-102-slug.md     (Component - feature)
└── TOC.md
```

---

## Task 1: Create Reference Template

**Files:**
- Create: `templates/ref.md`
- Delete: `templates/component-auxiliary.md`

**Step 1: Create ref.md template**

```markdown
---
id: ref-${SLUG}
title: ${TITLE}
---

# ${TITLE}

## Goal

{What problem does this reference solve? Be specific.}

<!--
WHY DOCUMENT:
- Enforce consistency (current and future work)
- Enforce quality (current and future work)
- Support auditing (verifiable, cross-referenceable)
- Be maintainable (worth the upkeep cost)

ANTI-GOALS:
- Over-documenting → stale quickly, maintenance burden
- Text walls → hard to review, hard to maintain
- Isolated content → can't verify from multiple angles

PRINCIPLES:
- Diagrams over text. Always.
- Fewer meaningful sections > many shallow sections
- Add sections that elaborate the Goal - remove those that don't
- Cross-content integrity: same fact from different angles aids auditing

GUARDRAILS:
- Must have: Goal section
- Prefer: 2-4 focused sections
- Each section must serve the Goal - if not, delete
- If a section grows large, consider: diagram? split?

Common sections (create whatever serves your Goal):
- Rationale, Constraints, Boundaries, Examples

But create whatever sections best elaborate your goal.

Delete this comment block after drafting.
-->
```

**Step 2: Delete component-auxiliary.md**

Run: `rm templates/component-auxiliary.md`

**Step 3: Commit**

```bash
git add templates/ref.md
git rm templates/component-auxiliary.md
git commit -m "feat: add ref template, remove auxiliary template"
```

---

## Task 2: Update Component Template (Unified)

**Files:**
- Delete: `templates/component-foundation.md`
- Delete: `templates/component-feature.md`
- Create: `templates/component.md`

**Step 1: Create unified component.md template**

```markdown
---
id: c3-${N}${NN}
c3-version: 3
title: ${COMPONENT_NAME}
type: component
category: foundation | feature
parent: c3-${N}
summary: ${SUMMARY}
---

# ${COMPONENT_NAME}

## Goal

{Why does this component exist? What problem does it solve?}

<!--
WHY DOCUMENT:
- Enforce consistency (current and future work)
- Enforce quality (current and future work)
- Support auditing (verifiable, cross-referenceable)
- Be maintainable (worth the upkeep cost)

ANTI-GOALS:
- Over-documenting → stale quickly, maintenance burden
- Text walls → hard to review, hard to maintain
- Isolated content → can't verify from multiple angles

PRINCIPLES:
- Diagrams over text. Always.
- Fewer meaningful sections > many shallow sections
- Add sections that elaborate the Goal - remove those that don't
- Cross-content integrity: same fact from different angles aids auditing

GUARDRAILS:
- Must have: Goal section
- Prefer: 2-4 focused sections
- Each section must serve the Goal - if not, delete
- If a section grows large, consider: diagram? split? ref-*?

REF HYGIENE (component level = component-specific concerns):
- Before writing: does a ref-* already cover this? Cite, don't duplicate.
- Each ref-* cited must directly serve the Goal - no tangential refs.
- If you're duplicating a ref, cite it instead.

Common sections (create whatever serves your Goal):
- Contract (Provides/Expects), Dependencies, Behavior, Edge Cases, Constraints

Delete this comment block after drafting.
-->

## References

<!-- Code symbols, patterns, paths. Each ref-* cited should serve the Goal. -->
```

**Step 2: Delete old component templates**

Run: `rm templates/component-foundation.md templates/component-feature.md`

**Step 3: Commit**

```bash
git add templates/component.md
git rm templates/component-foundation.md templates/component-feature.md
git commit -m "feat: unify component templates with Goal-first structure"
```

---

## Task 3: Update Container Template

**Files:**
- Modify: `templates/container.md`
- Modify: `templates/container-service.md`
- Modify: `templates/container-database.md`
- Modify: `templates/container-queue.md`

**Step 1: Update container.md**

Replace entire file with:

```markdown
---
id: c3-${N}
c3-version: 3
title: ${CONTAINER_NAME}
type: container
parent: c3-0
summary: ${SUMMARY}
---

# ${CONTAINER_NAME}

## Goal

{Why does this container exist? What bounded context does it own?}

<!--
WHY DOCUMENT:
- Enforce consistency (current and future work)
- Enforce quality (current and future work)
- Support auditing (verifiable, cross-referenceable)
- Be maintainable (worth the upkeep cost)

ANTI-GOALS:
- Over-documenting → stale quickly, maintenance burden
- Text walls → hard to review, hard to maintain
- Isolated content → can't verify from multiple angles

PRINCIPLES:
- Diagrams over text. Always.
- Fewer meaningful sections > many shallow sections
- Add sections that elaborate the Goal - remove those that don't
- Cross-content integrity: same fact from different angles aids auditing

GUARDRAILS:
- Must have: Goal + Components table
- Prefer: 3-5 focused sections
- Each section must serve the Goal - if not, delete
- If a section grows large, consider: diagram? split? ref-*?

REF HYGIENE (container level = cross-component concerns):
- Cite refs that govern how components in this container interact
  (communication patterns, error propagation, shared data flow)
- Component-specific ref usage belongs in component docs, not here
- If a pattern only affects one component, document it there instead

Common sections (create whatever serves your Goal):
- Overview (diagram), Components, Complexity Assessment, Fulfillment, Linkages

Delete this comment block after drafting.
-->

## Components

| ID | Name | Category | Status | Responsibility |
|----|------|----------|--------|----------------|
<!-- Category: foundation | feature -->
```

**Step 2: Update container-service.md similarly**

Add Goal section, universal guidance, update Components table (remove Auxiliary subsection).

**Step 3: Update container-database.md similarly**

**Step 4: Update container-queue.md similarly**

**Step 5: Commit**

```bash
git add templates/container*.md
git commit -m "feat: update container templates with Goal-first, remove auxiliary section"
```

---

## Task 4: Update Context Template

**Files:**
- Modify: `templates/context.md`

**Step 1: Update context.md**

Replace entire file with:

```markdown
---
id: c3-0
c3-version: 3
title: ${PROJECT}
summary: ${SUMMARY}
---

# ${PROJECT}

## Goal

{Why does this system exist? What business problem does it solve?}

<!--
WHY DOCUMENT:
- Enforce consistency (current and future work)
- Enforce quality (current and future work)
- Support auditing (verifiable, cross-referenceable)
- Be maintainable (worth the upkeep cost)

ANTI-GOALS:
- Over-documenting → stale quickly, maintenance burden
- Text walls → hard to review, hard to maintain
- Isolated content → can't verify from multiple angles

PRINCIPLES:
- Diagrams over text. Always.
- Fewer meaningful sections > many shallow sections
- Add sections that elaborate the Goal - remove those that don't
- Cross-content integrity: same fact from different angles aids auditing

GUARDRAILS:
- Must have: Goal + Containers table
- Prefer: 3-5 focused sections
- This is the entry point - navigable, not exhaustive

REF HYGIENE (context level = system-wide concerns):
- Cite refs that govern cross-container behavior
  (system-wide error strategy, auth patterns, inter-container data flow)
- Container-specific patterns belong in container docs
- If a ref only applies within one container, cite it there instead

Common sections (create whatever serves your Goal):
- Overview (diagram), Actors, Containers, External Systems, Linkages

Delete this comment block after drafting.
-->

## Containers

| ID | Name | Type | Status | Purpose |
|----|------|------|--------|---------|
<!-- Type: service | app | library | external -->
```

**Step 2: Commit**

```bash
git add templates/context.md
git commit -m "feat: update context template with Goal-first structure"
```

---

## Task 5: Update v3-structure.md

**Files:**
- Modify: `references/v3-structure.md`

**Step 1: Add refs folder to structure**

Find the file structure section and add:

```markdown
| Refs | `ref-{slug}` | `.c3/refs/ref-{slug}.md` | c3-0 |
```

**Step 2: Add refs frontmatter documentation**

```markdown
## Reference Frontmatter

```yaml
id: ref-{slug}
title: {Reference Name}
```

**Step 3: Update component categories**

Remove `auxiliary` from valid categories. Update to: `foundation | feature`

**Step 4: Commit**

```bash
git add references/v3-structure.md
git commit -m "docs: add refs to v3-structure, remove auxiliary category"
```

---

## Task 6: Update component-types.md

**Files:**
- Modify: `references/component-types.md`

**Step 1: Remove Auxiliary section**

Delete the entire Auxiliary section from the decision flowchart and descriptions.

**Step 2: Update dependency rules**

Remove references to Auxiliary in dependency direction rules.

**Step 3: Add Refs section**

Add new section explaining refs:

```markdown
## References (.c3/refs/)

References are system-wide patterns and conventions cited by components at any level.

**What belongs in refs:**
- Design patterns (strategy choices)
- Coding conventions
- Data flow patterns
- External standards references

**Refs are NOT components** - they don't have code references or implementations.
Components cite refs; refs explain patterns.
```

**Step 4: Commit**

```bash
git add references/component-types.md
git commit -m "docs: remove auxiliary, add refs documentation"
```

---

## Task 7: Update layer-navigation.md

**Files:**
- Modify: `references/layer-navigation.md`

**Step 1: Add ref lookup to navigation**

Add section:

```markdown
## Reference Resolution

When navigating and a pattern/convention is mentioned:

1. Check if component cites a `ref-*`
2. Look up ref in `.c3/refs/ref-{slug}.md`
3. Refs explain patterns; components explain usage
```

**Step 2: Commit**

```bash
git add references/layer-navigation.md
git commit -m "docs: add ref resolution to layer navigation"
```

---

## Task 8: Update audit-checks.md

**Files:**
- Modify: `references/audit-checks.md`

**Step 1: Remove auxiliary validation**

Update category validation to only check for `foundation | feature`.

**Step 2: Add ref validation**

Add new audit check:

```markdown
## Reference Validation

- [ ] All `ref-*` files have Goal section
- [ ] All cited refs (`ref-*` in component docs) exist in `.c3/refs/`
- [ ] No orphan refs (refs not cited by any component)
- [ ] Refs don't duplicate component content
```

**Step 3: Commit**

```bash
git add references/audit-checks.md
git commit -m "docs: update audit checks for refs, remove auxiliary"
```

---

## Task 9: Update c3 Skill

**Files:**
- Modify: `skills/c3/SKILL.md`

**Step 1: Update component categories**

Change all references from `foundation/auxiliary/feature` to `foundation/feature`.

**Step 2: Add refs to adopt flow**

In the adopt workflow, add step to identify and create refs:

```markdown
### Refs Discovery

After components are documented:
1. Identify repeated patterns across components
2. Extract to `.c3/refs/ref-{pattern}.md`
3. Update components to cite refs instead of duplicating
```

**Step 3: Update templates reference**

Point to new `templates/component.md` (unified) and `templates/ref.md`.

**Step 4: Commit**

```bash
git add skills/c3/SKILL.md
git commit -m "feat: update c3 skill for refs system"
```

---

## Task 10: Update c3-query Skill

**Files:**
- Modify: `skills/c3-query/SKILL.md`

**Step 1: Add ref lookup to query flow**

When navigating, include refs in lookup:

```markdown
### Reference Lookup

If query relates to patterns/conventions:
1. Check `.c3/refs/` for `ref-*` matching topic
2. Return ref content + citing components
```

**Step 2: Commit**

```bash
git add skills/c3-query/SKILL.md
git commit -m "feat: add ref lookup to c3-query"
```

---

## Task 11: Update c3-alter Skill

**Files:**
- Modify: `skills/c3-alter/SKILL.md`

**Step 1: Add ref maintenance to alter flow**

When making changes that affect patterns:

```markdown
### Ref Maintenance

If change affects a pattern:
1. Check if `ref-*` exists for pattern
2. Update ref if pattern changes
3. Create new ref if pattern is new and reusable
```

**Step 2: Remove auxiliary from affected categories**

Update any validation that checks for auxiliary category.

**Step 3: Commit**

```bash
git add skills/c3-alter/SKILL.md
git commit -m "feat: add ref maintenance to c3-alter"
```

---

## Task 12: Update Eval Fixtures

**Files:**
- Modify: `eval/fixtures/realistic-project/.c3/`

**Step 1: Migrate auxiliary components to refs**

For each auxiliary component (c3-111 through c3-115):
1. Create corresponding `ref-*.md` in `.c3/refs/`
2. Update citing components to reference `ref-*`
3. Delete old auxiliary component files

**Step 2: Update container README**

Remove Auxiliary section from components table, add refs citations.

**Step 3: Commit**

```bash
git add eval/fixtures/
git commit -m "test: migrate eval fixtures from auxiliary to refs"
```

---

## Task 13: Final Verification

**Step 1: Run type check**

```bash
bunx @typescript/native-preview
```

Expected: No type errors

**Step 2: Build plugin**

```bash
bun run build
```

Expected: Build succeeds

**Step 3: Manual verification**

- Check all templates have Goal section
- Check all templates have universal guidance
- Check refs folder documented in v3-structure
- Check audit-checks validates refs

**Step 4: Final commit**

```bash
git add -A
git commit -m "chore: reference system implementation complete"
```

---

## Summary of Changes

| Area | Change |
|------|--------|
| Component types | `foundation/auxiliary/feature` → `foundation/feature` |
| Templates | All Goal-first with universal guidance |
| New folder | `.c3/refs/` with `ref-{slug}.md` files |
| New template | `templates/ref.md` |
| Deleted | `templates/component-auxiliary.md`, `component-foundation.md`, `component-feature.md` |
| Unified | `templates/component.md` (covers both categories) |
| Skills | Updated c3, c3-query, c3-alter for refs |
| References | Updated v3-structure, component-types, layer-navigation, audit-checks |
