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

## Dependencies

- `src/lib/jwt.ts` for token verification

## Security Notes

- Tokens must be passed as `Bearer <token>`
- Invalid tokens result in 401 response
- User ID is extracted from token's `sub` claim
