---
id: c3-201
c3-version: 3
title: Entry Point
type: component
category: foundation
parent: c3-2
summary: Server bootstrap with scope creation, signal handling, and graceful shutdown
---

# Entry Point

Provides the application entry point (entry.ts) that creates the root scope with extensions and tags, starts the appropriate server (dev/production), and handles process signals for graceful shutdown.

## Contract

| Provides | Expects |
|----------|---------|
| Root scope with extensions | Environment variables (PORT, NODE_ENV) |
| Server startup (dev/prod) | PGURI, OTEL_* environment |
| Signal handlers (SIGTERM, SIGINT) | OS signals |
| Graceful shutdown with dispose | Scope cleanup support |

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Missing PGURI | Database connection fails at startup |
| Double shutdown signal | shuttingDown flag prevents re-entry |
| Uncaught exception | Logged and triggers shutdown |
| Unhandled rejection | Logged and triggers shutdown |

## Testing

| Scenario | Verifies |
|----------|----------|
| Dev mode startup | NODE_ENV=development uses devServer |
| Prod mode startup | NODE_ENV=production uses productionServer |
| Signal handling | Send SIGTERM, verify dispose called |
| OTEL configuration | Tags set from environment |

## References

- `apps/start/src/server.tsx` - Main entry point
- `apps/start/src/server/resources/*Server.ts` - Server configurations
