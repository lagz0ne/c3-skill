# c3-1-api: REST API Server

## Purpose

Main API server handling all HTTP requests for task management.

## Technology

- **Framework**: Hono
- **Runtime**: Bun
- **Auth**: JWT tokens

## Components

| ID | Name | Type | Purpose |
|----|------|------|---------|
| c3-101 | entry-point | Foundation | App bootstrap and route mounting |
| c3-102 | auth-middleware | Foundation | JWT verification |
| c3-103 | task-routes | Feature | Task CRUD operations |
| c3-104 | user-routes | Feature | User profile management |
| c3-105 | database | Foundation | Database abstraction layer |

## Data Flow

```
Request → Auth Middleware → Route Handler → Database → Response
```

## External Dependencies

- Database (SQL)
- JWT secret (environment)
