# Container Layer Defaults

## Abstraction Level

**Inside view** - Zoom into one container to see its structure.

This layer explains WHAT components exist, their RESPONSIBILITIES, and their RELATIONSHIPS. It also briefly connects to adjacent containers (referencing Context). It answers: "What's inside this container and how do the parts work together?"

## Include

| Element | Purpose |
|---------|---------|
| Component responsibilities | WHAT each component does (not HOW) |
| Component relationships | How components depend on and call each other |
| Data flows | How data moves across components |
| Business flows | Key workflows spanning multiple components |
| Inner patterns | Shared approaches (logging, config, error handling) |
| Adjacent connections | Brief: how this container connects to others (from Context) |

## Exclude

| Element | Push To | Why |
|---------|---------|-----|
| WHY this container exists | Context | Higher abstraction |
| Container-to-container details | Context | Bird's-eye view concern |
| HOW components work | Component | Lower abstraction |
| Implementation details | Component | Abstract implementation |
| Code | Auxiliary docs | Code changes, adds context load |

## Litmus Test

> "Is this about WHAT components do and HOW they relate to each other?"

- **Yes** → Container level
- **No (container relationships)** → Push to Context
- **No (how it works)** → Push to Component

**View check:** If you need to explain the internal logic of one component, it's too detailed for Container.

## Diagrams

| Type | Use For |
|------|---------|
| **Required: Component Relationships** | Flowchart showing how components interact |
| **Required: Data Flow** | Sequence diagram showing request paths |
| **Avoid** | System context, actor diagrams, detailed class diagrams |
