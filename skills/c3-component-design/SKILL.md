---
name: c3-component-design
description: Use when exploring Component level impact during scoping - implementation details, configuration, dependencies, and technical specifics
---

# C3 Component Level Exploration

## Overview

Component is the **leaf layer** - it inherits all constraints from above and implements actual behavior.

**Position:** LEAF (c3-{N}{NN}) | Parent: Container (c3-{N}) | Grandparent: Context (c3-0)

---

## ⛔ NO CODE ENFORCEMENT (MANDATORY)

**This is non-negotiable. Component docs describe HOW things work, not the actual implementation.**

### What Counts as Code (PROHIBITED)

| Prohibited | Example | Why Prohibited |
|------------|---------|----------------|
| Implementation code | `function handle() { ... }` | Code lives in codebase |
| Type definitions | `interface User { id: string }` | Types change, adds context load |
| Config snippets | `{ "port": 3000, ... }` | Config is implementation detail |
| SQL/queries | `SELECT * FROM users` | Query syntax is implementation |
| Shell commands | `npm run build` | Operational, not architectural |
| **JSON/YAML schemas** | `{ "eventId": "uuid", "type": "..." }` | Schema syntax is implementation |
| **Example payloads** | `{ "user": { "name": "..." } }` | Payload structure is implementation |
| **Pseudocode schemas** | `{ field: type, nested: { } }` | Still code syntax |
| **Wire format examples** | Request/response JSON bodies | Protocol detail is implementation |

### What to Write Instead (REQUIRED)

| Instead of Code | Write This |
|-----------------|------------|
| Function implementation | Flow diagram showing steps |
| Type definitions | Table: Field \| Type \| Purpose \| Required |
| Config snippets | Table of settings and their effects |
| SQL queries | Access pattern description |
| API handlers | Request/response contract table |
| **JSON/YAML schemas** | Table: Field \| Type \| Purpose (use dot notation for nesting: `user.address.city`) |
| **Example payloads** | Table with Field \| Type \| Example Value columns |
| **Nested structures** | Flatten with dot notation: `parent.child.field` |

### Validation Checklist (RUN BEFORE FINALIZING)

Before completing any component doc, verify:

- [ ] **No triple-backtick code blocks** (except mermaid diagrams)
- [ ] **No inline code showing implementation** (variable names, function bodies)
- [ ] **No file paths to source code** (paths change with refactoring)
- [ ] **No language-specific syntax** (TypeScript, SQL, JSON config)

### Red Flags - STOP and Rewrite If You See

🚩 `function`, `class`, `interface`, `type` keywords
🚩 `=>` arrow functions or `{ }` code blocks
🚩 `import`, `require`, `export` statements
🚩 File extensions like `.ts`, `.js`, `.py`, `.go`
🚩 Package names like `express`, `socket.io`, `yjs`
🚩 Variable declarations or assignments
🚩 **"Example payload"** or **"sample request/response"** in code blocks
🚩 **`{ field: type }`** pseudocode notation
🚩 **"Wire format"** or **"protocol structure"** as justification for JSON
🚩 **Any nested JSON/YAML** regardless of purpose

### Why Mermaid is Allowed but JSON is Not

- **Mermaid** = visual flow/state diagrams (architectural, shows relationships)
- **JSON/YAML** = data structure syntax (implementation detail)
- **Rule:** If it could be parsed by a JSON/YAML parser, it's code → use a table instead

### Where Code DOES Belong

If implementation details are **truly necessary** for understanding:

```
.c3/references/[component-name]-implementation.md
```

Reference it from the component doc: `See [implementation details](../references/...)`

**⚠️ references/ is NOT an escape hatch for schemas.** Use it for:
- Complex implementation patterns that genuinely need code
- Library-specific configuration examples
- **NOT** for schema documentation that fits in a table

---

As the leaf:
- I INHERIT from Container: technology, patterns, interfaces
- I INHERIT from Context (via Container): boundary, protocols, cross-cutting
- I implement HOW things work
- Changes here are CONTAINED unless they break inherited contracts

**Announce:** "I'm using the c3-component-design skill to explore Component-level impact."

**📁 File Location:** Component is `.c3/c3-{N}-{slug}/c3-{N}{NN}-{slug}.md` - INSIDE the container folder, NOT `components/c3-NNN.md`.

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
2. **⛔ NO CODE** - See [NO CODE ENFORCEMENT](#-no-code-enforcement-mandatory) above. This is MANDATORY.
3. **PREFER DIAGRAMS** - A flowchart beats paragraphs. Mermaid is allowed.
4. **Edge cases and errors** - Document non-obvious behavior

**Component doc:** flows (diagram), decisions, dependencies, edge cases
**.c3/references/:** schemas, code examples, configs, library patterns (code goes HERE, not in component docs)

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
- [ ] **⛔ NO CODE validation passed** (no code blocks, no implementation syntax, no file paths)

---

## Related

- [core-principle.md](../../references/core-principle.md) - The C3 principle
- [defaults.md](./defaults.md) - Component layer rules
- [container-archetypes.md](../../references/container-archetypes.md) - Container types and component patterns
- [diagram-patterns.md](../../references/diagram-patterns.md) - Diagram guidance
