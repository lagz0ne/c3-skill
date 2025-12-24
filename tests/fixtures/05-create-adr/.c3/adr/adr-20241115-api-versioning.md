---
id: adr-20241115-api-versioning
type: adr
status: implemented
title: API Versioning Strategy
affects: [c3-1, c3-2]
---

# ADR: API Versioning Strategy

## Context

Partner integrations require stable API contracts. Need versioning strategy.

## Decision

Use URL path versioning: `/api/v1/`, `/api/v2/`

## Rationale

- Clear and visible in URLs
- Easy to route in gateway
- Partners can specify exact version

## Consequences

### Positive
- Partners have stable contracts
- Can iterate on new versions without breaking existing

### Negative
- Must maintain multiple versions
- Documentation per version

## Changes Across Layers

| Layer | Change |
|-------|--------|
| c3-0 | Document versioning in interactions |
| c3-1 | Version-specific route modules |
| c3-2 | Gateway routing by version |

## Verification Checklist

- [x] Gateway routes by version prefix
- [x] API docs per version
- [x] Partner contracts specify version

## Audit Record

| Phase | Date | Status |
|-------|------|--------|
| Proposed | 2024-11-10 | Done |
| Accepted | 2024-11-12 | Done |
| Implemented | 2024-11-20 | Done |
