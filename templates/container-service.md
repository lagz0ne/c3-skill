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

<!-- service purpose: one sentence -->

## Complexity Assessment

<!-- REQUIRED: See skill-harness.md for levels and rules -->

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

### Foundation
> Primitives others build on. High impact when changed.

| ID | Name | Status | Responsibility |
|----|------|--------|----------------|

### Auxiliary
> Conventions for using external tools. "How we use X here."

| ID | Name | Status | Responsibility |
|----|------|--------|----------------|

### Feature
> Domain-specific. Uses Foundation + Auxiliary.

| ID | Name | Status | Responsibility |
|----|------|--------|----------------|

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
