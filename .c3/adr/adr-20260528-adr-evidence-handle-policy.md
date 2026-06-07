---
id: adr-20260528-adr-evidence-handle-policy
c3-seal: 54a4304ac66627eda53bef98e9de955374ed82075d36ae852908ae55d9e69cce
title: adr-evidence-handle-policy
type: adr
goal: Require ADR governance rows to cite current C3 content with mechanically verifiable evidence handles so release-targeted documents cannot misquote, reference missing docs, or silently drift from the graph.
status: proposed
date: "2026-05-28"
template: implementation-change
---

## Goal

Require ADR governance rows to cite current C3 content with mechanically verifiable evidence handles so release-targeted documents cannot misquote, reference missing docs, or silently drift from the graph.

## Context

The canvas work depends on compact declarations whose output can still be verified by tooling. `c3x read --cite` now emits node handles, but ADR tables still accepted plain prose in Affected Topology, Compliance Refs, and Compliance Rules. That left the exact user failure mode open: an LLM could name a doc that does not exist, quote content that changed, or skip reading the referenced source.

## Decision

Make Evidence a required cite column in the ADR governance tables, and make ADR validation parse the handle against the current store. Existing targets must cite the same entity, current version, existing node id, matching node hash, and exact snippet. Rows for to-be-created refs or rules may use `N.A - <reason>` evidence only when the action is a create action.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-113 | component | Owns schema registry and check validation that now enforce ADR evidence handles. | c3-113#n1403@v1:sha256:4518d11fd4ebabdef5c56416ed228b867c4847f82e97a390e91004803ef38602 "Provide durable agent-ready documentation for check-cmd so generated code, tests, and follow-up docs preserve ownership, boundaries, governance, and verificatio" | Review schema, add/check validation, and focused tests. |
| c3-117 | component | Owns docs-state commands and the read --cite output contract consumed by Evidence rows. | c3-117#n1574@v1:sha256:3a377de9129c41b25fa893dc4431afb0aad2814a699e530c70d0de2567204ff6 "Read, write, set, validate schema, and report status for canonical C3 documents." | Confirm schema/read contract stays usable for agents. |
| c3-108 | component | Test fixture helper gained citation generation support for CLI validation tests. | c3-108#n1139@v1:sha256:ae80704ae7172ccccc82f6ba7b67f4fe434e41a8e3571e164a7b2165e4e4f06b "Provide CLI bootstrap, option parsing, output shaping, config resolution, and human/agent presentation helpers." | Review fixture-only helper impact. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| N.A - no existing ref governs ADR evidence handles. | N.A - this slice promotes a CLI validation contract before extracting a cross-domain ref. | N.A - no existing ref governs this slice. | N.A - no ref update in this slice. |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| N.A - no existing rule governs ADR evidence handles. | N.A - enforcement is introduced by tests and check validation in this slice. | N.A - no existing rule governs this slice. | N.A - no rule update in this slice. |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| ADR schema | Add Evidence cite columns to Affected Topology, Compliance Refs, and Compliance Rules. | focused ADR/add/check/schema tests |
| ADR validation | Validate evidence handles against entity id, node id, current version, node hash, and snippet. | go test ./... from cli/ |
| Test fixtures | Add a helper that generates valid in-memory evidence handles for CLI tests. | go test ./cmd -run TestValidateADREvidence |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| cli/internal/schema/schema.go | ADR registry includes Evidence cite columns and fill guidance for governance rows. | Focused schema/add tests pass. |
| cli/cmd/adr_linkage.go | ADR linkage validation parses citation handles and rejects missing, stale, wrong-entity, wrong-node, wrong-hash, or mismatched snippets. | go test ./cmd -run TestValidateADREvidence |
| cli/cmd/add_test.go and cli/cmd/check_enhanced_test.go | ADR fixtures now use the new Evidence columns and assert validator behavior. | focused ADR/add/check tests |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| c3x schema adr | Shows Evidence cite columns before agents draft ADR governance rows. | C3X_MODE=agent bash skills/c3/bin/c3x.sh schema adr |
| c3x add adr | Rejects governance tables missing Evidence columns and rejects invalid authored evidence. | Focused add tests pass. |
| c3x check --include-adr | Warns when active ADR governance rows do not cite current graph content. | Focused check tests pass. |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Keep evidence as prose. | Prose cannot prove the doc exists, the quote matches, or the cited source was current. |
| Add many domain-specific validators now. | The canvas goal is lean; a single cite contract carries current software/design/PM examples better than per-domain ceremony. |
| Require historical version lookup now. | Current release validation needs the latest graph to match; historical cite replay can be added later as verify <cite>. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Existing ADR fixtures break because table columns changed. | Update fixtures to include Evidence and generate valid citations where rows target real docs. | focused ADR/add/check tests |
| Agents paste stale handles after editing docs. | Check compares current entity version and node hash. | TestValidateADREvidence_RejectsWrongEntityAndStaleHash |
| Future PM/PRD traces need typed relationships, not just point citations. | Keep this slice as point evidence; next CLI slice can add compact edge/verify semantics. | Subagent null-hypothesis review recorded this as a later need. |

## Verification

| Check | Result |
| --- | --- |
| focused ADR/add/check/schema tests | pass |
| go test ./... from cli/ | pass |
| bash scripts/build.sh | pass |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --only adr-20260528-adr-evidence-handle-policy --include-adr | pass |
