# Task API - C3 Architecture

## Overview

Task API is a RESTful backend service for task management. Built with Hono framework on Bun runtime.

## Architecture Style

- **Type**: Modular Monolith
- **Framework**: Hono (lightweight web framework)
- **Runtime**: Bun
- **Database**: SQL (abstracted)

## Key Decisions

- JWT-based authentication
- Route-based modularization
- Middleware for cross-cutting concerns

## Containers

| ID | Name | Purpose |
|----|------|---------|
| c3-1 | api | REST API server |

## Quick Links

- [API Container](c3-1-api/README.md)
- [ADR: Initial Architecture](adr/adr-00000000-c3-adoption.md)
