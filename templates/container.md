---
id: c3-${N}
c3-version: 3
title: ${CONTAINER_NAME}
type: container
parent: c3-0
summary: ${SUMMARY}
---

# ${CONTAINER_NAME}

<!-- AI: 2-3 sentences on what this container does and its role in the system -->

## Overview

```mermaid
graph TD
    %% Foundation
    c3-${N}01[Entry Point]

    %% Auxiliary
    c3-${N}02[Shared Utility]

    %% Business
    c3-${N}03[Domain Service]

    %% Flows
    c3-${N}01 -->|delegates| c3-${N}03
    c3-${N}03 -->|uses| c3-${N}02
```

## Components

### Foundation

| ID | Name | Concern | Status | Responsibility |
|----|------|---------|--------|----------------|
<!-- AI: Entry, identity, integration concerns -->
<!-- Concern: entry | identity | integration -->

### Auxiliary

| ID | Name | Concern | Status | Responsibility |
|----|------|---------|--------|----------------|
<!-- AI: Framework usage, library wrappers, shared utilities -->
<!-- Concern: library-wrapper | framework | cross-cutting -->
<!-- Other components MUST use these for consistency -->

### Business

| ID | Name | Concern | Status | Responsibility |
|----|------|---------|--------|----------------|
<!-- AI: Domain services, business flows -->
<!-- Concern: domain -->

### Presentation

N/A
<!-- For frontend containers, uncomment and fill: -->
<!-- | ID | Name | Concern | Status | Responsibility | -->
<!-- Concern: styling | composition | state -->

## Fulfillment

<!-- How this container fulfills connections from Context (c3-0) -->

| Link (from c3-0) | Fulfilled By | Constraints |
|------------------|--------------|-------------|
<!-- AI: Map context-level links to components -->

## Linkages

| From | To | Contract | Reasoning |
|------|-----|----------|-----------|
<!-- AI: WHY components connect internally -->
