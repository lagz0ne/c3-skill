---
id: c3-402
c3-seal: 6ff47a60e012e6bbbc7b9758e1d79693f79534bd210272861282e5868ad10745
title: semantic-assets
type: component
category: foundation
parent: c3-4
goal: Build the ONNX embedding-model assets that the store's semantic search uses at runtime — produce the model, vocab, and per-platform runtime in the form go:embed expects, or the release-named asset set with checksums, so a build can ship or embed the pinned semantic model.
uses:
    - ref-fat-thin-distribution
---

# semantic-assets

## Goal

Build the ONNX embedding-model assets that the store's semantic search uses at runtime — produce the model, vocab, and per-platform runtime in the form go:embed expects, or the release-named asset set with checksums, so a build can ship or embed the pinned semantic model.

## Parent Fit

| Field | Value |
| --- | --- |
| Parent | c3-4 |
| Role | The semantic-asset builder: a standalone main that materializes the embedding model and onnxruntime payload for a target platform, invoked during release packaging, not part of the shipped binary. |
| Boundary | Owns the build-time command surface and the choice of embed-mode vs release-mode; the actual asset preparation logic lives in the store, which this tool drives. |
| Collaboration | It parses --embed-dir/--release-dir (exactly one) plus --os/--arch, then calls the store's PrepareEmbeddedSemanticAssets or PrepareReleaseSemanticModelAssets to write the assets the runtime semantic index later consumes. |

## Purpose

Owns the build-time entry point that produces the semantic-search assets: an embed mode that writes `model.onnx`, `vocab.txt`, and the target-platform onnxruntime into a directory for `go:embed` (requiring `--os` and `--arch`), and a release mode that writes release-named assets plus checksums. It enforces that exactly one of the two modes is selected. Non-goals: the asset-preparation logic itself (owned by the store's `PrepareEmbeddedSemanticAssets`/`PrepareReleaseSemanticModelAssets`), runtime embedding/search, and asset distribution — it only invokes preparation with the right mode and target.

## Governance

| Reference | Type | Governs | Precedence | Notes |
| --- | --- | --- | --- | --- |
| ref-fat-thin-distribution | ref | The two build modes this tool drives produce exactly the assets the fat/thin split needs — embed mode feeds the fat skill build, release mode feeds the thin-client download. | Standard applies to both build modes | The asset-preparation logic itself lives in the store; this driver only selects the mode the distribution model requires. |

## Contract

| Surface | Direction | Contract | Boundary | Evidence |
| --- | --- | --- | --- | --- |
| flags | IN | Accept exactly one of --embed-dir or --release-dir; embed mode additionally requires --os and --arch. | Wiring only; it rejects an ambiguous mode and otherwise delegates. | cli/tools/semantic-assets/main.go |
| asset preparation | OUT | Invoke the store to write the embedded or release semantic assets for the chosen mode and target. | It calls the store's prepare functions; it does not build the assets itself. | cli/tools/semantic-assets/main.go |

## Derived Materials

| Material | Must derive from | Allowed variance | Evidence |
| --- | --- | --- | --- |
| cli/tools/semantic-assets/**.go | Contract | The flag set may grow as long as the one-mode-and-target / delegate-to-store contract holds | go test ./tools/... |
