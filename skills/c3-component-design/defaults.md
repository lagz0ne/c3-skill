# Component Layer Defaults

## Abstraction Level

**Close-up view** - Zoom into one component to see HOW it works.

This layer explains HOW the component implements the contract defined at Container level. It's abstract (no code) but detailed enough to understand before making changes. It answers: "How does this component fulfill its responsibility?"

## Include

| Element | Purpose |
|---------|---------|
| Flows | Step-by-step: how it processes requests/data |
| Dependencies | What other components/services it calls and why |
| Decision logic | Key branching points and rules |
| Edge cases | Non-obvious scenarios and their handling |
| Error scenarios | What can go wrong and how it's handled |
| State/Lifecycle | If stateful, how state transitions work |

## Exclude

| Element | Push To | Why |
|---------|---------|-----|
| WHAT this component does | Container | Already defined there |
| Component relationships | Container | Higher abstraction |
| Container relationships | Context | Even higher abstraction |
| Code | Auxiliary docs | Code changes, adds context load |
| File paths | Codebase | Changes with refactoring |

## Litmus Test

> "Is this about HOW this component implements its contract?"

- **Yes** → Component level
- **No (what it does)** → Push to Container
- **No (relationships)** → Push to Container or Context

**View check:** Component docs IMPLEMENT what Container described. If Container says "UserService handles user operations", Component explains HOW UserService does that.

## Diagrams

| Type | Use For |
|------|---------|
| Flowchart | Processing steps, decision logic |
| Sequence | Calls to dependencies |
| State chart | Lifecycle/state transitions |
| **Avoid** | System context, container overview, component relationships |

## Documentation Pattern

Since Component implements Container's contract:

1. **Reference the contract** - "As defined in Container, this component handles X"
2. **Explain the flow** - Step-by-step how it accomplishes X
3. **List dependencies** - What it calls to accomplish X
4. **Cover edge cases** - What happens when things go wrong
