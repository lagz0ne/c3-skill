---
id: c3-0
c3-version: 3
title: expansive.ly
summary: AI-powered spatial knowledge canvas for knowledge management
---

# expansive.ly

## Overview

expansive.ly is a spatial knowledge canvas (Miro/FigJam style) where users position knowledge nodes freely in 2D space. Knowledge units are "Concepts" - abstract ideas that can contain multiple resources. The platform supports progressive disclosure where knowledge can be pre-defined and revealed as suggestions (onboarding), or freely discovered. All interactions are enriched by AI via suggestions and chat.

## Containers

| ID | Name | Purpose |
|----|------|---------|
| c3-1 | Web Frontend | React-based spatial canvas application with TanStack Router |
| c3-2 | API Backend | Bun runtime backend handling business logic and orchestration |
| c3-3 | Graph Database | Neo4j for native concept/relationship storage |
| c3-4 | Real-time Service | Live cursor sync and collaborative editing |
| c3-5 | AI Service | Multi-provider AI integration for suggestions and chat |

## Container Interactions

```mermaid
graph TB
    subgraph Users["External Actors"]
        U[User]
        I[Import Sources]
    end

    subgraph System["expansive.ly"]
        C1[c3-1 Web Frontend]
        C2[c3-2 API Backend]
        C3[c3-3 Graph Database]
        C4[c3-4 Real-time Service]
        C5[c3-5 AI Service]
    end

    subgraph External["External Systems"]
        E1[AI Providers]
        E2[OAuth Providers]
    end

    U -->|Canvas interactions| C1
    U -->|Real-time cursors| C4
    I -->|External data| C2

    C1 -->|API calls| C2
    C1 <-->|WebSocket| C4

    C2 -->|Graph queries| C3
    C2 -->|AI requests| C5
    C2 -->|Auth| E2

    C4 -->|Presence sync| C3

    C5 -->|LLM calls| E1
```

## External Actors

| Actor | Type | Interaction |
|-------|------|-------------|
| User | Human | Interacts with canvas, creates concepts, asks AI questions |
| Import Sources | System | Notion, Roam, Obsidian, browser bookmarks |

## External Systems

| ID | System | Type | Purpose |
|----|--------|------|---------|
| E1 | AI Providers | service | OpenAI, Anthropic, etc. for LLM capabilities |
| E2 | OAuth Providers | auth | Google, GitHub for user authentication |
| E3 | Redis | cache/pubsub | Real-time Pub/Sub and presence state |

## Linkages

| From | To | Protocol | Purpose |
|------|-----|----------|---------|
| c3-1 | c3-2 | HTTPS/REST | Canvas operations, concept CRUD |
| c3-1 | c3-4 | WebSocket | Live cursor positions, presence |
| c3-2 | c3-3 | Bolt | Graph queries and mutations |
| c3-2 | c3-5 | Internal | AI request orchestration |
| c3-4 | c3-3 | Bolt | Presence state persistence |
| c3-4 | E3 | Redis | Pub/Sub for presence broadcast |
| c3-5 | E1 | HTTPS | LLM API calls |
| c3-2 | E2 | HTTPS/OAuth | Authentication flow |

## Key Data Flows

### Concept Creation Flow

```mermaid
sequenceDiagram
    participant U as User
    participant FE as c3-1 Frontend
    participant API as c3-2 Backend
    participant DB as c3-3 Graph DB
    participant AI as c3-5 AI Service
    participant RT as c3-4 Real-time

    U->>FE: Create concept at (x,y)
    FE->>API: POST /concepts
    API->>DB: CREATE (c:Concept)
    DB-->>API: concept_id
    API->>AI: Enrich concept (async)
    AI-->>API: suggestions, tags
    API->>DB: UPDATE concept
    API-->>FE: concept created
    FE->>RT: Broadcast new concept
    RT-->>FE: Sync to collaborators
```

### AI Chat Flow

```mermaid
sequenceDiagram
    participant U as User
    participant FE as c3-1 Frontend
    participant API as c3-2 Backend
    participant DB as c3-3 Graph DB
    participant AI as c3-5 AI Service

    U->>FE: Ask question about knowledge
    FE->>API: POST /chat
    API->>DB: Get relevant concepts (vector search)
    DB-->>API: context nodes
    API->>AI: Query with context
    AI-->>API: response + suggested links
    API-->>FE: chat response
    FE->>U: Display answer with citations
```
