---
target: c3-201
scope: whole
type: component
parent: c3-2
title: skill-definition
---
# skill-definition

## Goal

Be the skill's entry document — classify the agent's intent, teach the one-model / three-act story, and state the contract every operation reference cites.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-2 |
| Role | The skill's front door: the intent router and shared contract an agent loads before any C3 work. |
| Boundary | Owns routing, the three-act narrative, and the named gates; teaches no per-operation procedure — that lives in the references it dispatches to. |
| Collaboration | Routes each classified op to operation-references; names the CLI handle that cli-wrapper provides; states the freeze and membership rules the references cite back to. |

## Purpose

Carry SKILL.md: the frontmatter triggers, the one-model / three-act framing (build the model and freeze the facts, change-units drive progress, the canvas grows and evolves), the intent-classification table that maps keywords to an op and its reference, the dispatch steps and read-only fast path, the shared contract (frozen facts, membership-by-construction, rung completeness, the ADR status set), and the command table. Non-goals: executing operations (the packaged CLI does that), teaching a single operation's procedure (operation-references), or selecting and invoking the binary (cli-wrapper).

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-fat-thin-distribution | ref | Why the skill is the router over a packaged single-source binary rather than re-implementing behavior in prose | Skill teaches; the bundled binary executes | SKILL.md states "the packaged CLI is the single source; this skill is the router," matching the fat-skill packaging. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| intent router | IN | An agent message classifies to exactly one op via the keyword table; ambiguity escalates to AskUserQuestion | Routes only — never mutates `.c3/` or bypasses the CLI | skills/c3/SKILL.md §Intent Classification, §Dispatch |
| shared contract | OUT | States the freeze, membership-by-construction, rung-completeness, and ADR-status rules once, for every reference to cite | Names gates; defines no operation's step-by-step | skills/c3/SKILL.md §The shared contract, §Command table |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| skills/c3/SKILL.md | Goal | Wording and table layout may vary while routing and the three-act contract stay intact | skills/c3/references/*.md — the Intent table resolves every op to a file here |
