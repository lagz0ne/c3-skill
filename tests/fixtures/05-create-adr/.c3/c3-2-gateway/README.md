---
id: c3-2
c3-version: 3
title: API Gateway
type: container
parent: c3-0
summary: API gateway for routing, auth, rate limiting
---

# API Gateway

## Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| Runtime | Node.js | JavaScript runtime |
| Gateway | Express Gateway | API gateway |
| Auth | JWT | Token validation |

## Components

| ID | Name | Responsibility |
|----|------|----------------|
| c3-201 | Router | Request routing to backends |
| c3-202 | Auth Middleware | JWT validation |
| c3-203 | Rate Limiter | Request throttling |

## Notes

- Currently forwards all requests to c3-1
- Partner API versioning handled here
