---
name: c3-design
description: Design or update system architecture using C3 methodology - iterative scoping through hypothesis, exploration, and discovery across Context/Container/Component layers
---

# C3 Architecture Design

## Overview

Transform requirements into structured C3 (Context-Container-Component) architecture documentation through iterative scoping.

**Core principle:** Form hypothesis, explore to validate, discover impacts, iterate until stable. Uncertainty is expected and healthy.

**Announce at start:** "I'm using the c3-design skill to guide you through architecture design."

## Quick Reference

| Phase | Key Activities | Output |
|-------|---------------|--------|
| **1. Surface Understanding** | Read TOC, parse request, form hypothesis | Initial hypothesis |
| **2. Iterative Scoping** | HYPOTHESIZE → EXPLORE → DISCOVER loop | Stable scope |
| **3. ADR with Stream** | Document journey, changes, verification | ADR in `.c3/adr/` |

## Prerequisites

**Required:** `.c3/` directory with `TOC.md` must exist.

If `.c3/` doesn't exist:
- Stop and inform user to initialize structure first
- Suggest: "Create `.c3/` directory to start, or use `/c3-init` to initialize"

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
   - Which specific element? (CTX-001? CON-002? COM-003?)
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
- "Based on CON-001, the auth middleware handles tokens. Is that still accurate?"
- "I see COM-002 depends on this. Does that dependency need to change?"

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

### Phase 3: ADR with Stream

**Goal:** Document the decision capturing the full scoping journey.

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
- [CTX-XXX]: [What changes, why]

### Container Level
- [CON-XXX]: [What changes, why]

### Component Level
- [COM-XXX]: [What changes, why]

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

**After ADR:**
1. Update affected documents (CTX/CON/COM) as specified
2. Regenerate TOC: `.c3/scripts/build-toc.sh`

## Sub-Skill Invocation

Use the Skill tool to invoke during exploration:

| Skill | When to Use |
|-------|-------------|
| `c3-locate` | Retrieve content by document/heading ID |
| `c3-context-design` | Explore Context-level impact |
| `c3-container-design` | Explore Container-level impact |
| `c3-component-design` | Explore Component-level impact |

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

## Common Patterns

### New Feature
Surface → Hypothesis at Container level → Explore up to Context, down to Components → ADR with cross-layer changes

### Bug Fix
Surface → Hypothesis at Component level → Explore if isolated or upstream cause → ADR focused on Component, verify no upstream issues

### Architectural Change
Surface → Hypothesis at Context level → Explore all downstream Containers/Components → Large ADR with many changes

### Refactoring
Surface → Hypothesis at Component level → Explore adjacent components → ADR focused on Component layer
