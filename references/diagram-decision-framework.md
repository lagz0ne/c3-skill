# Diagram Decision Framework

Strategic guidance for choosing container-level diagrams.

## Philosophy

There is NO one-size-fits-all diagram. Different containers need different diagrams based on complexity and role.

**THE GOAL:** Complement prose with visual clarity.
- If prose is clear → skip the diagram
- If relationships are non-obvious → diagram helps
- If there's a critical flow → show it

## Quick Reference

| Container Type | Recommended Diagrams |
|---------------|---------------------|
| Simple (1-3 components, linear) | Component table only, maybe one flowchart |
| Moderate (4-6 components, branching) | Component flowchart, maybe one sequence |
| Complex (7+ components, mesh) | Component flowchart + ONE critical sequence + maybe state |

## Diagram Capabilities

### Flowchart / Box-and-Arrow

| Can Show | Cannot Show |
|----------|-------------|
| Static structure (components exist) | Temporal ordering |
| Dependencies (A depends on B) | Conditional paths clearly |
| Data flow direction | Request/response pairing |
| Groupings (subgraphs) | Async vs sync distinction |
| Entry/exit points | Error paths |

**Best for:** "What exists and what talks to what"

### Sequence Diagram

| Can Show | Cannot Show |
|----------|-------------|
| Temporal ordering | Overall structure |
| Request/response pairing | Dependencies beyond this flow |
| Sync vs async | Parallel paths (gets messy) |
| Activation periods | Multiple entry points |
| Alt/opt blocks | State persistence |

**Best for:** "What happens when X occurs"

### State Diagram

| Can Show | Cannot Show |
|----------|-------------|
| States an entity can be in | WHO performs transition |
| Transitions between states | Multiple entities |
| Triggers for transitions | Data payload |
| Terminal states | Component structure |

**Best for:** "What states can this entity be in"

### Table / Matrix

| Can Show | Cannot Show |
|----------|-------------|
| Mapping between dimensions | Relationships or flow |
| Responsibility assignment | Hierarchy or nesting |
| Feature coverage | Process or sequence |

**Best for:** "What does what" or "Who owns what"

## Combination Patterns

### Pattern 1: Overview + Critical Flow
**When:** Complex structure AND non-obvious key flow
**Use:** Flowchart (structure) + Sequence (one critical path)

### Pattern 2: Overview + State Machine
**When:** Manages stateful entities with complex lifecycles
**Use:** Flowchart (components) + State diagram (entity lifecycle)

### Pattern 3: External + Internal
**When:** Boundary service with both external and internal complexity
**Use:** External diagram (boundaries) + Internal flowchart (components)

## Anti-Patterns

1. **Two flowcharts at different zoom levels** - Can't mentally map between them
2. **Sequence for every endpoint** - Noise, maintenance nightmare
3. **State diagram for simple states** - Overkill if < 4 states
4. **Table + diagram showing same thing** - Redundant, will drift

## Decision Checklist

For each potential diagram, evaluate:

| Factor | Question |
|--------|----------|
| **Clarity** | Can prose alone convey this? Is this non-obvious? |
| **Value** | Will readers return to this? Is it a "north star"? |
| **Cost** | How often will it change? How hard to update? |

**Decision:** INCLUDE / SKIP / SIMPLIFY

## Placement Rules

| Diagram Type | Section | Anchor |
|--------------|---------|--------|
| Component Overview | ## Component Organization | `{#c3-n-organization}` |
| Critical Flow | ## Key Flows | `{#c3-n-flows}` |
| State Diagram | ## Entity Lifecycle | `{#c3-n-lifecycle}` |
| External Interactions | ## External Dependencies | `{#c3-n-external}` |

- Diagram comes AFTER introductory prose
- Diagram comes BEFORE detailed tables/lists
- Caption/legend if notation is non-obvious
- Reference in prose: "As shown in the diagram below..."
