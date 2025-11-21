# C3 Documentation Table of Contents

> **AUTO-GENERATED** - Do not edit manually. Regenerate with: `.c3/scripts/build-toc.sh`
>
> Last generated: 2025-11-21 10:31:42

## Context Level

### [CTX-001-system-overview](./CTX-001-system-overview.md) - TaskFlow System Architecture Overview
> Explains the overall TaskFlow system landscape, how users interact with the
application, and how different containers communicate. Read this to understand
the bird's-eye view before diving into individual containers.

**Sections**:
- [Overview](#ctx-001-overview) - 
Describes the system at the highest level - what it does, who uses it,
and what the major components are. Read to understand the big picture.
- [Architecture](#ctx-001-architecture) - 
Shows the complete system diagram with all containers, external systems,
and their relationships. Read to understand how pieces fit together.
- [Containers](#ctx-001-containers) - 
Lists all containers with brief descriptions and links. Read to navigate
to specific container details.
- [Protocols & Communication](#ctx-001-protocols) - 
Explains communication protocols used across the system and why chosen.
Read to understand integration patterns.
- [Cross-Cutting Concerns](#ctx-001-cross-cutting) - 
Describes concerns that span multiple containers like authentication,
logging, and monitoring. Read to understand system-wide patterns.
- [Deployment](#ctx-001-deployment) - 
High-level deployment architecture - cloud vs on-prem, scaling approach,
infrastructure patterns. Read to understand operational context.
- [Related](#ctx-001-related)

---

## Container Level

### [CON-001-backend](./containers/CON-001-backend.md) - Backend API Container
> Describes the backend API service architecture, middleware pipeline,
and component organization. Read this to understand how the backend
handles requests, manages authentication, and structures its components.

**Sections**:
- [Overview](#con-001-overview) - 
High-level description of container purpose and responsibilities.
- [Technology Stack](#con-001-technology-stack) - 
Lists languages, frameworks, and key libraries used. Read to understand
the technical foundation.
- [Middleware Pipeline](#con-001-middleware) - 
Describes the request processing pipeline through authentication, cookie
handling, and rate limiting layers. Read this to understand how requests
flow through the backend before reaching business logic.
- [Component Organization](#con-001-components) - 
Shows how components are structured inside the container.
- [API Endpoints](#con-001-api-endpoints) - 
Documents the main API endpoints exposed by this container.
- [Communication Patterns](#con-001-communication) - 
Explains how this container talks to other containers.
- [Data Responsibilities](#con-001-data) - 
What data this container owns and manages.
- [Configuration](#con-001-configuration) - 
Environment-based configuration for this container.
- [Deployment](#con-001-deployment) - 
Container-specific deployment characteristics.
- [Related](#con-001-related)

---

### [CON-002-frontend](./containers/CON-002-frontend.md) - Web Frontend Container
> Describes the React-based web frontend architecture, component organization,
state management, and API integration patterns. Read this to understand how
the frontend handles user interactions, manages application state, and
communicates with the backend API.

**Sections**:
- [Overview](#con-002-overview) - 
High-level description of container purpose and responsibilities.
- [Technology Stack](#con-002-technology-stack) - 
Lists languages, frameworks, and key libraries used. Read to understand
the technical foundation.
- [Component Organization](#con-002-components) - 
Shows how components are structured inside the container.
- [State Management](#con-002-state) - 
How application state is organized and managed.
- [Routing](#con-002-routing) - 
Client-side routing configuration.
- [Communication Patterns](#con-002-communication) - 
How this container communicates with backend.
- [Configuration](#con-002-configuration) - 
Environment-based configuration for this container.
- [Build & Deployment](#con-002-deployment) - 
Container-specific deployment characteristics.
- [Related](#con-002-related)

---

## Component Level

### Backend Components

#### [COM-001-db-pool](./components/backend/COM-001-db-pool.md) - Database Connection Pool Component
> Explains PostgreSQL connection pooling strategy, configuration, and
retry behavior. Read this to understand how the backend manages database
connections efficiently and handles connection failures.

**Sections**:
- [Overview](#com-001-overview) - 
What this component does and why it exists.
- [Purpose](#com-001-purpose) - 
Specific responsibilities and goals.
- [Technical Implementation](#com-001-implementation) - 
How it's built - libraries, patterns, architecture.
- [Configuration](#com-001-configuration) - 
Explains environment variables, configuration loading strategy, and
differences between development and production. Read this section to
understand how to configure the connection pool for different environments.
- [Connection Pool Behavior](#com-001-pool-behavior) - 
How the pool manages connections, sizing strategy, lifecycle.
- [Error Handling](#com-001-error-handling) - 
How errors are handled, retry strategy, error types.
- [Performance](#com-001-performance) - 
Performance characteristics, benefits, monitoring.
- [Health Checks](#com-001-health-checks) - 
Health check implementation and frequency.
- [Usage Example](#com-001-usage) - 
How to use this component in application code.
- [Related](#com-001-related)

---

#### [COM-002-auth-middleware](./components/backend/COM-002-auth-middleware.md) - Authentication Middleware Component
> Explains JWT validation middleware, token extraction, and user context
injection. Read this to understand how the backend authenticates requests
and protects API endpoints.

**Sections**:
- [Overview](#com-002-overview) - 
What this component does and why it exists.
- [Purpose](#com-002-purpose) - 
Specific responsibilities and goals.
- [Technical Implementation](#com-002-implementation) - 
How it's built - libraries, patterns, architecture.
- [Configuration](#com-002-configuration) - 
Environment variables and configuration options.
- [Token Extraction](#com-002-token-extraction) - 
How tokens are extracted from requests.
- [Token Validation](#com-002-token-validation) - 
JWT verification process.
- [User Context](#com-002-user-context) - 
How user information is injected into request.
- [Public Routes](#com-002-public-routes) - 
Routes that bypass authentication.
- [Main Middleware](#com-002-main-middleware) - 
The assembled middleware function.
- [Error Handling](#com-002-error-handling) - 
Authentication-specific errors.
- [Testing](#com-002-testing) - 
How to test this component.
- [Related](#com-002-related)

---

#### [COM-003-task-service](./components/backend/COM-003-task-service.md) - Task Service Component
> Explains the business logic layer for task operations including validation,
authorization, and domain rules. Read this to understand how task CRUD
operations are processed and what business rules are enforced.

**Sections**:
- [Overview](#com-003-overview) - 
What this component does and why it exists.
- [Purpose](#com-003-purpose) - 
Specific responsibilities and goals.
- [Technical Implementation](#com-003-implementation) - 
How it's built - libraries, patterns, architecture.
- [Service Interface](#com-003-interface) - 
Public API of the service.
- [Business Rules](#com-003-business-rules) - 
Domain rules enforced by the service.
- [Authorization](#com-003-authorization) - 
How permissions are checked.
- [Service Implementation](#com-003-service-impl) - 
Core service logic.
- [Events](#com-003-events) - 
Events emitted by the service.
- [Error Handling](#com-003-error-handling) - 
Service-specific error scenarios.
- [Testing](#com-003-testing) - 
How to test this service.
- [Related](#com-003-related)

---

### Frontend Components

#### [COM-004-api-client](./components/frontend/COM-004-api-client.md) - API Client Component
> Explains the Axios-based HTTP client configuration, request/response
interceptors, and error handling. Read this to understand how the frontend
communicates with the backend API and handles authentication.

**Sections**:
- [Overview](#com-004-overview) - 
What this component does and why it exists.
- [Purpose](#com-004-purpose) - 
Specific responsibilities and goals.
- [Technical Implementation](#com-004-implementation) - 
How it's built - libraries, patterns, architecture.
- [Client Configuration](#com-004-configuration) - 
How the Axios client is configured.
- [Request Interceptor](#com-004-request-interceptor) - 
How outgoing requests are modified.
- [Response Interceptor](#com-004-response-interceptor) - 
How responses are processed.
- [Token Refresh](#com-004-token-refresh) - 
How expired tokens are refreshed.
- [Error Handling](#com-004-error-handling) - 
How API errors are processed.
- [API Methods](#com-004-api-methods) - 
Typed methods for specific endpoints.
- [Usage Examples](#com-004-usage) - 
How to use the API client in components.
- [Testing](#com-004-testing) - 
How to test API client usage.
- [Related](#com-004-related)

---

## Architecture Decisions

### [ADR-004-jwt-auth](./adr/ADR-004-jwt-auth.md) - Use JWT for API Authentication
> Documents the decision to use JWT tokens for API authentication over
session-based auth. Read this to understand the token structure,
refresh flow, and security considerations.

**Status**: Accepted

**Sections**:
- [Status](#adr-004-status)
- [Context](#adr-004-context) - 
Current situation and why change/decision is needed.
- [Decision](#adr-004-decision) - 
High-level approach with reasoning.
- [Alternatives Considered](#adr-004-alternatives) - 
What else was considered and why rejected.
- [Token Security](#adr-004-security) - 
Security measures for token handling.
- [Implementation Details](#adr-004-implementation-details) - 
Technical implementation specifics.
- [Consequences](#adr-004-consequences) - 
Positive, negative, and mitigation strategies.
- [Cross-Cutting Concerns](#adr-004-cross-cutting) - 
Impacts that span multiple levels.
- [Revisit Triggers](#adr-004-revisit)
- [Related](#adr-004-related)

---

### [ADR-003-caching-strategy](./adr/ADR-003-caching-strategy.md) - Implement Redis Caching for Performance
> Proposes adding Redis as a caching layer to reduce database load and
improve response times. Read this to understand the caching strategy,
cache invalidation approach, and implementation plan.

**Status**: Proposed

**Sections**:
- [Status](#adr-003-status)
- [Context](#adr-003-context) - 
Current situation and why change/decision is needed.
- [Decision](#adr-003-decision) - 
High-level approach with reasoning.
- [Cache Patterns](#adr-003-patterns) - 
How caching will be implemented.
- [Alternatives Considered](#adr-003-alternatives) - 
What else was considered and why rejected.
- [Consequences](#adr-003-consequences) - 
Positive, negative, and mitigation strategies.
- [Implementation Plan](#adr-003-implementation) - 
Ordered steps for implementation.
- [Success Metrics](#adr-003-metrics)
- [Risk Assessment](#adr-003-risks)
- [Revisit Triggers](#adr-003-revisit)
- [Related](#adr-003-related)

---

### [ADR-002-postgresql](./adr/ADR-002-postgresql.md) - Use PostgreSQL for Primary Data Storage
> Documents the decision to use PostgreSQL as the primary database over
alternatives like MySQL, MongoDB, or SQLite. Read this to understand
the database choice rationale and operational considerations.

**Status**: Accepted

**Sections**:
- [Status](#adr-002-status)
- [Context](#adr-002-context) - 
Current situation and why change/decision is needed.
- [Decision](#adr-002-decision) - 
High-level approach with reasoning.
- [Alternatives Considered](#adr-002-alternatives) - 
What else was considered and why rejected.
- [Consequences](#adr-002-consequences) - 
Positive, negative, and mitigation strategies.
- [Implementation Notes](#adr-002-implementation) - 
Ordered steps for implementation.
- [Cross-Cutting Concerns](#adr-002-cross-cutting) - 
Impacts that span multiple levels.
- [Revisit Triggers](#adr-002-revisit)
- [Related](#adr-002-related)

---

### [ADR-001-rest-api](./adr/ADR-001-rest-api.md) - Use REST API for Client-Server Communication
> Documents the decision to use REST API over alternatives like GraphQL or gRPC
for client-server communication in TaskFlow. Read this to understand the
reasoning, trade-offs, and when this decision might be revisited.

**Status**: Accepted

**Sections**:
- [Status](#adr-001-status)
- [Context](#adr-001-context) - 
Current situation and why change/decision is needed.
- [Decision](#adr-001-decision) - 
High-level approach with reasoning.
- [Alternatives Considered](#adr-001-alternatives) - 
What else was considered and why rejected.
- [Consequences](#adr-001-consequences) - 
Positive, negative, and mitigation strategies.
- [Cross-Cutting Concerns](#adr-001-cross-cutting) - 
Impacts that span multiple levels.
- [Implementation Notes](#adr-001-implementation) - 
Ordered steps for implementation.
- [Revisit Triggers](#adr-001-revisit) - 
When should this decision be reconsidered?
- [Related](#adr-001-related)

---

## Quick Reference

**Total Documents**: 11
**Contexts**: 1 | **Containers**: 2 | **Components**: 4 | **ADRs**: 4
