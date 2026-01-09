---
id: c3-102
type: component
parent: c3-1
category: feature
title: Handlers
---

# Handlers

Request handlers implementing business logic for each API endpoint.

## Purpose

Process incoming requests, interact with models, and return responses.

## Key Behaviors

- CRUD operations for users
- Input validation
- Error responses with proper status codes

## References

- `src/api/handlers/` - Handler implementations
- `UserHandler` class in `src/api/handlers/user.ts:10`
- `createUser` function in `src/api/handlers/user.ts:25`
