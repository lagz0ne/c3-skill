---
id: c3-111
c3-version: 4
title: add-cmd
type: component
category: feature
parent: c3-1
goal: Create new containers, components, refs, rules, or ADRs with correct numbering and wired into the parent doc
summary: Assigns IDs via the numbering library, creates stub docs from templates, and updates parent component tables
uses: [c3-101, c3-102, c3-103, c3-104]
---

# add-cmd

## Goal

Create new containers, components, refs, rules, or ADRs with correct numbering and wired into the parent doc.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Frontmatter parsing | c3-101 |
| IN (uses) | Entity discovery | c3-102 |
| IN (uses) | Scaffolding templates | c3-103 |
| OUT (provides) | New entity files with correct numbering |  |

## Code References

| File | Purpose |
|------|---------|
| `cli/cmd/add.go` | Core add logic |
| `cli/cmd/add_rich.go` | Interactive prompting |
| `cli/internal/numbering/numbering.go` | ID assignment |
| `cli/internal/writer/writer.go` | Safe file creation |
