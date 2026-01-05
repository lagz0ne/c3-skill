---
name: c3-alter
description: |
  Use when project has .c3/ and user intends to change code or architecture.
  Triggers: "add", "change", "modify", "update", "fix", "refactor", "remove", "implement", "create new", "rename", "new feature", "bug fix".
---

# C3 Alter - Change Through ADR

**Every change flows through an ADR.** No exceptions.

## NOT This Skill

| Intent | Use Instead |
|--------|-------------|
| Understand/explore | `c3-query` |
| Set up C3 | `c3` (Adopt) or `/onboard` |
| Validate docs | `c3` (Audit) |

## REQUIRED: Load Navigation Reference

**Before any work, read `../../references/layer-navigation.md`.**

## Critical: Always Use AskUserQuestionTool

All Socratic questioning MUST use the `AskUserQuestion` tool - never plain text questions.
- Structured questions get structured answers
- Multiple-choice reduces ambiguity
- User can select rather than type

## Core Pattern

Every stage uses the same recursive learning loop:

```
ANALYZE → ASK (until confident) → SYNTHESIZE → REVIEW
              │                        │
              └── until no open ───────┘
                  questions

On conflict during execution:
ASCEND → fix earlier stage → re-descend
```

**Confident = No open questions.** Don't proceed if any field would be "TODO" or unclear.

---

## Stage 1: Intent

### 1.1 Analyze

What is the user asking for? Form hypotheses:
- Add, modify, remove, fix, or refactor?
- What problem are they solving?
- What's the scope hint?

Build open questions list.

### 1.2 Ask (Socratic)

Use AskUserQuestion until no open questions about intent.

Example questions:
- "Is this a new feature or a fix to existing behavior?"
- "What problem does this solve?"
- "Is there urgency? Breaking change risk?"

### 1.3 Synthesize

Create clear intent statement:
```
**Intent:** Add OAuth login option to the authentication flow
**Goal:** Allow users to log in with Google/GitHub instead of password
**Type:** New feature
```

### 1.4 Review

Present intent. User confirms or corrects.

---

## Stage 2: Understand Current State

### 2.1 Analyze

Follow layer navigation to read affected C3 docs:
- Context (c3-0) - for system-level context
- Relevant Container(s) - for component inventory
- Relevant Component(s) - for current behavior

Build open questions list about current architecture.

### 2.2 Ask (Socratic)

Use AskUserQuestion until confident about current state.

Example questions:
- "The docs say AuthProvider handles JWT only. Is that still accurate?"
- "Container c3-2 lists 'LoginFlow' but no OAuth. Has OAuth been added without docs?"
- "Are there any recent code changes not reflected in C3 docs?"

### 2.3 Synthesize

Summarize current state:
```
**Current State:**
- Auth handled by: c3-102 AuthProvider (Foundation)
- Login flow: c3-205 LoginFlow (Feature) - password only
- External auth: None currently documented
- Dependencies: E1 PostgreSQL (user storage)
```

### 2.4 Review

Present current state summary. User confirms or corrects.

---

## Stage 3: Scope Impact

### 3.1 Analyze

Map the change to affected layers. Form hypotheses:
- Which layers will change?
- What depends on what's changing?
- Will linkages change?
- Will diagrams change?

Build open questions list.

### 3.2 Ask (Socratic)

Use AskUserQuestion until confident about full scope.

Example questions:
- "Adding OAuth means a new External System. Should this be Google, GitHub, or both?"
- "Will existing password login remain, or be replaced?"
- "Does the API contract change? Any breaking changes for clients?"

### 3.3 Synthesize

Impact assessment:
```
**Scope:**
Layers affected:
- c3-0 Context: Add External System (Google OAuth)
- c3-1 backend: Add linkage to OAuth
- c3-102 AuthProvider: Extend contract for OAuth tokens
- c3-205 LoginFlow: Add OAuth trigger path
- NEW: c3-206 OAuthCallback (Feature)

Breaking changes: None (additive)
```

### 3.4 Review

Present scope. User confirms or expands.

---

## Stage 4: Create ADR

### 4.1 Synthesize

Generate ADR at `.c3/adr/adr-YYYYMMDD-{slug}.md`.

Use template from `../../references/adr-template.md`.

Key sections:
- **Problem:** Why this change (2-3 sentences)
- **Decision:** What we're doing (clear, direct)
- **Rationale:** Why this approach over alternatives
- **Affected Layers:** List all c3 IDs from scope
- **Verification Checklist:** How to confirm success

### 4.2 Review

Present ADR for user acceptance.

**On accept:** Update status to `accepted`, proceed to Plan.

**On reject:**
- If scope changed → return to Stage 3
- If intent changed → return to Stage 1
- If wording issue → revise ADR, re-present

---

## Stage 5: Create Plan

### 5.1 Analyze

Determine order of operations:
- Which docs update first?
- Which code changes depend on which?
- Any migrations or deployment steps?

Build open questions list.

### 5.2 Ask (Socratic)

Use AskUserQuestion until confident about execution order.

Example questions:
- "Should we update the AuthProvider contract before or after adding OAuthCallback?"
- "Are there tests that need updating? Run before or after code changes?"
- "Any feature flags or rollout considerations?"

### 5.3 Synthesize

Generate plan at `.c3/adr/adr-YYYYMMDD-{slug}.plan.md`.

```
## Order of Operations

1. Update Context: Add E2 Google OAuth
2. Update Container c3-1: Add linkage to E2
3. Update c3-102 AuthProvider: Extend contract
4. Create c3-206 OAuthCallback: New component doc
5. Implement code changes:
   - auth/oauth.ts (new)
   - auth/provider.ts (extend)
   - routes/callback.ts (new)
6. Update c3-205 LoginFlow: Add OAuth trigger
7. Update diagrams
8. Run tests
9. Verify
```

### 5.4 Review

Present plan. User approves or adjusts.

---

## Stage 6: Execute

### 6.1 Apply Changes

Follow plan order. For each item:
1. Make the change (doc or code)
2. Check for conflicts with earlier assumptions

### 6.2 Handle Conflicts (Ascent)

During execution, discoveries may require ascending:

```
Updating c3-102 AuthProvider...
  → Discovers it also uses Redis for session cache
  → But Context doesn't list Redis
  → ASCEND: Is this in scope?
```

**Tiered assumptions:**

| High-Impact (ask first) | Low-Impact (auto-proceed) |
|-------------------------|---------------------------|
| Scope expansion | Minor doc wording fix |
| New affected layer discovered | Fixing ID inconsistency |
| Breaking change detected | Updating diagram to match |
| ADR needs revision | Adding linkage reasoning |

For high-impact: Ask user, potentially update ADR, then continue.
For low-impact: Fix it, note in execution log, continue.

### 6.3 Execution Log

Track what was changed:
```
**Executed:**
- [x] Context: Added E2 Google OAuth
- [x] Container c3-1: Added linkage
- [x] c3-102: Extended contract (also noted Redis dependency - added to Context)
- [ ] c3-206: In progress...
```

---

## Stage 7: Verify

### 7.1 Run Audit

```
/c3 audit
```

Check:
- Diagrams match inventory tables
- All IDs consistent
- Linkages have reasoning
- Code matches docs

### 7.2 Handle Failures

**On pass:** Update ADR status to `implemented`. Done.

**On fail:**
- Identify what failed
- Fix the issue
- Re-run audit
- Loop until pass

---

## Tiered Assumption Rules

### High-Impact (ask user first)
- Scope expansion (new layer affected)
- Breaking change detected
- ADR needs significant revision
- New External System discovered
- Container boundary change

### Low-Impact (auto-proceed, note in log)
- Adding linkage reasoning
- Minor doc wording improvements
- Fixing diagram inconsistency
- Updating IDs for consistency
- Adding missing test scenarios

---

## Red Flags - STOP

| Situation | Action |
|-----------|--------|
| "Just a small change" | Still needs ADR. Small changes cascade. |
| "I'll update docs later" | No. ADR first. |
| "Skip ADR, it's obvious" | Not obvious to future readers. |
| Changing code before ADR accepted | Stop. Complete Stage 4 first. |
| Uncertain about scope | Loop in Stage 3 until confident. |

---

## Response Format

At each stage, show:

```
**Stage N: {Name}**

{Analysis findings}

**Open Questions:** {list, or "None - confident to proceed"}

**Next:** {What happens next}
```
