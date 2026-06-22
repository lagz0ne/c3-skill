---
id: ref-cross-compiled-binary
c3-seal: d3ef0600a734750458802bad4799cd6419377523aef6ee243deb42cab7d58432
title: Cross-Compiled Binary Distribution
type: ref
goal: Ship c3x as standalone per-platform binaries with no runtime dependency.
---

# Cross-Compiled Binary Distribution

## Goal

Ship c3x as standalone per-platform binaries with no runtime dependency.

## Choice

Cross-compile the Go CLI for linux/amd64, linux/arm64, and darwin/arm64 in CI, and attach the checksummed artifacts to the GitHub Release.

## Why

A single static binary installs anywhere with no toolchain present; cross-compiling in CI keeps every release reproducible and independent of any one developer machine.

## How

A `v*` version tag triggers the distribute workflow, which builds the platform matrix and uploads the assets the skill wrapper and npm client download on demand.
