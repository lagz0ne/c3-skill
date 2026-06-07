---
id: adr-20260528-read-cite-evidence-handles
c3-seal: f5e9e73487cefbebc04e2eb7cadfb481e8a32abacc91ef63f022c3a00ff24a5f
title: read-cite-evidence-handles
type: adr
goal: Add a compact `c3x read --cite` path that lets agents obtain canonical evidence handles from existing C3 documents before filling ADR and canvas/change-record evidence fields.
status: implemented
date: "2026-05-28"
template: implementation-change
---

## Goal

Add a compact `c3x read --cite` path that lets agents obtain canonical evidence handles from existing C3 documents before filling ADR and canvas/change-record evidence fields.

## Context

The canvas model needs source-bound evidence so agents cannot claim context they did not read, cite nonexistent docs, or silently reuse stale context. The current CLI already stores node hashes and entity versions, and `c3x read` owns document/section reads, but it does not emit copyable evidence handles. The first CLI slice should use the existing read surface rather than adding command sprawl.

## Decision

Extend `c3x read` with `--cite`. Full-entity reads append an entity citation using the entity version and root merkle. Section reads append node-level citations for the smallest meaningful body nodes in that section, using the existing node tree, entity version, node hash, and an exact source snippet. Keep normal read output intact and append citations plus existing agent hints.

## Affected Topology

| Entity | Type | Why affected | Governance review |
| --- | --- | --- | --- |
| c3-117 | component | read command behavior and docs-state command tests change to output evidence handles for selected docs/sections. | Review docs-state command contract and keep read/write/schema behavior deterministic. |
| c3-108 | component | Runtime option parsing and dispatcher wiring gain the --cite flag without adding a new command or output mode. | Review option parsing and help output so human and agent modes stay compact. |
| c3-106 | component | Citation output reads node-tree hashes for section body nodes produced by content parsing. | Review content node assumptions; avoid changing content storage semantics. |

## Compliance Refs

| Ref | Why required | Action |
| --- | --- | --- |
| ref-frontmatter-docs | read --cite depends on canonical C3 docs remaining frontmatter-backed markdown entities with stable IDs, versions, and node content. | comply |

## Compliance Rules

| Rule | Why required | Action |
| --- | --- | --- |
| N.A - no current rule governs read citation handles | The repository has no existing rule document for evidence-handle formatting or citation generation. | create-rule later if evidence policy becomes a reusable strict rule |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| CLI options | Add --cite to parsed options and route it through main.go into ReadOptions. | cli/cmd/options.go; cli/main.go; options tests |
| Read command | Append entity or section node citations while preserving existing read output and agent hints. | cli/cmd/read.go; readwrite tests |
| Help surface | Document --cite under c3x read help without adding a separate evidence command. | cli/cmd/help.go; help tests |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| read command | Add citation formatting for entity and section reads, backed by entity version/root merkle and node hashes. | go test ./cmd -run TestRunRead_ |
| runtime dispatch | Parse and pass --cite through the existing read command path. | go test ./cmd -run TestParseArgs; go test . -run TestRun_Read |
| help/tests | Update read help assertions and add citation output coverage. | go test ./cmd |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| c3x read <id> --cite | Appends citation: <entity>@v<version>:sha256:<root> to normal entity read output. | read command tests |
| c3x read <id> --section <name> --cite | Appends citations: rows for cited body nodes under the section. | read command tests |
| Option parser | Accepts --cite only as a flag and does not require a value. | options tests |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Add c3x evidence command | It increases API surface before evidence validation exists; current direction prefers compact read/check APIs. |
| Put citations in JSON-only output | Agents and humans both need copyable citation handles in normal read output; JSON-only would create a second workflow. |
| Handcraft evidence handles in ADRs | It does not force a read and allows bogus node/version/hash content. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Citation drift | Include entity version and node hash in every citation so later check work can distinguish stale from invalid evidence. | Unit tests assert version/hash formatting. |
| Noisy section output | Emit citations only after section content and only for body/table/list/code nodes, not headings alone. | Section citation tests cover concise output. |
| Existing read regressions | Preserve current read output before appended citation block and keep existing agent hints. | Existing read tests plus new cite tests. |

## Verification

| Check | Result |
| --- | --- |
| cd cli && go test ./cmd -run TestRunRead | passed |
| cd cli && go test ./cmd -run TestParseArgs | passed |
| cd cli && go test ./cmd -run TestShowHelp_Read | passed |
| cd cli && go test ./... | passed |
| bash scripts/build.sh | passed |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh read c3-117 --section Goal --cite | passed |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr --only adr-20260528-read-cite-evidence-handles | passed |
