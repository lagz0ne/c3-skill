---
id: ref-frontmatter-docs
c3-version: 4
title: Frontmatter Docs Pattern
goal: Every .c3/ doc uses YAML frontmatter for machine-readable metadata and Markdown body for human-readable content
via: [c3-101, c3-103]
---

# Frontmatter Docs Pattern

## Goal

Every `.c3/` doc uses YAML frontmatter for machine-readable metadata and a Markdown body for human-readable content.

## Choice

All `.c3/` architecture docs use YAML frontmatter (between `---` delimiters) followed by a Markdown body.

## Why

- **Machine-readable identity**: `id`, `type`, `parent`, `goal` fields enable CLI traversal without parsing prose
- **Human-readable content**: Markdown body allows rich documentation with tables and diagrams
- **Separation of concerns**: Metadata (frontmatter) vs content (body) are independent

## How

```markdown
---
id: c3-NNN
c3-version: 4
title: My Component
type: component
category: foundation
parent: c3-N
goal: One-line goal statement
summary: Brief summary
---

# My Component

## Goal

One-line goal statement.
```

## Not This

Do not put structural metadata (id, type, parent) in the markdown body — it won't be parseable by the CLI.
