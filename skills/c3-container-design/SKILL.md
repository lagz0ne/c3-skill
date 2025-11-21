---
name: c3-container-design
description: Explore Container level impact during scoping - technology choices, component organization, middleware, and inter-container communication
---

# C3 Container Level Exploration

## Overview

Explore Container-level impact during the scoping phase of c3-design. Container is the middle layer: individual services, their technology, and component organization.

**Abstraction Level:** WHAT and WHY, not HOW. Characteristics and architecture, not implementation code.

## When Invoked

Called during EXPLORE phase of c3-design when:
- Hypothesis suggests Container-level impact
- Need to understand service-level implications
- Exploring downstream from Context
- Exploring upstream from Component
- Change affects technology stack or middleware

## Container Level Defines

| Concern | Examples |
|---------|----------|
| **Container identity** | Name, purpose, responsibilities |
| **Technology stack** | Language, framework, runtime |
| **Component organization** | Internal structure, layers |
| **Middleware pipeline** | Auth, rate limiting, request flow |
| **APIs** | Endpoints exposed and consumed |
| **Data responsibilities** | What data this container owns |
| **Deployment specifics** | Container-level deployment |

## Exploration Questions

When exploring Container level, investigate:

### Isolated (at Container)
- What container responsibilities change?
- What middleware pipeline affected?
- What APIs need modification?

### Upstream (to Context)
- Does this change system boundaries?
- Do protocols need updating?
- Are cross-cutting concerns affected?

### Adjacent (same level)
- What sibling containers related?
- What inter-container communication affected?
- What shared dependencies exist?

### Downstream (to Components)
- Which components inside this container affected?
- What new components needed?
- How does component organization change?

## Reading Container Documents

Use c3-locate to retrieve:

```
c3-locate CON-001                    # Overview
c3-locate #con-001-technology-stack  # Tech choices
c3-locate #con-001-middleware        # Request pipeline
c3-locate #con-001-components        # Internal structure
c3-locate #con-001-api-endpoints     # API surface
c3-locate #con-001-communication     # Inter-container
c3-locate #con-001-data              # Data ownership
c3-locate #con-001-configuration     # Config approach
c3-locate #con-001-deployment        # Deployment details
```

## Impact Signals

| Signal | Meaning |
|--------|---------|
| New middleware layer needed | Cross-component change |
| API contract change | Consumers affected |
| Technology stack change | Major container rewrite |
| Data ownership change | Migration needed |
| New container needed | Context-level impact |

## Output for c3-design

After exploring Container level, report:
- What Container-level elements are affected
- Impact on adjacent containers
- Components that need deeper exploration
- Whether Context level needs revisiting
- Whether hypothesis needs revision

## Document Template Reference

Container documents follow this structure:

```markdown
## Overview {#con-nnn-overview}
## Technology Stack {#con-nnn-technology-stack}
## Middleware Pipeline {#con-nnn-middleware}
## Component Organization {#con-nnn-components}
## API Endpoints {#con-nnn-api-endpoints}
## Communication Patterns {#con-nnn-communication}
## Data Responsibilities {#con-nnn-data}
## Configuration {#con-nnn-configuration}
## Deployment {#con-nnn-deployment}
## Related {#con-nnn-related}
```

Use these heading IDs for precise exploration.
