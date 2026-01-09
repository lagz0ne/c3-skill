---
id: c3-${N}
c3-version: 3
title: ${CONTAINER_NAME}
type: container
category: service
parent: c3-0
summary: ${SUMMARY}
---
<!-- USE: Code services, APIs, backend applications -->
<!-- ADAPT: This is a starting point. Discover and document what actually exists. -->

# ${CONTAINER_NAME}

## Goal

{Why does this service exist? What bounded context does it own?}

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

## Complexity Assessment

<!-- OPTIONAL: See skill-harness.md for levels and rules -->

**Level:** <!-- trivial | simple | moderate | complex | critical -->

**Signals observed:**
<!-- Scan for: external calls, state mgmt, security, error handling, caching -->

## Overview

```mermaid
graph TD
    c3-${N}01[Component]
    c3-${N}02[Component]
    c3-${N}01 --> c3-${N}02
```

## Components

| ID | Name | Category | Status | Responsibility |
|----|------|----------|--------|----------------|
<!-- Category: foundation | feature -->

## Fulfillment

| Link (from c3-0) | Fulfilled By | Constraints |
|------------------|--------------|-------------|

## Linkages

```mermaid
graph TD
    %% Component → Component flows
    %% Edge labels: "reasoning"
```

## Discovered Aspects

<!--
SKIP THIS SECTION if complexity is trivial or simple.

For moderate+, discover aspects through code analysis:
- What patterns actually exist in the code?
- What would a new developer miss?
- What has non-trivial implementation worth documenting?

Document only what exists and matters. Reference code locations.
-->

## Testing (if warranted)

<!-- SKIP IF: trivial/simple complexity, no integration points -->
