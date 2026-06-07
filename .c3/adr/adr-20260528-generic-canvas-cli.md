---
id: adr-20260528-generic-canvas-cli
c3-seal: e89aefbb21362ff69a796000637548a3cffa04db2d134de6592fec26a70fe1b1
title: generic-canvas-cli
type: adr
goal: Add a generic canvas definition surface to the CLI so C3 can govern knowledge graphs beyond software ADRs while keeping the primitive set lean and mechanically checkable.
status: proposed
date: "2026-05-28"
template: implementation-change
---

## Goal

Add a generic canvas definition surface to the CLI so C3 can govern knowledge graphs beyond software ADRs while keeping the primitive set lean and mechanically checkable.

## Context

The current ADR template work proves schema-driven authoring, but it is too ADR-specific for the broader goal. The target model is a canvas: users can define project-local knowledge graph contracts for design systems, PM requirements, PRDs, user stories, and software decisions without turning C3 into a broad workflow platform. The CLI needs a small place to list, read, add, and replace those canvas definitions before eval can ask agents to author documents against them.

## Decision

Introduce `c3x canvas list/read/add/write` as a generic definition manager. Canvas files live flat under `.c3/canvases/`, use sealed canonical markdown, and carry YAML section definitions using the existing section and column schema. Built-in canvases cover `c3-adr`, `atomic-design-change`, `pm-requirement`, `prd`, and `user-story`. The allowed column primitives stay small: text, date, enum, table sections, edge, cite, check, and entity_id.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-113 | component | Owns schema definitions and validation for the new canvas registry and primitive whitelist. | c3-113#n1827@v1:sha256:4518d11fd4ebabdef5c56416ed228b867c4847f82e97a390e91004803ef38602 "Provide durable agent-ready documentation for check-cmd so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verificatio" | Review primitive set and parser validation. |
| c3-117 | component | Owns docs-state commands; canvas is a docs-state definition command alongside read/write/schema/template. | c3-117#n1998@v1:sha256:3a377de9129c41b25fa893dc4431afb0aad2814a699e530c70d0de2567204ff6 "Read, write, set, validate schema, and report status for canonical C3 documents." | Review command behavior, output, tests, and code-map ownership. |
| c3-108 | component | Owns CLI dispatch, option parsing, help, and agent/human presentation touched by the new command. | c3-108#n1563@v1:sha256:ae80704ae7172ccccc82f6ba7b67f4fe434e41a8e3571e164a7b2165e4e4f06b "Provide CLI bootstrap, option parsing, output shaping, config resolution, and human/agent presentation helpers." | Review main dispatch and help addition. |
| c3-119 | component | Owns sync/export/repair behavior; project-local canvases must survive canonical export like ADR templates. | c3-119#n2089@v1:sha256:af616512ffa6834ca8959b4eb4a2e900bcc6b039bd34d4426d95e0bcca47ec59 "Handle import, export, sync, repair, migrate, delete, and git guardrail flows around canonical C3 state." | Review canvas preservation in sync. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| N.A - no existing ref governs generic canvas definitions. | N.A - this slice establishes the first generic canvas contract before extracting a reusable ref. | N.A - no existing ref governs this slice. | N.A - no ref update in this slice. |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| N.A - no existing rule governs generic canvas definitions. | N.A - enforcement is covered by parser validation, command tests, and local C3 checks. | N.A - no existing rule governs this slice. | N.A - no rule update in this slice. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Canvas schema | Add schema.Canvas, project canvas parsing, built-ins, and lean primitive validation. | go test ./cmd -run TestRunCanvas |
| Canvas command | Add c3x canvas list/read/add/write with sealed canonical markdown output. | go test ./cmd -run TestRunCanvas |
| CLI integration | Dispatch command, accept --file, help text, and mutation classification. | go test ./... from cli/ |
| Check and sync behavior | Preserve .c3/canvases/*.md across canonical export/check and fail check on invalid project canvas definitions. | TestRunSyncExport_PreservesCanvases and TestRunCheck_ValidatesProjectCanvases |
| Canonical rendering | Keep canvas frontmatter minimal by omitting empty title fields. | c3x canvas read prd |
| Code map | Map new canvas command files to c3-117 ownership. | c3x lookup cli/cmd/canvas.go |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| cli/internal/schema/canvas.go | New generic canvas registry, document parser, built-ins, and primitive whitelist. | go test ./cmd -run TestRunCanvas |
| cli/cmd/canvas.go | New canvas CLI command with list/read/add/write. | go test ./cmd -run TestRunCanvas |
| cli/main.go and cli/cmd/help.go | Command dispatch, file support, mutation classification, and help. | go test ./... |
| cli/cmd/document_format.go | Canvas canonical docs omit empty title frontmatter so builtin canvas output can round-trip through add/write. | c3x canvas read prd |
| cli/cmd/check_enhanced.go | Load and validate project canvases during c3x check. | TestRunCheck_ValidatesProjectCanvases |
| cli/cmd/sync.go | Preserve project-local canvas files during sync export/check. | TestRunSyncExport_PreservesCanvases |
| .c3/code-map.yaml | Assign canvas command files to c3-117. | c3x lookup cli/cmd/canvas.go |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| c3x canvas list | Shows built-in and project canvases for software, design, PM, PRD, and story use cases. | C3X_MODE=agent bash skills/c3/bin/c3x.sh canvas list |
| c3x canvas read | Emits sealed YAML-compatible canvas definitions with lean primitives. | C3X_MODE=agent bash skills/c3/bin/c3x.sh canvas read prd |
| c3x canvas add/write | Rejects unsupported primitive types and built-in id collisions. | TestRunCanvasAdd_RejectsUnsupportedPrimitive |
| c3x check | Ensures canonical files, code map, and seals remain coherent. | C3X_MODE=agent bash skills/c3/bin/c3x.sh check |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Keep extending ADR templates. | That preserves an ADR-specific mental model and misses design/PM/PRD/story canvases. |
| Add domain-specific commands for design, PRD, and story. | That expands primitives and command surface before eval proves need. |
| Allow arbitrary column validators now. | User explicitly wants Docker Compose-level flexibility, not Kubernetes-level policy machinery. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Canvas becomes a renamed template system. | Command and file type are generic canvas, with built-ins spanning non-software domains and no ADR-only assumptions. | c3x canvas list includes target use cases. |
| Primitive set grows too quickly. | Parser only accepts the agreed lean set plus enum values and edge target syntax. | Unsupported primitive test rejects script. |
| Eval still cannot prove 90 percent success. | This slice provides authoring contracts; the next slice must wire eval cases against canvas read output. | Subagent review tracks eval readiness as next work. |

## Verification

| Check | Result |
| --- | --- |
| focused canvas command tests | pass |
| project canvas check validation test | pass |
| go test ./... from cli/ | pass |
| bash scripts/build.sh | pass |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh canvas list | pass |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh canvas read prd | pass |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh lookup cli/cmd/canvas.go | pass |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check | pass |
