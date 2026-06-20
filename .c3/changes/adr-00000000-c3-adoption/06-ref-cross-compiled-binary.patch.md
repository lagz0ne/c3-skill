---
target: ref-cross-compiled-binary
scope: whole
type: ref
title: Cross-Compiled Binary Distribution
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
