# Onboard Reference

## Precondition

`c3x list` returns entities → already onboarded. `AskUserQuestion`: re-onboard or cancel (skip if ASSUMPTION_MODE). Cancel → suggest audit/query.

## File Structure

```
.c3/
├── README.md                    # Context (c3-0)
├── adr/adr-00000000-c3-adoption.md
├── refs/ref-<pattern>.md
├── rules/rule-<pattern>.md
└── c3-N-<container>/
    ├── README.md
    └── c3-NNN-<component>.md
```

Each component = separate file. Each container = separate directory.

## Component Categories

| Can name concrete file? | Category |
|------------------------|---------|
| Yes | Foundation (01-09) or Feature (10+) |
| No (rules only) | **Ref** — code-map entry optional |

Foundation: infrastructure others depend on. Feature: business logic. Ref: conventions or shared utilities. Rule: coding standards and constraints that must be followed. Refs with concrete implementation files (shared middleware, utility libraries) should have code-map entries; pure-convention refs and rules may leave them empty.

## Progress Checklist

```
- [ ] Stage 0: inventory complete, ADR-000 tables filled
- [ ] Gate 0: proceed to Details
- [ ] Stage 1: all container/component/ref docs created
- [ ] Gate 1: no new items discovered
- [ ] Stage 2: code-map scaffolded + patterns filled, integrity + audit pass
- [ ] Gate 2: ADR-000 marked implemented
```

---

## Stage 0: Inventory

### 0.1 Scaffold

```bash
bash <skill-dir>/bin/c3x.sh init
```
Creates `.c3/` with config, README, refs/, rules/, adr/. Update ADR-000 via `c3x write` to fill discovery tables.

### 0.2 Context Discovery

Capture in ADR-000:

| Arg | Value |
|-----|-------|
| PROJECT | System name |
| GOAL | Why it exists |
| SUMMARY | One sentence |

Also find **Abstract Constraints** — system-level non-negotiables.

Use `AskUserQuestion` for gaps (ASSUMPTION_MODE: assume, mark `[ASSUMED]`).

### 0.3 Container Discovery

Container = deployment/runtime boundary. Capture:

| N | CONTAINER_NAME | BOUNDARY | GOAL | SUMMARY |
|---|----------------|----------|------|---------|

### 0.4 Component Discovery

| N | NN | COMPONENT_NAME | CATEGORY | GOAL | SUMMARY |
|---|----|----|----------|------|---------|

Foundation (01-09): others depend on it. Feature (10+): business logic.

### 0.5 Ref Discovery

Patterns repeating across components:

| SLUG | TITLE | GOAL | Scope | Applies To |
|------|-------|------|-------|------------|

Common: error handling, form patterns, data fetching, design system. Each ref requires Choice + Why minimum.

### 0.6 Rule Discovery

Coding standards and constraints that must be followed project-wide:

| SLUG | TITLE | GOAL | Scope | Applies To |
|------|-------|------|-------|------------|

Common: naming conventions, forbidden patterns, required lint rules, security constraints. Look for repeated code-review feedback, linter configs, and "we always/never do X" statements.

### 0.7 Overview Diagram

Mermaid: Actors → Containers → External Systems.

### Gate 0

- [ ] Context args filled (PROJECT, GOAL, SUMMARY)
- [ ] Abstract Constraints identified
- [ ] All containers with args (including BOUNDARY)
- [ ] All components (brief) with category
- [ ] Cross-cutting refs (Choice + Why minimum)
- [ ] Coding standards as rules
- [ ] Overview diagram

---

## Stage 1: Details

### 1.1 Context Doc

Update c3-0 via `c3x set c3-0 --section "Goal" <text>` and `c3x write c3-0 < content.md` for full body.

### 1.2 Container Docs

**Create container:**
```bash
bash <skill-dir>/bin/c3x.sh add container <slug>
```
Fill via `c3x write <id>`: Goal, Responsibilities, Complexity, Components table.

**Create components:**
```bash
# Foundation (01-09):
bash <skill-dir>/bin/c3x.sh add component <slug> --container c3-N
# Feature (10+):
bash <skill-dir>/bin/c3x.sh add component <slug> --container c3-N --feature
```
Fill via `c3x write <id>`: Goal, Container Connection, Related Refs table. Add code-map via `c3x codemap`.
Bracket paths (`[id]`, `[...slug]`) for Next.js/SvelteKit routes work automatically in code-map patterns.

**Extract Refs:** "Would this change if we swapped the underlying tech?" Yes → extract to ref.

**Extract Rules:** "Is this a coding standard or constraint rather than a pattern choice?" Yes → extract to rule.

| Signal | Action |
|--------|--------|
| "We use X with..." | ref-X |
| "Our convention is..." | new/existing ref |
| Same pattern in 2+ components | create ref, cite both |
| "We always/never do X" | rule |
| Lint rule, naming convention, security constraint | rule |

### 1.3 Ref Docs

```bash
bash <skill-dir>/bin/c3x.sh add ref <slug>
```
Fill via `c3x write <id>`: Goal, Choice (required), Why (required), How/Scope/Not This/Override as needed.

### 1.4 Rule Docs

```bash
bash <skill-dir>/bin/c3x.sh add rule <slug>
```
Fill via `c3x write <id>`: Goal, Constraint (required), Why (required), Scope/Exceptions as needed.

### Gate 1

- [ ] All container README.md created
- [ ] All component docs created
- [ ] All refs documented
- [ ] All rules documented
- [ ] No new items (else update ADR-000, return to Stage 0)

---

## Stage 2: Finalize

### 2.1 Code-Map Scaffold

```bash
bash <skill-dir>/bin/c3x.sh codemap
```

Scaffolds `.c3/code-map.yaml` with empty stubs for every component, ref, and rule.
Idempotent — safe to re-run; existing patterns are preserved.

After scaffolding, fill in glob patterns for each entry, then verify:
```bash
bash <skill-dir>/bin/c3x.sh coverage          # how many files are mapped
bash <skill-dir>/bin/c3x.sh lookup 'src/**'   # spot-check the mapping
```

### 2.2 Structural

```bash
bash <skill-dir>/bin/c3x.sh check
```

### 2.3 Semantic

| Check | Verify |
|-------|--------|
| Context ↔ Container | ADR-000 containers match README.md |
| Container ↔ Component | Each component in container README has doc |
| * ↔ Refs | Citations match Related Refs |

### 2.4 Audit

Run audit operation. Pass → mark ADR-000 `implemented`.

### Gate 2

- [ ] Code-map scaffolded and patterns filled
- [ ] Coverage % acceptable (or exclusions documented)
- [ ] Integrity checks pass
- [ ] Audit passes

Issues → Inventory (Gate 0) or Detail (Gate 1).

---

## Final Checks

```bash
bash <skill-dir>/bin/c3x.sh codemap                    # scaffold/update code-map.yaml stubs
bash <skill-dir>/bin/c3x.sh list
bash <skill-dir>/bin/c3x.sh check
bash <skill-dir>/bin/c3x.sh lookup <any-mapped-file>   # spot-check single file
bash <skill-dir>/bin/c3x.sh lookup 'src/**'            # check entire source tree
bash <skill-dir>/bin/c3x.sh coverage                   # code-map coverage gaps
```

**Fix before completing:**

| Signal | Problem | Fix |
|--------|---------|-----|
| No system goal | Missing `goal:` in README.md | `c3x set <id> <field> <value>` |
| No `files:` | Missing code-map stubs | Run `c3x codemap`, then fill in patterns |
| No `uses:` | Ref not wired | `c3x wire <component> <ref>` |
| Ref has no `via:` | Uncited ref | Wire or delete |
| `[provisioning]` | Design-only | Expected or implement |
| `lookup <file>` returns nothing | No codemap or bad glob | Run `c3x codemap`; fix patterns; try `lookup 'src/**'` to see what IS mapped |
| Low coverage % | Many unmapped files | Add `_exclude` for tests/configs, map remaining to components |

---

## Post-Onboard

### CLAUDE.md Injection

```markdown
# Architecture
This project uses C3 docs in `.c3/`.
For architecture questions, changes, audits, file context -> `/c3`.
Operations: query, audit, change, ref, rule, sweep.
File lookup: `c3x lookup <file-or-glob>` maps files/directories to components + refs.
```

### Capabilities Reveal

Show the user the typical workflow, then point to self-discovery:

```
## Your C3 toolkit is ready

**Typical flow:**

1. Understand what exists: `c3x list` → topology, then `c3x lookup <file>` → which component owns it
2. Make changes: `c3x add` / `c3x set` / `c3x wire` to create and connect entities
3. Validate: `c3x check` catches broken links, schema gaps, orphans
4. Explore impact: `c3x graph <id>` shows what connects to what

For architecture questions, changes, audits → just say `/c3` + what you want.

Run `c3x capabilities` to see all available commands.
Run `c3x <command> --help` for detailed usage.
```

## Complexity Guide

| Level | Signals | Aspect Doc |
|-------|---------|------------|
| trivial/simple | Single purpose | Skip aspects |
| moderate | Multiple concerns | 2-3 key aspects |
| complex | Orchestration | Full discovery + code-map |
| critical | Distributed/compliance | + rationale each |

Discover aspects from code, don't assume from templates.
