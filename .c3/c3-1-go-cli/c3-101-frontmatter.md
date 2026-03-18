---
id: c3-101
c3-version: 4
title: frontmatter
type: component
category: foundation
parent: c3-1
goal: Parse and write YAML frontmatter embedded in .c3/ markdown files
summary: Provides Get/Set access to frontmatter fields; used by every command that reads entity metadata
uses: []
---

# frontmatter

## Goal

Parse and write YAML frontmatter embedded in `.c3/` markdown files.

## Container Connection

Every command reads entity identity (id, title, goal, type) from frontmatter. Without this, no command can identify or navigate entities.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Markdown file bytes |  |
| OUT (provides) | Parsed YAML fields + body text |  |

## Code References

| File | Purpose |
|------|---------|
| `cli/internal/frontmatter/frontmatter.go` | Parse/write frontmatter |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-frontmatter-docs | Defines the frontmatter schema that this component parses |
