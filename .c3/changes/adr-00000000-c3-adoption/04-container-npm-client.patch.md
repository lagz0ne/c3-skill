---
target: c3-3
scope: whole
type: container
parent: c3-0
title: npm @c3x/cli
---
# npm @c3x/cli

## Goal

Install and run the c3x binary from npm — a thin client that downloads the right platform build and forwards arguments.

## Components

| ID | Name | Category | Status | Goal Contribution |
| --- | --- | --- | --- | --- |

## Responsibilities

Publish the `@c3x/cli` package; resolve and cache the platform binary for the pinned version; forward the user's command line to it. Holds no architecture behavior — distribution only.

## Complexity Assessment

Low: a TypeScript downloader/forwarder with a pinned version. The only real risk surface is platform and version resolution.
