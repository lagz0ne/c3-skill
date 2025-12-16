---
name: c3-context-design
description: Explore Context level impact during scoping - system boundaries, actors, cross-component interactions, and high-level concerns
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

## Load Settings

Read `.c3/settings.yaml` and merge with `defaults.md`.

```bash
cat .c3/settings.yaml 2>/dev/null
```

**Default litmus:** "Would changing this require coordinating multiple containers or external parties?"

---

## Decision: Is This Context Level?

Context changes are **rare**. Most changes happen at Container or Component level.

**Context-level triggers (ANY):**
- Adding or removing a container
- Changing how containers talk (new/changed protocol)
- Changing system boundary
- Adding a new actor type

**Delegate to Container if (ALL):**
- Change is within existing container
- No new protocols
- Boundary unchanged
- Same actors

---

## Context's Two Core Jobs

### Job 1: Container Inventory

| Container | ID | Type | Responsibility (one sentence) |
|-----------|-----|------|------------------------------|
| [Name] | c3-{N} | Code/Infra | [What it does] |

**Context says:** "These are the boxes"
**Container says:** "Here's what's inside each box"

### Job 2: Container Protocols

| From | To | Protocol | Purpose |
|------|-----|----------|---------|
| c3-1 | c3-2 | [REST/Queue/etc] | [Why] |

**Context says:** "Container A talks to Container B via Protocol X"
**Container says:** "Our IntegrationClient component implements that protocol"

---

## Exploration Process

### Phase 1: Load Current Context

```bash
cat .c3/README.md
```

Extract: Container inventory, protocols, actors, boundary

### Phase 2: Analyze Change Impact

| Change Type | Impact | Action |
|-------------|--------|--------|
| New container | Create doc | Delegate to c3-container-design |
| Remove container | Remove doc | Audit affected protocols |
| New protocol | Both containers | Update both docs |
| Protocol change | All consumers/providers | Coordinate updates |
| Boundary change | All containers | Full audit |

### Phase 3: Document Downstream Impact

```xml
<contract container="c3-{N}">
  <protocol name="[name]" role="[consumer|provider]"/>
  <boundary>
    <can_access>[internal]</can_access>
    <cannot_access>[external]</cannot_access>
  </boundary>
</contract>
```

---

## Socratic Discovery

**Containers:** "What would be separately deployed? What has its own codebase?"
**Protocols:** "How do containers talk? Sync or async? What's the contract?"
**Boundary:** "What's inside vs external? What external systems integrate?"
**Actors:** "Who initiates interactions? Humans? Other systems?"

---

## Template

```markdown
---
id: c3-0
c3-version: 3
title: [System Name] Overview
---

# [System Name] Overview

## Overview {#c3-0-overview}
[System purpose in 1-2 sentences]

## System Boundary {#c3-0-boundary}
### Inside (Our System)
- [Container 1]
- [Container 2]

### Outside (External)
- [External System 1]

## Actors {#c3-0-actors}
| Actor | Type | Interacts Via |
|-------|------|---------------|

## Containers {#c3-0-containers}
| Container | ID | Type | Responsibility |
|-----------|-----|------|----------------|

## Container Interactions {#c3-0-interactions}
| From | To | Protocol | Purpose |
|------|-----|----------|---------|

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

- [core-principle.md](../../references/core-principle.md) - The C3 principle (upper defines WHAT, lower implements HOW)
- [container-archetypes.md](../../references/container-archetypes.md) - Container types and patterns
- [hierarchy-model.md](../../references/hierarchy-model.md) - C3 layer inheritance
- [v3-structure.md](../../references/v3-structure.md) - Document structure
- [diagram-patterns.md](../../references/diagram-patterns.md) - Diagram guidance
