# Rule + Graph Injection via Skill Instructions

**Date:** 2026-04-07
**Status:** Proposed
**Scope:** Skill instruction changes only (SKILL.md, change.md, audit.md)

## Problem

C3's codemap achieves 100% file coverage, and `c3x lookup` already returns rules wired to components. But the lookup output only gives `{id, goal}` for rules — a pointer, not the constraint. The LLM sees that a rule exists but doesn't absorb its content unless it follows up with `c3x read`.

The skill instructions don't enforce this follow-up. Result: the infrastructure for rule-aware code editing exists, but the LLM doesn't naturally load rule constraints before writing code.

## Solution

Enhance three skill instruction files to make rule reading and graph injection automatic after every `c3x lookup`:

```
lookup <file>  →  read <each-rule>  →  graph <component> --depth 1
```

No CLI changes. No schema changes. Instruction-only enhancement.

## Changes

### 1. SKILL.md — File Context section

Expand the mandatory file context block from:

```
c3x lookup <file-path>
→ Returned refs = hard constraints
```

To:

```
c3x lookup <file-path>
→ For each rule-* returned: c3x read <rule-id>
→ For each component returned: c3x graph <component-id> --depth 1 (first match if multiple)
→ Returned refs + loaded rule content = hard constraints
```

This applies universally to ALL operations that touch files, not just change.

### 2. change.md — Phase 3 (Execute)

After the existing "REQUIRED before touching any file" lookup block, add:

**Rule loading:**
- For every `rule-*` in the lookup response, run `c3x read <rule-id>`
- Extract `## Rule`, `## Golden Example`, and `## Not This` sections
- Hold these constraints in working context while editing
- Code MUST match the golden pattern; deviations require an Override section in the rule or a new ADR

**Graph context:**
- Run `c3x graph <component-id> --depth 1` for the first matched component (or each if multiple)
- Shows what depends on this component (consumers) and what it depends on (providers)
- If changes affect the component's interface, check consumers before proceeding

### 3. change.md — Phase 3b (Ref Compliance Gate)

Minor enhancement: connect the verification to the injection:
- "Compare against the rule content loaded in Phase 3" (not re-read)
- Reinforce that the golden example and not-this patterns are already in context

### 4. audit.md — Phase 7b (Ref Compliance)

Add at the start of rule compliance checking:
- `c3x read <rule-id>` for each rule before compliance checking
- Derive 1-3 YES/NO compliance questions from `## Rule` + `## Golden Example`
- If questions can't be derived, WARN: rule is too vague for enforcement

## What does NOT change

- c3x CLI code (Go source)
- c3x lookup output format
- Rule/ref entity schemas
- Codemap structure
- Any other skill reference files (onboard.md, query.md, ref.md, rule.md, sweep.md, migrate.md)

## Flow with enhancement

```
User: "add rate limiting to the API gateway"
  │
  ├─ c3 change: ADR → topology → Phase 3
  │
  ├─ c3x lookup cli/cmd/gateway.go
  │   → component: c3-150 (api-gateway)
  │   → rules: [rule-error-handling, rule-structured-logging]
  │
  ├─ c3x read rule-error-handling
  │   → ## Rule: All commands return structured CmdError
  │   → ## Golden Example: return &CmdError{Code: ..., Msg: ...}
  │   → ## Not This: fmt.Errorf("failed: %w", err)
  │
  ├─ c3x read rule-structured-logging
  │   → ## Rule: Use slog with component context
  │   → ## Golden Example: slog.Info("op completed", "component", id)
  │
  ├─ c3x graph c3-150 --depth 1
  │   → depends on: c3-106 (store), c3-101 (frontmatter)
  │   → used by: c3-2 (skill container)
  │
  ├─ LLM writes code with constraints + blast radius in context
  │
  └─ Phase 3b: verify compliance against loaded rules
```

## Validation

After implementing the instruction changes:
1. Run `/c3 audit` on this project — should pass structural checks
2. Create a test rule (e.g., `rule-error-handling`) and wire it to a component
3. Invoke `/c3 change` to modify a file governed by that rule
4. Verify the LLM loads the rule content before editing and checks compliance after
