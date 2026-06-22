---
target: recipe-validation
scope: whole
type: recipe
title: Validation
affects: c3-101, c3-105, c3-103, c3-110, c3-102
---
# Validation

## Goal

Trace the `c3 check` flow that no single component owns: walker (c3-105) discovers every fact file in `.c3/`, doc-model (c3-101) parses each into a node tree, store (c3-102) supplies the sealed state, schema (c3-103) validates each body against its canvas, and read-cmds (c3-110) reports the structural verdict — errors versus warnings.
