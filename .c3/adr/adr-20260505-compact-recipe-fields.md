---
id: adr-20260505-compact-recipe-fields
c3-seal: 01c5f161a8fa1bccb074c68e89881fc866c8521a9c81d0491828aee9d8e1a53c
title: compact-recipe-fields
type: adr
goal: Expose recipe-only `description` and `sources` in compact/agent `c3x list` output so agents can match recipe shortcuts and see their trace sources without immediately switching to full structured output.
status: implemented
date: "2026-05-05"
---

## Goal

Expose recipe-only `description` and `sources` in compact/agent `c3x list` output so agents can match recipe shortcuts and see their trace sources without immediately switching to full structured output.

## Context

`recipe-validation-system` now gives the validation system a shortcut, but compact/agent list output only includes `id`, `type`, `title`, `goal`, `parent`, and `status`. The query workflow says agents should match recipes by title and description, then read the recipe and serve its sources as a trace. Without compact `description` and `sources`, recipe discovery still has a gap: agents can see the recipe row but cannot cheaply judge source coverage or description match from the default agent topology.

## Decision

Update `c3-112` list-cmd compact projection to include `description` and `sources` only for recipe entities. Keep non-recipe compact rows unchanged to preserve bounded agent output. Add focused tests around recipe compact output so the shortcut remains mechanically visible.

## Affected Topology

| Entity | Type | Why affected | Governance review |
| --- | --- | --- | --- |
| c3-112 | component | Owns c3x list compact/agent topology projection changed by this ADR. | Verify compact output remains bounded and recipe-only fields do not alter other entity rows. |
| c3-1 | container | Parent Go CLI output contract changes within one command component. | Parent delta reviewed as no new component; responsibility remains list output. |
| c3-101 | component | Recipe description comes from frontmatter/metadata already parsed by this component. | Review metadata semantics stay read-only from list-cmd. |
| c3-102 | component | Recipe entities discovered by walker remain list input. | Review no walker behavior changes are needed. |

## Compliance Refs

| Ref | Why required | Action |
| --- | --- | --- |
| ref-frontmatter-docs | Recipe description and sources are frontmatter-derived fields projected by list-cmd. | comply |

## Compliance Rules

| Rule | Why required | Action |
| --- | --- | --- |
| N.A - no rule entity currently governs compact list output shape | N.A - no rule entity currently governs compact list output shape | N.A - no rule entity currently governs compact list output shape |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| cli/cmd/list.go | Add optional compact recipe fields and populate them only when e.Type == "recipe". | c3x lookup cli/cmd/list.go; go test ./cmd -run TestListTopology_RecipesCompact |
| cli/cmd/list_test.go | Extend compact recipe tests to assert description and sources are present for recipe discovery. | c3x lookup cli/cmd/list_test.go; focused Go test |
| ADR lifecycle | Create and implement this work order with verification evidence. | c3x check --include-adr --only adr-20260505-compact-recipe-fields |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| list-cmd compact output | Add recipe-only compact fields, preserving bounded non-recipe rows. | C3X_MODE=agent bash skills/c3/bin/c3x.sh list --compact |
| list-cmd tests | Add/extend tests proving compact recipe discovery fields. | cd cli && go test ./cmd -run TestListTopology_RecipesCompact and cd cli && go test ./cmd -run TestRunListStructuredAgentCompactIncludesRecipeDiscoveryFields |
| Autoresearch benchmark | Score compact recipe description/sources, C3 check, and focused Go tests. | bash .autoresearch/sessions/recipe-compact-fields/benchmark.sh |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| TestListTopology_RecipesCompact | Human compact recipe output exposes discovery source detail. | go test ./cmd -run TestListTopology_RecipesCompact |
| TestRunListStructuredAgentCompactIncludesRecipeDiscoveryFields | Agent compact structured output exposes recipe fields. | go test ./cmd -run TestRunListStructuredAgentCompactIncludesRecipeDiscoveryFields |
| c3x check | Verifies C3 docs remain structurally valid. | C3X_MODE=agent bash skills/c3/bin/c3x.sh check |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Add a new c3x recipes command | More command surface than needed; query workflow already starts from c3x list. |
| Put sources on every compact entity | Would inflate agent output and duplicate full structured output for non-recipe rows. |
| Require agents to always run list --json | Defeats agent-mode compact default and increases token pressure before a recipe match is proven. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Compact output grows too much | Limit new fields to recipe entities only. | Tests inspect recipe behavior and existing compact rows remain unchanged by construction. |
| Description leaks unbounded text | Reuse the existing short goal limit for compact description. | Focused tests assert expected compact description text. |
| Sources drift from recipe frontmatter | Continue using existing relationships from store and C3 check. | c3x check validates source references resolve. |

## Verification

| Check | Result |
| --- | --- |
| bash .autoresearch/sessions/recipe-compact-fields/benchmark.sh | Passed: compact_recipe_fields_score 5, recipe_hit 1, description_hit 1, sources_hit 1, c3_check_pass 1, go_test_pass 1. |
| cd cli && go test ./cmd -run TestListTopology_RecipesCompact | Passed. |
| cd cli && go test ./cmd -run TestRunListStructuredAgentCompactIncludesRecipeDiscoveryFields | Passed. |
| cd cli && go test ./... | Passed. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh list --compact | Passed: compact recipe row includes description and sources with c3-113. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check | Passed: total 66, no issues. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr --only adr-20260505-compact-recipe-fields | Passed: total 66, no issues. |
