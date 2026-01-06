---
id: c3-0
c3-version: 3
title: ${PROJECT}
summary: ${SUMMARY}
---

# ${PROJECT}

<!-- system purpose -->

## Overview

```mermaid
graph LR
    %% Actors
    A1[User]

    %% Containers
    c3-1[${C1_NAME}]

    %% External
    E1[(Database)]

    %% Connections
    A1 -->|HTTP| c3-1
    c3-1 -->|SQL| E1
```

## Actors

| ID | Actor | Type | Purpose |
|----|-------|------|---------|
<!-- user | system | scheduled | external-service -->

## Containers

| ID | Name | Type | Status | Purpose |
|----|------|------|--------|---------|
| c3-1 | ${C1_NAME} | service | | <!-- purpose --> |
<!-- service | app | library | external -->

## External Systems

| ID | Name | Type | Purpose | Docs |
|----|------|------|---------|------|
<!-- database | cache | queue | api | auth | storage -->
<!-- Docs: "inline" for simple, "externals/E1-name/" for complex -->

## Linkages

```mermaid
graph LR
    %% Actor → Container → External flows
    %% Edge labels: "protocol: reasoning"
```

## E2E Testing (if warranted)

<!-- SKIP IF: single container, no critical user journeys, internal tool -->

<!--
trivial: skip section
simple: "E2E covers: login flow, checkout"
moderate: boundaries + key flows prose
complex: sequence diagram showing cross-container flow
-->
