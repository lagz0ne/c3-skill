---
id: adr-20260415-rebuild-layer-disconnect-diagnostics
c3-seal: fb830f099d6ba5432356c58c1e3d1a0e646d4b7a18fee12b07c19222095da0f5
title: rebuild-layer-disconnect-diagnostics
type: adr
goal: When c3 rebuilds the local database from canonical markdown, surface layer-disconnect issues between context, containers, and components so agents fix integration drift instead of trusting a clean cache rebuild.
status: implemented
date: "2026-04-15"
affects:
    - c3-1
    - c3-113
    - c3-119
---

# Rebuild Layer Disconnect Diagnostics

## Goal

When c3 rebuilds the local database from canonical markdown, surface layer-disconnect issues between context, containers, and components so agents fix integration drift instead of trusting a clean cache rebuild.

## Work Breakdown

- Add `check` warnings for parent/child table drift: context `Containers` and container `Components` must match actual child entities.
- Add hints explaining that rebuild proves storage only, not layer integration.
- Make direct `c3x import --force` print layer integration issues immediately after rebuilding the database.
- Verify/repair already surface the same issues because rebuild is followed by the verification suite.
- Add tests for missing child rows, stale child rows, direct import surfacing, and verify rebuild output surfacing layer disconnects.
- Fix current repo drift by updating `c3-1` Components table to include `c3-106`, `c3-107`, `c3-108`, `c3-117`, `c3-118`, `c3-119`, and `c3-120`.

## Parent Delta

updated: `c3-1` Components now matches all Go CLI child components. `c3-113` owns the new layer-disconnect check. `c3-119` owns direct import plus verify/rebuild surfacing.

## Verification

- Red tests first: layer-disconnect check tests and rebuild-surfacing test failed before implementation.
- Green targeted tests: `go test ./cmd -run 'TestRunCheck_LayerDisconnect|TestHintFor'`, `go test ./cmd -run 'TestRunImport_SurfacesLayerDisconnectAfterRebuild|TestRunImport_RebuildsDBFromMarkdown'`, and `go test . -run 'TestRun_VerifyRebuildSurfacesLayerDisconnect'` passed.
- Full suite: `go test ./...` passed in `cli/`.
- Current repo smoke: `C3X_MODE=agent go run . --c3-dir ../.c3 check --include-adr` returned zero issues after fixing `c3-1` Components.
- Current repo smoke: `C3X_MODE=agent go run . --c3-dir ../.c3 verify` returned zero issues and canonical markdown sync OK.
