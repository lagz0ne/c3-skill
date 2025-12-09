---
name: c3-container-design
description: Explore Container level impact during scoping - technology choices, component organization, middleware, and inter-container communication
---

# C3 Container Level Exploration

## Overview

The Container level is the **architectural command center** of C3. It sits at the perfect abstraction level:
- **Full context awareness** from above (knows everything from Context)
- **Complete control** over component responsibilities below
- **Mediator** for all interactions (components implement, container orchestrates)

<container_position>
Layer: MIDDLE (c3-{N})
Parent: Context (c3-0)
Children: Components (c3-{N}01, c3-{N}02, ...)

As the command center:
- I INHERIT everything from Context (boundary, protocols, cross-cutting)
- I DEFINE what exists and why (components, their responsibilities)
- I ORCHESTRATE how things interact (internal + cross-container)
- I DECIDE strategic diagrams (few, critical, high-impact)
</container_position>

**Announce at start:** "I'm using the c3-container-design skill to explore Container-level impact."

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
<settings_merge layer="container">
  <!-- Step 1: Load defaults from this skill's defaults.md -->
  <defaults source="skills/c3-container-design/defaults.md">
    <include>Technology stack, responsibilities, component relationships, data flow, API surface</include>
    <exclude>System boundary (→Context), cross-cutting (→Context), implementation code (→Component)</exclude>
    <litmus>"Is this about WHAT this container does and WITH WHAT, not HOW internally?"</litmus>
    <diagrams>Component Relationships (required), Data Flow (required)</diagrams>
  </defaults>

  <!-- Step 2: Check settings.yaml for container section -->
  <project_settings source=".c3/settings.yaml">
    <if key="container.useDefaults" value="false">
      <!-- Don't load defaults, use only project settings -->
    </if>
    <if key="container.useDefaults" value="true" OR missing>
      <!-- Merge: project settings extend defaults -->
      <include>defaults + container.include</include>
      <exclude>defaults + container.exclude</exclude>
      <litmus>container.litmus OR default</litmus>
      <diagrams>container.diagrams OR default</diagrams>
      <guidance>container.guidance (layer-specific prose)</guidance>
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
Container Layer Configuration:
├── Include: [merged list]
├── Exclude: [merged list]
├── Litmus: [active test]
├── Diagrams: [tool] - [types]
├── Guidance: [layer-specific notes]
└── Guardrails: [if any]
```
</output>
</chain_prompt>

<apply_throughout>
Use loaded settings when:
- Deciding what belongs at Container level (litmus test)
- Making diagram decisions in Phase 4 (override defaults.md if settings specify)
- Applying team guardrails
- Writing documentation (guidance from settings)
- Checking include/exclude rules for content placement
</apply_throughout>

---

## Why Container is the Sweet Spot

<extended_thinking>
<goal>Understand why Container is the most impactful architectural layer</goal>

<rationale>
**Context is too high:**
- Defines boundaries and constraints
- Changes rarely (system-wide impact)
- Doesn't guide day-to-day development

**Component is too low:**
- Implementation details
- Changes frequently
- Trees obscure the forest

**Container is JUST RIGHT:**
- Has all the context needed for decisions
- Defines component responsibilities (readable abstraction)
- Controls interactions (internal and cross-container)
- Where readers get the architectural picture
- Where diagrams have the highest ROI
</rationale>

<architectural_truth>
Inter-container interactions are NOT container-level concerns.
They are COMPONENT-level concerns that Container mediates.

Example: Container A talks to Container B
- Reality: Component in A (e.g., "Integration Client") calls Component in B (e.g., "Request Handler")
- Container A documents: "We have an Integration Client that calls Container B"
- Container B documents: "We expose a Request Handler that others call"
- The actual protocol is in Context (where both containers inherit it from)
</architectural_truth>
</extended_thinking>

---

## Critical Decision: Is This a Container-Level Change?

<extended_thinking>
<goal>Determine if the proposed change belongs at Container level or should escalate/delegate</goal>

<container_level_indicators>
DEFINITELY Container level if ANY are true:
- [ ] Changes technology stack (runtime, framework, language)
- [ ] Reorganizes component structure
- [ ] Modifies internal patterns (error handling, data access)
- [ ] Changes API contracts between components
- [ ] Affects multiple components in the same way
- [ ] Adds/removes a component
- [ ] Changes how this container interacts with others

PROBABLY Context level (escalate UP) if ANY are true:
- [ ] Requires new inter-container protocol
- [ ] Changes how actors interact with the system
- [ ] Violates or needs to change system boundary
- [ ] Requires modifying cross-cutting concerns system-wide

PROBABLY Component level (delegate DOWN) if ALL are true:
- [ ] Change is within a single component
- [ ] Technology stack unchanged
- [ ] Internal patterns followed
- [ ] Interface contract unchanged
</container_level_indicators>

<decision>
IF any_context_indicator THEN escalate to c3-context-design
ELIF any_container_indicator THEN proceed with Container exploration
ELSE delegate to c3-component-design
</decision>

<output>
"This change is at [Context/Container/Component] level because [specific indicator].
[Action]: [Escalating to c3-context-design | Proceeding with exploration | Delegating to c3-component-design]"
</output>
</extended_thinking>

---

## Phase 1: Inherit From Context

<critical>
**ALWAYS START HERE** - Before exploring this container, load what Context expects from it.
</critical>

<chain_prompt id="load_parent_context">
<instruction>Read Context to understand inherited constraints</instruction>
<action>
```bash
cat .c3/README.md
```
</action>
<extract_for_this_container>
```xml
<inherited_from_context container="c3-{N}">
  <boundary>
    <can_access>[internal resources from Context boundary]</can_access>
    <cannot_access>[external to system boundary]</cannot_access>
  </boundary>

  <protocols>
    <implements role="provider">[protocols this container exposes]</implements>
    <implements role="consumer">[protocols this container calls]</implements>
  </protocols>

  <actors>
    <serves actor="[actor]">[endpoints/interfaces]</serves>
  </actors>

  <cross_cutting>
    <must_follow pattern="auth">[how to implement]</must_follow>
    <must_follow pattern="logging">[how to implement]</must_follow>
    <must_follow pattern="errors">[how to implement]</must_follow>
  </cross_cutting>
</inherited_from_context>
```
</extract_for_this_container>
</chain_prompt>

<extended_thinking>
<goal>Verify this container honors Context constraints</goal>

<verification_questions>
1. Does my change violate the system boundary?
2. Do I still implement required protocols correctly?
3. Do I still serve the actors correctly?
4. Am I following cross-cutting patterns?
</verification_questions>

<escalation_triggers>
IF violates_boundary THEN escalate to c3-context-design
IF breaks_protocol_contract THEN escalate to c3-context-design
IF changes_actor_interface THEN escalate to c3-context-design
IF deviates_from_cross_cutting THEN escalate to c3-context-design
</escalation_triggers>
</extended_thinking>

---

## Phase 2: Load and Understand Current Container State

<chain_prompt id="load_container">
<instruction>Read and internalize current Container document</instruction>
<action>
```bash
find .c3 -name "c3-{N}-*" -type d | head -1 | xargs -I {} cat {}/README.md
```
</action>
<extract>
- Current technology stack
- Current component inventory (with responsibilities)
- Current internal patterns
- Current API contracts
- Existing diagrams (what do they show?)
</extract>
</chain_prompt>

<extended_thinking>
<goal>Build mental model of this container's architecture</goal>

<understanding_checklist>
Before proceeding, I must be able to answer:

TECHNOLOGY:
- What runtime/framework is this?
- What are the critical dependencies?
- What constraints does the tech stack impose?

COMPONENTS:
- What are the major components?
- What is each component's single responsibility?
- How do components interact internally?

EXTERNAL INTERACTIONS:
- How does this container expose itself to others?
- What other containers does this container depend on?
- Which COMPONENT handles each external interaction?

PATTERNS:
- What patterns are established (error handling, data access, etc.)?
- Are these patterns enforced consistently?
</understanding_checklist>
</extended_thinking>

---

## Phase 3: Analyze Change Impact

<extended_thinking>
<goal>Determine how change affects this Container and its children</goal>

<analysis_framework>
**UPSTREAM CHECK (to Context):**
Does this change require Context-level updates?
- New protocol needed? → Escalate
- Actor interface change? → Escalate
- Boundary violation? → Escalate
- Cross-cutting deviation? → Escalate

**ISOLATED CHECK (this Container):**
What changes at Container level?
- Technology stack change?
- Internal pattern change?
- API contract change?
- Component organization change?
- New component needed?

**ADJACENT CHECK (sibling Containers):**
Does this affect other Containers?
- Are we changing a component that talks to another container?
- Does the sibling's component need to change too?
- Note: The actual protocol is in Context - if protocol changes, escalate!

**DOWNSTREAM CHECK (to Components):**
What must Components now do?
- New component needed?
- Existing component behavior change?
- Component interface change?
- Component removal?
</analysis_framework>

<decision_tree>
IF upstream_impact THEN escalate to c3-context-design, STOP
IF isolated_only THEN document container change
IF adjacent_impact THEN coordinate with sibling containers
IF downstream_impact THEN prepare delegation to c3-component-design
</decision_tree>
</extended_thinking>

---

## Phase 3b: Analyze Sibling Container Impact

<extended_thinking>
<goal>Understand cross-container interactions as component-mediated</goal>

<key_insight>
Inter-container communication is NOT Container-to-Container.
It's Component-to-Component, mediated by Container.

When Container A "talks to" Container B:
- A Component in Container A (e.g., "B Client") initiates
- A Component in Container B (e.g., "Request Handler") receives
- The protocol they use comes from Context
- Container A documents the client component
- Container B documents the handler component
- Neither container "owns" the interaction - Context does
</key_insight>

<sibling_analysis>
For each cross-container interaction:

1. IDENTIFY THE COMPONENTS
   - Which component in THIS container initiates/handles?
   - Which component in OTHER container responds?

2. VERIFY PROTOCOL ALIGNMENT
   - Is this using a protocol defined in Context?
   - If new protocol needed → Escalate to Context

3. ASSESS CHANGE IMPACT
   - If we change our component, does their component need to change?
   - Coordinate changes, document in both containers
   - Consider ADR if significant
</sibling_analysis>

<sibling_impact_matrix>
| Our Component | Their Container | Their Component | Impact | Action |
|---------------|-----------------|-----------------|--------|--------|
| [name] | c3-{M} | [name] | [none/update/breaking] | [coordinate/ADR] |
</sibling_impact_matrix>
</extended_thinking>

---

## Phase 4: Strategic Diagram Decisions

<critical>
Diagrams at Container level have the **highest architectural ROI**.
This is where readers understand the system.
Choose diagrams strategically - few but impactful.
</critical>

<extended_thinking>
<goal>Deeply analyze what diagrams this container needs and why</goal>

<diagram_philosophy>
There is NO one-size-fits-all diagram.
Different containers need different diagrams based on their complexity and role.

THE GOAL: Complement the prose with visual clarity.
- If prose is clear, skip the diagram
- If relationships are non-obvious, diagram helps
- If there's a critical flow, show it
</diagram_philosophy>
</extended_thinking>

### Diagram Type Analysis

<extended_thinking>
<goal>Understand what each diagram type can and cannot express</goal>

<diagram_capabilities>

**FLOWCHART / BOX-AND-ARROW**
```
┌─────────────────────────────────────────────────────────────────┐
│                     WHAT IT CAN SHOW                            │
├─────────────────────────────────────────────────────────────────┤
│ ✓ Static structure (components exist)                           │
│ ✓ Dependencies (A depends on B)                                 │
│ ✓ Data flow direction (A → B)                                   │
│ ✓ Groupings (subgraphs, boundaries)                             │
│ ✓ Entry/exit points                                             │
├─────────────────────────────────────────────────────────────────┤
│                     WHAT IT CANNOT SHOW                         │
├─────────────────────────────────────────────────────────────────┤
│ ✗ Temporal ordering (what happens first?)                       │
│ ✗ Conditional paths clearly (if/else branching)                 │
│ ✗ Request/response pairing                                      │
│ ✗ Async vs sync distinction                                     │
│ ✗ Error paths                                                   │
├─────────────────────────────────────────────────────────────────┤
│                     BEST FOR                                    │
├─────────────────────────────────────────────────────────────────┤
│ "What exists and what talks to what"                            │
│ Component overview, dependency graphs, system boundaries        │
└─────────────────────────────────────────────────────────────────┘
```

**SEQUENCE DIAGRAM**
```
┌─────────────────────────────────────────────────────────────────┐
│                     WHAT IT CAN SHOW                            │
├─────────────────────────────────────────────────────────────────┤
│ ✓ Temporal ordering (this, then that)                           │
│ ✓ Request/response pairing (call and return)                    │
│ ✓ Participant lifelines                                         │
│ ✓ Sync vs async (solid vs dashed)                               │
│ ✓ Activation (when component is "working")                      │
│ ✓ Alt/opt blocks for conditionals                               │
├─────────────────────────────────────────────────────────────────┤
│                     WHAT IT CANNOT SHOW                         │
├─────────────────────────────────────────────────────────────────┤
│ ✗ Overall structure (only shows participants in one flow)       │
│ ✗ Dependencies (only what's used in THIS flow)                  │
│ ✗ Parallel paths well (gets messy)                              │
│ ✗ Multiple entry points                                         │
│ ✗ State persistence between flows                               │
├─────────────────────────────────────────────────────────────────┤
│                     BEST FOR                                    │
├─────────────────────────────────────────────────────────────────┤
│ "What happens when X occurs"                                    │
│ Critical flows, request handling, multi-step processes          │
└─────────────────────────────────────────────────────────────────┘
```

**STATE DIAGRAM**
```
┌─────────────────────────────────────────────────────────────────┐
│                     WHAT IT CAN SHOW                            │
├─────────────────────────────────────────────────────────────────┤
│ ✓ States an entity can be in                                    │
│ ✓ Transitions between states                                    │
│ ✓ Triggers for transitions                                      │
│ ✓ Terminal states                                               │
├─────────────────────────────────────────────────────────────────┤
│                     WHAT IT CANNOT SHOW                         │
├─────────────────────────────────────────────────────────────────┤
│ ✗ WHO performs the transition                                   │
│ ✗ Multiple entities simultaneously                              │
│ ✗ Data payload on transitions                                   │
│ ✗ Component structure                                           │
├─────────────────────────────────────────────────────────────────┤
│                     BEST FOR                                    │
├─────────────────────────────────────────────────────────────────┤
│ "What states can this entity be in and how does it change"      │
│ Order lifecycle, workflow states, entity status machines        │
└─────────────────────────────────────────────────────────────────┘
```

**TABLE / MATRIX**
```
┌─────────────────────────────────────────────────────────────────┐
│                     WHAT IT CAN SHOW                            │
├─────────────────────────────────────────────────────────────────┤
│ ✓ Mapping between two dimensions                                │
│ ✓ Responsibility assignment                                     │
│ ✓ Feature coverage                                              │
│ ✓ Comparison of options                                         │
├─────────────────────────────────────────────────────────────────┤
│                     WHAT IT CANNOT SHOW                         │
├─────────────────────────────────────────────────────────────────┤
│ ✗ Relationships or flow                                         │
│ ✗ Hierarchy or nesting                                          │
│ ✗ Process or sequence                                           │
├─────────────────────────────────────────────────────────────────┤
│                     BEST FOR                                    │
├─────────────────────────────────────────────────────────────────┤
│ "What does what" or "Who owns what"                             │
│ RACI, responsibility matrix, feature breakdown                  │
└─────────────────────────────────────────────────────────────────┘
```
</diagram_capabilities>
</extended_thinking>

### Diagram Combination Patterns

<extended_thinking>
<goal>When and how to combine diagram types effectively</goal>

<combination_patterns>

**PATTERN 1: Overview + Critical Flow**
When: Container has complex structure AND a non-obvious key flow
Use: Flowchart (structure) + Sequence (one critical path)

Example:
```
Flowchart shows: Handler → Auth → Service → [Cache | DB | Queue]
Sequence shows: "What happens when user places an order" (the flow through those boxes)

WHY THIS WORKS: Structure gives the map, sequence gives the journey.
One without the other: Structure alone doesn't show dynamics, sequence alone doesn't show alternatives.
```

**PATTERN 2: Overview + State Machine**
When: Container manages stateful entities with complex lifecycles
Use: Flowchart (components) + State diagram (entity lifecycle)

Example:
```
Flowchart shows: Order Service → Payment → Fulfillment → Notification
State diagram shows: Order states (pending → paid → shipped → delivered)

WHY THIS WORKS: Components don't reveal state transitions, state diagram doesn't show who changes state.
```

**PATTERN 3: External Interaction + Internal Flow**
When: Container is a boundary service with both external and internal complexity
Use: External diagram (boundaries) + Internal flowchart (components)

Example:
```
External shows: [Frontend] → [This API] → [Payment Service] + [Inventory Service]
Internal shows: Handler → Validator → Orchestrator → [PaymentClient | InventoryClient]

WHY THIS WORKS: External shows the "edges" of this container, internal shows what's inside.
```

**ANTI-PATTERNS**

1. TWO FLOWCHARTS AT DIFFERENT ZOOM LEVELS
   Why bad: Readers can't mentally map between them
   Fix: Use one flowchart with subgraphs, or pick ONE level

2. SEQUENCE FOR EVERY ENDPOINT
   Why bad: Noise, maintenance nightmare
   Fix: Pick 1-2 CRITICAL flows only

3. STATE DIAGRAM FOR SIMPLE STATES
   Why bad: Overkill, prose is clearer
   Fix: Just list states in prose if < 4 states with obvious transitions

4. TABLE + DIAGRAM SHOWING SAME THING
   Why bad: Redundant, will drift apart
   Fix: Pick one representation, not both

</combination_patterns>
</extended_thinking>

### Diagram Decision Framework

<extended_thinking>
<goal>Make explicit, justified diagram decisions for this container</goal>

<decision_framework>

**STEP 1: Characterize the Container**

Questions to answer:
- How many components? (1-3: simple, 4-6: moderate, 7+: complex)
- What type of relationships? (linear, branching, mesh)
- Are there external interactions? (none, few, many)
- Is there a dominant flow? (yes: sequence candidate, no: skip sequence)
- Does it manage stateful entities? (yes: state diagram candidate)

Classification:
```
SIMPLE CONTAINER (skip most diagrams):
- 1-3 components
- Linear relationships (A → B → C)
- 0-1 external interactions
- Obvious flows
→ Probably just needs a component table, maybe one flowchart

MODERATE CONTAINER (selective diagrams):
- 4-6 components
- Some branching
- 2-3 external interactions
- 1-2 critical flows
→ Needs: Component flowchart, maybe one sequence diagram

COMPLEX CONTAINER (strategic diagrams):
- 7+ components OR mesh relationships
- Multiple external interactions
- Multiple critical flows
- Possibly stateful entities
→ Needs: Component flowchart + ONE critical sequence + maybe state diagram
```

**STEP 2: For Each Potential Diagram, Evaluate**

```
DIAGRAM: [type]
┌────────────────────────────────────────────────────────────────┐
│ CLARITY ADDED                                                  │
│ Can prose alone convey this? [yes/no]                          │
│ Is this non-obvious? [yes/no]                                  │
│ Does visual help more than text? [yes/no]                      │
├────────────────────────────────────────────────────────────────┤
│ READER VALUE                                                   │
│ Will readers return to this? [often/sometimes/rarely]          │
│ Does this serve as "north star"? [yes/no]                      │
├────────────────────────────────────────────────────────────────┤
│ MAINTENANCE COST                                               │
│ How often will this change? [rarely/sometimes/often]           │
│ How hard to update? [trivial/moderate/complex]                 │
├────────────────────────────────────────────────────────────────┤
│ DECISION: [INCLUDE / SKIP / SIMPLIFY]                          │
│ Justification: [one sentence why]                              │
└────────────────────────────────────────────────────────────────┘
```

**STEP 3: Check Combinations**

If including multiple diagrams:
- Do they serve different purposes? (not redundant)
- Do they reference same terminology? (consistent naming)
- Can reader mentally connect them? (clear relationship)
- Is total count ≤ 3? (more = noise)

</decision_framework>
</extended_thinking>

### Diagram Decision Output

<extended_thinking>
<goal>Document the diagram decisions with full justification</goal>

For this container, document each decision:

```xml
<diagram_analysis container="c3-{N}">
  <container_characteristics>
    <component_count>[N]</component_count>
    <relationship_type>[linear|branching|mesh]</relationship_type>
    <external_interactions>[N]</external_interactions>
    <has_dominant_flow>[yes|no]</has_dominant_flow>
    <manages_state>[yes|no]</manages_state>
    <classification>[simple|moderate|complex]</classification>
  </container_characteristics>

  <diagram_evaluation type="component_overview">
    <clarity_added>[yes|no] - [reason]</clarity_added>
    <reader_value>[high|medium|low] - [reason]</reader_value>
    <maintenance_cost>[low|medium|high] - [reason]</maintenance_cost>
    <decision>[INCLUDE|SKIP|SIMPLIFY]</decision>
    <justification>[one sentence]</justification>
  </diagram_evaluation>

  <diagram_evaluation type="sequence_critical_flow">
    <which_flow>[name of flow if applicable]</which_flow>
    <clarity_added>[yes|no] - [reason]</clarity_added>
    <reader_value>[high|medium|low] - [reason]</reader_value>
    <maintenance_cost>[low|medium|high] - [reason]</maintenance_cost>
    <decision>[INCLUDE|SKIP|SIMPLIFY]</decision>
    <justification>[one sentence]</justification>
  </diagram_evaluation>

  <diagram_evaluation type="state_diagram">
    <which_entity>[entity if applicable]</which_entity>
    <clarity_added>[yes|no] - [reason]</clarity_added>
    <reader_value>[high|medium|low] - [reason]</reader_value>
    <maintenance_cost>[low|medium|high] - [reason]</maintenance_cost>
    <decision>[INCLUDE|SKIP|SIMPLIFY]</decision>
    <justification>[one sentence]</justification>
  </diagram_evaluation>

  <diagram_evaluation type="external_interactions">
    <clarity_added>[yes|no] - [reason]</clarity_added>
    <reader_value>[high|medium|low] - [reason]</reader_value>
    <maintenance_cost>[low|medium|high] - [reason]</maintenance_cost>
    <decision>[INCLUDE|SKIP|SIMPLIFY]</decision>
    <justification>[one sentence]</justification>
  </diagram_evaluation>

  <combination_check>
    <total_diagrams>[N]</total_diagrams>
    <redundancy_check>[pass|fail - reason]</redundancy_check>
    <consistency_check>[pass|fail - reason]</consistency_check>
    <reader_navigation>[clear|confusing - reason]</reader_navigation>
  </combination_check>

  <final_diagram_list>
    <diagram type="[type]" placement="[section]"/>
    <!-- or -->
    <none reason="[justification for no diagrams]"/>
  </final_diagram_list>
</diagram_analysis>
```
</extended_thinking>

### Diagram Placement

<diagram_placement>
Once decisions are made, diagrams go in Container README.md under specific sections:

| Diagram Type | Placement Section | Anchor |
|--------------|-------------------|--------|
| Component Overview | ## Component Organization | `{#c3-n-organization}` |
| Critical Flow (Sequence) | ## Key Flows | `{#c3-n-flows}` |
| State Diagram | ## Entity Lifecycle | `{#c3-n-lifecycle}` |
| External Interactions | ## External Dependencies | `{#c3-n-external}` |

**Diagram Placement Rules:**
- Diagram comes AFTER the introductory prose for that section
- Diagram comes BEFORE detailed tables/lists
- Caption/legend if diagram uses non-obvious notation
- Reference diagram in prose: "As shown in the diagram below..."
</diagram_placement>

---

## Phase 5: Define Downstream Contracts

<downstream_contract_template>
For each affected component, document what Container now expects:

```xml
<contract component="c3-{N}{NN}">
  <inherits_from>c3-{N}</inherits_from>

  <technology>
    <runtime>[version]</runtime>
    <framework>[version]</framework>
    <must_use>[required libraries]</must_use>
  </technology>

  <patterns>
    <pattern name="[name]">[how to implement]</pattern>
  </patterns>

  <interface>
    <exposes>[methods/endpoints]</exposes>
    <accepts>[input types]</accepts>
    <returns>[output types]</returns>
  </interface>

  <cross_cutting_implementation>
    <auth>[specific implementation for this component]</auth>
    <logging>[specific implementation for this component]</logging>
    <errors>[specific implementation for this component]</errors>
  </cross_cutting_implementation>
</contract>
```
</downstream_contract_template>

---

## Socratic Discovery

<chain_prompt id="socratic_discovery">
<instruction>Ask questions based on container type and gaps</instruction>

<gap_analysis>
Container type: [Code|Infrastructure]

If Code Container:
- Technology stack: [clear|unclear]
- Component organization: [clear|unclear]
- Internal patterns: [clear|unclear]
- API contracts: [clear|unclear]
- Key flows identified: [yes|no]
- Diagram needs assessed: [yes|no]

If Infrastructure Container:
- Engine/version: [clear|unclear]
- Configuration: [clear|unclear]
- Features provided: [clear|unclear]
- Consumers: [clear|unclear]
</gap_analysis>

<question_bank>
**For Code Containers:**
- "What is this container's single main responsibility?"
- "What are the 3-5 most important components and what does each do?"
- "What's the most critical flow through this container?"
- "How does this container interact with others?"
- "Which component handles each external interaction?"

**For Infrastructure Containers:**
- "What engine/version is this?"
- "What features does it provide to other containers?"
- "What components in other containers consume this?"
</question_bank>
</chain_prompt>

---

## Document Templates

### Code Container Template

<prefill_template type="code">
```markdown
---
id: c3-{N}
c3-version: 3
title: [Container Name]
type: code
---

# [Container Name]

## Inherited From Context {#c3-n-inherited}
<!--
What this container inherits from c3-0.
This is NOT optional - it's what Context expects of us.
-->

### Boundary Constraints
- Can access: [from Context]
- Cannot access: [from Context]

### Protocol Obligations
| Protocol | Role | Contract |
|----------|------|----------|
| [name] | provider/consumer | [what we must do] |

### Cross-Cutting Implementation
| Concern | Pattern | Our Implementation |
|---------|---------|-------------------|
| Auth | [from Context] | [how we do it] |
| Logging | [from Context] | [how we do it] |

## Overview {#c3-n-overview}
<!--
Purpose and responsibilities of THIS container.
One paragraph, clear statement of what this container does.
-->

## Technology Stack {#c3-n-stack}
<!--
Technology choices that Components inherit.
-->

- Runtime: [version]
- Framework: [version]
- Language: [version]
- Key Dependencies: [critical libraries]

## Component Organization {#c3-n-organization}
<!--
How components are structured and relate.

DIAGRAM DECISION (from Phase 4 analysis):
- Include component overview diagram if: 4+ components with non-trivial relationships
- Skip if: linear relationships (A → B → C) that prose describes clearly
-->

[Optional: Component overview diagram - mermaid flowchart]

| Component | ID | Responsibility |
|-----------|-----|----------------|
| [Name] | c3-{N}{NN} | [Single-sentence purpose] |

## Internal Patterns {#c3-n-patterns}
<!--
Patterns all components in this container must follow.
-->

### Error Handling
[Pattern that components must use]

### Data Access
[Pattern that components must use]

## Key Flows {#c3-n-flows}
<!--
The 1-2 CRITICAL flows that define this container's behavior.

DIAGRAM DECISION (from Phase 4 analysis):
- Include sequence diagram if: flow involves 3+ components, has branching, or is non-obvious
- Skip if: flow is standard CRUD or obvious from component descriptions
- NEVER diagram every endpoint - only the critical paths
-->

[Optional: Sequence diagram for the ONE most important flow]

## Entity Lifecycle {#c3-n-lifecycle}
<!--
ONLY include this section if container manages stateful entities with complex transitions.

DIAGRAM DECISION (from Phase 4 analysis):
- Include state diagram if: 4+ states with non-obvious transitions
- Skip if: simple status (active/inactive) or linear progression
-->

[Optional: State diagram for core entity]

## External Dependencies {#c3-n-external}
<!--
How this container interacts with others.
Remember: Cross-container = our component talks to their component.

DIAGRAM DECISION (from Phase 4 analysis):
- Include external interaction diagram if: 3+ external container relationships
- Skip if: 0-2 external relationships that table describes clearly
-->

[Optional: External interaction diagram showing this container's boundaries]

| External Container | Our Component | Their Component | Purpose |
|-------------------|---------------|-----------------|---------|
| c3-{M} | [our client] | [their handler] | [why] |

## Components {#c3-n-components}
<!--
Full inventory with links.
Each inherits Container constraints.
-->

| Component | ID | Location |
|-----------|-----|----------|
| [Name] | c3-{N}{NN} | [./c3-{N}{NN}-slug.md] |
```
</prefill_template>

### Infrastructure Container Template

<prefill_template type="infrastructure">
```markdown
---
id: c3-{N}
c3-version: 3
title: [Infrastructure Name]
type: infrastructure
---

# [Infrastructure Name]

## Inherited From Context {#c3-n-inherited}
<!--
What Context expects of this infrastructure.
-->

### Role in System
[How this supports other containers per Context]

## Engine {#c3-n-engine}
[Technology] [Version]

## Configuration {#c3-n-config}
| Setting | Value | Why |
|---------|-------|-----|
| [key] | [value] | [rationale] |

## Features Provided {#c3-n-features}
<!--
What this infrastructure offers to code containers.
-->

| Feature | Used By | Component |
|---------|---------|-----------|
| [feature] | c3-{M} | [their component that uses us] |

<!-- NO COMPONENTS - Infrastructure is a LEAF NODE -->
```
</prefill_template>

---

## Meta-Framework Detection

<chain_prompt id="meta_framework">
<instruction>Detect if this is a meta-framework requiring execution context documentation</instruction>

<detection>
```xml
<meta_framework_check>
  <nextjs if="next.config.* exists OR app/ OR pages/">
    Document by: Server Build-time, Server Runtime, Client Runtime
  </nextjs>
  <nuxt if="nuxt.config.* exists">
    Document by: Server, Client, Universal
  </nuxt>
  <sveltekit if="svelte.config.* exists">
    Document by: Server, Client, Shared
  </sveltekit>
  <remix if="remix.config.* exists">
    Document by: Loader, Action, Component
  </remix>
</meta_framework_check>
```
</detection>

<if_meta_framework>
Add section to template:

```markdown
## Execution Contexts {#c3-n-contexts}
<!--
Meta-framework execution contexts.
Components are tagged with their context.
-->

| Context | Description | Component Examples |
|---------|-------------|-------------------|
| Server Build-time | Runs at build | c3-{N}01 (SSG) |
| Server Runtime | Runs on request | c3-{N}02 (API) |
| Client Runtime | Runs in browser | c3-{N}03 (UI) |
```
</if_meta_framework>
</chain_prompt>

---

## Impact Assessment Output

<extended_thinking>
<goal>Summarize Container exploration for c3-design</goal>

<output_structure>
1. INHERITED CONTEXT (what we must honor)
   - Verified constraints from c3-0
   - Any violations detected?

2. CONTAINER CHANGES
   - What changed at Container level
   - Impact on adjacent containers

3. DIAGRAM DECISIONS
   - Which diagrams are needed (with justification)
   - Which diagrams are NOT needed (with justification)

4. DOWNSTREAM PROPAGATION
   - Which components inherit this change
   - Contracts for each component
   - Delegation list for c3-component-design

5. ESCALATION CHECK
   - Does Context need updating?
   - Does hypothesis need revision?
</output_structure>
</extended_thinking>

<output_format>
```xml
<container_exploration_result container="c3-{N}">
  <inherited_verification>
    <context_constraint type="boundary" honored="[yes|no]"/>
    <context_constraint type="protocol" honored="[yes|no]"/>
    <context_constraint type="cross_cutting" honored="[yes|no]"/>
    <escalation_needed>[yes|no]</escalation_needed>
  </inherited_verification>

  <changes>
    <change type="[stack|pattern|api|organization|component]">
      [Description of what changed]
    </change>
  </changes>

  <adjacent_impact>
    <container id="c3-{M}">
      <our_component>[component handling interaction]</our_component>
      <their_component>[component on their side]</their_component>
      <impact>[description]</impact>
      <coordination>[how to coordinate]</coordination>
    </container>
  </adjacent_impact>

  <diagram_decisions>
    <diagram type="component_overview" include="[yes|no]">
      <reason>[why include or skip]</reason>
    </diagram>
    <diagram type="critical_flow" include="[yes|no]">
      <reason>[why include or skip]</reason>
    </diagram>
    <diagram type="external_interactions" include="[yes|no]">
      <reason>[why include or skip]</reason>
    </diagram>
  </diagram_decisions>

  <downstream_propagation>
    <component id="c3-{N}{NN}" action="[update|create|remove]">
      <inherited_change>[What this component must now do]</inherited_change>
    </component>
  </downstream_propagation>

  <delegation>
    <to_skill name="c3-context-design" if="[escalation needed]">
      <reason>[Why Context needs update]</reason>
    </to_skill>
    <to_skill name="c3-component-design">
      <component_ids>[c3-{N}01, c3-{N}02, ...]</component_ids>
      <reason>[Why these need deeper exploration]</reason>
    </to_skill>
  </delegation>

  <hypothesis_revision needed="[yes|no]">
    <reason>[Why revision needed or not]</reason>
  </hypothesis_revision>
</container_exploration_result>
```
</output_format>

---

## Checklist

<verification_checklist>
Before completing Container exploration:

**Context Inheritance:**
- [ ] Context constraints loaded and verified
- [ ] All inherited contracts still honored
- [ ] No escalation needed (or escalation triggered)

**Container Understanding:**
- [ ] Container type determined (Code/Infrastructure)
- [ ] Current container state fully understood
- [ ] All components and their responsibilities clear

**Change Analysis:**
- [ ] Upstream impact analyzed (Context)
- [ ] Isolated impact documented (this Container)
- [ ] Adjacent impact assessed (sibling Containers)
- [ ] Downstream impact identified (Components)

**Diagram Decisions:**
- [ ] Each potential diagram evaluated (clarity + reference + cost)
- [ ] Decisions documented with justifications
- [ ] Diagrams placed in appropriate sections

**Handoff:**
- [ ] Downstream contracts documented for components
- [ ] Delegation list prepared for c3-component-design
- [ ] Hypothesis revision decision made
</verification_checklist>

---

## Related

- [hierarchy-model.md](../../references/hierarchy-model.md) - C3 layer inheritance
- [derivation-guardrails.md](../../references/derivation-guardrails.md) - Core principles
- [v3-structure.md](../../references/v3-structure.md) - Document structure
- [archetype-hints.md](../../references/archetype-hints.md) - Container types
- [diagram-patterns.md](../../references/diagram-patterns.md) - Diagram syntax reference
