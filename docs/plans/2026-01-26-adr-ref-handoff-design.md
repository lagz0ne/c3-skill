# ADR Reference Handoff Design

**Date:** 2026-01-26
**Status:** Draft
**Problem:** Components lack embedded ref constraints, causing implementation drift

---

## Problem Statement

Current workflow has a handoff gap:

1. **Refs exist** as cross-cutting patterns in `.c3/refs/`
2. **Components** don't explicitly declare which refs they must follow
3. **ADRs** don't surface applicable refs for changes
4. **No enforcement** at commit time

**Result:** Code "runs wild" without architectural guardrails. Implementers don't know what constraints apply.

---

## Goals

| Goal | Measure |
|------|---------|
| Components declare their refs | 100% of components have `## References` section |
| ADRs surface applicable refs | Inherited from affected components + discovered for new ones |
| Hard enforcement | c3-gate blocks code that violates component refs |

---

## Core Principle: Component as Source of Truth

**Refs are component-level constraints, not ADR-level.**

The ADR is transient (captures a change). The component is permanent. Refs must be embedded in component documentation.

```
┌─────────────────────────────────┐
│  Component Doc (permanent)      │
│  c3-201-auth-middleware.md      │
│                                 │
│  ## References                  │  ← Source of truth
│  - ref-error-handling           │
│  - ref-auth                     │
│  - ref-middleware-chain         │
│                                 │
│  ## Code References             │
│  - src/middleware/auth.ts       │
└─────────────────────────────────┘
              │
              │ Any change to c3-201
              ▼
┌─────────────────────────────────┐
│  ADR (transient)                │
│  adr-20260126-fix-timeout.md    │
│                                 │
│  ## References to Follow        │  ← Inherited from c3-201
│  - ref-error-handling           │
│  - ref-auth                     │
│  - ref-middleware-chain         │
│                                 │
│  ## Approved Files              │
│  - src/middleware/auth.ts       │
└─────────────────────────────────┘
              │
              │ On commit
              ▼
┌─────────────────────────────────┐
│  c3-gate (enforcement)          │
│                                 │
│  1. Find affected component     │
│  2. Read ## References          │
│  3. Validate code follows refs  │
│  4. Block if violated           │
└─────────────────────────────────┘
```

---

## Design

### 1. Component Documentation Enhancement

Every component doc MUST have a `## References` section:

```markdown
## References

Patterns this component must follow:

| Ref | Applies To |
|-----|------------|
| ref-error-handling | Error responses in auth failures |
| ref-middleware-chain | Request/response pipeline |
```

**Placement:** After `## Responsibilities`, before `## Code References`

**Rules:**
- Required for all components (Foundation and Feature)
- Empty section is valid (no refs apply) but must be explicit
- Refs must exist in `.c3/refs/`

---

### 2. c3-analysis Enhancement

Add Part 4 to analysis output: **Reference Discovery**

**For existing components:**
```markdown
## Part 4: Required References

### From Component Docs
| Component | Refs |
|-----------|------|
| c3-201 | ref-error-handling, ref-auth |
| c3-203 | ref-error-handling |

### Combined (deduplicated)
- ref-error-handling (cited by c3-201, c3-203)
- ref-auth (cited by c3-201)
```

**For new components:**
```markdown
## Part 4: Required References

### Discovered for New Component
| Ref | Source | Why |
|-----|--------|-----|
| ref-error-handling | Keyword match | "error" in component responsibility |
| ref-middleware-chain | Container pattern | c3-2-api uses middleware pattern |

### Recommended for c3-206
- ref-error-handling
- ref-middleware-chain
```

---

### 3. ADR Template Enhancement

Add `## References to Follow` section:

```markdown
## References to Follow

Refs inherited from affected components (enforced by c3-gate):

| Ref | Source |
|-----|--------|
| ref-error-handling | c3-201, c3-203 |
| ref-auth | c3-201 |

**For new components (c3-206):**
| Ref | Reason |
|-----|--------|
| ref-middleware-chain | Container pattern |
```

Add to frontmatter:
```yaml
references-to-follow:
  - ref-error-handling
  - ref-auth
  - ref-middleware-chain
```

---

### 4. c3-gate Hook Enhancement

**Current:** Validates approved-files list only

**New:** Also validates ref compliance

```bash
# Pseudo-code for c3-gate

# 1. Find accepted ADR
adr=$(find_accepted_adr)

# 2. Get affected components from ADR
components=$(extract_affects "$adr")

# 3. For each component, get its refs
for component in $components; do
  refs=$(extract_refs "$component")

  # 4. For each changed file in that component
  for file in $(get_component_files "$component"); do
    for ref in $refs; do
      # 5. Validate compliance
      if ! check_compliance "$file" "$ref"; then
        echo "BLOCKED: $file violates $ref (required by $component)"
        exit 1
      fi
    done
  done
done
```

---

### 5. Workflow: Modify Existing Component

```
User: "Fix timeout in auth middleware"

1. c3-analysis reads c3-201
   → Extracts: ref-error-handling, ref-auth, ref-middleware-chain

2. ADR created with:
   ## References to Follow
   | Ref | Source |
   | ref-error-handling | c3-201 |
   | ref-auth | c3-201 |
   | ref-middleware-chain | c3-201 |

3. Implementation must honor all three refs

4. c3-gate on commit:
   - Reads c3-201's ## References
   - Validates changed code
   - Blocks if any ref violated
```

---

### 6. Workflow: Create New Component

```
User: "Add rate limiting middleware"

1. c3-analysis discovers refs:
   - ref-middleware-chain (from c3-2-api container pattern)
   - ref-error-handling (rate limit errors need standard format)

2. ADR created with:
   ## References to Follow
   | Ref | Reason |
   | ref-middleware-chain | Container pattern |
   | ref-error-handling | Error responses |

3. New component doc (c3-206-rate-limiter.md) created with:
   ## References
   | Ref | Applies To |
   | ref-middleware-chain | Middleware pipeline integration |
   | ref-error-handling | Rate limit error responses |

4. c3-gate enforces these refs for c3-206 going forward
```

---

## Changes Required

| File | Change |
|------|--------|
| `references/component-types.md` | Add `## References` as required section |
| `templates/component.md` | Add `## References` section template |
| `agents/c3-analysis.md` | Add Part 4: Reference Discovery |
| `references/adr-template.md` | Add `## References to Follow` section |
| `agents/c3-orchestrator.md` | Phase 5: Include refs from analysis |
| `skills/c3-alter/SKILL.md` | Stage 4: Require refs in ADR |
| `scripts/c3-gate` | Add ref compliance validation |
| `skills/onboard/SKILL.md` | Ensure new components get refs |

---

## Example: Component with References

```markdown
---
id: c3-201
title: Auth Middleware
container: c3-2-api
---

# Auth Middleware

## Responsibilities

- Validate JWT tokens on protected routes
- Attach user context to request
- Handle auth failures gracefully

## References

Patterns this component must follow:

| Ref | Applies To |
|-----|------------|
| ref-error-handling | Auth failure responses (401, 403) |
| ref-auth | Token validation, session handling |
| ref-middleware-chain | Request pipeline integration |

## Code References

- `src/middleware/auth.ts:15-89` - Main middleware function
- `src/middleware/auth.ts:91-120` - Token validation helper

## Hand-offs

| From | Data | To |
|------|------|-----|
| HTTP Router | Raw request | This |
| This | Authenticated request | Route handlers |
```

---

## Enforcement Depth

**Recommended approach: Checklist + Audit**

| Stage | Check | Depth |
|-------|-------|-------|
| c3-gate (commit) | Refs acknowledged | Presence check (fast) |
| c3-dev (implementation) | TDD follows refs | Manual during development |
| /c3 audit (verification) | Full compliance | Deep check (thorough) |

**Why not full static analysis at commit?**
- Refs are semantic (hard to automate)
- False positives block legitimate work
- Audit catches issues before ADR completion

---

## Migration Path

For existing C3 projects without `## References` in components:

1. Run `/c3 audit` to identify components without refs
2. For each component:
   - Identify applicable refs from code patterns
   - Add `## References` section
3. Create migration ADR documenting the additions

---

## Ref Creation: When and How

**Principle: Ref-first is preferred.**

If a pattern will be reused, create the ref BEFORE implementing. This ensures consistency from the start.

### Valid Paths

| Path | When to Use | Flow |
|------|-------------|------|
| **Ref-first (preferred)** | Pattern will be reused across components | `/c3-ref` → Create ref → Then `/c3-alter` → Component cites ref |
| **During change** | Discover pattern mid-implementation | c3-alter pauses → `/c3-ref` → Resume with ref |
| **Retroactive** | Pattern already exists inconsistently | `/c3-ref` → Create ref → Update existing components |

### Ref-First Workflow

```
User: "Add rate limiting to API"

c3-analysis discovers:
  - No ref-rate-limiting exists
  - Pattern will likely be reused (multiple endpoints)

c3-orchestrator asks:
  "Rate limiting looks like a reusable pattern. Create ref first?"
  - "Yes - create ref-rate-limiting first" (preferred)
  - "No - implement without ref (one-off)"

If yes:
  1. Pause c3-alter
  2. Switch to /c3-ref → Create ref-rate-limiting
  3. Resume c3-alter → Component cites ref-rate-limiting
```

### During-Change Workflow

```
User: "Add rate limiting to API"
(User chose to proceed without ref)

Mid-implementation, realizes:
  "This retry logic should be standardized"

c3-alter pauses:
  1. Create ref-retry-pattern via /c3-ref
  2. Update current component to cite ref
  3. Resume implementation
```

### Signals for Ref Creation

c3-analysis should surface ref opportunities when:

| Signal | Interpretation |
|--------|----------------|
| Similar code in 2+ components | Pattern already emerging - standardize |
| User says "like we do in X" | Implicit pattern reference - make explicit |
| Cross-cutting concern (auth, logging, errors) | Almost always needs ref |
| "Should be consistent" language | User wants standardization |

### Ref Creation Gate

When creating a ref during c3-alter:

1. Ref MUST have its own mini-ADR (via c3-ref skill)
2. Ref MUST be created before component cites it
3. Component doc updated to cite new ref
4. Original c3-alter ADR references the ref ADR

---

## Open Questions

1. **Empty refs:** Should components with no applicable refs have empty section or omit it?
   - Recommendation: Explicit empty section (`## References\n\nNone.`)

2. **Ref inheritance:** Should container-level refs auto-apply to all components?
   - Recommendation: No, explicit per-component (avoids hidden constraints)

3. **Ref versioning:** What if a ref changes after component was created?
   - Recommendation: Ref changes trigger audit of citing components (via c3-ref skill)
