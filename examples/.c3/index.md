---
layout: home
title: C3 Architecture Documentation Example
---

# C3 Architecture Documentation Example

This is an example C3 documentation structure for a simple web application called **TaskFlow** - a task management system with backend, frontend, and database components.

## Navigation

- [CTX-001: System Overview](./CTX-001-system-overview.md)
- [Containers](./containers/)
  - [CON-001: Backend](./containers/CON-001-backend.md)
- [Components](./components/)
  - [COM-001: DB Connection Pool](./components/backend/COM-001-db-pool.md)
- [ADRs](./adr/)
  - [ADR-001: REST API Choice](./adr/ADR-001-rest-api.md)

## About

This example demonstrates:
- Unique ID conventions (CTX-XXX, CON-XXX, COM-XXX, ADR-XXX)
- Simplified frontmatter with summaries
- Heading-level IDs and optional summaries
- Cross-references between levels
- Mermaid diagrams at appropriate abstraction

## Getting Started

To explore this example:

1. Start with **CTX-001** to understand the system landscape
2. Dive into **CON-001** to see backend container details
3. Check **COM-001** for implementation-level component specs
4. Read **ADR-001** to understand architectural decisions

## Conventions Used

| Level | ID Pattern | Location | Purpose |
|-------|-----------|----------|---------|
| Context | `CTX-NNN-slug` | `.c3/` | Bird's-eye system view |
| Container | `CON-NNN-slug` | `.c3/containers/` | Container characteristics |
| Component | `COM-NNN-slug` | `.c3/components/{container}/` | Implementation details |
| ADR | `ADR-NNN-slug` | `.c3/adr/` | Architecture decisions |
