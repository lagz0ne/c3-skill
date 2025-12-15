# C3 Hierarchy Model

The C3 hierarchy (Context-Container-Component) is a **zoom hierarchy**. Each layer is a different zoom level into the system. The purpose is **understanding how things work**, not documenting code.

## Core Principle: Zoom Levels

```
Context   = Bird's-eye view (very far)  → WHY containers exist, relationships
Container = Inside view (mid zoom)      → WHAT components do, how they relate
Component = Close-up view (zoomed in)   → HOW each component works

Code lives in the codebase, not in C3 documents.
C3 documents enable UNDERSTANDING before making changes.
```

## The Three Layers

```
┌─────────────────────────────────────────────────────────────────┐
│  CONTEXT (c3-0) - BIRD'S-EYE VIEW                               │
│  See the system from very far                                   │
│  • WHY each container exists                                    │
│  • Relationships between containers                             │
│  • Connecting points (interfaces between containers)            │
│  • External actors                                              │
│                              │                                  │
│                              ▼ ZOOM IN                          │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  CONTAINERS (c3-1, c3-2, ...) - INSIDE VIEW              │   │
│  │  See the structure inside one container                  │   │
│  │  • WHAT each component does                              │   │
│  │  • Relationships between components                      │   │
│  │  • Data flows across components                          │   │
│  │  • Business flows (workflows)                            │   │
│  │  • Inner patterns (logging, config, errors)              │   │
│  │                              │                           │   │
│  │                              ▼ ZOOM IN                   │   │
│  │  ┌────────────────────────────────────────────────────┐  │   │
│  │  │  COMPONENTS (c3-101, c3-102, ...) - CLOSE-UP VIEW  │  │   │
│  │  │  See HOW one component implements its contract     │  │   │
│  │  │  • Flows (step-by-step processing)                 │  │   │
│  │  │  • Dependencies (what it calls)                    │  │   │
│  │  │  • Decision points                                 │  │   │
│  │  │  • Edge cases and error handling                   │  │   │
│  │  └────────────────────────────────────────────────────┘  │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## The Contract Chain

Each layer defines a contract that the next layer implements:

| Layer | Defines | Implemented By |
|-------|---------|----------------|
| **Context** | WHY containers exist, their relationships | Container |
| **Container** | WHAT components do, their relationships | Component |
| **Component** | HOW it works (the implementation) | (Code in codebase) |

## Layer Responsibilities

### Context (c3-0) - BIRD'S-EYE VIEW

See the system from very far. Document WHY and relationships.

| Documents | Example |
|-----------|---------|
| WHY containers exist | "Backend handles business logic" |
| Container relationships | "Backend calls DB for persistence" |
| Connecting points | "REST API between Frontend and Backend" |
| External actors | "Users, Admin, External Payment API" |

**Does NOT document:** What's inside containers

### Container (c3-N) - INSIDE VIEW

Zoom into one container. Document WHAT components do.

| Documents | Example |
|-----------|---------|
| Component responsibilities | "UserService handles user operations" |
| Component relationships | "UserService calls DBAdapter" |
| Data flows | "Request → Handler → Service → DB" |
| Business flows | "Registration flow across components" |
| Inner patterns | "How logging/config/errors work here" |

**Does NOT document:** How each component works internally

### Component (c3-NNN) - CLOSE-UP VIEW

Zoom into one component. Document HOW it implements its contract.

| Documents | Example |
|-----------|---------|
| Flows | Step-by-step: validate → check exists → create → notify |
| Dependencies | "Calls DBAdapter, EmailService" |
| Decision points | "Skip email in test environment" |
| Edge cases | "Duplicate email returns specific error" |
| Error handling | "DB timeout → retry 3x with backoff" |

**Does NOT document:** Code (that's in the codebase)

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

Before documenting, verify you're at the right zoom level:

### "Is this Context level?" (Bird's-eye)
- Is it about WHY a container exists?
- Is it about how containers relate to each other?
- Is it about external actors or connecting points?

→ If yes, document in Context

### "Is this Container level?" (Inside view)
- Is it about WHAT a component does?
- Is it about how components relate within this container?
- Is it about data flows or business flows?

→ If yes, document in Container

### "Is this Component level?" (Close-up)
- Is it about HOW a component works internally?
- Is it about flows, dependencies, or edge cases?
- Does it implement the contract from Container?

→ If yes, document in Component
