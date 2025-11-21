---
name: c3-context-design
description: Explore Context level impact during scoping - system boundaries, actors, cross-component interactions, and high-level concerns
---

# C3 Context Level Exploration

## Overview

Explore Context-level impact during the scoping phase of c3-design. Context is the bird's-eye view: system boundaries, actors, and cross-component interactions.

**Abstraction Level:** WHAT exists and HOW they relate. No implementation details.

## When Invoked

Called during EXPLORE phase of c3-design when:
- Hypothesis suggests Context-level impact
- Need to understand system-wide implications
- Exploring upstream from Container/Component
- Change affects system boundaries or protocols

## Context Level Defines

| Concern | Examples |
|---------|----------|
| **System boundaries** | What's inside vs outside the system |
| **Actors** | Users, external systems, third parties |
| **Containers** | High-level view of all containers |
| **Cross-cutting concerns** | Auth strategy, logging, monitoring |
| **Protocols** | REST, gRPC, WebSocket, message queues |
| **Deployment model** | Cloud, on-prem, hybrid (high level) |

## Exploration Questions

When exploring Context level, investigate:

### Isolated (at Context)
- What system boundaries change?
- What actors are affected?
- What protocols need modification?

### Upstream (external)
- What external systems depend on this?
- What third-party integrations affected?
- What user-facing contracts change?

### Adjacent (same level)
- What other cross-cutting concerns related?
- What other protocol decisions affected?

### Downstream (to Containers)
- Which containers does this affect?
- How do container responsibilities change?
- What new containers might be needed?

## Reading Context Documents

Use c3-locate to retrieve:

```
c3-locate CTX-001                    # Overview
c3-locate #ctx-001-architecture      # System diagram
c3-locate #ctx-001-containers        # Container list
c3-locate #ctx-001-protocols         # Communication patterns
c3-locate #ctx-001-cross-cutting     # System-wide concerns
c3-locate #ctx-001-deployment        # Deployment overview
```

## Impact Signals

| Signal | Meaning |
|--------|---------|
| Change affects system boundary | Major architectural shift |
| New actor type introduced | Interface design needed |
| Protocol change | All containers using it affected |
| Cross-cutting concern change | Ripples through all layers |

## Output for c3-design

After exploring Context level, report:
- What Context-level elements are affected
- Impact magnitude (boundary change = high, protocol tweak = medium)
- Downstream containers that need exploration
- Whether hypothesis needs revision

## Document Template Reference

Context documents follow this structure:

```markdown
## Overview {#ctx-nnn-overview}
## Architecture {#ctx-nnn-architecture}
## Containers {#ctx-nnn-containers}
## Protocols & Communication {#ctx-nnn-protocols}
## Cross-Cutting Concerns {#ctx-nnn-cross-cutting}
## Deployment {#ctx-nnn-deployment}
## Related {#ctx-nnn-related}
```

Use these heading IDs for precise exploration.
