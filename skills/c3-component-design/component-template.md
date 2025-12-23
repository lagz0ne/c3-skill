# Component Template

```markdown
---
id: c3-{N}{NN}
c3-version: 3
title: [Component Name]
type: component
parent: c3-{N}
summary: >
  [One-line description of what this component does]
---

# [Component Name]

## Contract
From Container (c3-{N}): "[responsibility from parent container]"

## Interface
[REQUIRED: Diagram showing boundary and hand-off points]

` ` `mermaid
flowchart LR
    subgraph IN["Receives From"]
        Input1[Component/External]
    end

    subgraph SELF["Owns"]
        Process[What this component does]
    end

    subgraph OUT["Provides To"]
        Output1[Component/Caller]
    end

    IN --> SELF --> OUT
` ` `

## Hand-offs
| Direction | What | To/From |
|-----------|------|---------|
| Receives | [data/request] | [source component] |
| Provides | [data/response] | [target component] |

## Conventions
| Rule | Applies To | Why |
|------|------------|-----|
| [Convention name] | [What it governs] | [Rationale] |

## Edge Cases & Errors
| Scenario | Behavior |
|----------|----------|

## [Optional Sections - Include Only When Relevant]

### Organization
[Include if internal structure is complex - show layers, subsystems]

### Configuration
[Include if component has significant config surface - table of settings]

### Dependencies
[Include if external dependencies matter - what it requires from outside]

### Invariants
[Include if there are key guarantees to verify - testable truths]

### Performance
[Include if performance characteristics matter - throughput, latency bounds]
```

## Required Sections Checklist

1. ☐ Frontmatter (id, c3-version, title, type, parent, summary)
2. ☐ Contract (from parent Container)
3. ☐ Interface (diagram showing boundary - REQUIRED)
4. ☐ Hand-offs table (what exchanges with whom)
5. ☐ Conventions table (rules for consistency)
6. ☐ Edge Cases & Errors table

## Optional Sections (include based on component nature)

- Additional diagrams (Sequence, State, Organization) - where needed
- Configuration - significant config surface
- Dependencies - external dependencies matter
- Invariants - key guarantees to verify
- Performance - throughput/latency matters
