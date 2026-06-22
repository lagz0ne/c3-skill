---
target: c3-1
scope: whole
type: container
parent: c3-0
title: Go CLI
---
# Go CLI

## Goal

Provide every c3x operation as a single cross-compiled Go binary — the engine that reads, writes, validates, and freezes the architecture graph.

## Components

| ID | Name | Category | Status | Goal Contribution |
| --- | --- | --- | --- | --- |

## Responsibilities

Own the entire behavior of C3: parse and render `.c3/` documents, persist the entity-relationship graph, validate canvas conformance, run the change-unit saga that is the only legal mutation path, and map facts to the code they govern. The skill (c3-2) and npm client (c3-3) only invoke this binary; no architecture logic lives outside it.

## Complexity Assessment

High-cohesion layered design: foundation libraries (doc-model, store, schema, changeset, walker, codemap, runtime-support) under a thin command surface grouped by read / author / change / lifecycle intent.
