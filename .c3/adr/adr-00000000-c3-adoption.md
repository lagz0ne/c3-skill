---
id: adr-00000000-c3-adoption
c3-seal: 815f2eaf307132b6a3beef5e5be5a470c50aaf44284671f7a9c141b1bada446b
title: C3 Architecture Documentation Adoption
type: adr
goal: Adopt C3 to document c3-design's own architecture as frozen, verifiable facts — rebuilt from scratch against the current code after the prior model drifted.
status: accepted
---

# C3 Architecture Documentation Adoption

## Goal

Adopt C3 to document c3-design's own architecture as frozen, verifiable facts — rebuilt from scratch against the current code after the prior model drifted.

## Context

This repository is the C3 source: a Go CLI engine, a Claude skill, and an npm installer. Its own `.c3/` model had drifted badly — a mechanical sweep found ten-plus components whose code-map pointed at deleted or moved code, and several facts described commands that had been merged or removed (coverage-cmd, analysis-cmds, history-marketplace-cmds). Patching each in place was polishing rot, so the model was wiped and re-onboarded from the code as the single source of truth.

## Decision

Establish a lean topology: three containers — Go CLI (c3-1), Claude Skill (c3-2), npm @c3x/cli (c3-3) — over the system c3-0. The CLI carries seven foundation libraries (doc-model, store, schema, changeset, walker, codemap-lib, runtime-support) and four command-group features (read, author, change, lifecycle); the skill carries skill-definition, operation-references, and cli-wrapper; npm carries binary-downloader. The graph is wired by three refs (frontmatter-docs, cross-compiled-binary, fat-thin-distribution), three rules (dispatcher-error-hint, output-via-helpers, wrap-error-cause), and three recipes (validation, change-unit-saga, release-distribution). Every fact is materialized and frozen in one atomic flip; code-maps are set after the flip as the one mutable binding.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-0 c3-design | system | Re-onboarded: system goal and abstract constraints authored | N.A - created by this genesis unit | check clean; constraints reviewed |
| c3-1 Go CLI | container | Re-onboarded: 11 components rebuilt against current cli/ | N.A - created by this genesis unit | check clean; code-maps resolve via lookup |
| c3-2 Claude Skill | container | Re-onboarded: 3 components for SKILL.md, references/, bin/ | N.A - created by this genesis unit | check clean |
| c3-3 npm @c3x/cli | container | Re-onboarded: thin-client binary-downloader | N.A - created by this genesis unit | check clean |

## Verification

| Check | Result | Verification evidence |
| --- | --- | --- |
| Structural integrity | issues[0], canonical markdown in sync | c3x check |
| Fact ↔ code binding resolves | passes | cli/internal/store/store.go → c3-102; cli/cmd/change.go → c3-112; skills/c3/SKILL.md → c3-201; packages/cli/src/manager.ts → c3-301 |
| Membership and governance edges | synthesized and formed | reverse graph on each ref/rule lists its citing components |
