---
id: c3-${N}
c3-version: 3
title: ${CONTAINER_NAME}
type: container
category: database
parent: c3-0
summary: ${SUMMARY}
---
<!-- USE: Databases, data stores, persistence layers -->
<!-- ADAPT: This is a starting point. Discover and document what actually exists. -->

# ${CONTAINER_NAME}

<!-- database purpose and technology: one sentence -->

## Complexity Assessment

<!-- REQUIRED: See skill-harness.md for levels and rules -->

**Level:** <!-- trivial | simple | moderate | complex | critical -->

**Signals observed:**
<!-- Scan for: schema complexity, access patterns, constraints, data lifecycle, multi-tenancy -->

## Overview

```mermaid
erDiagram
    %% Entity relationships
    %% Discover actual schema, don't assume
```

## Components

> For databases, "components" are logical groupings (domains, bounded contexts)

### Foundation
> Core entities that others reference

| ID | Name | Status | Responsibility |
|----|------|--------|----------------|

### Feature
> Domain-specific data models

| ID | Name | Status | Responsibility |
|----|------|--------|----------------|

## Linkages

```mermaid
graph LR
    %% Which services connect to this database
    %% Edge labels: "read/write purpose"
```

## Discovered Aspects

<!--
SKIP THIS SECTION if complexity is trivial or simple.

For moderate+, discover aspects through analysis:
- What schema patterns exist?
- What access patterns are used?
- What constraints are enforced?
- What lifecycle policies apply?

Document only what exists and matters. Reference specific tables/columns.
-->

## Testing (if warranted)

<!-- SKIP IF: trivial/simple complexity, standard ORM usage -->
