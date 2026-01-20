---
id: ref-graph-patterns
title: Graph Data Patterns
---

# Graph Data Patterns

## Goal

Define query and write patterns for the Neo4j knowledge graph ensuring tenant isolation, performance optimization, and consistent data access across all services.

## Overview

Conventions for working with the Neo4j knowledge graph.

## Query Patterns

### Tenant Isolation

All queries MUST include tenant filtering. The graph client injects this automatically.

| Pattern | When |
|---------|------|
| `WHERE n.tenantId = $tid` | All node queries |
| `MATCH path WHERE ALL(n IN nodes(path) WHERE n.tenantId = $tid)` | Path queries |

### Common Query Shapes

| Query | Pattern |
|-------|---------|
| Single concept | `MATCH (c:Concept {id: $id, tenantId: $tid})` |
| Canvas concepts | `MATCH (c:Concept {canvasId: $cid, tenantId: $tid})` |
| Linked concepts | `MATCH (c:Concept)-[:LINKS_TO]->(related) WHERE c.id = $id` |
| Path between | `MATCH path = shortestPath((a)-[*..5]-(b)) WHERE a.id = $from AND b.id = $to` |

### Pagination

| Approach | Use Case |
|----------|----------|
| SKIP/LIMIT | Small result sets, UI pagination |
| Cursor-based | Large sets, streaming |
| Time-based | Activity feeds |

## Write Patterns

### Atomic Operations

| Operation | Pattern |
|-----------|---------|
| Create with relations | Single transaction, MERGE for idempotency |
| Cascade delete | Application-controlled, not DB constraints |
| Position update | Batch multiple concepts in one transaction |

### Conflict Resolution

| Scenario | Strategy |
|----------|----------|
| Concurrent position updates | Last-write-wins with timestamp |
| Link creation race | Both succeed (allow duplicates, dedupe on read) |
| Concept deletion during link | Delete wins, orphan link cleaned |

## Performance Guidelines

| Guideline | Rationale |
|-----------|-----------|
| Index all ID fields | O(1) lookups |
| Composite index (tenantId, canvasId) | Common filter combo |
| Limit relationship traversal depth | Prevent runaway queries |
| Use EXPLAIN for complex queries | Understand query plan |

## Cited By

- c3-203 (Graph Client)
- c3-211 (Concept Service)
- c3-301 (Graph Schema)
- c3-302 (Vector Index)
- c3-513 (Suggestion Engine)
