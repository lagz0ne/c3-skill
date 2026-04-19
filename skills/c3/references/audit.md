# Audit Reference

Validate C3 docs: consistency, drift, completeness.

Three tiers: **structural** (CLI) → **inventory** (CLI) → **semantic** (reasoning).

## Progress

- [ ] Phase 0: Structural (`c3x check`)
- [ ] Phase 1: Inventory (`c3x list`)
- [ ] Phase 2: Inventory vs Code
- [ ] Phase 3: Component Categorization
- [ ] Phase 4: Code Map Validation
- [ ] Phase 5: Diagram Accuracy
- [ ] Phase 6: ADR Lifecycle
- [ ] Phase 7: Ref & Rule Validation
- [ ] Phase 7b: Ref Compliance
- [ ] Phase 8: Abstraction Boundaries
- [ ] Phase 9: Content Separation
- [ ] Phase 10: CLAUDE.md

---

## Phase 0: Structural

```bash
bash <skill-dir>/bin/c3x.sh check
```
Detects: broken links, orphans, dup IDs, missing parents. Overlaps Phases 2,4,7 — skip re-check.

## Phase 1: Inventory

```bash
bash <skill-dir>/bin/c3x.sh list
```
Source of truth for all subsequent phases. No manual Glob+Read of `.c3/`.

**Topology Graphs:** Per container, run `c3x graph <container-id> --format mermaid` — include as mermaid blocks. Visual baseline for audit; subsequent phases reference these.

## Phase 2: Inventory vs Code

Compare Containers table ↔ actual dirs. Flag drift.
Per Container: Components inventory ↔ actual modules. Major module missing → FAIL.

## Phase 3: Component Categorization

Foundation (01-09): "Would changing this break many others?"
Feature (10+): "Specific to what product DOES?"
Wrong category → WARN.

## Phase 4: Code Map Validation

Per Component: `c3x lookup <file>` per mapped path — verify resolution, load constraint chain.
- Symbol: grep definition, flag if missing
- Pattern: glob, flag if zero matches
- Path: check exists, flag if missing
- Report: valid / stale / broken

Coverage:
```bash
bash <skill-dir>/bin/c3x.sh coverage
```
Low coverage → WARN. Formula: `mapped / (total - excluded)` — `_exclude` patterns don't penalize score. Suggest `_exclude` for test/config files, map remaining.

## Phase 5: Diagram Accuracy

All diagram IDs → verify exist in inventory. Stale reference → FAIL.

## Phase 6: ADR Lifecycle (--include-adr only)

ADRs = ephemeral work orders, hidden from default `c3x` ops.
Only audit when explicitly requested or with `c3x check --include-adr`.

`status=accepted` + >30 days without `implemented` → WARN.

## Phase 7: Ref & Rule Validation

- Ref: requires Choice + Why sections
- Ref: cited by ≥1 component (orphan → WARN)
- Citing component: ref entity exists in store (verify via `c3x list`)
- Rule: requires Rule + Golden Example sections
- Rule: cited by ≥1 component (orphan → WARN)
- Citing component: rule entity exists in store (verify via `c3x list`)

## Phase 7b: Ref Compliance

Per ref with `## How` containing golden patterns:
1. Find citing components via `c3x list`
2. Per citing component, spot-check 1-2 mapped files
3. Compare code against `## How` pattern

| Result | Meaning |
|--------|---------|
| COMPLIANT | Matches golden pattern structure |
| DRIFT | Diverges from pattern (may be intentional) |
| NOT CHECKED | No code-map mapping or no `## How` |

**Quality check:** Per ref `## How`, can you derive 1-3 YES/NO compliance questions?
- Yes → pattern actionable
- No → WARN: `## How` needs rework (too vague)

**Rule Compliance:** Per rule with `## Golden Example`:
1. Load rule:
   ```bash
   bash <skill-dir>/bin/c3x.sh read <rule-id>
   ```
   Extract `## Rule`, `## Golden Example`, `## Not This`.
2. Derive 1-3 YES/NO questions from `## Rule` + `## Golden Example` (e.g., "Does error return use CmdError struct?" / "Is slog used with component context?"). Can't derive → WARN: rule too vague.
3. Find citing components via `c3x list`
4. Per citing component, spot-check 1-2 mapped files
5. Apply YES/NO questions to spot-checked code

| Result | Meaning |
|--------|---------|
| COMPLIANT | All questions YES |
| VIOLATION | ≥1 question NO |
| INCOMPLETE | No Golden Example or can't derive questions |
| NOT CHECKED | No code-map mapping or no citing components |

Rules = STRICT enforcement (must match golden pattern exactly). Ref = directional alignment. Rule VIOLATION = always FAIL severity — non-negotiable constraints.

## Phase 8: Abstraction Boundaries

| Signal | Check | Violation | Severity |
|--------|-------|-----------|----------|
| Cross-container imports | Grep imports from other c3-* | Container bleeding | WARN |
| Global config definition | Grep exported constants used 3+ files | Context bleeding | WARN |
| Multi-component orchestration | Orchestrating vs handing off | Container job | FAIL |
| Pattern redefinition | Compare to cited refs | Ref bypass | FAIL |

## Phase 9: Content Separation

Code-map test:
- Component WITH code-map → implemented (Foundation/Feature)
- Component WITHOUT code-map → provisioned or misclassified
- Ref WITH code-map file patterns → VIOLATION (scaffold stubs OK)
- Ref with code examples in body → VALID
- Rule WITH code-map file patterns → VALID (rules govern code)
- Rule WITHOUT Golden Example → WARN (incomplete)

Missing refs: scan deps for tech used in 3+ components. Does ref explain "how we use it HERE"?

| Signal | Indicates | Action |
|--------|-----------|--------|
| "We use X for..." | Tech usage pattern | Extract to ref |
| "Our convention is..." | Cross-cutting pattern | Extract to ref |
| Same pattern in 2+ components | Duplicated knowledge | Create ref |

Missing rules: scan for patterns enforced in 3+ components without rule doc. Pattern has single correct form (not just preference)?

| Signal | Indicates | Action |
|--------|-----------|--------|
| Identical boilerplate in 3+ components | Enforced pattern without rule | Create rule with Golden Example |
| PR reviews citing same pattern | Implicit rule | Extract to rule |
| Linter/CI check without C3 rule doc | External enforcement gap | Create rule linking to tooling |

## Phase 10: CLAUDE.md

1. Extract expected dirs from code-map entries
2. Check CLAUDE.md exists per directory
3. Check `<!-- c3-generated: c3-NNN -->` matches expected component
4. Check orphan blocks referencing deleted components

Expected block:
```markdown
<!-- c3-generated: c3-201 -->
# c3-201: Component Title

Before modifying this code, run: c3x read c3-201
Patterns: ref-error-handling, ref-logging (run: c3x read ref-error-handling)
<!-- end-c3-generated -->
```

---

## Output

```
**C3 Audit Results**

| Phase | Status | Issues |
|-------|--------|--------|
| Structural | PASS/WARN/FAIL | [details] |
| ... | ... | ... |

**Summary:** N passes, M warnings, K failures
**Action Items:** [fixes]
```

---

## Drift Resolution

| Situation | Cause | Action |
|-----------|-------|--------|
| Code changed, docs outdated | Undocumented change | Create ADR, update docs |
| Docs describe removed code | Rot | Remove stale sections |
| New module not in inventory | Recent addition | Add to inventory |
| Orphan ADR (accepted, never implemented) | Abandoned | Close with reason |

Intentional arch change → ADR. Doc rot → direct fix.

---

## Audit Scope

| Scope | Focus | Phases |
|-------|-------|--------|
| Full | All layers | All |
| Single container | Container + components | 2-9 scoped |
| ADR-specific | ADR + affected | 6 + affected |
