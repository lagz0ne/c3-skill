# c3-103: Task Routes

## Purpose

CRUD operations for task management.

## Location

`src/routes/tasks.ts`

## Responsibilities

- Task CRUD operations
- Task validation
- Task business rules

## Does NOT Own

- Authentication (handled by c3-102)
- Request coordination (handled by container)

## Dependencies

- c3-102: Provides authenticated user context

## Hand-offs

- Receives: Authenticated request from c3-102
- Returns: Task data or errors
