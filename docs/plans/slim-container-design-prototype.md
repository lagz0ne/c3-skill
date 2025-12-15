# Slim c3-container-design Prototype

This is a prototype of a slimmed-down c3-container-design skill.

**Original:** 43,571 bytes (~10,900 tokens)
**Target:** ~15,000 bytes (~3,750 tokens) - **65% reduction**

## Changes Made

1. Removed verbose `<extended_thinking>` blocks - replaced with inline comments
2. Extracted settings merge logic → `references/settings-merge.md`
3. Extracted diagram framework → `references/diagram-decision-framework.md`
4. Removed XML-heavy templates - replaced with concise checklists
5. Added explicit "Read more" references instead of inline documentation

---

## Proposed Slim Skill

```markdown
---
name: c3-container-design
description: Explore Container level impact during scoping - technology choices, component organization, middleware, and inter-container communication
---

# C3 Container Level Exploration

## Overview

The Container level is the **architectural command center** of C3:
- **Full context awareness** from above (inherits from Context)
- **Complete control** over component responsibilities below
- **Mediator** for all interactions

**Position:** MIDDLE (c3-{N}) | Parent: Context (c3-0) | Children: Components (c3-{N}NN)

**Announce:** "I'm using the c3-container-design skill to explore Container-level impact."

---

## Load Settings

Read `.c3/settings.yaml` and merge with `defaults.md` per [settings-merge.md](../../references/settings-merge.md).

Display active config:
```
Container Layer Configuration:
├── Include: [merged list]
├── Exclude: [merged list]
├── Litmus: [active test]
└── Diagrams: [tool] - [types]
```

---

## Decision: Is This Container Level?

**Container-level indicators (ANY = yes):**
- Changes technology stack
- Reorganizes component structure
- Modifies internal patterns
- Changes API contracts between components
- Adds/removes components
- Changes cross-container interactions

**Escalate to Context if (ANY):**
- Requires new inter-container protocol
- Changes actor interfaces
- Violates system boundary

**Delegate to Component if (ALL):**
- Single component change
- Tech stack unchanged
- Patterns followed
- Interface unchanged

---

## Phase 1: Inherit From Context

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

---

## Phase 2: Load Current Container

```bash
find .c3 -name "c3-{N}-*" -type d | head -1 | xargs -I {} cat {}/README.md
```

**Must be able to answer:**
- What runtime/framework?
- What are the major components?
- How do components interact internally?
- How does this container interact with others?
- What patterns are established?

---

## Phase 3: Analyze Change Impact

| Direction | Check | Action |
|-----------|-------|--------|
| **Upstream** (Context) | New protocol? Boundary violation? | Escalate |
| **Isolated** (this Container) | Stack/pattern/API/org change? | Document |
| **Adjacent** (siblings) | Component-to-component impact? | Coordinate |
| **Downstream** (Components) | New/changed/removed components? | Delegate |

**Key insight:** Inter-container = Component-to-Component mediated by Container.
- Our component initiates/handles
- Their component responds
- Protocol defined in Context

---

## Phase 4: Diagram Decisions

See [diagram-decision-framework.md](../../references/diagram-decision-framework.md) for full guidance.

**Quick decision:**

| Container Complexity | Diagrams |
|---------------------|----------|
| Simple (1-3 components, linear) | Table only |
| Moderate (4-6 components) | Component flowchart |
| Complex (7+ components) | Flowchart + ONE sequence |

**For each diagram, evaluate:** Clarity added? Reader value? Maintenance cost?

---

## Phase 5: Downstream Contracts

For each affected component, document what Container expects:

```yaml
component: c3-{N}{NN}
inherits_from: c3-{N}
technology:
  runtime: [version]
  framework: [version]
patterns:
  - [pattern name]: [how to implement]
interface:
  exposes: [methods/endpoints]
  accepts: [input types]
  returns: [output types]
```

---

## Output Format

```xml
<container_exploration_result container="c3-{N}">
  <inherited_verification honored="[yes|no]"/>
  <changes>[list of container changes]</changes>
  <adjacent_impact>[sibling containers affected]</adjacent_impact>
  <diagram_decisions>[which diagrams, why]</diagram_decisions>
  <downstream_propagation>[components to update]</downstream_propagation>
  <escalation_needed>[yes|no]</escalation_needed>
</container_exploration_result>
```

---

## Checklist

- [ ] Context constraints loaded and verified
- [ ] Container type determined (Code/Infrastructure)
- [ ] Change impact analyzed (upstream/isolated/adjacent/downstream)
- [ ] Diagram decisions made with justification
- [ ] Downstream contracts documented
- [ ] Delegation list prepared

---

## Templates

### Code Container

```markdown
---
id: c3-{N}
title: [Container Name]
type: code
---

# [Container Name]

## Inherited From Context
[Boundary, protocols, cross-cutting from c3-0]

## Overview
[Single paragraph purpose]

## Technology Stack
- Runtime: [version]
- Framework: [version]

## Component Organization
[Diagram if needed per framework decision]

| Component | ID | Responsibility |
|-----------|-----|----------------|

## Internal Patterns
[Error handling, data access patterns]

## Key Flows
[1-2 critical flows, diagram if non-obvious]

## External Dependencies
| External Container | Our Component | Their Component | Purpose |
```

### Infrastructure Container

```markdown
---
id: c3-{N}
title: [Infrastructure Name]
type: infrastructure
---

# [Infrastructure Name]

## Engine
[Technology] [Version]

## Configuration
| Setting | Value | Why |

## Features Provided
| Feature | Used By | Component |
```

---

## Related

- [settings-merge.md](../../references/settings-merge.md)
- [diagram-decision-framework.md](../../references/diagram-decision-framework.md)
- [hierarchy-model.md](../../references/hierarchy-model.md)
- [archetype-hints.md](../../references/archetype-hints.md)
```

---

## Size Comparison

| Section | Original | Slim | Savings |
|---------|----------|------|---------|
| Settings merge | 2.5KB | 0.3KB (ref) | 88% |
| Extended thinking blocks | 15KB | 0 (removed) | 100% |
| Diagram framework | 14KB | 0.5KB (ref) | 96% |
| Templates | 6KB | 2KB | 67% |
| Core workflow | 6KB | 4KB | 33% |
| **Total** | **43.5KB** | **~7KB** | **84%** |

## Trade-offs

**Pros:**
- Much faster to load and process
- Forces reference docs to be used
- Cleaner separation of concerns
- Easier to maintain

**Cons:**
- Requires reading reference docs for full context
- Less self-contained (depends on refs being available)
- Extended thinking guidance is lost (may reduce reasoning quality)

## Recommendation

**Hybrid approach:**
1. Keep ONE extended_thinking block for the most critical decision (diagram selection)
2. Extract everything else to references
3. Add "Quick Mode" flag in settings.yaml that skips verbose output
4. Target: ~15KB skill file (65% reduction)
