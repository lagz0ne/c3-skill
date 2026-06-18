---
id: adr-20260617-update-frontmatter-goal-html-comments
c3-seal: 24a2c3af9f5b99f132b0fae2f9950d04b3304d0b4a43fbf811caafc4ca2e396a
title: update-frontmatter-goal-html-comments
type: adr
goal: 'Update component c3-101 (`frontmatter`)''s `## Goal` section so it states that the component strips HTML comments, in addition to parsing and writing YAML frontmatter embedded in `.c3/` markdown files. The change is documentation-only: it edits the Goal block of one frozen component fact to reflect the requested behavior.'
status: proposed
date: "2026-06-17"
---

# ADR: update-frontmatter-goal-html-comments

## Goal

Update component c3-101 (`frontmatter`)'s `## Goal` section so it states that the component strips HTML comments, in addition to parsing and writing YAML frontmatter embedded in `.c3/` markdown files. The change is documentation-only: it edits the Goal block of one frozen component fact to reflect the requested behavior.

## Context

c3-101's current Goal reads "Parse and write YAML frontmatter embedded in `.c3/` markdown files." The request is to extend it to also mention HTML-comment stripping. c3-101 is a frozen fact (a component), so its Goal block can only change through a change-unit patch landed by `change apply`. The affected topology is limited to c3-101 and its parent container c3-1 (Go CLI), whose `## Components` row carries a one-line goal contribution for c3-101.

NOTE FOR REVIEWER (code-check): A grep of `cli/internal/frontmatter/` (c3-101's mapped code) found no HTML-comment handling — the `frontmatter` package neither strips nor mentions HTML comments. HTML-comment logic lives in a different package, `cli/internal/markdown/`, whose test asserts comments are *preserved*, not stripped (`markdown_test.go`: "HTML comments should be preserved in section content"). The requested Goal text therefore does not match c3-101's current code. This ADR is authored to carry the requested doc edit through the correct change-unit flow; the doc/code mismatch is flagged here so it is resolved (correct the component, correct the code, or redirect the edit to the markdown component) before the unit is accepted and landed.

## Decision

Author a single block patch against c3-101's Goal block that replaces it with text covering both YAML frontmatter parsing/writing and HTML-comment stripping, anchored by the Goal block's current cite handle. Land it via `c3 change apply` after the doc/code mismatch noted in Context is resolved. Use a block patch (the only legal edit to a live fact); do not use `set`/`write` (refused on a fact) and do not use `whole`-with-base or `insert` (rejected/unimplemented).

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-101 | component | Its Goal block is the target of the edit. | c3-101#n1956@v1:sha256:c4b27cf43e62def3c641ca7fc35ac245da11668f2f6efe6fa788ba3e171d4048 "Parse and write YAML frontmatter embedded in .c3/ markdown files." | Parent Delta against c3-1 Components/Responsibilities. |
| c3-1 | container | Parent of c3-101; its ## Components row for c3-101 states a one-line goal contribution that may need to match. | c3-1#n1920@v1:sha256:b31de913a9eaa864d4de5f10894602485b8bff069d1c91f691bff809ce59c5b9 "Provide all c3x commands as a single cross-compiled binary that reads and writes .c3/ architecture docs." | Decide updated vs no-delta with evidence. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-frontmatter-docs | c3-101 uses this ref; it governs the frontmatter+body doc contract the component implements. | ref-frontmatter-docs#n3410@v1:sha256:e45eb7cfd2d112f2acc1a6e499d5f3906a484e958430b2277be0b8071d142909 "All .c3/ architecture docs use YAML frontmatter (between --- delimiters) followed by a Markdown body." | review — confirm the new Goal does not contradict the frontmatter/body split. |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| N.A - no rule-* surfaced by c3 lookup on cli/internal/frontmatter | No coding rule governs this doc-only Goal edit. | c3 lookup cli/internal/frontmatter (refs only, no rules) | N.A |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Patch | Author 01-goal-html-comments.patch.md in .c3/changes/<adr-id>/, scope block, base = c3-101 Goal cite handle, body = new Goal text. | patch file |
| Land | c3 change accept then c3 change apply then c3 check. | change-unit flow |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| N.A - no CLI surface changes | This ADR edits a component doc Goal block only; no command, validator, schema, hint, template, or test changes. | git diff shows only .c3/ doc + change-unit files |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| c3 check | Validates c3-101 against the component canvas after apply; non-empty Goal section required. | c3 check post-land |
| c3 change apply drift+canvas gates | Rejects the patch if the Goal block drifted or the merged body breaks the component canvas. | apply preflight |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| c3 set c3-101 goal "..." | Refused — c3-101 is a frozen fact; set returns "c3-101 is a fact — facts are frozen and change only through a change-unit". |
| whole-with-base patch | Rejected by the apply gate — a live fact edit must be block-anchored, not full-replace. |
| Editing the markdown component instead | Out of scope of the request, which names c3-101; flagged in Context as the likely true home of HTML-comment behavior. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Goal claims behavior the code does not have. | Context flags the mismatch; resolve before accept/apply. | grep cli/internal/frontmatter for comment handling (currently none). |
| Parent container row drifts from the new Goal. | Decide Parent Delta (updated vs no-delta) before completing the unit. | c3 read c3-1 Components row for c3-101. |

## Verification

| Check | Result |
| --- | --- |
| c3 change apply <adr-id> --dry-run | Preflight passes (drift + canvas) before land. |
| c3 check | c3-101 valid against component canvas after land. |
| grep -niE 'comment' cli/internal/frontmatter/frontmatter.go | Must return a match before the Goal can truthfully claim stripping (currently returns nothing). |
