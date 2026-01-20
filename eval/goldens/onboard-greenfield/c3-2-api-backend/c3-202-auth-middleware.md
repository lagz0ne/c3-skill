---
id: c3-202
c3-version: 3
title: Auth Middleware
type: component
parent: c3-2
category: foundation
summary: Token validation and session management for API requests
---

# Auth Middleware

## Goal

Secure API endpoints with JWT validation, OAuth flow handling, and tenant isolation enforcement.

## Contract

From c3-2 (API Backend): "Token validation, session management"

## Interface Diagram

```mermaid
flowchart LR
    subgraph IN["Receives"]
        I1[Authorization header]
        I2[OAuth callback]
        I3[Session cookie]
    end

    subgraph PROCESS["Owns"]
        P1[Token validation]
        P2[Session lookup]
        P3[Tenant resolution]
    end

    subgraph OUT["Provides"]
        O1[User context]
        O2[Tenant ID]
        O3[Auth errors]
    end

    IN --> PROCESS --> OUT
```

## Hand-offs

| Direction | What | To/From |
|-----------|------|---------|
| IN | Bearer token | c3-201 Router (from header) |
| IN | OAuth code | OAuth provider callback |
| OUT | User context object | All protected route handlers |
| OUT | Tenant ID | c3-203 Graph Client (for isolation) |
| OUT | 401/403 response | c3-201 Router |

## Auth Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant R as Router
    participant A as Auth Middleware
    participant O as OAuth Provider
    participant DB as Session Store

    C->>R: Request with token
    R->>A: Check auth
    alt Valid JWT
        A->>A: Decode, validate
        A->>DB: Lookup session
        A-->>R: User context
    else Expired
        A-->>R: 401 Expired
    else Invalid
        A-->>R: 401 Invalid
    end

    Note over C,O: OAuth Flow
    C->>R: /auth/google
    R->>O: Redirect
    O->>R: Callback with code
    R->>A: Exchange code
    A->>O: Token exchange
    A->>DB: Create session
    A-->>C: Set token cookie
```

## Conventions

| Rule | Why |
|------|-----|
| Short-lived JWTs (15min) | Limit exposure window |
| Refresh tokens in httpOnly cookie | XSS protection |
| Tenant ID in JWT claims | Avoid lookup on every request |
| Log auth failures without PII | Security audit |

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Missing token | 401 with www-authenticate header |
| Malformed token | 401, log attempt |
| Valid token, revoked session | 401, clear cookie |
| OAuth provider down | 503 with retry-after |
| Cross-tenant access attempt | 403, log security event |

## References

- Auth middleware: `src/api/middleware/auth.ts`
- OAuth handlers: `src/api/auth/oauth.ts`
- Session store: `src/services/session.ts`
- Cites: ref-auth-patterns, ref-security
