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

Create a new C3 project with a local cache and canonical .c3/ markdown.
Normal users rarely need this after initial setup.`,
	},
	{
		Name:     "read",
		Args:     "<entity-id>",
		OneLiner: "Output full entity content (frontmatter + body)",
		Help: `Usage: c3x read <entity-id> [--section <name>] [--json] [--full] [--cite]

Output the full content of an entity as markdown (default) or structured data.
Markdown output includes YAML frontmatter + body — same format accepted by write.

In agent mode (C3X_MODE=agent), structured output is TOON and body is truncated to 1500 chars with size hint.
Use --full to get the complete body.

Options:
  --section <name>   Output only the named section's content
  --json             Structured JSON output outside agent mode; agent mode stays TOON
  --full             Disable body truncation in agent mode
  --cite             Append canonical entity or section evidence handles

Examples:
  c3x read c3-101                        # full markdown output
  c3x read ref-jwt --json                # structured JSON
  c3x read c3-101 --section Goal         # just the Goal section
  c3x read c3-101 --section Goal --cite  # section content plus evidence handle
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
  set <id> <field> <value>      Single frontmatter field
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
  --json         Explicit JSON compatibility outside agent mode; default structured output is TOON
  --include-adr  Include ADR entities (hidden by default)

Use c3x canvas list to inspect available entity definitions.`,
	},
	{
		Name:     "check",
		OneLiner: "Validate docs, schema, citations, consistency",
		Help: `Usage: c3x check [--json] [--include-adr] [--fix] [--only <id>] [--only-touched [--since <ref>]] [--rule <rule-id>]

Layered validation (ADRs excluded by default; use --include-adr to validate them):
  Canonical seal + cache integrity
  Broken links, orphans, duplicates, missing parents
  Required sections empty/missing per schema
  Entity IDs in graph, cite consistency
(Fact-vs-code conformance is not here — that is a separate 'c3x eval' run.)

Options:
  --fix              Auto-fix entity/ref references that match by title (e.g., "API" → c3-1)
  --include-adr      Include ADR entities in validation
  --only <id>        Scope check to specific entity IDs (repeatable)
  --only-touched     Scope to entities affected by uncommitted changes.
  --since <ref>      Widen --only-touched window (e.g. --since main)
  --rule <rule-id>   Scope check to the set of entities that cite a rule (repeatable).
                     Errors if the rule has no citers. Composes with --only as union.`,
	},
	{
		Name:     "repair",
		OneLiner: "Rebuild local cache from canonical .c3/ and reseal",
		Help: `Usage: c3x repair [--json] [--include-adr] [--only <id>]

Rebuild the local C3 cache (.c3/c3.db) from canonical markdown, then re-export
canonical files so seals match. Use after a branch switch, selective merge, or
conflict resolution when 'c3x check' reports seal drift or cache divergence.

Repair does not invent fixes for content errors — it only realigns the cache
and seals to the current canonical text. If validation still fails afterwards,
the canonical files themselves need editing.

Options:
  --json             Structured output
  --include-adr      Include ADR entities in post-repair verification
  --only <id>        Scope verification to specific entity IDs (repeatable)`,
	},
	{
		Name:     "add",
		Args:     "<type> <slug>",
		OneLiner: "Create entity (auto-numbering; edges from body)",
		Help: `Usage: c3x add <type> <slug> [options]

Types come from c3x canvas list, including built-ins and project-defined document types.

Options:
  --container <id>       Parent container (component only)
  --feature              Feature numbering (10+) vs foundation (01-09)
  --file <path>          Read body from file instead of stdin
  --json                 Output as JSON (id, type, sections list)
  --dry-run              Validate content without creating the entity

Examples:
  c3x schema container > container.md
  c3x add container payments --file container.md
  c3x schema component > component.md
  c3x add component auth --container c3-1 --dry-run --file component.md
  c3x schema adr > adr.md
  c3x add adr config-change --file adr.md`,
	},
	{
		Name:     "canvas",
		Args:     "<list|read|add|write>",
		OneLiner: "Manage generic canvas definitions",
		Help: `Usage: c3x canvas list
       c3x canvas read <id>
       c3x canvas add <id> < canvas.md
       c3x canvas write <id> < canvas.md

Canvases are sealed canonical C3 markdown files under .c3/canvases/.
They are generic knowledge-graph contracts, not ADR-only templates.

Supported primitive column types:
  text, date, enum, table sections, edge<...>, cite, check, entity_id

Built-in examples cover C3 ADRs, atomic design changes, PM requirements,
PRDs, and user stories.

Examples:
  c3x canvas list
  c3x canvas read atomic-design-change
  c3x canvas read prd > prd-canvas.md
  c3x canvas add release-note --file canvas.md`,
	},
	{
		Name:     "set",
		Args:     "<id> <field> <value>",
		OneLiner: "Update a frontmatter field",
		Help: `Usage: c3x set <id> <field> <value>

Frontmatter fields: goal, summary, status, boundary, category, title, date, description

Note: set does NOT sync relationships. Citations come from the body — a column
the canvas marks as an edge (see 'c3x schema <type>'); author it at add, or
ride a change-unit patch for a frozen fact. Use 'c3x write <id> --section <name>'
for section body updates.

Examples:
  c3x set c3-101 goal "Handle JWT auth"`,
	},
	{
		Name:     "schema",
		Args:     "<type>",
		OneLiner: "Show required sections plus how to fill them",
		Help: `Usage: c3x schema <type> [--json]

Show known sections for an entity type.
Types come from c3x canvas list (system, container, component, ref, rule, adr, and project-defined document types).

Output includes:
  - REJECT IF: leading rejection contract for adr, ref, rule (read these BEFORE drafting)
  - purpose: what the section is for
  - fill: what the author must put there
  - rejected when: the failure that triggers rejection of that section

Structured output includes section guidance plus column types (text, date, enum, cite, check, entity_id, reference, evidence, edge<...>).

Examples:
  c3x schema component --json
  c3x canvas read adr
  c3x schema adr`,
	},
	{
		Name:     "lookup",
		Args:     "<file-path>",
		OneLiner: "Map file to component(s) + refs",
		Help: `Usage: c3x lookup <file-or-glob> [--json]

Map a file path (or glob pattern) to owning fact(s) via the eval-spec code
bindings (.c3/eval/<fact>.yaml 'code:'). Shows the fact's goal and cited refs
with their goals. Glob arguments expand against the project and show a file map.
Bracket paths ([id], [...slug]) for Next.js/SvelteKit routes work automatically.

Examples:
  c3x lookup src/auth/login.ts
  c3x lookup 'src/auth/**/*.ts'`,
	},
	{
		Name:     "eval",
		Args:     "[fact-id]",
		OneLiner: "Check a fact's claim against its external (conformance)",
		Help: `Usage: c3x eval [fact-id] [--json] [--policy]

Check a frozen fact's claim against the uncontrolled external it governs and
report a one-off verdict — holds / drift / needs-judgement — stamped with the
external state measured. A check you run on CI cadence, never an apply gate:
c3x eval always exits success; the verdict is the signal.

Specs live in .c3/eval/<fact-id>.yaml. The simplest is just a code binding:

  fact: c3-102
  code: [ cli/internal/store/** ]     # also what 'lookup' resolves a file against
  # no pipeline => default check: every glob must resolve to >=1 file

A pipeline composes five ops (gather/filter/transform/eval; loop fans them out):

  gather     acquire values — file (raw read) | command (mechanical, sh -c at repo
             root) | files (glob) | facts (id-glob) | code (a fact's globs) |
             literal | each (several, concatenated)
  filter     keep values matching   contains | matches
  transform  reshape each value     trim | first | lines
  eval       assert -> verdict      exists | equals | all_equal | contains_all |
             contains | count "<op> N" | judgement (surfaces; never auto-scored)
  loop       run a sub-pipeline per item of 'over', binding $item

Loop example — one roll-up verdict over a container's components:

  pipeline:
    - loop:
        over: { facts: "c3-1[0-9][0-9]" }   # each CLI component id
        do:
          - gather: { code: "$item" }        # its declared globs -> files
          - eval:   { exists: true }          # all resolve

Options:
  --json     Structured output for CI (TOON in agent mode)
  --policy   Report cache trust coverage instead of running verdicts

Examples:
  c3x eval                 # every spec
  c3x eval c3-203          # one fact's spec
  c3x eval --json          # machine output
  c3x eval --policy        # compact cache policy audit`,
	},
	{
		Name:     "search",
		Args:     "<query>",
		OneLiner: "Search content with semantic, keyword, and graph context",
		Help: `Usage: c3x search <query> [--hybrid] [--semantic] [--no-semantic] [--type <type>] [--limit N] [--json]

Search entity metadata and indexed markdown content. By default, results fuse
local semantic similarity, keyword/BM25 matches, and graph relationships, then
decorate hits with governing refs/rules.

Search auto-builds or refreshes the local semantic index on first use and reuses
fresh vectors on repeat runs. If the semantic model is unavailable, search falls
back to keyword plus graph without failing.

Options:
  --hybrid        Compatibility flag; graph context is already included by default
  --semantic      Compatibility flag; semantic is already enabled by default
  --no-semantic   Force keyword/graph ranking and skip semantic index refresh
  --type <type>   Restrict metadata search by entity type
  --limit N       Maximum number of results (default 5 in agent mode, 20 otherwise)
  --json          Structured output outside agent mode; agent mode stays TOON

Examples:
  c3x search "pool wait p95 latency"
  c3x search "owns a source path"
  c3x search traceparent --json --limit 3`,
	},
	{
		Name:     "index",
		OneLiner: "Build local semantic embeddings",
		Hidden:   true,
		Help: `Usage: c3x index [--json]

Maintenance only — never a correctness step. search self-heals the index on
demand; you do not need to re-index after a change.

Download the pinned all-MiniLM-L6-v2 ONNX model and matching onnxruntime shared
library into the user cache if missing, then rebuild SQLite entity embeddings.
Fat builds already embed these assets; thin builds fetch them on first semantic use.

Examples:
  c3x index
  c3x index --json`,
	},
	{
		Name:     "graph",
		Args:     "<entity-id>",
		OneLiner: "Subgraph from entity (LLM-friendly output)",
		Help: `Usage: c3x graph <entity-id> [--depth N] [--direction forward|reverse] [--format mermaid] [--json]

Emit a subgraph rooted at the given entity. Shows typed neighbors
and relationship edges.

Options:
  --depth N              BFS traversal depth (default: 1)
                         0 = entity only, 1 = direct neighbors, 2+ = multi-hop
  --direction forward    Impact analysis — children, affects, cited-by only
  --direction reverse    Reverse deps — what points to this entity only
                         (default: all neighbors in both directions)
  --format mermaid       Mermaid flowchart output (pipe to diashort for rendering)
  --json                 Explicit JSON compatibility outside agent mode; agent/default structured output is TOON

Examples:
  c3x graph c3-1                          # container + direct children
  c3x graph c3-101 --depth 0             # single entity detail
  c3x graph ref-jwt --depth 2            # ref + citers + their containers
  c3x graph c3-1 --format mermaid        # visual diagram
  c3x graph c3-101 --direction reverse   # what points to this component`,
	},
	{
		Name:     "git",
		Args:     "<subcommand>",
		OneLiner: "Install thin Git guardrails for C3 workflow",
		Hidden:   true,
		Help: `Usage: c3x git install

Installs a small pre-commit hook and Git file rules around the existing C3
workflow. The hook rejects staged c3.db, runs c3x check, and aborts the commit
if .c3/ still has unstaged changes. No custom merge driver or DB merge logic is installed.

What it installs:
  - pre-commit hook: c3x check
  - .gitattributes: mark legacy tracked .c3/c3.db as binary/generated
  - .c3/.gitignore: ignore c3.db and backup files within the C3 tree

Example:
  c3x git install`,
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
  - Removes Compliance Refs / Compliance Rules rows citing this entity
  - Removes row from parent container's Components table
  - Deletes the entity file

Options:
  --dry-run   Show cleanup plan without making changes

Examples:
  c3x delete c3-101
  c3x delete ref-jwt --dry-run`,
	},
	{
		Name:     "supersede",
		Args:     "<new-id> <old-id>",
		OneLiner: "Mark a terminal change doc superseded by a successor",
		Help: `Usage: c3x supersede <new-id> <old-id>

Record that <new-id> supersedes a terminal (done/superseded) change doc <old-id>:
flips <old-id> to superseded and writes the backlink <new-id> --supersedes--> <old-id>.

This is a mechanical operation. It never judges whether <new-id> is a legitimate
successor; it only enforces that:
  - <old-id> is terminal (a still-open decision cannot be superseded — finish it first)
  - the supersede would not form a cycle

Examples:
  c3x supersede adr-20260601-new adr-20260101-old`,
	},
	{
		Name:     "change",
		Args:     "<new|view|accept|apply|status|rebase|scaffold> <id>",
		OneLiner: "Author, review, and apply a change-unit (the only path that mutates a fact)",
		Help: `Usage: c3x change <new|view|accept|apply|status|rebase|scaffold> <change-unit-id>

A change-unit = reasoning (the doc) + change material (patch files in its folder,
.c3/changes/<id>/*.patch.md). Applying it is the ONLY legal mutation of a fact.

Each patch is one primitive, scoped: a no-base patch CREATES a new fact (full
content); a block patch (anchored by a cite handle) replaces / inserts / deletes
one block, leaving every sibling block frozen; frontmatter and retire patches
rename / re-edge / remove. The folder of *.patch.md files is the source of truth.

  new     scaffold the change-unit's patch folder
  view    the "files changed" panel: per-patch drift + state
  status  per-patch state derived from seal state (pending/applied/drifted/new)
  accept  record the one stored human judgment (status → accepted)
  apply   the switcher: drift + canvas + morph + retire gates, atomic all-or-nothing
  rebase  emit the drift bundle for re-authoring drifted patches
  scaffold stage a rung-climb: one empty insert patch per fact below its canvas bar

Apply runs four mechanical gates before any write — drift (every anchor fresh),
canvas (the merged body stays valid), morph (a reshaped fact-type leaves no
instance invalid), retire (no destruction strands a child or citer) — and is
atomic: one failing gate blocks every patch. --dry-run reports without writing.

Examples:
  c3x change view adr-20260601-auth
  c3x change apply adr-20260601-auth
  c3x change status adr-20260601-auth`,
	},
	{
		Name:     "migrate",
		OneLiner: "Sweep legacy statuses onto the canonical set (loud, one-time)",
		Help: `Usage: c3x migrate

One-time SWEEP & CLEAR ALL migration over the store. For every entity it:
  - CLEARS each fact's legacy 'active' status (facts have no status)
  - MAPS each change doc's legacy status onto the canonical set
    {open, accepted, done, superseded}, recording the lossy provisioned->done collapse
  - GRANDFATHERS old terminal ADRs (implemented/provisioned) to 'done' with no retro check
  - RECONCILES uncustomized materialized canvases to the current grammar (re-sealed);
    a customized canvas is left intact and reported for manual reconcile
  - RE-SEALS every entity

Every step is itemized in a loud, non-silent report. An unmappable status FAILS loud
and coerces nothing. migrate is the only path that may rewrite a terminal status.

Examples:
  c3x migrate`,
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
		if c.Hidden {
			continue
		}
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
Entity Types: container, component, ref, rule, adr (context created by init)

Global Options:
  --json                     Explicit JSON compatibility outside agent mode; agent/default structured output is TOON
  --c3-dir <path>            Override .c3/ auto-detection
  --force                    Confirm advanced reset commands
  -h, --help                 Show help
  -v, --version              Print version

Workflows:

  Understand the architecture before making changes:
    c3x list                # topology: goals, file coverage, ref usage
    c3x canvas list         # available entity definitions
    c3x schema component    # required sections for a given entity type
    c3x check               # validate refs, orphans, schema gaps
    c3x lookup src/auth.ts  # map code to owning component + refs

  Normal change flow:
    c3x schema component > auth.md   # author the cite in the edge-marked column (e.g. Governance)
    c3x add component auth --container c3-1 --file auth.md   # the cite edge wires at import
    c3x check

  After branch switch, selective merge, or conflict resolution:
    c3x check              # inspect canonical drift and consistency
    c3x repair             # rebuild cache and reseal if check reports seal drift

  Add a component to an existing container:
    c3x schema component > auth.md   # cite refs/rules in the edge-marked column
    c3x add component auth --container c3-1 --file auth.md
    c3x check

  Add a new domain (container + first component):
    c3x schema container > payments.md
    c3x add container payments --file payments.md
    c3x schema component > billing.md
    c3x add component billing --container c3-1 --file billing.md
    c3x check

  Remove an entity cleanly:
    c3x delete c3-101              # removes file + cleans all references
    c3x check
    c3x delete ref-jwt --dry-run   # preview cleanup plan without mutations

  Document a cross-cutting concern:
    c3x schema ref > rate-limiting.md
    c3x add ref rate-limiting --file rate-limiting.md
    # cite it from a component: author the row in its edge-marked column —
    # at add for a new component, or a change-unit patch for a frozen one
    c3x check

  Record an architectural decision:
    c3x schema adr > adr.md
    c3x add adr use-grpc --file adr.md`)

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
