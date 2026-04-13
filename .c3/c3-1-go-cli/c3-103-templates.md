---
id: c3-103
c3-version: 4
c3-seal: bddf89cc3df051403bf3a056299dc518358966554bb59c29a3577a1f9f721925
title: templates
type: component
category: foundation
parent: c3-1
goal: Provide embedded markdown templates for scaffolding new `.c3/` docs (context, container, component, ref, rule, ADR).
summary: Go embed bundle of all doc templates; used by init-cmd and add-cmd to create well-structured stubs
---

# templates
## Goal

Provide embedded markdown templates for scaffolding new `.c3/` docs (context, container, component, ref, rule, ADR).

## Container Connection

init-cmd and add-cmd produce correct, consistent doc stubs only because templates are embedded in the binary — no external files required.

## Dependencies

| Direction | What | From/To |
| --- | --- | --- |
| OUT (provides) | Template bytes for scaffolding |  |
## Code References

| File | Purpose |
| --- | --- |
| cli/internal/templates/templates.go | go:embed declarations |
| cli/internal/templates/*.md | Template source files |
## Related Refs

| Ref | How It Serves Goal |
| --- | --- |
| ref-embedded-templates | Pattern for bundling templates in binary via go:embed |
| ref-frontmatter-docs | Templates emit the standard frontmatter schema |
