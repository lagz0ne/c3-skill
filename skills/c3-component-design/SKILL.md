---
name: c3-component-design
description: Explore Component level impact during scoping - implementation details, configuration, dependencies, and technical specifics
---

# C3 Component Level Exploration

## Overview

Component is the **leaf layer** - it inherits all constraints from above and implements actual behavior.

**Position:** LEAF (c3-{N}{NN}) | Parent: Container (c3-{N}) | Grandparent: Context (c3-0)

As the leaf:
- I INHERIT from Container: technology, patterns, interfaces
- I INHERIT from Context (via Container): boundary, protocols, cross-cutting
- I implement HOW things work
- Changes here are CONTAINED unless they break inherited contracts

**Announce:** "I'm using the c3-component-design skill to explore Component-level impact."

---

## Load Settings

Read `.c3/settings.yaml` and merge with `defaults.md`.

```bash
cat .c3/settings.yaml 2>/dev/null
```

**Default litmus:** "Could a developer implement this from the documentation?"

---

## Decision: Is This Component Level?

**Component-level indicators (ALL must be true):**
- Change is about HOW to implement (not WHAT)
- Stays within Container's technology constraints
- Follows Container's patterns
- Keeps interface contract unchanged
- Isolated to this component

**Escalate to Container if (ANY):**
- Affects multiple components the same way
- Requires changing internal patterns
- Needs different technology
- Changes interface contract
- Is about WHAT/WHY not HOW

**Common "looks like Component, but isn't":**
- "Just add a new field to the API" → Interface change (Container)
- "Just change how auth works here" → Cross-cutting (Context)
- "Just use a different library" → Technology change (Container)

---

## Phase 1: Inherit From Container (and Context)

**ALWAYS START HERE.**

```bash
cat .c3/README.md  # Context constraints
find .c3 -name "c3-{N}-*" -type d | head -1 | xargs -I {} cat {}/README.md  # Container constraints
```

**Extract inheritance chain:**

| Source | What to Extract |
|--------|-----------------|
| Context | Boundary, cross-cutting patterns |
| Container | Technology, patterns, interface contract |

**Escalation triggers:** boundary violation, cross-cutting deviation, tech constraint violation, pattern deviation, interface change

---

## Phase 2: Determine Component Nature

See **Nature** section in `defaults.md` for indicators and focus areas.

Use decision questions:
1. External resources? -> **Resource**
2. Business rules? -> **Business**
3. Request handling? -> **Framework**
4. Shared utility? -> **Cross-cutting**

Nature determines which template to use in Phase 5.

---

## Phase 3: Load Current State & Explore Code

```bash
find .c3/c3-{N}-* -name "c3-{N}{NN}-*.md" 2>/dev/null | head -1 | xargs cat 2>/dev/null
```

**Code exploration checklist:**
- [ ] What files implement this component?
- [ ] Does current code match documentation?
- [ ] Which files would the change affect?
- [ ] Can this be done within Container's technology?

---

## Phase 4: Analyze Change Impact

| Check | Question | If Yes |
|-------|----------|--------|
| Breaks interface? | Does Container need to know? | Escalate |
| Breaks patterns? | Does it deviate from Container's patterns? | Escalate |
| Needs new tech? | Outside Container's tech stack? | Escalate |
| Affects siblings? | Multiple components same way? | Escalate |
| Affects boundary? | Outside system boundary? | Escalate (Context) |
| Contained? | Just implementation details? | Proceed |

---

## Phase 5: Document Understanding

### Documentation Principles

1. **No code** - Code lives in the codebase and changes frequently
2. **Explain behavior** - How it works conceptually, not implementation details
3. **Enable understanding** - Reader should understand before making changes
4. **Diagrams over code** - Flowcharts and tables convey behavior better than snippets

### Template by Nature

**Resource Component:**
```markdown
---
id: c3-{N}{NN}
title: [Resource Name]
nature: resource
---

# [Resource Name]

## Inherited Constraints
From Container: [technology, patterns, interface]
From Context: [boundary, cross-cutting]

## Overview
[What this resource does and why it exists]

## Connection Behavior
[How connections are managed - pooling strategy, lifecycle, health checks]

## Retry Strategy
[How failures are handled - retry logic, backoff, circuit breaker]

## Failure Modes
| Failure | Detection | Recovery |
```

**Business Component:**
```markdown
---
id: c3-{N}{NN}
title: [Service Name]
nature: business
---

# [Service Name]

## Inherited Constraints
[Same as above]

## Overview
[What business problem this solves]

## Domain Flow
[Mermaid flowchart showing the conceptual flow]

## Business Rules
| Rule | Condition | Outcome |

## Edge Cases
| Case | Behavior | Rationale |
```

**Framework Component:**
```markdown
---
id: c3-{N}{NN}
title: [Handler Name]
nature: framework
---

# [Handler Name]

## Inherited Constraints
[Same as above]

## Overview
[What requests this handles and why]

## Request Flow
[Mermaid flowchart showing request processing]

## Validation Logic
| Input | Validation | On Failure |

## Error Handling
| Scenario | Response | Rationale |
```

---

## Socratic Discovery

**For Resource:** "What does this connect to? What env vars? What errors?"
**For Business:** "What problem does this solve? What rules? What edge cases?"
**For Framework:** "What endpoints? What auth? What request/response format?"

---

## Output Format

```xml
<component_exploration_result component="c3-{N}{NN}">
  <inherited_verification>
    <container_constraints honored="[yes|no]"/>
    <context_constraints honored="[yes|no]"/>
    <escalation_needed>[yes|no]</escalation_needed>
  </inherited_verification>

  <changes>
    <change type="[config|behavior|dependency]">[description]</change>
    <contained>[yes|no]</contained>
  </changes>

  <sibling_impact>
    <component id="c3-{N}{MM}" impact="[none|needs_update]"/>
  </sibling_impact>

  <delegation>
    <to_skill name="c3-container-design" if="[escalation needed]"/>
  </delegation>
</component_exploration_result>
```

---

## Checklist

- [ ] Container constraints loaded and understood
- [ ] Context constraints loaded (via Container)
- [ ] Component nature determined
- [ ] Current state loaded (if exists)
- [ ] Change impact analyzed
- [ ] All inherited constraints still honored
- [ ] Documentation matches component nature
- [ ] Escalation decision made (if needed)

---

## Related

- [hierarchy-model.md](../../references/hierarchy-model.md) - C3 layer inheritance
- [role-taxonomy.md](../../references/role-taxonomy.md) - Component roles
- [testing-discovery.md](../../references/testing-discovery.md) - Test patterns
- [diagram-patterns.md](../../references/diagram-patterns.md) - Diagram guidance
