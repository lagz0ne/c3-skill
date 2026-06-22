---
id: c3-4
c3-seal: 44d4a8cff9223730ba72e2b31eca04ac44ac3bba872b1ec56869622229e0b829
title: dev-tooling
type: container
boundary: service
parent: c3-0
goal: Hold the standalone build/test programs that support the c3x CLI but ship separately from the binary — the search-ranking quality harness and the embedding-asset builder — so they are first-class facts with their own code surfaces rather than undescribed corners of the tree.
---

# dev-tooling

## Goal

Hold the standalone build/test programs that support the c3x CLI but ship separately from the binary — the search-ranking quality harness and the embedding-asset builder — so they are first-class facts with their own code surfaces rather than undescribed corners of the tree.

## Components

| ID | Name | Category | Status | Goal Contribution |
| --- | --- | --- | --- | --- |
| c3-401 | search-eval | foundation | active | Measure whether a change to c3x search ranking is a real, numeric win — run a fixed set of labelled queries through the live search path against the local store and report Hit@1/3/5 and MRR so retrieval changes are kept only on a measured improvement. |
| c3-402 | semantic-assets | foundation | active | Build the ONNX embedding-model assets that the store's semantic search uses at runtime — produce the model, vocab, and per-platform runtime in the form go:embed expects, or the release-named asset set with checksums, so a build can ship or embed the pinned semantic model. |

## Responsibilities

This container is accountable for the developer-facing tooling that lives under `cli/tools/` and links the CLI's libraries but is never compiled into the released `c3x` binary. It owns two concerns: measuring whether a change to search ranking is a real, numeric win (the search-eval harness, run during development against the local store), and producing the ONNX semantic-model assets that the store's runtime semantic search depends on (the semantic-assets builder, run during release packaging). Each tool is a separate `main` program with its own command surface; the container draws the line between "ships in the binary" (the CLI containers) and "supports the binary from the side" (these tools). It is not accountable for the runtime semantic search itself (that is the store), nor for the CLI commands that invoke search.

## Complexity Assessment

Low structural complexity: two small, independent `main` programs with no shared state. The risk that matters is coupling — both tools import CLI/store internals, so a breaking change to those internals can break a tool's build without touching the binary; the container exists partly to keep that dependency visible and under its own eval.
