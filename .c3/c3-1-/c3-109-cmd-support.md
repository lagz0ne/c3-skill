---
id: c3-109
c3-seal: 915e81bec366413c2037a9cf23da32cba4a49bc036f243d94cbc2e8b99a1b4ac
title: cmd-support
type: component
category: foundation
parent: c3-1
goal: Provide the shared command-layer scaffolding for the c3x CLI — the authoritative command registry that drives help text, the global argument parser that turns argv into typed options, and the common output helper every command reuses.
uses:
    - rule-output-via-helpers
---

# cmd-support

## Goal

Provide the shared command-layer scaffolding for the c3x CLI — the authoritative command registry that drives help text, the global argument parser that turns argv into typed options, and the common output helper every command reuses.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | The command-layer plumbing under the Go CLI: it sits beneath every concrete command, supplying the help/capabilities text, the parsed Options struct, and the JSON/TOON writer they all share. |
| Boundary | Owns the cross-command scaffolding only — the command metadata registry, argv parsing, and the shared writeJSON helper; it owns no single command's behavior and persists nothing. |
| Collaboration | main.go parses argv through ParseArgs, then dispatches to a concrete command which reads its flags off Options and renders help via ShowHelp; commands route structured output through the shared helper rather than printing it themselves. |

## Purpose

Owns three pieces of command-layer scaffolding shared across every c3x command: the `Commands` registry in `help.go` (the single source of truth for `--help`, per-command help, and `capabilities`), the `ParseArgs` flag/argument parser in `options.go` (one `Options` struct, agent-mode `--json` defaulting), and the `writeJSON` output helper in `helpers.go` (TOON in agent mode, indented JSON otherwise). Non-goals: implementing any concrete command's logic (each lives in its own cmd file), the full `WriteTableOutput`/`WriteObjectOutput` table-serialization helpers, dispatch and runtime wiring (runtime-support), and document or schema validation.

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| rule-output-via-helpers | rule | The shared writeJSON helper this component owns is the format-switch primitive every command's structured output funnels through, honoring agent-mode TOON vs explicit JSON. | Standard applies wherever a command emits structured output | help.go and capabilities text are prose-path output and are exempt; the rule binds the structured writeJSON path. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| ParseArgs | IN | Turn an argv slice into a typed Options struct, defaulting depth/limit and routing agent mode to --json-style structured output. | Parses flags only; it never validates command semantics or executes anything. | cli/cmd/options.go |
| ShowHelp / ShowCapabilities | OUT | Render global help, per-command help, and the capabilities table from the single Commands registry. | Reads the registry; it does not define command behavior. | cli/cmd/help.go |
| writeJSON | OUT | Serialize any value as TOON in agent mode and indented JSON otherwise, for commands to reuse. | The low-level writer only; it is not the table/object output-format switch. | cli/cmd/helpers.go |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/cmd/{help,helpers,options}.go | Contract | Internal field layout and flag set may grow as long as the registry/parse/write contract holds | go test ./cmd/... |
