# C3 Skill Harness

Shared behavioral constraints for all C3 skills.

## Complexity-First Documentation

### Rule: COMPLEXITY-BEFORE-ASPECTS

Assess container complexity BEFORE documenting aspects. Never assume aspects from templates.

**Complexity Levels:**

| Level | Signals | Documentation Depth |
|-------|---------|---------------------|
| trivial | Single purpose, stateless, no deps | Purpose + deps only |
| simple | Few concerns, basic state | + key components |
| moderate | Multiple concerns, caching, auth | + discovered aspects |
| complex | Orchestration, security-critical | Full aspect discovery |
| critical | Distributed txns, compliance | Comprehensive + rationale |

**Required in every container doc:**
```markdown
## Complexity Assessment
**Level:** [trivial|simple|moderate|complex|critical]
**Why:** [signals observed from code analysis]
```

### Rule: DISCOVERY-OVER-CHECKLIST

Aspects MUST be discovered through code analysis, not assumed from templates.

| Wrong | Right |
|-------|-------|
| Tick "caching: N/A" without checking | Scan for cache patterns, document if found |
| Copy aspect list from template | Analyze code, document what exists |
| "Assumed standard" | "Discovered: retry logic in api-client.ts:45" |

### Rule: DEPTH-MATCHES-COMPLEXITY

| Level | Aspect Documentation |
|-------|---------------------|
| trivial/simple | Skip aspects section entirely |
| moderate | Document 2-3 key discovered aspects |
| complex | Full discovery with code references |
| critical | + rationale for each aspect |

---

## Workflow Complexity

```
trivial ─────────────────────────────────────────── critical
│                                                        │
/onboard        /query         /alter add      /alter refactor
(init)          (navigate)     (single ADR)    (cross-layer ADR)
```

Match workflow depth to container complexity.

## STOP - Before Any Action

| Check | Action |
|-------|--------|
| No `.c3/README.md`? | Route to `c3` skill (Adopt mode) |
| Wrong intent? | Route to correct skill (see routing table) |
| Uncertain about user intent? | Use `AskUserQuestion` until clear |

## Skill Routing

| User Intent | Skill | Example |
|-------------|-------|---------|
| Explore / Explain / Find | `c3-query` | "where is auth?" |
| Add / Change / Implement | `c3-alter` | "add payment service" |
| Audit / Validate | `c3` | "check C3 docs" |
| Adopt / Onboard / Init | `c3` | "set up C3" |

## Red Flags - STOP Immediately

| Violation | Why Wrong | Correct Action |
|-----------|-----------|----------------|
| Editing code without ADR | Changes need reasoning trail | Create ADR first, then execute |
| Guessing user intent | Ambiguity causes wrong changes | Ask with `AskUserQuestion` |
| Jumping to component | Miss context, dependencies | Start from Context (c3-0) down |
| Updating docs without code check | Docs may be stale | Verify code matches before updating |

### Violation Examples

**Wrong:** "User says 'fix the login' → immediately edit auth code"

**Right:** "User says 'fix the login' → Stage 1: clarify what's broken → Stage 2: check current c3-auth docs → Stage 3: create ADR → Stage 4: execute with plan"

---

**Wrong:** "User asks 'where is payment?' → search codebase with grep"

**Right:** "User asks 'where is payment?' → read c3-0 → find container → read component docs → THEN explore code using References section"

## Required Tool: AskUserQuestion

All clarification MUST use `AskUserQuestion` tool:
- Structured answers reduce ambiguity
- Multiple-choice when options are clear
- Continue until no open questions

**When to ask:**
- Intent unclear (feature vs fix?)
- Scope ambiguous (which component?)
- Impact uncertain (breaking change?)

## Layer Navigation

Always load `layer-navigation.md` before traversing C3 docs.

Traversal order: **Context (c3-0) → Container (c3-N) → Component (c3-NNN)**

Never skip layers. Context provides container relationships. Container provides component inventory.
