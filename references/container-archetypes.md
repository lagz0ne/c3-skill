# Container Component Patterns

## The Principle

> **Every container has a relationship to content. That relationship determines its components.**

### The Contract Chain

```
Context   → defines WHAT containers exist and WHY
Container → defines WHAT components exist and WHAT they do
Component → defines HOW it implements its responsibility
```

### The One Question

> **"What is this container's relationship to content?"**

| Relationship | Container... | Component Focus |
|--------------|--------------|-----------------|
| **Processes** | Creates, transforms, orchestrates | Logic, flows, rules |
| **Stores** | Persists, structures, indexes | Schema, access patterns |
| **Transports** | Moves between places | Channels, routing, delivery |
| **Presents** | Displays to users | Views, interactions, state |
| **Integrates** | Bridges to external | Contracts, adapters, fallbacks |
| **Operates** | Manages other containers | Pipelines, config, observability |

Most containers combine 2-3 relationships. A backend service *processes* AND *integrates*. A frontend *presents* AND *stores* (client state).

---

## Component Inventory

Instead of rigid archetypes, pick components based on what your container actually does.

### Entry Points

How content enters the container.

| Component | When Needed | Documents |
|-----------|-------------|-----------|
| **Routes/Handler** | HTTP/API entry | Path mapping, methods, middleware |
| **UI/Views** | User-facing | Pages, layouts, navigation |
| **CLI** | Command-line entry | Commands, args, output |
| **Consumer** | Async/event entry | Subscriptions, message handling |
| **Scheduler** | Time-triggered | Cron patterns, job definitions |

### Logic

How content is processed.

| Component | When Needed | Documents |
|-----------|-------------|-----------|
| **Service/Domain** | Business rules | Rules, validation, orchestration |
| **Transform** | Data shaping | Mapping, enrichment, normalization |
| **Workflow** | Multi-step processes | Steps, conditions, compensation |
| **Calculation** | Computations | Algorithms, formulas, models |

### State

How content is stored or managed.

| Component | When Needed | Documents |
|-----------|-------------|-----------|
| **Schema** | Structured data | Models, relationships, constraints |
| **Cache** | Temporary storage | Keys, TTL, invalidation |
| **Session** | User state | Storage, expiry, security |
| **Config** | Runtime settings | Sources, precedence, secrets |

### Communication

How content moves to/from other containers.

| Component | When Needed | Documents |
|-----------|-------------|-----------|
| **Client/Adapter** | Calls other services | Endpoints, retry, timeout |
| **Publisher** | Sends events/messages | Topics, schemas, delivery |
| **Webhook** | Receives external events | Endpoints, validation, processing |
| **Contract** | External API boundary | Expected shape, SLAs, versioning |

### Resilience

How the container handles failure.

| Component | When Needed | Documents |
|-----------|-------------|-----------|
| **Fallback** | Degraded operation | Circuit breaker, defaults, retry |
| **Validation** | Input protection | Rules, sanitization, errors |
| **Error Handling** | Failure management | Categories, recovery, reporting |

### Operations

How the container is run and observed.

| Component | When Needed | Documents |
|-----------|-------------|-----------|
| **Deployment** | Release process | Rollout, health checks, rollback |
| **Observability** | Runtime insight | Logging, metrics, tracing |
| **Pipeline** | Build/test/publish | Stages, gates, artifacts |

---

## Deriving Components

For ANY container:

1. **What does it do?** (from Context - its responsibility)
2. **What relationships does it have?** (processes? stores? presents?)
3. **Scan the inventory** - which components apply?
4. **For each component** - document HOW it works

### Example: Backend API

**Responsibilities:** Handle user requests, apply business logic, persist data

**Relationships:** Processes + Stores + Integrates

**Components from inventory:**
- Entry: Routes/Handler
- Logic: Service/Domain
- State: Schema, Config
- Communication: Client (to database), Publisher (events)
- Resilience: Validation, Error Handling

### Example: Frontend App

**Responsibilities:** Present UI, capture interactions, manage client state

**Relationships:** Presents + Stores (client) + Integrates (API)

**Components from inventory:**
- Entry: UI/Views
- Logic: Transform (data formatting)
- State: Session, Cache, Config
- Communication: Client (to backend)
- Resilience: Fallback (offline), Error Handling

### Example: Worker/Job Processor

**Responsibilities:** Process async tasks from queue

**Relationships:** Processes + Integrates

**Components from inventory:**
- Entry: Consumer, Scheduler
- Logic: Workflow, Service/Domain
- State: Config
- Communication: Client, Publisher (results)
- Resilience: Fallback, Error Handling

### Example: Infrastructure/Platform

**Responsibilities:** Deploy and operate other containers

**Relationships:** Operates

**Components from inventory:**
- Operations: Pipeline, Deployment, Observability
- State: Config (secrets, env)
- Resilience: Fallback (rollback)

---

## Summary

1. **Ask the question** - "What's this container's relationship to content?"
2. **Identify relationships** - Processes? Stores? Presents? Integrates? Operates?
3. **Pick from inventory** - Select components that match
4. **Document each** - HOW does this component work?

No archetypes needed. The inventory covers all cases.
