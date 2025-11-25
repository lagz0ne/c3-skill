# Context Layer Defaults

## Include

| Element | Example |
|---------|---------|
| System boundary | "TaskFlow system includes..." |
| Actors | Users, Admin, External APIs |
| Container inventory | Links to container docs |
| Protocols | REST, gRPC, WebSocket |
| Cross-cutting concerns | Auth strategy, logging approach |
| Deployment topology | Cloud, multi-region |

## Exclude

| Element | Push To |
|---------|---------|
| Technology choices | Container |
| Middleware specifics | Container |
| API endpoints | Container |
| Configuration values | Component |
| Code examples | Component |

## Litmus Test

> "Would changing this require coordinating multiple containers or external parties?"

- **Yes** → Context level
- **No** → Push to Container

## Diagrams

| Type | Use For |
|------|---------|
| **Primary: System Context** | Bird's-eye view of system boundary and actors |
| **Secondary: Container Overview** | High-level container relationships |
| **Avoid** | Sequence diagrams with methods, class diagrams, flowcharts with logic |
