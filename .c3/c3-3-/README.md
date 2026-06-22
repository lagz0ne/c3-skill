---
id: c3-3
c3-seal: bc20073ea553605e1de7864b09dafa4dc205885a951fbb5135c4079c9e840358
title: npm @c3x/cli
type: container
parent: c3-0
goal: Install and run the c3x binary from npm — a thin client that downloads the right platform build and forwards arguments.
---

# npm @c3x/cli

## Goal

Install and run the c3x binary from npm — a thin client that downloads the right platform build and forwards arguments.

## Components

| ID | Name | Category | Status | Goal Contribution |
| --- | --- | --- | --- | --- |
| c3-301 | binary-downloader |  | active | Resolve the pinned platform binary, download and checksum-verify it into a per-version cache, then exec it with the user's arguments forwarded. |

## Responsibilities

Publish the `@c3x/cli` package; resolve and cache the platform binary for the pinned version; forward the user's command line to it. Holds no architecture behavior — distribution only.

## Complexity Assessment

Low: a TypeScript downloader/forwarder with a pinned version. The only real risk surface is platform and version resolution.
