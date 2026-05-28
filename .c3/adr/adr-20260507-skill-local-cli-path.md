---
id: adr-20260507-skill-local-cli-path
c3-seal: db538e422da84fdc3874492229422911a8a7c1d044a0c47c82ae0039cb26fb5c
title: skill-local-cli-path
type: adr
goal: Change the C3 skill instructions so agents invoke the versioned CLI through the skill-local packaged binary path, with a session alias or function permitted only as a token-saving shortcut to that resolved path.
status: implemented
date: "2026-05-07"
---

## Goal

Change the C3 skill instructions so agents invoke the versioned CLI through the skill-local packaged binary path, with a session alias or function permitted only as a token-saving shortcut to that resolved path.

## Context

The skill is distributed with its binary so C3 remains complete and operable in restricted environments without external downloads. Current skill prose still uses c3x as the apparent command name in several places, which can make agents treat a user convenience shortcut or global install as the authoritative executable. The affected topology is the Claude Skill container, its skill-router component, and operation reference docs because they define agent entry, preconditions, and command routing.

## Decision

The skill will define a resolved command handle named c3 as a session-local alias or shell function pointing at C3X_MODE agent bash skill-dir bin c3x.sh. Agent-facing prose will describe bare c3x only as the underlying packaged CLI or user-facing shorthand, not as the agent invocation path. This keeps the binary version tied to the skill installation while still reducing token cost during agent runs.

## Affected Topology

| Entity | Type | Why affected | Governance review |
| --- | --- | --- | --- |
| c3-2 | container | The container describes routing through C3 commands and must remain correct for packaged binary use. | Parent Delta required; responsibilities still valid with local command handle wording. |
| c3-201 | component | skills c3 SKILL md belongs to skill-router and defines command entry and dispatch behavior. | Verify command prose no longer instructs agents to use bare c3x. |
| c3-210 | component | Operation reference contract is affected because all operation docs inherit the command invocation convention. | Verify references use the c3 handle for agent examples. |
| c3-211 | component | Onboard workflow examples include command snippets agents execute. | Verify examples use the c3 handle. |
| c3-212 | component | Query workflow examples include command snippets agents execute. | Verify examples use the c3 handle. |
| c3-213 | component | Audit workflow examples include command snippets agents execute. | Verify examples use the c3 handle. |
| c3-214 | component | Change workflow examples include command snippets agents execute. | Verify examples use the c3 handle. |
| c3-216 | component | Ref workflow examples include command snippets agents execute. | Verify examples use the c3 handle. |
| c3-217 | component | Rule workflow examples include command snippets agents execute. | Verify examples use the c3 handle. |
| c3-218 | component | Sweep workflow examples include command snippets agents execute. | Verify examples use the c3 handle. |

## Compliance Refs

| Ref | Why required | Action |
| --- | --- | --- |
| ref-cross-compiled-binary | The packaged binary and offline distribution contract is the reason agent invocation must use the skill-local binary path. | comply |

## Compliance Rules

| Rule | Why required | Action |
| --- | --- | --- |
| N.A - no governing rule surfaced by c3 lookup skills c3 SKILL md | The lookup returned the owning component and no rule IDs, so there is no rule-specific golden example to apply. | N.A - no applicable rule |

## Work Breakdown

| Area | Detail | Evidence |
| --- | --- | --- |
| Skill CLI instructions | Update skills c3 SKILL md to introduce a token-saving c3 alias or function bound to skill-dir bin c3x.sh. | Diff plus rg check for remaining bare-agent command examples. |
| Operation references | Replace agent-facing c3x examples in skills c3 references md files with c3 command handle examples. | rg review of c3x and skill-dir references in skills c3 markdown. |
| Verification | Run focused ADR check and repo-level c3 check. | Command outputs recorded before implementation closure. |

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| Skill instructions | Change agent command contract in skills c3 SKILL md; no Go CLI validator or template behavior changes. | rg confirms instruction wording; C3 checks confirm docs structure. |
| Operation references | Change command examples in audit, change, onboard, query, ref, rule, and sweep workflow docs to use the c3 handle. | rg confirms no bare c3x agent examples remain in skill markdown. |
| Validators tests help | N.A - this is skill routing documentation, not CLI runtime behavior. | c3 check remains the enforcement surface for doc integrity. |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| skills c3 SKILL md | Agents are told to create and use a local command handle backed by the packaged binary. | File diff and targeted text search. |
| operation references | Workflow examples route through c3 handle instead of bare c3x or direct wrapper calls. | Targeted text search. |
| focused ADR check | ADR remains structurally valid while work is active. | Must pass before status transition. |
| repo c3 check | Existing C3 topology remains valid after the skill instruction edit. | Must pass before final response. |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| Keep documenting bare c3x for agents | It can resolve to a global install or user shortcut, breaking the versioned packaged-binary contract. |
| Spell out the full skill-local c3x shell command everywhere | It is correct but wastes tokens and makes workflows noisier; a local alias or function preserves correctness with lower token cost. |
| Require external CLI install | It violates the restricted and offline environment reason for bundling the binary with the skill. |

## Risks

| Risk | Mitigation | Verification |
| --- | --- | --- |
| Agents may confuse alias name with global executable | Define the alias or function as session-local and explicitly backed by skill-dir bin c3x.sh. | rg review of command contract wording. |
| User-facing shorthand may disappear from docs entirely | Keep the underlying binary name visible only as implementation detail and not the agent command path. | Human diff review. |
| ADR could drift from actual skill prose | Verify after edit with targeted search and C3 checks. | rg plus focused ADR check. |

## Verification

| Check | Result |
| --- | --- |
| focused ADR check for adr 20260507 skill local cli path | PASS - c3 check --include-adr --only adr-20260507-skill-local-cli-path returned zero issues before implementation status transition. |
| rg command wording review for skills c3 markdown | PASS - only packaged wrapper mentions remain: skill binary path, function definition, and packaged c3x.sh source-of-truth sentence. |
| git diff whitespace check | PASS - git diff --check returned clean. |
| repo-level c3 check | PASS - c3 check returned zero issues. |
| CLI regression tests | PASS - go test ./... from cli passed. |
