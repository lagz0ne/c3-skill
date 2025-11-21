---
description: Design Component level - implementation details, configuration, technical specifics
---

Design or update a Component in your C3 architecture.

## What Component Level Covers

- **Component identity**: Name, purpose within container
- **Technical implementation**: Libraries, patterns, algorithms
- **Configuration**: Environment variables, initialization
- **Dependencies**: External libraries, other components
- **Interfaces**: Methods, data structures, APIs
- **Error handling**: Failure modes, retry logic, fallbacks
- **Performance**: Caching, optimization, resource usage
- **Health checks**: Monitoring, observability

## Abstraction Level

Implementation details are appropriate here. Code examples, configuration snippets, and specific library usage are expected.

Document:
- Library choices and usage
- Configuration with environment variables
- Code patterns and examples
- Error handling strategies
- Performance characteristics

## Output

Creates or updates: `.c3/components/{container}/COM-{NNN}-{slug}.md`

With sections:
- Overview
- Purpose
- Technical Implementation
- Configuration (detailed)
- Behavior (pool management, lifecycle, etc.)
- Error Handling
- Performance
- Health Checks
- Usage Example

## Related Commands

- `/c3-context-design` - System-wide architecture
- `/c3-container-design` - Parent container specification
