# C3 Hierarchy Model

The C3 hierarchy (Context-Container-Component) defines clear boundaries and inheritance patterns. Each layer inherits constraints from above and defines contracts for below.

## The Three Layers

```
┌─────────────────────────────────────────────────────────────────┐
│  CONTEXT (c3-0)                                                 │
│  ROOT - Defines system-wide contracts                           │
│  • System boundary (what's in/out)                              │
│  • Actors (who interacts)                                       │
│  • Protocols (how containers communicate)                       │
│  • Cross-cutting concerns (auth, logging, errors)               │
│                              │                                  │
│                              ▼ PROPAGATES TO                    │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  CONTAINERS (c3-1, c3-2, c3-3, ...)                      │   │
│  │  MIDDLE - Inherits from Context, defines for Components  │   │
│  │  • Technology stack (runtime, framework)                 │   │
│  │  • Component organization                                │   │
│  │  • Internal patterns & conventions                       │   │
│  │  • API contracts (what components expose)                │   │
│  │                              │                           │   │
│  │                              ▼ PROPAGATES TO             │   │
│  │  ┌────────────────────────────────────────────────────┐  │   │
│  │  │  COMPONENTS (c3-101, c3-102, c3-201, ...)          │  │   │
│  │  │  LEAF - Inherits all, implements actual behavior   │  │   │
│  │  │  • HOW things work (the actual code)               │  │   │
│  │  │  • Configuration details                           │  │   │
│  │  │  • Error handling specifics                        │  │   │
│  │  │  • Usage patterns                                  │  │   │
│  │  └────────────────────────────────────────────────────┘  │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Inheritance Flow

| From | To | What's Inherited |
|------|-----|------------------|
| **Context** | All Containers | Boundary, actors, protocols, cross-cutting |
| **Container** | Its Components | Technology stack, patterns, interface contracts |
| **Context** | Components (via Container) | System-wide constraints flow through |

## Layer Responsibilities

### Context (c3-0) - ROOT

Defines WHAT exists and HOW they relate. Changes here propagate to ALL descendants.

| Contract Type | What Context Decides | What Children Must Honor |
|---------------|---------------------|-------------------------|
| **Boundary** | What's inside the system | Containers cannot expose outside boundary |
| **Actors** | Who interacts with system | Containers implement actor interfaces |
| **Protocols** | How containers communicate | Containers must implement specified protocols |
| **Cross-cutting** | System-wide concerns | All containers follow these patterns |

### Container (c3-N) - MIDDLE

Inherits from Context. Defines WHAT/WHY for its domain. Changes propagate to its Components.

| Inherits From Context | Defines For Components |
|----------------------|------------------------|
| System boundary constraints | Technology stack choices |
| Protocol contracts to implement | Component organization |
| Cross-cutting patterns to follow | Internal patterns & conventions |
| Actor interfaces to support | API contracts |

### Component (c3-NNN) - LEAF

Inherits from both Container and Context (via Container). Implements HOW things work.

| Inherits From Container | Inherits From Context | Implements |
|------------------------|----------------------|------------|
| Technology stack | Boundary constraints | Actual code behavior |
| Internal patterns | Cross-cutting patterns | Configuration |
| Interface contracts | Protocol requirements | Error handling |

## Impact Propagation Rules

### Upstream Discovery (scope may be bigger)

| Finding | Signal | Action |
|---------|--------|--------|
| Impact at Context | Higher-level change | Escalate, revisit hypothesis |
| Impact at Container (from Component) | Parent change | Escalate to container-design |
| Impact on siblings | Horizontal expansion | Expand scope |

### Downstream Discovery (expected)

| Finding | Signal | Action |
|---------|--------|--------|
| Impact on child Containers | Normal propagation | Document, delegate to container-design |
| Impact on child Components | Normal propagation | Document, delegate to component-design |
| No further impact | Contained change | Proceed with documentation |

## Escalation Triggers

Changes should escalate UP when they would break inherited contracts:

| Level | Escalate If... |
|-------|----------------|
| **Component → Container** | Violates technology/patterns/interface contract |
| **Container → Context** | Violates boundary/protocol/cross-cutting contract |
| **Any → Context** | Affects system boundary or adds new actors |

## Verification Questions

Before making changes, verify you're at the right level:

### "Is this a Context change?"
- Does it affect system boundary?
- Does it add/remove actors?
- Does it change how containers communicate?
- Does it affect cross-cutting concerns system-wide?

### "Is this a Container change?"
- Does it change technology stack?
- Does it reorganize components?
- Does it change internal patterns?
- Does it affect multiple components the same way?

### "Is this a Component change?"
- Is it about HOW to implement something?
- Is it isolated to this component?
- Does it keep the interface unchanged?
- Does it follow existing patterns?
