# C3 Documentation Table of Contents

> **⚠️ AUTO-GENERATED** - Do not edit manually. Regenerate with: `.c3/scripts/build-toc.sh`
>
> Last generated: 2025-11-21 10:11:07

## Context Level

### [CTX-001-system-overview](./CTX-001-system-overview.md) - TaskFlow System Architecture Overview
> Explains the overall TaskFlow system landscape, how users interact with the
application, and how different containers communicate. Read this to understand
the bird's-eye view before diving into individual containers.

**Sections**:

---

## Container Level

### [CON-001-backend](./containers/CON-001-backend.md) - Backend API Container
> Describes the backend API service architecture, middleware pipeline,
and component organization. Read this to understand how the backend
handles requests, manages authentication, and structures its components.

**Sections**:

---

## Component Level

### Backend Components

#### [COM-001-db-pool](./components/backend/COM-001-db-pool.md) - Database Connection Pool Component
> Explains PostgreSQL connection pooling strategy, configuration, and
retry behavior. Read this to understand how the backend manages database
connections efficiently and handles connection failures.

**Sections**:

---

## Architecture Decisions

### [ADR-001-rest-api](./adr/ADR-001-rest-api.md) - Use REST API for Client-Server Communication
> Documents the decision to use REST API over alternatives like GraphQL or gRPC
for client-server communication in TaskFlow. Read this to understand the
reasoning, trade-offs, and when this decision might be revisited.

**Status**: Accepted

**Sections**:

---

## Quick Reference

**Total Documents**: 4
**Contexts**: 1 | **Containers**: 1 | **Components**: 1 | **ADRs**: 1
