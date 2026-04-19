# Onboard Reference

## Precondition

`c3x list` returns entities → already onboarded. `AskUserQuestion`: re-onboard or cancel (skip if ASSUMPTION_MODE). Cancel → suggest audit/query.

## Component Categories

| Can name concrete file? | Category |
|------------------------|---------|
| Yes | Foundation (01-09) or Feature (10+) |
| No (rules only) | **Ref** — code-map entry optional |

Foundation: infra others depend on. Feature: biz logic. Ref: conventions/shared utils. Rule: coding standards/constraints. Refs with concrete files (shared middleware, util libs) → code-map entries; pure-convention refs and rules → empty.

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
Creates `.c3/` with config, README, refs/, rules/, adr/. Update ADR-000 via `c3x write`.

### 0.2 Context Discovery

Capture in ADR-000:

| Arg | Value |
|-----|-------|
| PROJECT | System name |
| GOAL | Why it exists |
| SUMMARY | One sentence |

Find **Abstract Constraints** — system-level non-negotiables.

`AskUserQuestion` for gaps (ASSUMPTION_MODE: assume, mark `[ASSUMED]`).

### 0.3 Container Discovery

Container = deployment/runtime boundary.

| N | CONTAINER_NAME | BOUNDARY | GOAL | SUMMARY |
|---|----------------|----------|------|---------|

### 0.4 Component Discovery

| N | NN | COMPONENT_NAME | CATEGORY | GOAL | SUMMARY |
|---|----|----|----------|------|---------|

Foundation (01-09): others depend on. Feature (10+): biz logic.

### 0.5 Ref Discovery

Patterns repeating across components:

| SLUG | TITLE | GOAL | Scope | Applies To |
|------|-------|------|-------|------------|

Common: error handling, form patterns, data fetching, design system. Each ref requires Choice + Why minimum.

### 0.6 Rule Discovery

Project-wide coding standards/constraints:

| SLUG | TITLE | GOAL | Scope | Applies To |
|------|-------|------|-------|------------|

Common: naming conventions, forbidden patterns, lint rules, security constraints. Look for repeated review feedback, linter configs, "always/never" statements.

### 0.7 Overview Diagram

Per container:
```bash
bash <skill-dir>/bin/c3x.sh graph <container-id> --format mermaid
```
Include each as mermaid code block.

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

Update c3-0 via `c3x set c3-0 --section "Goal" <text>` and `c3x write c3-0 < content.md`.

### 1.2 Container Docs

**Create container** (body via stdin — atomic):
```bash
cat <<'EOF' | bash <skill-dir>/bin/c3x.sh add container <slug>## Goal
<goal description>

## Components
| ID | Name | Goal |
|---|---|---|
| <id> | <name> | <goal> |

## Responsibilities
- <responsibility>
EOF
```

**Create components** (body via stdin):
```bash
# Foundation (01-09):
cat <<'EOF' | bash <skill-dir>/bin/c3x.sh add component <slug> --container c3-N## Goal
<goal>

## Dependencies
| Target | Why |
|--------|-----|
| <target> | <reason> |
EOF

# Feature (10+): add --feature flag
cat <<'EOF' | bash <skill-dir>/bin/c3x.sh add component <slug> --container c3-N --feature## Goal
<goal>

## Dependencies
| Target | Why |
|--------|-----|
| <target> | <reason> |
EOF
```
Code-map via `c3x codemap`. Bracket paths (`[id]`, `[...slug]`) work automatically.

**Extract Refs:** "Would this change if we swapped underlying tech?" Yes → ref.

**Extract Rules:** "Coding standard or constraint, not pattern choice?" Yes → rule.

| Signal | Action |
|--------|--------|
| "We use X with..." | ref-X |
| "Our convention is..." | new/existing ref |
| Same pattern in 2+ components | create ref, cite both |
| "We always/never do X" | rule |
| Lint rule, naming, security | rule |

### 1.3 Ref Docs

```bash
cat <<'EOF' | bash <skill-dir>/bin/c3x.sh add ref <slug>## Goal
<goal>

## Choice
<choice>

## Why
<rationale>
EOF
```
Optional: How, Scope, Not This, Override.

### 1.4 Rule Docs

```bash
cat <<'EOF' | bash <skill-dir>/bin/c3x.sh add rule <slug>## Goal
<goal>

## Rule
<rule description>

## Golden Example
<example code or pattern>
EOF
```
Optional: Not This, Scope, Override.

### Gate 1

- [ ] All container README.md created
- [ ] All component docs created
- [ ] All refs documented
- [ ] All rules documented
- [ ] No new items (else update ADR-000, return Stage 0)

---

## Stage 2: Finalize

### 2.1 Code-Map Scaffold

```bash
bash <skill-dir>/bin/c3x.sh codemap
```

Scaffolds code-map entries for every component, ref, rule. Idempotent — existing patterns preserved.

Fill glob patterns, then verify:
```bash
bash <skill-dir>/bin/c3x.sh coverage          # file coverage
bash <skill-dir>/bin/c3x.sh lookup 'src/**'   # spot-check mapping
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

Run audit. Pass → mark ADR-000 `implemented`.

### Gate 2

- [ ] Code-map scaffolded + patterns filled
- [ ] Coverage % acceptable (or exclusions documented)
- [ ] Integrity checks pass
- [ ] Audit passes

Issues → Inventory (Gate 0) or Detail (Gate 1).

---

## Final Checks

```bash
bash <skill-dir>/bin/c3x.sh codemap
bash <skill-dir>/bin/c3x.sh list
bash <skill-dir>/bin/c3x.sh check
bash <skill-dir>/bin/c3x.sh lookup <any-mapped-file>
bash <skill-dir>/bin/c3x.sh lookup 'src/**'
bash <skill-dir>/bin/c3x.sh coverage
```

**Fix before completing:**

| Signal | Problem | Fix |
|--------|---------|-----|
| No system goal | Missing `goal:` in README.md | `c3x set <id> <field> <value>` |
| No `files:` | Missing code-map stubs | `c3x codemap`, fill patterns |
| No `uses:` | Ref not wired | `c3x wire <component> <ref>` |
| Ref has no `via:` | Uncited ref | Wire or delete |
| `[provisioning]` | Design-only | Expected or implement |
| `lookup` returns nothing | Bad/missing codemap | `c3x codemap`; fix patterns; `lookup 'src/**'` |
| Low coverage % | Many unmapped files | `_exclude` for tests/configs, map rest |

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

```
## Your C3 toolkit is ready

**Typical flow:**

1. Understand what exists: `c3x list` → topology, then `c3x lookup <file>` → which component owns it
2. Make changes: `c3x add` / `c3x set` / `c3x wire` to create and connect entities
3. Validate: `c3x check` catches broken links, schema gaps, orphans
4. Visualize: `c3x graph <container-or-component> --format mermaid` renders architecture as mermaid diagrams

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

Discover aspects from code, never assume from templates.
