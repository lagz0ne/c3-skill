---
target: c3-104
scope: whole
type: component
parent: c3-1
title: changeset
---
# changeset

## Goal

Run the change-unit saga: parse a folder of patch material, gate it against the frozen facts, and commit it atomically as the only legal mutation of a fact.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | The mutation engine: the single path through which a frozen fact ever changes, gated and transactional. |
| Boundary | Owns parsing patch/codemap/inspect material and the apply transaction; it never decides canvas shape (schema) or renders facts (doc-model) — it commits through the store. |
| Collaboration | doc-model supplies the node trees a patch rewrites; schema's canvas validates the merged body at the gate; the store opens the transaction this saga runs inside. |

## Purpose

Parse `*.patch.md` primitives (block, insert, whole, frontmatter, retire, canvas) plus the unit's external `*.codemap.md` and `*.inspect.md` carriers, then commit them inside one transaction: check every anchor for drift, apply each primitive, re-derive body-owned edges, and reconcile each affected parent's membership table — so the frozen result is membership-consistent by construction. Non-goals: deciding when to mutate (the CLI dispatcher), validating canvas section sets (schema), or parsing markdown bodies (doc-model).

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| rule-wrap-error-cause | rule | Every apply/parse failure returned from the saga must wrap its cause with file-and-stage context | Diagnostic obligation binds all returned errors | A drifted anchor or a mid-apply failure surfaces which patch file and stage failed via a wrapped `%w`. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| ParsePatch / ReadPatchDir | IN | Callers pass a patch folder; each file parses to a typed Patch, and a no-base edit of an existing fact is refused — only whole/canvas scope may omit the base anchor | Integrity by construction: an edit that does not anchor is rejected at parse, never applied | patch_test.go |
| Apply | OUT | Given the parsed patches and codemap carriers, checks drift then writes the whole unit in one store transaction — a single drifted anchor or landing-hash mismatch rolls back every prior write | The unit lands completely or not at all; no half-matched state between facts and code bindings | apply_test.go atomic rollback |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/internal/changeset/**.go | Contract | Anchor representation and the order of internal saga steps may vary as long as the unit stays atomic and drift-gated | go test ./internal/changeset/... |
