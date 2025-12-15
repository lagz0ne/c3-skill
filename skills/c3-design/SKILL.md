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
| **3. ADR + Plan Creation** | Document journey, changes, verification, **AND implementation plan** | ADR with Plan in `.c3/adr/` | **ADR file exists with Implementation Plan section** |
| **4. Handoff** | Execute settings.yaml handoff steps | Tasks/notifications created | Handoff complete |

**MANDATORY:** You MUST create TodoWrite items for each phase. No phase can be skipped.

## ADR + Plan: Inseparable Pair

Every design change produces **both** an ADR and an Implementation Plan. They are two sides of the same coin:

```
ADR (medium abstraction)              Plan (low abstraction)
├── Changes Across Layers ───────────→ Code Changes
├── Verification ────────────────────→ Acceptance Criteria
└────────────────── Mutual Reference ─────────────────────┘
```

**ADR Lifecycle:**
```
proposed → accepted → implemented
    ↓         ↓
superseded/deprecated
```

- **proposed**: Created by c3-design, awaiting review
- **accepted**: Approved by team, ready for implementation
- **implemented**: Verified by c3-audit, appears in TOC

> **Note:** Only `implemented` ADRs appear in the Table of Contents. Use `c3-audit` to verify implementation and transition status.

**Rules:**
- No ADR is complete without an Implementation Plan section
- Every "Changes Across Layers" item must map to a Code Change
- Every Verification item must map to an Acceptance Criterion
- Audit will verify this coherence

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

   <extended_thinking>
   <goal>Form initial hypothesis about which C3 layer and element is affected</goal>

   <layer_detection>
   Parse user request for layer signals:

   CONTEXT signals (system-wide):
   - "add a new service/container"
   - "change how services communicate"
   - "add authentication across the system"
   - "new type of user/actor"
   - Words: boundary, protocol, cross-cutting, system-wide

   CONTAINER signals (service-level):
   - "change the backend/frontend/worker"
   - "add new functionality to [service]"
   - "refactor [service] internals"
   - Words: technology, patterns, components within

   COMPONENT signals (implementation-level):
   - "fix this bug in [specific module]"
   - "optimize [specific function]"
   - "add a field to [specific endpoint]"
   - Words: code, implementation, configuration
   </layer_detection>

   <hypothesis_formation>
   Based on signals:
   - Primary layer: [Context/Container/Component]
   - Specific element: [map to TOC ID]
   - Evidence: [what in the request points here]
   - Uncertainty: [what could make this wrong]
   - Upstream risk: [could this be bigger?]
   </hypothesis_formation>

   <confidence_assessment>
   HIGH if: Clear layer signal, maps directly to TOC ID
   MEDIUM if: Some ambiguity, multiple possible elements
   LOW if: Vague request, could be multiple layers

   If LOW confidence: Explore more before committing to hypothesis
   </confidence_assessment>
   </extended_thinking>

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

<thinking>
Update hypothesis based on what I've learned:
- Current hypothesis: [X] because [reasoning]
- Alternative: Could be [Y] if [condition]
- Document IDs involved: [list from TOC]
- Confidence level: High/Medium/Low
- What would change my mind?
</thinking>

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

<extended_thinking>
<goal>Evaluate exploration results and decide whether to iterate or exit</goal>

<impact_classification>
Classify each finding by direction:

UPSTREAM (higher level) findings:
- Needs new protocol → Context impact
- Changes actor interface → Context impact
- Violates boundary → Context impact
- Changes cross-cutting pattern → Context impact
→ SIGNAL: Hypothesis was too narrow

ADJACENT (same level) findings:
- Affects sibling containers/components
- Changes shared interfaces
- Impacts shared data structures
→ SIGNAL: Scope expanding horizontally

DOWNSTREAM (lower level) findings:
- Child documents need updates
- Implementation details affected
- Normal propagation expected
→ SIGNAL: Contained, proceed

NO NEW findings:
- Exploration confirmed hypothesis
- No surprises
→ SIGNAL: Scope stable
</impact_classification>

<scope_stability_check>
Questions to determine stability:
1. Can I name ALL affected documents by ID? [list them]
2. For each affected document, do I know WHY it's affected?
3. Did exploration reveal any UPSTREAM impacts? [if yes, unstable]
4. Are there documents I'm UNSURE about? [if yes, explore more]
5. Have Socratic questions validated my understanding?
</scope_stability_check>

<decision>
IF upstream_impacts_found THEN
  - Scope is UNSTABLE
  - Form new hypothesis at higher level
  - LOOP BACK to HYPOTHESIZE

ELIF adjacent_impacts_expanding THEN
  - Scope is GROWING
  - Add affected siblings to hypothesis
  - Continue EXPLORE phase

ELIF only_downstream_or_none THEN
  - Scope is STABLE
  - EXIT to Phase 3 (ADR Creation)
</decision>

<iteration_counter>
Track iterations to prevent infinite loops:
- Iteration 1: Initial hypothesis
- Iteration 2: [if upstream found]
- Iteration 3: [if still expanding]
- Iteration 4+: Consider if scope is unbounded → ask user for constraints
</iteration_counter>
</extended_thinking>

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

**EXIT GATE - ALL must be true:**

| Criterion | Check | If No |
|-----------|-------|-------|
| Complete ID list | Can name ALL affected documents by ID | Continue exploring |
| Impact clarity | Know WHY each document is affected | Ask clarifying questions |
| No upstream surprises | Last round revealed no new upstream impacts | Revise hypothesis |
| Socratic confirmation | Questions validated understanding | Ask more questions |

> **Gate passed?** Exit to Phase 3 (ADR Creation)
> **Gate failed?** Another iteration needed

<thinking>
Internal verification:
- Documents affected: [list IDs]
- Why each affected: [brief reason per ID]
- New upstream impacts this round: [yes/no, list if yes]
- Socratic confirmation status: [confirmed/pending]
</thinking>

**MANDATORY:** Phase 3 (ADR Creation) cannot be skipped.

You CANNOT:
- Update any C3 documents directly
- Skip to handoff
- Consider the design "done"

Until you have created an ADR file in `.c3/adr/`.

### Phase 3: ADR + Plan Creation (MANDATORY)

**Goal:** Document the decision AND the implementation plan as an inseparable pair.

**This phase is NOT optional.** You must create an ADR with Implementation Plan before any document updates.

**Determine ADR filename:**
```bash
today=$(date +%Y%m%d)
# Use: .c3/adr/adr-YYYYMMDD-{slug}.md
```

**Create ADR using template:** See [adr-template.md](../../references/adr-template.md) for full template.

**ADR Structure Overview:**

```yaml
---
id: adr-YYYYMMDD-{slug}
title: [Decision Title]
status: proposed
date: YYYY-MM-DD
---

# [Decision Title]
## Status
## Problem/Requirement       # What triggered this
## Exploration Journey       # hypothesis -> explore -> discover
## Solution                  # Approach formed through exploration
## Changes Across Layers     # Specific changes per C3 document
## Verification              # Checklist from scoping
## Implementation Plan       # MANDATORY
  ### Code Changes           # Maps from Changes Across Layers
  ### Dependencies           # Order of operations
  ### Acceptance Criteria    # Maps from Verification
## Related
```

**Key principle:** Changes Across Layers -> Code Changes (1:1), Verification -> Acceptance Criteria (1:1)

<extended_thinking>
<goal>Verify ADR and Plan are coherent before marking Phase 3 complete</goal>

<coherence_check>
For each item in "Changes Across Layers":
- Is there a corresponding entry in "Code Changes"?
- Is the code location specific (file:function, not "somewhere")?

For each item in "Verification":
- Is there a corresponding entry in "Acceptance Criteria"?
- Is the criterion testable (command/test, not "should work")?

Mapping must be complete:
- Orphan doc changes (no code work) = incomplete
- Orphan verifications (no criteria) = incomplete
- Vague code locations = incomplete
</coherence_check>

<mutual_reference_verification>
ADR references Plan:
- "Changes Across Layers" points to "Code Changes"
- "Verification" points to "Acceptance Criteria"

Plan references ADR:
- Each Code Change traces back to a Layer Change
- Each Acceptance Criterion traces back to a Verification item

If mapping is incomplete → DO NOT proceed to Phase 4
</mutual_reference_verification>
</extended_thinking>

**Phase 3 Gate - Verify ADR + Plan coherence:**

```bash
# Verify ADR was created
ls .c3/adr/adr-*.md | tail -1

# Verify Implementation Plan section exists
grep -q "## Implementation Plan" .c3/adr/adr-*.md && echo "Plan section: OK"

# Verify Code Changes table exists
grep -q "### Code Changes" .c3/adr/adr-*.md && echo "Code Changes: OK"

# Verify Acceptance Criteria table exists
grep -q "### Acceptance Criteria" .c3/adr/adr-*.md && echo "Acceptance Criteria: OK"
```

**Gate criteria (ALL must pass):**
- [ ] ADR file exists
- [ ] Implementation Plan section exists
- [ ] Code Changes maps to all "Changes Across Layers"
- [ ] Acceptance Criteria maps to all Verification items

If any gate fails, **STOP**. Complete the ADR+Plan before proceeding.

**After ADR+Plan verified:**
1. Update affected documents (CTX/CON/COM) as specified
2. Regenerate TOC using the c3-toc skill
3. **Proceed to Phase 4: Handoff**

### Phase 4: Handoff (MANDATORY)

**Goal:** Execute post-ADR steps configured in settings.yaml.

**Step 1: Load handoff configuration**

```bash
# Check for settings.yaml
cat .c3/settings.yaml 2>/dev/null | grep -A 20 "^handoff:" || echo "NO_HANDOFF_CONFIG"
```

**Step 2: Determine handoff actions**

<thinking>
Determine handoff approach:
- Does settings.yaml exist? [yes/no]
- Does it have a `handoff:` section? [yes/no]
- If yes: What steps are configured?
- If no: Use default handoff (summarize, list docs, offer tasks)

Handoff steps to execute:
1. [step]
2. [step]
...
</thinking>

| If settings.yaml has... | Then do... |
|------------------------|------------|
| `handoff:` section exists | Execute each step listed |
| No `handoff:` section | Use defaults below |
| No settings.yaml | Use defaults below |

**Default handoff steps (when no config):**
1. Summarize ADR to user
2. List affected documents
3. Ask: "Would you like me to create implementation tasks?"

**Step 3: Execute handoff**

For each step in handoff config (or defaults):
- If "create tasks" → Use vibe_kanban MCP or ask user for task system
- If "notify team" → Ask user how to notify (Slack, email, etc.)
- If "update docs" → Already done in Phase 3
- If custom step → Execute as described

**Step 4: Verify handoff complete**

Confirm with user:
- "ADR created: `.c3/adr/adr-YYYYMMDD-slug.md`"
- "Documents updated: [list]"
- "Handoff actions completed: [list what was done]"

**Phase 4 Gate:** User acknowledges handoff is complete.

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
- **Skipping Phase 3 (ADR+Plan creation)** and updating documents directly.
- **Creating ADR without Implementation Plan** - they are inseparable.
- **Orphan layer changes** - "Changes Across Layers" without corresponding Code Changes.
- **Orphan verifications** - Verification items without corresponding Acceptance Criteria.
- **Vague code locations** - "update auth" instead of `src/handlers/auth.ts:validateToken()`.
- **Skipping Phase 4 (Handoff)** and ending the session without executing settings.yaml steps.
- **Not creating TodoWrite items** for phase tracking.
- **Ignoring settings.yaml** handoff configuration.

## Red Flags & Counters

| Rationalization | Counter |
|-----------------|---------|
| "No time to refresh the TOC, I'll just skim files" | Stop and build/read the TOC first; C3 navigation depends on it. |
| "Examples can keep duplicate IDs, they're just sample data" | IDs must be unique or locate/anchor references break—fix collisions before scoping. |
| "I'll ask the user where to change docs instead of hypothesizing" | Hypothesis bounds exploration and prevents confirmation bias; form it before asking questions. |
| "The scope is clear, I can skip the ADR" | **NO.** ADR is mandatory. It documents the journey and enables review. |
| "I'll just update the docs and mention what I did" | **NO.** ADR first, then doc updates. This is non-negotiable. |
| "I'll add the Implementation Plan later" | **NO.** ADR and Plan are created together. Plan is part of Phase 3 gate. |
| "The code changes are obvious, no need to list them" | **NO.** Explicit mapping enables audit verification. List them. |
| "Handoff is just cleanup, I can skip it" | **NO.** Handoff ensures tasks are created and team is informed. Execute it. |
| "No settings.yaml means no handoff needed" | **NO.** Use default handoff steps. Always confirm completion with user. |

**Red flags that mean you should pause:**
- `.c3/TOC.md` missing or obviously stale.
- Component IDs reused across containers or layers.
- ADR being drafted without notes from hypothesis, exploration, and discovery.
- **Updating C3 documents without an ADR file existing.**
- **ADR without Implementation Plan section.**
- **"Changes Across Layers" count ≠ "Code Changes" count.**
- **"Verification" count ≠ "Acceptance Criteria" count.**
- **Ending the session without executing handoff.**
- **No TodoWrite items for the 4 phases.**

## Common Patterns

### New Feature
Surface → Hypothesis at Container level → Explore up to Context, down to Components → ADR with cross-layer changes

### Bug Fix
Surface → Hypothesis at Component level → Explore if isolated or upstream cause → ADR focused on Component, verify no upstream issues

### Architectural Change
Surface → Hypothesis at Context level → Explore all downstream Containers/Components → Large ADR with many changes

### Refactoring
Surface → Hypothesis at Component level → Explore adjacent components → ADR focused on Component layer

---

## Related

- [adr-template.md](../../references/adr-template.md) - ADR template and sections
- [hierarchy-model.md](../../references/hierarchy-model.md) - C3 layer inheritance
- [v3-structure.md](../../references/v3-structure.md) - Document structure
- [skill-protocol.md](../../references/skill-protocol.md) - Skill conventions
