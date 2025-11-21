---
id: COM-004-api-client
title: API Client (Resource)
summary: >
  Frontend HTTP client for TaskFlow. Wraps fetch with auth header injection, JSON parsing,
  retries, and error normalization for UI consumers.
nature: Resource
---

# [COM-004-api-client] API Client (Resource)

## Overview {#com-004-overview}
- Implements the CTX REST protocol to the backend from [CON-002-frontend#con-002-api-calls](../../containers/CON-002-frontend.md#con-002-api-calls).
- Centralized HTTP client for all backend API calls. Handles auth token injection, JSON parsing, retries, and consistent error formatting.

## Stack {#com-004-stack}
- Library: Native `fetch` API with small wrapper
- Language: TypeScript 5.x

## Configuration {#com-004-config}
| Env Var | Dev | Prod | Why |
|---------|-----|------|-----|
| VITE_API_URL | http://localhost:3000/api | /api | Backend base URL |
| VITE_API_TIMEOUT_MS | 10000 | 30000 | Request timeout ms |
| VITE_RETRY_COUNT | 1 | 3 | Max retries for retriable errors |

### Config Loading {#com-004-config-loading}

```typescript
const apiConfig = {
  baseUrl: import.meta.env.VITE_API_URL || '/api',
  timeoutMs: Number(import.meta.env.VITE_API_TIMEOUT_MS) || 30000,
  retryCount: Number(import.meta.env.VITE_RETRY_COUNT) || 3,
};
```

## Interfaces & Types {#com-004-interfaces}

```typescript
interface ApiClient {
  get<T>(path: string): Promise<T>;
  post<T>(path: string, body: unknown): Promise<T>;
  put<T>(path: string, body: unknown): Promise<T>;
  delete(path: string): Promise<void>;
}

interface ApiError {
  code: string;
  message: string;
  correlationId?: string;
}
```

## Behavior {#com-004-behavior}

```mermaid
sequenceDiagram
    participant Component
    participant ApiClient
    participant Backend

    Component->>ApiClient: get('/tasks')
    ApiClient->>ApiClient: Attach auth header
    ApiClient->>Backend: GET /api/v1/tasks
    Backend-->>ApiClient: 200 + JSON (or 401/5xx)
    ApiClient->>ApiClient: Parse/normalize error
    ApiClient-->>Component: Data or ApiError

    Note over ApiClient,Backend: On 401 → refresh then retry once
```

### Behavior Notes
- Injects `Authorization: Bearer <token>` when available.
- Retries idempotent requests on network/5xx up to `VITE_RETRY_COUNT`.
- Maps HTTP errors into UI-friendly objects with correlation IDs when provided by backend.

## Error Handling {#com-004-errors}
| Error | Retriable | Action/Code |
|-------|-----------|-------------|
| Network error | Yes | Retry with backoff up to `VITE_RETRY_COUNT` |
| 401 unauthorized | Yes | Trigger refresh flow, retry once |
| 4xx validation | No | Surface normalized error |
| 5xx | Yes | Retry up to retry count, then surface |

## Usage {#com-004-usage}

```typescript
import { apiClient } from './api';

const tasks = await apiClient.get<Task[]>('/tasks');
await apiClient.post('/tasks', { title: 'Demo' });
```

## Health Checks {#com-004-health}
| Check | Probe | Expectation |
|-------|-------|-------------|
| Backend reachable | `GET /health` | 200 OK |
| Config valid | `baseUrl` non-empty | Non-empty string |

## Metrics & Observability {#com-004-metrics}
| Metric | Type | Description |
|--------|------|-------------|
| `api_requests_total` | Counter | Requests by method |
| `api_errors_total` | Counter | Errors by status code |
| `api_latency_ms` | Histogram | Request duration |
| `api_retries_total` | Counter | Retry attempts |

## Dependencies {#com-004-deps}
- Consumes backend REST API at [CON-001-backend](../../containers/CON-001-backend.md)
- Used by frontend data-fetching hooks and pages
