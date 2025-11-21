---
id: ADR-001-rest-api
title: REST API for Frontend-Backend Communication
status: accepted
date: 2024-01-15
---

# ADR-001: REST API for Frontend-Backend Communication

## Status

Accepted

## Context

We need to choose a communication protocol between the Frontend and Backend containers.

## Decision

Use REST over HTTPS with JSON payloads.

## Rationale

- **Simplicity**: REST is well-understood, easy to debug
- **Tooling**: Excellent browser devtools, Postman, curl support
- **Caching**: HTTP caching works out of the box
- **Compatibility**: Works with any client (web, mobile, CLI)

## Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| GraphQL | Flexible queries | Complexity, caching harder |
| gRPC | Performance | Browser support limited |
| WebSocket | Real-time | Overkill for CRUD |

## Consequences

- Need API versioning strategy (`/api/v1/`)
- Must design consistent error format
- Consider rate limiting from start
