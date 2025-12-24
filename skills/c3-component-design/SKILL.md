---
name: c3-component-design
description: Use when documenting component implementation patterns, internal structure, or hand-off points - enforces NO CODE rule and diagram-first approach for leaf-level C3 documentation
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

**⚠️ DO NOT read ADRs** unless user specifically asks:
- Component work focuses on HOW, not historical WHY
- ADRs add unnecessary context → hallucination risk
- Only read: Context, parent Container, sibling Components (if dependencies)

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

> See `references/core-principle.md` for full details.
>
> **Upper layer defines WHAT. Lower layer implements HOW.**

At Component: Container defines WHAT I am. I define HOW I implement it. Code implements my documentation. I must be listed in Container; my Contract references what Container says about me.

---

## Component Boundaries (Context Loading Units)

> **Components are stepping stones for loading context, not code documentation.**
>
> Group what you'd want to load together. Split what you'd want to load separately.

A component doc is a **unit of context**. When someone needs to understand one concern, they load that component. The boundary question is: "Would I want this context loaded together?"

### Boundary Litmus Test

| Question | Decision |
|----------|----------|
| Would someone working on X need to understand Y? | If NO → separate docs |
| Do X and Y change for the same reason? | If NO → separate docs |
| Is this a technology choice (library/framework)? | → Container tech stack, NOT a component |
| Would loading this together save navigation? | If YES and concerns are related → group |

### The Economic Tradeoff

**Too grouped:** "To understand TaskForm, I must load Header/Sidebar/TaskCard context" → wasted cognitive load

**Too split:** "50 tiny component docs to navigate" → navigation overhead

**Right balance:** Each doc = context you'd naturally want together

### What Should NOT Be a Component Doc

| Don't Document As Component | Why | Where It Belongs |
|-----------------------------|-----|------------------|
| Logger/Logging library | Technology choice, not business concern | Container → Tech Stack table |
| Config parsing library | DI framework detail | Container → Tech Stack + patterns |
| Database driver | Technology choice | Container → Tech Stack |
| HTTP framework (Hono, Express) | Foundation documented at Container | Container → Internal Structure diagram |
| Generic utilities | Not a business concern | No doc needed |

### What SHOULD Be a Component Doc

| Document As Component | Why |
|-----------------------|-----|
| Renderer service | Has unique business logic, clear IN/OUT, specific conventions |
| Cache service | Business rules (TTL, eviction), not just "we use Redis" |
| Queue/backpressure | Business constraints (limits, behavior), not just "semaphore" |
| API flow orchestration | Coordinates multiple concerns, has hand-offs |
| External integration | Mapping, error handling, retry logic |

### Foundation vs Business Components

Both are valid component docs, but they serve different purposes:

| Aspect | Foundation Component | Business Component |
|--------|---------------------|-------------------|
| **Purpose** | What it PROVIDES to consumers | HOW it implements logic |
| **Audience** | Business components using it | Maintainers of this component |
| **Content** | Interface + conventions for consumers | Processing + hand-offs + edge cases |
| **Examples** | Logger, Config, HTTP Framework | Renderer, Cache, Queue, Flows |

**Foundation component doc answers:**
- "What does this provide to business components?"
- "What conventions must consumers follow?"
- "What's the hand-off interface?"

**Foundation component doc does NOT include:**
- Implementation details ("we use pino")
- Configuration code
- Library-specific patterns

**Example - Logger as Foundation:**
- ✅ Interface: provides Logger instance to all business components
- ✅ Conventions: structured fields (requestId, component), log levels (when to use each)
- ✅ Hand-offs: what business components receive and how to use it
- ❌ NOT: "here's the pino config code"

### Technology Choice (No Doc Needed)

Pure technology choice = just "we use X library". No unique conventions, no hand-off complexity.

- If there are conventions for consumers → Foundation component
- If it's just "we use X" → Container Tech Stack row only

---

## Include/Exclude

> See `defaults.md` for full include/exclude rules and litmus test.

**Quick check:** "Is this about HOW this component implements its contract?"

---

## Diagram-First Principle

**Components are documented visually.** The diagram tells the story, text supports it.

Every component explains:
- **Its boundary** - what it's responsible for (and what it's not)
- **Hand-off points** - where information exchanges happen with other components
- **Internal processing** - what happens inside (self-contained)

### Interface Diagram (Required)

Shows the boundary and hand-offs:

```mermaid
flowchart LR
    subgraph IN["Receives"]
        I1[From Component A]
        I2[From External]
    end

    subgraph SELF["Owns"]
        S1[Processing]
        S2[Transformation]
    end

    subgraph OUT["Provides"]
        O1[To Component B]
        O2[To Caller]
    end

    IN --> SELF --> OUT
```

### Additional Diagrams (Where Applicable)

Use multiple diagrams when needed - just enough to explain:

| Diagram Type | Use When |
|--------------|----------|
| **Sequence** | Hand-off between multiple components matters |
| **State** | Component has lifecycle/state transitions |
| **Organization** | Internal structure is complex (layers, subsystems) |

**Foundation vs Business components:**
- **Foundation** (e.g., Hono): Shows the transition point - where framework ends and business begins. What it provides TO business components.
- **Business** (e.g., User Handler): Shows the flow - receives from foundation, processes, hands off results.

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

See `component-template.md` for complete structure with frontmatter, diagrams, and examples.

**Required sections:**
1. Contract (from parent Container)
2. Interface (diagram showing boundary - REQUIRED)
3. Hand-offs table
4. Conventions table
5. Edge Cases & Errors table

**Optional sections (include based on component nature):**
- Additional diagrams (Sequence, State, Organization)
- Configuration, Dependencies, Invariants, Performance

**Keep it lean:** Simple component = 5 sections. Complex framework = multiple diagrams + optional sections.

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

### Harness 2: Diagram-First

**Rule:** Interface diagram (IN → Processing → OUT) is REQUIRED.

**Required sections:**
1. Frontmatter (id, c3-version, title, type, parent, summary)
2. Contract (from parent Container)
3. Interface (Mermaid diagram REQUIRED - boundary and hand-offs)
4. Hand-offs (table - what exchanges with whom)
5. Conventions (table of rules)
6. Edge Cases & Errors (table)

**Optional sections (include based on component nature):**
- Additional diagrams (Sequence, State, Organization) - where needed
- Configuration - significant config surface
- Dependencies - external dependencies matter
- Invariants - key guarantees to verify
- Performance - throughput/latency matters

🚩 **Red Flags:**
- Missing Interface diagram
- No IN/OUT structure in diagram
- Text-heavy without diagrams
- Code blocks present

---

## Verification Checklist

Before claiming completion, execute:

```bash
# Verify component doc exists in correct location
ls .c3/c3-{N}-*/c3-{N}{NN}-*.md

# Verify frontmatter
grep -E "^id:|^type:|^parent:" .c3/c3-{N}-*/c3-{N}{NN}-*.md

# Verify Interface diagram exists with IN/OUT structure
grep -c '```mermaid' .c3/c3-{N}-*/c3-{N}{NN}-*.md  # Should be >= 1
grep -E "subgraph.*(IN|OUT)" .c3/c3-{N}-*/c3-{N}{NN}-*.md  # Should find IN/OUT

# Verify NO non-mermaid code blocks
non_mermaid=$(grep -E '```[a-z]+' .c3/c3-{N}-*/c3-{N}{NN}-*.md | grep -v mermaid | wc -l)
echo "Non-mermaid code blocks: $non_mermaid (should be 0)"
```

- [ ] Critical gate executed (Container + Context loaded)
- [ ] Component IS listed in parent Container's inventory
- [ ] **Context unit verified** - would I want this context loaded together?
- [ ] **Not over-grouped** - no unrelated concerns forcing unnecessary context load
- [ ] **Not a tech choice** - if just "we use X library", belongs in Container tech stack
- [ ] "Contract" section references Container's description
- [ ] **Interface diagram present** (boundary and hand-offs)
- [ ] **Hand-offs table present** (what exchanges with whom)
- [ ] **Conventions table present** (rules for consistency)
- [ ] Edge Cases & Errors table present
- [ ] **NO code blocks** (except Mermaid)
- [ ] **Additional diagrams justified** - only if Sequence/State/Organization needed
- [ ] **Optional sections justified** - only included if relevant to this component

---

## 📚 Reading Chain Output

**At the end of component work, output a reading chain for related docs.**

Format:
```
## 📚 To Go Deeper

This component (c3-NNN) is part of:

**Ancestors (understand constraints):**
└─ c3-0 (Context) → c3-N (Container) → c3-NNN (this)

**Sibling components (if dependencies):**
├─ c3-NMM - [why this sibling matters]
└─ c3-NKK - [dependency relationship]

**Parent container (for context):**
└─ c3-N-{slug} - [what contract this component fulfills]

*Reading chain generated from component's dependencies and parent.*
```

**Rules:**
- List siblings only if this component depends on or affects them
- Always include parent Container for contract reference
- Include Context only if cross-cutting concerns are relevant
- Never include ADRs unless user asked

---

## When NOT to Use

**Escalate to c3-container-design if:**
- Change affects component relationships (not just one component's internals)
- New technology/pattern needed
- Component not yet listed in parent Container

**Escalate to c3-context-design if:**
- Cross-cutting concern (affects multiple containers)
- New protocol or boundary needed

**Wrong layer symptoms:**
- Describing WHAT component does → Container's job
- Documenting WHY component exists → Container's job
- Including cross-container flows → Context's job

---

## Change Impact Decision

```mermaid
flowchart TD
    A[Analyze Change] --> B{Breaks interface<br/>or patterns?}
    B -->|yes| C[Escalate to<br/>c3-container-design]
    B -->|no| D{Needs new<br/>technology?}
    D -->|yes| C
    D -->|no| E[Document at<br/>Component level]
```

---

## Common Rationalizations

| Excuse | Reality |
|--------|---------|
| "JSON is clearer than a table" | JSON is code. Use table with Field \| Type \| Example columns. |
| "Just one example payload won't hurt" | Examples become implementation. Describe structure, don't show syntax. |
| "Code is easier to understand" | Code is implementation. Component doc describes patterns, not implements. |
| "Type definitions help readers" | Types are code. Use tables with Field, Type, Purpose columns. |
| "I'll use a simple config snippet" | Config syntax is code. Table: Setting \| Default \| Purpose. |
| "Mermaid is code too" | Mermaid is architectural diagram (allowed). JSON/YAML is data structure (prohibited). |
| "Interface diagram not needed for simple component" | REQUIRED. Even simple components have IN → OUT boundary. |
| "Every atom/service needs a component doc" | Only if it's a context-loading unit. Logger is a tech choice → Container tech stack. |
| "I should document all the pieces" | Document context units, not implementation units. Ask: "Would I load this together?" |
| "The flow is already in another component" | If flow is documented elsewhere, don't duplicate. Reference it. |
| "I'll group related things together" | Related by what? If they change independently, they're separate context units. |
| "One doc per code file/module" | Code structure ≠ context structure. Group by what you'd want to understand together. |

---

## Common Mistakes

| Mistake | Fix |
|---------|-----|
| **Including code blocks** | Use tables. Field \| Type \| Purpose for structures. |
| **Skipping Interface diagram** | REQUIRED. Shows IN → Processing → OUT boundary. |
| **Not checking Container first** | Component must be listed in parent Container inventory. |
| **Text-heavy docs** | Lead with diagrams. Text supports visuals, not vice versa. |
| **Missing Hand-offs table** | REQUIRED. Shows what exchanges with whom. |
| **Describing WHAT not HOW** | WHAT is Container's job. Component explains HOW. |
| **Creating doc for every code unit** | Only create for context units. Tech choices → Container. |
| **Over-grouping unrelated concerns** | If working on X doesn't need Y context, split them. |
| **Documenting tech choice as component** | Logger, config lib → Container tech stack, not component doc. |
| **Grouping by code structure** | Group by context need, not by file/folder structure. |

---

## Related

- `references/core-principle.md` - The C3 principle
- `defaults.md` - Component layer rules
- `references/container-archetypes.md` - How archetype shapes component docs
- `references/diagram-patterns.md` - Diagram guidance
