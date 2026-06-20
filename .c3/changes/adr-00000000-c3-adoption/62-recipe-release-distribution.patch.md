---
target: recipe-release-distribution
scope: whole
type: recipe
title: Release Distribution
affects: c3-203, c3-301
---
# Release Distribution

## Goal

Trace how a `v*` version tag becomes installed software: CI cross-compiles the binary for the platform matrix and publishes a checksummed GitHub Release, then the two distribution surfaces converge on it — cli-wrapper (c3-203) selects the bundled per-platform binary in the fat skill zip, and binary-downloader (c3-301) fetches the pinned build for thin npm installs.
