---
id: c3-110
c3-version: 4
c3-seal: 456e0eb92870a8255fd0d45d91fcf84163a5946df4068572f89199b712f2cce6
title: init-cmd
type: component
category: feature
parent: c3-1
goal: Scaffold the `.c3/` directory structure from scratch for a new project.
summary: Creates .c3/config.yaml, README.md, refs/, adr/, and adr-000 stub in one command
uses:
    - c3-103
---

# init-cmd
## Goal

Scaffold the `.c3/` directory structure from scratch for a new project.

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| IN (uses) | Scaffolding templates | c3-103 |
| OUT (provides) | .c3/ directory structure |  |
## Code References

| File | Purpose |
| --- | --- |
| cli/cmd/init.go | Init command implementation |
