---
name: c3-component-design
description: Use when exploring Component level impact during scoping - implementation details, configuration, dependencies, and technical specifics
---

# C3 Component Level Exploration

## ⛔ CRITICAL GATE: Load Parent Container + Context First

> **STOP** - Before ANY component-level work, execute:
> ```bash
> # Load grandparent Context
> cat .c3/README.md 2>/dev/null || echo "NO_CONTEXT"
> 
> # Load parent Container (REQUIRED - components inherit from here)
> cat .c3/c3-{N}-*/README.md 2>/dev/null || echo "NO_CONTAINER"
> 
> # Check if component is listed in Container
> grep "c3-{N}{NN}" .c3/c3-{N}-*/README.md 2>/dev/null || echo "COMPONENT_NOT_IN_CONTAINER"
> 
> # Load existing component doc (if exists)
> cat .c3/c3-{N}-*/c3-{N}{NN}-*.md 2>/dev/null || echo "NO_COMPONENT_DOC"
> ```

**Based on output:**
- If "NO_CONTAINER" → **STOP.** Container must exist first. Escalate to c3-container-design.
- If "COMPONENT_NOT_IN_CONTAINER" → **STOP.** Add component to Container first.
- If component doc exists → Read it completely before proposing changes

**Why this gate exists:** Components INHERIT from Container (and transitively from Context). A component cannot exist without being listed in its parent Container.

**Self-check before proceeding:**
- [ ] I executed the commands above
- [ ] I read parent Container doc
- [ ] This component IS listed in Container's component inventory
- [ ] I know what contract/responsibility this component has (from Container)
- [ ] I read existing component doc (if exists)

---

## Overview

Component is the **leaf layer** - it inherits all constraints from above and implements actual behavior.

**Position:** LEAF (c3-{N}{NN}) | Parent: Container (c3-{N}) | Grandparent: Context (c3-0)

**📁 File Location:** Component is `.c3/c3-{N}-{slug}/c3-{N}{NN}-{slug}.md` - INSIDE the container folder.

**Announce:** "I'm using the c3-component-design skill to explore Component-level impact."

---

## The Principle

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

| Include (Component Level) | Exclude (Push Up) |
|---------------------------|-------------------|
| Flows (how processing works) | WHAT component does (Container) |
| Dependencies | Component relationships (Container) |
| Decision logic | WHY patterns chosen (Container/ADR) |
| Edge cases | Code snippets |
| Error scenarios | File paths |
| State/lifecycle | |

**Litmus test:** "Is this about HOW this component implements its contract?"

---

## ⛔ NO CODE ENFORCEMENT (MANDATORY)

**Component docs describe HOW things work, NOT the actual implementation.**

### What Counts as Code (PROHIBITED)

| Prohibited | Example | Write Instead |
|------------|---------|---------------|
| Implementation code | `function handle() {...}` | Flow diagram |
| Type definitions | `interface User {...}` | Table: Field \| Type \| Purpose |
| Config snippets | `{ "port": 3000 }` | Table of settings |
| SQL/queries | `SELECT * FROM...` | Access pattern description |
| JSON/YAML schemas | `{ "eventId": "uuid" }` | Table with dot notation |
| Example payloads | Request/response JSON | Table: Field \| Type \| Example |

### Why Mermaid is Allowed but JSON is Not

- **Mermaid** = visual flow/state diagrams (architectural)
- **JSON/YAML** = data structure syntax (implementation)
- **Rule:** If it could be parsed by JSON/YAML parser → use table instead

---

## Exploration Process

### Phase 1: Verify Integrity

From loaded Container, extract:
- Component's responsibility (from Container's component table)
- Related components (siblings)
- Technology constraints
- Pattern constraints

**If component not in Container:** STOP. Escalate to c3-container-design.

### Phase 2: Analyze Change Impact

| Check | If Yes |
|-------|--------|
| Breaks interface/patterns? | Escalate to c3-container-design |
| Needs new tech? | Escalate to c3-container-design |
| Affects siblings? | Coordinate (may need Container update) |
| Implementation only? | Proceed |

### Phase 3: Socratic Discovery

**By container archetype:**
- **Service:** Processing steps? Dependencies? Error paths?
- **Data:** Structure? Queries? Migrations?
- **Boundary:** External API? Mapping? Failures?
- **Platform:** Process? Triggers? Recovery?

---

## Template

```markdown
---
id: c3-{N}{NN}
c3-version: 3
title: [Component Name]
type: component
parent: c3-{N}
summary: >
  [One-line description of what this component does]
---

# [Component Name]

## Contract
From Container (c3-{N}): "[responsibility from parent container]"

## How It Works

### Flow
[REQUIRED: Mermaid diagram showing processing steps and decisions]

` ` `mermaid
flowchart TD
    Start([Request]) --> Validate{Validate?}
    Validate -->|Yes| Process[Process Data]
    Validate -->|No| Error[Return Error]
    Process --> Result([Return Result])
    Error --> Result
` ` `

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

## ⛔ Enforcement Harnesses

### Harness 1: No Code

**Rule:** No code blocks except Mermaid diagrams.

```bash
# Check for non-mermaid code blocks
grep -E '```[a-z]+' .c3/c3-{N}-*/c3-{N}{NN}-*.md | grep -v mermaid
# Should return nothing
```

🚩 **Red Flags:**
- `function`, `class`, `interface`, `type` keywords
- `import`, `require`, `export` statements
- File extensions like `.ts`, `.js`, `.py`
- JSON/YAML blocks (even for "schemas")
- "Example payload" in code blocks

### Harness 2: Template Fidelity

**Rule:** Output MUST match template structure exactly.

**Required sections (in order):**
1. Frontmatter (id, c3-version, title, type, parent, summary)
2. Contract (from parent Container)
3. How It Works
4. Flow (Mermaid diagram REQUIRED)
5. Dependencies
6. Decision Points
7. Edge Cases
8. Error Handling
9. References

🚩 **Red Flags:**
- Missing "Contract" section
- No flow diagram
- Code blocks present

---

## Verification Checklist

Before claiming completion, execute:

```bash
# Verify component doc exists in correct location
ls .c3/c3-{N}-*/c3-{N}{NN}-*.md

# Verify frontmatter
grep -E "^id:|^type:|^parent:" .c3/c3-{N}-*/c3-{N}{NN}-*.md

# Verify flow diagram exists
grep -c '```mermaid' .c3/c3-{N}-*/c3-{N}{NN}-*.md  # Should be >= 1

# Verify NO non-mermaid code blocks
non_mermaid=$(grep -E '```[a-z]+' .c3/c3-{N}-*/c3-{N}{NN}-*.md | grep -v mermaid | wc -l)
echo "Non-mermaid code blocks: $non_mermaid (should be 0)"
```

- [ ] Critical gate executed (Container + Context loaded)
- [ ] Component IS listed in parent Container's inventory
- [ ] "Contract" section references Container's description
- [ ] Template sections present in correct order
- [ ] Flow diagram included (Mermaid)
- [ ] **NO code blocks** (except Mermaid)
- [ ] Tables used instead of JSON/YAML for data structures

---

## Related

- `references/core-principle.md` - The C3 principle
- `defaults.md` - Component layer rules
- `references/container-archetypes.md` - How archetype shapes component docs
- `references/diagram-patterns.md` - Diagram guidance
