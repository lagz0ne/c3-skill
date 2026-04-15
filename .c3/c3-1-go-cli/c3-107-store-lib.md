---
id: c3-107
c3-seal: f8ad094d8f529e47f4db9d60da9b7991bcc73e603c5f4a1a69be41757f660af1
title: store-lib
type: component
category: foundation
parent: c3-1
goal: Provide persistent entity, relationship, changelog, codemap, hash, node, and version storage operations for the CLI.
---

# store-lib
## Goal

Provide persistent entity, relationship, changelog, codemap, hash, node, and version storage operations for the CLI.

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| OUT | Persistence APIs for all command layers |  |
## Container Connection

Acts as storage backbone for canonical import/export, query, and mutation commands.
