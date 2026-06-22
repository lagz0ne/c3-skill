---
id: ref-frontmatter-docs
c3-seal: 5ec5d290948009c467723bd212d951aa625d3b25240d6bf86dcc361866a8e605
title: Frontmatter Docs Pattern
type: ref
goal: Standardize every `.c3/` document as YAML frontmatter plus canvas-shaped markdown sections.
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
