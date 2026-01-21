# Consolidate Validation: Merge Auditor into Synthesizer

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Move all validation checks before ADR generation so ADR is always valid when presented. Remove redundant auditor phase.

**Architecture:** Expand c3-impact with boundary checks, expand c3-patterns with ref checks, consolidate validation in c3-synthesizer. Remove c3-adr-auditor.

**Tech Stack:** Markdown agent definitions

---

## Summary of Changes

| Component | Change |
|-----------|--------|
| c3-impact | Add 6 boundary checks |
| c3-patterns | Add 4 ref checks |
| c3-synthesizer | Add validation table, context check |
| c3-orchestrator | Remove Phase 5a, update flow |
| c3-adr-auditor | Remove entirely |

---

## Task 1: Expand c3-impact with Boundary Checks

**Files:**
- Modify: `agents/c3-impact.md`

**Add to output format:**

```markdown
## Boundary Analysis

| Check | Status | Evidence |
|-------|--------|----------|
| Ownership | ✓/✗ | Component owns this change per Responsibilities section |
| Redundancy | ✓/✗ | No duplicate capability exists elsewhere |
| Sibling overlap | ✓/✗ | No sibling already handles this |
| Composition | ✓/✗ | Hand-off pattern, not orchestration |
| Leaky abstraction | ✓/✗ | Internals not exposed in interface |
| Correct layer | ✓/✗ | Logic at appropriate level (context/container/component) |

### Issues Found
- [Issue with evidence if any check fails]
```

**Add to process section:**

```markdown
### Step 5: Boundary Analysis

For each affected component:

1. **Ownership:** Read component's "Responsibilities" or "Owns" section. Does the change fall within stated responsibilities?

2. **Redundancy:** Check sibling components in same container. Does any sibling already provide this capability?

3. **Sibling overlap:** Compare change description against all sibling responsibilities. Flag if >30% overlap.

4. **Composition:** Check for orchestration signals in change:
   - FAIL: "coordinate", "orchestrate", "manage flow", "control", "decide which to call"
   - PASS: "receives from", "passes to", "returns to", "hands off"

5. **Leaky abstraction:** Check if change exposes internal implementation details in the interface.

6. **Correct layer:**
   - Context-level: system-wide decisions, external boundaries
   - Container-level: coordination between components, composition rules
   - Component-level: single responsibility implementation
```

**Commit message:** `feat(impact): add boundary analysis checks`

---

## Task 2: Expand c3-patterns with Ref Checks

**Files:**
- Modify: `agents/c3-patterns.md`

**Update output format:**

```markdown
## Ref Compliance

| Check | Status | Evidence |
|-------|--------|----------|
| Follows ref | ✓/✗ | Change aligns with ref-X conventions |
| Ref usage correct | ✓/✗ | Ref applied as documented |
| Missing ref | ✓/✗ | No new reusable pattern without ref |
| Stale ref | ✓/✗ | Referenced ref still matches codebase |

### Issues Found
- [Issue with evidence if any check fails]
```

**Add to process section:**

```markdown
### Step 5: Ref Health Checks

1. **Follows ref:** Does the change follow the conventions documented in the relevant ref?

2. **Ref usage correct:** Is the ref being applied as intended, or misused?

3. **Missing ref:** Does this change introduce a reusable pattern that should become a ref?
   - Signals: pattern used in 2+ places, cross-cutting concern, convention others should follow

4. **Stale ref:** Is the referenced ref still accurate?
   - Check ref's examples against current codebase
   - Flag if ref documents patterns no longer in use
```

**Commit message:** `feat(patterns): add ref health checks`

---

## Task 3: Expand c3-synthesizer with Validation Hub

**Files:**
- Modify: `agents/c3-synthesizer.md`

**Replace output format with:**

```markdown
## Comprehensive Picture

### What You're Actually Changing
[2-3 sentences - the real scope]

### True Complexity
**Level:** [low|medium|high|critical]
**Hidden factors:** [what wasn't obvious]

### Key Decision Points
1. [Decision with options]
2. [Decision with options]

---

## Validation Status

### Context Alignment
**Checked:** .c3/README.md Key Decisions
**Contradictions:** [none | list]
**Override required:** [yes/no]

### Validation Summary

| Principle | Source | Status | Notes |
|-----------|--------|--------|-------|
| Ownership | c3-impact | ✓/✗ | |
| Redundancy | c3-impact | ✓/✗ | |
| Sibling overlap | c3-impact | ✓/✗ | |
| Composition | c3-impact | ✓/✗ | |
| Leaky abstraction | c3-impact | ✓/✗ | |
| Correct layer | c3-impact | ✓/✗ | |
| Follows ref | c3-patterns | ✓/✗ | |
| Ref usage | c3-patterns | ✓/✗ | |
| Missing ref | c3-patterns | ✓/✗ | |
| Stale ref | c3-patterns | ✓/✗ | |
| Context alignment | self | ✓/✗ | |

### Validation Outcome
**Ready for ADR:** [yes | no]
**Blockers:** [list any ✗ that must resolve before ADR]

---

## Suggested Verification Criteria
- [ ] [Criterion 1]
- [ ] [Criterion 2]

## Open Questions
[Any ✗ validation becomes an open question here]
[Or "None - ready for ADR" if all ✓]
```

**Add to process section:**

```markdown
### Step 5: Context Alignment Check

Read `.c3/README.md` Key Decisions section. Compare each key decision against the proposed change:
- Does the change contradict any key decision?
- If yes, is there justification for override?

### Step 6: Build Validation Table

Collect validation results from c3-impact and c3-patterns outputs:
1. Parse their "Boundary Analysis" and "Ref Compliance" tables
2. Add context alignment from Step 5
3. Any ✗ status = blocker for ADR

### Step 7: Determine Readiness

**Ready for ADR** only if:
- All validation checks are ✓
- OR all ✗ items have documented override justification

**Not ready** if:
- Any ✗ without justification
- Convert each ✗ to an Open Question for user resolution
```

**Commit message:** `feat(synthesizer): add validation hub with all checks`

---

## Task 4: Update c3-orchestrator Flow

**Files:**
- Modify: `agents/c3-orchestrator.md`

**Changes:**

1. **Remove Phase 5a (Audit)** - Delete the entire section

2. **Update workflow diagram** - Remove auditor box and loop

3. **Update Phase 4 (Refinement)** to check validation:

```markdown
## Phase 4: Socratic Refinement

Review synthesizer output.

**If validation blockers exist (any ✗):**
- Surface each blocker to user
- Use `AskUserQuestion` to resolve:
  - "Rethink scope" → return to Phase 1
  - "Justify override" → add to ADR's Pattern Overrides section
  - "Fix the issue" → return to Phase 1 with narrower scope

**If "Ready for ADR: yes":**
- Confirm understanding with user
- Proceed to Phase 5

**Loop until synthesizer outputs "Ready for ADR: yes"**
```

4. **Update Phase 5** - Remove audit step reference:

```markdown
## Phase 5: Generate ADR

**Precondition:** Synthesizer output shows "Ready for ADR: yes"

Create ADR at `.c3/adr/adr-YYYYMMDD-{slug}.md`...

[ADR is guaranteed valid - no audit needed]
```

5. **Remove Phase 5a entirely**

6. **Update anti-patterns table** - Remove auditor references

**Commit message:** `refactor(orchestrator): remove audit phase, validate before ADR`

---

## Task 5: Remove c3-adr-auditor

**Files:**
- Delete: `agents/c3-adr-auditor.md`

**Also update:**
- Remove from any orchestrator references
- Remove from plugin.json if explicitly listed (auto-discovery handles it)

**Commit message:** `refactor(agents): remove c3-adr-auditor (consolidated into synthesizer)`

---

## Task 6: Update Documentation

**Files:**
- Modify: `references/adr-template.md` - Remove audit references if any

**Commit message:** `docs: update references for consolidated validation`

---

## Verification

After implementation:

1. **Test the flow:**
   - Create a change that violates "ownership" → should loop at synthesizer
   - Create a change with stale ref → should loop at synthesizer
   - Create a valid change → should go straight to ADR

2. **Verify auditor removed:**
   - `agents/c3-adr-auditor.md` should not exist
   - Orchestrator should not reference Phase 5a

3. **Verify validation coverage:**
   - All 11 checks present in synthesizer output
   - Each check traceable to source (impact or patterns)

---

## Diagram

```
Phase 1: Clarify ◄─────────────────────────────┐
     │                                         │
     ▼                                         │
Phase 2: Parallel Analysis                     │
     │                                         │
     ├──► c3-impact (+ boundary checks)        │
     ├──► c3-patterns (+ ref checks)           │
     └──► c3-analyzer                          │
               │                               │
               ▼                               │
Phase 3: Synthesizer (VALIDATION HUB)          │
     │                                         │
     ▼                                         │
╔════════════════════════════════════╗         │
║   All 11 validations pass?         ║         │
╠════════════════════════════════════╣         │
║  NO ──► Surface to user ───────────╫─────────┘
║         (loop back)                ║
║                                    ║
║  YES ──► Phase 5: Generate ADR     ║
║          (always valid)            ║
╚════════════════════════════════════╝
     │
     ▼
Phase 5b: Accept (+ base-commit)
     │
     ▼
Phase 6: Delegate
     │
     ▼
Phase 7: Complete
```
