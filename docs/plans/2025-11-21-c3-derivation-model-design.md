# C3 Derivation Model Design

> Concrete, ready-to-apply spec for how Context, Container, and Component docs derive from each other, including templates, checklists, and required skill updates.

## Intent
- Make derivation explicit: higher layers **constrain** lower layers; lower layers **implement** higher-layer contracts.
- Provide repeatable templates and checklists so authors can produce consistent CTX/CON/COM documents without re-deciding structure.
- Ensure traceability: every protocol, cross-cutting decision, and component has a clear downward link to its implementation.

---

## Core Rules
- **Reading order:** Context → Container → Component.
- **Reference direction:** Only downward links. Context → Container sections; Container → Component docs. No upward links.
- **Single source:** Define relationships once at the highest layer that owns the decision; lower layers only implement/link.
- **Infrastructure containers are leaf nodes:** they have no components; their features are consumed by code containers/components.

---

## Templates and Checklists

### Context Document (CTX-###)
**Purpose:** Bird’s-eye view of the system: what exists, boundaries, and protocols that must be implemented below.

**Template (copy/paste and fill):**
```markdown
# CTX-XXX <Name>

## System Boundary {#ctx-xxx-boundary}
- Inside: ...
- Outside: ...

## Actors {#ctx-xxx-actors}
| Actor | Role |
|-------|------|
| ...   | ...  |

## Containers {#ctx-xxx-containers}
| Container | Type (Code/Infra) | Description |
|-----------|-------------------|-------------|
| [C3-X-name](./containers/C3-X-name.md) | Code | ... |
| ... | ... | ... |

## Protocols {#ctx-xxx-protocols}
| From | To | Protocol | Implementations |
|------|----|----------|-----------------|
| Frontend | Backend | REST/HTTPS | [C3-2#api], [C3-1#rest] |
| ... | ... | ... | ... |

## Cross-Cutting {#ctx-xxx-cross-cutting}
- Auth: ... implemented in [C3-X#auth]
- Logging: ... implemented in [C3-X#logging]
- Error strategy: ... implemented in [C3-X#errors]

## Deployment Topology {#ctx-xxx-deployment}
- Diagram or bullets for high-level infra layout
```

**Checklist (must be true to call CTX done):**
- System boundary and actors listed.
- Container inventory table includes every container with type (Code/Infra) and link.
- Protocols table lists every inter-container communication with links to implementing container sections.
- Cross-cutting decisions listed with downward links to container sections.
- Topology described (diagram or text).

### Code Container Document (C3-C, type=Code)
**Purpose:** What this container does and with what components; how it fulfills Context protocols and cross-cutting choices.

**Template:**
```markdown
# C3-X <Name> (Code)

## Technology Stack {#c3-x-stack}
- Runtime, language, framework

## Protocol Implementations {#c3-x-protocols}
| Protocol (from CTX) | Implemented In |
|---------------------|----------------|
| REST/HTTPS auth | [C3-X02-auth] |
| ... | ... |

## Component Relationships {#c3-x-relationships}
```mermaid
flowchart LR
    ... component flow ...
```

## Data Flow {#c3-x-data-flow}
```mermaid
sequenceDiagram
    ... request/response path ...
```

## Container Cross-Cutting {#c3-x-cross-cutting}
- Logging: implemented by [C3-Xxx-logger]
- Error handling: implemented by [C3-Xxx-errors]
- Validation/observability/etc.: links to components

## Components {#c3-x-components}
| Component | Nature | Responsibility |
|-----------|--------|----------------|
| [C3-X01-name](../components/C3-X01-name.md) | Entrypoint | ... |
| ... | ... | ... |
```

**Checklist:**
- Stack recorded.
- Protocols table maps every CTX protocol to specific components/sections.
- Flowchart shows component relationships (must exist).
- Sequence diagram shows data flow (must exist).
- Cross-cutting choices mapped to components.
- Component inventory complete with nature + responsibility.

### Infrastructure Container Document (C3-C, type=Infra)
**Purpose:** Leaf node describing platform service features that code containers consume.

**Template:**
```markdown
# C3-X <Name> (Infrastructure)

## Engine {#c3-x-engine}
- Version/edition, deployment mode

## Configuration {#c3-x-config}
| Setting | Value | Why |
|---------|-------|-----|
| ... | ... | ... |

## Features Provided {#c3-x-features}
| Feature | Consumed By |
|---------|-------------|
| WAL logical replication | [C3-1#components] → [C3-105-event-streaming] |
| ... | ... |
```

**Checklist:**
- Engine/version stated.
- Config table with rationale.
- Features table lists capabilities with links to consuming code containers/components.
- No component-level sections; this is a leaf.

### Component Document (C3-CNN)
**Purpose:** HOW the component works; implementation detail level.

**Template:**
```markdown
# C3-XNN <Name> (<Nature>)

## Overview {#c3-xnn-overview}
- Responsibility and how it fits container protocols/cross-cutting.

## Stack {#c3-xnn-stack}
- Library/version choices; why selected.

## Configuration {#c3-xnn-config}
| Env Var | Dev | Prod | Why |
|---------|-----|------|-----|
| ... | ... | ... |

## Interfaces & Types {#c3-xnn-interfaces}
- Signatures, DTOs, events; link to schemas.

## Behavior {#c3-xnn-behavior}
- Narrative plus diagram where helpful.
```mermaid
stateDiagram-v2
    ... if useful ...
```
- Alternative: flowchart/sequence/ERD depending on need.

## Error Handling {#c3-xnn-errors}
| Error | Retriable | Action/Code |
|-------|-----------|-------------|
| ... | ... | ... |

## Usage {#c3-xnn-usage}
```typescript
// exemplar usage
```

## Dependencies {#c3-xnn-deps}
- Upstream/downstream components; infra features consumed.
```

**Checklist:**
- Nature chosen and informs focus (resource, business, cross-cutting, entrypoint, testing, deployment, contextual, etc.).
- Stack and configuration fully documented (with env differences).
- Interfaces/types specified.
- Behavior explained with at least one diagram if non-trivial.
- Error handling table present.
- Usage example shows intended invocation.
- Dependencies and consumed infra features listed.

---

## Derivation Enforcement
- **Context → Container:** Every protocol and cross-cutting item in CTX links to specific container sections describing implementation. Containers must not invent protocols absent from CTX without updating CTX.
- **Container → Component:** Every protocol or cross-cutting implementation in a code container maps to specific components/sections. Component inventory must cover all behavior shown in relationships/data-flow diagrams.
- **Infrastructure:** Features listed in infra containers must be cited by consuming components; infra docs are final leaves.
- **Anchors:** Use `{#ctx-xxx-...}`, `{#c3-x-...}`, `{#c3-xnn-...}` so downward links are stable.
- **No upward duplication:** lower layers do not redefine relationships; they implement and link back downwards only.

---

## Skill Updates (must implement)
1. `skills/c3-context-design/SKILL.md`
   - Add checklist enforcing container inventory, protocol table with downward links, cross-cutting with links, topology.
   - Include CTX template above; require anchor usage guidance and downward-only linking.
2. `skills/c3-container-design/SKILL.md`
   - Describe Code vs Infrastructure container types and the “infra is leaf” rule.
   - For Code containers, require protocol mapping table, flowchart, sequence diagram, cross-cutting mapping, and component inventory with nature/responsibility.
   - For Infra containers, require engine/config/features table with consuming links, and explicitly forbid components.
3. `skills/c3-component-design/SKILL.md`
   - Include nature taxonomy guidance (open-ended, examples above).
   - Require stack/config ownership, interfaces, behavior diagrams when non-trivial, error handling table, usage example, dependencies.
4. Examples
   - Update sample CTX/CON/COM docs to match the templates (or add new examples mirroring them).
5. `skills/c3-adopt/SKILL.md`
   - Ensure adoption flow instructs teams to create CTX → CON → COM following these templates and checklists, including anchor conventions and downward-link validation.

---

## Quick Reference: Derivation Chain
```
Context (WHAT exists, HOW they relate)
│  ├─ Protocols → CON#section
│  └─ Cross-cutting → CON#section
↓
Container (WHAT it does, WITH WHAT)
│  ├─ Code containers → components, relationships (flowchart), data flow (sequence)
│  ├─ Cross-cutting → COM links
│  └─ Infra containers → features consumed by Code
↓
Component (HOW it works)
   ├─ Nature-driven focus
   ├─ Stack, config, interfaces, behavior, errors, usage
   └─ Terminal leaf
```
