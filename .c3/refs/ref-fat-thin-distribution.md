---
id: ref-fat-thin-distribution
c3-seal: d44c96a16ce83bf55f6b955666f7b4b3781636886eb89a22b7f2a014dd33d296
title: Fat Skill / Thin Client Distribution
type: ref
goal: Distribute C3 as a fat skill zip (binaries bundled) and a thin npm client (binary downloaded on demand).
---

# Fat Skill / Thin Client Distribution

## Goal

Distribute C3 as a fat skill zip (binaries bundled) and a thin npm client (binary downloaded on demand).

## Choice

The skill ships per-platform binaries inside the plugin zip; the npm `@c3x/cli` client ships only a downloader that fetches the pinned binary at install time.

## Why

The skill must work offline at any install path with its binary and templates bundled, while npm users want a small package that pulls only the binary they need — two audiences, two packaging strategies, one underlying binary.

## How

CI assembles the plugin zip with binaries for the release matrix; the npm client pins `C3X_VERSION` and downloads the matching build from the GitHub Release.
