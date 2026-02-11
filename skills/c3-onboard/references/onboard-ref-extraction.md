# Onboard: Ref Extraction Guidance

## The Separation Test

When documenting components, proactively identify content that belongs in refs.

> "Would this content change if we swapped the underlying technology?"
> - **Yes** → Integration/usage pattern → extract to ref
> - **No** → Business/domain logic → keep in component

## Signals to Extract

| Signal | Action |
|--------|--------|
| "We use [technology] with..." | Extract to ref-[technology] |
| "Our convention is..." | Extract to existing or new ref |
| "Always handle errors by..." | Extract to ref-error-handling |
| "The retry policy is..." | Extract to ref-retry |
| "Logging should include..." | Extract to ref-logging |
| "Auth tokens must..." | Extract to ref-auth |

## Extraction Flow

```
Discover Pattern in Code → Check if Ref Exists → Create or Cite
                                    │
                    ┌───────────────┴───────────────┐
                    │                               │
               Ref Exists                    No Ref Exists
                    │                               │
            Add to Cited By                 Create New Ref
            Update Component                (if pattern repeats)
```

## Example

**During component documentation:**

> "The auth service validates JWTs using RS256 and checks token expiry..."

**Ask:** Would JWT validation change if we switched from Firebase to Cognito?

- **Yes** → Extract to `ref-jwt-validation`
- **No** → Keep in component (business rule)

## Common Extractions

| Pattern Category | Typical Ref Name | Examples |
|-----------------|------------------|----------|
| Error handling | ref-error-handling | Error classes, error codes, error responses |
| Authentication | ref-auth | Token format, validation rules, session handling |
| Database access | ref-database | Connection pooling, query patterns, transactions |
| API conventions | ref-api | Request/response format, versioning, pagination |
| Logging | ref-logging | Log levels, structured logging, correlation IDs |
| Retry/resilience | ref-retry | Backoff strategies, circuit breakers |
