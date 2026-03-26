# Changelog

All notable changes to the C3 Skill plugin will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [8.0.7] - 2026-03-26

### Added

- **`c3x set <id> codemap` support** — update code-map patterns via CLI instead of going through the database directly. Replace all (`"a,b"`), append (`--append`), remove (`--remove`), clear (`""`), and batch mode (`"codemap": [...]` in `--stdin` JSON).
- **Batch mode codemap clearing** — `{"codemap": []}` now correctly clears all patterns (previously ignored empty arrays).
- **`--append`/`--remove` mutual exclusion** — using both flags together now returns a clear error.

## [8.0.6] - 2026-03-26

### Fixed

- **Migration warnings go to stdout, not stderr** — all c3x output (warnings + JSON) goes to stdout. With `--json`, parse failure warnings appear as text lines before the JSON blob. Removed incorrect `2>&1` and stderr references throughout migrate.md.
- **Tightened migration reference** — 401→215 lines. Runbook tone, no hand-holding. Markdown is source of truth framing scoped correctly (before migration; after migration, database is authoritative).
- **Phase B export verification** — B1 now verifies export file count matches entity count. B5 restores FTS5 search check.

## [8.0.5] - 2026-03-26

### Fixed

- **Migration reference hardened against silent data loss** — complete rewrite of `references/migrate.md` (98→401 lines). Now version-aware with two distinct paths: v6→v7 (`migrate-legacy`) and v7→v8 (`migrate`). Adds evidence gates at every phase, recovery table for 6 failure scenarios, and "warnings are errors" enforcement throughout.
- **Correct CLI command routing in migration docs** — Phase A now uses `c3x migrate-legacy` (not `c3x migrate`), and documents that `c3x check` is unavailable before database exists.
- **`--keep-originals` enforced during migration** — source `.md` files preserved until post-migration verification passes, preventing irrecoverable data loss on migration failure.
- **SKILL.md command table** — now lists both `migrate` and `migrate-legacy` with correct descriptions and version requirements.

## [8.0.4] - 2026-03-24

### Added

- **Help entries for all new commands** — `nodes`, `hash`, `versions`, `version`, `prune` now appear in `c3x --help` with full usage docs
- **Updated `migrate` help** — reflects node tree migration (not legacy file import)
- **`migrate-legacy` command** — the old file-based migration is now a separate hidden command

### Changed

- **README** — updated read/write descriptions, "Why" section mentions node trees and element-level tracking, content database section fully rewritten

## [8.0.3] - 2026-03-24

### Fixed

- **Strip stale frontmatter from body during migration** — old v7 entities stored YAML frontmatter (`goal:`, `status:`, `parent:`, etc.) inside the body text. `WriteEntity` now strips `---` delimited blocks and leading YAML lines before parsing into nodes. Verified on sft (26 entities) and remmd (39 entities).

## [8.0.2] - 2026-03-24

### Fixed

- **`c3x migrate` output is actionable** — distinguishes "already have nodes (ok)" from "no content yet" with `c3x write <id>` guidance. No more opaque "skipped N".

## [8.0.1] - 2026-03-24

### Fixed

- **`c3x migrate` reads legacy body column via raw SQL** — the v8.0.0 migration command couldn't read the old `body` column since it was removed from the Entity struct. Now uses direct SQL query to read pre-v8 body content for node tree population.

## [8.0.0] - 2026-03-24

### BREAKING

- **Content Database** — entity content is now stored as an element-level node tree in SQLite. Every heading, paragraph, list item, table row, and code block has its own ID and SHA256 content hash. The `Body`, `Summary`, and `Description` fields have been removed from the Entity struct.
- **`c3x migrate`** now populates the node tree (was `migrate-v2`). The old file-based migration is `c3x migrate-legacy`.
- **FTS on entities** trimmed to `title` + `goal` only. Content search now uses `content_fts` over the node tree.

### Added

- **Node tree storage** — `nodes` table with element-level content decomposition via goldmark AST parser. Parent-child relationships, sequence ordering, per-node content hashing.
- **Version history** — `versions` table stores full content snapshots on every write. `c3x versions`, `c3x version <n>`, `c3x prune --keep <n>`.
- **Content hashing** — per-node SHA256 hashes + entity-level root merkle. `c3x hash` with `--recompute` for integrity checks.
- **`c3x nodes <entity-id>`** — inspect the content tree with IDs, types, hashes. JSON and text modes.
- **`c3x hash <entity-id>`** — root merkle hash with optional recompute/drift detection.
- **`c3x versions <entity-id>`** — version history with timestamps and commit marks.
- **`c3x version <entity-id> <n>`** — retrieve content at a specific version.
- **`c3x prune <entity-id> --keep <n>`** — prune old versions, preserving git-marked ones.
- **Content-level FTS** — `c3x query` now searches both entity metadata and node content, merging results.
- **`content` package** — `ParseMarkdown` (goldmark AST to nodes), `RenderMarkdown` (nodes to markdown), `WriteEntity`/`ReadEntity` bridge layer with transactional node insertion.

### Changed

- **Write path** — `write`, `set`, `wire` commands now route through the node tree via `content.WriteEntity`. Each write creates a version snapshot and updates the root merkle.
- **Read path** — `read`, `export`, `check` commands reconstruct content from the node tree via `content.ReadEntity`.
- **Schema** — `entities_fts` indexes only `title` and `goal`. `content_fts` indexes node content with auto-sync triggers.

### Removed

- `Entity.Body`, `Entity.Summary`, `Entity.Description` fields and corresponding database columns
- `internal/writer/` package (dead code, zero imports)
- `internal/wiring/` package (dead code, zero imports)
- `chunks` table (replaced by `nodes`)
- `truncateForLog` helper (no body field to truncate)

## [7.0.4] - 2026-03-23

### Fixed

- **Body corruption on `read | write` roundtrip** — templates stored YAML frontmatter in `entity.Body`; `read` prepended another frontmatter block; each cycle nested another `---` block. Now stripped at insert time
- **Validation catch-22** — `set --section` and `write --section` validated ALL required sections, blocking incremental filling. Section-level updates now skip full-document validation
- **Template comment bloat** — HTML comment blocks (40-50 lines per template) stored in Body and returned on every `read`. Stripped at insert time; `buildDocument` path already clean
- **Changelog body bloat** — `UpdateEntity` logged full old+new body text in changelog. Now truncated to 200 chars

### Added

- **`read --section <name>`** — extract a single section's content (text or JSON mode) without reading the full entity body
- **`set --stdin`** batch mode — pipe `{"fields":{...},"sections":{...}}` JSON to update multiple fields and sections in one call
- **`wire` multiple targets** — `c3x wire c3-101 ref-jwt ref-error-handling` wires to multiple targets in a single invocation
- **`add --json` returns sections** — response now includes `type` and `sections[]` from schema, eliminating the follow-up `read` call
- **`list --json --compact`** — lightweight JSON output (id, type, title, parent, status) skipping per-entity relationship and codemap queries
- **Compact JSON in agent mode** — `C3X_MODE=agent` now uses `json.Marshal` (no indentation) instead of `json.MarshalIndent`

### Removed

- Orphaned `cli/templates/` directory (6 files) — source of truth is `cli/internal/templates/` via `go:embed`

## [7.0.3] - 2026-03-23

### Changed

- **Add commands are DB-only** — `c3x add container` and `c3x add component` no longer write backward-compat `.md` files to disk; all entity data lives exclusively in SQLite
- **`c3x init` creates only `c3.db`** — removed vestigial `config.yaml` write and dead legacy markdown scaffolding (`RunInit`)
- **Removed `AddResult.Path` from JSON output** — field was semantically broken after disk write removal; `--json` add now returns `{"id": "..."}` only
- **Trimmed SKILL.md description** to 460 chars (was 4372) — fixes "exceeds maximum length of 1024" error in some plugin loaders

### Fixed

- **Midnight race in ADR ID generation** — `addAdr` and `addRichAdr` called `time.Now()` twice (once for ID, once for date field); could produce mismatched dates at midnight boundary. Now captures a single instant
- **Inline `regexp.MustCompile` in `addComponent`** — hoisted to package-level `reContainer` var, consistent with `validSlug` pattern
- **Removed legacy format detection** — `hasMarkdownFiles()`, `runLegacyCheck()`, and the legacy block in `main.go` removed; no-DB case now returns a clean error pointing to `c3x init` or `c3x migrate`

## [7.0.2] - 2026-03-23

### Changed

- **Codemap is DB-only** — `c3x codemap` no longer writes `.c3/code-map.yaml`; all code-map data lives exclusively in the SQLite store. `c3x export` still produces the YAML file when needed.
- Removed no-op `SetCodeMap` calls for empty scaffolds — eliminates unnecessary DB transactions

### Fixed

- Skill docs enforced c3x-only access — bare `Read`/`Glob`/`Grep` on `.c3/` files replaced with `c3x` commands throughout all operation references
- Updated stale hint text referencing `code-map.yaml` in validation output

## [7.0.0] - 2026-03-20

### Breaking

- Architecture data now stored in embedded SQLite (`c3.db`) instead of raw markdown files
- Existing file-based `.c3/` directories must be migrated with `c3x migrate`
- `write` and `set --section` enforce schema validation — incomplete content is rejected

### Added

- **Embedded SQLite store** — entities, relationships, code-map, and changelog in a single file
- **8 new commands** — `query`, `diff`, `impact`, `export`, `graph`, `read`, `write`, `marketplace`
- **Rules** — new entity type for enforceable coding standards with golden examples
- **Marketplace** — browse and install community rule collections from git repos
- **Migrate dry-run** — `c3x migrate --dry-run` reports all quality gaps before committing
- **Schema enforcement** — `write` and `set --section` reject missing/empty required sections
- **Goal auto-promotion** — body `## Goal` content auto-fills frontmatter `goal:` on write
- **Full-text search** — `c3x query` with BM25 ranking across titles, goals, bodies
- **Impact analysis** — `c3x impact <id>` finds all transitively affected entities
- **Changelog tracking** — `c3x diff` shows mutations; `--mark` stamps with commit hash

### Changed

- All commands rewired from file-based walker to `*store.Store`
- `main.go` refactored into testable `run()` function
- Relationship sync on `write` diffs and removes stale edges
- Test coverage: 73.4% → 89.0% across 15 packages

### Fixed

- Non-atomic marketplace registration — rename cache before registering source
- Anchor-stripping consistency across migrate and write paths
- TTY detection on `c3x write` — errors early instead of hanging

## [6.12.1] - 2026-03-18

### Fixed
- **`uses:` frontmatter field**: CLI now reads `uses:` as the canonical field (matching skill docs and all `.c3/` files). `refs:` still accepted for backward compat with dedup merge when both present. Fixes 93 false warnings from `c3x check` on projects following the docs. (#26)
- **User-facing output migrated**: All CLI output (`list`, `graph`, `lookup`, `check`), JSON tags, help text, and error messages now use `uses:` consistently

### Removed
- **Historical binaries purged**: 44 cross-compiled Go binaries removed from git history via `git filter-repo`, significantly reducing clone size

## [6.12.0] - 2026-03-17

### Added
- **`@c3x/cli` npm package**: Thin Node.js CLI (`npx @c3x/cli`) that discovers installed c3x Go binaries across Claude/Codex skill paths and marketplace installations, picks the highest version, and delegates. Humans get text output; agents get JSON.
- **`--agent` flag**: Restrict binary discovery to a specific agent type (`--agent claude` or `--agent codex`). Project scope is always included.
- **`C3X_MODE` env var**: Go binary respects `C3X_MODE=agent` to output JSON by default for commands that support it. Explicit `--json`/`--compact` flags override.
- **Automated npm publishing**: CI publishes `@c3x/cli` to npm alongside GitHub releases on version bumps.

## [6.11.1] - 2026-03-17

### Fixed
- **Onboard CLAUDE.md injection**: Removed dev-only `CLI: bash skills/c3/bin/c3x.sh` path that broke c3x resolution in installed plugins. The skill resolves the binary path via `<skill-dir>` at runtime.

## [6.11.0] - 2026-03-17 [YANKED]

### Added
- **`c3x capabilities` command**: Emits a markdown table of all non-hidden CLI commands. Single source of truth for feature documentation — onboard and README both consume this instead of maintaining separate lists.

### Changed
- **Registry-driven command metadata**: `help.go` refactored from hardcoded strings to a `[]CommandMeta` registry. Both `c3x --help` and `c3x capabilities` render from the same data. Adding a new command means adding one struct entry.
- **Onboard post-reveal**: Replaced static capabilities table with a flow-based introduction (understand → change → validate → explore) and a pointer to `c3x capabilities` for self-discovery.

## [6.10.2] - 2026-03-17

### Changed
- **Single VERSION source of truth**: Consolidated from root `VERSION` + `skills/c3/bin/VERSION` to just `skills/c3/bin/VERSION`. CI workflows (`release.yml`, `distribute.yml`), `build.sh`, and `/release` command all read from the same file now.

### Removed
- **Root `VERSION` file**: Eliminated redundant version file. `skills/c3/bin/VERSION` is the sole source of truth for version detection.

## [6.10.1] - 2026-03-16

### Changed
- **CLI surface rationalized to 12 LLM-visible commands**: Adversarial triage (triage-three) evaluated all 14 commands across mergeability, breakage risk, and LLM cognitive load. Result: single-purpose verbs beat merged flag-heavy commands for LLM use.
- **`unwire` merged into `wire --remove`**: Symmetric pair collapsed. `unwire` remains as hidden backward-compat alias.
- **`graph` demoted from LLM surface**: `list --json` now carries all relationship data (refs, affects, scope, files), making `graph --json` redundant for LLM workflows. `graph` stays in CLI for `--format mermaid` diagrams.
- **`list --json` enriched**: Now includes `files` from code-map and `refs`/`affects`/`scope` arrays in frontmatter — one call gives the LLM everything it needs.
- **`wire`/`unwire` cite now optional**: `wire <src> <tgt>` works (cite is the default). 3-arg form still accepted.
- **`set --section` format detection**: Uses actual JSON parse instead of `[` prefix sniffing — plain text starting with `[` no longer misroutes to JSON table mode.
- **Entity Types in help**: "context" removed from addable types, noted as "(created by init)".
- **ADR date format unified**: `add_rich.go` now uses ISO `2006-01-02` matching `add.go` (was `20060102`).

### Removed
- **`_index/notes/` sunset**: Notes validation removed from `c3x check`. Cross-cutting concern traces are now handled exclusively by `recipes/`. Existing notes should be migrated to recipes via onboard.

### Added
- **`add --json`**: Returns `{"id":"c3-101","path":"..."}` for programmatic entity creation workflows.

## [6.10.0] - 2026-03-12

### Added
- **Ref governance metric**: `c3x coverage` now reports what percentage of components are governed by at least one ref, with an ungoverned components list. Appears in both JSON (`ref_governance` field) and human-readable output.
- **Scope cross-check**: `c3x check` warns when a ref scopes a container but a child component doesn't cite that ref (e.g. "ref-jwt scopes c3-1 but c3-110 does not cite it").
- **Ref quality rubric**: Ref template now includes a 7-criteria quality rubric (compliance questions, mechanism over outcome, violation examples, scope grounding, brevity, dependency visibility, single compliance path).
- **Ref compliance gate (Phase 3b)**: Change workflow now includes an adversarial ref compliance check before audit — for each file touched, lookup applicable refs and verify compliance with structured verdict output.
- **Ref compliance audit (Phase 7b)**: Audit workflow now spot-checks code against golden patterns in ref `## How` sections, with quality check for pattern actionability.

### Changed
- **Discovery-first ref creation**: Ref Add flow rewritten to discover existing implementations before drafting — `Scaffold → Discover → Fill → Usage → Cite → ADR`. Includes quality gate (must be able to derive YES/NO compliance questions from `## How`).
- **Format-flexible `## How`**: Ref template `## How` section no longer prescribes a table — supports code blocks, do/don't pairs, checklists, or tables. The test: can a reviewer check compliance in under 10 seconds?
- **Dual-purpose `## Not This`**: Ref template `## Not This` now serves both rejected alternatives and concrete anti-examples.
- **Schema purpose**: `How` section purpose updated to "Golden pattern — prescriptive examples and implementation guidance".
- **Codemap gaps fixed**: `cli/internal/schema/**` and `cli/internal/index/**` now mapped to c3-113 (check-cmd).

### Fixed
- **`ref_id` column validation**: `c3x check` now validates `ref_id` typed columns in Related Refs tables, with `--fix` auto-correcting bad references via title matching.
- **CI**: 5 fixes — PR merge branching, YAML heredoc parsing, distribute branch conflicts, release step, merge strategy.

## [6.9.0] - 2026-03-11

### Added
- **CoT Harness**: Context-led reasoning reflex in SKILL.md — before touching any file, `c3x lookup` it and follow what C3 knows. Re-enters when context shifts mid-task. Replaces assumptions with topology-driven decisions.
- **Frontmatter examples for audit + ref**: Skill description now covers all 6 operations (was 4), improving trigger reliability for audit and ref invocations.

### Changed
- **CLAUDE.md Injection + Capabilities Reveal moved to `references/onboard.md`**: These onboard-specific blocks no longer live in the main SKILL.md — keeps the skill file focused on dispatch and reasoning.

## [6.8.0] - 2026-03-05

### Added
- **`--include-adr` flag**: `c3x list` and `c3x check` now exclude ADRs by default. Use `--include-adr` to include them in output and validation. ADRs are ephemeral work orders — they drive changes, then stay out of the way.
- **Lightweight ADR template**: `c3x add adr <slug>` now creates a minimal template (Goal, Work Breakdown, Risks) instead of the heavy onboarding template. Fast to create, drives the change, throwaway.
- **11 new tests**: Full coverage for ADR filtering in list (topology, flat, JSON), check (skip/include), and `--include-adr` flag parsing.

### Changed
- **ADR-first enforcement**: Change operations now require `c3x add adr <slug>` as their first action (HARD RULE in skill instructions). No code reads, no file edits, no exploration before the ADR exists.
- **Change flow reordered**: `ADR → Understand → Approve → Execute → Audit` (was `Understand → ADR → ...`). ADR creation is Phase 1, non-negotiable.
- **Audit Phase 6 opt-in**: ADR Lifecycle audit only runs with `--include-adr`, since ADRs are ephemeral.
- **Ref-add exception documented**: Ref operations create their adoption ADR at completion (not upfront), clearly noted in shared rules.

## [6.7.0] - 2026-03-04

### Added
- **Recipe entity type**: Cross-cutting concern traces that link entity sections into end-to-end narratives. Created via `c3x add recipe <topic>`, stored in `.c3/recipes/recipe-*.md`. Includes `description` and `sources` frontmatter for LLM-driven query matching. Validated by `c3x check` (source references must resolve). Shown in `c3x list` topology and JSON output.

### Fixed
- **Walker skips `_index/`**: `WalkC3Docs` now skips `.c3/_index/` directory, preventing index files from being parsed as entities (#7).
- **Frontmatter delimiter consistency**: `parseNoteSources` aligned to use `\n---\n` delimiter matching (was `\n---`) with EOF edge case handling (#8).

### Changed
- **Code cleanup**: Removed AI-generated slop — unnecessary godoc on unexported functions, restating comments, single-use variables inlined, dead if/else branch collapsed.

## [6.6.0] - 2026-03-04

### Added
- **Structural index** (`cli/internal/index/`): Precomputed entity→files→refs→reverse-deps→constraints index at `.c3/_index/structural.md`. Gives LLMs instant architecture discovery without multiple CLI calls. Auto-rebuilt after mutating commands (`add`, `set`, `wire`, `unwire`).
- **Note health checking**: `c3x check` validates that topic notes in `.c3/_index/notes/` reference entities that exist in the graph. Orphaned source citations reported as warnings with actionable hints.
- **Design doc**: `docs/plans/2026-03-04-knowledge-index-design.md` — two-layer knowledge index design (structural index + topic notes).

### Changed
- **Index rebuild scoped to mutations**: Structural index only rebuilds after `add`, `set`, `wire`, `unwire` — read-only commands (`list`, `check`, `lookup`) skip the rebuild for faster execution.
- **`parseNoteSources` simplified**: Replaced hand-rolled YAML line parser with `yaml.Unmarshal`.
- **File map reuses computed refs**: Index build no longer re-traverses the graph for refs already computed in entity entries.

### Removed
- **`c3x query` command**: Eliminated — structural index + direct file reads replace the routing and block extraction that `query` provided.
- **`cli/internal/blocks/` package**: Block extraction logic inlined into the index package; standalone package removed.

## [6.5.3] - 2026-03-03

### Added
- **`c3x query` command**: Extract structured blocks from C3 entities. Modes: catalog (all entities with fill status), snapshot (all blocks for one entity), single block, chain walk (component → container → context + refs), and file resolution (file path → entity → snapshot). Supports `--chain` and `--json` flags.
- **`check` output quality**: Summary header (`Checked N docs — all clear` or `Checked N docs — X errors, Y warnings`), actionable hint lines below each issue (`→ fix suggestion`), and a legend footer. JSON output now includes `hint` field on issues.

### Changed
- **Schema extracted to `internal/schema`**: Schema registry moved from `cmd/schema.go` to a standalone internal package, enabling reuse across `add`, `check`, and `query` commands.
- **`writeJSON` helper**: Consolidated 5 duplicate `json.MarshalIndent` + `Fprintln` patterns into a single `writeJSON` helper in `helpers.go`.

## [6.5.2] - 2026-02-27

### Fixed
- **CI: stale binaries accumulating on main**: Each release added new versioned binaries without removing old ones, doubling the plugin zip size. The distribute workflow now `git rm`s previous version binaries before adding new ones.
- **Nonexistent `c3x wire` command in onboard reference**: `onboard.md` troubleshooting table referenced `c3x wire <id> cite <ref-id>` which doesn't match the skill's CLI reference. Replaced with correct instruction to edit component frontmatter directly.
- **Ambiguous ref code-map audit rule**: `audit.md` Phase 9 flagged any ref with a code-map entry as a violation, but `c3x codemap` scaffolds empty stubs for refs by default. Clarified that scaffold stubs are acceptable — only filled-in file patterns are violations.

## [6.5.1] - 2026-02-27

### Fixed
- **`c3x.sh` cleanup destroying cross-compiled binaries**: When run from the source directory (where all 4 platform binaries exist), the stale-version cleanup loop deleted binaries for other platforms. Now detects multi-binary source dirs and skips cleanup.
- **Broken YAML frontmatter silently dropped entities**: Files with `---` delimiters but invalid YAML (e.g. unquoted `via:` colon-space in values) were silently excluded from the entity graph. `c3x check` now reports these as errors (`✗`) with a hint to check for unquoted colons.

### Changed
- **SKILL.md**: Added shared rule to run `c3x check` frequently after creating/editing `.c3/` docs. Check now catches broken YAML, missing sections, bad entity references, and codemap issues.

### Documentation
- **README.md**: Added Layer 0 (Parse) to the validation table — broken YAML frontmatter detection.
- **CLAUDE.md**: Added Architecture block pointing to `.c3/` directory and `/c3` skill.

## [6.5.0] - 2026-02-27

### Added
- **`c3x codemap` command**: Scaffold or update `.c3/code-map.yaml` with stubs for every component and ref in the C3 graph. Idempotent — existing patterns are preserved, missing entries get empty lists. Output groups components then refs with commented `_exclude` example. JSON by default, human-readable with `HUMAN=1`.
- **Versioned binary naming**: Binaries are now named `c3x-{version}-{os}-{arch}` (e.g. `c3x-6.5.0-linux-amd64`). A `VERSION` file alongside the wrapper tells `c3x.sh` which binary to use.
- **Stale binary cleanup**: `c3x.sh` removes binaries from previous versions on every invocation, eliminating the caching issue where plugin updates left old binaries in place.
- **Refs in code-map**: Refs (e.g. `ref-jwt`) can now have code-map entries alongside components. Useful for refs that have concrete implementation files (shared middleware, utility libraries).

### Changed
- **`c3x.sh` simplified**: Removed download fallback — binaries are always bundled with the plugin release, so network fetching was unnecessary dead code.
- **Onboarding flow**: Stage 2 now starts with `c3x codemap` scaffold step before structural checks. Gate 2 checklist includes code-map coverage as a requirement.
- **`lookup` hint**: When `code-map.yaml` is empty or missing, `c3x lookup` now prints a hint to run `c3x codemap`.

### Fixed
- **Refs no longer warned in code-map validation**: `validate.go` previously flagged any non-component ID. Now only warns for `container`, `context`, and `adr` types in the codemap.

## [6.4.0] - 2026-02-27

### Added
- **`c3x lookup` command**: Map any file path or glob pattern to its owning component(s) and governing refs. Single file returns component details + cited refs with goals. Glob pattern expands against project and shows file-to-component map.
- **`c3x coverage` command**: Code-map coverage statistics — shows how many project files are mapped, excluded, or unmapped. Uses `git ls-files` for fast file discovery with filesystem walk fallback. JSON output by default (human-readable with `HUMAN=1`).
- **`_exclude` key in code-map.yaml**: Mark files as intentionally unmapped (tests, build output, configs). Excluded files don't count against coverage percentage. Formula: `mapped / (total - excluded)`.
- **Bracket path support**: Literal `[id]`, `[...slug]` paths (Next.js, SvelteKit route params) now work in code-map patterns. Double-try matching: glob interpretation first, then escaped brackets as fallback.
- **`--compact` flag for `c3x list`**: Goals-only topology tree without file/ref detail.
- **`doublestar/v4` dependency**: Powers `**` glob matching for code-map patterns.

### Changed
- **`c3x list` refactored**: Now accepts `ListOptions` struct with `C3Dir` for code-map integration. Shows file coverage and ref usage in topology output.
- **Code-map validation**: `isGlobPattern` excludes brackets from glob detection — `[id]` paths are treated as literal, not character classes.
- **Skill docs streamlined**: All 6 reference docs (onboard, query, audit, change, ref, sweep) trimmed for conciseness while adding coverage/lookup guidance.

### Fixed
- **`.c3/` files excluded from coverage**: `git ls-files` output now filtered by `skipPrefixes` (`.c3/`, `.git/`, `node_modules/`, `dist/`), preventing architecture docs from inflating file counts.
- **Coverage percentage denominator**: Uses `(total - excluded)` not `total` — `_exclude` patterns don't penalize the coverage score.
- **`c3x list` swallows code-map parse errors**: Now propagates malformed `code-map.yaml` errors instead of silently using empty map.
- **`c3x lookup` with no argument**: Now exits with usage error instead of silently returning "no mapping found".
- **JSON null for empty unmapped files**: `unmapped_files` is always `[]` in JSON output, never `null`.
- **Duplicate `isGlobPattern`**: Exported as `IsGlobPattern` from codemap package, removed duplicate in cmd/lookup.go.

## [6.3.0] - 2026-02-26

### Changed
- **CI pipeline hardened**: `distribute.yml` now does `git pull --rebase` before pushing to main, preventing push failures when manual main commits race with CI runs.
- **Binary integrity**: Go binaries marked as `binary` in `.gitattributes` — prevents line-ending corruption on Windows CI runners.
- **Dead code removed**: Removed v1 `RunCheck()` and its tests; `Issue`/`CheckResult` types consolidated into `check_enhanced.go`. Removed unused `internal/output` package.
- **Node.js residues removed**: Deleted `package.json`, removed `node_modules/` from `.gitignore`, removed stale pre-2026 design docs (28 files).

## [6.2.0] - 2026-02-26

### Added
- **`.c3/code-map.yaml`**: New centralized file mapping component IDs to their source files. Replaces the `## Code References` section in component docs — keeps architecture reasoning in docs, code pointers in a machine-managed file.
- **`c3x check` validates code-map**: Validates entity IDs exist in graph, IDs are components (not refs/containers), file paths exist on disk, no empty/absolute/traversal paths.
- **`CheckOptions.C3Dir` field**: Explicit c3 directory path in check options — no longer reconstructed from ProjectDir, supports custom `--c3-dir` usage.

### Changed
- **Removed `## Code References` section**: Component docs no longer contain file path tables. File-to-component linkage moves to `.c3/code-map.yaml`.
- **Removed `## Cited By` section**: Ref docs no longer require manual "Cited By" tables — citation tracking is derived from `refs:` frontmatter on components. Bidirectional consistency check removed from `c3x check`.
- **Component type validation**: `code-map.yaml` validation now checks entity type from the graph (not a regex) — prevents false-positives on container IDs with 3+ digits.
- **Skill docs updated**: All `skills/c3/references/` docs updated to reference `.c3/code-map.yaml` instead of `## Code References`.

### Fixed
- **Path validation in code-map**: Rejects empty paths, absolute paths, `../` traversal, and directory paths (must be regular files).

## [6.1.0] - 2026-02-26

### Breaking Changes
- **Go CLI replaces npm package**: `c3x` is now a native Go binary (cross-compiled for linux/darwin × amd64/arm64), replacing the previous npm-based CLI
- **Unified `/c3` skill**: All separate skills (c3-onboard, c3-query, c3-ref, c3-audit, c3-change) consolidated into a single `c3` skill with intent router

### Added
- **CLI v2 structured document engine**: Three-layer architecture for C3 markdown documents
  - Layer 1: Controlled frontmatter mutations + atomic 3-sided cite wiring
  - Layer 2: Section/table schema registry with known sections per entity type
  - Layer 3: Typed column validation (filepath, entity_id, ref_id, enum)
- **New commands**: `set`, `wire`, `unwire`, `schema`
  - `set`: Update frontmatter fields and section content (text or JSON tables)
  - `wire`/`unwire`: Create/remove cite relationships with atomic 3-sided updates
  - `schema`: Show known sections and column types per entity type
- **Enhanced `add` command**: Rich content scaffolding with `--goal`, `--summary`, `--boundary` flags, auto-wiring to parent containers
- **Enhanced `check` command**: 3-layer validation — broken links/orphans, required sections per schema, code refs + entity IDs + cite consistency
- **Auto-download binary**: `c3x.sh` downloads pre-built binary from GitHub releases when not found locally — no Go toolchain required
- **Internal packages**: `markdown` (section/table parser), `writer` (frontmatter mutations), `schema` (registry with typed columns)

### Changed
- **Binary distribution**: Binaries built in CI, force-added to `main` branch, included in release zips. `dev` branch has source only.
- **Help text**: Compact, agent-browser style help for all 8 commands

### Fixed
- **CI workflow**: Merged release into distribute workflow for single-pipeline release
- **npm package naming**: Resolved registry conflicts (c3x → c3-kit → @lagz0ne/c3x)

## [4.3.0] - 2026-02-11

### Changed
- **Renamed `living-entity` to `c3-sweep`**: All skills, agents, and commands now use the `c3-` prefix consistently. `living-entity` skill becomes `c3-sweep`, agents become `c3-sweep-lead`, `c3-sweep-container`, `c3-sweep-component`, `c3-sweep-ref`
- **Cross-skill routing**: c3-change and c3-audit now route to c3-sweep for impact assessment
- **Version sync**: All version files (plugin.json, marketplace.json, VERSION, package.json) synchronized

## [4.2.1] - 2026-02-11

### Fixed
- **Agent spawning broken**: Worker agent types in `living-entity-lead`, `living-entity-container`, and `c3-lead` used unprefixed names (e.g., `living-entity-container`) instead of fully-qualified names (`c3-skill:living-entity-container`), causing all worker agent spawns to fail silently

## [4.2.0] - 2026-02-10

### Added
- **Persistent entity agents**: c3-lead and living-entity-lead now spawn named, persistent agents per C3 entity (container, component, ref) that stay alive across phases and operations within a session
- **Entity roster pattern**: Leads check team config before spawning — reuse existing agents via `SendMessage` instead of creating duplicates
- **`/c3-change` command**: Slash command for easy invocation of architectural change workflow
- **`/living-entity` command**: Slash command for impact assessment (marked experimental)

### Changed
- **c3-lead agent**: Replaced disposable analyst/reviewer/implementer workers with persistent entity agents. Lead absorbs adversarial reviewer role (cross-entity challenge at synthesis). Phase 1 uses entity self-assessment, Phase 3 messages existing agents for implementation
- **living-entity-lead agent**: Updated to spawn persistent named entity agents with roster check pattern. Both leads share `c3-session` team name so entity agents are reused across assessment and implementation
- **c3-change skill**: Added explicit `## Execution` section with HARD RULE to spawn c3-lead as first action (fixes unreliable team startup)
- **living-entity skill**: Added explicit `## Execution` section with HARD RULE to spawn living-entity-lead as first action

## [4.1.0] - 2026-02-09

### Added
- **living-entity-lead agent**: Team lead for living-entity skill — orchestrates impact assessment via Agent Teams with adaptive `TeamCreate`/`Task` fallback
- **Golden code examples in refs**: Refs may now include canonical code snippets as prescriptive review standards

### Changed
- **living-entity skill**: Restructured from skill-as-orchestrator to Agent Teams pattern (skill describes team, lead agent orchestrates)
- **`## Code References` semantics**: Clarified as implementation indicator — present = implemented, absent on component = provisioned, never on refs
- **Ref scope**: Refs are scoped conventions (apply where stated), not global enforcement
- **c3-ref skill**: Trigger phrase tightened ("document this as a standard"), Mode: List now instructs Grep for counting citings, discovery relaxed to 2-3 Grep calls
- **c3-lead agent**: Added `TeamCreate` and `SendMessage` to tool list to match Agent Teams body text
- **living-entity worker agents**: Delegation attribution corrected to reference lead agent, `.c3/ref/` paths fixed to `.c3/refs/`

### Removed
- **experiments/living-entity/generated/**: Outdated standalone plugin export (superseded by main agents)

## [4.0.0] - 2026-02-09

### Breaking Changes
- **Architecture overhaul**: Reduced from 7 agents + 7 skills to 5 agents + 6 skills
- **Commands eliminated**: All slash commands (`/c3`, `/onboard`, `/query`, `/alter`, `/apply`) removed — routing is now entirely through skill descriptions
- **Hooks removed**: `hooks.json` and all hook scripts deleted — context propagation moved to CLAUDE.md
- **Legacy skills removed**: `c3` (router), `c3-alter`, `c3-provision`, `onboard` skills replaced by `c3-change`, `c3-onboard`, `c3-audit`
- **Legacy agents removed**: `c3-orchestrator`, `c3-analysis`, `c3-synthesizer`, `c3-summarizer`, `c3-content-classifier`, `c3-adr-transition`, `c3-dev` — replaced by `c3-lead` and `c3-navigator`
- **Legacy references removed**: `component-lifecycle.md`, `component-types.md`, `container-patterns.md`, `content-separation.md`, `implementation-guide.md`, `plan-template.md`, `v3-structure.md`
- **Templates simplified**: Removed `container-database.md`, `container-queue.md`, `container-service.md`, `external.md`, `external-aspect.md` — single container and component templates now

### Added
- **Living Entity**: Architecture-aware impact assessment system
  - `living-entity` skill: context-tier orchestrator that reads `.c3/` dynamically
  - `living-entity-container` agent: container-tier subagent that identifies affected components
  - `living-entity-component` agent: component-tier subagent that inspects code and enforces conventions
  - `living-entity-ref` agent: ref-tier subagent that validates reference pattern compliance
  - Four-layer constraint chain: code ownership → behavioral refs → relationships → ADR history
  - Tiered delegation: context → container → component + ref (parallel)
- **c3-lead agent**: Team lead for architectural changes — orchestrates 4-phase workflow (Understand → ADR → Execute → Audit)
- **c3-change skill**: ADR-first change workflow replacing c3-alter and c3-provision
- **c3-onboard skill**: Staged Socratic discovery replacing the old onboard skill
- **c3-audit skill**: Architecture documentation audit with bundled references
- **Self-contained skills**: Each skill bundles its own `references/` and `templates/` subdirectories
- **check-refs script**: `bun run check-refs` validates bundled references match shared source; `bun run fix-refs` auto-fixes drift
- **constraint-chain.md reference**: Documents the four-layer constraint model

### Changed
- **c3-navigator agent**: Streamlined description (3968 → 657 chars), added targeted examples for reliable routing
- **c3-query skill**: Simplified with bundled references, outcome-focused instructions
- **c3-ref skill**: Simplified with bundled references and templates
- **Build system**: Simplified `build.ts`, removed `build-toc.sh` and legacy scripts
- **Templates**: Goal-first structure throughout, simplified to essentials
- **Skill descriptions**: All use `<example>` blocks and `DO NOT use for` routing guards for reliable triggering

### Removed
- All slash commands (`commands/` directory)
- All hooks (`hooks/hooks.json`, gate/verifier/context-loader scripts)
- `src/opencode/plugin.ts` (OpenCode support)
- 13 legacy scripts (c3-gate, c3-verifier, c3-init, etc.)
- `build-toc.sh` and `pre-commit-toc` hook

### Documentation
- **CLAUDE.md**: Updated with skill development philosophy, build system docs, plugin structure checklist
- **Skill descriptions**: All under 1024-char limit with examples and routing exclusions

## [3.8.0] - 2026-01-29

### Added
- **c3-provision skill**: Architecture-first workflow for designing components before implementation
  - Supports "provision", "design", "plan", "envision", "architect" trigger words
  - Creates provisioned ADRs that can be promoted to accepted when ready to implement
  - Detects and promotes provisioned components in c3-alter
- **Context propagation via CLAUDE.md files**: New approach to ensure Claude loads architecture context
  - `/c3 apply` command generates CLAUDE.md files in Code Reference directories
  - Solves subagent context loss (CLAUDE.md loads at session start for all agents)
  - Phase 10 audit check verifies CLAUDE.md presence and freshness
  - Minimal pointer template with c3-generated block markers for safe regeneration
- **Provisioned ADR status**: New ADR lifecycle stage between draft and accepted
  - `provisioned-by` and `supersedes` fields for ADR linking
  - Component lifecycle reference documentation
- **Build system**: Self-contained skills build with Codex target
  - Generates standalone skill files for distribution
  - CI workflow documentation
- **Test framework**: Skill triggering test infrastructure
  - 18 test prompts covering all routing scenarios
  - Fixture creation script and test helpers

### Changed
- **c3-navigator**: Filters provisioned components by default, surfaces on explicit request
- **c3-transition**: Supports superseded status for provisioned ADRs
- **Audit Phase 10**: New Context Files check for CLAUDE.md validation
- **Proactive pattern awareness**: CLAUDE.md files now recommended over hooks (hooks fire too late)

### Fixed
- **Routing**: All 18 skill triggering tests pass
  - Ref keywords now prioritized over provision
  - Negative examples added to c3-navigator for ref-related prompts
  - Explicit keyword filter in c3-navigator
- **ADR lifecycle**: Correct flow with accepted before provisioned
- **Supersedes paths**: Fixed relative path from provisioned/ directory
- **Transition scoping**: Implements check now scoped to current ADR only

## [3.7.0] - 2026-01-26

### Added
- **c3-dev agent**: New TDD execution agent for implementing ADR-approved changes
  - Creates tasks per work item linked to ADR
  - Implements RED-GREEN TDD cycle with Socratic dialogue
  - Validates code against patterns via c3-patterns dispatch
  - Creates summary task with integrity check before ADR transition
  - Granular task states for parallel visibility: pending → in_progress → blocked → testing → implementing → completed
- **README.md**: Comprehensive plugin documentation with d2 diagrams via diashort
  - Agent ecosystem diagram
  - TDD workflow diagram
  - C3 structure diagram
  - Simple 4-step example walkthrough

### Changed
- **c3-orchestrator Phase 6**: Now dispatches c3-dev when user chooses "Execute now"
- **c3-orchestrator Phase 7**: Clarified to only apply for manual implementation path
- **c3-adr-transition**: Added integrity check for summary task (Step 1b)
  - Verifies summary task exists with correct metadata before transition
  - Added TaskList and TaskGet tools for integrity verification
  - Added example blocks for better triggering

### Fixed
- **Command descriptions**: Added trigger phrases to query.md and alter.md
- **c3.md command**: Added missing argument-hint field
- **c3 skill description**: Added "verify docs", "check documentation" triggers
- **c3-synthesizer**: Added token limit constraint for consistency

## [3.6.0] - 2026-01-21

### Added
- **Content Separation Verification (Phase 9)**: New audit phase that validates proper separation between components (domain logic) and refs (usage patterns)
  - The Separation Test: "Would this content change if we swapped the underlying technology?"
  - Detects missing refs for technologies, integration patterns in components, duplicated patterns
- **c3-content-classifier agent**: LLM-based content classification for Phase 9 audit
  - Identifies misplaced content between components and refs
  - Includes worked example with reasoning
  - Accepts optional technology context for better classification
- **content-separation.md reference**: Canonical definition for component vs ref separation
- **references/README.md**: Index of all reference files for discoverability
- **Proactive ref extraction in onboarding**: Section 1.2.3 guides extracting refs during component documentation

### Changed
- **ref.md template**: Added When/Why/Conventions structure for clearer ref documentation
- **Skill/agent relationship clarification**: c3-query and c3-alter skills now document their relationship to c3-navigator and c3-orchestrator agents
- **c3/SKILL.md Adopt mode simplified**: Now routes directly to /onboard skill without duplicating workflow
- **CLAUDE.md plugin guidance**: Updated to emphasize auto-discovery (explicit component paths break loading)

### Fixed
- **Skill name typo**: Fixed `onboarding-c3` → `/onboard` in c3/SKILL.md
- **Template typo**: Fixed `Applicable atterns` → `Applicable Patterns` in component.md

## [3.5.1] - 2026-01-21

### Fixed
- **Plugin loading broken by explicit paths**: Removed `commands`, `skills`, `agents`, `hooks` declarations from plugin.json - Claude Code uses auto-discovery
- **Release command validation**: Added Step 4 to validate plugin.json and remove explicit paths before every release

## [3.5.0] - 2026-01-21

### Added
- **ADR Lifecycle Tracking**: Complete tracking from acceptance to implementation
  - `base-commit` field captured when ADR transitions to accepted
  - `c3-verifier` script compares approved-files vs actual git changes
  - `c3-adr-transition` agent handles status transitions with verification
  - Verification results documented: matched, unplanned, untouched files
- **Phase 7 in c3-orchestrator**: Implementation completion with verification
- **Simplified c3-gate**: Now requires any accepted ADR (not specific file approval)
- **11-check validation hub in c3-synthesizer**: Consolidates all validation before ADR generation
  - 6 boundary checks from c3-impact: ownership, redundancy, sibling overlap, composition, leaky abstraction, correct layer
  - 4 ref checks from c3-patterns: follows ref, ref usage, missing ref, stale ref
  - 1 context alignment check in synthesizer

### Changed
- **Validation happens BEFORE ADR**: Loop closes at synthesizer, ADR is always valid when generated
- **c3-impact expanded**: Now performs boundary analysis (6 checks)
- **c3-patterns expanded**: Now performs ref health checks (4 checks)
- **c3-alter workflow simplified**: Removed Stage 4b (audit), validation now implicit in ADR creation

### Removed
- **c3-adr-auditor agent**: Consolidated into c3-synthesizer validation hub
- **c3-audit-adr skill**: No longer needed
- **Phase 5a in c3-orchestrator**: Audit phase removed, validation moved earlier

## [3.4.3] - 2026-01-21

### Added
- **c3-adr-auditor agent**: New agent that validates ADRs against C3 architectural principles before approval
  - Checks abstraction boundaries (component doing sibling's job)
  - Checks composition rules (orchestration vs hand-off)
  - Checks context alignment (contradicting Key Decisions)
  - Checks ref compliance (touching pattern domain without citation)
- **c3-audit-adr skill**: Direct invocation wrapper for ADR auditing (`/c3 audit-adr`)
- **Phase 5a in c3-orchestrator**: ADR audit gate between generation and acceptance - ADR cannot be accepted until auditor returns PASS
- **Stage 4b in c3-alter**: Audit step integrated into alter workflow checklist

### Changed
- **c3-orchestrator workflow**: Now includes mandatory audit step before user can accept ADR
- **c3-alter workflow**: Progress checklist updated with Stage 4b (audit) and Stage 4c (accept)

## [3.4.2] - 2026-01-21

### Fixed
- **SessionStart hook**: Use JSON `hookSpecificOutput.additionalContext` format for context injection - plain text output was showing as "Success" but not being injected into conversation context
- **hooks.json**: Add `matcher: "startup|resume"` to SessionStart hook for proper event matching

## [3.4.1] - 2026-01-20

### Fixed
- **plugin.json**: Removed explicit component paths - auto-discovery looks at parent of `.claude-plugin/` for skills/agents/commands/hooks directories, explicit paths were breaking loading when plugin.json is nested

## [3.4.0] - 2026-01-20

### Added
- **c3-ref skill**: New skill for managing cross-cutting patterns as first-class architecture artifacts
  - `add` mode: Create refs from discovered patterns with automatic citing component updates
  - `update` mode: Modify refs with impact analysis across all citing components
  - `list` mode: Show all refs with citation counts
  - `usage` mode: Show which components cite a specific ref

- **Constraint Chain Query**: New query type in c3-query skill
  - Ask "what constraints apply to c3-XXX" to see full inheritance chain
  - Shows Context → Container → Refs constraints with MAY/MUST NOT boundaries
  - Generates visual diagram of constraint inheritance

- **Phase 4b: Pattern Violation Gate** in c3-orchestrator
  - Refs are now blocking constraints - violations cannot be silently bypassed
  - Changes that break patterns require explicit override with justification in ADR
  - Options: update the pattern, override with justification, or rethink approach

- **Phase 8: Abstraction Boundaries** in audit-checks
  - Detects when components take on container/context responsibilities
  - Checks for: container bleeding, context bleeding, component orchestrating peers, ref bypass
  - FAIL on clear violations, WARN on potential issues needing review

- **Layer Constraints sections** in component.md and container.md templates
  - Explicit MUST/MUST NOT boundaries for each layer
  - Prevents abstraction violations through clear documentation

### Changed
- **ADR template**: Added "Pattern Overrides" section for documenting justified pattern violations

### Documentation
- **CLAUDE.md**: Updated plugin testing instructions with correct structure explanation

## [3.3.5] - 2026-01-20

### Fixed
- **plugin.json**: Added explicit `commands`, `skills`, `agents` declarations - auto-discovery wasn't working reliably, explicit declarations ensure Claude Code finds all components

### Documentation
- **CLAUDE.md**: Added comprehensive plugin structure checklist with pre-release requirements, plugin.json template, and common issues troubleshooting table

## [3.3.4] - 2026-01-20

### Fixed
- **plugin.json**: Added missing `hooks` field referencing `hooks/hooks.json` - hooks were defined but not loaded because plugin manifest didn't reference the hooks file

## [3.3.3] - 2026-01-20

### Fixed
- **plugin.json**: Removed explicit component paths - paths were relative to `.claude-plugin/` not plugin root, causing auto-discovery to fail. Now uses Claude Code's default auto-discovery.
- **onboard skill**: Aligned frontmatter `name: onboard` with directory name (was `onboarding-c3`)
- **onboard command**: Updated skill reference to match corrected skill name
- **c3-orchestrator agent**: Changed invalid color `orange` to `yellow`

### Documentation
- **CLAUDE.md**: Added plugin troubleshooting section with validation workflow

## [3.3.2] - 2026-01-20

### Fixed
- **plugin.json**: Added missing `skills` and `hooks` declarations - skills and hooks were not loading due to missing manifest entries

## [3.3.1] - 2026-01-19

### Fixed
- **TOC hook timing**: Moved TOC rebuild from Stop hook to PreToolUse on Bash (git commit) - ensures TOC changes are included in commits instead of being left behind
- **build-toc.sh**: Removed timestamp from TOC header, added content comparison to skip updates when nothing changed

## [3.3.0] - 2026-01-19

### Added
- **c3-orchestrator agent**: Multi-agent system for orchestrating architectural changes with Socratic dialogue before ADR generation
  - `c3-analyzer`: Current state extraction sub-agent (sonnet)
  - `c3-impact`: Dependency tracing and risk assessment sub-agent (sonnet)
  - `c3-patterns`: Convention checking against refs sub-agent (sonnet)
  - `c3-synthesizer`: Critical thinking synthesis sub-agent (opus)
- **ADR-gated code changes**: New `approved-files` field in ADR frontmatter
  - c3-gate now blocks Edit/Write unless file is in accepted ADR's approved-files list
  - Enforces analysis-before-change workflow in C3-adopted projects

### Changed
- **SessionStart hook**: Now routes to agents (c3-navigator for questions, c3-orchestrator for changes) instead of skills
- **c3 skill routing**: Updated mode selection to use agents for navigation and changes
- **Session harness**: Simplified with inline mermaid diagram showing routing flow

### Fixed
- Mermaid diagram syntax in session hooks (removed quotes in labels)

## [3.2.0] - 2026-01-19

### Added
- **c3-navigator agent**: Dedicated agent that triggers on any question in projects with `.c3/` directory. Provides architecture answers with visual diagrams via diashort service. Runs in separate context window for token efficiency.
- **c3-summarizer sub-agent**: Haiku-powered extraction agent that efficiently summarizes C3 documentation (~500 tokens output). Called by c3-navigator via Task tool for optimal token usage.

### Architecture
- Two-agent orchestration pattern: navigator identifies scope, summarizer extracts relevant facts
- 70-90% token savings compared to reading all `.c3/` docs in main context
- Auto-generated Mermaid diagrams rendered via diashort service
- Adaptive output format based on query type (structural, behavioral, flow)

## [3.1.0] - 2026-01-14

### Added
- **Goal-driven templates with staged onboarding workflow**: Templates now guide through staged adoption with clear goals at each step

### Fixed
- **Removed auxiliary references**: Completed migration from auxiliary category to refs system throughout all documentation and skills

## [3.0.0] - 2026-01-09

### Breaking Changes
- **Refs system replaces auxiliary category**: Components no longer have auxiliary category. Code references are now managed through explicit `refs/` documents that link architecture to code.
- **Templates restructured**: Merged `component-foundation.md` and `component-feature.md` into unified `component.md` with Goal-first structure.

### Added
- **Refs system**: New `templates/ref.md` for creating explicit code-to-architecture references
- **Ref lookup in c3-query**: Query skill now resolves refs to find relevant code locations
- **Ref maintenance in c3-alter**: Alter skill updates refs when code moves or changes
- **Goal-first template structure**: All templates now start with Goal section for clearer intent

### Changed
- **Context template**: Restructured with Goal-first approach, clearer actor/system sections
- **Container templates**: Service, database, queue templates all use Goal-first structure
- **Component template**: Unified template replaces separate foundation/feature templates
- **Audit checks**: Updated to validate refs instead of auxiliary category

### Documentation
- Layer navigation updated for refs resolution
- V3 structure documentation reflects refs system
- Removed all auxiliary category references

### Rationale
The auxiliary category was ambiguous and led to inconsistent categorization. Refs provide explicit, traceable links between architecture documentation and actual code, improving maintainability and audit accuracy.

## [2.5.0] - 2026-01-06

### Added
- **Complexity-first documentation approach**: Assess container complexity BEFORE documenting aspects
- **Harness complexity rules**: COMPLEXITY-BEFORE-ASPECTS, DISCOVERY-OVER-CHECKLIST, DEPTH-MATCHES-COMPLEXITY
- **Type-specific container templates**: `container-service.md`, `container-database.md`, `container-queue.md` with discovery prompts
- **External system templates**: `external.md` and `external-aspect.md` for documenting external dependencies
- **Complexity levels**: trivial → simple → moderate → complex → critical with clear documentation depth rules

### Changed
- **Harness as single source**: All complexity rules defined in `skill-harness.md`, templates reference it
- **Templates use discovery prompts**: No more pre-populated aspect checklists that bias AI
- **Container templates simplified**: Type-specific signals to scan, not assumptions to check off
- **c3 skill Adopt flow**: Now explicitly requires complexity-first and discovery-over-checklist

### Rationale
This release prevents AI bias from pre-populated checklists. By requiring complexity assessment first and discovery through code analysis, documentation reflects what actually exists rather than what templates assume.

## [2.4.0] - 2026-01-06

### Added
- **Component types reference**: New `references/component-types.md` consolidates Foundation/Feature/Auxiliary guidance with decision flowchart and dependency rules
- **Progressive complexity diagram**: Skill harness now shows simple → complex skill selection visually
- **Violation examples in harness**: Concrete wrong vs right examples for common mistakes

### Changed
- **Templates: diagram-first approach**: Linkage tables replaced with mermaid flowcharts, testing tables replaced with strategy prose
- **Skill descriptions**: All three skills now use third-person format with specific trigger phrases per plugin-dev guidelines
- **Template comments**: Multi-line AI hints consolidated to single-line format for token efficiency
- **Component templates**: Added type selector hints at top (e.g., `<!-- USE: Core primitives -->`)

### Documentation
- Added `docs/plans/2026-01-06-diagram-first-templates-design.md` capturing the design rationale

## [2.3.0] - 2026-01-05

### Added
- **Component references section**: All component templates now include `## References` section for explicit code-to-architecture links
- **Reference validation in audit**: New Phase 4 validates reference validity (symbols/paths/globs exist) and coverage (major code areas referenced)
- **Pre-execution checklist**: Plan template includes checklist item to update component references before implementation

### Changed
- Onboarding workflow now includes step to populate `## References` after drafting each component
- ADR workflow tracks "References Affected" and updates references during execution when code moves
- Audit procedure renumbered: References Validation (Phase 4), Diagram Accuracy (Phase 5), ADR Lifecycle (Phase 6)

### Documentation
- Added `docs/plans/2026-01-05-component-references-design.md` capturing the design rationale

## [2.2.0] - 2026-01-05

### Added
- **Staged onboarding with recursive learning loop**: Adoption now progresses through 5 stages (Context → Containers → Auxiliary → Foundation → Feature) with analysis and validation at each level
- **Socratic questioning**: All skills now use `AskUserQuestion` tool to clarify understanding before proceeding, continuing until confident (no open questions)
- **Bidirectional navigation**: When conflicts are discovered at deeper levels, the system ascends to fix parent documentation, then re-descends
- **Tiered assumptions**: High-impact changes (new External Systems, container boundaries) require user confirmation; low-impact changes (linkage reasoning, naming) auto-proceed
- **Progress tracking in ADR-000**: Adoption ADR now includes a Progress section showing documented vs remaining items per level
- **Intent clarification for queries**: c3-query now clarifies ambiguous queries before searching

### Changed
- `commands/onboard.md`: Complete rewrite implementing 5-stage recursive learning loop
- `skills/c3-alter/SKILL.md`: Rewritten with 7-stage workflow (Intent → Understand → Scope → ADR → Plan → Execute → Verify) using same loop pattern
- `skills/c3-query/SKILL.md`: Added Step 0 for Socratic intent clarification before navigation
- `templates/adr-000.md`: Added Adoption Progress table and Open Questions section

### Documentation
- Added `docs/plans/2026-01-05-staged-onboarding-design.md` capturing the full design rationale

## [2.1.0] - 2026-01-05

### Added
- Component file creation during onboarding (subagent creates docs for each discovered component)
- Category-specific component templates: `component-foundation.md`, `component-auxiliary.md`, `component-feature.md`
- Testing documentation at each layer: E2E (Context), Integration/Mocking/Fixtures (Container), Unit (Component)

### Changed
- Component categories simplified to Foundation/Auxiliary/Feature (removed Presentation)
- Audit checks streamlined to 4 essential checks (removed template-enforced redundancies)
- Container template now uses Foundation/Auxiliary/Feature sections with descriptions

### Removed
- Generic `component.md` template (replaced by category-specific templates)

## [2.0.1] - 2026-01-05

### Fixed
- `c3-init.sh` now falls back to `sed` when `envsubst` is not available

## [2.0.0] - 2026-01-02

### Added
- OpenCode support with dual-platform distribution
- `/release` command for version bumping and changelog generation
- Activation harnesses in all skills for consistent behavior
- Shared layer navigation pattern extracted to `references/layer-navigation.md`

### Changed
- **BREAKING**: Restructured skills from 5 to 3 (`c3`, `c3-query`, `c3-alter`)
- Skills now invoke other skills directly (not agents)
- Commands delegate to skills (`/c3`, `/query`, `/alter`, `/onboard`)

### Removed
- Dead code: orphan templates, duplicate references
- Unused agent definitions

### Documentation
- Added comprehensive CLAUDE.md with migration awareness
- Session history tracking in CLAUDE.md
