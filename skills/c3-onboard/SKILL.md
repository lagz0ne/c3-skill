---
name: c3-onboard
description: |
  Creates C3 architecture documentation from scratch using staged Socratic discovery.

  This skill should be used when the user asks to:
  - "adopt C3", "init C3", "initialize C3", "start C3", "bootstrap C3"
  - "scaffold C3 docs", "create C3 docs", "create .c3", "set up .c3"
  - "set up architecture docs", "create architecture documentation"
  - "document this project", "document the codebase", "document the architecture from scratch"
  - "new C3 project", "first time with C3", "never used C3", "getting started with C3"
  - "onboard", "onboard this project", "C3 onboarding"

  DO NOT use when .c3/ already exists (route to c3-audit, c3-change, or c3-query).
  DO NOT use for updating existing docs (route to c3-change).
  Implements 3-stage workflow: Inventory → Details → Finalize.
---

# C3 Onboarding

Initialize C3 architecture documentation using a **3-stage workflow with gates**.

## Precondition: Fresh Project

**If `.c3/README.md` already exists**, this project already has C3 docs.

Use `AskUserQuestion`: "This project already has C3 docs (.c3/ exists). What do you want to do?" with options "Re-onboard from scratch (will overwrite existing docs)" and "Cancel (use c3-audit to check existing docs instead)".

**On Cancel:** Stop and suggest c3-audit or c3-query.
**On Re-onboard:** Proceed with warning that existing docs will be overwritten.

## CRITICAL: File Structure and Categories

**DO NOT create a single file for all documentation.** Each component gets its OWN file. Each container gets its OWN directory.

**Required structure:**

```
.c3/
├── README.md                    # System context
├── adr/
│   └── adr-00000000-c3-adoption.md
├── refs/                        # Cross-cutting patterns
│   └── ref-<pattern>.md         # One file per ref (if any discovered)
└── c3-1-<container-name>/       # Container directory
    ├── README.md                # Container overview
    └── c3-101-<component>.md    # One file per component
```

**Non-negotiable rules:**
- Each component = separate file (c3-101-auth.md, c3-102-users.md, etc.)
- Each container = separate directory (c3-1-api/, c3-2-web/, etc.)
- refs/ = separate directory with pattern files (ref-auth.md, ref-errors.md)

**The Three Categories:** Load `references/component-categories.md` for the full Foundation vs Feature vs Ref rules. Components (Foundation/Feature) MUST have `## Code References`. Refs must NOT. If you cannot name a concrete file, create a ref instead.

---

## REQUIRED: Load References

**Load immediately** (Read these files relative to this skill's directory):
1. `references/skill-harness.md` - Red flags and complexity rules (ignore its ".c3/README.md" check — onboard creates it)
2. `references/layer-navigation.md` - How to traverse C3 docs (ignore its activation check — onboard creates the docs it describes)
3. `references/component-categories.md` - Foundation vs Feature vs Ref classification rules

**Load for templating** (Read when needed, relative to this skill's directory):
- `templates/adr-000.md` - Adoption ADR (drives the workflow)
- `templates/context.md` - Context template
- `templates/container.md` - Container template
- `templates/component.md` - Component template
- `templates/ref.md` - Reference template

---

## Workflow: Inventory → Details → Finalize

**Questioning:** Use `AskUserQuestion` tool for all questions — unless it has been denied.

**HARD RULE — AskUserQuestion Denial:**
If `AskUserQuestion` is denied even once, set `ASSUMPTION_MODE = true` for the rest of this session. When `ASSUMPTION_MODE` is true:
- NEVER call `AskUserQuestion` again — not with fewer questions, not with simpler options, not at all
- For high-impact decisions: state assumption clearly, mark `[ASSUMED]` in ADR-000 Conflict Resolution table
- For low-impact decisions: auto-proceed silently, note in Conflict Resolution
- Every instruction below that says "Use AskUserQuestion" becomes "make your best assumption and mark [ASSUMED]"

---

## Progress Checklist

Copy and track as you work:

```
Onboarding Progress:
- [ ] Stage 0: All items inventoried (context, containers, components, refs)
- [ ] Stage 0: ADR-000 created with discovery tables filled
- [ ] Gate 0: Inventory complete, proceed to Details
- [ ] Stage 1: All container README.md created
- [ ] Stage 1: All component docs created
- [ ] Stage 1: All refs documented
- [ ] Gate 1: No new items discovered
- [ ] Stage 2: Integrity checks pass
- [ ] Gate 2: c3-audit passes
- [ ] ADR-000 marked implemented
```

---

## Stage 0: Inventory

**Goal:** Discover EVERYTHING before creating any docs.

### 0.1 Scaffold and Create ADR-000

Scaffold the base `.c3/` directory structure using the CLI:

```bash
npx -y c3-kit init
```

This creates `.c3/` with `config.yaml`, `README.md` (context template), `refs/` subdirectory, and `adr/adr-00000000-c3-adoption.md` (adoption ADR template). It does NOT create container directories — those are created in Stage 1 after discovery is confirmed at Gate 0.

After init, **Edit** `.c3/adr/adr-00000000-c3-adoption.md` to fill in the adoption ADR content from `templates/adr-000.md` template (discovery tables, workflow diagram, etc.).

### 0.2 Context Discovery

Read codebase structure. Capture template arguments in **ADR-000 Context Discovery table**:

| Arg | Value |
|-----|-------|
| PROJECT | What is this system called? |
| GOAL | Why does this system exist? |
| SUMMARY | One-sentence description |

Also discover **Abstract Constraints** — system-level non-negotiable requirements that cascade to container responsibilities:

| Constraint | Rationale | Affected Containers |
|------------|-----------|---------------------|
| e.g. "All API responses < 200ms" | SLA requirement | c3-1-api |

Use `AskUserQuestion` for gaps (skip if `ASSUMPTION_MODE`; assume and mark `[ASSUMED]`). For example: "I see apps/api/ and apps/web/. Are these separate containers?" with options like "Separate containers", "Single monolith", "Monorepo with shared code".

### 0.3 Container Discovery

A container is a **deployment/runtime boundary** that **allocates responsibilities** derived from context abstract constraints. Each container owns a set of responsibilities that satisfy one or more abstract constraints.

For each potential container, capture in **ADR-000 Container Discovery table**:

| N | CONTAINER_NAME | BOUNDARY | GOAL | SUMMARY |
|---|----------------|----------|------|---------|

Note: `N` is the container number (1, 2, 3, etc.)

Use `AskUserQuestion` to confirm each container (skip if `ASSUMPTION_MODE`; assume based on directory structure and mark `[ASSUMED]`). For example: "Is apps/api/ a deployable backend API?" with options like "Yes - Backend API", "No - Library/shared code", "Need more context".

### 0.4 Component Discovery (Brief)

For each container, scan for components. Capture in **ADR-000 Component Discovery table**:

| N | NN | COMPONENT_NAME | CATEGORY | GOAL | SUMMARY |
|---|----|--------------  |----------|------|---------|

**Categorization (use lowercase in table):**
- **foundation** (NN = 01-09): Provides platform capabilities that others depend on (Router, AuthProvider, Database)
- **feature** (NN = 10+): Composes business logic using foundation capabilities (UserService, CheckoutFlow, Dashboard)

Component IDs encode category: `c3-N01` through `c3-N09` = foundation, `c3-N10`+ = feature.

Use `AskUserQuestion` to confirm categorization (skip if `ASSUMPTION_MODE`; categorize by import count and mark `[ASSUMED]`). For example: "AuthMiddleware is imported by 15 files. Foundation or Feature?" with options like "Foundation - others depend on it", "Feature - domain logic", "Need more context".

### 0.5 Ref Discovery

Look for patterns that repeat across components. Capture in **ADR-000 Ref Discovery table**:

| SLUG | TITLE | GOAL | Scope | Applies To |
|------|-------|------|-------|------------|

Common ref candidates:
- Error handling patterns
- Form patterns
- Query/data fetching patterns
- Design system principles
- User flows / IA

**Ref structure:** Each ref should use sections that serve its Goal: Choice (what we chose), Why (rationale), How (conventions/rules), Not This (anti-patterns), Scope (where it applies), Override (how to deviate). **Choice and Why are required minimum sections.** Other sections are optional based on what serves the Goal.

### 0.6 Overview Diagram

Generate mermaid diagram showing:
- Actors → Containers → External Systems
- Key linkages

### Gate 0

Before proceeding, verify:
- [ ] Context args filled (PROJECT, GOAL, SUMMARY)
- [ ] Abstract Constraints identified
- [ ] All containers identified with args (including BOUNDARY)
- [ ] All components identified (brief) with args and category
- [ ] Cross-cutting refs identified (with Choice and Why minimum)
- [ ] Overview diagram generated

**If gaps:** Return to discovery with Socratic questioning.

---

## Stage 1: Details

**Goal:** Create all docs from inventory.

### 1.1 Context Doc

`.c3/README.md` was created by `npx -y c3-kit init` from the context template. **Edit** it to fill in discovered args:
- Goal section (from PROJECT, GOAL)
- Abstract Constraints table (system-level non-negotiable requirements)
- Overview diagram
- Containers table with Boundary, Responsibilities, and Goal Contribution columns

### 1.2 Container Docs

For EACH container in inventory:

**1.2.1 Create container**

Create the container using the CLI:

```bash
npx -y c3-kit add container <slug>
```

This auto-numbers (c3-N), creates the directory, and generates `README.md` from the container template. After creation, **Edit** `.c3/c3-N-{slug}/README.md` to fill in:
- Goal section
- Responsibilities (what this container owns to satisfy context constraints)
- Complexity Assessment (level + why)
- Components table with Goal Contribution column

**1.2.2 Create component docs**

For each component in this container, create using the CLI:

```bash
# Foundation component (NN = 01-09):
npx -y c3-kit add component <slug> --container c3-N

# Feature component (NN = 10+):
npx -y c3-kit add component <slug> --container c3-N --feature
```

This auto-numbers the component (c3-NNN) and generates the file from the component template. After creation, **Edit** `.c3/c3-N-{slug}/c3-NNN-{component}.md` to fill in:
- Goal section
- Container Connection
- Code References section (REQUIRED — list concrete file paths, classes, or modules that implement this component)
- Related Refs table

**1.2.3 Extract Refs During Component Documentation**

While documenting components, proactively identify content that belongs in refs.

Load `references/onboard-ref-extraction.md` for the separation test, signals, and common extractions.

**Quick test:** "Would this content change if we swapped the underlying technology?"
- **Yes** → Extract to ref
- **No** → Keep in component

**Common extractions:** error handling, auth patterns, database usage, API conventions, state management.

**1.2.4 Handle discoveries**

If new component found during documentation:
- Add to ADR-000 Component Discovery table
- Document conflict in Conflict Resolution table
- Continue (will verify at Gate 2)

### 1.3 Ref Docs

For each ref in inventory, create using the CLI:

```bash
npx -y c3-kit add ref <slug>
```

This creates `.c3/refs/ref-{slug}.md` from the ref template. After creation, **Edit** the file to fill in:
- Goal section
- Choice (what option was chosen and context)
- Why (rationale over alternatives)
- How (implementation guidance for this codebase)
- Scope (where it applies and doesn't)
- Not every ref needs all sections, but Choice and Why are required
- Update Cited By as you create component docs

### Gate 1

Before proceeding, verify:
- [ ] All container README.md created
- [ ] All component docs created
- [ ] All refs documented
- [ ] No new items discovered (else → update ADR-000 inventory)

**If new items discovered:** Return to Stage 0, update ADR-000 discovery tables.

---

## Stage 2: Finalize

**Goal:** Verify integrity across all docs.

### 2.1 Integrity Checks

First, run structural validation using the CLI:

```bash
npx -y c3-kit check
```

This detects broken links, orphans, and duplicate IDs automatically. Fix any structural issues reported before proceeding to semantic checks.

Then verify semantic integrity:

| Check | How to Verify |
|-------|---------------|
| Context ↔ Container | Every container in ADR-000 appears in c3-0 (`.c3/README.md`) Containers table |
| Container ↔ Component | Every component in container README has a doc |
| Component ↔ Component | Linkages documented in Container README |
| * ↔ Refs | Refs Cited By section matches component Related Refs |

### 2.2 Run Audit

Invoke the **c3-audit skill** to validate integrity of all created docs. Note: audit Phase 10 (CLAUDE.md propagation) will produce warnings on fresh onboard — this is expected and can be resolved later via c3-change.

### Gate 2

Before marking complete, verify:
- [ ] All integrity checks pass
- [ ] audit passes

**If issues:**
- Inventory issues (missing container/component) → Gate 0
- Detail issues (wrong Goal, missing connection) → Gate 1
- Pass → Mark ADR-000 as `implemented`

---

## FINAL CHECKLIST - Verify Before Completing

**Before completing, verify structure exists:**

```bash
npx -y c3-kit list                    # Visual topology with goals
npx -y c3-kit check                   # Structural validation (broken links, orphans)
```

**If refs/ is empty:** Most projects have at least one cross-cutting chosen option (error handling, auth, API conventions). If discovery genuinely found none, that's fine — don't create synthetic refs.

**If any file is missing, create it before completing!**
