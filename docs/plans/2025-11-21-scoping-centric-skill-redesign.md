# Scoping-Centric Skill Redesign

> Design for improving c3-design skill to make iterative scoping the core workflow.

## Problem

Current skill treats phases too linearly (Understand → Confirm → Scope → ADR). Real architectural decisions require bouncing between layers to truly understand impact. Being "too sure" early leads to low quality decisions.

## Core Insight

**Uncertainty is expected and healthy.** Scoping is where understanding develops through:
- Forming hypotheses from TOC/surface understanding
- Exploring to validate (not asking directly)
- Discovering impacts that revise the hypothesis
- Iterating until stable

**Abstraction level signals impact magnitude:**
- Change belongs at higher level than thought → bigger systemic impact
- Change belongs at lower level → more isolated

## New Phase Structure

### Before (4 phases, linear)
```
1. Understand → 2. Confirm → 3. Scope → 4. ADR
```

### After (3 phases, scoping-centric)
```
1. Surface Understanding
2. Iterative Scoping (core loop)
3. ADR with Stream
```

## Phase Details

### Phase 1: Surface Understanding

- Read existing `.c3/TOC.md`
- Parse user request (what do they think they want?)
- Form initial hypothesis: which layer, which element
- Hypothesis is abstract, high-level

### Phase 2: Iterative Scoping (Core Loop)

```
┌────────────────────────────────────────────────────────┐
│                                                        │
│   HYPOTHESIZE (abstract, high-level)                   │
│        ↓                                               │
│   EXPLORE (investigate the hypothesis)                 │
│        │                                               │
│        ├── Use c3-locate for ID-based content lookup   │
│        ├── Socratic questions as needed                │
│        │   (confirm understanding along the way)       │
│        ↓                                               │
│   DISCOVER (what did exploration reveal?)              │
│        │                                               │
│        ├── Need to revise? → Update hypothesis, loop   │
│        │                                               │
│        └── Stable? → Exit to ADR                       │
│                                                        │
└────────────────────────────────────────────────────────┘
```

#### HYPOTHESIZE
- From Surface Understanding + TOC
- Map user's words to existing documents
- Form hypothesis: "This likely affects [X] because [reasoning]"
- Identify uncertainty: "But could also be [Y] if [condition]"
- This is THINKING, not asking

#### EXPLORE
Four directions from current location:
- **Isolated**: What changes directly at this element?
- **Upstream**: What feeds into this? (dependencies, data sources)
- **Adjacent**: What's at same level? (sibling containers/components)
- **Downstream**: What does this affect? (consumers, dependents)

Use `c3-locate` sub-skill for ID-based content retrieval during exploration.

Socratic questions happen during exploration to confirm understanding, not as a separate gate.

#### DISCOVER
Impact assessment triggers next action:
- Impact at higher abstraction → form new hypothesis at that level, loop
- Impact at same level widely → scope is bigger than thought, expand
- Impact only downstream → scope is contained, can proceed
- Stable → exit to ADR

**Key principle:** Upstream/higher-level impacts signal revisit. Downstream/lower-level impacts are expected.

### Phase 3: ADR with Stream

ADR captures the full journey, not just the answer:

```markdown
## Problem/Requirement
What user asked for (starting point)

## Exploration Journey
- Initial hypothesis: [what we thought]
- Explored: [directions investigated]
- Discovered: [what we found - upstream/downstream/adjacent impacts]
- Revised to: [updated hypothesis if changed]
- Confirmed: [what Socratic questions validated]

## Solution
Formed through iteration above

## Changes (across layers)
- CTX-001: [what changes, why]
- C3-2: [what changes, why]
- C3-203: [what changes, why]

## Verification
- [ ] Check: Is [X] at right abstraction level?
- [ ] Check: Does [Y] upstream still hold?
- [ ] Check: Are [Z] downstream consumers updated?
```

## New Sub-skill: c3-locate

Purpose: Retrieve content by ID during exploration.

### Primary Mode (ID-based)
- Input: Document ID (CTX-001, C3-2, C3-203)
- Input: Heading ID (#c3-1-middleware, #c3-102-configuration)
- Output: Frontmatter + section content

Examples:
```
c3-locate C3-1              → frontmatter + overview
c3-locate #c3-1-middleware  → heading content + summary
c3-locate C3-102 #c3-102-error-handling → specific section
```

### Secondary Mode (discovery, rare)
- When you don't know the ID yet
- Search by concept to find relevant IDs
- Then switch to ID-based lookup

## Sub-skills Role

Sub-skills become **exploration tools** rather than separate workflows:

| Sub-skill | Role in Exploration |
|-----------|---------------------|
| c3-locate | ID-based content retrieval |
| c3-context-design | Understand Context-level impact |
| c3-container-design | Understand Container-level impact |
| c3-component-design | Understand Component-level impact |

Main `c3-design` skill orchestrates the loop, sub-skills provide layer-specific exploration.

## Updated Skill Structure

```
c3-design (main skill)
├── Phase 1: Surface Understanding
│   - Read .c3/TOC.md
│   - Parse user request
│   - Form initial hypothesis
│
├── Phase 2: Iterative Scoping (loop)
│   - HYPOTHESIZE → EXPLORE → DISCOVER
│   - Use c3-locate for ID-based lookup
│   - Invoke layer sub-skills as exploration tools
│   - Socratic questions during exploration
│   - Loop until stable
│
└── Phase 3: ADR with Stream
    - Problem/Requirement
    - Exploration Journey (captured from loop)
    - Solution
    - Changes across layers
    - Verification checklist
```

## Implementation Tasks

1. Rewrite `skills/c3-design/SKILL.md` with new 3-phase structure
2. Create `skills/c3-locate/SKILL.md` sub-skill
3. Update layer sub-skills to be exploration-focused
4. Update ADR template in skill to use stream format
5. Update examples to reflect new workflow
