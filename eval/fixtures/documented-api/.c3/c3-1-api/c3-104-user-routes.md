# c3-104: User Routes

## Purpose

User profile management endpoints.

## Location

`src/routes/users.ts`

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/users/me | Get current user profile |
| PATCH | /api/users/me | Update current user profile |

## Dependencies

- c3-102 (auth-middleware) - for userId
- c3-105 (database) - for persistence
