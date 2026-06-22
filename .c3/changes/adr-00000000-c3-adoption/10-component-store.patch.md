---
target: c3-102
scope: whole
type: component
parent: c3-1
title: store
---
# store

## Goal

Persist the architecture graph — entities, relationships, nodes, versions, code-map, and semantic vectors — behind one SQLite-backed API.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | The persistence foundation every other CLI component reads and writes through. |
| Boundary | Owns the database schema and all SQL; exposes typed operations, never raw rows, to callers. |
| Collaboration | doc-model bridges documents to nodes here; changeset commits mutations here; read/check query here. |

## Purpose

Own the SQLite store and its schema: insert/get/update/delete entities and relationships, write and hash node trees, snapshot versions, hold the code-map rows, and back hybrid semantic search. Non-goals: parsing markdown (doc-model), validating canvases (schema), or deciding when to mutate (changeset).

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-frontmatter-docs | ref | How document identity and fields map onto stored entities and nodes | Convention informs schema | Frontmatter fields become entity columns. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| store API | IN | Callers pass typed entities/relationships/nodes; the store validates ids and writes atomically | No raw SQL escapes the package | internal/store/*.go |
| store API | OUT | Returns typed results and stable errors; seals node trees with a content merkle | Callers never see database/sql types | store_test.go round-trips |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/internal/store/**.go | Contract | Implementation detail (indexes, query plans) may vary | go test ./internal/store/... |
