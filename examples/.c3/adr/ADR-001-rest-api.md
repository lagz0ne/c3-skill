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

TaskFlow's frontend communicates with the backend. We need a protocol that is simple to implement, well-supported by tooling, and fits our latency/throughput needs across web and mobile clients.

## Decision

Use REST over HTTPS with JSON payloads, versioned via URL prefix (`/api/v1`), authenticated via JWT bearer tokens.

## Rationale

- **Simplicity:** Broadly understood, easy to debug
- **Tooling:** Browser devtools, Postman, curl, and SDK support
- **Caching:** HTTP caching/CDN friendly for GETs
- **Compatibility:** Works with any client (web, mobile, CLI)

## Alternatives Considered

| Option | Pros | Cons |
|--------|------|------|
| GraphQL | Flexible queries | Complexity, caching harder |
| gRPC | Performance | Browser support limited |
| WebSocket | Real-time | Overkill for CRUD |

## Consequences

- Versioning required for breaking changes (`/api/v1/`)
- Standardized error format and correlation IDs needed
- Rate limiting from the start (100 requests/minute per user)
- Less efficient than gRPC for streaming/binary payloads
- Pagination strategies required for large lists

## Mitigations

- Add ETags/caching headers for GET responses
- Consider gRPC/websockets for future real-time features
- HTTP error contract covered by [C3-1-backend#c3-1-error-handling](../containers/C3-1-backend.md#c3-1-error-handling)
