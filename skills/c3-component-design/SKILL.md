---
name: c3-component-design
description: Explore Component level impact during scoping - implementation details, configuration, dependencies, and technical specifics
---

# C3 Component Level Exploration

## Overview

Explore Component-level impact during the scoping phase of c3-design. Component is the implementation layer: detailed specifications, configuration, and technical behavior.

**Abstraction Level:** Implementation details. Code examples, configuration snippets, and library usage are appropriate here.

## When Invoked

Called during EXPLORE phase of c3-design when:
- Hypothesis suggests Component-level impact
- Need to understand implementation implications
- Exploring downstream from Container
- Change affects specific technical behavior

## Component Level Defines

| Concern | Examples |
|---------|----------|
| **Component identity** | Name, purpose within container |
| **Technical implementation** | Libraries, patterns, algorithms |
| **Configuration** | Environment variables, config files |
| **Dependencies** | External libraries, other components |
| **Interfaces** | Methods, data structures, APIs |
| **Error handling** | Failures, retry logic, fallbacks |
| **Performance** | Caching, optimization, resources |
| **Health checks** | Monitoring, observability |

## Exploration Questions

When exploring Component level, investigate:

### Isolated (at Component)
- What implementation details change?
- What configuration affected?
- What error handling needs updating?

### Upstream (to Container)
- Does this change container responsibilities?
- Does middleware need modification?
- Are container APIs affected?

### Adjacent (same level)
- What sibling components related?
- What shared utilities affected?
- What component interactions change?

### Downstream (consumers)
- What code uses this component?
- What tests need updating?
- What documentation affected?

## Reading Component Documents

Use c3-locate to retrieve:

```
c3-locate COM-001                    # Overview
c3-locate #com-001-implementation    # Technical details
c3-locate #com-001-configuration     # Config options
c3-locate #com-001-pool-behavior     # Specific behavior
c3-locate #com-001-error-handling    # Error strategies
c3-locate #com-001-performance       # Performance characteristics
c3-locate #com-001-health-checks     # Health monitoring
c3-locate #com-001-usage             # Usage examples
```

## Impact Signals

| Signal | Meaning |
|--------|---------|
| Interface change | Consumers need updating |
| Configuration change | Deployment affected |
| Dependency change | Security/compatibility review |
| Error handling change | Monitoring may need updates |
| Component should be higher level | Revisit hypothesis at Container |

## Output for c3-design

After exploring Component level, report:
- What Component-level elements are affected
- Impact on sibling components
- Whether Container level needs changes
- Whether this is truly Component-level or should be higher
- Whether hypothesis needs revision

## Abstraction Check

**Critical question:** Does this change belong at Component level?

Signs it should be **higher** (Container/Context):
- Affects multiple components similarly
- Changes middleware behavior
- Alters container responsibilities
- Impacts system protocols

Signs it's correctly at **Component**:
- Isolated to single component
- Implementation detail only
- No upstream contract changes
- Configuration/behavior tweak

If change belongs higher, report this to c3-design for hypothesis revision.

## Document Template Reference

Component documents follow this structure:

```markdown
## Overview {#com-nnn-overview}
## Purpose {#com-nnn-purpose}
## Technical Implementation {#com-nnn-implementation}
## Configuration {#com-nnn-configuration}
## [Behavior Section] {#com-nnn-behavior}
## Error Handling {#com-nnn-error-handling}
## Performance {#com-nnn-performance}
## Health Checks {#com-nnn-health-checks}
## Usage Example {#com-nnn-usage}
## Related {#com-nnn-related}
```

Use these heading IDs for precise exploration.
