---
id: c3-224
c3-version: 3
title: Auth Flows
type: component
category: feature
parent: c3-2
summary: Authentication flows for Google OAuth and test token auth
---

# Auth Flows

Provides authentication business logic for Google OAuth token exchange and profile fetching, plus test token authentication for development/testing.

## Uses

| Category | Component | For |
|----------|-----------|-----|
| Ref | ref-flow-patterns | Flow structure |
| Ref | ref-query-patterns | userQueries for user lookup |
| Foundation | c3-203 Database Layer | User record access |

## Behavior

| Trigger | Result |
|---------|--------|
| authenticateWithGoogle(code) | Exchanges code for tokens, fetches profile, returns user |
| authenticateWithTestToken(token, email) | Validates token, looks up user by email |
| Invalid OAuth code | Returns failure with reason |
| User not in database | Returns failure (no auto-create) |
| Valid auth | Sets session cookie, returns user data |

## Testing

| Scenario | Verifies |
|----------|----------|
| Valid OAuth code | Token exchange succeeds, profile fetched |
| Invalid OAuth code | Returns error, no cookie set |
| Valid test token | User returned from database |
| Invalid test token | Returns unauthorized |
| User not found | Returns not found, no crash |

## References

- `apps/start/src/server/flows/authenticate.ts` - Auth flows
- `apps/start/src/server/resources/auth.ts` - OAuth service
