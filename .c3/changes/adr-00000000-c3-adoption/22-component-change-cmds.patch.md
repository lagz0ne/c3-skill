---
target: c3-112
scope: whole
type: component
parent: c3-1
title: change-cmds
---
# change-cmds

## Goal

Drive the change-unit lifecycle from the command line: scaffold and view a unit, run the apply gates, preview the result, inspect derivation obligations, supersede a decision, and enforce the freeze that makes a change-unit the only way to mutate a fact.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | The user-facing front end of the changeset engine: the verbs an agent runs to stage, review, and land a change-unit. |
| Boundary | Owns the command surface, the gate orchestration, and the freeze guard; the atomic apply transaction itself lives in the changeset library, which these commands call. |
| Collaboration | The changeset library parses patches and runs the apply saga; the store provides the preview transaction the overlay replays into; schema's canvas validates each gated body. |

## Purpose

Serve the change-unit command surface: `change new`/`scaffold` stage a unit and its climb patches, `change view`/`status` project each patch's drift and apply state, `change inspect` shows the derivation obligations and code-map territory to attest, and `change apply` runs the gates (drift, canvas, morph, retire, inspection) before committing atomically. Around them: the freeze guard refuses a direct write/set/delete of a frozen fact, `supersede` flips a terminal decision under a successor, the overlay previews the graph as a unit would leave it, and `conflict`/`materialize` support rebase and seed materialization. Non-goals: the apply transaction's internals (changeset), pre-freeze authoring (author-cmds), or read-only queries (read-cmds).

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| rule-wrap-error-cause | rule | Every gate or apply failure surfaced from a change command wraps its cause with the unit and stage that failed | A wrapped saga error names which patch and phase broke | `change apply` wraps overlay, inspection-gate, and apply errors with `fmt.Errorf(... : %w, err)`. |
| rule-dispatcher-error-hint | rule | A misuse of a change subcommand returns an actionable hint naming the concrete next command | The user always learns the recoverable next step | A missing unit id or unknown subcommand returns an `error:` + `hint: c3x change <sub> <id>` line. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| change apply | IN | Given a unit id, reads its patch/codemap/inspect folder and runs all gates; a single gate failure rejects the whole unit and writes nothing | All-or-nothing across the file/store boundary — canvas morphs roll back if the store apply fails | cli/cmd/change.go RunChangeApply |
| view / status / inspect | OUT | Projects each patch's state, drift, and derivation obligations for review, computed on a rolled-back preview overlay that replays the real apply path | Read-only previews never commit; the overlay is always discarded | cli/cmd/change_view.go buildChangeUnitView; cli/cmd/inspect.go RunChangeInspect |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/cmd/{change,change_scaffold,change_view,inspect,morph}.go | Contract | Gate ordering and preview rendering may vary as long as apply stays atomic and gated | go test ./cmd/... |
| cli/cmd/{overlay,conflict,supersede,destruction,materialize,freeze,adr_linkage,cascade_hints}.go | Purpose | Freeze-guard carve-outs and supersede chain-walk details may vary while the freeze invariant holds | go test ./cmd/... |
