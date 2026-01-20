# expansive.ly Architecture

> AI-powered spatial knowledge canvas for knowledge management

## Context

- [README](./README.md) - System overview and container interactions

## Containers

### c3-1 Web Frontend
React-based spatial canvas application with TanStack Router

- [README](./c3-1-web-frontend/README.md) - Container overview
- [c3-101 Canvas Engine](./c3-1-web-frontend/c3-101-canvas-engine.md) - Spatial rendering
- [c3-102 Concept Node](./c3-1-web-frontend/c3-102-concept-node.md) - Visual concept representation
- [c3-104 State Atoms](./c3-1-web-frontend/c3-104-state-atoms.md) - State management
- [c3-111 Canvas Screen](./c3-1-web-frontend/c3-111-canvas-screen.md) - Main workspace
- [c3-113 AI Chat Panel](./c3-1-web-frontend/c3-113-ai-chat-panel.md) - AI chat interface
- [c3-115 Onboarding Flow](./c3-1-web-frontend/c3-115-onboarding-flow.md) - Guided discovery

### c3-2 API Backend
Bun runtime backend handling business logic and orchestration

- [README](./c3-2-api-backend/README.md) - Container overview
- [c3-201 HTTP Router](./c3-2-api-backend/c3-201-http-router.md) - Request routing
- [c3-202 Auth Middleware](./c3-2-api-backend/c3-202-auth-middleware.md) - Authentication
- [c3-203 Graph Client](./c3-2-api-backend/c3-203-graph-client.md) - Neo4j abstraction
- [c3-211 Concept Service](./c3-2-api-backend/c3-211-concept-service.md) - Concept CRUD
- [c3-214 Chat Handler](./c3-2-api-backend/c3-214-chat-handler.md) - AI chat orchestration

### c3-3 Graph Database
Neo4j for native concept/relationship storage

- [README](./c3-3-graph-database/README.md) - Container overview
- [c3-301 Graph Schema](./c3-3-graph-database/c3-301-graph-schema.md) - Data model
- [c3-302 Vector Index](./c3-3-graph-database/c3-302-vector-index.md) - Semantic search

### c3-4 Real-time Service
Live cursor sync and collaborative editing

- [README](./c3-4-realtime-service/README.md) - Container overview
- [c3-401 Connection Manager](./c3-4-realtime-service/c3-401-connection-manager.md) - WebSocket lifecycle
- [c3-402 Presence State](./c3-4-realtime-service/c3-402-presence-state.md) - Cursor tracking

### c3-5 AI Service
Multi-provider AI integration for suggestions and chat

- [README](./c3-5-ai-service/README.md) - Container overview
- [c3-501 Provider Abstraction](./c3-5-ai-service/c3-501-provider-abstraction.md) - Unified AI interface
- [c3-502 Prompt Templates](./c3-5-ai-service/c3-502-prompt-templates.md) - Prompt management
- [c3-513 Suggestion Engine](./c3-5-ai-service/c3-513-suggestion-engine.md) - AI suggestions

## References

- [ref-ai-integration](./refs/ref-ai-integration.md) - AI provider conventions
- [ref-api-conventions](./refs/ref-api-conventions.md) - REST API patterns
- [ref-canvas-interactions](./refs/ref-canvas-interactions.md) - Spatial interaction patterns
- [ref-design-system](./refs/ref-design-system.md) - Visual design conventions
- [ref-error-handling](./refs/ref-error-handling.md) - Error handling patterns
- [ref-graph-patterns](./refs/ref-graph-patterns.md) - Neo4j query conventions
- [ref-keyboard-shortcuts](./refs/ref-keyboard-shortcuts.md) - Keyboard shortcuts
- [ref-onboarding-patterns](./refs/ref-onboarding-patterns.md) - Progressive disclosure
- [ref-realtime-patterns](./refs/ref-realtime-patterns.md) - WebSocket conventions
- [ref-state-management](./refs/ref-state-management.md) - Frontend state patterns
- [ref-streaming-patterns](./refs/ref-streaming-patterns.md) - SSE response handling

## ADRs

- [adr-00000000-c3-adoption](./adr/adr-00000000-c3-adoption.md) - C3 methodology adoption
