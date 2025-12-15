# Context Layer Defaults

## Abstraction Level

**Bird's-eye view** - See the system from very far.

This layer explains WHY containers exist, their RELATIONSHIPS, and their CONNECTING POINTS. It answers: "What's in this system and why do the parts exist?"

## Include

| Element | Purpose |
|---------|---------|
| Container responsibilities | WHY each container exists, what problem it solves |
| Container relationships | How containers depend on each other |
| Connecting points | Interfaces between containers (APIs, events, data flows) |
| External actors | Who/what interacts with the system from outside |
| System boundary | What's inside vs outside |

## Exclude

| Element | Push To | Why |
|---------|---------|-----|
| What components exist | Container | Container's internal structure |
| How containers work internally | Container | Lower abstraction |
| Implementation details | Component | Even lower abstraction |
| Code | Auxiliary docs | Code changes, adds context load |

## Litmus Test

> "Is this about WHY containers exist and HOW they relate to each other?"

- **Yes** → Context level
- **No (internal structure)** → Push to Container

**View check:** If you need to zoom into a container to explain it, it's too detailed for Context.

## Diagrams

| Type | Use For |
|------|---------|
| **Primary: System Context** | Bird's-eye view of system boundary and actors |
| **Secondary: Container Overview** | High-level container relationships |
| **Avoid** | Sequence diagrams with methods, class diagrams, flowcharts with logic |
