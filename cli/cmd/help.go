package cmd

import (
	"fmt"
	"io"
	"strings"
)

// CommandMeta describes a single CLI command. This is the single source of
// truth for both `c3x --help` and `c3x capabilities`.
type CommandMeta struct {
	Name     string // e.g. "lookup"
	Args     string // e.g. "<file-or-glob>"
	OneLiner string // e.g. "Map file to component(s) + refs"
	Help     string // detailed help text shown by `c3x <cmd> --help`
	Hidden   bool   // if true, excluded from capabilities output
}

// Commands is the authoritative registry of all c3x commands.
var Commands = []CommandMeta{
	{
		Name:     "status",
		OneLiner: "Project dashboard (entity counts, coverage, pending ADRs)",
		Help: `Usage: c3x status [--json]

Project dashboard showing entity counts, code-map coverage, pending ADRs, and validation warnings.
This is the default output when c3x is invoked with no arguments.

Options:
  --json    Structured JSON output`,
	},
	{
		Name:     "init",
		OneLiner: "Scaffold .c3/ skeleton",
		Hidden:   true,
		Help: `Usage: c3x init

Create a new C3 project with a local cache and canonical .c3/ markdown.
Normal users rarely need this after initial setup.`,
	},
	{
		Name:     "verify",
		OneLiner: "Validate canonical .c3/ truth and refresh local cache if needed",
		Help: `Usage: c3x verify [--include-adr] [--only <id-or-path-or-glob>]

Verify that canonical .c3/ markdown is sealed, structurally valid, and in sync
with the local cache. If c3.db is missing or stale, verify rebuilds it from the
canonical text automatically.

ADRs are excluded by default so in-progress work orders do not block same-branch
work. Use --include-adr when finishing ADR work or before release/commit.

Use --only to verify a focused set of canonical docs by entity ID or path. Repeat
--only for multiple docs.

Use this in CI and pre-commit.

User rule: if you want confidence before commit, run c3x verify --include-adr.`,
	},
	{
		Name:     "repair",
		OneLiner: "Rebuild local cache from canonical .c3/, reseal, then verify",
		Help: `Usage: c3x repair

Recovery path after branch switches, selective merges, or manual conflict edits.
Rebuilds c3.db from canonical .c3/ markdown, rewrites sealed canonical files,
then verifies the result.

User rule: if C3 gets weird after Git operations, run c3x repair.`,
	},
	{
		Name:     "read",
		Args:     "<entity-id>",
		OneLiner: "Output full entity content (frontmatter + body)",
		Help: `Usage: c3x read <entity-id> [--section <name>] [--json] [--full]

Output the full content of an entity as markdown (default) or structured data.
Markdown output includes YAML frontmatter + body — same format accepted by write.

In agent mode (C3X_MODE=agent), structured output is TOON and body is truncated to 1500 chars with size hint.
Use --full to get the complete body.

Options:
  --section <name>   Output only the named section's content
  --json             Structured JSON output outside agent mode; agent mode stays TOON
  --full             Disable body truncation in agent mode

Examples:
  c3x read c3-101                        # full markdown output
  c3x read ref-jwt --json                # structured JSON
  c3x read c3-101 --section Goal         # just the Goal section
  c3x read c3-101 --full                 # full body in agent mode`,
	},
	{
		Name:     "write",
		Args:     "<entity-id>",
		OneLiner: "Replace entity content with validation (stdin)",
		Help: `Usage: c3x write <entity-id> [--section <name>] < content


Full write: replace entity content from stdin. Parses YAML frontmatter for
structured fields and validates required sections before accepting.
Relationships (uses, affects, scope) in frontmatter are updated consistently.

Section write (--section): replace only the named section. Skips full-document
validation — allows incremental filling of sections.

When to use set vs write:
  set --field <f> <v>           Single frontmatter field
  set --section <s> <v>         Single section (supports JSON tables, --append)
  set --stdin                   Batch fields + sections from JSON
  write --section <s> < stdin   Single section from stdin (plain text)
  write < stdin                 Full body + frontmatter (validates, syncs rels)

Examples:
  echo "New goal text" | c3x write c3-101 --section Goal
  cat updated-ref.md | c3x write ref-jwt`,
	},
	{
		Name:     "list",
		OneLiner: "Topology view with relationships",
		Help: `Usage: c3x list [--compact] [--flat] [--json] [--include-adr]

Topology view with system goal, entity goals, file coverage, and ref usage.
  --compact      Goals-only tree (no files/uses detail); with --json: lightweight output (id, type, title, parent, status only)
  --flat         Simple file list (id, type, path)
  --json         Machine-readable output
  --include-adr  Include ADR entities (hidden by default)`,
	},
	{
		Name:     "check",
		OneLiner: "Validate docs, schema, code refs, consistency",
		Help: `Usage: c3x check [--json] [--include-adr] [--fix]

Three-layer validation (ADRs excluded by default; use --include-adr to validate them):
  Layer 1: Broken links, orphans, duplicates, missing parents
  Layer 2: Required sections empty/missing per schema
  Layer 3: Code refs exist on disk, entity IDs in graph, cite consistency

Options:
  --fix            Auto-fix entity/ref references that match by title (e.g., "API" → c3-1)
  --include-adr    Include ADR entities in validation`,
	},
	{
		Name:     "add",
		Args:     "<type> <slug>",
		OneLiner: "Create entity (auto-numbering + wiring)",
		Help: `Usage: c3x add <type> <slug> [options]

Types: container, component, ref, rule, adr, recipe

Options:
  --container <id>       Parent container (component only)
  --feature              Feature numbering (10+) vs foundation (01-09)
  --goal <text>          Pre-fill goal in frontmatter + body
  --summary <text>       Pre-fill summary
  --boundary <text>      Pre-fill boundary (container only)
  --json                 Output as JSON (id, type, sections list)
  --dry-run              Validate content without creating the entity

ADR workflow:
  c3x schema adr                         # CLI-owned ADR creation contract
  cat complete-adr.md | c3x add adr <slug> # create the complete work order
  c3x read adr-YYYYMMDD-<slug> --full    # inspect before execution
  c3x verify --only adr-YYYYMMDD-<slug> --include-adr # prove focused ADR work
  c3x check --include-adr && c3x verify --include-adr # full validation before done

Examples:
  c3x add container payments --goal "Process payments" --boundary service
  c3x template component | c3x add component auth --container c3-1
  c3x add ref rate-limiting --goal "Consistent rate limiting"
  c3x add rule structured-logging --goal "Consistent structured logging"
  c3x template adr | c3x add adr use-grpc --json
  c3x add component auth --container c3-1 --dry-run < auth.md   # validate only
  c3x add recipe auth-flow`,
	},
	{
		Name:     "set",
		Args:     "<id> <field> <value>",
		OneLiner: "Update frontmatter field, section, or codemap patterns",
		Help: `Usage: c3x set <id> <field> <value>
       c3x set <id> --section <name> <value> [--append]
       c3x set <id> codemap "<patterns>" [--append|--remove]
       echo '{"fields":{...},"sections":{...},"codemap":[...]}' | c3x set <id> --stdin

Frontmatter fields: goal, summary, status, boundary, category, title, date, description

Special field "codemap" updates code-map patterns (comma-separated):
  Replace:  c3x set c3-101 codemap "src/auth/**,src/auth.go"
  Append:   c3x set c3-101 codemap "src/utils.go" --append
  Remove:   c3x set c3-101 codemap "src/old/**" --remove
  Clear:    c3x set c3-101 codemap ""

Section mode accepts text or JSON (array for replace, object for --append).
Component section updates validate the full resulting document.

Batch mode (--stdin) accepts a JSON payload with fields, sections, and codemap:
  {"fields": {"goal": "X"}, "sections": {"Choice": "..."}, "codemap": ["src/**"]}

Note: set does NOT sync relationships. Use wire/unwire for relationship changes,
or write with full frontmatter for bulk updates including relationship sync.

Examples:
  c3x set c3-101 goal "Handle JWT auth"
  c3x set c3-101 codemap "src/auth/**,src/auth.go"
  c3x set c3-101 codemap "src/new/**" --append
  c3x set c3-101 --section "Choice" "Use RS256 signed JWTs"
  c3x set c3-101 --section "Governance" --append '{"Reference":"ref-jwt","Type":"ref","Governs":"Token validation","Precedence":"ref beats local prose","Notes":"Required for auth"}'
  echo '{"fields":{"goal":"X"},"codemap":["src/**"]}' | c3x set c3-101 --stdin`,
	},
	{
		Name:     "wire",
		Args:     "<src> <tgt> [tgt2 ...]",
		OneLiner: "Link component to ref(s) (--remove to unlink)",
		Help: `Usage: c3x wire <source> <target> [target2 ...]
       c3x wire <source> cite <target> [target2 ...]
       c3x unwire <source> <target> [target2 ...]

Creates or removes cite relationships (updated atomically per target):
  1. source uses[] += target
  2. component source "Governance" table += row

Supports multiple targets in a single call for batch wiring.

"cite" is optional (it's the only supported relation type).

Examples:
  c3x wire c3-101 ref-jwt                            # single target
  c3x wire c3-101 ref-jwt ref-error-handling          # multiple targets
  c3x wire c3-101 cite ref-jwt ref-error-handling     # explicit cite
  c3x unwire c3-101 ref-jwt                           # remove link`,
	},
	{
		Name:     "schema",
		Args:     "<type>",
		OneLiner: "Show known sections for entity type",
		Help: `Usage: c3x schema <type> [--json]

Show known sections for an entity type.
Types: context, container, component, ref, rule, adr, recipe

JSON output includes column types (filepath, entity_id, enum, ref_id).

Example: c3x schema component --json`,
	},
	{
		Name:     "template",
		Args:     "<type>",
		OneLiner: "Output a fillable entity template that passes validation",
		Help: `Usage: c3x template <type>

Output a validation-ready markdown template for the given entity type.
The template passes c3x's own strict validation, so LLMs can use it
as a one-shot scaffold — fill in real content and pipe to c3x add.

Types: component, container, ref, rule, adr, recipe

Examples:
  c3x template component                    # see the scaffold
  c3x template component | c3x add component auth --container c3-1
  c3x template adr | c3x add adr use-grpc`,
	},
	{
		Name:     "codemap",
		OneLiner: "Scaffold code-map entries for all components, refs + rules",
		Help: `Usage: c3x codemap [--json]

Scaffold or update code-map entries in the store for every component, ref,
and rule in the C3 graph. Existing entries (patterns already set) are preserved.
New entries are added with empty pattern lists for you to fill in.

After scaffolding, edit patterns manually or with your LLM, then check:
  c3x coverage   # see how many files are mapped vs. unmapped

JSON output lists added and existing IDs. Default output is JSON;
set HUMAN=1 for human-readable text.

Example: c3x codemap`,
	},
	{
		Name:     "lookup",
		Args:     "<file-path>",
		OneLiner: "Map file to component(s) + refs",
		Help: `Usage: c3x lookup <file-or-glob> [--json]

Map a file path (or glob pattern) to owning component(s) from the code-map.
Shows component goal, summary, and cited refs with their goals.
Glob arguments expand against the project and show a file map.
Bracket paths ([id], [...slug]) for Next.js/SvelteKit routes work automatically.

Examples:
  c3x lookup src/auth/login.ts
  c3x lookup 'src/auth/**/*.ts'`,
	},
	{
		Name:     "coverage",
		OneLiner: "Code-map coverage stats",
		Help: `Usage: c3x coverage [--json]

Show code-map coverage: how many project files are mapped, excluded, or unmapped.
Coverage % = mapped / (total - excluded), so _exclude patterns don't penalize your score.
Uses git ls-files for file discovery (falls back to filesystem walk).
Default output is JSON; set HUMAN=1 for human-readable text.

Add _exclude patterns to mark intentional exclusions:
  _exclude:
    - "**/*.test.ts"
    - "**/*.spec.ts"
    - dist/**

Example: c3x coverage --json`,
	},
	{
		Name:     "graph",
		Args:     "<entity-id>",
		OneLiner: "Subgraph from entity (LLM-friendly output)",
		Help: `Usage: c3x graph <entity-id> [--depth N] [--direction forward|reverse] [--format mermaid] [--json]

Emit a subgraph rooted at the given entity. Shows typed neighbors,
file paths from code-map, and relationship edges.

Options:
  --depth N              BFS traversal depth (default: 1)
                         0 = entity only, 1 = direct neighbors, 2+ = multi-hop
  --direction forward    Impact analysis — children, affects, cited-by only
  --direction reverse    Reverse deps — what points to this entity only
                         (default: all neighbors in both directions)
  --format mermaid       Mermaid flowchart output (pipe to diashort for rendering)
  --json                 Machine-readable JSON output

Examples:
  c3x graph c3-1                          # container + direct children
  c3x graph c3-101 --depth 0             # single entity detail
  c3x graph ref-jwt --depth 2            # ref + citers + their containers
  c3x graph c3-1 --format mermaid        # visual diagram
  c3x graph c3-101 --direction reverse   # what points to this component`,
	},
	{
		Name:     "query",
		Args:     "<search-terms>",
		OneLiner: "Full-text search across entities (ADRs excluded by default)",
		Help: `Usage: c3x query <search-terms> [--type <type>] [--limit N] [--include-adr] [--json]

Search entity titles, goals, summaries, and bodies using FTS5 with BM25 ranking.
ADRs are excluded by default — they are historical records, not source of truth.

Operators:
  auth security     Implicit AND — both terms must match
  auth OR security  Either term matches
  auth | security   Same as OR
  auth AND handler  Explicit AND
  auth NOT jwt      Exclude matches containing "jwt"

Special characters (commas, periods, brackets, etc.) are stripped automatically.

Options:
  --type <type>     Filter results to entity type (component, ref, adr, etc.)
  --include-adr     Include ADR entities in results (excluded by default)
  --limit N         Max results (default: 20)
  --json            Machine-readable output

Examples:
  c3x query authentication
  c3x query "error handling" --type ref
  c3x query "store OR walker"
  c3x query frontmatter --limit 5 --json
  c3x query migration --include-adr`,
	},
	{
		Name:     "diff",
		OneLiner: "Show uncommitted changes (human-readable changelog)",
		Help: `Usage: c3x diff [--mark <commit-hash>] [--json]

Render all entity mutations since the last commit mark.

Options:
  --mark <hash>   Stamp current changes with commit hash (use in post-commit hook)
  --json          Output raw changelog entries as JSON

Examples:
  c3x diff                    # show pending changes
  c3x diff --mark abc123      # mark changes as committed`,
	},
	{
		Name:     "impact",
		Args:     "<entity-id>",
		OneLiner: "Transitive impact analysis (who depends on this?)",
		Help: `Usage: c3x impact <entity-id> [--depth N] [--include-code] [--json]

Find all entities affected by changes to the given entity.
Traverses reverse 'uses' + forward 'affects' relationships.

With --include-code, merges the documented citation graph with a grep-derived
import graph over the target's code-map sources. Components that call into
the target but are not documented in .c3/ are flagged [uncited]. Caller files
with no owning component are listed separately as codemap coverage gaps.

Options:
  --depth N         Max traversal depth (default: 3)
  --include-code    Merge documented citations with grep-derived callers (off by default)
  --json            Machine-readable output

Examples:
  c3x impact c3-101                  # what breaks if auth changes?
  c3x impact ref-jwt --depth 5       # deep impact of JWT ref
  c3x impact c3-201 --include-code   # include undocumented callers as [uncited]`,
	},
	{
		Name:     "export",
		OneLiner: "Render canonical markdown from the local cache (advanced)",
		Hidden:   true,
		Help: `Usage: c3x export [<output-dir>]

Render canonical markdown from the local cache to a target directory.
This is an advanced debug command. The shared truth is sealed .c3/ text;
normal workflows should use c3x verify and c3x repair instead.

Default output dir is the current .c3/ directory.

Examples:
  c3x export /tmp/c3-export    # inspect a temporary snapshot
  c3x export                   # rewrite current .c3/ from local cache`,
	},
	{
		Name:     "sync",
		Args:     "<subcommand>",
		OneLiner: "Sync canonical markdown from the local cache",
		Hidden:   true,
		Help: `Usage: c3x sync export [<output-dir>]
       c3x sync check [<output-dir>]

Sync commands write the canonical text representation used for Git review
and branch merges. Canonical markdown is sealed; sync check verifies both
the exported bytes and each file's semantic seal.

Subcommands:
  export [dir]   Export the current DB to canonical markdown (default: .c3/)
  check [dir]    Verify the current tree matches canonical markdown

Examples:
  c3x sync export
  c3x sync export /tmp/c3-export
  c3x sync check`,
	},
	{
		Name:     "git",
		Args:     "<subcommand>",
		OneLiner: "Install thin Git guardrails for C3 workflow",
		Help: `Usage: c3x git install

Installs a small pre-commit hook and Git file rules around the existing C3
workflow. The hook rejects staged c3.db, runs c3x verify, and aborts the commit
if .c3/ still has unstaged changes. No custom merge driver or DB merge logic is installed.

What it installs:
  - pre-commit hook: c3x verify
  - .gitattributes: mark legacy tracked .c3/c3.db as binary/generated
  - .c3/.gitignore: ignore c3.db and backup files within the C3 tree

Example:
  c3x git install`,
	},
	{
		Name:     "nodes",
		Args:     "<entity-id>",
		OneLiner: "List content nodes with IDs, types, and hashes",
		Help: `Usage: c3x nodes <entity-id> [--json]

Display the content node tree for an entity. Every heading, paragraph, list item,
table row, and code block has a unique ID and SHA256 content hash.

Options:
  --json   Machine-readable JSON array

Examples:
  c3x nodes c3-101           # text table with indentation
  c3x nodes c3-101 --json    # full node data as JSON`,
	},
	{
		Name:     "hash",
		Args:     "<entity-id>",
		OneLiner: "Show entity content hash (root merkle)",
		Help: `Usage: c3x hash <entity-id> [--recompute]

Output the root merkle hash for an entity's content. This is a single SHA256 that
changes whenever any content in the entity changes — useful for change detection.

Options:
  --recompute   Recompute from node content and compare with stored hash.
                Shows OK if they match, DRIFT if they differ.

Examples:
  c3x hash c3-101              # stored root hash
  c3x hash c3-101 --recompute  # verify integrity`,
	},
	{
		Name:     "versions",
		Args:     "<entity-id>",
		OneLiner: "List content version history",
		Help: `Usage: c3x versions <entity-id> [--json]

Show all content versions for an entity, newest first. Each write creates a new
version with a full content snapshot and root merkle hash.

Options:
  --json   Machine-readable JSON array

Examples:
  c3x versions c3-101
  c3x versions ref-jwt --json`,
	},
	{
		Name:     "version",
		Args:     "<entity-id> <n>",
		OneLiner: "Show content at a specific version",
		Help: `Usage: c3x version <entity-id> <version-number>

Output the full content of an entity as it was at a specific version.

Examples:
  c3x version c3-101 1    # content at version 1
  c3x version c3-101 3    # content at version 3`,
	},
	{
		Name:     "prune",
		Args:     "<entity-id>",
		OneLiner: "Delete old content versions",
		Help: `Usage: c3x prune <entity-id> --keep <n>

Delete old content versions, keeping the most recent N. Versions marked with a
git commit hash (via c3x diff --mark) are always preserved.

Options:
  --keep <n>   Number of recent versions to keep (required, minimum 1)

Examples:
  c3x prune c3-101 --keep 10   # keep last 10, delete the rest`,
	},
	{
		Name:     "migrate",
		OneLiner: "Populate content node tree for all entities",
		Help: `Usage: c3x migrate [--dry-run] [--json]
       c3x migrate repair-plan
       c3x migrate repair <id> --section <name> < content.md
       c3x migrate --continue

Populates the content node tree for entities that don't have one yet.
Reads legacy body content, parses into element-level nodes with hashing
and versioning. Warns about entities with stale frontmatter in body.

Options:
  --dry-run     Show what would be migrated without making changes
  --json        Emit machine-readable blockers outside agent mode; agent mode stays TOON
  --continue    Resume the same migration intent after scoped repairs and import

Repair flow:
  c3x migrate --dry-run --json       # machine-readable blockers, writesMade:false
  c3x migrate repair-plan            # exact safe steps plus current blockers
  c3x migrate repair <id> --section Goal < goal.md
  c3x cache clear
  c3x import --force
  c3x migrate --continue

Examples:
  c3x migrate --dry-run --json   # preview blockers
  c3x migrate repair-plan        # safe repair loop
  c3x migrate --continue         # resume after repair/import`,
	},
	{
		Name:     "cache",
		Args:     "clear",
		OneLiner: "Clear disposable local C3 cache files",
		Help: `Usage: c3x cache clear

Deletes local cache files under .c3/ without touching canonical markdown:
  .c3/c3.db*
  .c3/.c3.import.tmp.db*

Use this instead of manual rm commands during migration repair.`,
	},
	{
		Name:     "migrate-legacy",
		OneLiner: "Import .c3/ markdown files into SQLite database",
		Hidden:   true,
		Help: `Usage: c3x migrate-legacy [--keep-originals]

Import legacy .c3/ markdown files into a local cache.
Advanced migration path for older file-based projects.

Options:
  --keep-originals   Don't delete original .md files after import`,
	},
	{
		Name:     "import",
		OneLiner: "Rebuild SQLite database from tracked markdown tree",
		Hidden:   true,
		Help: `Usage: c3x import [--force]

Rebuild the local c3.db cache from canonical .c3/ markdown.
Advanced plumbing behind c3x repair. Normal users should prefer c3x repair.

Imports entities, relationships, code-map entries, and content nodes, then
clears the changelog to establish a clean baseline.
Sealed files are trusted by default; broken or missing seals require --force.

Safety:
  --force   Required when c3.db already exists. Creates a timestamped backup
            before replacing the database.

Example:
  c3x import --force`,
	},
	{
		Name:     "delete",
		Args:     "<id>",
		OneLiner: "Remove entity + clean all references",
		Help: `Usage: c3x delete <id> [--dry-run]

Remove an entity and clean all references to it across the graph.

Safety:
  - Refuses to delete c3-0 (context root)
  - Refuses to delete containers with children (lists them)

Cleanup:
  - Removes id from uses[], affects[], scope[], sources[] on referencing entities
  - Removes Governance table rows citing this entity
  - Removes row from parent container's Components table
  - Removes code-map.yaml entry
  - Deletes the entity file

Options:
  --dry-run   Show cleanup plan without making changes

Examples:
  c3x delete c3-101
  c3x delete ref-jwt --dry-run`,
	},
	{
		Name:     "marketplace",
		Args:     "<subcommand>",
		OneLiner: "Manage marketplace rule sources",
		Help: `Usage: c3x marketplace <subcommand> [options]

Subcommands:
  add <github-url>          Clone marketplace repo, register as source
  list [--source] [--tag]   List available rules across sources
  show <rule-id>            Preview a rule's content
  update [<source-name>]    Pull latest from registered sources
  remove <source-name>      Unregister source + delete cache

Options:
  --source <name>   Filter by source name
  --tag <tag>       Filter rules by tag
  --json            Machine-readable output

Examples:
  c3x marketplace add https://github.com/org/go-patterns
  c3x marketplace list --tag reliability
  c3x marketplace show rule-error-handling
  c3x marketplace update
  c3x marketplace remove go-patterns`,
	},
}

// buildGlobalHelp generates the global help text from the command registry.
func buildGlobalHelp() string {
	var b strings.Builder
	b.WriteString(`c3x - architecture-aware CLI for C3 projects

Usage: c3x <command> [args] [options]

Commands:
`)
	maxWidth := 0
	for _, c := range Commands {
		w := len(c.Name)
		if c.Args != "" {
			w += 1 + len(c.Args)
		}
		if w > maxWidth {
			maxWidth = w
		}
	}

	for _, c := range Commands {
		if c.Hidden {
			continue
		}
		label := c.Name
		if c.Args != "" {
			label += " " + c.Args
		}
		fmt.Fprintf(&b, "  %-*s  %s\n", maxWidth, label, c.OneLiner)
	}

	b.WriteString(`
Entity Types: container, component, ref, rule, adr, recipe (context created by init)

Global Options:
  --json                     Machine-readable output
  --c3-dir <path>            Override .c3/ auto-detection
  --force                    Confirm advanced reset commands like import
  -h, --help                 Show help
  -v, --version              Print version

Workflows:

  Understand the architecture before making changes:
    c3x list                # topology: goals, file coverage, ref usage
    c3x schema component    # required sections for a given entity type
    c3x check               # validate refs, orphans, schema gaps
    c3x lookup src/auth.ts  # map code to owning component + refs

  Normal change flow:
    c3x add component auth --container c3-1 --goal "JWT auth for all services"
    c3x wire c3-101 cite ref-jwt
    c3x set c3-101 --section "Code References" '[{"File":"src/auth.ts","Purpose":"Auth middleware"}]'
    c3x verify

  After branch switch, selective merge, or conflict resolution:
    c3x repair             # rebuild cache, reseal canonical text, verify

  Add a component to an existing container:
    c3x add component auth --container c3-1 --goal "JWT auth for all services"
    c3x wire c3-101 cite ref-jwt
    c3x set c3-101 --section "Code References" '[{"File":"src/auth.ts","Purpose":"Auth middleware"}]'
    c3x verify

  Add a new domain (container + first component):
    c3x add container payments --goal "Process payments" --boundary service
    c3x add component billing --container c3-1 --goal "Invoice generation via Stripe"
    c3x verify

  Remove an entity cleanly:
    c3x delete c3-101              # removes file + cleans all references
    c3x verify
    c3x delete ref-jwt --dry-run   # preview cleanup plan without mutations

  Document a cross-cutting concern:
    c3x add ref rate-limiting --goal "Consistent rate limiting across services"
    c3x wire c3-101 cite ref-rate-limiting
    c3x set ref-rate-limiting --section "Code References" '[{"File":"src/middleware/rate.ts","Purpose":"Rate limiter"}]'
    c3x verify

  Trace an end-to-end concern:
    c3x add recipe auth-flow
    printf '# Auth Flow\n\n## Goal\n\nTrace auth end to end.\n\n## Sources\n\n- c3-101\n' | c3x write recipe-auth-flow
    c3x verify

  Record an architectural decision:
    c3x add adr use-grpc --goal "Migrate to gRPC for internal services" --json
    # use the returned ADR id for follow-up set/write commands`)

	b.WriteString(`

  Browse and adopt marketplace rules:
    c3x marketplace add https://github.com/org/go-patterns
    c3x marketplace list --tag reliability
    c3x marketplace show rule-error-handling`)

	return b.String()
}

var globalHelp = buildGlobalHelp()

// ShowHelp prints command help or global help.
func ShowHelp(command string, w io.Writer) {
	for _, c := range Commands {
		if c.Name == command {
			if c.Hidden {
				break
			}
			fmt.Fprintln(w, c.Help)
			return
		}
	}
	fmt.Fprintln(w, globalHelp)
}

// ShowCapabilities writes a markdown table of all non-hidden commands.
func ShowCapabilities(w io.Writer) {
	fmt.Fprintln(w, "| Command | What it does |")
	fmt.Fprintln(w, "|---------|-------------|")
	for _, c := range Commands {
		if c.Hidden {
			continue
		}
		label := "c3x " + c.Name
		if c.Args != "" {
			label += " " + c.Args
		}
		fmt.Fprintf(w, "| `%s` | %s |\n", label, c.OneLiner)
	}
}
