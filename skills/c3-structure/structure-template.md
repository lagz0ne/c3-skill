# Structure Templates

## Context Template (c3-0)

```markdown
---
id: c3-0
c3-version: 3
title: [System Name]
type: context
summary: >
  [One-line system description]
---

# [System Name]

## Overview

[Single paragraph: what this system does, its boundary]

## Containers

| ID | Name | Responsibility |
|----|------|----------------|
| c3-1 | [Name] | [What it does] |
| c3-2 | [Name] | [What it does] |

## Interactions

` ` `mermaid
flowchart LR
    C1[Container 1]
    C2[Container 2]
    Ext[External System]

    C1 -->|protocol| C2
    C2 -->|protocol| Ext
` ` `

## External Actors

| Actor | Interaction |
|-------|-------------|
| [User/System] | [How they interact] |
```

---

## Container Template (c3-N)

```markdown
---
id: c3-{N}
c3-version: 3
title: [Container Name]
type: container
parent: c3-0
summary: >
  [One-line description]
---

# [Container Name]

## Inherited From Context

- **Responsibility:** [from Context inventory]
- **Connects to:** [adjacent containers]

## Overview

[Single paragraph purpose]

## Technology Stack

| Category | Choice |
|----------|--------|
| Runtime | [e.g., Node.js 20] |
| Framework | [e.g., Hono] |

## Components

| ID | Name | Type | Responsibility |
|----|------|------|----------------|
| c3-{N}01 | [Name] | Foundation | [What it does] |
| c3-{N}02 | [Name] | Business | [What it does] |

## Internal Structure

` ` `mermaid
flowchart TD
    subgraph Container["[Name]"]
        C1[Component 1]
        C2[Component 2]
        C1 --> C2
    end
` ` `

## Key Flows

**[Flow Name]:** [Brief description of what happens]
```

---

## Notes

- **Inventory first:** List all containers/components even without detailed docs
- **No code:** Tables only, no snippets
- **Mermaid only:** No ASCII diagrams
