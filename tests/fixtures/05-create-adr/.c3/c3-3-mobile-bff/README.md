---
id: c3-3
c3-version: 3
title: Mobile BFF
type: container
parent: c3-0
summary: Backend-for-Frontend optimized for mobile clients
---

# Mobile BFF

## Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| Runtime | Node.js | JavaScript runtime |
| Framework | Express | HTTP server |
| Protocol | REST | Currently REST, needs improvement |

## Components

| ID | Name | Responsibility |
|----|------|----------------|
| c3-301 | Aggregator | Combines multiple c3-1 calls |
| c3-302 | Transformer | Shapes data for mobile |
| c3-303 | Cache | Response caching |

## Pain Points

- Mobile team makes 5-10 REST calls per screen
- Over-fetching: REST returns more data than needed
- Under-fetching: Need multiple calls for related data
- Aggregator (c3-301) is complex workaround

## Notes

- Mobile team has explicitly requested GraphQL
- Current aggregator pattern is becoming unmaintainable
