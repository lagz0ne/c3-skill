---
id: adr-20260610-force-toon-default
c3-seal: 5c7c3ef9f78a1b2b19a48ce55c2d28875a6af222246ef7d92b349862bf37cdf7
title: force-toon-default
type: adr
goal: Force TOON to be the default structured CLI output format and sunset JSON as the agent-facing default for now, including current `show`/marketplace-facing copy that presents JSON as the machine-friendly path.
status: implemented
date: "2026-06-10"
---

## Goal

Force TOON to be the default structured CLI output format and sunset JSON as the agent-facing default for now, including current `show`/marketplace-facing copy that presents JSON as the machine-friendly path.

## Context

The runtime contract already says agent-mode structured output must be TOON, but the implementation and help text still carry legacy JSON framing. `ResolveFormat(false, false)` returns `FormatHuman`, and the shared structured writers treat non-TOON formats as JSON, so structured-capable commands can default to JSON-shaped output instead of TOON. `marketplace show/list` also sits outside normal `.c3` dispatch, so its output and help need an explicit review while this contract changes. The affected topology is the Go CLI runtime output layer and the marketplace/show command family.

## Decision

Make TOON the default structured output format in the shared runtime helpers. Keep explicit `--json` available only as non-agent compatibility where command tests still prove it, but do not present JSON as the agent/default path. Update focused tests so default structured output is TOON, explicit non-agent JSON remains parseable, and agent mode still overrides explicit JSON. Update help/rule text that still says JSON is the default or best machine-readable form.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-1 | container | The Go CLI owns command output behavior and the shared rules governing agent versus human command surfaces. | c3-1#n1700@v1:sha256:bc556f922698f71ff7228c639417b2971e01a6b8711748ef052c19909515250d "Validate structural integrity of the doc tree" | Confirm this is a runtime/output contract change only, not a container topology change. |
| c3-108 | component | Runtime support owns output format resolution, agent-mode detection, and shared TOON/JSON writers. | c3-108#n2042@v1:sha256:ae80704ae7172ccccc82f6ba7b67f4fe434e41a8e3571e164a7b2165e4e4f06b "Provide CLI bootstrap, option parsing, output shaping, config resolution, and human/agent presentation helpers." | Update implementation/tests and rule golden example to align with TOON default. |
| c3-112 | component | List command routing needs to preserve human topology output while still honoring TOON for agent/structured output. | c3-112#n2260@v1:sha256:4ca566268fd3cda084d28e199c74bf84e62f43245ceb8e184ba9f7a3decfb2e6 "topology in human-readable or machine-readable formats." | Verify list human, explicit JSON, and agent TOON tests. |
| c3-115 | component | Codemap command had a hidden default JSON path and needs to move to TOON by default with explicit JSON compatibility. | c3-115#n2396@v1:sha256:fbf4d1d7c70602570a656b33759806e718c55bb828332a8617ec5aa316c19043 "empty stubs for all entities." | Verify codemap default TOON, explicit JSON, and legacy HUMAN=1 behavior. |
| c3-120 | component | Marketplace/show commands are the surfaced show path and must not keep advertising or depending on JSON as the agent-friendly default. | c3-120#n2632@v1:sha256:bf9271b0adc789557a88a4da5abbac12be6bec496c2140cb1af60d6aa27b7b50 "Handle versions, hash, nodes, prune, and marketplace command families." | Review marketplace list/show tests and help copy while preserving rule preview content. |

## Compliance Refs

| Ref | Why required | Evidence | Action |
| --- | --- | --- | --- |
| ref-cross-compiled-binary | Inherited Go CLI distribution governance; this output change must not alter npm wrapper or release binary behavior already present in the worktree. | ref-cross-compiled-binary#n3157@v1:sha256:d4e8aff3ebc7fabf348e6da10383569e74e9cafae95aba2c02a73f285f7e30be "Publish two install paths with one default." | review |
| ref-embedded-templates | Inherited Go CLI governance because runtime/help output changes must not alter scaffold template embedding. | ref-embedded-templates#n3167@v1:sha256:db227a0598059041ce49f2746f25fc8501e6ae86b9c9f688f593279ecde35ff8 "Doc templates are bundled in the CLI binary so scaffolding works without external files at any install path." | review |
| ref-frontmatter-docs | Inherited Go CLI governance because read/schema/check command output changes must not alter frontmatter/body handling. | ref-frontmatter-docs#n3182@v1:sha256:2113120972193c5dccc7b63aec1ce6de17e31f2aa403b0262fa3fee03501307f "YAML frontmatter for machine-readable metadata" | review |

## Compliance Rules

| Rule | Why required | Evidence | Action |
| --- | --- | --- | --- |
| rule-output-via-helpers | This is the controlling rule for structured output format resolution and it contained the old FormatHuman default example. | rule-output-via-helpers#n3633@v2:sha256:ca4a0c295ace95c5e32a1c7ca5b92a583f8637573bdec86d8513c1c6d7d486be "Commands must emit results via" | comply and update-rule |
| rule-wrap-error-cause | Runtime and marketplace errors touched during the change must continue wrapping causes across layer boundaries. | rule-wrap-error-cause#n3236@v1:sha256:20a5bd788231e5b7b7403d387c6414f0d5b8b31303d720d83369abbe18c9ab26 "Every returned error in the Go CLI preserves its cause and context so failures stay traceable across the dispatcher, store, and command layers." | comply |
| rule-dispatcher-error-hint | Any dispatcher-facing error or usage text touched by output routing must remain actionable. | rule-dispatcher-error-hint#n3196@v1:sha256:c88ba0daf18b558254516480e305ec64cacb996ef674a8540d5b99e103488cb4 "User-facing CLI errors from the top-level dispatcher guide the user to a next step, so a failure is actionable rather than a bare message." | review |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| RED tests | Add or update focused tests proving default structured output is TOON and agent mode does not need or honor JSON. | Expected initial failures before changing ResolveFormat/help. |
| Runtime helper | Change shared format resolution so no explicit JSON means TOON, and explicit JSON remains non-agent compatibility. | Focused output tests for ResolveFormat and writers. |
| Marketplace/show copy | Stop presenting marketplace --json as the machine-readable/default path and add a regression for agent/default list output if needed. | go test ./cmd -run Marketplace. |
| Rule governance | Update rule-output-via-helpers golden example to show TOON as default. | c3x read rule-output-via-helpers --section "Golden Example". |
| Parent delta | Parent Delta: none expected - c3-1 responsibilities already cover command output, validation, and agent-mode boundaries; no topology change. | c3x read c3-1, c3x lookup cli/cmd/output.go, and c3x check. |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| output format resolver | Change ResolveFormat and tests so TOON is default structured output and JSON requires explicit non-agent opt-in. | Focused output tests. |
| command help | Remove or revise copy that says JSON is the default or machine-readable path where TOON is now default. | Help tests and grep for stale JSON-default wording. |
| rule-output-via-helpers | Rewrite the golden example section through c3x write --section so the rule matches code. | Focused rule check and c3x check. |
| marketplace/show tests | Ensure marketplace list/show behavior does not regress while JSON is not positioned as the agent default. | Focused marketplace tests. |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| cli/cmd/output_test.go | Default structured format is TOON; explicit non-agent JSON stays JSON; agent explicit JSON stays TOON. | go test ./cmd -run TestResolveFormat. |
| cli/cmd/marketplace_test.go | Marketplace list/show remains usable and agent/default structured list output is TOON-shaped. | go test ./cmd -run Marketplace. |
| cli/cmd/help_test.go | Help no longer advertises JSON as default machine-readable output for affected paths. | go test ./cmd -run TestShowHelp. |
| grep stale wording | Stale strings such as "Default output is JSON" are removed from active help. | rg "Default output is JSON" cli/cmd/help.go and rg "Machine-readable output" cli/cmd/help.go. |
| full Go suite | Runtime output changes do not break command behavior. | go test ./... from cli/. |
| C3 checks | ADR, rule, codemap, and canonical docs remain valid. | c3x check --include-adr --only <adr> and c3x check. |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Only change help text and leave ResolveFormat(false,false) as FormatHuman. | The default structured writer would still fall through to JSON, so the runtime contract would remain misleading. |
| Remove JSON support entirely. | Existing non-agent explicit JSON tests and compatibility callers still depend on parseable JSON, and the user asked to sunset JSON "for the time being" as the agent/default path rather than destroy explicit compatibility. |
| Change only marketplace show/list. | The root default-selection behavior lives in runtime-support, so marketplace-only changes would leave search/index/list/canvas behavior inconsistent. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Human users lose familiar prose output on commands that previously used generic structured writers. | Scope this to structured-capable helpers and keep command-specific human renderers where they intentionally exist. | Focused command tests and review diff. |
| Explicit JSON compatibility breaks for scripts. | Keep ResolveFormat(true, false) returning JSON and keep parse tests. | Existing JSON tests plus focused output tests. |
| Rule and code drift. | Update rule-output-via-helpers in the same change and run C3 check. | c3x check --include-adr --only <adr> and c3x check. |
| Marketplace show content becomes over-structured and harder to read. | Preserve rule preview markdown for show; focus structured TOON default on list/results. | Marketplace show tests. |

## Verification

| Check | Result |
| --- | --- |
| RED focused tests before implementation | PASS - initial focused run failed on missing MarketplaceOptions.JSONExplicit before implementation. |
| focused Go command tests | PASS - output, codemap, marketplace, help, and parser-level TOON tests passed. |
| go test ./cmd | PASS. |
| go test ./... from cli/ | PASS. |
| cd packages/cli && npm test && npm run build | PASS. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh list --compact | PASS - emitted TOON table beginning with totalCount and entities rows, not JSON. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh marketplace show rule-output-via-helpers --json | PASS - rejected explicit show JSON with error and hint. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh search "real-time sync" | PASS - still returns results and no SQL hyphen error. |
| stale help wording grep | PASS - no Default output is JSON or Machine-readable output matches in active help. |
| C3X_MODE=agent bash skills/c3/bin/c3x.sh check --include-adr --only adr-20260610-force-toon-default | PASS - focused ADR check issues empty after citation/ref updates. |
