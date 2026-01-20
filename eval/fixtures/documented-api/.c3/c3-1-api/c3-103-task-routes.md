# c3-103: Task Routes

## Purpose

CRUD operations for task management.

## Location

`src/routes/tasks.ts`

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /api/tasks | List user's tasks |
| POST | /api/tasks | Create new task |
| PATCH | /api/tasks/:id | Update task |
| DELETE | /api/tasks/:id | Delete task |

## Dependencies

- c3-102 (auth-middleware) - for userId
- c3-105 (database) - for persistence

## Data Model

```typescript
interface Task {
  id: number;
  title: string;
  description: string;
  user_id: string;
}
```
