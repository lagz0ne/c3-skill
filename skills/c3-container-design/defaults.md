# Container Layer Defaults

## Abstraction Level

**Architectural abstraction** - Component organization and internal contracts.

This layer IMPLEMENTS Context interfaces and DEFINES interfaces for Components. It answers: "How is this container organized and why?"

## Include

| Element | Purpose |
|---------|---------|
| Technology choices | What stack and why (Node.js for X, PostgreSQL for Y) |
| Component inventory | What components exist and their responsibilities |
| Component relationships | How components interact (flowchart) |
| Data flow patterns | How requests/data move through (sequence) |
| API surface | What this container exposes to others |
| Data ownership | What data this container is responsible for |
| Internal contracts | Patterns components must follow |

## Exclude

| Element | Push To | Why |
|---------|---------|-----|
| System boundary | Context | Higher abstraction |
| Cross-cutting concerns | Context | System-wide, not container-specific |
| How components work internally | Component | Lower abstraction |
| Configuration values | Component | Implementation detail |
| Code examples | Auxiliary docs | Code changes, adds context load |

## Litmus Test

> "Does understanding this require knowing how components are organized?"

- **Yes** → Container level (architecture, tech choices, component contracts)
- **No (system-wide)** → Push to Context
- **No (inner workings)** → Push to Component

**Abstraction check:** If it's about HOW a specific component works internally, it's too detailed for Container.

## Diagrams

| Type | Use For |
|------|---------|
| **Required: Component Relationships** | Flowchart showing how components interact |
| **Required: Data Flow** | Sequence diagram showing request paths |
| **Avoid** | System context, actor diagrams, detailed class diagrams |
