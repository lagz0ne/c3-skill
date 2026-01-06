---
id: c3-${N}
c3-version: 3
title: ${CONTAINER_NAME}
type: container
parent: c3-0
summary: ${SUMMARY}
---

# ${CONTAINER_NAME}

<!-- container purpose -->

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

## Testing Strategy

**Integration scope:** <!-- what component interactions to test -->

**Mocking approach:** <!-- how to isolate dependencies -->

**Fixtures:** <!-- test data sources and factories -->
