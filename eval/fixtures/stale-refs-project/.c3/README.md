---
id: c3-0
c3-version: 3
title: StaleApp
summary: A test app with intentionally stale references
---

# StaleApp

Application with outdated documentation for testing degradation detection.

## Actors

- **User**: End user

## Containers

```mermaid
graph TD
    U[User] --> API[c3-1 API]
```

## Container Index

| ID | Name | Purpose |
|----|------|---------|
| c3-1 | API | API service |
