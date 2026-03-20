# C3 Structural Index
<!-- hash: sha256:19feb472218ed4ef5276e55c572b5584f89d2268f31029437260dfb4f06ae088 -->

## adr-00000000-c3-adoption — C3 Architecture Documentation Adoption (adr)
blocks: Goal ✓

## adr-20260312-upgrade-refs-golden-patterns — Upgrade Refs — Golden Patterns, Enforcement, Measurement (adr)
blocks: Goal ✓

## c3-0 — c3-design (context)
reverse deps: adr-00000000-c3-adoption, c3-1, c3-2
blocks: Abstract Constraints ✓, Containers ✓, Goal ✓

## c3-1 — Go CLI (container)
context: c3-0
reverse deps: c3-101, c3-102, c3-103, c3-104, c3-105, c3-110, c3-111, c3-112, c3-113, c3-114, c3-115, c3-116
constraints from: c3-0
blocks: Complexity Assessment ✓, Components ✓, Goal ✓, Responsibilities ✓

## c3-101 — frontmatter (component)
container: c3-1 | context: c3-0
files: cli/internal/frontmatter/**, cli/internal/markdown/**
constraints from: c3-0, c3-1
blocks: Container Connection ✓, Dependencies ✓, Goal ✓, Related Refs ✓

## c3-102 — walker (component)
container: c3-1 | context: c3-0
files: cli/internal/walker/**
constraints from: c3-0, c3-1
blocks: Container Connection ✓, Dependencies ✓, Goal ✓, Related Refs ○

## c3-103 — templates (component)
container: c3-1 | context: c3-0
files: cli/internal/templates/**, cli/templates/**
constraints from: c3-0, c3-1
blocks: Container Connection ✓, Dependencies ✓, Goal ✓, Related Refs ✓

## c3-104 — wiring (component)
container: c3-1 | context: c3-0
files: cli/internal/wiring/**, cli/cmd/wire.go
constraints from: c3-0, c3-1
blocks: Container Connection ✓, Dependencies ✓, Goal ✓, Related Refs ○

## c3-105 — codemap-lib (component)
container: c3-1 | context: c3-0
files: cli/internal/codemap/**
constraints from: c3-0, c3-1
blocks: Container Connection ✓, Dependencies ✓, Goal ✓, Related Refs ○

## c3-110 — init-cmd (component)
container: c3-1 | context: c3-0
files: cli/cmd/init.go, cli/cmd/init_test.go
constraints from: c3-0, c3-1
blocks: Container Connection ○, Dependencies ✓, Goal ✓, Related Refs ○

## c3-111 — add-cmd (component)
container: c3-1 | context: c3-0
files: cli/cmd/add.go, cli/cmd/add_rich.go, cli/cmd/add_test.go, cli/cmd/add_rich_test.go, cli/internal/numbering/**, cli/internal/writer/**
constraints from: c3-0, c3-1
blocks: Container Connection ○, Dependencies ✓, Goal ✓, Related Refs ○

## c3-112 — list-cmd (component)
container: c3-1 | context: c3-0
files: cli/cmd/list.go, cli/cmd/list_test.go
constraints from: c3-0, c3-1
blocks: Container Connection ○, Dependencies ✓, Goal ✓, Related Refs ○

## c3-113 — check-cmd (component)
container: c3-1 | context: c3-0
reverse deps: adr-20260312-upgrade-refs-golden-patterns
files: cli/cmd/check_enhanced.go, cli/cmd/check_enhanced_test.go, cli/internal/schema/**, cli/internal/index/**
constraints from: c3-0, c3-1
blocks: Container Connection ○, Dependencies ✓, Goal ✓, Related Refs ○

## c3-114 — lookup-cmd (component)
container: c3-1 | context: c3-0
files: cli/cmd/lookup.go, cli/cmd/lookup_test.go
constraints from: c3-0, c3-1
blocks: Container Connection ○, Dependencies ✓, Goal ✓, Related Refs ○

## c3-115 — codemap-cmd (component)
container: c3-1 | context: c3-0
files: cli/cmd/codemap.go
constraints from: c3-0, c3-1
blocks: Container Connection ○, Dependencies ✓, Goal ✓, Related Refs ○

## c3-116 — coverage-cmd (component)
container: c3-1 | context: c3-0
reverse deps: adr-20260312-upgrade-refs-golden-patterns
files: cli/cmd/coverage.go, cli/cmd/coverage_test.go
constraints from: c3-0, c3-1
blocks: Container Connection ○, Dependencies ✓, Goal ✓, Related Refs ○

## c3-2 — Claude Skill (container)
context: c3-0
reverse deps: c3-201, c3-210
constraints from: c3-0
blocks: Complexity Assessment ✓, Components ✓, Goal ✓, Responsibilities ✓

## c3-201 — skill-router (component)
container: c3-2 | context: c3-0
files: skills/c3/SKILL.md
constraints from: c3-0, c3-2
blocks: Container Connection ✓, Dependencies ✓, Goal ✓, Related Refs ○

## c3-210 — operation-refs (component)
container: c3-2 | context: c3-0
reverse deps: adr-20260312-upgrade-refs-golden-patterns
files: skills/c3/references/**
constraints from: c3-0, c3-2
blocks: Container Connection ○, Dependencies ✓, Goal ✓, Related Refs ○

## ref-cross-compiled-binary — Cross-Compiled Binary Distribution (ref)
files: scripts/build.sh, skills/c3/bin/c3x.sh, skills/c3/bin/VERSION, .github/workflows/distribute.yml
blocks: Choice ✓, Goal ✓, How ✓, Why ✓

## ref-embedded-templates — Embedded Templates via go:embed (ref)
files: cli/internal/templates/**, cli/templates/**
blocks: Choice ✓, Goal ✓, How ✓, Why ✓

## ref-frontmatter-docs — Frontmatter Docs Pattern (ref)
files: cli/internal/frontmatter/**, cli/internal/config/**
blocks: Choice ✓, Goal ✓, How ✓, Why ✓

## File Map
.github/workflows/distribute.yml → ref-cross-compiled-binary
cli/cmd/add.go → c3-111
cli/cmd/add_rich.go → c3-111
cli/cmd/add_rich_test.go → c3-111
cli/cmd/add_test.go → c3-111
cli/cmd/check_enhanced.go → c3-113
cli/cmd/check_enhanced_test.go → c3-113
cli/cmd/codemap.go → c3-115
cli/cmd/coverage.go → c3-116
cli/cmd/coverage_test.go → c3-116
cli/cmd/init.go → c3-110
cli/cmd/init_test.go → c3-110
cli/cmd/list.go → c3-112
cli/cmd/list_test.go → c3-112
cli/cmd/lookup.go → c3-114
cli/cmd/lookup_test.go → c3-114
cli/cmd/wire.go → c3-104
cli/internal/codemap/** → c3-105
cli/internal/config/** → ref-frontmatter-docs
cli/internal/frontmatter/** → c3-101, ref-frontmatter-docs
cli/internal/index/** → c3-113
cli/internal/markdown/** → c3-101
cli/internal/numbering/** → c3-111
cli/internal/schema/** → c3-113
cli/internal/templates/** → c3-103, ref-embedded-templates
cli/internal/walker/** → c3-102
cli/internal/wiring/** → c3-104
cli/internal/writer/** → c3-111
cli/templates/** → c3-103, ref-embedded-templates
scripts/build.sh → ref-cross-compiled-binary
skills/c3/SKILL.md → c3-201
skills/c3/bin/VERSION → ref-cross-compiled-binary
skills/c3/bin/c3x.sh → ref-cross-compiled-binary
skills/c3/references/** → c3-210

## Ref Map
ref-cross-compiled-binary
ref-embedded-templates
ref-frontmatter-docs
