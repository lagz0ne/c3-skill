---
name: c3-design
description: Use when designing or updating system architecture with the C3 methodology - iteratively scope through hypothesis, exploration, and discovery across Context/Container/Component layers
---

# C3 Architecture Design

## Overview

Transform requirements into structured C3 (Context-Container-Component) architecture documentation through iterative scoping.

**Core principle:** Form hypothesis, explore to validate, discover impacts, iterate until stable. Uncertainty is expected and healthy.

**Announce at start:** "I'm using the c3-design skill to guide you through architecture design."

## Mandatory Phase Tracking

**IMMEDIATELY after announcing, create TodoWrite items:**

```
Phase 1: Surface Understanding - Read TOC, form hypothesis
Phase 2: Iterative Scoping - HYPOTHESIZE → EXPLORE → DISCOVER until stable
Phase 3: ADR Creation - Create ADR file in .c3/adr/ (MANDATORY)
Phase 4: Handoff - Execute settings.yaml handoff steps
```

**Rules:**
- Mark each phase `in_progress` when starting
- Mark `completed` only when gate criteria met
- **Phase 3 gate:** ADR file MUST exist before marking complete
- **Phase 4 gate:** Handoff steps MUST be executed before marking complete

**Skipping phases = skill failure. No exceptions.**

## Quick Reference

| Phase | Key Activities | Output | Gate |
|-------|---------------|--------|------|
| **1. Surface Understanding** | Read TOC, parse request, form hypothesis | Initial hypothesis | TodoWrite: phases tracked |
| **2. Iterative Scoping** | HYPOTHESIZE → EXPLORE → DISCOVER loop | Stable scope | Scope stable, all IDs named |
| **3. ADR Creation** | Document journey, changes, verification | ADR in `.c3/adr/` | **ADR file exists** |
| **4. Handoff** | Execute settings.yaml handoff steps | Tasks/notifications created | Handoff complete |

**MANDATORY:** You MUST create TodoWrite items for each phase. No phase can be skipped.

## Prerequisites

**Required:** `.c3/` directory with `TOC.md` must exist.

**Pre-flight sanity checks (especially when auditing examples):**
- Open the TOC first (e.g., `examples/.c3/TOC.md` in this repo) instead of hopping between files.
- Scan for duplicate IDs across containers/components and fix them before scoping so navigation stays unambiguous.

If `.c3/` doesn't exist:
- Stop and inform user to initialize structure first
- Suggest: "Use the `c3-adopt` skill to initialize C3 documentation for this project"
- The `c3-adopt` skill will discover the project structure and create initial documents

**Load project settings (if exists):**
```bash
cat .c3/settings.yaml 2>/dev/null
```

If `.c3/settings.yaml` exists, apply preferences throughout the session:
- `diagrams` - Use specified tool and patterns when creating diagrams
- `context`/`container`/`component` - Follow layer-specific guidance when writing docs
- `guard` - Respect guardrails as constraints during design
- `handoff` - Follow post-ADR steps after decision is accepted

Settings are optional - if missing, use sensible defaults.

## The Process

### Phase 1: Surface Understanding

**Goal:** Form initial hypothesis about what's affected.

**Actions:**

1. **Read TOC for current state**
   ```bash
   cat .c3/TOC.md
   ```

2. **Parse user request**
   - What do they think they want?
   - What words/concepts map to existing documents?

3. **Form initial hypothesis** (THINKING, not asking)
   - Which layer? (Context / Container / Component)
   - Which specific element? (CTX-system-overview? C3-2-backend? C3-103-db-pool?)
   - Why do you think so?
   - What's uncertain?

**Output:** Abstract, high-level hypothesis to explore.

### Phase 2: Iterative Scoping (Core Loop)

**Goal:** Validate and refine hypothesis until scope is stable.

```
┌────────────────────────────────────────────────────────┐
│                                                        │
│   HYPOTHESIZE (abstract, from TOC + understanding)     │
│        ↓                                               │
│   EXPLORE (investigate with c3-locate + sub-skills)    │
│        │                                               │
│        ├── Socratic questions as needed                │
│        │   (confirm understanding along the way)       │
│        ↓                                               │
│   DISCOVER (what did exploration reveal?)              │
│        │                                               │
│        ├── Need to revise? → Update hypothesis, loop   │
│        │                                               │
│        └── Stable? → Exit to Phase 3                   │
│                                                        │
└────────────────────────────────────────────────────────┘
```

#### HYPOTHESIZE

Form or update hypothesis based on current understanding:
- "This likely affects [X] because [reasoning]"
- "But could also be [Y] if [condition]"
- Map to specific document IDs from TOC

**This is internal reasoning, not questions to user.**

#### EXPLORE

Investigate hypothesis in 4 directions:

| Direction | Question | Tool |
|-----------|----------|------|
| **Isolated** | What changes directly at this element? | c3-locate by ID |
| **Upstream** | What feeds into this? Dependencies? | c3-locate related IDs |
| **Adjacent** | What's at same level? Siblings? | c3-locate same-layer IDs |
| **Downstream** | What does this affect? Consumers? | c3-locate dependent IDs |

**Use sub-skills for layer-specific exploration:**
- `c3-locate` - ID-based content retrieval
- `c3-context-design` - Explore Context-level impact
- `c3-container-design` - Explore Container-level impact
- `c3-component-design` - Explore Component-level impact

**Socratic questions during exploration** to confirm understanding:
- "Based on C3-1-backend, the auth middleware handles tokens. Is that still accurate?"
- "I see C3-102-auth-service depends on this. Does that dependency need to change?"

#### DISCOVER

Assess what exploration revealed:

| Discovery | Signal | Action |
|-----------|--------|--------|
| Impact at **higher** abstraction | Bigger than thought | Form new hypothesis at higher level, loop |
| Impact at **same level** widely | Scope expansion | Expand hypothesis, continue exploring |
| Impact only **downstream** | Contained | Scope is stable, proceed |
| No new impacts | Complete | Exit to Phase 3 |

**Key principle:**
- Upstream/higher-level impacts → revisit hypothesis
- Downstream/lower-level impacts → expected, proceed

#### Loop Exit Criteria

Scope is stable when:
- You can name all affected documents (by ID)
- You understand why each is affected
- No exploration reveals new upstream/higher impacts
- Socratic confirmation validates understanding

**MANDATORY NEXT STEP:** Phase 3 (ADR Creation) is required.

You CANNOT:
- Update any C3 documents directly
- Skip to handoff
- Consider the design "done"

Until you have created an ADR file in `.c3/adr/`.

### Phase 3: ADR Creation (MANDATORY)

**Goal:** Document the decision capturing the full scoping journey.

**This phase is NOT optional.** You must create an ADR before any document updates.

**Determine ADR number:**
```bash
last_adr=$(find .c3/adr -name "ADR-*.md" | sed 's/.*ADR-\([0-9]*\).*/\1/' | sort -n | tail -1)
next_num=$(printf "%03d" $((10#${last_adr:-0} + 1)))
```

**ADR Template (Stream Format):**

```markdown
---
id: ADR-{NNN}-{slug}
title: [Decision Title]
summary: >
  [Why read this - what decision, what it affects]
status: proposed
date: YYYY-MM-DD
---

# [ADR-{NNN}] [Decision Title]

## Status {#adr-{nnn}-status}
**Proposed** - YYYY-MM-DD

## Problem/Requirement {#adr-{nnn}-problem}
<!--
Starting point - what user asked for, why change is needed.
-->

[What triggered this decision]

## Exploration Journey {#adr-{nnn}-exploration}
<!--
How understanding developed through scoping.
-->

**Initial hypothesis:** [What we first thought]

**Explored:**
- Isolated: [What we found at the element]
- Upstream: [Dependencies discovered]
- Adjacent: [Related elements at same level]
- Downstream: [Consumers/dependents affected]

**Discovered:** [Key insights that shaped the solution]

**Confirmed:** [What Socratic questions validated]

## Solution {#adr-{nnn}-solution}
<!--
Formed through exploration above.
-->

[The approach and why it fits]

## Changes Across Layers {#adr-{nnn}-changes}
<!--
Specific changes to each affected document.
-->

### Context Level
- [CTX-slug]: [What changes, why]

### Container Level
- [C3-<C>-slug]: [What changes, why]

### Component Level
- [C3-<C><NN>-slug]: [What changes, why]

## Verification {#adr-{nnn}-verification}
<!--
Checklist derived from scoping - what to inspect when implementing.
-->

- [ ] Is [X] at the right abstraction level?
- [ ] Does [Y] upstream dependency still hold?
- [ ] Are [Z] downstream consumers updated?
- [ ] [Specific checks from exploration]

## Related {#adr-{nnn}-related}
- [Links to affected documents]
```

**Phase 3 Gate - Verify ADR exists:**

```bash
# Verify ADR was created
ls .c3/adr/adr-*.md | tail -1
```

If no ADR file exists, **STOP**. Create the ADR before proceeding.

**After ADR verified:**
1. Update affected documents (CTX/CON/COM) as specified
2. Regenerate TOC: `.c3/scripts/build-toc.sh`
3. **Proceed to Phase 4: Handoff**

## Sub-Skill Invocation

Use the Skill tool to invoke during exploration:

| Skill | When to Use |
|-------|-------------|
| `c3-adopt` | Initialize C3 documentation for existing project (if `.c3/` doesn't exist) |
| `c3-toc` | Manage TOC, inspect document tree, rebuild index |
| `c3-locate` | Retrieve content by document/heading ID |
| `c3-context-design` | Explore Context-level impact |
| `c3-container-design` | Explore Container-level impact |
| `c3-component-design` | Explore Component-level impact |

## Example (Auditing bundled sample)

Goal: sanity check `examples/.c3` before sharing with others.
- **HYPOTHESIZE:** Index shows backend and frontend components reusing numbers; expect duplicate IDs and missing TOC.
- **EXPLORE:** Open `examples/.c3/TOC.md`, scan backend components, check frontend client; confirm legacy CON/COM prefixes caused collisions (e.g., **before:** `COM-004` reused by logger and API client, REST routes sharing numbers).
- **DISCOVER:** Add the TOC, align IDs to `C3-<C>`/`C3-<C><NN>` (e.g., `C3-106-rest-routes` for REST routes, `C3-201-api-client` for frontend API client), drop legacy duplicates, and refresh index links.
- **Outcome:** Unique IDs and clean downward links ready for ADR and implementation work.

## Key Principles

| Principle | Application |
|-----------|-------------|
| **Hypothesis first** | Form from TOC, don't ask directly for location |
| **Explore to validate** | Investigate before confirming |
| **Socratic during exploration** | Questions confirm understanding, not discover location |
| **ID-based navigation** | Use document/heading IDs, not keyword search |
| **Higher = bigger impact** | Upstream/higher-level discoveries trigger revisit |
| **ADR as stream** | Capture journey, not just final answer |
| **Iterate freely** | Loop until stable, don't force forward |

## Common Mistakes

- Skipping the TOC/ID sweep and diving into files, which hides upstream impacts and duplicate IDs.
- Asking the user to point to files instead of forming a hypothesis from the TOC and existing docs.
- Drafting an ADR before the hypothesis → explore → discover loop stabilizes.
- Treating examples as throwaway and allowing duplicate IDs or missing TOC to persist.

## Red Flags & Counters

| Rationalization | Counter |
|-----------------|---------|
| "No time to refresh the TOC, I'll just skim files" | Stop and build/read the TOC first; C3 navigation depends on it. |
| "Examples can keep duplicate IDs, they're just sample data" | IDs must be unique or locate/anchor references break—fix collisions before scoping. |
| "I'll ask the user where to change docs instead of hypothesizing" | Hypothesis bounds exploration and prevents confirmation bias; form it before asking questions. |

**Red flags that mean you should pause:**
- `.c3/TOC.md` missing or obviously stale.
- Component IDs reused across containers or layers.
- ADR being drafted without notes from hypothesis, exploration, and discovery.

## Common Patterns

### New Feature
Surface → Hypothesis at Container level → Explore up to Context, down to Components → ADR with cross-layer changes

### Bug Fix
Surface → Hypothesis at Component level → Explore if isolated or upstream cause → ADR focused on Component, verify no upstream issues

### Architectural Change
Surface → Hypothesis at Context level → Explore all downstream Containers/Components → Large ADR with many changes

### Refactoring
Surface → Hypothesis at Component level → Explore adjacent components → ADR focused on Component layer
