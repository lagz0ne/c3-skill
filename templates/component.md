---
id: c3-${N}${NN}
c3-version: 3
title: ${COMPONENT_NAME}
type: component
parent: c3-${N}
summary: ${SUMMARY}
---

# ${COMPONENT_NAME}

<!-- AI: 2-3 sentences on what this component does -->

## Overview

```mermaid
graph LR
    %% Inputs
    IN1([Input]) --> c3-${N}${NN}

    %% This component
    c3-${N}${NN}[${COMPONENT_NAME}]

    %% Outputs
    c3-${N}${NN} --> OUT1([Output])
```

## Interface

| Direction | What | Format | From/To |
|-----------|------|--------|---------|
<!-- AI: What goes in, what comes out -->

## Hand-offs

| To | What | Contract | Mechanism |
|----|------|----------|-----------|
<!-- AI: Who receives output, what contract -->

## Implementation

### Technology

| Aspect | Choice | Rationale |
|--------|--------|-----------|
<!-- AI: What libraries/frameworks, why chosen -->

### Conventions

<!-- AI: Rules consumers must follow -->

### Edge Cases

| Scenario | Behavior |
|----------|----------|
<!-- AI: What happens when things go wrong -->
