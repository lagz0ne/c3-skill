---
name: c3-component-design
description: Use when exploring Component level impact during scoping - implementation details, configuration, dependencies, and technical specifics
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

## The Principle

**Reference:** [core-principle.md](../../references/core-principle.md)

> **Upper layer defines WHAT. Lower layer implements HOW.**

At Component level:
- Container defines WHAT I am (my existence, my responsibility)
- I define HOW I implement that responsibility
- Code implements my documentation (in the codebase, not in C3)
- I do NOT invent my responsibility - that's Container's job

**Integrity rules:**
- I must be listed in Container before I can exist
- My "Contract" section must reference what Container says about me
- If Container doesn't mention me, I don't exist as a C3 document

---

## Include/Exclude

See [defaults.md](./defaults.md) for complete rules.

**Quick reference:**
- **Include:** Flows, dependencies, decision logic, edge cases, error scenarios, state/lifecycle
- **Exclude:** WHAT component does (Container), component relationships (Container), code, file paths
- **Litmus:** "Is this about HOW this component implements its contract?"

---

## Load Settings

Read `.c3/settings.yaml` and merge with `defaults.md`.

```bash
cat .c3/settings.yaml 2>/dev/null
```

---

## Integrity Check: Component ↔ Container

**BEFORE proceeding, verify integrity with parent Container.**

```bash
# 1. Load parent Container
cat .c3/c3-{N}-*/README.md

# 2. Check this component is listed
grep "c3-{N}{NN}" .c3/c3-{N}-*/README.md
```

**Integrity requirements:**

| Check | Pass | Fail |
|-------|------|------|
| Component listed in Container inventory | Proceed | STOP - add to Container first |
| Component responsibility matches | Proceed | STOP - align with Container |
| Container archetype identified | Proceed | Ask: "What's the relationship to content?" |

**If component not in Container:** Escalate to c3-container-design to add the component first.

---

## Container Archetype Awareness

**Reference:** [container-archetypes.md](../../references/container-archetypes.md)

The parent Container's archetype shapes what this component documents:

| Container Archetype | Component Documents |
|--------------------|---------------------|
| **Service** | Processing flows, business logic, orchestration |
| **Data** | Structure details, query patterns, migration steps |
| **Boundary** | Integration mechanics, API mapping, resilience |
| **Platform** | Operational procedures, configs, runbooks |

---

## Exploration Process

### Phase 1: Inherit From Container (and Context)

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

### Phase 2: Understand the Contract

```bash
cat .c3/c3-{N}-*/README.md | grep -A5 "c3-{N}{NN}"
```

Extract: responsibility, related components, flows

### Phase 3: Load Current State & Explore Code

```bash
find .c3/c3-{N}-* -name "c3-{N}{NN}-*.md" 2>/dev/null | head -1 | xargs cat 2>/dev/null
```

Check: implementation files, code/doc alignment, affected files, tech constraints

### Phase 4: Analyze Change Impact

| Check | If Yes |
|-------|--------|
| Breaks interface/patterns? | Escalate |
| Needs new tech? | Escalate |
| Affects siblings? | Escalate |
| Implementation only? | Proceed |

### Phase 5: Socratic Discovery

**Confirm integrity:** Listed in Container? Responsibility defined? Archetype identified?

**By container archetype:**
- **Service:** Processing steps? Dependencies? Error paths?
- **Data:** Structure? Queries? Migrations?
- **Boundary:** External API? Mapping? Failures?
- **Platform:** Process? Triggers? Recovery?

---

## Documentation Principles

1. **Implement the contract** - Container says WHAT, Component explains HOW
2. **NO CODE** - Code lives in codebase, not C3 docs
3. **PREFER DIAGRAMS** - A flowchart beats paragraphs
4. **Edge cases and errors** - Document non-obvious behavior

**Component doc:** flows (diagram), decisions, dependencies, edge cases
**.c3/references/:** schemas, code examples, configs, library patterns

---

## Template

```markdown
---
id: c3-{N}{NN}
title: [Component Name]
type: component
parent: c3-{N}
---

# [Component Name]

## Contract
From Container (c3-{N}): "[responsibility from parent container]"

## How It Works

### Flow
[REQUIRED: Mermaid diagram showing processing steps and decisions]

### Dependencies
| Dependency | Container | Purpose |
|------------|-----------|---------|

### Decision Points
| Decision | Condition | Outcome |
|----------|-----------|---------|

## Edge Cases
| Scenario | Behavior | Rationale |
|----------|----------|-----------|

## Error Handling
| Error | Detection | Recovery |
|-------|-----------|----------|

## References
- [Link to .c3/references/ if detailed implementation docs exist]
```

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
- [ ] Component integrity verified (listed in Container)
- [ ] Current state loaded (if exists)
- [ ] Change impact analyzed
- [ ] All inherited constraints still honored
- [ ] Escalation decision made (if needed)

---

## Related

- [core-principle.md](../../references/core-principle.md) - The C3 principle
- [defaults.md](./defaults.md) - Component layer rules
- [container-archetypes.md](../../references/container-archetypes.md) - Container types and component patterns
- [diagram-patterns.md](../../references/diagram-patterns.md) - Diagram guidance
