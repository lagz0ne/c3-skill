# C3 Architecture Documentation

This directory contains the C3 (Context, Containers, Components) architecture documentation for the T3 Blog Application.

## What is C3?

C3 is a simplified version of the C4 model, focusing on three levels of abstraction:

| Level | Description | Audience |
|-------|-------------|----------|
| **Context** | System context showing external actors and systems | Everyone |
| **Containers** | High-level technology choices and deployable units | Developers, Architects |
| **Components** | Key components within each container | Developers |

## Documentation Structure

- [`01-context.md`](./01-context.md) - System context and external dependencies
- [`02-containers.md`](./02-containers.md) - Container architecture (Next.js, Database, Auth)
- [`03-components.md`](./03-components.md) - Component breakdown by domain
- [`decisions/`](./decisions/) - Architecture Decision Records (ADRs)

## Quick Start

For a high-level understanding, start with the Context diagram, then drill into Containers and Components as needed.

## Diagrams

All diagrams are rendered using [diashort](https://diashort.apps.quickable.co) with Mermaid syntax for easy maintenance and version control.
