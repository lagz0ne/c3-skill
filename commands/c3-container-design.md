---
description: Design Container level - technology choices, component organization, middleware
---

Design or update a Container in your C3 architecture.

## What Container Level Covers

- **Container identity**: Name, purpose, responsibilities
- **Technology stack**: Language, framework, runtime
- **Component organization**: Internal structure overview
- **Middleware pipeline**: Auth, rate limiting, request processing
- **APIs**: Endpoints exposed and consumed
- **Data responsibilities**: What data this container owns
- **Deployment**: Container-specific deployment characteristics

## Abstraction Level

Focus on WHAT and WHY - not HOW. Describe characteristics and architectural decisions, not implementation code.

Document:
- Technology choices and rationale
- Component organization
- API surface (endpoints, protocols)
- Middleware/proxy layer behavior
- Data ownership boundaries

Do NOT include:
- Detailed code implementations
- Line-by-line configurations
- Algorithm implementations

## Output

Creates or updates: `.c3/containers/CON-{NNN}-{slug}.md`

With sections:
- Overview
- Technology Stack
- Middleware Pipeline
- Component Organization
- API Endpoints
- Communication Patterns
- Data Responsibilities
- Configuration (high-level)
- Deployment

## Related Commands

- `/c3-context-design` - System-wide architecture
- `/c3-component-design` - Detail specific components within this container
