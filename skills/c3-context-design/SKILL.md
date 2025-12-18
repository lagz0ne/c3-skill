---
name: c3-context-design
description: Use when exploring Context level impact during scoping - system boundaries, actors, cross-container interactions, and high-level concerns
---

# C3 Context Level Exploration

## Overview

Context is the **eagle-eye introduction** to your architecture. Two core jobs:

1. **What containers exist and what are they responsible for?**
2. **How do containers interact with each other?**

**Position:** ROOT (c3-0) | Parent: None | Children: All Containers

As the introduction:
- I provide the MAP of the system
- I define WHO exists (containers) and HOW they talk (protocols)
- I set boundaries that children inherit
- Changes here are RARE but PROPAGATE to all descendants

**Announce:** "I'm using the c3-context-design skill to explore Context-level impact."

**📁 File Location:** Context is `.c3/README.md` - NOT `context/c3-0.md` or any subfolder.

---

## ⛔ MERMAID-ONLY DIAGRAM ENFORCEMENT (MANDATORY)

**Reference:** [diagram-patterns.md](../../references/diagram-patterns.md) - Full harness

**This is non-negotiable:**
- ALL diagrams MUST use Mermaid syntax in ` ```mermaid ` blocks
- ASCII art, Unicode box drawing, text-based flowcharts are PROHIBITED

### Quick Validation

Before finalizing any Context doc:
- [ ] All diagrams are Mermaid
- [ ] No ASCII art anywhere

### Red Flags

🚩 `+---+` boxes or `├──` trees
🚩 `-->` arrows outside Mermaid blocks
🚩 Text describing flow without an actual diagram

---

## The Principle

**Reference:** [core-principle.md](../../references/core-principle.md)

> **Upper layer defines WHAT. Lower layer implements HOW.**

At Context level:
- I define WHAT containers exist and WHY
- Container implements my definitions (WHAT components exist)
- I do NOT define what's inside containers - that's Container's job

**Integrity rule:** Containers cannot exist without being listed in Context.

---

## Include/Exclude

See [defaults.md](./defaults.md) for complete rules.

**Quick reference:**
- **Include:** Container responsibilities, container relationships, connecting points (APIs/events), external actors
- **Exclude:** Component details, internal patterns, implementation, code
- **Litmus:** "Is this about WHY containers exist and HOW they relate to each other?"

---

## Load Settings

Read `.c3/settings.yaml` and merge with `defaults.md`.

```bash
cat .c3/settings.yaml 2>/dev/null
```

---

## Exploration Process

### Phase 1: Load Current Context

```bash
cat .c3/README.md
```

Extract: Container inventory, protocols, actors, boundary

### Phase 2: Analyze Change Impact

| Change Type | Action |
|-------------|--------|
| New/remove container | Delegate to c3-container-design / Audit protocols |
| Protocol change | Update all consumers/providers |
| Boundary change | Full system audit |

### Phase 3: Socratic Discovery

- **Containers:** "What would be separately deployed?"
- **Protocols:** "How do containers talk? What's the contract?"
- **Boundary:** "What's inside vs external?"
- **Actors:** "Who initiates interactions?"

---

## Diagram Requirement

**A container relationship diagram is REQUIRED at Context level.**

Must show: containers, external systems, protocols, actors.

See [diagram-patterns.md](../../references/diagram-patterns.md) for examples.

---

## Template

```markdown
---
id: c3-0
title: [System Name] Overview
---

# [System Name] Overview

## Overview {#c3-0-overview}
[System purpose in 1-2 sentences]

## System Architecture {#c3-0-architecture}

[REQUIRED: Mermaid diagram showing containers and relationships]

` ` `mermaid
flowchart TB
    subgraph System["[System Name]"]
        C1[Container 1 c3-1]
        C2[Container 2 c3-2]
        C3[Container 3 c3-3]
    end

    User((User)) --> C1
    C1 --> C2
    C2 --> C3
` ` `

## Actors {#c3-0-actors}
| Actor | Type | Interacts Via |
|-------|------|---------------|

## Containers {#c3-0-containers}
| Container | ID | Archetype | Responsibility |
|-----------|-----|-----------|----------------|

## Container Interactions {#c3-0-interactions}
| From | To | Protocol | Purpose |
|------|-----|----------|---------|

Note: Use Container names (e.g., "Backend → Database"), NOT component IDs.

## Cross-Cutting Concerns {#c3-0-cross-cutting}
- **Auth:** [approach]
- **Logging:** [approach]
- **Errors:** [approach]
```

---

## Output Format

```xml
<context_exploration_result>
  <changes>
    <change type="[container|protocol|boundary|actor]">[description]</change>
  </changes>

  <downstream_impact>
    <container id="c3-{N}" action="[update|create|remove]">
      <reason>[what this container must do]</reason>
    </container>
  </downstream_impact>

  <delegation>
    <to_skill name="c3-container-design" container_ids="[c3-1, c3-2]"/>
  </delegation>
</context_exploration_result>
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
| "Simplifying" the template for small systems | Small docs grow; structure must be ready | Full template always |

### Red Flags

🚩 Document has sections not in the template
🚩 Template sections missing entirely (not even marked N/A)
🚩 Section order differs from template

### Required Sections (Context)

1. Frontmatter (id, title)
2. Overview
3. System Architecture (with Mermaid diagram)
4. Actors
5. Containers
6. Container Interactions
7. Cross-Cutting Concerns

### Self-Check

- [ ] Did I read the template in this session?
- [ ] Does my output have exactly the template sections, in order?
- [ ] Are missing-content sections marked N/A, not deleted?

### Escape Hatch

User explicitly requests deviation: "Skip the Actors section."

---

## ⛔ REFERENCE LOADING ENFORCEMENT (MANDATORY)

**Rule:** If a reference is mentioned, READ it before using its content.

### Anti-Patterns

| Pattern | Why It's Wrong | Correct Action |
|---------|----------------|----------------|
| "Per core-principle.md..." without reading | You're guessing the content | `cat references/core-principle.md` first |
| "The defaults.md says..." | Defaults may have changed | Read the file |
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
| "Created Context doc" | Write command to `.c3/README.md` visible |
| "Structure is correct" | Validation checklist executed with results |
| "Diagram included" | Mermaid block visible in output |
| "Template followed" | Section-by-section match verified |

### Anti-Patterns

| Pattern | Why It's Wrong | Correct Action |
|---------|----------------|----------------|
| "Context doc complete" (no file ops visible) | No evidence of creation | Show the write command |
| "Following template" (no verification) | Template drift is common | Verify section-by-section |
| "Diagram included" (no mermaid block) | May have used ASCII art | Show the mermaid code |

### Red Flags

🚩 Completion claim without corresponding tool usage
🚩 "Done" without checklist execution
🚩 Describing artifacts that weren't created in this conversation

### Self-Check

- [ ] For each artifact I claim exists, is there evidence of its creation?
- [ ] Did I run the skill's validation checklist?
- [ ] Can a reviewer see proof in this conversation?

### Escape Hatch

None. Unverified completion = not complete.

---

## Checklist

- [ ] Container inventory complete
- [ ] Container responsibilities clear (one sentence each)
- [ ] Protocols documented
- [ ] Actors identified
- [ ] Boundary defined
- [ ] Cross-cutting concerns named
- [ ] Downstream containers identified for delegation
- [ ] **Template fidelity verified** (all sections present, in order)
- [ ] **References loaded** (not assumed from memory)
- [ ] **Output verified** (file creation evidence visible)

---

## Related

- [core-principle.md](../../references/core-principle.md) - The C3 principle
- [defaults.md](./defaults.md) - Context layer rules
- [container-archetypes.md](../../references/container-archetypes.md) - Container types
- [diagram-patterns.md](../../references/diagram-patterns.md) - Diagram guidance
