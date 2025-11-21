---
name: c3-container-design
description: Explore Container level impact during scoping - technology choices, component organization, middleware, and inter-container communication
---

# C3 Container Level Exploration

## Overview

Explore Container-level impact during the scoping phase of c3-design. Container is the middle layer: individual services, their technology, and component organization.

**Abstraction Level:** WHAT and WHY, not HOW. Characteristics and architecture, not implementation code.

**Announce at start:** "I'm using the c3-container-design skill to explore Container-level impact."

## When Invoked

Called during EXPLORE phase of c3-design when:
- Hypothesis suggests Container-level impact
- Need to understand service-level implications
- Exploring downstream from Context
- Exploring upstream from Component
- Change affects technology stack or middleware

Also called by c3-adopt to CREATE initial Container documentation.

---

## Core Principles for Container Documents
- **Reading order:** consume CTX first; CON derives from CTX and constrains COM.
- **Reference direction:** Downward-only. Containers link to components; CTX links to containers. No upward links needed.
- **Two container types:** Code vs Infrastructure. Infra is a leaf (no components).
- **Anchors:** Use `{#con-xxx-*}` anchors for all sections you will link to (protocols, cross-cutting, relationships, components, config).
- **Single source:** Do not redefine protocols/cross-cutting; map CTX items to implementations in this container.

---

## Container Types

| Type | Has Components? | Focus |
|------|-----------------|-------|
| **Code** | Yes | Tech stack, protocols implemented, component relationships, data flow, cross-cutting mapping |
| **Infrastructure** | No (leaf) | Engine/config/features offered to code containers; consumers must be linked |

Infra containers never define components. Their features are consumed by code containers/components.

---

## Required Outputs

### Code Container (type=Code)
- Technology stack.
- Protocol implementations table mapping each CTX protocol to specific component sections.
- Component relationships flowchart (must exist).
- Data flow sequence diagram (must exist).
- Container cross-cutting mapped to components.
- Component inventory with nature and responsibility.

### Infrastructure Container (type=Infra)
- Engine/technology and deployment mode.
- Configuration table with rationale.
- Features provided table with links to consuming code containers/components.
- No component sections (leaf).

---

## Templates (copy/paste and fill)

### Code Container
````markdown
# CON-XXX <Name> (Code)

## Technology Stack {#con-xxx-stack}
- Runtime, language, framework

## Protocol Implementations {#con-xxx-protocols}
| Protocol (from CTX) | Implemented In |
|---------------------|----------------|
| REST/HTTPS auth | [COM-002-auth] |
| ... | ... |

## Component Relationships {#con-xxx-relationships}
```mermaid
flowchart LR
    ... component flow ...
```

## Data Flow {#con-xxx-data-flow}
```mermaid
sequenceDiagram
    ... request/response path ...
```

## Container Cross-Cutting {#con-xxx-cross-cutting}
- Logging: implemented by [COM-0xx-logger]
- Error handling: implemented by [COM-0xx-errors]
- Validation/observability/etc.: links to components

## Components {#con-xxx-components}
| Component | Nature | Responsibility |
|-----------|--------|----------------|
| [COM-001-name](../components/COM-001-name.md) | Entrypoint | ... |
| ... | ... | ... |
````

### Infrastructure Container
```markdown
# CON-XXX <Name> (Infrastructure)

## Engine {#con-xxx-engine}
- Version/edition, deployment mode

## Configuration {#con-xxx-config}
| Setting | Value | Why |
|---------|-------|-----|
| ... | ... | ... |

## Features Provided {#con-xxx-features}
| Feature | Consumed By |
|---------|-------------|
| WAL logical replication | [CON-001#components] → [COM-005-event-streaming] |
| ... | ... |
```

---

## Checklists

### Code Container Checklist
- Stack recorded.
- CTX protocols mapped to specific components/sections.
- Flowchart shows component relationships (exists).
- Sequence diagram shows data flow (exists).
- Cross-cutting choices mapped to components.
- Component inventory complete with nature/responsibility.
- Anchors follow `{#con-xxx-*}` scheme for link targets.

### Infrastructure Container Checklist
- Engine/version stated.
- Config table with rationale.
- Features table with consumer links.
- Explicitly no components included (leaf).
- Anchors follow `{#con-xxx-*}` scheme for link targets.

---

## Diagram Guidance (Code container)
- **Flowchart (required):** show component relationships/connection paths.
- **Sequence diagram (required):** show request/data flow through the container.
- Use mermaid snippets above; keep labels aligned to component names in Components table.

---

## Derivation & Linking
- CTX protocols/cross-cutting → map here with links to component sections.
- Component inventory must cover everything shown in relationships/data-flow diagrams.
- Infra features listed here must be cited by consuming components in code containers.
- Downward-only references; no duplication of protocol definitions from CTX.

---

## Exploration Questions (use to fill templates)
- Identity: Single responsibility? If it vanished, what breaks?
- Technology: Language/framework/runtime? Why chosen?
- Protocols: Which CTX protocols does this implement? Which components handle them?
- Relationships: How do components connect? What is the data path?
- Cross-cutting: Which components implement logging, error handling, validation?
- Data: What data does this own/read? Where is it stored?
- Infra (if type=Infra): Which features do code containers consume? What config matters?
