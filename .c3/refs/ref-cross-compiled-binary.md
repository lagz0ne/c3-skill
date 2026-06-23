---
id: ref-cross-compiled-binary
c3-seal: 4739c3260fe1ca298bd03a40013913249479f620e10a9551322eb5812223b96d
title: Cross-Compiled Binary Distribution
type: ref
goal: Ship c3x as named per-platform release binaries, with standard runtime-manager binaries, full-fat semantic skill binaries, and Linux portable binaries where distro/sandbox compatibility matters more than local ONNX semantic search.
---

# Cross-Compiled Binary Distribution

## Goal

Ship c3x as named per-platform release binaries, with standard runtime-manager binaries, full-fat semantic skill binaries, and Linux portable binaries where distro/sandbox compatibility matters more than local ONNX semantic search.

## Choice

Build the standard Go CLI release binary for linux/amd64, linux/arm64, and darwin/arm64; build full-fat embedmodel skill binaries for that same matrix; and build additional pure-Go Linux portable binaries for linux/amd64 and linux/arm64 named `c3x-{VERSION}-linux-{arch}-portable`.

## Why

Standard and full-fat builds preserve the existing feature-complete runtime path, including embedded semantic assets for self-contained skill installs. Pure-Go Linux portable builds give musl, Alpine, distroless-like, and tightly sandboxed environments a bundled core runtime without forcing a heavier native ONNX/musl build.

## How

A `v*` version tag triggers the distribute workflow, which runs the release build variant: thin plus full-fat binaries for each platform in the matrix, and a `CGO_ENABLED=0` portable binary for each Linux arch. Release assembly publishes the standard binaries for the npm manager and packages full-fat and Linux portable binaries into their matching skill ZIPs.
