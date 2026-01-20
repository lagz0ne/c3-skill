---
id: c3-${N}
c3-version: 3
title: ${CONTAINER_NAME}
type: container
parent: c3-0
goal: ${GOAL}
summary: ${SUMMARY}
---

# ${CONTAINER_NAME}

## Goal

${GOAL}

## Context Connection

<!-- How does this container advance the parent context's Goal? -->
<!-- What would break in the system without this container? -->

<!--
WHY DOCUMENT:
- Enforce consistency (current and future work)
- Enforce quality (current and future work)
- Support auditing (verifiable, cross-referenceable)
- Be maintainable (worth the upkeep cost)

ANTI-GOALS:
- Over-documenting → stale quickly, maintenance burden
- Text walls → hard to review, hard to maintain
- Isolated content → can't verify from multiple angles

PRINCIPLES:
- Diagrams over text. Always.
- Fewer meaningful sections > many shallow sections
- Add sections that elaborate the Goal - remove those that don't
- Cross-content integrity: same fact from different angles aids auditing

GUARDRAILS:
- Must have: Goal + Components table
- Prefer: 3-5 focused sections
- Each section must serve the Goal - if not, delete
- If a section grows large, consider: diagram? split? ref-*?

REF HYGIENE (container level = cross-component concerns):
- Cite refs that govern how components in this container interact
  (communication patterns, error propagation, shared data flow)
- Component-specific ref usage belongs in component docs, not here
- If a pattern only affects one component, document it there instead

Common sections (create whatever serves your Goal):
- Overview (diagram), Components, Complexity Assessment, Fulfillment, Linkages

Delete this comment block after drafting.
-->

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
<!-- Category: foundation | feature -->
<!-- Goal Contribution: How this component advances the container Goal -->

## Layer Constraints

This container operates within these boundaries:

**MUST:**
- Coordinate components within its boundary
- Define how context linkages are fulfilled internally
- Own its technology stack decisions

**MUST NOT:**
- Define system-wide policies (context responsibility)
- Implement business logic directly (component responsibility)
- Bypass refs for cross-cutting concerns
- Orchestrate other containers (context responsibility)

## Criteria

Before finalizing, verify:
- [ ] Goal is specific and bounded
- [ ] Context Connection explains contribution to parent
- [ ] Every component has Goal Contribution filled
- [ ] Refs cited are container-wide (not component-specific)
- [ ] Layer Constraints are respected (no boundary violations)
