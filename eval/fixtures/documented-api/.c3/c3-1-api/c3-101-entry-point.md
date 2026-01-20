# c3-101: Entry Point

## Purpose

Application entry point that bootstraps the Hono server and mounts all routes.

## Location

`src/index.ts`

## Responsibilities

- Create Hono app instance
- Apply global middleware
- Mount route modules
- Export app for server startup

## Dependencies

- c3-102 (auth-middleware)
- c3-103 (task-routes)
- c3-104 (user-routes)

## Code Pattern

```typescript
const app = new Hono();
app.use("/api/*", authMiddleware);
app.route("/api/tasks", taskRoutes);
app.route("/api/users", userRoutes);
```
