# Task API - C3 Architecture

## Overview

Task API is a RESTful backend service for task management.

## Architecture Style

- **Type**: Modular Monolith
- **Runtime**: Bun

## Key Decisions

- JWT-based authentication (all routes use JWT tokens)
- Middleware handles cross-cutting concerns (auth, logging)
- Route handlers focus only on business logic

## Containers

| ID | Name | Purpose |
|----|------|---------|
| c3-1 | api | REST API server |

## Quick Links

- [API Container](c3-1-api/README.md)
