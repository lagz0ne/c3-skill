# Container Component Patterns

## Relationship to Content

> **"What is this container's relationship to content?"**

| Relationship | Container... | Component Focus |
|--------------|--------------|-----------------|
| **Processes** | Creates, transforms, orchestrates | Logic, flows, rules |
| **Stores** | Persists, structures, indexes | Schema, access patterns |
| **Transports** | Moves between places | Channels, routing, delivery |
| **Presents** | Displays to users | Views, interactions, state |
| **Integrates** | Bridges to external | Contracts, adapters, fallbacks |
| **Operates** | Manages other containers | Pipelines, config, observability |

Most containers combine 2-3 relationships.

---

## Component Inventory

Pick components based on what your container does.

### Entry Points

| Component | When Needed |
|-----------|-------------|
| **Routes/Handler** | HTTP/API entry |
| **UI/Views** | User-facing |
| **CLI** | Command-line entry |
| **Consumer** | Async/event entry |
| **Scheduler** | Time-triggered |

### Logic

| Component | When Needed |
|-----------|-------------|
| **Service/Domain** | Business rules |
| **Transform** | Data shaping |
| **Workflow** | Multi-step processes |
| **Calculation** | Computations |

### State

| Component | When Needed |
|-----------|-------------|
| **Schema** | Structured data |
| **Cache** | Temporary storage |
| **Session** | User state |
| **Config** | Runtime settings |

### Communication

| Component | When Needed |
|-----------|-------------|
| **Client/Adapter** | Calls other services |
| **Publisher** | Sends events/messages |
| **Webhook** | Receives external events |
| **Contract** | External API boundary |

### Resilience

| Component | When Needed |
|-----------|-------------|
| **Fallback** | Degraded operation |
| **Validation** | Input protection |
| **Error Handling** | Failure management |

### Operations

| Component | When Needed |
|-----------|-------------|
| **Deployment** | Release process |
| **Observability** | Runtime insight |
| **Pipeline** | Build/test/publish |

---

## Usage

1. Identify container's relationships (processes? stores? presents?)
2. Scan inventory for matching components
3. Add to Container's component inventory
4. Create component docs when conventions mature
