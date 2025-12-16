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

[REQUIRED: Mermaid diagram showing all containers, external systems, and their relationships]

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

## Checklist

- [ ] Container inventory complete
- [ ] Container responsibilities clear (one sentence each)
- [ ] Protocols documented
- [ ] Actors identified
- [ ] Boundary defined
- [ ] Cross-cutting concerns named
- [ ] Downstream containers identified for delegation

---

## Related

- [core-principle.md](../../references/core-principle.md) - The C3 principle
- [defaults.md](./defaults.md) - Context layer rules
- [container-archetypes.md](../../references/container-archetypes.md) - Container types
- [diagram-patterns.md](../../references/diagram-patterns.md) - Diagram guidance
