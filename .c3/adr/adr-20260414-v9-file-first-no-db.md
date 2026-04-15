---
id: adr-20260414-v9-file-first-no-db
c3-seal: 2c0dccaf3b6430501b42ed68ce32b5835a5ce2bf2441f635b18b3c66b610a727
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
