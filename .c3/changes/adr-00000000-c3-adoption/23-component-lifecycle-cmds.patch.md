---
target: c3-113
scope: whole
type: component
parent: c3-1
title: lifecycle-cmds
---
# lifecycle-cmds

## Goal

Keep the `.c3/` store and its canonical markdown coherent across its lifecycle: import a tree into the database, export and sync it back out, repair drift, run the status migration, delete an entity safely, and install Git guardrails.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | The state-lifecycle plumbing of the CLI: the commands that move the project between its canonical markdown form and its SQLite cache and keep the two in sync. |
| Boundary | Owns the round-trip between disk and store (export render, import rebuild, seal verification) and the on-disk guardrails; it leaves entity rendering to doc-model and persistence to the store. |
| Collaboration | doc-model seals and parses each canonical document; the store is opened, rebuilt, and exported through here; walker enumerates the on-disk `.c3/` tree the import and seal checks consume. |

## Purpose

Serve the lifecycle around `.c3/`: `import` rebuilds `c3.db` from the sealed markdown tree (refusing a broken or unsealed seal unless forced), `export`/`sync` render the store back to canonical markdown and reconcile the on-disk tree, `repair` re-imports then re-seals to heal drift, `migrate` sweeps every entity's status onto the canonical set (clearing facts, mapping change docs, failing loud on an unmappable status), `delete` removes an entity and strips every reference to it after refusing to strand a parent or root, and `git install` writes the pre-commit hook and ignore guardrails. Non-goals: validating canvas shape (read-cmds `check`), authoring facts (author-cmds), or running a change-unit (change-cmds).

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| rule-wrap-error-cause | rule | Every filesystem, store, or seal failure crossing a lifecycle boundary wraps its cause with the stage and the file or entity that failed | The import/export/repair chain stays diagnosable to its root cause | import/export/sync/repair/migrate all wrap their errors with `fmt.Errorf(... : %w, err)` naming the path or id. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| import / repair | IN | Reads the canonical `.c3/` tree, verifies each seal, then rebuilds the database into a temp file and atomically swaps it in; a broken seal aborts unless `--force` reseals | Never replaces the live DB until the rebuild succeeds; backs up the prior database first | cli/cmd/import.go RunImport; cli/cmd/repair.go RunRepair |
| export / sync / migrate / delete | OUT | Renders entities to sealed markdown, reconciles the on-disk tree, sweeps statuses, or removes an entity and its references, reporting every change | Refuses to delete the context root or a parent with children; the status sweep is loud and non-silent | cli/cmd/export.go RunExport; cli/cmd/delete.go RunDelete |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/cmd/{import,export,sync,repair}.go | Contract | Temp-file swap mechanics and diff rendering may vary as long as the round-trip stays seal-verified and atomic | go test ./cmd/... |
| cli/cmd/{migrate_status,delete,git,document_format}.go | Purpose | Migration mapping and guardrail block contents may vary while the sweep stays loud and delete stays non-stranding | go test ./cmd/... |
