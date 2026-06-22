---
target: ref-frontmatter-docs
scope: whole
type: ref
title: Frontmatter Docs Pattern
---
# Frontmatter Docs Pattern

## Goal

Standardize every `.c3/` document as YAML frontmatter plus canvas-shaped markdown sections.

## Choice

Each fact is a markdown file whose frontmatter carries identity and structured fields (id, type, title, parent, seal) and whose body carries the canvas's required sections.

## Why

A single representation that is human-diffable, git-friendly, and machine-parseable — the frontmatter gives stable structured fields without a separate schema file, and the body stays readable to a person reviewing a diff.

## How

The doc-model component parses frontmatter and markdown into a node tree; the store seals that tree with a content merkle; check validates the required sections per canvas.
