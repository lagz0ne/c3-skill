---
id: ref-embedded-templates
c3-version: 4
c3-seal: 672dbb981b5f8e27df82fd23405b79d9809cc03be3e889d1256bccfa0a04d2b2
title: Embedded Templates via go:embed
type: ref
goal: Doc templates are bundled in the CLI binary so scaffolding works without external files at any install path.
via:
    - c3-103
---

# Embedded Templates via go:embed

## Goal

Doc templates are bundled in the CLI binary so scaffolding works without external files at any install path.

## Choice

Use Go's `//go:embed` directive to bundle `.md` template files into the compiled binary.

## Why

- **Self-contained binary**: `c3x init` and `c3x add` work with no external template files
- **Template versioning**: Templates ship with the binary version — no drift between binary and templates
- **Consistent scaffolding**: All users get the same templates regardless of their local setup

## How

```go
//go:embed *.md
var templateFS embed.FS
```

Templates live in `cli/internal/templates/*.md` and are accessed via `templateFS.ReadFile(name)`.

## Not This

Do not read templates from the filesystem at runtime — templates must travel with the binary.
