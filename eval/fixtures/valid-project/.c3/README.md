---
id: c3-0
c3-version: 3
title: TestApp
summary: A test application for C3 skill evaluation
---

# TestApp

A simple test application with API and database components.

## Actors

- **User**: End user accessing the application
- **Admin**: Administrator managing the system

## Containers

```mermaid
graph TD
    U[User] --> API[c3-1 API Service]
    A[Admin] --> API
    API --> DB[(Database)]
```

## External Systems

- **Database**: PostgreSQL database for persistence

## Container Index

| ID | Name | Purpose |
|----|------|---------|
| c3-1 | API Service | REST API handling user requests |
