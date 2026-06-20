---
target: c3-105
scope: whole
type: component
parent: c3-1
title: walker
---
# walker

## Goal

Discover every fact file under the `.c3/` directory tree and parse each one for check, list, and import.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | The discovery layer: it is the one place that recursively walks `.c3/` and turns the files on disk into a list of parsed documents. |
| Boundary | Owns the directory traversal and per-file frontmatter parse; it does not interpret a document beyond parsing its frontmatter, and it never writes anything. |
| Collaboration | It calls doc-model's frontmatter parser on each file; check, list, and import consume the parsed-doc list it returns to load or verify the graph. |

## Purpose

Recursively walk a `.c3/` directory with filepath.Walk, read every `.md` file, parse its frontmatter and body, and return the successfully parsed documents alongside warnings for files that have frontmatter delimiters but fail to parse. Skip the generated `_index` directory so derived artifacts are never re-ingested, and derive a slug from a file path by stripping the id prefix (or the parent directory name for README containers). Non-goals: parsing markdown bodies into node trees (doc-model), resolving canvas shape (schema), and persisting anything (store).

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-frontmatter-docs | ref | The frontmatter shape every walked `.md` file must carry to be recognized as a fact | Convention decides what counts as a fact file | A file with `---` delimiters that fails to parse is surfaced as a warning, not silently dropped. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| WalkC3Docs / WalkC3DocsWithWarnings | IN | Callers pass a `.c3` directory; the walker reads every `.md` under it and returns parsed docs (path relative to `.c3/`) plus parse warnings | Read-only filesystem traversal; the `_index` directory is skipped and a read or walk error aborts cleanly | walker_test.go |
| SlugFromPath | OUT | Returns a stable slug for a file path by stripping the recognized id prefix, deriving a README container's slug from its parent directory | Pure string derivation over a path; touches no filesystem and returns "" for the top-level context README | walker_test.go slug cases |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/internal/walker/**.go | Contract | Traversal internals (skip rules, slug prefix pattern) may vary as long as discovery returns the same docs and warnings | go test ./internal/walker/... |
