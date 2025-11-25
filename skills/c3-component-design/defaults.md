# Component Layer Defaults

## Include

| Element | Example |
|---------|---------|
| Stack details | `pg: 8.11.x` - why chosen |
| Environment config | `DB_POOL_MAX=50` (dev vs prod) |
| Implementation patterns | Connection pooling algorithm |
| Interfaces/Types | Method signatures, DTOs |
| Error handling | Retry strategies, error catalog |
| Usage examples | TypeScript snippets |

## Exclude

| Element | Push To |
|---------|---------|
| Container purpose | Container |
| API endpoint list | Container |
| Technology choice rationale | Container |
| System protocols | Context |

## Litmus Test

> "Could a developer implement this from the documentation?"

- **Yes** → Correct level
- **No, needs more detail** → Add specifics
- **No, it's about structure** → Push to Container

## Diagrams

| Type | Use For |
|------|---------|
| Flowchart | Decision logic |
| Sequence | Method calls |
| State chart | Lifecycle/state |
| ERD | Data structures |
| Class diagram | Type relationships |
| **Avoid** | System context, container overview, deployment diagrams |
