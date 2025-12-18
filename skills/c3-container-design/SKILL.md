---
name: c3-container-design
description: Use when exploring Container level impact during scoping - technology choices, component organization, middleware, and inter-container communication
---

# C3 Container Level Exploration

## Overview

The Container level is the **architectural command center** of C3:
- **Full context awareness** from above (inherits from Context)
- **Complete control** over component responsibilities below
- **Mediator** for all interactions

**Position:** MIDDLE (c3-{N}) | Parent: Context (c3-0) | Children: Components (c3-{N}NN)

**Announce:** "I'm using the c3-container-design skill to explore Container-level impact."

**📁 File Location:** Container is `.c3/c3-{N}-{slug}/README.md` - a folder with README inside, NOT `containers/c3-N.md`.

---

## ⛔ MERMAID-ONLY DIAGRAM ENFORCEMENT (MANDATORY)

**Reference:** [diagram-patterns.md](../../references/diagram-patterns.md) - Full harness

**This is non-negotiable:**
- ALL diagrams MUST use Mermaid syntax in ` ```mermaid ` blocks
- ASCII art, Unicode box drawing, text-based flowcharts are PROHIBITED

### Quick Validation

Before finalizing any Container doc:
- [ ] All diagrams are Mermaid
- [ ] No ASCII art anywhere

### Red Flags

🚩 `+---+` boxes or `├──` trees
🚩 `-->` arrows outside Mermaid blocks
🚩 Text describing flow without an actual diagram
🚩 "See diagram below" with ASCII art following

---

## The Principle

**Reference:** [core-principle.md](../../references/core-principle.md)

> **Upper layer defines WHAT. Lower layer implements HOW.**

At Container level:
- Context defines WHAT I am (my existence, my responsibility)
- I define WHAT components exist and WHAT they do
- Component implements my definitions (HOW it works)
- I do NOT define how components work internally - that's Component's job

**Integrity rules:**
- I must be listed in Context before I can exist
- Components cannot exist without being listed in my inventory

---

## Include/Exclude

See [defaults.md](./defaults.md) for complete rules.

**Quick reference:**
- **Include:** Component responsibilities, component relationships, data flows, business flows, inner patterns
- **Exclude:** WHY container exists (Context), HOW components work internally (Component), code
- **Litmus:** "Is this about WHAT components do and HOW they relate to each other?"

---

## Container Archetypes

**Reference:** [container-archetypes.md](../../references/container-archetypes.md)

Different containers have different relationships to content:

| Archetype | Relationship | Typical Components |
|-----------|--------------|-------------------|
| **Service** | Creates/processes | Handlers, Services, Adapters |
| **Data** | Stores/structures | Schema, Indexes, Migrations |
| **Boundary** | Interface to external | Contract, Client, Fallback |
| **Platform** | Operates on containers | CI/CD, Deployment, Networking |

---

## Load Settings

Read `.c3/settings.yaml` and merge with `defaults.md`.

```bash
cat .c3/settings.yaml 2>/dev/null
```

---

## Exploration Process

### Phase 1: Inherit From Context

**ALWAYS START HERE.**

```bash
cat .c3/README.md  # Load Context constraints
```

Extract for this container:
- **Boundary:** What can/cannot access
- **Protocols:** What we implement (provider/consumer)
- **Actors:** Who we serve
- **Cross-cutting:** Patterns we must follow (auth, logging, errors)

**Escalation triggers:** boundary violation, protocol break, actor change, cross-cutting deviation

### Phase 2: Load Current Container

```bash
find .c3 -name "c3-{N}-*" -type d | head -1 | xargs -I {} cat {}/README.md
```

Extract: runtime, components, interactions, patterns

### Phase 3: Analyze Change Impact

| Direction | Action |
|-----------|--------|
| **Upstream** | New protocol/boundary violation → Escalate |
| **Isolated** | Stack/pattern/API/org change → Document |
| **Adjacent** | Component-to-component impact → Coordinate |
| **Downstream** | New/changed components → Delegate |

### Phase 4: Diagram Decisions

**Reference:** [diagram-decision-framework.md](../../references/diagram-decision-framework.md)

Use the framework's Quick Reference table to select diagrams based on container complexity (simple/moderate/complex).

**For each potential diagram, ask:**
1. Does this clarify what prose cannot?
2. Will readers return to this as a "north star"?
3. Is maintenance cost justified?

**Document decision:** INCLUDE / SKIP / SIMPLIFY with one-sentence justification.

### Phase 5: Socratic Discovery

**Identify archetype:** "What is this container's relationship to content?"

**By archetype:**
- **Service:** Responsibility? Key components? Critical flows?
- **Data:** Engine/version? Schema? Access patterns?
- **Boundary:** Contract? Client? Fallback?
- **Platform:** Processes? Affected containers?

---

## Diagram Requirements

**Container level REQUIRES two diagrams:**

1. **External Relationships** - Shows connections to other containers/external systems
2. **Internal Components** - Shows how components relate to each other

See [diagram-patterns.md](../../references/diagram-patterns.md) for examples.

---

## Template

```markdown
---
id: c3-{N}
title: [Container Name]
type: container
parent: c3-0
---

# [Container Name]

## Inherited From Context
- **Boundary:** [what this container can/cannot access]
- **Protocols:** [what protocols this container uses]
- **Cross-cutting:** [patterns inherited from Context]

## Overview
[Single paragraph purpose]

## Technology Stack
| Technology | Version | Purpose |
|------------|---------|---------|

## Architecture

### External Relationships
[REQUIRED: Mermaid diagram showing connections to other containers]

` ` `mermaid
flowchart LR
    ThisContainer[This Container c3-N]
    OtherContainer[Other Container c3-M]
    ExtSystem[External System]

    ThisContainer --> OtherContainer
    ThisContainer --> ExtSystem
` ` `

### Internal Structure
[REQUIRED: Mermaid diagram showing component relationships]

` ` `mermaid
flowchart TD
    subgraph Container["[Container Name] (c3-{N})"]
        C1[Component 1 c3-N01]
        C2[Component 2 c3-N02]
        C3[Component 3 c3-N03]
    end

    C1 --> C2
    C2 --> C3
` ` `

## Components
| Component | ID | Responsibility |
|-----------|-----|----------------|

## Key Flows
[1-2 critical flows - describe WHAT happens, not HOW (that's Component's job)]
```

---

## Output Format

```xml
<container_exploration_result container="c3-{N}">
  <inherited_verification>
    <context_constraints honored="[yes|no]"/>
    <escalation_needed>[yes|no]</escalation_needed>
  </inherited_verification>

  <changes>
    <change type="[stack|pattern|api|organization|component]">
      [Description]
    </change>
  </changes>

  <adjacent_impact>
    <container id="c3-{M}">
      <our_component>[name]</our_component>
      <their_component>[name]</their_component>
      <impact>[description]</impact>
    </container>
  </adjacent_impact>

  <diagram_decisions>
    <diagram type="[type]" include="[yes|no]">
      <reason>[one sentence justification]</reason>
    </diagram>
  </diagram_decisions>

  <downstream_propagation>
    <component id="c3-{N}{NN}" action="[update|create|remove]">
      <inherited_change>[what component must do]</inherited_change>
    </component>
  </downstream_propagation>

  <delegation>
    <to_skill name="c3-context-design" if="[escalation needed]"/>
    <to_skill name="c3-component-design" components="[list]"/>
  </delegation>
</container_exploration_result>
```

---

---

## ⛔ TEMPLATE FIDELITY ENFORCEMENT (MANDATORY)

**Rule:** Output documents MUST match the skill's template structure exactly.

### Anti-Patterns

| Pattern | Why It's Wrong | Correct Action |
|---------|----------------|----------------|
| Adding creative sections not in template | Breaks consistency, confuses users | Stick to template sections |
| Omitting "optional" template sections | They're optional content, not optional structure | Include section, mark N/A if empty |
| Reordering template sections | Users expect consistent navigation | Maintain template order |
| "Simplifying" the template for small containers | Small docs grow; structure must be ready | Full template always |

### Red Flags

🚩 Document has sections not in the template
🚩 Template sections missing entirely (not even marked N/A)
🚩 Section order differs from template
🚩 Missing REQUIRED diagrams (External Relationships, Internal Structure)

### Required Sections (Container)

1. Frontmatter (id, title, type, parent)
2. Inherited From Context
3. Overview
4. Technology Stack
5. Architecture (External Relationships diagram - REQUIRED)
6. Architecture (Internal Structure diagram - REQUIRED)
7. Components
8. Key Flows

### Self-Check

- [ ] Did I read the template in this session?
- [ ] Does my output have exactly the template sections, in order?
- [ ] Are missing-content sections marked N/A, not deleted?
- [ ] Are BOTH required diagrams present (External + Internal)?

### Escape Hatch

User explicitly requests deviation: "Skip the Technology Stack section."

---

## ⛔ REFERENCE LOADING ENFORCEMENT (MANDATORY)

**Rule:** If a reference is mentioned, READ it before using its content.

### Anti-Patterns

| Pattern | Why It's Wrong | Correct Action |
|---------|----------------|----------------|
| "Per core-principle.md..." without reading | You're guessing the content | `cat references/core-principle.md` first |
| "The container-archetypes.md says..." | Archetypes may have updated | Read the file |
| "Following diagram-decision-framework.md..." | You need current guidance | Read the file |
| Citing reference by memory from previous session | Context may have compacted | Re-read in this session |

### Red Flags

🚩 Quoting a reference file without a `cat` or Read command preceding it
🚩 "As documented in X.md" without file content visible in conversation
🚩 Paraphrasing reference content that wasn't loaded this session

### Self-Check

- [ ] For each reference I cite, is there a file read in this conversation?
- [ ] Am I working from actual file content, not memory?

### Escape Hatch

If reference was read earlier in THIS conversation and context hasn't compacted, re-reading is optional.

---

## ⛔ OUTPUT VERIFICATION ENFORCEMENT (MANDATORY)

**Rule:** Claiming completion requires verification evidence in the conversation.

### Verification Requirements

| Claim | Required Evidence |
|-------|-------------------|
| "Created Container doc" | Write command to `.c3/c3-{N}-*/README.md` visible |
| "Structure is correct" | Validation checklist executed with results |
| "Diagrams included" | Both Mermaid blocks visible in output |
| "Template followed" | Section-by-section match verified |
| "Delegated to c3-component-design" | Skill tool invocation visible |

### Anti-Patterns

| Pattern | Why It's Wrong | Correct Action |
|---------|----------------|----------------|
| "Container doc complete" (no file ops visible) | No evidence of creation | Show the write command |
| "Following template" (no verification) | Template drift is common | Verify section-by-section |
| "Both diagrams included" (only one visible) | May have forgotten one | Verify both mermaid blocks |
| "Components delegated" (no skill invocation) | Hallucination | Show Skill tool usage |

### Red Flags

🚩 Completion claim without corresponding tool usage
🚩 "Done" without checklist execution
🚩 Describing artifacts that weren't created in this conversation
🚩 Only one diagram when two are REQUIRED

### Self-Check

- [ ] For each artifact I claim exists, is there evidence of its creation?
- [ ] Did I run the skill's validation checklist?
- [ ] Can a reviewer see proof in this conversation?
- [ ] Are BOTH required diagrams present?

### Escape Hatch

None. Unverified completion = not complete.

---

## Checklist

- [ ] Context constraints loaded and verified
- [ ] Container archetype determined
- [ ] Change impact analyzed (upstream/isolated/adjacent/downstream)
- [ ] Diagram decisions made with justification
- [ ] Downstream contracts documented
- [ ] Delegation list prepared
- [ ] **Template fidelity verified** (all sections present, in order, both diagrams)
- [ ] **References loaded** (not assumed from memory)
- [ ] **Output verified** (file creation evidence visible)

---

## Related

- [core-principle.md](../../references/core-principle.md) - The C3 principle
- [defaults.md](./defaults.md) - Container layer rules
- [container-archetypes.md](../../references/container-archetypes.md) - Container types
- [diagram-decision-framework.md](../../references/diagram-decision-framework.md) - Full diagram guidance
