---
id: c3-${N}
c3-version: 3
title: ${CONTAINER_NAME}
type: container
parent: c3-0
summary: ${SUMMARY}
---

# ${CONTAINER_NAME}

<!-- AI: 2-3 sentences on what this container does -->

## Overview

```mermaid
graph TD
    c3-${N}01[Component]
    c3-${N}02[Component]
    c3-${N}01 --> c3-${N}02
```

## Components

### Foundation
> Primitives others build on. High impact when changed. Reusable within this container.
> Examples: Layout, Button, EntryPoint, Router, AuthProvider

| ID | Name | Status | Responsibility |
|----|------|--------|----------------|
<!-- AI: What are the building blocks? -->

### Auxiliary
> Conventions for using external tools HERE. "This is how we use X in this project."
> Examples: "We use Tailwind like this", "prefer type over interface", API client patterns

| ID | Name | Status | Responsibility |
|----|------|--------|----------------|
<!-- AI: What patterns/conventions do Feature components follow? -->

### Feature
> Domain-specific. Uses Foundation + Auxiliary. Not reusable outside this context.
> Examples: ProductCard, CheckoutScreen, useCart, OrderHistory

| ID | Name | Status | Responsibility |
|----|------|--------|----------------|
<!-- AI: What does this container actually DO for users? -->

## Fulfillment

| Link (from c3-0) | Fulfilled By | Constraints |
|------------------|--------------|-------------|
<!-- AI: How does this container fulfill Context connections? -->

## Linkages

| From | To | Reasoning |
|------|-----|-----------|
<!-- AI: WHY do components connect? -->

## Testing

> Tests component ↔ component linkages within this container.

### Integration Tests

| Scenario | Components Involved | Verifies |
|----------|---------------------|----------|
<!-- AI: How components work together -->

### Mocking

| Dependency | How to Mock | When |
|------------|-------------|------|
<!-- AI: How to replace external dependencies in tests -->

### Fixtures

| Entity | Factory/Source | Notes |
|--------|----------------|-------|
<!-- AI: Test data for this container's domain -->
