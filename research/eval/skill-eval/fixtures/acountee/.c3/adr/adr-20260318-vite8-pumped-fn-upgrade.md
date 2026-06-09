---
id: adr-20260318-vite8-pumped-fn-upgrade
c3-seal: 73669a51549e42e0a0abb40beefa6c0a0e0382f63d0c9de4e1af00b04f881e63
title: Vite 8 + pumped-fn/lite Ecosystem Upgrade
type: adr
goal: Upgrade from Vite 8.0.0-beta.16 → 8.0.0 stable (Rolldown bundler), upgrade ecosystem deps to latest compatible versions, update @pumped-fn/lite 2.1.4 → 2.1.5, remove vinxi, and update C3 ref-pumped-fn to reflect current API surface from CLI.
status: proposed
date: "2026-03-18"
affects:
    - c3-101
    - c3-201
    - c3-202
    - ref-pumped-fn
---

# Vite 8 + pumped-fn/lite Ecosystem Upgrade

## Goal

Upgrade from Vite 8.0.0-beta.16 → 8.0.0 stable (Rolldown bundler), upgrade ecosystem deps to latest compatible versions, update @pumped-fn/lite 2.1.4 → 2.1.5, remove vinxi, and update C3 ref-pumped-fn to reflect current API surface from CLI.

## Current State

| Package | Current | Target | Notes |
| --- | --- | --- | --- |
| vite | 8.0.0-beta.16 | 8.0.0 | Stable: Rolldown bundler, Oxc minifier |
| @vitejs/plugin-react | ^5.0.4 | ^6.0.1 | Requires vite ^8.0.0 |
| vitest | ^3.0.5 | ^4.1.0 | Supports vite ^8.0.0-0 |
| @tailwindcss/vite | ^4.0.6 (installed 4.1.17) | 4.2.1 | Blocker: peer dep ^5.2.0 |
| @tanstack/react-start | ^1.157.15 | ^1.166.16 | peer dep >=7.0.0 — OK |
| @tanstack/react-router | ^1.157.15 | ^1.167.4 | OK |
| @tanstack/router-plugin | ^1.157.15 | ^1.166.13 | peer dep >=5.0.0 |
| @tanstack/devtools-vite | ^0.3.12 | latest | peer dep ^6.0.0 |
| nitro | latest | latest | peer dep ^7 |
| vinxi | ^0.4.3 | REMOVE | Only used for deleteCookie in logout.tsx; not used in build/dev |
| @pumped-fn/lite | ^2.1.4 | ^2.1.5 | Minor bump |
| @pumped-fn/react-lite | ^0.3.0 | ^0.3.0 | No change |
| vite-node (root) | ^5.3.0 | ^6.0.0 | Updated with vitest ecosystem |
| vite-tsconfig-paths | ^5.1.4 | ^5.1.4 | peer dep * — OK, but Vite 8 has built-in resolve.tsconfigPaths |
| tailwindcss | ^4.0.6 | ^4.2.1 | OK |

## Vite 8 Breaking Changes (from migration guide)

1. **`build.rollupOptions` → `build.rolldownOptions`** — auto-compat layer exists but deprecated
2. **Oxc replaces esbuild** for JS transforms — `esbuild` config → `oxc`
3. **Lightning CSS** is default CSS minifier
4. **Plugin authors**: `load`/`transform` hooks should return `moduleType: 'js'` when converting to JS
5. **CJS interop**: default import from CJS = `module.exports` value
6. **`build.rollupOptions.output.manualChunks`** object form removed; function form deprecated
7. **`build.commonjsOptions`** is now no-op
8. **Node.js 20.19+, 22.12+** required

## Work Breakdown

### Phase 1: Vite 8 Core (apps/start)

1. **Upgrade vite** `8.0.0-beta.16 → 8.0.0`
2. **Upgrade @vitejs/plugin-react** `^5.0.4 → ^6.0.1`
3. **Migrate vite.config.ts**: `build.rollupOptions` → `build.rolldownOptions`

- The `onwarn` for `UNUSED_EXTERNAL_IMPORT` needs to move to rolldownOptions equivalent

1. **Upgrade @tailwindcss/vite** to `4.2.1` — use `pnpm.overrides` or `--force` to bypass peer dep mismatch (plugin is CSS-only, functionally compatible with vite 8)
2. **Upgrade tailwindcss** to `^4.2.1`

### Phase 2: TanStack Ecosystem

1. **Upgrade @tanstack/react-start** `^1.157.15 → ^1.166.16`
2. **Upgrade @tanstack/react-router** `^1.157.15 → ^1.167.4`
3. **Upgrade @tanstack/router-plugin** `^1.157.15 → ^1.166.13`
4. **Upgrade @tanstack/react-router-ssr-query** to matching version
5. **Upgrade @tanstack/react-router-devtools** to matching version
6. **Upgrade @tanstack/devtools-vite** to latest

### Phase 3: Test Tooling

1. **Upgrade vitest** `^3.0.5 → ^4.1.0`
2. **Upgrade vite-node** (root) `^5.3.0 → ^6.0.0`
3. **Check vitest config** for any breaking changes (v3 → v4)

### Phase 4: Cleanup

1. **Remove vinxi** from dependencies — replace `deleteCookie` import in `logout.tsx` with cookie utility from existing deps (e.g., `cookie` package already in deps)
2. **Evaluate vite-tsconfig-paths removal** — Vite 8 has built-in `resolve.tsconfigPaths`; consider migrating and removing the plugin

### Phase 5: pumped-fn/lite Update

1. **Upgrade @pumped-fn/lite** `^2.1.4 → ^2.1.5`
2. **Update ref-pumped-fn** to reflect current API surface from CLI:

- Add `scope.select(atom, selector, { eq })` — derived state slices
- Add `scope.on(event, atom, listener)` — atom state observation
- Add `scope.flush()` — await pending invalidation chains
- Add tag static methods: `tag.get()`, `tag.find()`, `tag.collect()`
- Add tag introspection: `tag.atoms()`, `getAllTags()`
- Add `tags.all(tag)` — collect from all hierarchy levels
- Document `scope.release(atom)` — explicit release
- Ensure extension `wrapResolve` event kinds (`"atom"` vs `"resource"`) are documented

### Phase 6: Verify

1. **`pnpm dev`** — dev server starts, HMR works
2. **`pnpm build`** — production build succeeds
3. **`pnpm test`** — all tests pass
4. **`bunx @typescript/native-preview --noEmit`** — type check passes
5. **`c3x check`** — no structural issues

## Risks

| Risk | Severity | Mitigation |
| --- | --- | --- |
| @tailwindcss/vite peer dep mismatch with vite 8 | Medium | Use pnpm overrides; CSS plugin is bundler-agnostic. Monitor for official vite 8 support release. |
| TanStack Start 1.157 → 1.166 may have breaking changes | Medium | Read changelog; test routing, SSR, server functions end-to-end |
| vitest 3 → 4 may have config/API changes | Medium | Review vitest migration guide; run full test suite |
| build.rolldownOptions.onwarn API may differ from rollupOptions | Low | Check Rolldown docs; the onwarn pattern may need adaptation |
| vinxi removal — deleteCookie replacement | Low | cookie package already a dependency; straightforward swap |

## Context

N.A - historical ADR; original context is captured in the git log around the ADR date and in the current code that implements the decision.

## Decision

N.A - historical ADR; the decision matches the Goal above and has already shipped. Current .c3 topology reflects the implemented outcome.

## Underlay C3 Changes

| Underlay area | Exact C3 change | Verification evidence |
| --- | --- | --- |
| N.A - historical | Current .c3 entities, refs, and code-map are the post-change state. | c3x verify and c3x check |

## Enforcement Surfaces

| Surface | Behavior | Evidence |
| --- | --- | --- |
| N.A - historical | Enforcement is implicit in the currently linked components and refs. | c3x graph and cited ref ids on the relevant components |

## Alternatives Considered

| Alternative | Rejected because |
| --- | --- |
| N.A - historical | Alternatives were considered at decision time; rationale is preserved in the original commit message or branch discussion. |

## Verification

| Check | Result |
| --- | --- |
| Merged and running in production | PASS - see git log for the merge commit |
