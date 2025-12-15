# Component Layer Defaults

## Abstraction Level

**Implementation abstraction** - How things work internally.

This layer IMPLEMENTS Container interfaces. It answers: "How does this component work and why?"

**Key principle:** Document understanding, not code. Code lives in the codebase and changes frequently. C3 documents HOW things work conceptually so readers can understand before making changes.

## Include

| Element | Purpose |
|---------|---------|
| Behavior description | What this component does and why |
| Internal flow | How it processes requests/data (conceptual) |
| Decision logic | Key branching and rules |
| Error handling strategy | How errors are handled and why |
| Edge cases | Non-obvious scenarios and their handling |
| Dependencies | What it depends on and why |
| Configuration concepts | What can be configured and why |

## Exclude

| Element | Push To | Why |
|---------|---------|-----|
| Container organization | Container | Higher abstraction |
| Technology choice rationale | Container | Architectural decision |
| System protocols | Context | System-wide concern |
| **Code snippets** | Auxiliary docs or codebase | Code changes, adds context load |
| **Exact file paths** | Codebase | Changes with refactoring |
| **Type definitions** | Codebase | Implementation detail |

## Litmus Test

> "Does understanding this require knowing how it works internally?"

- **Yes** → Component level (behavior, flow, edge cases)
- **No (organization)** → Push to Container
- **No (system-wide)** → Push to Context

**Abstraction check:** If someone needs this to UNDERSTAND the component before modifying it, it belongs here. If they need it to WRITE code, it belongs in auxiliary docs or the codebase.

## Diagrams

| Type | Use For |
|------|---------|
| Flowchart | Decision logic |
| Sequence | Method calls |
| State chart | Lifecycle/state |
| ERD | Data structures |
| Class diagram | Type relationships |
| **Avoid** | System context, container overview, deployment diagrams |

## Nature

Determines documentation focus (no code in any nature).

| Nature | Indicators | Documentation Focus |
|--------|-----------|---------------------|
| **Resource** | DB, cache, API client, queue | Connection behavior, retry strategy, failure modes |
| **Business** | Services, domain models, workflows | Domain flows, business rules, edge cases |
| **Framework** | Controllers, handlers, middleware | Request flow, validation logic, error responses |
| **Cross-cutting** | Utilities, helpers, shared modules | Contract stability, usage patterns, evolution strategy |

**Decision questions:**
1. Does it connect to external resources? -> Resource
2. Does it encode business rules? -> Business
3. Does it handle requests/responses? -> Framework
4. Is it used by multiple other components? -> Cross-cutting
