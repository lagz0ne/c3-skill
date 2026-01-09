---
id: adr-20260106-dev-browser-e2e
title: Simplify E2E Tests to Focused Smoke Tests
status: implemented
date: 2026-01-06
affects: [c3-0, c3-3]
---

# Simplify E2E Tests to Focused Smoke Tests

## Status

**Implemented** - 2026-01-06

## Problem

The previous Playwright-based e2e test suite was slow (several minutes), complex (multiple fixtures, helpers, and scattered test files), and difficult to run quickly during development. Tests frequently failed due to timing issues with WebSocket sync. We needed a faster, more reliable way to verify critical user journeys.

## Decision

Simplify the Playwright e2e test suite to a single focused smoke test file:

1. Consolidate all critical happy-path tests into `tests/smoke.spec.ts`
2. Remove complex fixtures, helpers, and multiple test files
3. Use simpler waiting strategies (timeouts + selectors instead of WebSocket listeners)
4. Run in ~1 minute total (vs several minutes)
5. Cover the same critical flows: login, invoice upload, PR creation, 3-stage approval, payment methods, invoice-PR linking

## Rationale

| Considered | Rejected Because |
|------------|------------------|
| Keep complex structure, optimize | Complexity was the problem - simpler is better |
| Cypress migration | Different tool, no significant benefit |
| Unit tests only | Need real browser interaction for critical user journeys |
| Dev-browser skill | More complexity than needed - Playwright directly is sufficient |

**Why simplified Playwright:**
- Same familiar Playwright API
- Single test file is easier to maintain
- Simpler waiting strategies are more reliable
- Standard `npx playwright test` command
- Built-in debugging with headed mode, UI mode

## Affected Layers

| Layer | Document | Change |
|-------|----------|--------|
| Container | c3-3 | Simplified architecture - single smoke test file |
| Components | c3-301 to c3-326 | Consolidated into c3-301 (config), c3-310 (smoke tests), c3-320 (fixtures) |

## Verification

- [x] All 6 smoke tests pass (Login, Invoice, PR, Approval, Payment, Linking)
- [x] Total runtime ~1 minute (66 seconds)
- [x] Old complex test files removed from apps/e2e/
- [x] New tests documented in c3-3 README
- [x] PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS set for WSL2/Nix
