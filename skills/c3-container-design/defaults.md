# Container Layer Defaults

## Include

| Element | Example |
|---------|---------|
| Technology stack | Node.js 20, Express 4.18 |
| Container responsibilities | "Handles API requests" |
| Component relationships | Flowchart of connections |
| Data flow | Sequence diagram |
| Component inventory | Links to component docs |
| API surface | Endpoints exposed |
| Data ownership | "Owns User accounts, Tasks" |
| Inter-container communication | "REST to Backend, SQL to DB" |

## Exclude

| Element | Push To |
|---------|---------|
| System boundary | Context |
| Cross-cutting concerns | Context |
| Implementation code | Component |
| Library specifics | Component |
| Configuration values | Component |

## Litmus Test

> "Is this about WHAT this container does and WITH WHAT, not HOW internally?"

- **Yes** → Container level
- **No (system-wide)** → Push to Context
- **No (implementation)** → Push to Component

## Diagrams

| Type | Use For |
|------|---------|
| **Required: Component Relationships** | Flowchart showing how components interact |
| **Required: Data Flow** | Sequence diagram showing request paths |
| **Avoid** | System context, actor diagrams, detailed class diagrams |
