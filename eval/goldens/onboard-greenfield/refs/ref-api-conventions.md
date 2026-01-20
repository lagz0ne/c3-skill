---
id: ref-api-conventions
c3-version: 3
title: API Conventions
---

# API Conventions

## Goal

Define REST API conventions for consistent endpoint design, request/response formats, versioning, and error handling across all backend services.

## URL Structure

### Base Pattern

```
/api/v1/{resource}[/{id}][/{sub-resource}]
```

### Naming Conventions

| Convention | Example |
|------------|---------|
| Plural resources | /concepts, /canvases |
| Kebab-case | /canvas-templates |
| No trailing slash | /concepts (not /concepts/) |

### Resource Examples

| Resource | URL |
|----------|-----|
| List concepts | GET /api/v1/concepts |
| Get concept | GET /api/v1/concepts/:id |
| Create concept | POST /api/v1/concepts |
| Update concept | PATCH /api/v1/concepts/:id |
| Delete concept | DELETE /api/v1/concepts/:id |
| Concept links | GET /api/v1/concepts/:id/links |

## Request Format

### Headers

| Header | Value | Required |
|--------|-------|----------|
| Content-Type | application/json | Yes (POST/PATCH) |
| Authorization | Bearer {token} | Yes (authenticated) |
| X-Request-ID | UUID | Recommended |
| X-Canvas-ID | UUID | Context-dependent |

### Query Parameters

| Parameter | Type | Usage |
|-----------|------|-------|
| page | number | Pagination offset |
| limit | number | Items per page (max 100) |
| sort | string | Field to sort by |
| order | asc/desc | Sort direction |
| filter[field] | string | Field filtering |
| include | string | Related resources |

### Body Format

```json
{
  "data": {
    "type": "concept",
    "attributes": {
      "title": "Example",
      "position": { "x": 100, "y": 200 }
    }
  }
}
```

## Response Format

### Success Response

```json
{
  "data": {
    "id": "uuid",
    "type": "concept",
    "attributes": { ... },
    "relationships": { ... }
  },
  "meta": {
    "requestId": "uuid",
    "timestamp": "ISO8601"
  }
}
```

### List Response

```json
{
  "data": [...],
  "meta": {
    "total": 100,
    "page": 1,
    "limit": 20,
    "hasMore": true
  }
}
```

### Error Response

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable message",
    "details": [
      { "field": "title", "message": "Required" }
    ]
  },
  "meta": {
    "requestId": "uuid"
  }
}
```

## HTTP Status Codes

### Success

| Code | Meaning | When |
|------|---------|------|
| 200 | OK | Successful GET, PATCH |
| 201 | Created | Successful POST |
| 204 | No Content | Successful DELETE |

### Client Errors

| Code | Meaning | When |
|------|---------|------|
| 400 | Bad Request | Invalid request body |
| 401 | Unauthorized | Missing/invalid auth |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 409 | Conflict | Resource conflict |
| 422 | Unprocessable | Validation failed |
| 429 | Too Many Requests | Rate limited |

### Server Errors

| Code | Meaning | When |
|------|---------|------|
| 500 | Internal Error | Unexpected server error |
| 502 | Bad Gateway | Upstream failure |
| 503 | Service Unavailable | Maintenance/overload |

## Versioning

### Strategy

| Approach | Implementation |
|----------|----------------|
| URL versioning | /api/v1/, /api/v2/ |
| Deprecation header | X-API-Deprecated: true |
| Sunset header | Sunset: date |

### Compatibility

| Change Type | Version Impact |
|-------------|----------------|
| New field | None (additive) |
| Optional param | None |
| Remove field | Major version |
| Change behavior | Major version |

## Rate Limiting

### Headers

| Header | Description |
|--------|-------------|
| X-RateLimit-Limit | Requests per window |
| X-RateLimit-Remaining | Remaining requests |
| X-RateLimit-Reset | Window reset time |

### Limits

| Tier | Limit |
|------|-------|
| Free | 100/minute |
| Pro | 1000/minute |
| Team | 5000/minute |

## Cited By

- c3-201 (HTTP Router)
