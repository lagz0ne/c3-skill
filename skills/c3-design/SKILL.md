---
name: c3-design
description: Use when designing system architecture from scratch, updating existing designs, migrating systems, or auditing current architecture - guides through Context→Container→Component methodology with mermaid diagrams and structured documentation
---

# C3 Architecture Design

## Overview

Transform system requirements into structured C3 (Context-Container-Component) architecture documentation through intelligent state detection and phased workflows.

**Core principle:** Detect current state, understand intention through Socratic questioning, scope changes precisely, document decisions via ADR.

**Announce at start:** "I'm using the c3-design skill to guide you through architecture design."

## Quick Reference

| Phase | Key Activities | Socratic? | Output |
|-------|---------------|-----------|--------|
| **1. Understand** | Read `.c3/`, analyze request, infer affected levels | Only if ambiguous | Current state picture |
| **2. Confirm** | Map to C3 structure, present understanding | REQUIRED | Validated intention |
| **3. Scope** | Cross-cutting concerns, boundaries | If uncertain | Change scope |
| **4. ADR** | Document decision with progressive detail | No | ADR in `.c3/adr/` |

## Prerequisites

**Required Location:** `.c3/` directory must exist in project root.

If `.c3/` doesn't exist AND user didn't force mode via `/c3-from-scratch`:
- Stop and inform user to initialize structure first
- Suggest: "Create `.c3/` directory to start, or use `/c3-from-scratch` to initialize"

## The Process

### Phase 1: Understand Current Situation

**Goal:** Build clear picture of what exists and what user wants

**Actions:**

1. **Check for `.c3/` directory**
   ```bash
   if [ ! -d ".c3" ]; then
     echo "Error: .c3/ directory not found. Initialize with /c3-from-scratch or create manually."
     exit 1
   fi
   ```

2. **Read existing documents**
   - List all files: `find .c3 -name "*.md" -type f`
   - Parse document IDs from frontmatter:
     ```bash
     awk '/^---$/,/^---$/ {if (/^id:/) print $2}' .c3/CTX-*.md
     ```
   - Extract status and summaries to understand current state

3. **Analyze user request**
   - Which C3 level does this touch?
     - Context: Cross-component, protocols, deployment, system boundaries
     - Container: Individual container tech/structure/middleware
     - Component: Implementation details, config, dependencies
   - May span multiple levels (e.g., "add authentication" → all three)

4. **Use Socratic questions ONLY if:**
   - Request is ambiguous
   - Scope is unclear
   - Conflicting signals detected

**Output:** Clear understanding of current state + user's goal

### Phase 2: Analyze & Confirm Intention (REQUIRED: Socratic)

**Goal:** Validate understanding before proceeding

**Actions:**

1. **Map request to C3 structure**
   - Determine impacted levels
   - Identify affected containers/components
   - Recognize intention pattern:
     - No docs + "design system" = from-scratch
     - Docs exist + "add X" = update-existing
     - Different structure + "convert" = migrate-system
     - Docs exist + "review" = audit-current

2. **Present understanding concisely**
   ```
   "I see you have [current state]. You want to [goal].
   This will affect:
   - Context level: [impact or none]
   - Container level: [which containers]
   - Component level: [which components]

   Is this understanding correct?"
   ```

3. **Ask targeted Socratic questions**
   - Use AskUserQuestion tool for structured choices
   - Be surgical - don't ask everything
   - Example: If adding auth, ask "Cookie-based or token-based?"

4. **Iterate until aligned**

**Output:** Validated, mutual understanding of what needs to change

### Phase 3: Scoping the Change

**Goal:** Define precise boundaries of impact

**Actions:**

1. **Identify cross-cutting concerns**
   - What touches multiple levels?
   - Examples: Auth spans Context (protocol) → Container (middleware) → Component (JWT handler)

2. **Define boundaries**
   - What will NOT change?
   - Be explicit about out-of-scope items

3. **If scope uncertain:**
   - Use Socratic clarification
   - Present options with AskUserQuestion

**Output:** Clear boundaries - what changes and what doesn't

### Phase 4: Suggest Changes via ADR

**Goal:** Document the architectural decision as planning artifact

**Actions:**

1. **Determine ADR number**
   ```bash
   # Find next ADR number
   last_adr=$(find .c3/adr -name "ADR-*.md" | sed 's/.*ADR-\([0-9]*\).*/\1/' | sort -n | tail -1)
   next_num=$(printf "%03d" $((10#$last_adr + 1)))
   ```

2. **Create ADR with progressive detail**

   File: `.c3/adr/ADR-{NNN}-{slug}.md`

   Template:
   ```markdown
   ---
   id: ADR-{NNN}-{slug}
   title: [Decision Title]
   summary: >
     Documents the decision to [what]. Read this to understand [why],
     what alternatives were considered, and the trade-offs involved.
   status: proposed
   date: YYYY-MM-DD
   related-components: [CON-XXX, COM-YYY]
   ---

   # [ADR-{NNN}] [Decision Title]

   ## Status {#adr-{nnn}-status}
   **Proposed** - YYYY-MM-DD

   ## Context {#adr-{nnn}-context}
   Current situation and why change is needed.

   ## Decision {#adr-{nnn}-decision}

   ### High-Level Approach (Context Level)
   System-wide implications with diagram.

   ### Container Level Details
   Affected containers: [CON-001-backend](../containers/CON-001-backend.md)
   Technology choices and architecture.

   ### Component Level Impact
   New/modified components: [COM-010-new-component](../components/backend/COM-010-new-component.md)

   ## Alternatives Considered {#adr-{nnn}-alternatives}
   What else was considered and why rejected.

   ## Consequences {#adr-{nnn}-consequences}
   Positive, negative, and mitigation strategies.

   ## Cross-Cutting Concerns {#adr-{nnn}-cross-cutting}
   Impacts that span multiple levels.

   ## Implementation Notes {#adr-{nnn}-implementation}
   Ordered steps for implementation.

   ## Related {#adr-{nnn}-related}
   - [Other ADRs]
   - [Affected containers/components]
   ```

3. **Invoke sub-skills as needed**

   Based on ADR scope:
   ```bash
   # If Context level affected
   /c3-context-design

   # If Container level affected
   /c3-container-design

   # If Component level affected
   /c3-component-design
   ```

4. **Regenerate TOC**
   ```bash
   .c3/scripts/build-toc.sh
   ```

**Output:** ADR document + updated/new C3 docs + regenerated TOC

## Iteration & Back-tracking

**You can go backward at any phase:**
- Phase 2 reveals new constraint → Return to Phase 1
- Phase 3 shows fundamental gap → Return to Phase 1
- Phase 4 questions approach → Return to Phase 2

**Don't force forward** when going backward clarifies better.

## Sub-Skill Invocation

Use SlashCommand tool to invoke sub-skills:

```
Use SlashCommand tool with command: "/c3-context-design"
Use SlashCommand tool with command: "/c3-container-design"
Use SlashCommand tool with command: "/c3-component-design"
```

Each sub-skill focuses on its level with appropriate abstraction.

## Key Principles

| Principle | Application |
|-----------|-------------|
| **Detect before ask** | Read `.c3/` first, infer intention, ask only if unclear |
| **Socratic when mandatory** | Phase 2 always, others only if ambiguous |
| **Progressive detail** | ADR starts Context→Container→Component |
| **Unique IDs** | Every document/heading has ID |
| **Regenerate TOC** | After any document changes |
| **Iterate freely** | Go backward when needed |

## Common Patterns

### From Scratch
No `.c3/` exists → Initialize structure → Create Context → Containers → Components

### Update Existing
Docs exist → Detect change scope → Update affected levels → New ADR

### Migrate System
Different structure exists → Map to C3 → Create migration ADR → Progressive conversion

### Audit Current
Docs exist → Review for completeness/accuracy → Identify gaps → Suggest improvements

## Announce Usage

At start of session: "I'm using the c3-design skill to guide you through architecture design."
