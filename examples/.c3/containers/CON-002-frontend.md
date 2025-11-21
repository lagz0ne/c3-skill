---
id: CON-002-frontend
title: Frontend Container (Code)
summary: >
  Web user interface for TaskFlow. React-based SPA that communicates
  with the backend API.
---

# [CON-002-frontend] Frontend Container (Code)

## Overview {#con-002-overview}

The Frontend container provides the web user interface for TaskFlow. It's a React-based single-page application that communicates with the backend API.

## Technology Stack {#con-002-stack}

- Runtime: Browser
- Framework: React 18
- Language: TypeScript 5.x
- Build: Vite

## Protocol Implementations {#con-002-protocols}

| Protocol (from CTX) | Implemented In |
|---------------------|----------------|
| REST/HTTPS to Backend | [COM-004-api-client#com-004-behavior](../components/frontend/COM-004-api-client.md#com-004-behavior) |

## Component Relationships {#con-002-relationships}

```mermaid
flowchart LR
    App[App] --> Router[Router]
    Router --> Pages[Pages]
    Pages --> Components[UI Components]
    Pages --> Hooks[Custom Hooks]
    Hooks --> APIClient[API Client]
```

## Container Cross-Cutting {#con-002-cross-cutting}

### Auth Handling {#con-002-auth-handling}

- JWT token storage in httpOnly cookies
- Automatic token refresh on 401
- Auth context provider for components

### Logging {#con-002-logging}

- Console logging in development
- Error boundary for uncaught errors
- Correlation ID forwarding to backend

### Error Handling {#con-002-error-handling}

- Toast notifications for user errors
- Error boundaries for component failures
- Retry UI for network errors

## API Calls {#con-002-api-calls}

Uses fetch API with wrapper for:
- Automatic auth header injection
- JSON parsing
- Error normalization

## Components {#con-002-components}

| Component | Nature | Responsibility |
|-----------|--------|----------------|
| [COM-004-api-client](../components/frontend/COM-004-api-client.md) | Resource | Backend communication |

## Related {#con-002-related}

- [CTX-001: System Overview](../CTX-001-system-overview.md)
- [CON-001: Backend](./CON-001-backend.md) - API provider
