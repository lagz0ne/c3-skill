---
target: c3-107
scope: whole
type: component
parent: c3-1
title: runtime-support
---
# runtime-support

## Goal

Bootstrap the CLI and host the shared runtime — resolve the `.c3/` directory, dispatch the command, serialize output, and serialize concurrent mutations behind a coordinator.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | The process entrypoint and shared plumbing every command runs on top of: argument dispatch, directory resolution, output formatting, and the write coordinator. |
| Boundary | Owns `main`, config resolution, the TOON serializer, and the coordinator; it knows nothing about canvas shape or fact content — it routes argv to a command handler and renders what comes back. |
| Collaboration | The dispatcher hands a resolved store and `.c3/` path to each command; changeset's mutations flow through the coordinator gate; command results render through the TOON output helpers. |

## Purpose

Parse argv into options and dispatch to a command handler (`main.go`), resolve the project's `.c3/` directory by walking up from the working dir (`config`), serialize a mutating command through a per-`.c3/` unix-socket leader so two writers never race (`coord`), and marshal command results to TOON for machine consumers (`toon`). Non-goals: owning any command's behavior, validating canvases (schema), or persisting facts (store) — it is the bootstrap and the shared runtime, not the work.

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| rule-output-via-helpers | rule | Command results serialize through the shared output helpers, never a direct fmt.Print of structured data | Output contract binds every command result path | WriteTableOutput / WriteObjectOutput pick TOON vs JSON; agent mode forces TOON. |
| rule-dispatcher-error-hint | rule | A user-facing error returned from the dispatcher carries an actionable next-step hint | Dispatcher errors must guide the next action | Usage / missing-arg / unknown-command errors append a `hint:` line. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| run / runWithIO | IN | The dispatcher takes argv plus io, resolves the `.c3/` directory and store, routes to one command handler, and returns an error (never calls os.Exit) so it stays testable | A missing `.c3/` or unknown command returns a hint-bearing error, not a panic | main.go returns errors; runWithIO is exercised by cmd tests |
| coord leader / TOON output | OUT | A mutating command is forwarded to (or becomes) the per-`.c3/` write leader so writes serialize; results marshal to TOON for machine consumers | One writer at a time per `.c3/`; the coordinator is a no-op on Windows and degrades to in-process | coord_test.go; toon_test.go round-trips |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/main.go | Contract | Dispatch table layout and the coordinator transport (socket paths, timeouts) may vary as long as writes stay serialized and errors stay testable | go build ./... && go test ./... |
| cli/internal/{config,coord,toon}/**.go | Purpose | Directory-walk, leader, and serializer internals may vary | go test ./internal/coord/... ./internal/toon/... |
