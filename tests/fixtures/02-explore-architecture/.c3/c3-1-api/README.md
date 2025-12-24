---
id: c3-1
c3-version: 3
title: API Backend
type: container
parent: c3-0
summary: REST API for all client operations
---

# API Backend

Handles all API requests.

## Technology Stack

| Tech | Purpose |
|------|---------|
| Express | Web framework |
| PostgreSQL | Database |

## Components

| ID | Name | Responsibility |
|----|------|----------------|
| c3-101 | Auth Middleware | JWT validation |
| c3-102 | User Routes | CRUD for users |
| c3-103 | Product Routes | Product catalog |
| OrderController | Order processing | Handles checkout |

## Notes

TODO: Add caching layer
TODO: Document payment integration
