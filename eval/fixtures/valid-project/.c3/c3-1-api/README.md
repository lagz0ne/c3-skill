---
id: c3-1
type: container
parent: c3-0
title: API Service
---

# API Service

REST API service handling all user requests.

## Complexity Assessment

**Level:** simple
**Why:** Single purpose API with basic CRUD operations, no complex state management.

## Technology Stack

- Runtime: Node.js
- Framework: Express
- Database: PostgreSQL via pg driver

## Components

```mermaid
graph TD
    R[c3-101 Router] --> H[c3-102 Handlers]
    H --> M[c3-103 Models]
    M --> DB[(Database)]
```

## Component Index

| ID | Name | Category | Purpose |
|----|------|----------|---------|
| c3-101 | Router | foundation | Route definitions and middleware |
| c3-102 | Handlers | feature | Request handlers for each endpoint |
| c3-103 | Models | foundation | Database models and queries |
