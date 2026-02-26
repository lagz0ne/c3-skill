# Onboard Reference

Initialize C3 architecture documentation using a 3-stage workflow with gates.

## Precondition: Fresh Project

If `.c3/README.md` already exists:
- `AskUserQuestion`: "This project already has C3 docs. What do you want to do?" (skip if ASSUMPTION_MODE)
  - "Re-onboard from scratch (overwrites existing docs)"
  - "Cancel (use audit to check existing docs)"
- On Cancel: stop, suggest audit or query
- On Re-onboard: proceed with overwrite warning

## File Structure

Each component = separate file. Each container = separate directory.

```
.c3/
├── README.md                    # System context
├── adr/
│   └── adr-00000000-c3-adoption.md
├── refs/                        # Cross-cutting patterns
│   └── ref-<pattern>.md
└── c3-N-<container>/            # Container directory
    ├── README.md                # Container overview
    └── c3-NNN-<component>.md    # One file per component
```

**Non-negotiable:**
- Each component = separate file (c3-101-auth.md, c3-102-users.md)
- Each container = separate directory (c3-1-api/, c3-2-web/)
- refs/ = separate directory with pattern files

## Component Categories

| Question | If Yes | Category |
|----------|--------|----------|
| Can you name a concrete code file? | Yes | **Foundation** (01-09) or **Feature** (10+) |
| Is it only rules/conventions? | Yes | **Ref** |

- **Foundation** (NN=01-09): infrastructure that others depend on. Has entry in `.c3/code-map.yaml`.
- **Feature** (NN=10+): business logic composing foundations. Has entry in `.c3/code-map.yaml`.
- **Ref**: conventions only. NO code-map entry. May include golden code examples.

Hard rules:
- If you cannot name a concrete file, you cannot create a component doc — create a ref
- Code-map entry = implemented. Refs never have one.

## Progress Checklist

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
- [ ] Gate 2: Audit passes
- [ ] ADR-000 marked implemented
```

---

## Stage 0: Inventory

**Goal:** Discover EVERYTHING before creating any docs.

### 0.1 Scaffold and Create ADR-000

```bash
bash <skill-dir>/bin/c3x.sh init
```

Creates `.c3/` with `config.yaml`, `README.md`, `refs/`, and `adr/adr-00000000-c3-adoption.md`.

After init, **Edit** the adoption ADR to fill discovery tables.

### 0.2 Context Discovery

Read codebase structure. Capture in ADR-000 Context Discovery table:

| Arg | Value |
|-----|-------|
| PROJECT | What is this system called? |
| GOAL | Why does this system exist? |
| SUMMARY | One-sentence description |

Also discover **Abstract Constraints** — system-level non-negotiable requirements:

| Constraint | Rationale | Affected Containers |
|------------|-----------|---------------------|

Use `AskUserQuestion` for gaps (skip if ASSUMPTION_MODE; assume and mark `[ASSUMED]`).

### 0.3 Container Discovery

A container = deployment/runtime boundary that allocates responsibilities.

For each container, capture in ADR-000:

| N | CONTAINER_NAME | BOUNDARY | GOAL | SUMMARY |
|---|----------------|----------|------|---------|

Use `AskUserQuestion` to confirm each container (skip if ASSUMPTION_MODE; assume from directory structure, mark `[ASSUMED]`).

### 0.4 Component Discovery (Brief)

For each container, scan for components:

| N | NN | COMPONENT_NAME | CATEGORY | GOAL | SUMMARY |
|---|----|--------------  |----------|------|---------|

- **foundation** (NN=01-09): others depend on it
- **feature** (NN=10+): business logic

Use `AskUserQuestion` to confirm categorization (skip if ASSUMPTION_MODE; categorize by import count, mark `[ASSUMED]`).

### 0.5 Ref Discovery

Look for patterns repeating across components:

| SLUG | TITLE | GOAL | Scope | Applies To |
|------|-------|------|-------|------------|

Common ref candidates: error handling, form patterns, query/data fetching, design system principles, user flows.

**Ref structure:** Each ref uses sections that serve its Goal: Choice (what we chose), Why (rationale), How (conventions/rules), Not This (anti-patterns), Scope (where it applies), Override (how to deviate). **Choice and Why are required minimum.**

### 0.6 Overview Diagram

Generate mermaid diagram: Actors -> Containers -> External Systems with key linkages.

### Gate 0

- [ ] Context args filled (PROJECT, GOAL, SUMMARY)
- [ ] Abstract Constraints identified
- [ ] All containers identified with args (including BOUNDARY)
- [ ] All components identified (brief) with args and category
- [ ] Cross-cutting refs identified (with Choice and Why minimum)
- [ ] Overview diagram generated

Gaps? Return to discovery with Socratic questioning.

---

## Stage 1: Details

**Goal:** Create all docs from inventory.

### 1.1 Context Doc

`.c3/README.md` created by `c3x init`. Edit to fill:
- Goal section (from PROJECT, GOAL)
- Abstract Constraints table
- Overview diagram
- Containers table with Boundary, Responsibilities, Goal Contribution

### 1.2 Container Docs

For EACH container:

**1.2.1 Create container:**
```bash
bash <skill-dir>/bin/c3x.sh add container <slug>
```
Auto-numbers (c3-N), creates directory + README.md. Edit to fill:
- Goal section
- Responsibilities
- Complexity Assessment (level + why)
- Components table with Goal Contribution

**1.2.2 Create components:**
```bash
# Foundation (NN=01-09):
bash <skill-dir>/bin/c3x.sh add component <slug> --container c3-N

# Feature (NN=10+):
bash <skill-dir>/bin/c3x.sh add component <slug> --container c3-N --feature
```
Auto-numbers (c3-NNN). Edit to fill:
- Goal section
- Container Connection
- Code map entry in `.c3/code-map.yaml` (REQUIRED — concrete file paths)
- Related Refs table

**1.2.3 Extract Refs During Component Documentation**

While documenting components, identify content for refs.

**Separation test:** "Would this content change if we swapped the underlying technology?"
- Yes -> extract to ref
- No -> keep in component

**Signals to extract:**

| Signal | Action |
|--------|--------|
| "We use [technology] with..." | Extract to ref-[technology] |
| "Our convention is..." | Extract to existing or new ref |
| "Always handle errors by..." | Extract to ref-error-handling |
| Same pattern in 2+ components | Create ref, cite in both |

**1.2.4 Handle discoveries:**
- New component found -> add to ADR-000, document conflict in Conflict Resolution table

### 1.3 Ref Docs

```bash
bash <skill-dir>/bin/c3x.sh add ref <slug>
```
Creates `.c3/refs/ref-{slug}.md`. Edit to fill:
- Goal, Choice (required), Why (required)
- How, Scope, Not This, Override (as relevant)
- Verify citing components listed in code-map.yaml as you create component docs

### Gate 1

- [ ] All container README.md created
- [ ] All component docs created
- [ ] All refs documented
- [ ] No new items discovered (else -> update ADR-000 inventory)

New items discovered? Return to Stage 0.

---

## Stage 2: Finalize

**Goal:** Verify integrity.

### 2.1 Structural Validation

```bash
bash <skill-dir>/bin/c3x.sh check
```
Detects broken links, orphans, duplicate IDs. Fix structural issues before semantic checks.

### 2.2 Semantic Integrity

| Check | How to Verify |
|-------|---------------|
| Context <-> Container | Every container in ADR-000 appears in README.md Containers table |
| Container <-> Component | Every component in container README has a doc |
| Component <-> Component | Linkages documented in Container README |
| * <-> Refs | Ref citations match component Related Refs |

### 2.3 Run Audit

Execute the audit operation to validate all created docs.

### Gate 2

- [ ] All integrity checks pass
- [ ] Audit passes

Issues? Inventory issues -> Gate 0. Detail issues -> Gate 1. Pass -> Mark ADR-000 `implemented`.

---

## Final Checklist

```bash
bash <skill-dir>/bin/c3x.sh list     # Visual topology with goals
bash <skill-dir>/bin/c3x.sh check    # Structural validation
```

If refs/ empty: most projects have at least one cross-cutting pattern. If discovery genuinely found none, fine.

If any file missing, create it before completing.

---

## Post-Onboard

1. Inject CLAUDE.md routing block (see SKILL.md)
2. Show capabilities reveal (see SKILL.md)

## Complexity-First Documentation

Assess container complexity BEFORE documenting aspects:

| Level | Signals | Aspect Documentation |
|-------|---------|---------------------|
| trivial/simple | Single purpose, few concerns | Skip aspects section |
| moderate | Multiple concerns, caching | 2-3 key discovered aspects |
| complex | Orchestration, security-critical | Full discovery with code-map |
| critical | Distributed txns, compliance | + rationale for each aspect |

**Discovery over checklist:** Aspects MUST be discovered through code analysis, not assumed from templates.
