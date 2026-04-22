---
id: adr-20260415-clean-hook-cache-diff
c3-seal: 5540c659c83ae043248f3550749ad30da746e25326e3f4da34827132f0304e82
title: clean-hook-cache-diff
type: adr
goal: Fix the C3 pre-commit hook so successful verification does not block commits just because `.c3/c3.db` local cache changed.
status: implemented
date: "2026-04-15"
affects:
    - c3-119
---

# Clean Hook Cache Diff

## Goal

Fix the C3 pre-commit hook so successful verification does not block commits just because `.c3/c3.db` local cache changed.

## Work Breakdown

- Keep staged-cache guard: staged `.c3/c3.db` still blocks commit.
- Change post-verify dirty check to exclude disposable cache files: `.c3/c3.db`, `.c3/c3.db-*`, `.c3/*.tmp.db`, `.c3/*.tmp.db-*`.
- Update installed local pre-commit hook through `c3x git install`.
- Add regression test proving generated hook does not diff all `.c3` blindly.

## Parent Delta

none: `c3-119` already owns git guardrail and sync lifecycle flows; no parent responsibility change needed.

## Verification

- Red test first: generated hook test failed because hook used `git diff --quiet -- .c3`.
- Green targeted hook tests passed.
- Full suite: `go test ./...` passed in `cli/`.
- Source verify: `C3X_MODE=agent go run . --c3-dir ../.c3 verify` returned zero issues and canonical sync OK.
