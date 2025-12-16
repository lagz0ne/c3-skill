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
[REQUIRED: Diagram showing connections to other containers and external systems]

### Internal Structure
[REQUIRED: Diagram showing how components relate to each other]

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

## Checklist

- [ ] Context constraints loaded and verified
- [ ] Container archetype determined
- [ ] Change impact analyzed (upstream/isolated/adjacent/downstream)
- [ ] Diagram decisions made with justification
- [ ] Downstream contracts documented
- [ ] Delegation list prepared

---

## Related

- [core-principle.md](../../references/core-principle.md) - The C3 principle
- [defaults.md](./defaults.md) - Container layer rules
- [container-archetypes.md](../../references/container-archetypes.md) - Container types
- [diagram-decision-framework.md](../../references/diagram-decision-framework.md) - Full diagram guidance
