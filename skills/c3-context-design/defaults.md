# Context Layer Defaults

## Abstraction Level

**Highest abstraction** - System-level contracts and boundaries.

This layer defines the INTERFACES that containers must honor. It answers: "What exists and how do they relate?"

## Include

| Element | Purpose |
|---------|---------|
| System boundary | What's inside vs outside the system |
| Actors | Who/what interacts with the system |
| Container inventory | What containers exist and their responsibilities |
| Inter-container protocols | How containers communicate (REST, gRPC, events) |
| Cross-cutting contracts | System-wide patterns (auth, logging, error handling) |
| Deployment topology | Where things run conceptually |

## Exclude

| Element | Push To | Why |
|---------|---------|-----|
| Technology choices | Container | Implementation detail |
| How components work | Container/Component | Lower abstraction |
| Configuration values | Component | Implementation detail |
| Code examples | Auxiliary docs | Code changes, adds context load |

## Litmus Test

> "Does understanding this require seeing the full system picture?"

- **Yes** → Context level (boundaries, protocols, cross-cutting)
- **No** → Push to Container

**Abstraction check:** If it's about HOW a specific container works internally, it's too detailed for Context.

## Diagrams

| Type | Use For |
|------|---------|
| **Primary: System Context** | Bird's-eye view of system boundary and actors |
| **Secondary: Container Overview** | High-level container relationships |
| **Avoid** | Sequence diagrams with methods, class diagrams, flowcharts with logic |
