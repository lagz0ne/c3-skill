# Container Template

```markdown
---
id: c3-{N}
c3-version: 3
title: [Container Name]
type: container
parent: c3-0
summary: >
  [One-line description of container purpose]
---

# [Container Name]

## Inherited From Context
- **Boundary:** [what this container can/cannot access]
- **Protocols:** [what protocols this container uses]
- **Cross-cutting:** [patterns inherited from Context]

## Overview
[Single paragraph purpose]

## Technology Stack
| Technology | Version | Purpose |
|------------|---------|---------|

## Architecture

### External Relationships
[REQUIRED: Map Context-defined interfaces to components]

**Context defines the external interfaces. Container maps which component owns each interface.**

` ` `mermaid
flowchart LR
    subgraph Container["This Container (c3-N)"]
        C1[Component c3-N01]
        C2[Component c3-N02]
    end

    ExtA[External System A]
    ExtB[Container c3-M]

    C1 -->|"protocol"| ExtA
    C2 -->|"protocol"| ExtB
` ` `

| External Interface (from Context) | Owning Component | Protocol |
|-----------------------------------|------------------|----------|
| [External system/container] | [Component ID] | [How they communicate] |

### Internal Structure
[REQUIRED: Mermaid diagram showing component relationships with layering]

**Container defines how components relate** (just as Context defines how containers relate).

` ` `mermaid
flowchart TD
    subgraph Container["[Container Name] (c3-{N})"]
        subgraph Business["Business Layer"]
            B1[Handler c3-N03]
            B2[Service c3-N04]
        end

        subgraph Foundation["Foundation Layer"]
            F1[Framework c3-N01]
            F2[Data Access c3-N02]
        end

        B1 --> F1
        B2 --> F1
        B2 --> F2
    end
` ` `

### Component Layering Rules

| Layer | Contains | Rules |
|-------|----------|-------|
| **Foundation** | Framework integrations, data access, infrastructure | No business logic. Defines patterns others use. |
| **Business** | Domain handlers, services, middleware with business rules | Can depend on Foundation. Avoid cross-dependencies. |

**Dependency direction:** Business → Foundation only.

## Components
| Component | ID | Responsibility |
|-----------|-----|----------------|

## Key Flows
[1-2 critical flows - describe WHAT happens, not HOW]
```

## Required Sections Checklist

1. ☐ Frontmatter (id, c3-version, title, type, parent, summary)
2. ☐ Inherited From Context
3. ☐ Overview
4. ☐ Technology Stack
5. ☐ Architecture - External Relationships (diagram + table REQUIRED)
6. ☐ Architecture - Internal Structure (diagram REQUIRED)
7. ☐ Component Layering Rules
8. ☐ Components table
9. ☐ Key Flows
