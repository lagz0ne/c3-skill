---
id: c3-0
c3-version: 3
title: ${PROJECT}
summary: ${SUMMARY}
---

# ${PROJECT}

<!-- AI: 2-3 sentences on what this system does and why it exists -->

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

| ID | Actor | Type | Interacts With | Purpose |
|----|-------|------|----------------|---------|
<!-- AI: Who/what triggers this system -->
<!-- Type: user | system | scheduled | external-service -->

## Containers

| ID | Name | Type | Status | Purpose |
|----|------|------|--------|---------|
| c3-1 | ${C1_NAME} | service | | <!-- AI: purpose --> |
<!-- AI: Discover containers, assign IDs c3-1, c3-2, etc. -->
<!-- Type: service | app | library | external -->

## External Systems

| ID | Name | Type | Purpose |
|----|------|------|---------|
<!-- AI: Databases, APIs, third-party services -->
<!-- Type: database | cache | queue | api | storage -->

## Linkages

| From | To | Protocol | Reasoning |
|------|-----|----------|-----------|
<!-- AI: WHY these connect, not just THAT they do -->

## E2E Testing

> Tests container ↔ container linkages. Critical user flows that cross boundaries.

| Flow | Containers Involved | Verifies |
|------|---------------------|----------|
<!-- AI: Key user journeys that E2E tests cover -->
