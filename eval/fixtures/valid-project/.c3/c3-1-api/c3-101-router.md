---
id: c3-101
type: component
parent: c3-1
category: foundation
title: Router
---

# Router

Route definitions and Express middleware configuration.

## Purpose

Defines all API routes and applies middleware for authentication, validation, and error handling.

## Key Behaviors

- Route registration for /api/* endpoints
- Authentication middleware for protected routes
- Request validation using zod schemas
- Error handling middleware

## References

- `src/api/router.ts` - Main router configuration
- `src/api/middleware/` - Middleware implementations
- `Router` class in `src/api/router.ts:15`
