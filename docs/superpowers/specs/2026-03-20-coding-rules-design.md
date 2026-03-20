# Coding Rules: First-Class Entity Type for C3

## Problem

C3 refs currently serve two distinct purposes:
1. **Architectural decisions** — "we chose X over Y because Z" (e.g., ref-cross-compiled-binary)
2. **Enforceable coding standards** — "we do X, period" (e.g., logging conventions, error handling)

These have different templates, different enforcement postures, and different lifecycles. With code-map already mapping files to components, enforceable coding standards are a natural growth — but they need their own entity type.

## Design

### New Entity Type: `rule`

Rules are first-class entities in the C3 topology, alongside components, refs, ADRs, and recipes.

| Aspect | Refs (narrowed) | Rules (new) |
|--------|-----------------|-------------|
| Identity | "We chose X over Y because Z" | "We do X. Period." |
| Template | Choice + Why + How + Not This | Rule + Golden Example + Not This |
| Enforceability | Soft — structural | Hard — auditable at review time |
| Value center | Rationale | Enforcement |
| Code-map section | `# Refs` | `# Rules` |
| Directory | `.c3/refs/ref-*.md` | `.c3/rules/rule-*.md` |

### Separation Test (Updated)

The current test ("Is this a choice we made?") cannot distinguish refs from rules — both answer "yes." Updated heuristic:

> **Remove the Why section. Does the doc become useless?**
> - Yes — it's a **ref** (the value is in the rationale)
> - No — it's a **rule** (the value is in enforcement)

Concrete examples:
- "We chose cross-compiled binaries over runtime downloads" — remove Why and you lose the doc's purpose → **ref**
- "All API responses use `{ data, error, meta }` envelope" — remove Why, developers still know exactly what to do → **rule**
- "We use structured logging with pino, not console.log" — remove Why, the standard is still clear → **rule**

### Rule Template

```markdown
---
id: rule-${SLUG}
type: rule
c3-version: 4
title: ${TITLE}
goal: ${GOAL}
origin: []
scope: []
---

# ${TITLE}

## Goal

${GOAL}

## Rule

One-line statement of what must be true.

## Golden Example

Canonical code showing the correct pattern.
(Multiple code blocks for multi-context rules.)

## Not This

Anti-patterns — what's wrong and what to do instead.

| Anti-Pattern | Correct | Why Wrong Here |
|-------------|---------|----------------|
| ... | ... | ... |

## Scope

**Applies to:**
- ...

**Does NOT apply to:**
- ...

## Override

To override this rule:
1. Document justification in an ADR
2. Cite this rule and explain why the override is necessary
3. Specify the scope of the override (which components deviate)
```

### Key Template Decisions

1. **No `## Why`** — rules don't justify themselves. They are the law for this codebase. Provenance is tracked via `origin:` frontmatter field.

2. **`origin:` frontmatter field** — links to the ref or ADR that established the rule. Provides provenance trail without polluting the rule body. Example: `origin: [ref-naming-conventions]` or `origin: [adr-20260315-adopt-eslint]`.

3. **`## Override` included** — more authoritarian = more need for a documented escape hatch. Same 3-step process as refs.

4. **`## Golden Example` not `## How`** — "How" implies guidance; "Golden Example" implies the canonical standard. Multiple code blocks are fine within the section for multi-context rules.

5. **`type: rule` explicit in frontmatter** — required for CLI classification. Matches pattern used by other entity types.

### Code-Map Integration

Rules get their own section in `code-map.yaml`:

```yaml
# Components
c3-101:
  - cli/internal/frontmatter/**

# Refs
ref-cross-compiled-binary:
  - scripts/build.sh

# Rules
rule-structured-logging:
  - src/**/*.ts
  - src/**/*.tsx
```

Rules map to source files via globs, same as refs. The agent discovers applicable rules during `c3x lookup` and checks compliance at review time.

### Component Citation

Rules share the existing `uses:` frontmatter mechanism on components:

```yaml
---
id: c3-101
uses: [ref-frontmatter-docs, rule-frontmatter-format]
---
```

The CLI distinguishes rules from refs by ID prefix (`rule-` vs `ref-`). This avoids adding a separate `rules:` frontmatter field while keeping the plumbing simple. Governance metrics can filter by prefix.

### Dual-Nature Concerns

Some patterns are both a decision and a standard. Example: `ref-frontmatter-docs` is both "we chose YAML frontmatter" (ref) and "every file must have these fields" (rule).

Resolution: split into two artifacts linked by `origin:`.
- `ref-frontmatter-docs` → the architectural choice (why YAML frontmatter)
- `rule-frontmatter-format` → the enforceable standard (what fields, what format), with `origin: [ref-frontmatter-docs]`

This is expected to be rare (1 of 3 existing refs qualifies).

## CLI Changes

### New DocType

```go
// frontmatter.go — add to iota block
DocRule  // after existing DocRecipe

// Frontmatter struct — add field
Origin []string `yaml:"origin,omitempty" json:"origin,omitempty"`

// ClassifyDoc — insert BEFORE the ref- prefix check (line ~171)
if fm.Type == "rule" { return DocRule }
if strings.HasPrefix(fm.ID, "rule-") { return DocRule }
```

Note: `type: rule` check must precede the `ref-` prefix fallback in the if-chain, since `ClassifyDoc` returns on first match.

### Files Requiring Changes

| File | Change |
|------|--------|
| `cli/internal/frontmatter/frontmatter.go` | Add `DocRule` constant (iota), `Origin []string` field, classification branch before `ref-` check, add `fm.Origin...` to `DeriveRelationships()` so origin links are full graph edges (enables reverse traversal for delete cleanup) |
| `cli/internal/walker/walker.go` | Add `rule-` to `slugPattern` regex; add `DocRule` to `Forward()` cited-by traversal (line ~220) |
| `cli/internal/schema/schema.go` | Add `"rule"` to Registry with sections: Goal (req), Rule (req), Golden Example (req), Not This (opt), Scope (opt), Override (opt). Also add `"Related Rules"` section to the `"component"` schema entry (mirrors `"Related Refs"`) |
| `cli/internal/templates/rule.md` | New template (embedded via go:embed) |
| `cli/cmd/add.go` | Add `case "rule"` calling `addSubdirEntity` for `.c3/rules/` |
| `cli/cmd/help.go` | Update entity types hint text |
| `cli/internal/codemap/validate.go` | Add `"rule"` to accepted types (consider refactoring the negative condition into an allowlist set as types grow) |
| `cli/cmd/codemap.go` | Collect `DocRule` entities + write `# Rules` section (add `rules` param to `writeCodeMap`) |
| `cli/internal/index/index.go` | Add `Rules []string` to `FileEntry`, `RuleEntry` to `StructuralIndex`; filter `rule-` prefix IDs into `fe.Rules` during `Build()` |
| `cli/cmd/lookup.go` | Add `Rules []RefBrief` to `LookupMatch`; in `buildMatch()`, split `Frontmatter.Refs` by ID prefix — `rule-*` → Rules, `ref-*` → Refs |
| `cli/cmd/list.go` | Display rules in topology output |
| `cli/cmd/graph.go` | Add `DocRule` cases at lines ~189 (cited-by display), ~246 (JSON output), ~320 (mermaid node rendering) |
| `cli/cmd/coverage.go` | Add rule governance metric; in `index.go RefGovernance()`, separate rule vs ref governance counts by prefix |
| `cli/cmd/check_enhanced.go` | Rules validated via schema Registry (automatic once registered); add `origin:` existence check — verify each origin ref/ADR ID exists in graph |
| `cli/cmd/wire.go` | Detect target type by ID prefix; use `## Related Rules` section for rules, `## Related Refs` for refs |
| `cli/cmd/delete.go` | Include rules in cross-reference cleanup |
| `cli/cmd/init.go` | Scaffold `.c3/rules/` directory alongside `.c3/refs/` and `.c3/adr/` |

### Implementation Order (dependency chain)

1. `frontmatter.go` — DocRule enum + Origin field + ClassifyDoc (gate for everything)
2. `walker.go` — slugPattern regex + Forward() traversal
3. `schema.go` — Registry entry
4. `cli/internal/templates/rule.md` — embedded template
5. `add.go` + `help.go` + `init.go` — scaffolding
6. `validate.go` — codemap type check
7. `codemap.go` — scaffold generation
8. `index.go` — structural index + rule governance
9. `lookup.go` — differentiated output with prefix-based splitting
10. `list.go` + `graph.go` — topology display + graph rendering
11. `coverage.go` — governance metrics
12. `check_enhanced.go` — validation + origin existence check
13. `wire.go` + `delete.go` — citation + cleanup

## Skill Reference Changes

### New: `references/rule.md`

The ref operation reference (`references/ref.md`) handles ref lifecycle. A parallel `references/rule.md` handles rule lifecycle with adapted flows.

#### Mode Selection

| Intent | Mode |
|--------|------|
| "add/create/document a coding rule" | **Add** |
| "update/modify rule-X" | **Update** |
| "list rules", "what rules exist" | **List** |
| "who uses rule-X" | **Usage** |
| "remove/deprecate rule-X" | **change** (needs ADR) |

#### Add Flow

`Scaffold → Discover → Extract Golden Pattern → Confirm → Fill Content → Discover Usage → Update Citings → ADR`

**HARD RULE: First Bash call must be scaffold.**

1. **Scaffold**: `c3x add rule <slug>`
2. **Discover** (2-5 Grep calls): Search for existing implementations of the pattern.
   - 0 files → user describes the standard
   - 1 file → extract, flag for confirmation
   - 2+ files → structural intersection of top 3 = golden pattern
3. **Extract Golden Pattern**: From discovered code:
   - Shared structure → `## Golden Example`
   - Varies by context → multiple code blocks with context labels
   - Clearly wrong → `## Not This`
4. **Confirm**: `AskUserQuestion` — present extracted pattern for approval
5. **Quality Gate**: Write 1-3 YES/NO compliance questions derivable from `## Rule` + `## Golden Example`. If you can't write them, the standard is too vague.
6. **Fill Content**: Rule, Golden Example, Not This, Scope, Override
7. **Discover Usage** (2-3 Grep calls): Find components using this pattern
8. **Update Citings**: Add to `## Related Rules` in citing components
9. **Adoption ADR**: `status: implemented` — rule doc IS the deliverable

#### Update Flow

Same as ref Update, but compliance checking uses `## Golden Example` and `## Not This` for strict comparison (not directional alignment).

#### List/Usage

Same as ref List/Usage, filtered to `type: "rule"` entities.

### Updated: Separation Test in ref template

The embedded ref template (`cli/internal/templates/ref.md`) needs the updated Separation Test:

```
THE SEPARATION TEST:
"Remove the Why section. Does the doc become useless?"
- Yes → Belongs in ref (the value is in the rationale)
- No → Belongs in rule (the value is in enforcement)
- Neither → Belongs in component (business/domain logic)
```

### Updated: Change/Audit References

`references/change.md` Phase 3b (Ref Compliance Gate) and `references/audit.md` Phase 7b (Ref Compliance) need to process both refs and rules, with appropriate enforcement posture:
- Refs: check directional alignment with `## How`
- Rules: check strict compliance with `## Golden Example` and `## Not This`

## Migration

### Existing Refs

| Current | Action |
|---------|--------|
| `ref-cross-compiled-binary` | Stays as ref — pure architectural decision |
| `ref-embedded-templates` | Stays as ref — pure architectural decision |
| `ref-frontmatter-docs` | Split: keep ref for choice rationale, create `rule-frontmatter-format` for enforcement |

### For Other Projects Using C3

No breaking changes. Rules are additive. Projects without `.c3/rules/` continue to work. The CLI gracefully handles the absence of rules.

## Governance Semantics

`c3x coverage` reports rule governance separately from ref governance:

```json
{
  "ref_governance": { "governed": 5, "total": 8, "percentage": 62.5 },
  "rule_governance": { "governed": 3, "total": 8, "percentage": 37.5 }
}
```

A component is "rule-governed" if it cites at least one `rule-*` entity via `uses:`. This is independent of ref governance — a component can be ref-governed but not rule-governed, or both. The `RefGovernance()` function in `index.go` splits by ID prefix.

## Testing

- `c3x add rule test-rule` → scaffolds `.c3/rules/rule-test-rule.md` with correct template
- `c3x init` → creates `.c3/rules/` directory
- `c3x list --json` → rules appear with `type: "rule"` in topology
- `c3x check` → validates rule doc structure (required sections: Goal, Rule, Golden Example); validates `origin:` IDs exist in graph
- `c3x lookup <file>` → returns rules in `rules` field, refs in `uses` field (separated by prefix)
- `c3x codemap` → generates `# Rules` section; rule entries survive re-scaffold
- Code-map entries with `rule-*` IDs pass validation (no warning)
- `c3x coverage` → reports rule governance separately from ref governance
- `c3x graph` → rules render as distinct node type in mermaid output
- `c3x wire c3-101 rule-foo` → adds row to `## Related Rules` (not `## Related Refs`)
