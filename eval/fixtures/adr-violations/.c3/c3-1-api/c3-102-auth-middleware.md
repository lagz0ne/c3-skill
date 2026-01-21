# c3-102: Auth Middleware

## Purpose

JWT-based authentication middleware that protects API routes.

## Location

`src/middleware/auth.ts`

## Responsibilities

- Extract Bearer token from Authorization header
- Verify JWT signature
- Set userId in request context
- Return 401 for invalid/missing tokens

## Owns

All authentication concerns. Other components must NOT implement auth logic.

## Hand-offs

- Receives: Raw HTTP request
- Passes to: Route handlers with authenticated context
