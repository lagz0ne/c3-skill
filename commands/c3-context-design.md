---
description: Design Context level - system boundaries, actors, and cross-component interactions
---

Design or update the Context level of your C3 architecture.

## What Context Level Covers

- **System landscape**: Bird's-eye view of the entire system
- **Actors**: Users, external systems, third-party services
- **Containers**: High-level view of all containers in the system
- **Cross-cutting concerns**: Protocols, authentication strategy, deployment model
- **Communication patterns**: How containers and external systems interact

## Abstraction Level

Focus on WHAT exists and HOW they relate - not implementation details.

Document:
- System boundaries
- User types and their interactions
- External system dependencies
- Container-to-container relationships
- High-level protocols (REST, gRPC, WebSocket, etc.)

Do NOT include:
- Code examples
- Configuration details
- Implementation specifics

## Output

Creates or updates: `.c3/CTX-{NNN}-{slug}.md`

With sections:
- Overview
- Architecture diagram (mermaid)
- Containers list
- Protocols & Communication
- Cross-Cutting Concerns
- Deployment overview

## Related Commands

- `/c3-container-design` - Elaborate individual containers
- `/c3-component-design` - Detail specific components
