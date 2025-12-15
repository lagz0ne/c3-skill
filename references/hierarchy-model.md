# C3 Hierarchy Model

The C3 hierarchy (Context-Container-Component) is an **abstraction hierarchy**. Each layer defines interfaces that lower layers implement. The purpose is **understanding how things work**, not documenting code.

## Core Principle: Abstraction Levels

```
Higher abstraction = DEFINES interfaces/contracts
Lower abstraction  = IMPLEMENTS those interfaces (explains HOW)

Code lives in the codebase, not in C3 documents.
C3 documents enable UNDERSTANDING before making changes.
```

## The Three Layers

```
┌─────────────────────────────────────────────────────────────────┐
│  CONTEXT (c3-0) - HIGHEST ABSTRACTION                           │
│  Defines system-wide INTERFACES that containers must honor      │
│  • System boundary (what's in/out)                              │
│  • Actors (who interacts)                                       │
│  • Protocols (how containers communicate)                       │
│  • Cross-cutting contracts (auth, logging, errors)              │
│                              │                                  │
│                              ▼ IMPLEMENTS                       │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  CONTAINERS (c3-1, c3-2, c3-3, ...) - MIDDLE ABSTRACTION │   │
│  │  Implements Context interfaces, defines Component interfaces │
│  │  • Technology choices (runtime, framework)               │   │
│  │  • Component organization                                │   │
│  │  • Internal contracts (patterns components must follow)  │   │
│  │  • API surface (what this container exposes)             │   │
│  │                              │                           │   │
│  │                              ▼ IMPLEMENTS                │   │
│  │  ┌────────────────────────────────────────────────────┐  │   │
│  │  │  COMPONENTS (c3-101, c3-102, ...) - LOWEST ABSTRACT│  │   │
│  │  │  Implements Container interfaces                   │  │   │
│  │  │  • HOW it works (behavior, not code)               │  │   │
│  │  │  • Decision logic and edge cases                   │  │   │
│  │  │  • Error handling strategy                         │  │   │
│  │  │  • Failure modes and recovery                      │  │   │
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

### Context (c3-0) - HIGHEST ABSTRACTION

Defines system-wide INTERFACES. Documents understanding at the system level.

| Interface Type | What Context Defines | What Containers Must Implement |
|----------------|---------------------|-------------------------------|
| **Boundary** | What's inside the system | How to honor that boundary |
| **Actors** | Who interacts with system | How to serve those actors |
| **Protocols** | How containers communicate | How to implement those protocols |
| **Cross-cutting** | System-wide patterns | How to apply those patterns |

### Container (c3-N) - MIDDLE ABSTRACTION

Implements Context interfaces. Defines INTERFACES for its Components. Documents understanding at the architectural level.

| Implements From Context | Defines For Components |
|------------------------|------------------------|
| Boundary constraints | Technology choices and why |
| Protocol requirements | Component organization and why |
| Cross-cutting patterns | Internal contracts and patterns |
| Actor interface requirements | API surface and expectations |

### Component (c3-NNN) - LOWEST ABSTRACTION

Implements Container interfaces. Documents understanding at the behavioral level (how it works, not code).

| Implements From Container | Documents (No Code) |
|--------------------------|---------------------|
| Technology contracts | Behavioral flows |
| Internal patterns | Decision logic |
| Interface expectations | Edge cases and why |
| API surface requirements | Failure modes and recovery |

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
