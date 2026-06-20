---
target: c3-106
scope: whole
type: component
parent: c3-1
title: codemap-lib
---
# codemap-lib

## Goal

Parse the code-map that binds entities to source globs, match files against it, and validate that every binding resolves.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | The code-binding layer: it owns reading the code-map file and the glob matching that ties a source file to the component, ref, or rule that governs it. |
| Boundary | Owns code-map parsing, glob expansion, and binding validation; it decides nothing about doc shape (schema) and persists nothing (store) — it works against the filesystem and a passed entity-type map. |
| Collaboration | check passes it the entity-id-to-type map from the graph and the project dir, then surfaces the issues it reports; lookup uses its glob matching to map a file back to its governing entity. |

## Purpose

Parse a `code-map.yaml` into a map from entity id to a list of source-path globs, treating a missing or empty file as an empty map. Expand globs against the project filesystem with doublestar, including a literal-bracket fallback so `[id]`-style framework paths still match. Validate every binding: the id must exist in the C3 graph and be a component, ref, or rule, and each path must be relative, in-bounds, and resolve to a real file or a non-empty glob — each violation reported as a severity-tagged issue. Non-goals: deciding canvas shape (schema), parsing fact bodies (doc-model), and reading or writing the database (store).

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-frontmatter-docs | ref | The entity ids this layer binds to are the ids declared in fact frontmatter | Convention names the entities a binding may reference | A code-map key must resolve to a known entity id from the graph the frontmatter convention produces. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| ParseCodeMap / GlobFiles | IN | Callers pass a file path or a glob pattern; the layer returns the parsed id-to-globs map (empty when the file is absent) or the files a pattern matches | Tolerant of an absent or empty map file; the glob fallback re-tries with escaped brackets before giving up | codemap_test.go |
| Validate | OUT | Given the parsed map, the entity-id-to-type map, and the project dir, returns a list of issues for unknown ids, wrong-type ids, and unresolved or out-of-bounds paths | Returns findings only; it never mutates the map, the graph, or the filesystem, and underscore-prefixed keys are skipped | validate_test.go |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/internal/codemap/**.go | Contract | Matching internals (glob engine, bracket fallback, severity thresholds) may vary as long as bindings resolve the same way | go test ./internal/codemap/... |
