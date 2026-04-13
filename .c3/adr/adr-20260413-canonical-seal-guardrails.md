---
id: adr-20260413-canonical-seal-guardrails
c3-seal: 90cef08e51d46b3e48d255092450ce10638bdee6f239af5a4a1c91c2b68cfa54
title: canonical-seal-guardrails
type: adr
goal: Make canonical .c3 markdown safe to diff and merge by detecting manual edits and enforcing a verified reseal flow.
status: proposed
date: "2026-04-13"
---

# Canonical Seal Guardrails
## Goal

Make canonical .c3 markdown safe to diff and merge by detecting manual edits and enforcing a verified reseal flow.

## Context

Canonical .c3 markdown is now the Git merge surface. We need a seal so direct edits are detectable, and the skill harness must treat unverified markdown as untrusted.

## Decision

Add deterministic content seals to canonical exports, verify them during import and sync check, and document the break-glass reseal flow.

## Consequences

- Direct edits become detectable
- Import requires --force to reseal broken or missing seals
- Skill instructions must require sync check before trusting canonical files
