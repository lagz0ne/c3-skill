---
id: ADR-001-rest-api
title: Choose REST API for Frontend–Backend Communication
status: accepted
---

# [ADR-001] Choose REST API for Frontend–Backend Communication

## Context
TaskFlow frontend communicates with the backend. We need a protocol choice that is simple to implement, well supported by tooling, and fits our latency/throughput needs.

## Decision
Use REST over HTTPS with JSON payloads, versioned via URL prefix (`/api/v1`), authenticated via JWT bearer tokens.

## Consequences
- **Positive:** Broad tooling (browsers/fetch), easy debugging, caching/CDN friendly for GETs.
- **Negative:** Less efficient than gRPC for streaming/binary payloads; need pagination schemes for large lists.
- **Mitigations:** Add ETags/caching headers; consider gRPC/websockets for future real-time features.
