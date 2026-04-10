# Audit Reference

Validate C3 docs for consistency, drift, completeness.

Three tiers: **structural** (CLI) → **inventory** (CLI) → **semantic** (reasoning).

## Progress

- [ ] Phase 0: Structural (`c3x check`)
- [ ] Phase 1: Inventory (`c3x list --json`)
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
bash <skill-dir>/bin/c3x.sh check --json
```
Detects: broken links, orphans, duplicate IDs, missing parents. Issues here overlap Phases 2, 4, 7 — skip re-checking those.

## Phase 1: Inventory

```bash
bash <skill-dir>/bin/c3x.sh list --json
```
Source of truth for all subsequent phases. No manual Glob+Read of `.c3/`.

**Topology Graphs:** For each container in the inventory, run `c3x graph <container-id> --format mermaid` and include as mermaid code blocks. These establish the visual baseline for the audit — subsequent phases reference these graphs.

## Phase 2: Inventory vs Code

Context: Compare Containers table ↔ actual directories. Flag drift.
Each Container: Components inventory ↔ actual modules. Major module missing → FAIL.

## Phase 3: Component Categorization

Foundation (01-09): "Would changing this break many others?"
Feature (10+): "Is this specific to what this product DOES?"
Wrong category → WARN.

## Phase 4: Code Map Validation

For each Component: `c3x lookup <file>` per mapped path — verifies resolution, loads constraint chain.
- Symbol: grep for definition, flag if not found
- Pattern: glob, flag if zero matches
- Path: check exists, flag if missing
- Report: valid / stale / broken

Coverage check:
```bash
bash <skill-dir>/bin/c3x.sh coverage
```
Low coverage → WARN. Formula: `mapped / (total - excluded)` — `_exclude` patterns don't penalize the score. Suggest `_exclude` for test/config files, map remaining to components.

## Phase 5: Diagram Accuracy

All IDs in diagrams → verify exist in inventory. Stale reference → FAIL.

## Phase 6: ADR Lifecycle (--include-adr only)

ADRs are ephemeral work orders, hidden from default `c3x` operations.
Only audit ADR lifecycle when explicitly requested or when running `c3x check --include-adr`.

`status=accepted` + >30 days without `implemented` → WARN.

## Phase 7: Ref & Rule Validation

- Each ref: requires Choice + Why sections
- Each ref: cited by at least one component (orphan → WARN)
- Each citing component: ref entity exists in store (verify via `c3x list --json`)
- Each rule: requires Rule + Golden Example sections
- Each rule: cited by at least one component (orphan → WARN)
- Each citing component: rule entity exists in store (verify via `c3x list --json`)

## Phase 7b: Ref Compliance

For each ref with `## How` containing golden patterns:
1. Find citing components via `c3x list --json`
2. For each citing component, spot-check 1-2 mapped files from code-map
3. Compare code against `## How` pattern

| Result | Meaning |
|--------|---------|
| COMPLIANT | Code matches golden pattern structure |
| DRIFT | Code diverges from pattern (may be intentional) |
| NOT CHECKED | No code-map mapping or no `## How` section |

**Quality check:** For each ref `## How`, can you derive 1-3 YES/NO compliance questions?
- Yes → pattern is actionable
- No → WARN: `## How` needs rework (too vague for enforcement)

**Rule Compliance:** For each rule with `## Golden Example`:
1. **Load rule content:**
   ```bash
   bash <skill-dir>/bin/c3x.sh read <rule-id>
   ```
   Extract `## Rule`, `## Golden Example`, and `## Not This` sections.
2. **Derive compliance questions:** From `## Rule` + `## Golden Example`, write 1-3 YES/NO questions that code must satisfy (e.g., "Does the error return use CmdError struct?" / "Is slog used with component context?").
   - If questions can't be derived → WARN: rule is too vague for enforcement, needs rework.
3. Find citing components via `c3x list --json`
4. For each citing component, spot-check 1-2 mapped files from code-map
5. Apply the YES/NO questions to the spot-checked code

| Result | Meaning |
|--------|---------|
| COMPLIANT | All compliance questions answered YES |
| VIOLATION | One or more questions answered NO |
| INCOMPLETE | Rule lacks Golden Example or questions can't be derived |
| NOT CHECKED | No code-map mapping or no citing components |

Rules use STRICT enforcement (code must match golden pattern exactly), unlike refs which use directional alignment. A rule VIOLATION is always FAIL severity — rules are non-negotiable constraints.

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
- Rule WITHOUT Golden Example → WARN (rule is incomplete)

Missing refs: scan deps for tech used in 3+ components. Does ref explain "how we use it HERE"?

| Signal | Indicates | Action |
|--------|-----------|--------|
| "We use X for..." | Tech usage pattern | Extract to ref |
| "Our convention is..." | Cross-cutting pattern | Extract to ref |
| Same pattern in 2+ components | Duplicated knowledge | Create ref |

Missing rules: scan for patterns enforced in 3+ components without a rule doc. Does the pattern have a single correct form (not just a preference)?

| Signal | Indicates | Action |
|--------|-----------|--------|
| Identical boilerplate in 3+ components | Enforced pattern without rule | Create rule with Golden Example |
| PR reviews citing same pattern repeatedly | Implicit rule | Extract to rule |
| Linter/CI check without C3 rule doc | External enforcement gap | Create rule linking to tooling |

## Phase 10: CLAUDE.md

1. Extract expected dirs from code-map entries
2. Check CLAUDE.md exists in each directory
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
