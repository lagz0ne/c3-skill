---
id: adr-20260505-validation-recipe-shortcut
c3-seal: d4232746727d110c679a29dde4b8472962d17c93153a74042677941bf4c84e1a
title: validation-recipe-shortcut
type: adr
goal: Add a short-lived C3 recipe that gives agents a compact discovery shortcut for the current validation system without changing validation behavior.
status: implemented
date: "2026-05-05"
---

## Goal

Add a short-lived C3 recipe that gives agents a compact discovery shortcut for the current validation system without changing validation behavior.

## Context

The current topology has recipe support in the CLI but no recipe rows in this repository, so the query workflow cannot use the recipe fast path. The validation/check system is a high-pressure component because `c3-113` spans structural checks, schema checks, ADR coverage, recipe source validation, strict component docs, and supporting frontmatter/walker behavior. A recipe can temporarily gather the validation path so agents can start from one shortcut, then verify against the source entities to avoid stale shortcut drift.

## Decision

Create `recipe-validation-system` with a concise goal and `sources` relationships to the current validation owner and directly relevant supporting components. Keep this additive and documentation-only for this run. The recipe is intentionally a shortcut, not canonical truth; agents must read cited source entities before changing code.

## Affected Topology

| Entity | Type | Why affected | Governance review |
| --- | --- | --- | --- |
| c3-113 | component | Primary validation/check owner cited by the recipe for current validation behavior. | Review recipe source points at the validation owner before relying on it. |
| c3-101 | component | Frontmatter parsing governs fields such as sources and recipe classification. | Review because recipe discovery depends on frontmatter fields remaining synchronized. |
| c3-102 | component | Walker discovers recipe files under .c3/recipes. | Review because recipe discovery depends on walker coverage. |
| c3-103 | component | Embedded recipe template defines new recipe shape. | Review because recipe authoring uses this template and schema. |
| c3-104 | component | Wiring stores sources relationships used by recipe trace validation. | Review because stale shortcut risk is controlled by relationship checks. |

## Compliance Refs

| Ref | Why required | Action |
| --- | --- | --- |
| ref-frontmatter-docs | Recipe metadata and sources are frontmatter-governed C3 docs. | comply |

## Compliance Rules

| Rule | Why required | Action |
| --- | --- | --- |
| N.A - no rule entity currently governs recipe authoring in this topology | N.A - no rule entity currently governs recipe authoring in this topology | N.A - no rule entity currently governs recipe authoring in this topology |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| ADR | Create adr-20260505-validation-recipe-shortcut as the work order. | c3x add adr validation-recipe-shortcut --file /tmp/adr-validation-recipe-shortcut.md |
| Recipe | Create recipe-validation-system with a compact validation discovery goal. | c3x add recipe validation-system --file /tmp/recipe-validation-system.md |
| Sources | Write recipe frontmatter with sources pointing to c3-113, c3-101, c3-102, c3-103, and c3-104. | c3x write recipe-validation-system --file /tmp/recipe-validation-system.md |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| Recipe docs | Add one recipe entity through local C3 CLI; no Go CLI behavior changed. | c3x list --compact shows recipe-validation-system. |
| Relationship validation | Use existing sources relationship sync and recipe source check. | c3x check catches unresolved recipe source drift. |
| Autoresearch benchmark | Use .autoresearch/sessions/recipe-validation-shortcut/benchmark.sh to score recipe discovery. | bash .autoresearch/sessions/recipe-validation-shortcut/benchmark.sh |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| c3x check | Validates recipe source relationships resolve. | C3X_MODE=agent bash skills/c3/bin/c3x.sh check |
| c3x list --compact | Makes recipe ID and goal visible during query topology discovery. | C3X_MODE=agent bash skills/c3/bin/c3x.sh list --compact |
| Autoresearch benchmark | Fails to full score if recipe, validation text, sources, or check pass are missing. | .autoresearch/sessions/recipe-validation-shortcut/benchmark.sh |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Split c3-113 immediately | The user asked to test recipe pressure relief first; splitting validation ownership is a larger design change. |
| Change compact list output before adding a recipe | Without a real recipe in this repo, a projection change has no local validation-system shortcut to evaluate. |
| Keep relying on component topology only | It leaves agents opening broad validation surfaces instead of starting from one scoped trace. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Recipe drifts from source docs | Treat recipe as a shortcut and cite source entities through sources. | c3x check validates source IDs and agents read cited entities before changing code. |
| Compact list still lacks enough detail | Keep the experiment scoped, then decide whether a second CLI run should add recipe-only compact fields. | Autoresearch score and sidecar recommendation identify projection gap. |
| ADR/doc-only change hides needed code change | Stop after benchmark and report whether list/query projection needs a follow-up experiment. | Final no-slop, simplify, and review pass. |

## Verification

| Check | Result |
| --- | --- |
| bash .autoresearch/sessions/recipe-validation-shortcut/benchmark.sh | Passed: recipe_discovery_score 4, recipe_hits 1, validation_hits 1, source_hits 1, c3_check_pass 1. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh list --compact | Passed: compact topology includes recipe-validation-system and validation goal text. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh list --json | Passed: full structured topology includes recipe-validation-system with sources [c3-101 c3-102 c3-103 c3-104 c3-113]. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr --only adr-20260505-validation-recipe-shortcut | Passed: total 65, no issues. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check | Passed: total 65, no issues. |
