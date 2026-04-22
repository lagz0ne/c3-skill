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

Scaffold `.c3/` with config, README, refs/, rules/, adr/. Update ADR-000 via `c3x write <adr-000> --section <name> --file <path>` for any body with tables, mermaid, or code blocks; short single-sentence fields via `c3x set <adr-000> <field> <value>` or `echo "..." | c3x write <adr-000> --section <name>`.

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

Short text fields: `echo "<goal>" | c3x write c3-0 --section Goal`. Whole body rewrite: `c3x write c3-0 --file content.md`.

### 1.2 Container Docs

**Create container** (body in a file — tables and mermaid require `--file`):
```bash
# body.md contains: ## Goal / ## Components (table) / ## Responsibilities
bash <skill-dir>/bin/c3x.sh add container <slug> --file body.md
```

**Create components** (body in a file):
```bash
# Foundation (01-09):
bash <skill-dir>/bin/c3x.sh add component <slug> --container c3-N --file body.md

# Feature (10+): add --feature flag
bash <skill-dir>/bin/c3x.sh add component <slug> --container c3-N --feature --file body.md
```
Body should contain `## Goal` plus `## Dependencies` table. Any content with markdown tables, mermaid, or code fences MUST go through `--file <path>` — inline strings corrupt quoting.

Code-map patterns: `c3x set <id> codemap <pattern>`. Bracket paths (`[id]`, `[...slug]`) work automatically.

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
# body.md: ## Goal / ## Choice / ## Why (+ optional How, Scope, Not This, Override)
bash <skill-dir>/bin/c3x.sh add ref <slug> --file body.md
```

### 1.4 Rule Docs

```bash
# body.md: ## Goal / ## Rule / ## Golden Example (code fence) (+ optional Not This, Scope, Override)
bash <skill-dir>/bin/c3x.sh add rule <slug> --file body.md
```
Golden Example contains code fences -> `--file` is mandatory.

### Gate 1

- [ ] All container README.md created
- [ ] All component docs created
- [ ] All refs documented
- [ ] All rules documented
- [ ] No new items (else update ADR-000, return Stage 0)

---

## Stage 2: Finalize

### 2.1 Code-Map

Set glob patterns per component/ref/rule:
```bash
bash <skill-dir>/bin/c3x.sh set <id> codemap '<glob>'
bash <skill-dir>/bin/c3x.sh lookup 'src/**'   # spot-check mapping
```

### 2.2 Validate

```bash
bash <skill-dir>/bin/c3x.sh check
bash <skill-dir>/bin/c3x.sh list              # coverage + counts
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
bash <skill-dir>/bin/c3x.sh list
bash <skill-dir>/bin/c3x.sh check
bash <skill-dir>/bin/c3x.sh lookup <any-mapped-file>
bash <skill-dir>/bin/c3x.sh lookup 'src/**'
```

**Fix before completing:**

| Signal | Problem | Fix |
|--------|---------|-----|
| No system goal | Missing `goal:` in README.md | `c3x set <id> goal "<text>"` |
| No `files:` | Missing code-map pattern | `c3x set <id> codemap '<glob>'` |
| No `uses:` | Ref not wired | `c3x wire <component> <ref>` |
| Ref has no `via:` | Uncited ref | Wire or delete |
| `[provisioning]` | Design-only | Expected or implement |
| `lookup` returns nothing | Bad/missing codemap | Fix patterns via `c3x set <id> codemap '<glob>'`; re-check with `lookup 'src/**'` |
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

1. Understand what exists: `c3x list` → topology + coverage, then `c3x lookup <file>` → which component owns it
2. Make changes: `c3x add` / `c3x write` / `c3x set` / `c3x wire` to create and connect entities (use `--file <path>` for bodies with tables, mermaid, or code fences)
3. Validate: `c3x check` catches broken links, schema gaps, orphans
4. Visualize: `c3x graph <container-or-component> --format mermaid` renders architecture as mermaid diagrams

For architecture questions, changes, audits → just say `/c3` + what you want.

Run `c3x --help` to see all available commands.
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
