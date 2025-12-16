# Container Archetypes

## The Principle

> **This is what matters. Everything else is just examples.**

### The Contract Chain

```
Context  → defines WHAT containers exist and WHY
Container → defines WHAT components exist and WHAT they do
Component → defines HOW it implements its responsibility
```

Each layer implements the layer above. This never changes.

### The Question Every Container Must Answer

> **"What is this container's relationship to content?"**

Once you answer this, components become clear:

| Relationship | Components Document |
|--------------|---------------------|
| Creates/processes content | Processing logic, flows |
| Stores content | Structure, access patterns |
| Transports content | Channels, delivery mechanics |
| Operates on containers | Operational processes |
| Interfaces with external | Our side of the boundary |

### Deriving Components

For ANY container, ask:

1. What does this container DO? (its responsibility from Context)
2. What parts make that happen? (those are your components)
3. For each part, HOW does it work? (that's your component doc)

**You don't need an archetype to document a container.** The principle is sufficient.

---

## Why Archetypes Exist

Archetypes are **shortcuts, not rules**.

They provide:
- Sensible starting points
- Common component patterns others have found useful
- Vocabulary for discussing container types

They are NOT:
- Exhaustive (your container might not fit any)
- Prescriptive (you can deviate freely)
- Required (the principle is what matters)

---

## Common Archetypes

These patterns appear frequently. Use them as starting points, adapt as needed.

### Service

**Relationship:** Creates and processes content (business logic)

**Typical Components:**

| Component | Documents |
|-----------|-----------|
| Router/Handler | Request routing, entry points |
| Service/Logic | Business rules, orchestration |
| Adapter | Integration with other containers |

**Signs you have a Service:** You wrote code that does business logic.

---

### Data

**Relationship:** Stores and structures content

**Typical Components:**

| Component | Documents |
|-----------|-----------|
| Schema | Data model, relationships, constraints |
| Indexes | Query patterns, optimization |
| Migrations | Evolution strategy |

**Signs you have a Data container:** It persists state, you configure its structure.

---

### Boundary

**Relationship:** Interface to systems we don't control

**Typical Components:**

| Component | Documents |
|-----------|-----------|
| Contract | What we expect (API shape, SLAs) |
| Client | Our adapter/SDK usage |
| Fallback | Retry, circuit breaker, degraded mode |
| Webhook/Events | How we receive their events |

**Signs you have a Boundary:** It's external, we integrate with it, we document OUR side.

---

### Platform

**Relationship:** Operates on other containers

**Typical Components:**

| Component | Documents |
|-----------|-----------|
| CI Pipeline | Build, test, publish |
| Deployment | Rollout, health checks, rollback |
| Networking | Service mesh, ingress, DNS |
| Secrets | Storage, rotation, access |
| Observability | Logging, metrics, tracing |

**Signs you have a Platform:** It doesn't do business logic, it runs/operates other containers.

---

## Extended Archetypes

Less universal, but common enough to mention.

### Gateway

**Relationship:** Routes and guards content

**Typical Components:**

| Component | Documents |
|-----------|-----------|
| Routes | Path mapping, versioning |
| Auth | Token validation, API keys |
| Rate Limiting | Throttling rules |
| Transform | Request/response shaping |

**Signs you have a Gateway:** It sits in front, routes traffic, enforces policies.

---

### Messaging

**Relationship:** Transports content between containers

**Typical Components:**

| Component | Documents |
|-----------|-----------|
| Topics/Queues | Channel inventory, purpose |
| Schemas | Message structure, versioning |
| Consumers | Who reads what, ordering |
| Dead Letter | Failure handling |

**Signs you have a Messaging container:** Async communication infrastructure you configure.

**Note:** If using managed messaging (SQS, Pub/Sub), consider treating as Boundary instead.

---

## Custom Archetypes

Your container doesn't fit? That's fine. Apply the principle:

### Example: ML Pipeline

**Step 1: What's its relationship to content?**
> Transforms raw data into trained models

**Step 2: What parts make that happen?**

| Component | Documents |
|-----------|-----------|
| Data Ingestion | Sources, preprocessing |
| Training | Algorithm, hyperparameters |
| Evaluation | Metrics, validation |
| Serving | How models get deployed |

**Step 3: For each, document HOW it works.**

Done. No archetype needed.

---

### Example: Analytics

**Relationship:** Aggregates and reports on content

| Component | Documents |
|-----------|-----------|
| Collection | Event capture, SDKs |
| Pipeline | Transformation, enrichment |
| Storage | Warehouse schema |
| Dashboards | Key reports, access |

---

## Summary

1. **Learn the principle** - it applies everywhere
2. **Use archetypes as shortcuts** - when they fit
3. **Adapt freely** - your system, your components
4. **When in doubt** - ask "what is this container's relationship to content?"

The principle is the foundation. Archetypes are just furniture.
