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
| c3-102 | auth-middleware | Foundation | JWT verification for all protected routes |
| c3-103 | task-routes | Feature | Task CRUD operations |
| c3-104 | user-routes | Feature | User profile management |

## Data Flow

```
Request → c3-102 (Auth) → c3-103/c3-104 (Routes) → Response
```

## Component Responsibilities

- **c3-102 owns**: All authentication logic (JWT verification, token extraction, user context)
- **c3-103 owns**: Task business logic only (assumes user is authenticated)
- **c3-104 owns**: User profile logic only (assumes user is authenticated)

## Coordination

Request routing and flow coordination is handled at container level via c3-101 entry-point.
Components hand-off to each other; they do not orchestrate.
