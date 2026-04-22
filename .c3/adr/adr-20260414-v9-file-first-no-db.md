---
id: adr-20260414-v9-file-first-no-db
c3-seal: 40a69ef1c9d65125a6422f97e5751a2e1eeee9e1c95e71f30b7c1529cd6e46ea
title: v9-file-first-no-db
type: adr
goal: Align C3 docs and UX with v9 canonical-file workflow where DB is cache only and submitted state is .c3 files.
status: proposed
date: "2026-04-14"
---

# V9 File-First Workflow

## Goal

Align C3 docs and UX with v9 canonical-file workflow where DB is cache only and submitted state is .c3 files.

## Context

v9 uses canonical .c3 markdown as shared truth. Local cache may exist, but Git review and submission should ignore cache artifacts and trust verified files.

## Decision

Update skill docs and CLI-facing copy to describe DB as ignorable cache, prefer verify/repair over database migration framing, and make submitted artifact guidance explicitly file-first.

## Consequences

- Operators stop treating DB as authority
- Submitted changes stay in canonical .c3 files
- Migration guidance shifts to cache maintenance language
