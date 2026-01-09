# C3 Documentation Table of Contents

> **AUTO-GENERATED** - Do not edit manually. Rebuild using the c3-toc skill.
>
> Last generated: 2026-01-09 09:38:29

## Context Level

### [c3-0](./README.md) - Acountee
> Internal finance tool for managing invoices, payment requests, and multi-step approvals

**Sections**:

---

## Container Level

### [c3-1](./c3-1-web-frontend/) - Web Frontend
> React-based SPA for invoice management, payment requests, and approval workflows

**Sections**:

---

### [c3-2](./c3-2-api-backend/) - API Backend
> Server-side business logic, database operations, authentication, and real-time sync

**Sections**:

---

### [c3-3](./c3-3-e2e-tests/) - E2E Tests
> Playwright smoke tests for critical user journeys with isolated database support (~1 minute runtime)

**Sections**:

---

### [c3-4](./c3-4-postgresql/) - PostgreSQL
> Primary relational database for all business data

**Sections**:

---

### [c3-5](./c3-5-google-oauth/) - Google OAuth
> OAuth2 authentication provider for user identity

**Sections**:

---

### [c3-6](./c3-6-opentelemetry/) - OpenTelemetry
> Distributed tracing and observability infrastructure

**Sections**:

---

### [c3-7](./c3-7-shared/) - Shared Package
> Shared types, schemas, and utilities used across containers

**Sections**:

---

## Component Level

### Web Frontend Components

#### [c3-101](./c3-1-web-frontend/c3-101-router.md) - Router
> TanStack Router configuration with file-based routing and SSR support

**Sections**:

---

#### [c3-102](./c3-1-web-frontend/c3-102-auth-layout.md) - AuthLayout
> Authenticated layout wrapper with navigation, user menu, and state initialization

**Sections**:

---

#### [c3-103](./c3-1-web-frontend/c3-103-state-atoms.md) - State Atoms
> @pumped-fn/lite atoms for reactive client-side state management

**Sections**:

---

#### [c3-104](./c3-1-web-frontend/c3-104-ui-variants.md) - UI Variants
> tailwind-variants definitions for consistent component styling with DaisyUI

**Sections**:

---

#### [c3-121](./c3-1-web-frontend/c3-121-invoice-screen.md) - Invoice Screen
> Invoice listing with filtering, detail view, and status tracking

**Sections**:

---

#### [c3-122](./c3-1-web-frontend/c3-122-payment-requests-screen.md) - Payment Requests Screen
> Main screen for creating, viewing, editing, and managing payment request workflows

**Sections**:

---

#### [c3-124](./c3-1-web-frontend/c3-124-admin-screen.md) - Admin Screen
> Owner-only user administration screen with CRUD operations and ownership transfer

**Sections**:

---

#### [c3-131](./c3-1-web-frontend/c3-131-information-architecture.md) - Information Architecture
> Screen inventory with regions at medium abstraction level for UI development

**Sections**:

---

#### [c3-132](./c3-1-web-frontend/c3-132-user-flows.md) - User Flows
> Exhaustive user flows mapped to IA regions with preconditions and dependencies

**Sections**:

---

#### [c3-133](./c3-1-web-frontend/c3-133-ui-patterns.md) - UI Patterns
> Catalog of UI design patterns with implementation references

**Sections**:

---

### API Backend Components

#### [c3-201](./c3-2-api-backend/c3-201-entry-point.md) - Entry Point
> Server bootstrap with scope creation, signal handling, and graceful shutdown

**Sections**:

---

#### [c3-202](./c3-2-api-backend/c3-202-scope-di.md) - Scope & DI
> @pumped-fn/lite scope management and dependency injection patterns

**Sections**:

---

#### [c3-203](./c3-2-api-backend/c3-203-database-layer.md) - Database Layer
> Drizzle ORM setup with PostgreSQL, schema definitions, and transaction support

**Sections**:

---

#### [c3-204](./c3-2-api-backend/c3-204-logger.md) - Logger
> Pino logger configuration with structured logging and request correlation

**Sections**:

---

#### [c3-205](./c3-2-api-backend/c3-205-middleware.md) - Middleware
> TanStack middleware for execution context and user authentication from cookies

**Sections**:

---

#### [c3-221](./c3-2-api-backend/c3-221-pr-flows.md) - PR Flows
> Payment request business flows for CRUD operations, approvals, and status management

**Sections**:

---

#### [c3-224](./c3-2-api-backend/c3-224-auth-flows.md) - Auth Flows
> Authentication flows for Google OAuth and test token auth

**Sections**:

---

#### [c3-224](./c3-2-api-backend/c3-224-user-flows.md) - User Flows
> 5 user management flows for CRUD and ownership transfer with owner guard pattern

**Sections**:

---

## Refs

Shared conventions and patterns referenced by components across containers.

### [ref-form-patterns](./refs/ref-form-patterns.md) - Form Patterns
> Conventions for building forms with validation, state management, and UI components

---

### [ref-data-sync](./refs/ref-data-sync.md) - Data Sync
> Real-time data synchronization via WebSocket with delta updates and optimistic UI

---

### [ref-error-handling](./refs/ref-error-handling.md) - Error Handling
> Conventions for catching, displaying, and recovering from errors in the UI

---

### [ref-design-system](./refs/ref-design-system.md) - Design System
> DaisyUI-based design system with tailwind-variants for type-safe component styling

---

### [ref-visual-specs](./refs/ref-visual-specs.md) - Visual Specs
> Typography scale, spacing tokens, and visual wireframes for consistent screen generation

---

### [ref-flow-patterns](./refs/ref-flow-patterns.md) - Flow Patterns
> Conventions for defining business logic flows using @pumped-fn/lite

---

### [ref-query-patterns](./refs/ref-query-patterns.md) - Query Patterns
> Conventions for database queries using Drizzle ORM wrapped in service atoms

---

### [ref-realtime-sync](./refs/ref-realtime-sync.md) - Real-time Sync
> Server-side WebSocket sync for broadcasting data changes to connected clients

---

## Architecture Decisions

### [adr-20260108-visual-specs](./adr/adr-20260108-visual-specs.md) - Add Visual Specs Component for Screen Generation Consistency
> 

**Status**: Implemented

**Sections**:

---

### [adr-20260107-shadcn-ui-migration](./adr/adr-20260107-shadcn-ui-migration.md) - Migrate UI from DaisyUI to shadcn/ui
> 

**Status**: Implemented

**Sections**:

---

### [adr-20260106-user-role-management](./adr/adr-20260106-user-role-management.md) - User and Role Management System
> 

**Status**: Implemented

**Sections**:

---

### [adr-20260106-dev-browser-e2e](./adr/adr-20260106-dev-browser-e2e.md) - Simplify E2E Tests to Focused Smoke Tests
> 

**Status**: Implemented

**Sections**:

---

### [adr-00000000-c3-adoption](./adr/adr-00000000-c3-adoption.md) - C3 Architecture Documentation Adoption
> 

**Status**: Implemented

**Sections**:

---

## Quick Reference

**Total Documents**: 39
**Contexts**: 1 | **Containers**: 7 | **Components**: 18 | **Refs**: 8 | **ADRs**: 5 (implemented)
