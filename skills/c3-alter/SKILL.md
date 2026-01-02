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

**Before loading current state, you MUST read `references/layer-navigation.md`.**

This reference contains:
- Activation check
- Traversal order (Context → Container → Component)
- ID-to-path mapping

## Why ADR-First?

Changes have ripple effects. ADR forces you to:
1. Understand current state
2. Identify all affected layers
3. Document the decision
4. Plan the execution
5. Verify the result

## Core Principle

```
INTENT → UNDERSTAND → SCOPE → ADR → ACCEPT → PLAN → EXECUTE → VERIFY
         ↑                                                    │
         └────────────── If verify fails ─────────────────────┘
```

**No code changes before ADR is accepted.**

## Workflow

### Step 1: Understand Intent

Ask clarifying questions:

| Question | Why |
|----------|-----|
| What's the change? | Add, modify, remove, fix? |
| Why? | Problem being solved, goal |
| Scope? | Which part of the system? |
| Urgency? | Breaking change? Migration needed? |

Don't assume. Small changes can have large impacts.

### Step 2: Load Current State

Follow `references/layer-navigation.md` to traverse affected layers.

Load only what's relevant to the change scope identified in Step 1.

### Step 3: Scope Impact

Map the change to affected layers:

```
Change Type          │ Likely Impact
─────────────────────┼──────────────────────────────
New feature          │ Container + new Component(s)
Bug fix              │ Component only (usually)
Refactor             │ Container + Component(s)
New integration      │ Context + Container
Breaking change      │ Context + all downstream
Rename/restructure   │ All layers referencing it
```

**Impact Assessment Questions:**
- Which layers document this?
- What depends on this?
- What does this depend on?
- Will linkages change?
- Will diagrams change?

### Step 4: Create ADR

Generate ADR at `.c3/adr/adr-YYYYMMDD-{slug}.md`.

See `references/adr-template.md` for full template.

**Key sections:**
- Problem (2-3 sentences)
- Decision (clear, direct)
- Rationale (tradeoffs)
- Affected Layers (c3-0, c3-N, c3-NNN)
- Verification Checklist

### Step 5: User Accepts

Present ADR for review. User must explicitly accept.

On acceptance:
- Update status to `accepted`
- Proceed to planning

On rejection:
- Revise based on feedback
- Return to Step 1 if scope changed

### Step 6: Create Plan

Generate execution plan at `.c3/adr/adr-YYYYMMDD-{slug}.plan.md`.

See `references/plan-template.md` for full template.

**Key sections:**
- Order of Operations
- Changes Per File (docs first, then code)
- Verification Steps

### Step 7: Execute

Apply changes in order specified:

1. **Update C3 docs first** (if changing existing)
2. **Create new C3 docs** (if adding)
3. **Make code changes**
4. **Update diagrams** to match new state

### Step 8: Verify

Run audit on affected scope:

```
c3-skill:c3 audit container c3-N
```

Checks:
- Diagrams match inventory tables
- All IDs consistent
- Linkages have reasoning
- Fulfillment covers Context links
- Code matches what docs describe

**On pass:** Update ADR status to `implemented`

**On fail:** Fix issues, re-verify (loop back)

## Change Categories

### Code-First Changes
User wants to change code → Still needs ADR:
1. Understand what code change
2. Map to C3 layer(s)
3. Create ADR documenting the change
4. Accept → Plan → Execute (code + docs) → Verify

### Doc-First Changes
User wants to update architecture → ADR captures why:
1. Understand what doc change
2. Create ADR
3. Accept → Plan → Execute → Verify

### Correction Changes
User found docs are wrong → ADR documents drift:
1. Identify what's incorrect
2. ADR documents: "Docs said X, reality is Y"
3. Accept → Plan → Execute → Verify

## Red Flags - STOP

| Situation | Action |
|-----------|--------|
| "Just a small change" | Still needs ADR. Small changes cascade. |
| "I'll update docs later" | No. ADR first, or change isn't tracked. |
| "Skip ADR, it's obvious" | Not obvious to future readers. ADR. |
| Changing code without accepted ADR | Stop. Create ADR first. |

## Response Format

```
**Change Intent:** {summary of what user wants}

**Current State:**
- Loaded: {layers consulted}
- Affected: {layers that will change}

**Impact Assessment:**
{What will change and why}

**Next Step:** {Create ADR / Revise scope / etc.}
```
