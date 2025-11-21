id: C3-105-error-handler
title: Error Handler (Cross-cutting)
summary: >
  Unified error formatting middleware. Converts exceptions to consistent JSON error responses.
nature: Cross-cutting
---

# [C3-105-error-handler] Error Handler (Cross-cutting)

## Overview {#c3-105-overview}

Express error middleware that catches all errors and formats them into consistent JSON responses with correlation IDs.

## Stack {#c3-105-stack}

- Library: None (Express middleware)
- Why: Native Express error handling pattern

## Configuration {#c3-105-config}

| Env Var | Dev | Prod | Why |
|---------|-----|------|-----|
| EXPOSE_STACK | `true` | `false` | Include stack trace in dev |
| NODE_ENV | `development` | `production` | Environment detection |

### Config Loading {#c3-105-config-loading}

```typescript
import { z } from 'zod';

const errorConfigSchema = z.object({
  exposeStack: z.coerce.boolean().default(false),
  nodeEnv: z.enum(['development', 'production', 'test']).default('production'),
});

export const errorConfig = errorConfigSchema.parse({
  exposeStack: process.env.EXPOSE_STACK,
  nodeEnv: process.env.NODE_ENV,
});
```

## Interfaces & Types {#c3-105-interfaces}

```typescript
interface ErrorResponse {
  error: {
    code: string;
    message: string;
    correlationId: string;
    stack?: string; // Dev only
  };
}

interface AppError extends Error {
  code: string;
  statusCode: number;
  isOperational: boolean;
}
```

## Behavior {#c3-105-behavior}

```mermaid
flowchart TD
    Error --> IsAppError{Is AppError?}
    IsAppError -->|Yes| ExtractCode[Extract code/status]
    IsAppError -->|No| DefaultCode[Use 500/internal_error]
    ExtractCode --> Format[Format JSON response]
    DefaultCode --> Format
    Format --> Log[Log with correlation ID]
    Log --> Send[Send response]
```

## Error Handling {#c3-105-errors}

| Error Type | Status | Code |
|------------|--------|------|
| ValidationError | 400 | `validation_error` |
| AuthError | 401 | `auth_*` |
| ForbiddenError | 403 | `forbidden` |
| NotFoundError | 404 | `not_found` |
| Unknown | 500 | `internal_error` |

## Usage {#c3-105-usage}

```typescript
import { errorHandler } from './middleware/errorHandler';

// Must be last middleware
app.use(errorHandler);

// Throwing errors
throw new AppError('Task not found', 'task_not_found', 404);
```

## Health Checks {#c3-105-health}

| Check | Probe | Expectation |
|-------|-------|-------------|
| Middleware registered | Check app middleware stack | Handler present |

## Metrics & Observability {#c3-105-metrics}

| Metric | Type | Description |
|--------|------|-------------|
| `errors_total` | Counter | Errors by code |
| `errors_by_status` | Counter | Errors by HTTP status |
| `unhandled_errors_total` | Counter | Non-operational errors |

## Dependencies {#c3-105-deps}

- **Upstream:** [C3-104-logger](./C3-104-logger.md) for error logging
- **Downstream:** None (terminal middleware)
- **Infra features consumed:** None
