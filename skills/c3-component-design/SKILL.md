---
name: c3-component-design
description: Explore Component level impact during scoping - implementation details, configuration, dependencies, and technical specifics
---

# C3 Component Level Exploration

## Overview

Explore Component-level impact during the scoping phase of c3-design. Component is the implementation layer: detailed specifications, configuration, and technical behavior.

**Abstraction Level:** Implementation details. Code examples, configuration snippets, and library usage are appropriate here.

**Announce at start:** "I'm using the c3-component-design skill to explore Component-level impact."

## When Invoked

Called during EXPLORE phase of c3-design when:
- Hypothesis suggests Component-level impact
- Need to understand implementation implications
- Exploring downstream from Container
- Change affects specific technical behavior

Also called by c3-adopt to CREATE initial Component documentation.

---

## Core Principles for Component Documents
- **Abstraction:** Implementation detail. Show HOW components work to satisfy container obligations.
- **Nature-driven focus:** Nature tags are open-ended; choose what best explains behavior (resource, business, cross-cutting, entrypoint, testing, deployment, contextual, etc.).
- **Anchors:** Use `{#com-xxx-*}` anchors for linkable sections.
- **Derivation:** Components implement container protocols/cross-cutting; reference the container sections they fulfill.
- **Environment ownership:** Component doc owns stack details and configuration (with env diffs).

---

## Nature Types (open-ended, pick what fits)

| Nature | Focus |
|--------|-------|
| Resource/Integration | Configuration, env differences, how/why config loaded |
| Business Logic | Domain flows, rules, edge cases |
| Framework/Entrypoint | Mixed concerns: auth, errors, lifecycle |
| Cross-cutting | Integration patterns and how used everywhere |
| Build/Deployment | Build pipeline, deploy config, CI/CD specifics |
| Testing | Strategy, fixtures, mocking approach |
| Contextual | Caching, websocket, feature-specific behavior |
| Other | Add any nature that clarifies the focus |

---

## Required Outputs (Component)
- Overview tying to container protocols/cross-cutting.
- Stack (libraries/versions) with rationale.
- Configuration with env differences (tables encouraged).
- Interfaces & types (signatures, DTOs, events).
- Behavior narrative + at least one diagram when non-trivial.
- Error handling table (retriable? action/code).
- Usage example (code snippet).
- Dependencies and consumed infra features.
- Optional when relevant: health checks and metrics/observability (especially for resource/cross-cutting components).

---

## Template (copy/paste and fill)

````markdown
# COM-XXX <Name> (<Nature>)

## Overview {#com-xxx-overview}
- Responsibility and how it fits container protocols/cross-cutting.

## Stack {#com-xxx-stack}
- Library/version choices; why selected.

## Configuration {#com-xxx-config}
| Env Var | Dev | Prod | Why |
|---------|-----|------|-----|
| ... | ... | ... |

## Interfaces & Types {#com-xxx-interfaces}
- Signatures, DTOs, events; link to schemas.

## Behavior {#com-xxx-behavior}
- Narrative plus diagram where helpful.
```mermaid
stateDiagram-v2
    ... if useful ...
```
- Alternative: flowchart/sequence/ERD depending on need.

## Error Handling {#com-xxx-errors}
| Error | Retriable | Action/Code |
|-------|-----------|-------------|
| ... | ... | ... |

## Usage {#com-xxx-usage}
```typescript
// exemplar usage
```

## Dependencies {#com-xxx-deps}
- Upstream/downstream components; infra features consumed.
````

---

## Checklist
- Nature chosen and informs emphasis.
- Stack and configuration documented with env differences.
- Interfaces/types specified.
- Behavior explained with at least one diagram when non-trivial.
- Error handling table present.
- Usage example included.
- Dependencies and consumed infra features listed.
- Anchors follow `{#com-xxx-*}` scheme.
- Health checks/metrics captured when the component is a resource/cross-cutting element with availability/observability needs.

---

## Diagram Guidance
- Choose diagram type based on need: sequence (interactions), state (lifecycle), flowchart (logic), ERD (data), class diagram (types).
- Include at least one diagram when behavior is non-trivial.
- Align labels with component names and interface names used elsewhere.

---

## Derivation & Linking
- Tie back to the container sections (protocol implementation, cross-cutting) this component fulfills.
- Reference infra features consumed (from CON documents).
- No upward duplication; this is the terminal layer for HOW behavior works.

---

## Exploration Questions (to fill template)
- Purpose: What problem does this solve? If absent, what breaks?
- Implementation: Which library/framework/version? Why this choice?
- Interfaces: What are the signatures/DTOs/events?
- Configuration: Which env vars/config files? What differs dev vs prod?
- Behavior: What states/flows/algorithms matter? Which diagram best communicates it?
- Error handling: What can go wrong? Which errors are retriable? How signaled?
- Usage: How should other components call this? Provide realistic snippet.
- Dependencies: Which other components/infra features does this rely on?
