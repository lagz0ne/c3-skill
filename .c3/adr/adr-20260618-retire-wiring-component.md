---
id: adr-20260618-retire-wiring-component
c3-seal: 22a9fb66952239c0e6afcd37bafa3a4f56eaa0215f1d47d7956540d83fa29969
title: retire-wiring-component
type: adr
goal: Retire the `wiring` component (c3-104); its concern moved into `content` (c3-106) when the `wire` command was removed.
status: accepted
date: "2026-06-18"
---

## Goal

Retire the `wiring` component (c3-104); its concern moved into `content` (c3-106) when the `wire` command was removed.

## Context

The `wire` command was removed (superseded by canvas edge-columns — the column IS the citation, materialized by `SyncCanvasOwnedRelationships`). c3-104 "wiring" was the wire command's component; its code-map pointed only at the now-deleted `cli/cmd/wire.go`, and the real edge-materialization code lives under `cli/internal/content/**` (owned by c3-106) and `cli/internal/schema/canvas.go`. c3-104 no longer owns distinct territory. c3-111 (add-cmd) and c3-113 (check-cmd) referenced c3-104 via `uses`, and the c3-1 README lists it.

## Decision

Retire c3-104. Re-point c3-111 and c3-113 `uses` from c3-104 to c3-106 (where the edge-materialization code now lives). Remove the c3-104 row from the c3-1 README Components table. The "wiring" concern is documented by c3-106 (content) going forward.

## Affected Topology

| Entity | Type | Why affected | Evidence | Governance review |
| --- | --- | --- | --- | --- |
| c3-104 | component | Retired — its code (wire.go) was removed and the edge-materialization concern is subsumed by c3-106. | c3-104#n2143@v1:sha256:d640feb0aaeb592c08460b2ba9c2cd14c0da91aec43f8d3bea4a93651d3738e7 | N.A - retire |
| c3-111 | component | Re-edged: used c3-104; now uses c3-106 where the wiring code lives. | c3-111#n2487@v1:sha256:e6fad9e0289bcae50e2398556208897e696140a220eb4c0d3bb5d74085bf8d00 | No delta — same dependency, relocated target |
| c3-113 | component | Re-edged: used c3-104; now uses c3-106 where the wiring code lives. | c3-113#n2582@v1:sha256:cf24ddb428f8aecd7dc88a46115e6dfaf1a4a29525c628bdaed6f02ce94b4435 | No delta — same dependency, relocated target |
| c3-1 | container | Components table drops the retired c3-104 row. | c3-1#n1972@v1:sha256:b31de913a9eaa864d4de5f10894602485b8bff069d1c91f691bff809ce59c5b9 | No delta — membership only |

## Verification

| Check | Result |
| --- | --- |
| c3x check | canonical markdown in sync, no orphan/dangling refs |
| cd cli && go test ./... | all packages green |
