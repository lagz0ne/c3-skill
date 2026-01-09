# Plan: Simplify E2E Tests to Focused Smoke Tests

**References:** adr-20260106-dev-browser-e2e

## Order of Operations

### 1. Create New Smoke Test

- [x] Create `apps/e2e/tests/smoke.spec.ts` with 6 critical tests
- [x] Keep fixtures (sample-invoice.xml, etc.)

### 2. Remove Old Complex Test Structure

- [x] Remove old `apps/e2e/tests/*.spec.ts` files
- [x] Remove `apps/e2e/fixtures/auth.fixture.ts`
- [x] Remove `apps/e2e/fixtures/websocket.fixture.ts`
- [x] Remove `apps/e2e/utils/` directory
- [x] Remove `apps/e2e/scripts/` directory

### 3. Update Configuration

- [x] Update `apps/e2e/playwright.config.ts` - simplified config
- [x] Update `apps/e2e/package.json` - with PLAYWRIGHT_SKIP_VALIDATE_HOST_REQUIREMENTS

### 4. Update Documentation

- [x] Update `apps/e2e/README.md`
- [x] Update `.c3/c3-3-e2e-tests/README.md`

### 5. Verification

- [x] Run smoke tests: `cd apps/e2e && bun run test`
- [x] Confirm 6/6 tests pass (~1 minute runtime)
- [x] Mark ADR as implemented
