---
name: c3-context-design
description: Explore Context level impact during scoping - system boundaries, actors, cross-component interactions, and high-level concerns
---

# C3 Context Level Exploration

## Overview

Context is the **eagle-eye introduction** to your architecture. It's the first document readers see and should answer two fundamental questions:

1. **What containers exist and what are they responsible for?**
2. **How do containers interact with each other?**

<context_position>
Layer: ROOT (c3-0)
Parent: None (external world)
Children: All Containers (c3-1, c3-2, ...)

As the introduction:
- I provide the MAP of the system
- I define WHO exists (containers) and HOW they talk (protocols)
- I set boundaries that children inherit
- Changes here are RARE but PROPAGATE to all descendants
</context_position>

**Announce at start:** "I'm using the c3-context-design skill to explore Context-level impact."

---

## Load Settings & Defaults

<chain_prompt id="load_settings">
<instruction>Load project settings and merge with defaults</instruction>

<action>
```bash
# Check for project settings
cat .c3/settings.yaml 2>/dev/null
```
</action>

<merge_logic>
```xml
<settings_merge layer="context">
  <!-- Step 1: Load defaults from this skill's defaults.md -->
  <defaults source="skills/c3-context-design/defaults.md">
    <include>[default include list]</include>
    <exclude>[default exclude list]</exclude>
    <litmus>"Would changing this require coordinating multiple containers or external parties?"</litmus>
    <diagrams>System Context (primary), Container Overview (secondary)</diagrams>
  </defaults>

  <!-- Step 2: Check settings.yaml for context section -->
  <project_settings source=".c3/settings.yaml">
    <if key="context.useDefaults" value="false">
      <!-- Don't load defaults, use only project settings -->
    </if>
    <if key="context.useDefaults" value="true" OR missing>
      <!-- Merge: project settings extend defaults -->
      <include>defaults + context.include</include>
      <exclude>defaults + context.exclude</exclude>
      <litmus>context.litmus OR default</litmus>
      <diagrams>context.diagrams OR default</diagrams>
    </if>
  </project_settings>

  <!-- Step 3: Also load global settings -->
  <global>
    <diagrams_tool>settings.diagrams (e.g., mermaid)</diagrams_tool>
    <guard>settings.guard (team guardrails)</guard>
  </global>
</settings_merge>
```
</merge_logic>

<output>
Display active configuration:
```
Context Layer Configuration:
├── Include: [merged list]
├── Exclude: [merged list]
├── Litmus: [active test]
├── Diagrams: [tool] - [types]
└── Guardrails: [if any]
```
</output>
</chain_prompt>

<apply_throughout>
Use loaded settings when:
- Deciding what belongs at Context level (litmus test)
- Choosing diagram types
- Applying team guardrails
- Writing documentation (guidance)
</apply_throughout>

---

## Context's Two Core Jobs

### Job 1: Define Containers and Responsibilities

Context answers: "What are the major parts of this system?"

```
┌─────────────────────────────────────────────────────────────────┐
│                    CONTAINER INVENTORY                          │
├─────────────────────────────────────────────────────────────────┤
│  Container    │  Type         │  Responsibility (one sentence)  │
├───────────────┼───────────────┼─────────────────────────────────┤
│  c3-1 API     │  Code         │  Handles all client requests    │
│  c3-2 Worker  │  Code         │  Processes background jobs      │
│  c3-3 DB      │  Infra        │  Persists application state     │
└─────────────────────────────────────────────────────────────────┘
```

**What Context says:** "These are the boxes"
**What Container says:** "Here's what's inside each box"

### Job 2: Define How Containers Interact

Context answers: "How do containers talk to each other?"

```
┌─────────────────────────────────────────────────────────────────┐
│                    CONTAINER PROTOCOLS                          │
├─────────────────────────────────────────────────────────────────┤
│  From        │  To           │  Protocol     │  Purpose         │
├──────────────┼───────────────┼───────────────┼──────────────────┤
│  c3-1 API    │  c3-2 Worker  │  Redis Queue  │  Async jobs      │
│  c3-1 API    │  c3-3 DB      │  PostgreSQL   │  Data persistence│
│  c3-2 Worker │  c3-3 DB      │  PostgreSQL   │  Job state       │
└─────────────────────────────────────────────────────────────────┘
```

**What Context says:** "Container A talks to Container B via Protocol X"
**What Container says:** "Our IntegrationClient component implements that protocol"

---

## Hierarchical Position

> See [hierarchy-model.md](../../references/hierarchy-model.md) for full inheritance diagram.

<context_contracts>
| Contract Type | What Context Decides | What Children Inherit |
|---------------|---------------------|----------------------|
| **Containers** | What containers exist | Container docs must exist |
| **Protocols** | How containers communicate | Containers implement provider/consumer |
| **Boundary** | What's inside vs outside | Containers cannot reach outside |
| **Actors** | Who interacts with system | Containers implement interfaces |
| **Cross-cutting** | System-wide patterns | All containers follow these |
</context_contracts>

---

## Critical Decision: Is This a Context-Level Change?

Context changes are **rare**. Most changes happen at Container or Component level.

<extended_thinking>
<goal>Quickly determine if this belongs at Context level</goal>

<context_level_triggers>
CONTEXT level if ANY are true:
- [ ] Adding or removing a container
- [ ] Changing how containers talk to each other (new/changed protocol)
- [ ] Changing system boundary (what's in vs out)
- [ ] Adding a new actor type

CONTAINER level (delegate) if ALL are true:
- [ ] Change is within existing container
- [ ] No new protocols between containers
- [ ] Boundary unchanged
- [ ] Same actors
</context_level_triggers>

<output>
"This is [Context/Container] level because [reason]. [Proceeding/Delegating to c3-container-design]."
</output>
</extended_thinking>

---

## Exploration Process

### Phase 1: Load Current Context

<chain_prompt id="load_context">
<instruction>Read the current Context document</instruction>
<action>
```bash
cat .c3/README.md
```
</action>
<extract>
- Current container inventory
- Current protocols between containers
- Current actors
- Current boundary
</extract>
</chain_prompt>

### Phase 2: Analyze Change Impact

For Context-level changes, impact is usually **system-wide**:

| Change Type | Impact | Action |
|-------------|--------|--------|
| New container | Create container doc | Delegate to c3-container-design |
| Remove container | Remove doc, update protocols | Audit affected protocols |
| New protocol | Both containers affected | Update both container docs |
| Protocol change | All consumers/providers affected | Coordinate updates |
| Boundary change | All containers | Full audit |

### Phase 3: Document Downstream Impact

For each affected container:

```xml
<contract container="c3-{N}">
  <protocol name="[name]" role="[consumer|provider]"/>
  <boundary>
    <can_access>[internal]</can_access>
    <cannot_access>[external]</cannot_access>
  </boundary>
</contract>
```

---

## Socratic Discovery

When building or updating Context, ask:

**For Container Inventory:**
- "What would be separately deployed?"
- "What has its own codebase/runtime?"
- "What data stores exist?"

**For Protocols:**
- "How do containers talk to each other?"
- "Sync (REST/gRPC) or Async (queues/events)?"
- "What's the contract?"

**For Boundary:**
- "What's inside your system vs external?"
- "What external systems do you integrate with?"

**For Actors:**
- "Who/what initiates interactions with the system?"
- "Human users? Other systems? Both?"

---

## Document Template

<prefill_template>
```markdown
---
id: c3-0
c3-version: 3
title: [System Name] Overview
---

# [System Name] Overview

## Overview {#c3-0-overview}
<!--
System purpose in 1-2 sentences.
What problem does it solve? For whom?
-->

## System Boundary {#c3-0-boundary}
<!--
What's INSIDE vs OUTSIDE the system.
-->

### Inside (Our System)
- [Container 1]
- [Container 2]

### Outside (External)
- [External System 1]
- [External System 2]

## Actors {#c3-0-actors}
<!--
Who/what interacts with the system from outside.
-->

| Actor | Type | Interacts Via |
|-------|------|---------------|
| [User] | Human | Web UI (c3-1) |
| [External API] | System | REST (c3-2) |

## Containers {#c3-0-containers}
<!--
JOB 1: Define what containers exist and their responsibilities.
One sentence per container - details live in container docs.
-->

| Container | ID | Type | Responsibility |
|-----------|-----|------|----------------|
| [API Service] | c3-1 | Code | Handles client requests |
| [Worker] | c3-2 | Code | Processes background jobs |
| [Database] | c3-3 | Infra | Persists application state |

## Container Interactions {#c3-0-interactions}
<!--
JOB 2: Define how containers talk to each other.
The actual implementation is in container docs.
-->

| From | To | Protocol | Purpose |
|------|-----|----------|---------|
| c3-1 | c3-2 | Redis Queue | Async job dispatch |
| c3-1 | c3-3 | PostgreSQL | Data persistence |
| c3-2 | c3-3 | PostgreSQL | Job state |

<!--
Optional: System overview diagram if helpful.
Only if 4+ containers or complex interaction patterns.
-->

## Cross-Cutting Concerns {#c3-0-cross-cutting}
<!--
System-wide patterns. Just name them here.
Implementation details in container docs.
-->

- **Auth:** [JWT / OAuth / etc]
- **Logging:** [Structured JSON]
- **Errors:** [Standard error format]
```
</prefill_template>

---

## Impact Assessment Output

<output_format>
```xml
<context_exploration_result>
  <changes>
    <change type="[container|protocol|boundary|actor]">
      [Description]
    </change>
  </changes>

  <downstream_impact>
    <container id="c3-{N}" action="[update|create|remove]">
      <reason>[What this container must do]</reason>
    </container>
  </downstream_impact>

  <delegation>
    <to_skill name="c3-container-design">
      <container_ids>[c3-1, c3-2]</container_ids>
    </to_skill>
  </delegation>
</context_exploration_result>
```
</output_format>

---

## Checklist

<verification_checklist>
- [ ] Container inventory complete (all containers listed)
- [ ] Container responsibilities clear (one sentence each)
- [ ] Protocols between containers documented
- [ ] Actors identified
- [ ] Boundary defined (inside vs outside)
- [ ] Cross-cutting concerns named
- [ ] Downstream containers identified for delegation
</verification_checklist>

---

## Related

- [hierarchy-model.md](../../references/hierarchy-model.md) - C3 layer inheritance
- [v3-structure.md](../../references/v3-structure.md) - Document structure
- [diagram-patterns.md](../../references/diagram-patterns.md) - When/how to diagram
