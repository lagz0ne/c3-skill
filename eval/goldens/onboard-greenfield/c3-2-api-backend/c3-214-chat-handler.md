---
id: c3-214
c3-version: 3
title: Chat Handler
type: component
parent: c3-2
category: feature
summary: AI chat request orchestration with RAG context assembly
---

# Chat Handler

## Goal

Orchestrate AI chat requests by assembling relevant context from the knowledge graph, routing to AI providers, and streaming responses with citations.

## Contract

From c3-2 (API Backend): "AI chat request orchestration"

## Interface Diagram

```mermaid
flowchart LR
    subgraph IN["Receives"]
        I1[User message]
        I2[Canvas context IDs]
        I3[Chat history]
    end

    subgraph PROCESS["Owns"]
        P1[Context assembly]
        P2[RAG retrieval]
        P3[Prompt construction]
        P4[Response streaming]
    end

    subgraph OUT["Provides"]
        O1[Streamed response]
        O2[Citations]
        O3[Suggestions]
    end

    IN --> PROCESS --> OUT
```

## Hand-offs

| Direction | What | To/From |
|-----------|------|---------|
| IN | Chat request | c3-201 Router |
| IN | Visible concept IDs | c3-1 Frontend |
| IN | Previous messages | Request body |
| OUT | Concept lookups | c3-203 Graph Client |
| OUT | AI completion request | c3-204 AI Orchestrator |
| OUT | SSE stream | Client |

## RAG Pipeline

```mermaid
flowchart TB
    subgraph Retrieval["Context Retrieval"]
        R1[Get visible concepts]
        R2[Vector similarity search]
        R3[Graph neighborhood]
        R4[Merge and rank]
    end

    subgraph Assembly["Prompt Assembly"]
        A1[System prompt]
        A2[Context block]
        A3[Chat history]
        A4[User query]
    end

    subgraph Generation["Response"]
        G1[AI completion]
        G2[Citation extraction]
        G3[Suggestion generation]
    end

    R1 --> R4
    R2 --> R4
    R3 --> R4
    R4 --> A2
    A1 --> G1
    A2 --> G1
    A3 --> G1
    A4 --> G1
    G1 --> G2
    G2 --> G3
```

## Context Assembly

| Source | Weight | Purpose |
|--------|--------|---------|
| Visible concepts | High | Immediate context |
| Vector similar | Medium | Semantic relevance |
| Graph neighbors | Medium | Relationship context |
| Recent edits | Low | Activity relevance |

## Conventions

| Rule | Why |
|------|-----|
| Max 10 context concepts | Token budget |
| Stream with 100ms flush | Responsive UX |
| Citation format [c:id] | Parseable by frontend |
| History limited to 20 turns | Context window |

## Response Format

| Field | Type | Description |
|-------|------|-------------|
| content | string | AI response text (streamed) |
| citations | array | [{conceptId, title, relevance}] |
| suggestions | array | [{action, params, label}] |
| usage | object | Token counts |

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| No relevant concepts | General response, suggest exploration |
| Very long context | Truncate, prioritize by relevance |
| AI provider timeout | Return partial, offer retry |
| Rate limit hit | 429 with reset time |

## References

- Chat handler: `src/api/routes/chat.ts`
- RAG pipeline: `src/services/rag.ts`
- Cites: ref-ai-integration, ref-streaming-patterns
