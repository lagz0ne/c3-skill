---
name: c3-audit
description: Audit C3 documentation against codebase reality - find drift between docs and implementation. Also verify docs follow C3 methodology rules (layer content, structure, diagrams).
---

# C3 Audit

## Overview

The audit skill verifies C3 documentation quality from two angles:

1. **Methodology Compliance** - Do docs follow C3 layer rules?
   - Correct content at each layer (Context/Container/Component)
   - Proper structure (IDs, paths, frontmatter)
   - Appropriate diagrams per layer
   - Valid contract chain (no orphans/phantoms)

2. **Implementation Conformance** - Do docs match reality?
   - ADR Implementation - Are accepted ADRs actually implemented?
   - Architecture Conformance - Does code follow what docs describe?
   - Verification Checklists - Are ADR verification items satisfied?

<audit_position>
When to use:
- **Methodology Audit**: After creating/updating C3 docs, before review
- **Full/Container Audit**: After implementation, periodically to catch drift
- **ADR Audit**: After accepting an ADR (to verify it was implemented)
- **ADR-Plan Audit**: After c3-design, before handoff

This is NOT for:
- Active design work (use c3-design)
- Creating new documentation (use layer skills)
</audit_position>

**Announce at start:** "I'm using the c3-audit skill to audit documentation against codebase reality."

---

## Audit Modes

The audit can run in different modes based on what you want to verify:

| Mode | Purpose | When to Use |
|------|---------|-------------|
| **Methodology** | Verify docs follow C3 layer rules | After creating/updating docs |
| **Full** | Complete audit of all docs vs code | Periodic health check |
| **ADR** | Verify specific ADR was implemented | After ADR implementation |
| **ADR-Plan** | Verify ADR ↔ Plan coherence | After c3-design, before handoff |
| **Container** | Audit single container and its components | After container changes |
| **Quick** | High-level inventory check only | Fast sanity check |

---

## Mode 0: Methodology Audit

Verify that C3 documentation follows the C3 methodology rules - correct content at each layer, proper structure, and appropriate diagrams.

**Use this mode:**
- After creating new C3 documentation
- After significant doc updates
- To validate documentation quality before review
- When onboarding to ensure docs follow standards

### Phase 1: Load Layer Defaults

<chain_prompt id="load_defaults">
<instruction>Load the layer rules from plugin references</instruction>

<rules>
**Context Layer (c3-0) - Bird's-eye view**

MUST INCLUDE:
- Container responsibilities (WHY each exists)
- Container relationships (how they connect)
- Connecting points (APIs, protocols, events)
- External actors (who/what interacts)
- System boundary (inside vs outside)

MUST EXCLUDE:
- Component lists (push to Container)
- How containers work internally (push to Container)
- Implementation details (push to Component)

LITMUS: "Is this about WHY containers exist and HOW they relate?"

DIAGRAMS: System Context, Container Overview
AVOID: Sequence with methods, class diagrams, flowcharts with logic

---

**Container Layer (c3-N) - Inside view**

MUST INCLUDE:
- Component responsibilities (WHAT each does)
- Component relationships (how they interact)
- Data flows (how data moves)
- Business flows (workflows spanning components)
- Inner patterns (logging, config, errors)

MUST EXCLUDE:
- WHY this container exists (push to Context)
- Container-to-container details (push to Context)
- HOW components work (push to Component)

LITMUS: "Is this about WHAT components do and HOW they relate?"

DIAGRAMS: Component Relationships, Data Flow
AVOID: System context, actor diagrams

---

**Component Layer (c3-NNN) - Close-up view**

MUST INCLUDE:
- Flows (step-by-step processing)
- Dependencies (what it calls)
- Decision logic (branching points)
- Edge cases (non-obvious scenarios)
- Error handling (what can go wrong)

MUST EXCLUDE:
- WHAT this component does (already in Container)
- Component relationships (push to Container)
- Container relationships (push to Context)

LITMUS: "Is this about HOW this component implements its contract?"

DIAGRAMS: Flowcharts, Sequence (to dependencies), State charts
AVOID: System context, container overview
</rules>
</chain_prompt>

### Phase 2: Structure Compliance Check

<chain_prompt id="check_structure">
<instruction>Verify file structure, IDs, and frontmatter</instruction>

<action>
```bash
# List all C3 docs
find .c3 -name "*.md" -type f | sort
```
</action>

<checks>
**ID Pattern Checks:**
| Level | Pattern | Regex |
|-------|---------|-------|
| Context | `c3-0` | `^c3-0$` |
| Container | `c3-{N}` | `^c3-[1-9]$` |
| Component | `c3-{N}{NN}` | `^c3-[1-9][0-9]{2}$` |
| ADR | `adr-{YYYYMMDD}-{slug}` | `^adr-[0-9]{8}-[a-z0-9-]+$` |

**Path Pattern Checks:**
| Level | Expected Path |
|-------|---------------|
| Context | `.c3/README.md` |
| Container | `.c3/c3-{N}-{slug}/README.md` |
| Component | `.c3/c3-{N}-{slug}/c3-{N}{NN}-{slug}.md` |

**Frontmatter Required Fields:**
| Level | Required Fields |
|-------|-----------------|
| Context | `id: c3-0`, `title` |
| Container | `id: c3-N`, `title`, `parent: c3-0` |
| Component | `id: c3-NNN`, `title`, `parent: c3-N`, `type: component` |
</checks>

<output>
```xml
<structure_audit>
  <id_pattern_violations>
    <violation doc=".c3/c3-1-api/c3-1a-handler.md">
      <expected>c3-101, c3-102, etc.</expected>
      <actual>c3-1a</actual>
      <fix>Rename to c3-101-handler.md with id: c3-101</fix>
    </violation>
  </id_pattern_violations>

  <path_violations>
    <violation doc=".c3/components/auth.md">
      <expected>.c3/c3-N-slug/c3-NNN-slug.md</expected>
      <actual>Flat structure in components/</actual>
      <fix>Move to container folder</fix>
    </violation>
  </path_violations>

  <frontmatter_violations>
    <violation doc=".c3/c3-1-api/README.md">
      <missing>parent</missing>
      <fix>Add `parent: c3-0` to frontmatter</fix>
    </violation>
  </frontmatter_violations>
</structure_audit>
```
</output>
</chain_prompt>

### Phase 3: Layer Content Audit

<extended_thinking>
<goal>Verify each document contains appropriate content for its layer</goal>

<audit_process>
For each document, determine its layer from ID, then audit content:

**Context (c3-0) Audit:**

| Check | How to Detect | Violation |
|-------|---------------|-----------|
| Has container inventory | Table/list of containers with IDs | Missing = FAIL |
| Has actor list | External actors section | Missing = WARN |
| Has protocols | Container interactions table | Missing = FAIL |
| NO component lists | Mentions c3-NNN IDs in body | Found = VIOLATION (push down) |
| NO implementation details | Code blocks, algorithms | Found = VIOLATION (push down) |

**Container (c3-N) Audit:**

| Check | How to Detect | Violation |
|-------|---------------|-----------|
| Has component inventory | Table/list of components with IDs | Missing = FAIL |
| Has tech stack | Technology section | Missing = WARN |
| Has data/business flows | Flow descriptions or diagrams | Missing = WARN |
| NO "why container exists" | Purpose statement without context ref | Found = VIOLATION (push up) |
| NO step-by-step flows | Detailed algorithms for single component | Found = VIOLATION (push down) |

**Component (c3-NNN) Audit:**

| Check | How to Detect | Violation |
|-------|---------------|-----------|
| Has flow/algorithm | Step-by-step or flowchart | Missing = FAIL |
| Has dependencies | Dependencies table/list | Missing = WARN |
| Has edge cases | Edge cases section | Missing = WARN |
| Has error handling | Error handling section | Missing = WARN |
| References container contract | "As defined in c3-N" or similar | Missing = WARN |
| NO component relationships | Diagrams showing sibling components | Found = VIOLATION (push up) |
| NO container relationships | Mentions other containers | Found = VIOLATION (push up) |
</audit_process>

<output>
```xml
<layer_content_audit>
  <context doc=".c3/README.md">
    <required_present>
      <item check="container_inventory">PASS</item>
      <item check="protocols">PASS</item>
    </required_present>
    <violations>
      <violation type="content_too_detailed" severity="high">
        <location>Line 45-60</location>
        <content>Lists components of Backend (c3-101, c3-102...)</content>
        <rule>Context should NOT list components</rule>
        <fix>Move component inventory to Container doc</fix>
      </violation>
    </violations>
  </context>

  <container doc=".c3/c3-1-api/README.md" id="c3-1">
    <required_present>
      <item check="component_inventory">PASS</item>
      <item check="tech_stack">PASS</item>
    </required_present>
    <violations>
      <violation type="abstraction_leak_down" severity="medium">
        <location>Line 80-120</location>
        <content>Detailed sync algorithm with decision tree</content>
        <rule>Container documents WHAT, not HOW</rule>
        <fix>Move algorithm to component doc (c3-105)</fix>
      </violation>
    </violations>
  </container>

  <component doc=".c3/c3-1-api/c3-105-sync.md" id="c3-105">
    <required_present>
      <item check="flow">PASS</item>
      <item check="dependencies">PASS</item>
      <item check="edge_cases">PASS</item>
    </required_present>
    <violations>
      <violation type="abstraction_leak_up" severity="medium">
        <location>Line 15-30</location>
        <content>Describes relationships with all sibling components</content>
        <rule>Component docs focus on HOW this component works</rule>
        <fix>Move component relationships to Container doc</fix>
      </violation>
    </violations>
  </component>
</layer_content_audit>
```
</output>
</extended_thinking>

### Phase 4: Diagram Appropriateness Audit

<extended_thinking>
<goal>Verify diagrams match their layer's appropriate types</goal>

<diagram_rules>
| Layer | Appropriate | Inappropriate |
|-------|-------------|---------------|
| **Context** | System Context, Container Overview, Actor diagrams | Sequence with methods, Flowcharts with logic, Class diagrams |
| **Container** | Component Relationships, Data Flow, Sequence (high-level) | System Context, Actor diagrams, Detailed flowcharts |
| **Component** | Flowcharts, Sequence (to deps), State machines | System Context, Container Overview, Component relationships |
</diagram_rules>

<detection_patterns>
**Mermaid diagram type detection:**
- `C4Context` / `C4Container` → System/Container level
- `flowchart` with decision nodes → Flowchart
- `sequenceDiagram` → Sequence
- `stateDiagram` → State machine
- `graph` with components → Component relationships

**Inappropriate patterns:**
- Context with `sequenceDiagram` showing method calls → TOO DETAILED
- Container with `C4Context` showing actors → WRONG LEVEL
- Component with `graph` showing sibling relationships → WRONG LEVEL
</detection_patterns>

<output>
```xml
<diagram_audit>
  <doc path=".c3/README.md" layer="context">
    <diagrams_found>
      <diagram type="C4Context" line="50">APPROPRIATE</diagram>
      <diagram type="sequenceDiagram" line="80">
        <verdict>INAPPROPRIATE</verdict>
        <reason>Sequence diagram with method calls too detailed for Context</reason>
        <fix>Move to Container or remove method details</fix>
      </diagram>
    </diagrams_found>
  </doc>

  <doc path=".c3/c3-1-api/README.md" layer="container">
    <diagrams_found>
      <diagram type="flowchart" line="60">APPROPRIATE</diagram>
    </diagrams_found>
  </doc>

  <doc path=".c3/c3-1-api/c3-105-sync.md" layer="component">
    <diagrams_found>
      <diagram type="flowchart" line="30">APPROPRIATE</diagram>
      <diagram type="stateDiagram" line="80">APPROPRIATE</diagram>
    </diagrams_found>
    <missing_recommended>
      <diagram type="dependencies_sequence">Dependencies table exists but no sequence showing calls</diagram>
    </missing_recommended>
  </doc>
</diagram_audit>
```
</output>
</extended_thinking>

### Phase 5: Contract Chain Verification

<extended_thinking>
<goal>Verify each layer properly references and implements its parent's contract</goal>

<contract_chain_checks>
**Container → Context:**
- Container must be listed in Context's container inventory
- Container's "inherited from context" should match Context's definition
- Container's external dependencies should match Context's protocols

**Component → Container:**
- Component must be listed in Container's component inventory
- Component should reference "contract from Container"
- Component's responsibilities should match what Container says

**Orphan Detection:**
- Containers not listed in Context = ORPHAN
- Components not listed in their Container = ORPHAN
- Containers in Context but no folder = PHANTOM
- Components in Container but no file = PHANTOM
</contract_chain_checks>

<output>
```xml
<contract_chain_audit>
  <context_to_container>
    <container id="c3-1">
      <listed_in_context>yes</listed_in_context>
      <folder_exists>yes</folder_exists>
      <status>VALID</status>
    </container>
    <container id="c3-3">
      <listed_in_context>yes</listed_in_context>
      <folder_exists>no</folder_exists>
      <status>PHANTOM - documented but no folder</status>
      <fix>Create .c3/c3-3-slug/ or remove from Context</fix>
    </container>
  </context_to_container>

  <container_to_component>
    <component id="c3-105" container="c3-1">
      <listed_in_container>yes</listed_in_container>
      <file_exists>yes</file_exists>
      <references_contract>yes - "As defined in Container..."</references_contract>
      <status>VALID</status>
    </component>
    <component id="c3-107" container="c3-1">
      <listed_in_container>no</listed_in_container>
      <file_exists>yes</file_exists>
      <status>ORPHAN - file exists but not in Container inventory</status>
      <fix>Add to Container's component table or delete file</fix>
    </component>
  </container_to_component>

  <inheritance_mismatches>
    <mismatch container="c3-2">
      <context_says>Communicates via REST</context_says>
      <container_says>Inherited: gRPC</container_says>
      <fix>Align Container's inherited section with Context</fix>
    </mismatch>
  </inheritance_mismatches>
</contract_chain_audit>
```
</output>
</extended_thinking>

### Phase 6: Load Layer Skills for Suggestions

**BEFORE making suggestions, load the authoritative layer skills.**

The layer skills are the source of truth for what each layer requires. Loading them ensures suggestions stay in sync with actual requirements.

<chain_prompt id="load_layer_skills">
<instruction>Load all three layer skills to extract current requirements</instruction>

<action>
```bash
# Load Context skill for c3-0 requirements
cat skills/c3-context-design/SKILL.md

# Load Container skill for c3-N requirements
cat skills/c3-container-design/SKILL.md

# Load Component skill for c3-NNN requirements
cat skills/c3-component-design/SKILL.md
```
</action>

<extract>
From each skill, extract:
1. **Template section** - Required document structure
2. **Diagram Requirements** - What diagrams are mandatory
3. **The Principle section** - Layer integrity rules
4. **MUST INCLUDE / MUST EXCLUDE** - Content rules
5. **Socratic Discovery questions** - What the layer should answer

```xml
<layer_requirements>
  <context source="c3-context-design/SKILL.md">
    <required_sections>[from Template]</required_sections>
    <required_diagrams>[from Diagram Requirement]</required_diagrams>
    <must_include>[from The Principle + Template]</must_include>
    <must_exclude>[from The Principle]</must_exclude>
    <integrity_rule>[from The Principle]</integrity_rule>
  </context>

  <container source="c3-container-design/SKILL.md">
    <required_sections>[from Template]</required_sections>
    <required_diagrams>[from Diagram Requirements - TWO required]</required_diagrams>
    <must_include>[from The Principle + Template]</must_include>
    <must_exclude>[from The Principle]</must_exclude>
    <integrity_rule>[from The Principle]</integrity_rule>
  </container>

  <component source="c3-component-design/SKILL.md">
    <required_sections>[from Template]</required_sections>
    <required_diagrams>[from Documentation Principles]</required_diagrams>
    <must_include>[from The Principle + Template]</must_include>
    <must_exclude>[from Documentation Principles - NO CODE]</must_exclude>
    <integrity_rule>[from The Principle + Integrity Check]</integrity_rule>
  </component>
</layer_requirements>
```
</extract>
</chain_prompt>

### Phase 7: Generate Layer-Specific Suggestions

**Use the loaded layer requirements to generate suggestions.**

For each violation found in earlier phases, match it to the corresponding layer skill's requirements and generate a specific fix suggestion.

#### Suggestion Generation Process

```
For each violation:
1. Identify which layer (Context/Container/Component)
2. Look up that layer's requirements from Phase 6
3. Find the specific requirement that was violated
4. Generate suggestion pointing to the skill's Template or example
```

#### Example Suggestions (derived from layer skills)

**Context (from c3-context-design):**
| Issue | Suggestion (Reference) |
|-------|------------------------|
| Missing diagram | See c3-context-design "Diagram Requirement" - add mermaid showing containers, external systems, protocols, actors |
| Missing container inventory | See c3-context-design "Template" - add Containers table with ID, Archetype, Responsibility |
| Component IDs found | Violates c3-context-design "The Principle" - Context defines WHAT containers exist, not their internals |

**Container (from c3-container-design):**
| Issue | Suggestion (Reference) |
|-------|------------------------|
| Missing external diagram | See c3-container-design "Diagram Requirements" - External Relationships diagram is REQUIRED |
| Missing internal diagram | See c3-container-design "Diagram Requirements" - Internal Component diagram is REQUIRED |
| Missing component inventory | See c3-container-design "Template" - add Components table |
| Step-by-step algorithms | Violates c3-container-design "The Principle" - Container defines WHAT, not HOW |

**Component (from c3-component-design):**
| Issue | Suggestion (Reference) |
|-------|------------------------|
| Missing contract | See c3-component-design "Template" - add Contract section referencing Container |
| Missing flow diagram | See c3-component-design "Documentation Principles" - every component SHOULD have flow diagram |
| Code found | Violates c3-component-design "Documentation Principles" - NO CODE, move to .c3/references/ |
| Missing dependencies | See c3-component-design "Template" - add Dependencies table |

**Key benefit:** Suggestions always match current skill requirements. When skills evolve, audit suggestions automatically align.

### Phase 8: Methodology Audit Report

<report_template>
```markdown
# C3 Methodology Audit Report

**Date:** YYYY-MM-DD
**Target:** [.c3/ path or specific container]

## Executive Summary

| Category | Pass | Warn | Fail |
|----------|------|------|------|
| Structure (IDs, paths, frontmatter) | N | N | N |
| Layer Content | N | N | N |
| Diagrams | N | N | N |
| Contract Chain | N | N | N |

**Overall:** COMPLIANT / NEEDS_FIXES / NON_COMPLIANT

## Structure Violations

| Doc | Issue | Severity | Fix |
|-----|-------|----------|-----|
| path | ID pattern invalid | High | Rename to c3-NNN |

## Layer Content Violations

### Abstraction Leaks Down (too detailed)

| Doc | Layer | Content | Should Be In |
|-----|-------|---------|--------------|
| .c3/README.md | Context | Component list | Container |

### Abstraction Leaks Up (belongs higher)

| Doc | Layer | Content | Should Be In |
|-----|-------|---------|--------------|
| c3-105-sync.md | Component | Sibling relationships | Container |

### Missing Required Content

| Doc | Layer | Missing | Suggestion |
|-----|-------|---------|------------|
| c3-2/README.md | Container | Component inventory | Add table with component IDs |

## Diagram Issues

| Doc | Diagram | Issue | Fix |
|-----|---------|-------|-----|
| README.md | sequenceDiagram L80 | Too detailed for Context | Move to Container |

## Contract Chain Issues

| Type | ID | Issue | Fix |
|------|-----|-------|-----|
| PHANTOM | c3-3 | In Context, no folder | Create folder or remove |
| ORPHAN | c3-107 | File exists, not in Container | Add to inventory |
| MISMATCH | c3-2 | Protocol differs from Context | Align inherited section |

## Layer-Specific Recommendations

### Context (c3-0)
[List specific suggestions from Phase 6 Context table]

### Containers
[List specific suggestions for each container from Phase 6 Container table]

### Components
[List specific suggestions for each component from Phase 6 Component table]

## Priority Actions

1. **High:** [Critical fixes - phantoms, orphans, missing required content]
2. **Medium:** [Abstraction leaks, diagram issues]
3. **Low:** [Warnings, optional improvements]
```
</report_template>

---

## Mode 1: Full Audit

### Phase 1: Load Documentation State

<chain_prompt id="load_docs">
<instruction>Load all C3 documentation and build inventory</instruction>

<action>
```bash
# Read Context
cat .c3/README.md

# List all container docs
find .c3 -name "README.md" -path "*/c3-[0-9]-*/*"

# List all component docs
find .c3 -name "c3-[0-9][0-9][0-9]-*.md"

# List all ADRs
find .c3/adr -name "adr-*.md" 2>/dev/null
```
</action>

<extract>
```xml
<doc_inventory>
  <context path=".c3/README.md">
    <containers_listed>[c3-1, c3-2, ...]</containers_listed>
    <protocols_listed>[list]</protocols_listed>
  </context>

  <containers>
    <container id="c3-{N}" path=".c3/c3-{N}-slug/README.md">
      <type>code|infrastructure</type>
      <components_listed>[c3-{N}01, c3-{N}02]</components_listed>
      <technology_documented>[stack]</technology_documented>
    </container>
  </containers>

  <adrs>
    <adr id="adr-YYYYMMDD-slug" status="proposed|accepted|implemented">
      <changes_across_layers>[what should change]</changes_across_layers>
      <verification_checklist>[items to verify]</verification_checklist>
    </adr>
  </adrs>
</doc_inventory>
```
</extract>
</chain_prompt>

### Phase 2: Explore Codebase

<chain_prompt id="explore_codebase">
<instruction>Discover actual codebase structure</instruction>

<action>
Use Task tool with subagent_type="Explore" (very thorough):

```
Explore the codebase to identify:
1. All deployable units (containers) - look for package.json, go.mod, Dockerfile, etc.
2. Major code modules within each (components) - handlers, services, repositories
3. Infrastructure from docker-compose or config
4. Technology stack of each container

Return structured inventory of what actually exists in code.
```
</action>

<extract>
```xml
<code_inventory>
  <containers>
    <container path="/api" type="code">
      <technology>Node.js 20, Express 4.x</technology>
      <components>
        <component path="/api/src/handlers"/>
        <component path="/api/src/services"/>
        <component path="/api/src/db"/>
      </components>
    </container>
  </containers>
</code_inventory>
```
</extract>
</chain_prompt>

### Phase 3: Compare & Find Drift

<extended_thinking>
<goal>Systematically compare documented vs actual state</goal>

<comparison_categories>

**1. INVENTORY DRIFT**
Doc says X exists, but code doesn't (or vice versa)

| Finding | Severity | Example |
|---------|----------|---------|
| Phantom container | Critical | c3-3 documented, no /frontend code |
| Undocumented container | High | /services/notify exists, no c3 doc |
| Phantom component | High | c3-102 documented, code removed |
| Undocumented component | Medium | New handler added, not in docs |

**2. TECHNOLOGY DRIFT**
Doc says technology X, code uses Y

| Finding | Severity | Example |
|---------|----------|---------|
| Framework mismatch | High | Doc: Express, Code: Fastify |
| Version mismatch | Medium | Doc: Node 18, Code: Node 20 |
| Dependency drift | Low | New major dependency not documented |

**3. STRUCTURE DRIFT**
Code organization doesn't match documented patterns

| Finding | Severity | Example |
|---------|----------|---------|
| Pattern violation | Medium | Doc says repository pattern, code has direct DB calls |
| Component boundary blur | Medium | Service A imports Service B internals |

**4. PROTOCOL DRIFT**
Inter-container communication differs from Context

| Finding | Severity | Example |
|---------|----------|---------|
| Protocol mismatch | High | Context: REST, Code: gRPC |
| Undocumented protocol | High | New queue added between containers |

</comparison_categories>

<output>
```xml
<drift_findings>
  <finding type="phantom" severity="critical">
    <doc_id>c3-3</doc_id>
    <expected>/frontend</expected>
    <actual>NOT FOUND</actual>
    <action>Remove doc OR implement code</action>
  </finding>

  <finding type="technology_mismatch" severity="high">
    <doc_id>c3-1</doc_id>
    <documented>Express 4.x</documented>
    <actual>Fastify 4.x</actual>
    <action>Update doc technology section</action>
  </finding>
</drift_findings>
```
</output>
</extended_thinking>

---

## Mode 2: ADR Audit

When verifying a specific ADR was implemented correctly.

### Phase 1: Load the ADR

<chain_prompt id="load_adr">
<instruction>Load ADR and extract verification requirements</instruction>

<action>
```bash
cat .c3/adr/adr-YYYYMMDD-slug.md
```
</action>

<extract>
```xml
<adr_requirements>
  <id>adr-YYYYMMDD-slug</id>
  <status>[should be accepted or implemented]</status>

  <changes_across_layers>
    <context_changes>
      <change doc="c3-0">[what should change]</change>
    </context_changes>
    <container_changes>
      <change doc="c3-N-slug">[what should change]</change>
    </container_changes>
    <component_changes>
      <change doc="c3-NNN-slug">[what should change]</change>
    </component_changes>
  </changes_across_layers>

  <verification_checklist>
    <item>[verification item 1]</item>
    <item>[verification item 2]</item>
  </verification_checklist>
</adr_requirements>
```
</extract>
</chain_prompt>

### Phase 2: Verify Document Changes

<extended_thinking>
<goal>Verify each documented change was made</goal>

<verification_process>
For each item in "Changes Across Layers":

1. Read the target document
2. Check if the documented change exists
3. Mark as: IMPLEMENTED | PARTIAL | MISSING

```xml
<doc_change_verification>
  <change doc="c3-0" expected="Add new actor: Webhook System">
    <status>IMPLEMENTED</status>
    <evidence>Found in Actors table line 45</evidence>
  </change>

  <change doc="c3-1-api" expected="Add webhook handler component">
    <status>PARTIAL</status>
    <evidence>Component listed but missing details</evidence>
    <gap>Missing interface documentation</gap>
  </change>

  <change doc="c3-102-webhook" expected="Create component doc">
    <status>MISSING</status>
    <evidence>File does not exist</evidence>
  </change>
</doc_change_verification>
```
</verification_process>
</extended_thinking>

### Phase 3: Verify Against Code

<extended_thinking>
<goal>Verify code matches what ADR and updated docs describe</goal>

<code_verification_process>
For each verification checklist item in the ADR:

1. Identify what code artifact to check
2. Explore/read the code
3. Verify it matches the documented architecture

For structural checks:
- Does the component exist in code?
- Does it follow the documented patterns?
- Are the interfaces as documented?

For behavioral checks:
- Does the flow work as documented?
- Are protocols implemented correctly?
</code_verification_process>

<output>
```xml
<code_verification>
  <item check="Is webhook handler at correct abstraction level?">
    <status>PASS</status>
    <evidence>Located at /api/src/handlers/webhook.ts, follows handler pattern</evidence>
  </item>

  <item check="Does webhook handler use documented error pattern?">
    <status>FAIL</status>
    <evidence>Uses custom error class instead of documented AppError</evidence>
    <action>Refactor to use AppError from error handling pattern</action>
  </item>

  <item check="Are downstream consumers updated?">
    <status>PASS</status>
    <evidence>NotificationService updated to receive webhook events</evidence>
  </item>
</code_verification>
```
</output>
</extended_thinking>

### Phase 4: ADR Status Recommendation

<extended_thinking>
<goal>Determine if ADR can be marked as implemented</goal>

<decision_matrix>
| Doc Changes | Code Verification | Recommendation |
|-------------|-------------------|----------------|
| All IMPLEMENTED | All PASS | Mark ADR as `implemented` |
| Some PARTIAL/MISSING | Any | Do NOT mark, list gaps |
| All IMPLEMENTED | Some FAIL | Do NOT mark, fix code issues |
</decision_matrix>

<output>
```xml
<adr_audit_result adr="adr-YYYYMMDD-slug">
  <doc_changes_complete>yes|no</doc_changes_complete>
  <code_verification_pass>yes|no</code_verification_pass>

  <recommendation>
    [READY TO MARK IMPLEMENTED | GAPS REMAIN]
  </recommendation>

  <gaps_if_any>
    <gap type="doc">[missing doc change]</gap>
    <gap type="code">[code verification failure]</gap>
  </gaps_if_any>

  <next_steps>
    1. [Fix gap 1]
    2. [Fix gap 2]
    3. Re-run audit
  </next_steps>
</adr_audit_result>
```
</output>
</extended_thinking>

### Phase 5: Status Transition (if audit passes)

If the ADR audit passes completely (all doc changes implemented, all code verified):

**Automatically transition status:**

```bash
# Update ADR frontmatter from 'accepted' to 'implemented'
sed -i 's/^status: accepted$/status: implemented/' .c3/adr/adr-YYYYMMDD-slug.md

# Rebuild TOC to include the newly-implemented ADR
# (Use plugin's build-toc.sh script)
```

**Announce to user:**
- "ADR `adr-YYYYMMDD-slug` has been verified and marked as `implemented`."
- "The ADR will now appear in the Table of Contents."

**If audit fails:** Do NOT transition status. Report gaps and next steps.

---

## Mode 3: ADR-Plan Coherence Audit

Verify the mutual reference between ADR and Implementation Plan is complete and correct.

**When to use:**
- After c3-design creates an ADR+Plan
- Before handoff to ensure quality
- When reviewing someone else's ADR

### Phase 1: Load ADR and Extract Sections

<chain_prompt id="load_adr_plan">
<instruction>Load ADR and extract both ADR sections and Implementation Plan</instruction>

<action>
```bash
cat .c3/adr/adr-YYYYMMDD-slug.md
```
</action>

<extract>
```xml
<adr_plan_structure>
  <id>adr-YYYYMMDD-slug</id>

  <!-- ADR sections (medium abstraction) -->
  <changes_across_layers>
    <change layer="context" doc="c3-0">[what changes]</change>
    <change layer="container" doc="c3-N-slug">[what changes]</change>
    <change layer="component" doc="c3-NNN-slug">[what changes]</change>
  </changes_across_layers>

  <verification>
    <item>[verification item 1]</item>
    <item>[verification item 2]</item>
  </verification>

  <!-- Implementation Plan sections (low abstraction) -->
  <implementation_plan>
    <code_changes>
      <change layer_ref="c3-0" location="path/to/file" action="create|modify|delete">[details]</change>
      <change layer_ref="c3-N-slug" location="path/to/file" action="create|modify|delete">[details]</change>
    </code_changes>

    <dependencies>
      <step order="1">[first thing]</step>
      <step order="2" depends="1">[second thing]</step>
    </dependencies>

    <acceptance_criteria>
      <criterion verification_ref="item 1">[testable criterion]</criterion>
      <criterion verification_ref="item 2">[testable criterion]</criterion>
    </acceptance_criteria>
  </implementation_plan>
</adr_plan_structure>
```
</extract>
</chain_prompt>

### Phase 2: Verify Mapping Completeness

<extended_thinking>
<goal>Verify every ADR item has a corresponding Plan item</goal>

<mapping_checks>

**1. Changes Across Layers → Code Changes mapping**

For each item in "Changes Across Layers":
- Does a Code Change exist that references this layer change?
- Is the code location specific (file:function, not vague)?

| Layer Change | Has Code Change? | Location Specific? | Status |
|--------------|------------------|-------------------|--------|
| c3-0: [change] | yes/no | yes/no | PASS/FAIL |
| c3-N: [change] | yes/no | yes/no | PASS/FAIL |

**2. Verification → Acceptance Criteria mapping**

For each item in "Verification":
- Does an Acceptance Criterion exist that references this?
- Is the criterion testable (command/test, not "should work")?

| Verification Item | Has Criterion? | Testable? | Status |
|-------------------|----------------|-----------|--------|
| [item 1] | yes/no | yes/no | PASS/FAIL |
| [item 2] | yes/no | yes/no | PASS/FAIL |

**3. Orphan Detection**

Code Changes without Layer Change reference = ORPHAN (suspicious)
Acceptance Criteria without Verification reference = ORPHAN (suspicious)

</mapping_checks>

<coherence_score>
Calculate coherence:
- Layer Changes mapped: X/Y
- Verifications mapped: A/B
- Orphans found: N

COHERENT if: All mapped (X=Y, A=B) AND no orphans
PARTIAL if: Most mapped but some missing
INCOHERENT if: Major gaps or many orphans
</coherence_score>

<output>
```xml
<adr_plan_coherence adr="adr-YYYYMMDD-slug">
  <layer_to_code_mapping>
    <total_layer_changes>N</total_layer_changes>
    <mapped_to_code>M</mapped_to_code>
    <unmapped>[list of unmapped layer changes]</unmapped>
  </layer_to_code_mapping>

  <verification_to_criteria_mapping>
    <total_verifications>N</total_verifications>
    <mapped_to_criteria>M</mapped_to_criteria>
    <unmapped>[list of unmapped verifications]</unmapped>
  </verification_to_criteria_mapping>

  <orphans>
    <orphan_code_changes>[list]</orphan_code_changes>
    <orphan_criteria>[list]</orphan_criteria>
  </orphans>

  <quality_issues>
    <vague_locations>[list of non-specific code locations]</vague_locations>
    <untestable_criteria>[list of non-testable criteria]</untestable_criteria>
  </quality_issues>

  <coherence_verdict>COHERENT | PARTIAL | INCOHERENT</coherence_verdict>

  <gaps_to_fix>
    1. [specific gap to address]
    2. [specific gap to address]
  </gaps_to_fix>
</adr_plan_coherence>
```
</output>
</extended_thinking>

### Phase 3: Verify Plan Feasibility (Optional)

If coherence passes, optionally verify the Plan is feasible:

<extended_thinking>
<goal>Verify code locations in Plan actually exist or can be created</goal>

<feasibility_checks>
For each Code Change:
1. Does the target file/path exist (for modify/delete)?
2. Is the location reachable (not blocked by permissions, dependencies)?
3. Does the dependency order make sense?

For each Acceptance Criterion:
1. Can the test/command actually be run?
2. Are test dependencies available?
</feasibility_checks>

<output>
```xml
<plan_feasibility>
  <code_locations>
    <location path="src/handlers/auth.ts" action="modify">
      <exists>yes|no</exists>
      <feasible>yes|no</feasible>
      <issue_if_any>[description]</issue_if_any>
    </location>
  </code_locations>

  <dependency_order>
    <valid>yes|no</valid>
    <issue_if_any>[circular dependency, missing prerequisite, etc.]</issue_if_any>
  </dependency_order>

  <test_feasibility>
    <criterion test="npm test auth">[runnable: yes|no]</criterion>
  </test_feasibility>

  <overall_feasibility>FEASIBLE | NEEDS_ADJUSTMENT | BLOCKED</overall_feasibility>
</plan_feasibility>
```
</output>
</extended_thinking>

### ADR-Plan Audit Report

```markdown
# ADR-Plan Coherence Audit

**ADR:** adr-YYYYMMDD-slug
**Date:** YYYY-MM-DD

## Coherence Summary

| Mapping | Count | Mapped | Status |
|---------|-------|--------|--------|
| Layer Changes → Code Changes | N | M | ✓/✗ |
| Verifications → Acceptance Criteria | N | M | ✓/✗ |
| Orphans | - | N | ✓/✗ |

**Coherence Verdict:** COHERENT / PARTIAL / INCOHERENT

## Gaps Found

| Type | Item | Issue |
|------|------|-------|
| Unmapped Layer Change | c3-N: [change] | No Code Change entry |
| Unmapped Verification | [item] | No Acceptance Criterion |
| Vague Location | Code Change #3 | "auth code" not specific |
| Untestable Criterion | Criterion #2 | "should work" not testable |

## Quality Issues

- [List of vague locations that need specificity]
- [List of criteria that need testable form]

## Recommendation

[ ] **PASS** - Ready for handoff
[ ] **FIX REQUIRED** - Address gaps before handoff
[ ] **REJECT** - Major coherence issues, revise ADR+Plan

## Next Steps

1. [Specific action to fix gap 1]
2. [Specific action to fix gap 2]
```

---

## Mode 4: Container Audit

Focused audit of a single container and its components.

### Process

1. **Load container doc** and its component docs
2. **Explore container code** (just that path)
3. **Compare**:
   - Technology stack matches?
   - All documented components exist in code?
   - All code modules documented?
   - Patterns followed?
4. **Report** container-specific findings

<container_audit_output>
```xml
<container_audit container="c3-1-api">
  <technology_match>yes|no - [details]</technology_match>

  <component_coverage>
    <documented>5</documented>
    <found_in_code>4</found_in_code>
    <missing_in_code>[c3-105]</missing_in_code>
    <undocumented_code>[/src/middleware/rateLimit]</undocumented_code>
  </component_coverage>

  <pattern_compliance>
    <pattern name="error_handling">COMPLIANT|VIOLATION - [details]</pattern>
    <pattern name="data_access">COMPLIANT|VIOLATION - [details]</pattern>
  </pattern_compliance>

  <findings>
    <finding severity="medium">
      [Component c3-105 documented but code removed]
    </finding>
  </findings>
</container_audit>
```
</container_audit_output>

---

## Mode 4: Quick Audit

Fast inventory-only check.

```bash
# Just verify counts match
echo "=== Quick Audit ==="
echo "Containers in Context:"
grep -c "| c3-[0-9]" .c3/README.md

echo "Container folders:"
ls -d .c3/c3-[0-9]-*/ 2>/dev/null | wc -l

echo "ADRs by status:"
grep -h "^status:" .c3/adr/*.md 2>/dev/null | sort | uniq -c
```

---

## Audit Report Template

<report_template>
```markdown
# C3 Audit Report

**Date:** YYYY-MM-DD
**Mode:** [Full | ADR | Container | Quick]
**Target:** [all | adr-YYYYMMDD-slug | c3-N-slug]

## Executive Summary

[1-2 sentence summary]

## Statistics

| Metric | Documented | In Code | Match |
|--------|------------|---------|-------|
| Containers | N | N | ✓/✗ |
| Components | N | N | ✓/✗ |

## ADR Status

| ADR | Status | Docs Complete | Code Verified |
|-----|--------|---------------|---------------|
| adr-YYYYMMDD-slug | accepted | ✓/✗ | ✓/✗ |

## Findings by Severity

### Critical
| ID | Type | Issue | Action |
|----|------|-------|--------|

### High
| ID | Type | Issue | Action |
|----|------|-------|--------|

### Medium
| ID | Type | Issue | Action |
|----|------|-------|--------|

## Verification Checklist Results

[For ADR audits - show verification item results]

| Check | Result | Evidence |
|-------|--------|----------|
| [item] | PASS/FAIL | [details] |

## Recommended Actions

1. [Priority action 1]
2. [Priority action 2]
```
</report_template>

---

## Handoff

<chain_prompt id="handoff">
<instruction>Execute handoff based on settings or user choice</instruction>

<action>
```bash
cat .c3/settings.yaml 2>/dev/null | grep -A 5 "^audit:"
```
</action>

<handoff_options>
| Setting | Action |
|---------|--------|
| `audit: manual` | Present report only |
| `audit: tasks` | Create tracking tasks via vibe_kanban |
| `audit: fix` | Dispatch agents to fix (safe issues only) |
| Not set | Ask user preference |
</handoff_options>

<safe_to_auto_fix>
- Update technology version in docs
- Add missing component stub doc
- Update stale descriptions

NOT safe to auto-fix (require user decision):
- Remove phantom docs (might need to implement code instead)
- Structural changes
- Protocol changes
</safe_to_auto_fix>

</chain_prompt>

---

## Checklist

<verification_checklist>
**For Methodology Audit:**
- [ ] Layer defaults loaded (Context/Container/Component rules)
- [ ] Structure checked (IDs, paths, frontmatter)
- [ ] Layer content audited (include/exclude rules)
- [ ] Diagrams audited (appropriate types per layer)
- [ ] Contract chain verified (no orphans, no phantoms)
- [ ] Abstraction leaks identified (up and down)
- [ ] Report generated with fixes

**For Full Audit:**
- [ ] All C3 docs loaded
- [ ] Codebase explored
- [ ] Inventory compared
- [ ] Drift categorized by severity
- [ ] ADR statuses checked
- [ ] Report generated

**For ADR Audit:**
- [ ] ADR loaded and parsed
- [ ] Each "Changes Across Layers" item verified in docs
- [ ] Each verification checklist item checked against code
- [ ] Implementation recommendation made
- [ ] Gaps documented with next steps
- [ ] If all checks pass: transition status to `implemented`
- [ ] Rebuild TOC after status transition

**For ADR-Plan Coherence Audit:**
- [ ] ADR + Implementation Plan sections extracted
- [ ] Every "Changes Across Layers" mapped to Code Change
- [ ] Every Verification item mapped to Acceptance Criterion
- [ ] No orphan Code Changes (without Layer Change reference)
- [ ] No orphan Acceptance Criteria (without Verification reference)
- [ ] Code locations are specific (file:function, not vague)
- [ ] Acceptance Criteria are testable (command/test, not "should work")
- [ ] Coherence verdict determined
- [ ] Gaps documented with fix actions

**For Container Audit:**
- [ ] Container doc loaded
- [ ] Container code explored
- [ ] Technology match verified
- [ ] Component coverage checked
- [ ] Pattern compliance verified
</verification_checklist>

---

## Related

- [adr-template.md](../../references/adr-template.md) - ADR structure
- [v3-structure.md](../../references/v3-structure.md) - Expected doc structure
- [hierarchy-model.md](../../references/hierarchy-model.md) - C3 layer inheritance
