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
		Name:     "init",
		OneLiner: "Scaffold .c3/ skeleton",
		Hidden:   true,
		Help: `Usage: c3x init

Scaffold .c3/ skeleton (config, README, refs/, adr/).`,
	},
	{
		Name:     "list",
		OneLiner: "Topology view with relationships",
		Help: `Usage: c3x list [--compact] [--flat] [--json] [--include-adr]

Topology view with system goal, entity goals, file coverage, and ref usage.
  --compact      Goals-only tree (no files/uses detail)
  --flat         Simple file list (id, type, path)
  --json         Machine-readable full output
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

Types: container, component, ref, adr, recipe

Options:
  --container <id>       Parent container (component only)
  --feature              Feature numbering (10+) vs foundation (01-09)
  --goal <text>          Pre-fill goal in frontmatter + body
  --summary <text>       Pre-fill summary
  --boundary <text>      Pre-fill boundary (container only)
  --json                 Output created entity as JSON (id + path)

Examples:
  c3x add container payments --goal "Process payments" --boundary service
  c3x add component auth --container c3-1 --goal "JWT authentication"
  c3x add ref rate-limiting --goal "Consistent rate limiting"
  c3x add adr use-grpc --goal "Migrate to gRPC" --json
  c3x add recipe auth-flow`,
	},
	{
		Name:     "set",
		Args:     "<id> <field> <value>",
		OneLiner: "Update frontmatter field",
		Help: `Usage: c3x set <id> <field> <value>
       c3x set <id> --section <name> <value> [--append]

Frontmatter fields: goal, summary, status, boundary, category, title, date
Array via JSON:     scope, affects, refs

Section mode accepts text or JSON (array for replace, object for --append):
  c3x set c3-101 goal "Handle JWT auth"
  c3x set c3-101 --section "Code References" '[{"File":"src/auth.ts","Purpose":"Auth"}]'
  c3x set c3-101 --section "Dependencies" --append '{"Direction":"IN","What":"creds","From/To":"c3-102"}'`,
	},
	{
		Name:     "wire",
		Args:     "<src> [cite] <tgt>",
		OneLiner: "Link component to ref (--remove to unlink)",
		Help: `Usage: c3x wire <source> <target>
       c3x wire <source> cite <target>
       c3x wire --remove <source> <target>

Creates or removes a cite relationship (updated atomically):
  1. source frontmatter uses[] += target
  2. source "Related Refs" table += row

Options:
  --remove   Remove the cite relationship instead of creating it
             (idempotent — no error if not wired)

"cite" is optional (it's the only supported relation type).

Examples:
  c3x wire c3-101 ref-jwt            # create link
  c3x wire c3-101 cite ref-jwt       # same, explicit cite
  c3x wire --remove c3-101 ref-jwt   # remove link`,
	},
	{
		Name:     "schema",
		Args:     "<type>",
		OneLiner: "Show known sections for entity type",
		Help: `Usage: c3x schema <type> [--json]

Show known sections for an entity type.
Types: context, container, component, ref, adr, recipe

JSON output includes column types (filepath, entity_id, enum, ref_id).

Example: c3x schema component --json`,
	},
	{
		Name:     "codemap",
		OneLiner: "Scaffold code-map.yaml for all components + refs",
		Help: `Usage: c3x codemap [--json]

Scaffold or update .c3/code-map.yaml with stubs for every component and ref
in the C3 graph. Existing entries (patterns already set) are preserved.
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

Map a file path (or glob pattern) to owning component(s) from code-map.yaml.
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

Add _exclude patterns to code-map.yaml to mark intentional exclusions:
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
  - Removes Related Refs table rows citing this entity
  - Removes row from parent container's Components table
  - Removes code-map.yaml entry
  - Deletes the entity file

Options:
  --dry-run   Show cleanup plan without making changes

Examples:
  c3x delete c3-101
  c3x delete ref-jwt --dry-run`,
	},
}

// buildGlobalHelp generates the global help text from the command registry.
func buildGlobalHelp() string {
	var b strings.Builder
	b.WriteString(`c3x - architecture-aware CLI for C3 projects

Usage: c3x <command> [args] [options]

Commands:
`)
	// Find max width for alignment
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
		label := c.Name
		if c.Args != "" {
			label += " " + c.Args
		}
		fmt.Fprintf(&b, "  %-*s  %s\n", maxWidth, label, c.OneLiner)
	}

	b.WriteString(`
Entity Types: container, component, ref, adr, recipe (context created by init)

Global Options:
  --json                     Machine-readable output
  --c3-dir <path>            Override .c3/ auto-detection
  -h, --help                 Show help
  -v, --version              Print version

Workflows:

  Understand the architecture before making changes:
    c3x list              # topology: goals, file coverage, ref usage
    c3x schema component  # required sections for a given entity type
    c3x check             # validate refs, orphans, schema gaps

  Add a component to an existing container:
    c3x add component auth --container c3-1 --goal "JWT auth for all services"
    c3x wire c3-101 cite ref-jwt
    c3x set c3-101 --section "Code References" '[{"File":"src/auth.ts","Purpose":"Auth middleware"}]'
    c3x check

  Add a new domain (container + first component):
    c3x add container payments --goal "Process payments" --boundary service
    c3x add component billing --container c3-1 --goal "Invoice generation via Stripe"
    c3x check

  Remove an entity cleanly:
    c3x delete c3-101              # removes file + cleans all references
    c3x delete ref-jwt --dry-run   # preview cleanup plan without mutations

  Document a cross-cutting concern:
    c3x add ref rate-limiting --goal "Consistent rate limiting across services"
    c3x wire c3-101 cite ref-rate-limiting
    c3x set ref-rate-limiting --section "Code References" '[{"File":"src/middleware/rate.ts","Purpose":"Rate limiter"}]'

  Trace an end-to-end concern:
    c3x add recipe auth-flow
    # Edit .c3/recipes/recipe-auth-flow.md: add description + sources

  Record an architectural decision:
    c3x add adr use-grpc --goal "Migrate to gRPC for internal services"
    c3x set adr-1 status accepted
    c3x set adr-1 --section "Context" "We need lower latency between services"`)

	return b.String()
}

var globalHelp = buildGlobalHelp()

// ShowHelp prints command help or global help.
func ShowHelp(command string, w io.Writer) {
	for _, c := range Commands {
		if c.Name == command {
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
