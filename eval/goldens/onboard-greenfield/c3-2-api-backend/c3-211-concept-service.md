---
id: c3-211
c3-version: 3
title: Concept Service
type: component
parent: c3-2
category: feature
summary: CRUD operations and business logic for knowledge concepts
---

# Concept Service

## Goal

Handle concept lifecycle operations including creation, positioning, enrichment triggers, and deletion with relationship cleanup.

## Contract

From c3-2 (API Backend): "CRUD operations for concepts"

## Interface Diagram

```mermaid
flowchart LR
    subgraph IN["Receives"]
        I1[Concept data]
        I2[Position updates]
        I3[Canvas context]
    end

    subgraph PROCESS["Owns"]
        P1[Validation]
        P2[Graph mutations]
        P3[Enrichment trigger]
        P4[Event emission]
    end

    subgraph OUT["Provides"]
        O1[Concept entity]
        O2[Related concepts]
        O3[Events]
    end

    IN --> PROCESS --> OUT
```

## Hand-offs

| Direction | What | To/From |
|-----------|------|---------|
| IN | Create/update request | c3-201 Router |
| IN | User and tenant context | c3-202 Auth |
| OUT | Graph mutations | c3-203 Graph Client |
| OUT | Enrichment job | c3-205 Job Processor |
| OUT | Real-time events | c3-4 Real-time Service |

## Operations

| Operation | Validation | Side Effects |
|-----------|------------|--------------|
| Create | Title required, position valid | Trigger enrichment, emit event |
| Update | Concept exists, user owns | Emit update event |
| Move | Valid coordinates | Batch position update |
| Delete | Concept exists | Remove links, emit event |
| Bulk move | All IDs valid | Single transaction |

## Concept Model

| Field | Type | Description |
|-------|------|-------------|
| id | UUID | Unique identifier |
| tenantId | UUID | Tenant isolation |
| canvasId | UUID | Parent canvas |
| title | string | Display title |
| content | string | Markdown content |
| position | {x, y} | Canvas coordinates |
| tags | string[] | AI-generated + manual |
| embedding | vector | Semantic search |
| createdAt | timestamp | Creation time |
| updatedAt | timestamp | Last modification |

## Enrichment Flow

```mermaid
sequenceDiagram
    participant S as Concept Service
    participant Q as Job Queue
    participant AI as AI Service
    participant DB as Graph DB

    S->>Q: Enqueue enrichment job
    Q->>AI: Process concept
    AI->>AI: Generate tags, summary
    AI->>AI: Create embedding
    AI-->>Q: Enrichment result
    Q->>DB: Update concept
    Q->>S: Job complete event
```

## Conventions

| Rule | Why |
|------|-----|
| Soft delete with TTL | Recovery window |
| Position batched in transactions | Consistency |
| Enrichment async | Non-blocking create |
| Title max 200 chars | Display constraints |

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Duplicate title in canvas | Allow, warn user |
| Delete with many links | Cascade delete links |
| Position conflict | Last write wins |
| Enrichment fails | Mark concept, retry 3x |

## References

- Concept service: `src/services/concept.ts`
- Concept routes: `src/api/routes/concepts.ts`
- Cites: ref-graph-patterns, ref-async-processing
