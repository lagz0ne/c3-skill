---
id: adr-20260415-strict-enriched-component-docs
c3-seal: e279b5f825f39b13b9182efabb92eafede951b73b5181683a84fe21b0e145b5b
title: strict-enriched-component-docs
type: adr
goal: 'Apply strict enriched component documentation: replace thin component stubs with top-down, evidence-backed, derivation-ready sections and enforce them through add/write/set/check/verify gates.'
status: implemented
date: "2026-04-15"
---

# strict-enriched-component-docs
## Goal

Apply strict enriched component documentation: replace thin component stubs with top-down, evidence-backed, derivation-ready sections and enforce them through add/write/set/check/verify gates.

## Context

Component docs became the source material for subsequent code, tests, prompts, and migration repair. Thin sections such as a single Goal or generic dependency note were not enough for LLM runs because they omitted ownership, parent fit, governance, contracts, failure behavior, and verification evidence. The strict format was introduced to make each component usable as both documentation and an execution harness.

## Decision

Require component docs to use a strict, reviewer-ready section set:

| Section | Required role | Why it matters |
| --- | --- | --- |
| Goal | One clear behavior sentence. | Keeps component intent stable across generated code and docs. |
| Parent Fit | Parent, role, boundary, and collaboration. | Preserves top-down C3 topology: context -> container -> component. |
| Purpose | Ownership and non-goals in prose. | Prevents components from becoming vague stubs. |
| Foundational Flow | Preconditions, inputs, state/data, shared dependencies. | Captures base mechanics before business behavior. |
| Business Flow | Actor/caller, primary path, alternate paths, failure behavior. | Records what workflow the component demonstrates. |
| Governance | Cited refs, rules, ADRs, specs, policies, or examples. | Makes shared constraints explicit and precedence-driven. |
| Contract | IN/OUT/IN-OUT surfaces, boundaries, and evidence. | Lets future code and derived materials know what must remain stable. |
| Change Safety | Risks, triggers, detection, required verification. | Turns docs into a review checklist. |
| Derived Materials | What code/docs/tests/prompts must derive from. | Keeps subsequent generated work tied to the component contract. |
## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| Strict section registry | componentSectionOrder in cli/cmd/strict_component_docs.go defines the required section order and duplicate-section checks. | go test -count=1 ./cmd -run TestStrictComponent |
| Required table shape | strictTableHeaders and strictMinRows require concrete tables for Parent Fit, Foundational Flow, Business Flow, Governance, Contract, Change Safety, and Derived Materials. | cli/cmd/strict_component_docs_test.go |
| Component body validation | validateBodyContent calls validateStrictComponentDoc for component bodies. | cli/cmd/write.go; cli/cmd/readwrite_test.go |
| Add command gate | RunAdd rejects new component bodies that miss required strict sections. | cli/cmd/add.go; cli/cmd/add_test.go |
| Full write gate | RunWrite validates complete component bodies before node-tree storage. | cli/cmd/write.go; cli/cmd/readwrite_test.go |
| Section mutation gate | write --section and set --section validate the resulting full document, not only the edited section. | cli/cmd/write.go; cli/cmd/set.go; cli/cmd/strict_component_docs_test.go |
| Check/verify gate | RunCheckV2 reports strict component issues during check, and verify rebuilds then runs the same canonical validation path. | cli/cmd/check_enhanced.go; cli/cmd/check_enhanced_test.go |
| Migration generator | strictMigrationComponentBody creates strict docs for legacy or empty components before writing content nodes. | cli/cmd/migrate_v2.go; cli/cmd/migrate_v2_test.go |
| Migration all-or-nothing | collectMigrateBlockers preflights strict component output and blocks grouped failures before migration writes. | TestRunMigrateV2_AggregatesStrictComponentBlockers; TestRunMigrateV2_JSONBlockerReport |
| Scaffold/template impact | Embedded component templates and help text must teach the strict sections so new docs do not start invalid. | cli/internal/templates/component.md; cli/cmd/help.go |
## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| Add component | New components must include strict sections before acceptance. | cli/cmd/add.go; cli/cmd/strict_component_docs.go |
| Write full body | Full component writes validate strict content before storing. | cli/cmd/write.go |
| Write or set section | Section updates validate the resulting full component document. | cli/cmd/write.go; cli/cmd/set.go |
| Check and verify | Structural validation reports missing or weak strict sections. | cli/cmd/check_enhanced.go |
| Migrate | Legacy component docs are converted into strict form or blocked with repair guidance. | cli/cmd/migrate_v2.go |
## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Keep only Goal plus Dependencies | Too little detail for LLM continuity and code derivation. |
| Make sections advisory | Agents would continue accepting superficial stubs. |
| Require prose only | Harder to skim, validate, migrate, and repair in batches. |
| Make every section bespoke | Too flexible; tool could not provide a reliable harness. |
## Migration Impact

Legacy component docs must be upgraded into strict form. Migration should preserve all-or-nothing behavior: if generated strict docs contain weak language or invalid tables, C3 reports blockers and avoids partial writes. Repair should be scoped to listed sections, then followed by cache clear, import, continue migration, check, and verify.

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Boilerplate passes as useful content | Require evidence, references, section-backed derivation, and duplicate-boilerplate rejection in semantic tightening. | adr-20260415-strict-component-docs-semantic-tightening |
| Strict docs slow small changes | Section writes remain focused, but they validate the resulting document. | go test ./... |
| Migration produces unsafe partial state | Preflight blockers before writing and stop before canonical export on write failure. | cli/cmd/migrate_v2_test.go |
| Underlay drift from ADR | Keep the Underlay C3 Changes table aligned with validator, mutation, and migration files whenever strict docs change. | c3x lookup cli/cmd/strict_component_docs.go; c3x check --include-adr |
## Verification

| Check | Command or file |
| --- | --- |
| Strict validator coverage | go test -count=1 ./cmd -run TestStrict |
| Component add behavior | cli/cmd/add_test.go |
| Component write behavior | cli/cmd/readwrite_test.go |
| Component set behavior | cli/cmd/set_test.go |
| Migration behavior | go test -count=1 ./cmd -run TestRunMigrateV2 |
| C3 canonical health | c3x check --include-adr; c3x verify |
