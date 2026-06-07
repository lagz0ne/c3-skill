---
id: adr-20260514-managed-adr-templates
c3-seal: 3a2159a5d853af345d8af257a0dc6269fb7103c764284f0416ae4b6a047a79de
title: managed-adr-templates
type: adr
goal: Add a minimal c3x-managed ADR template path where project templates use tiny frontmatter metadata, are sealed as canonical C3 files, and can be selected by schema/add without expanding lifecycle or version policy.
status: proposed
date: "2026-05-14"
---

## Goal

Add a minimal c3x-managed ADR template path where project templates use tiny frontmatter metadata, are sealed as canonical C3 files, and can be selected by schema/add without expanding lifecycle or version policy.

## Context

The current spike exposes a built-in ADR template seam in Go, but project-specific templates are not yet managed by c3x. The desired template metadata is intentionally small: id, type, and description only. The existing canonical seal system applies to markdown files, so project ADR templates should use sealed markdown with a YAML contract body rather than a separate unsealed YAML side channel.

## Decision

Implement project ADR templates as sealed `.c3/adr-templates/<id>.md` files with frontmatter fields `id`, `type: adr-template`, and `description`. The body stores the template contract. Add CLI paths to list/read/add templates, select templates for `schema adr --template`, and stamp ADRs created with `add adr --template`.

## Affected Topology

| Entity | Type | Why affected | Governance review |
| --- | --- | --- | --- |
| c3-103 | component | Template scaffolding and embedded template policy are extended by project-managed ADR templates. | Review embedded-template ref and template ownership. |
| c3-111 | component | add-cmd must accept template selection for ADR creation and template file creation. | Preserve add validation and all-or-nothing behavior. |
| c3-113 | component | check-cmd/schema validation must resolve template contracts consistently. | Preserve schema-driven validation consumers. |
| c3-117 | component | docs-state-cmds exposes schema output for selected ADR templates. | Keep schema output compact and template-derived. |
| c3-119 | component | sync/import/export/repair must preserve and seal template files. | Preserve canonical seal and cache repair contracts. |

## Compliance Refs

| Ref | Why required | Action |
| --- | --- | --- |
| ref-embedded-templates | Template behavior must stay compatible with embedded built-ins while allowing project-managed overrides. | review |

## Compliance Rules

| Rule | Why required | Action |
| --- | --- | --- |
| N.A - no rule surfaced by lookup for the touched files. | N.A - current docs cite parent policy/ref only. | N.A - no rule action. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Template model | Parse sealed project ADR template markdown with minimal frontmatter and YAML body contract. | go test ./cmd ./internal/schema. |
| CLI commands | Add template list/read/add and --template selection for schema/add. | go test ./cmd. |
| Sync/seal | Preserve adr-template files during sync export and include seals in verification. | go test ./cmd. |
| ADR creation | Store selected template in ADR metadata/frontmatter and validate body against selected template. | go test ./cmd. |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| cli/internal/schema | Add project template parsing/loading and resolver. | Pending. |
| cli/cmd | Add template command and selected-template schema/add paths. | Pending. |
| sync/repair | Preserve sealed template files during export/check. | Pending. |
| tests | Cover minimal metadata, list/read/add/schema, seal preservation, and selected validation. | Pending. |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| c3x template add/list/read | Manages sealed project ADR template docs. | Pending. |
| c3x schema adr --template | Shows selected template description and sections. | Pending. |
| c3x add adr --template | Validates and stamps ADR with selected template. | Pending. |
| c3x check/repair | Verifies seals and preserves project templates. | Pending. |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Raw .yaml templates | Existing seal system covers canonical markdown docs; raw YAML would require a second seal path now. |
| Rich frontmatter metadata | User asked to keep metadata minimal and put prose/contract in body. |
| Template status/version fields | Too much lifecycle for current need; id/type/description is enough. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Sync export deletes template files because they are not DB entities | Preserve adr-template paths during sync export/check diff. | go test ./cmd. |
| Template body schema drifts from validator expectations | Parse into existing SectionDef/RejectRules and validate required fields. | go test ./internal/schema. |
| Existing ADR behavior breaks | Keep built-in default and run full Go tests. | go test ./... |

## Verification

| Check | Result |
| --- | --- |
| go test ./cmd ./internal/schema | Passed. |
| go test ./... | Passed. |
| go run . --c3-dir ../.c3 template list | Passed; output lists built-in implementation-change template and description. |
| go run . --c3-dir ../.c3 schema adr --template implementation-change | Passed; output shows selected template metadata. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr --only adr-20260514-managed-adr-templates | Passed. |
