---
id: c3-103
c3-version: 4
c3-seal: e5d99823565c0d904ca6bb27c88f5965b6e27260ffbfc37059dd38670bfce4f0
title: templates
type: component
category: foundation
parent: c3-1
goal: Provide embedded markdown templates for scaffolding new `.c3/` docs (context, container, component, ref, rule, ADR).
summary: Go embed bundle of all doc templates; used by init-cmd and add-cmd to create well-structured stubs
---

# templates
## Goal

Provide embedded markdown templates for scaffolding new `.c3/` docs (context, container, component, ref, rule, ADR).

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-1 |
| Role | Own templates behavior inside the parent container without taking over sibling responsibilities. |
| Boundary | Keep templates decisions inside this component and escalate container-wide policy to the parent. |
| Collaboration | Coordinate with cited governance and adjacent components before changing the contract. |
## Purpose

Provide durable agent-ready documentation for templates so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verification evidence.

## Foundational Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Preconditions | Parent container context is loaded before templates behavior is changed. | c3-1 |
| Inputs | Accept only the files, commands, data, or calls that belong to templates ownership. | c3-1 |
| State / data | Preserve explicit state boundaries and avoid hidden cross-component ownership. | c3-1 |
| Shared dependencies | Use lower-layer helpers and cited references instead of duplicating shared policy. | c3-1 |
## Business Flow

| Aspect | Detail | Reference |
| --- | --- | --- |
| Actor / caller | Agent, command, or workflow asks templates to deliver its documented responsibility. | c3-1 |
| Primary path | Follow the component goal, honor parent fit, and emit behavior through the documented contract. | c3-1 |
| Alternate paths | When a request falls outside templates ownership, hand it to the parent or sibling component. | c3-1 |
| Failure behavior | Surface mismatch through check, tests, lookup, or review evidence before derived work ships. | c3-1 |
## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| c3-1 | policy | Governs templates behavior, derivation, or review when applicable. | Explicit cited governance beats uncited local prose. | Migrated from legacy component form; refine during next component touch. |
## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| embedded ADR template | OUT | ADR scaffold must include Goal, Context, Decision, Work Breakdown, Underlay C3 Changes, Enforcement Surfaces, Alternatives Considered, Risks, and Verification headings with table headers where schema expects tables. | c3-103 template ownership; schema meaning belongs to c3-113. | cli/internal/templates/adr.md; cli/internal/templates/templates_test.go TestRead_ADRTemplateIncludesDecisionLedger. |
| template loading | OUT | Templates remain embedded in the CLI binary so c3x can create docs without skill-side file templates. | ref-embedded-templates governance. | cli/internal/templates/templates.go; go test ./internal/templates. |
## Change Safety

| Risk | Trigger | Detection | Required Verification |
| --- | --- | --- | --- |
| Contract drift | Goal, boundary, or derived material changes without matching component docs. | Compare Goal, Parent Fit, Contract, and Derived Materials. | Run c3x check and relevant project tests. |
| Governance drift | Cited references, rules, or parent responsibilities change. | Re-read Governance rows and parent container docs. | Run c3x verify plus targeted lookup for changed files. |
## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/internal/templates/adr.md | Contract embedded ADR template row and Governance c3-1 policy. | Placeholder text can stay sparse; headings and table headers must match the ADR schema. | go test ./internal/templates -run TestRead_ADRTemplateIncludesDecisionLedger. |
| cli/internal/templates/templates_test.go | Contract embedded ADR template row and Change Safety contract drift risk. | Test can assert more headings as schema grows. | go test ./internal/templates. |
