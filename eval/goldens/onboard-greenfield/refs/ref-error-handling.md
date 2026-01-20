---
id: ref-error-handling
c3-version: 3
title: Error Handling Patterns
---

# Error Handling Patterns

## Goal

Establish consistent error handling patterns including error codes, recovery strategies, user messaging, and logging conventions across all system layers.

## Error Categories

### Classification

| Category | Code Range | Recovery |
|----------|------------|----------|
| Validation | E1xxx | User correction |
| Authentication | E2xxx | Re-authenticate |
| Authorization | E3xxx | Request access |
| Resource | E4xxx | Check existence |
| External | E5xxx | Retry with backoff |
| Internal | E9xxx | Report and retry |

### Error Codes

| Code | Name | Description |
|------|------|-------------|
| E1001 | VALIDATION_REQUIRED | Required field missing |
| E1002 | VALIDATION_FORMAT | Invalid format |
| E1003 | VALIDATION_RANGE | Value out of range |
| E2001 | AUTH_TOKEN_MISSING | No auth token provided |
| E2002 | AUTH_TOKEN_EXPIRED | Token has expired |
| E2003 | AUTH_TOKEN_INVALID | Token malformed/invalid |
| E3001 | FORBIDDEN_RESOURCE | No access to resource |
| E3002 | FORBIDDEN_ACTION | Action not permitted |
| E4001 | NOT_FOUND | Resource not found |
| E4002 | ALREADY_EXISTS | Duplicate resource |
| E4003 | CONFLICT | Concurrent modification |
| E5001 | EXTERNAL_TIMEOUT | External service timeout |
| E5002 | EXTERNAL_ERROR | External service error |
| E9001 | INTERNAL_ERROR | Unexpected error |

## Error Structure

### Backend Format

```typescript
interface AppError {
  code: string;        // E1001, E2001, etc.
  message: string;     // Human-readable
  details?: object;    // Additional context
  stack?: string;      // Dev only
  requestId: string;   // Correlation
}
```

### Client Format

```typescript
interface ClientError {
  code: string;
  message: string;
  field?: string;      // For validation
  recoverable: boolean;
  action?: string;     // Suggested action
}
```

## Recovery Strategies

### By Category

| Category | Strategy |
|----------|----------|
| Validation | Show inline errors, focus first |
| Auth expired | Silent refresh, retry request |
| Auth invalid | Redirect to login |
| Not found | Show empty state, suggest action |
| Rate limited | Queue, retry after delay |
| External | Retry with exponential backoff |
| Internal | Show generic error, report |

### Retry Logic

| Attempt | Delay | Max |
|---------|-------|-----|
| 1 | 1s | 3 attempts |
| 2 | 2s | for recoverable |
| 3 | 4s | errors |

### Circuit Breaker

| State | Behavior |
|-------|----------|
| Closed | Normal operation |
| Open | Fail fast, skip calls |
| Half-open | Test with single request |

## User Messaging

### Principles

| Principle | Implementation |
|-----------|----------------|
| Be specific | "Title is required" not "Error" |
| Be actionable | "Click to retry" not "Try again later" |
| Be human | Friendly tone, no jargon |
| Be honest | Admit failures, don't blame user |

### Message Templates

| Scenario | Message |
|----------|---------|
| Validation | "{field} {problem}" |
| Network | "Connection lost. We'll retry automatically." |
| Auth | "Your session expired. Please sign in again." |
| Not found | "We couldn't find that. It may have been deleted." |
| Server | "Something went wrong on our end. We're looking into it." |

### Toast Duration

| Severity | Duration |
|----------|----------|
| Success | 3s |
| Info | 4s |
| Warning | 5s |
| Error | Manual dismiss |

## Logging

### Log Levels

| Level | Usage |
|-------|-------|
| error | Unhandled exceptions, critical failures |
| warn | Recoverable issues, deprecations |
| info | Request/response, important events |
| debug | Detailed troubleshooting |

### Error Context

| Field | Required |
|-------|----------|
| requestId | Yes |
| userId | If authenticated |
| errorCode | Yes |
| errorMessage | Yes |
| stack | Errors only |
| duration | Yes |

## Frontend Error Boundaries

### Boundary Placement

| Boundary | Catches |
|----------|---------|
| App root | Entire app crash |
| Route | Page-level errors |
| Feature | Component-level |
| Widget | Isolated failures |

### Fallback UI

| Level | Fallback |
|-------|----------|
| App | Full error page with reload |
| Route | Error message with back link |
| Feature | Retry button |
| Widget | Inline error message |

## Cited By

- c3-201 (HTTP Router)
