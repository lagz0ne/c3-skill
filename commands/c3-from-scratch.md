---
description: Design a new system architecture from scratch using C3 methodology
---

Initialize C3 documentation structure and guide through designing a complete system architecture from the ground up.

## What This Does

1. Creates `.c3/` directory structure if not exists
2. Guides through Context level design (system landscape)
3. Elaborates Container level for each container
4. Details Component level for key components
5. Documents decisions via ADRs

## Use When

- Starting a new project with no architecture documentation
- Documenting an existing undocumented system
- Replacing non-C3 documentation with C3 structure

## Process

This command initializes the C3 structure then invokes the c3-design skill to guide you through:

1. **Context Level**: Bird's-eye view of system, users, external dependencies
2. **Container Level**: Individual applications, services, databases
3. **Component Level**: Internal structure of key containers

Each level produces properly formatted documents with unique IDs and cross-references.

## First Steps

Create the directory structure:
```bash
mkdir -p .c3/{containers,components,adr,scripts}
```

Then describe your system requirements to begin the design process.
